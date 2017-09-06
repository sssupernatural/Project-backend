package server

import (
	"golang.org/x/net/context"
	umrpc "external/grpc/userManagerRPC"
	smrpc "external/grpc/searchManagerRPC"
	"external/comm"
	"dataCenter/dataClient"
	"userManager/rpc"
)

type UMServerConfig struct {
	Addr           string
    DataCenterConf dataClient.DataCenterDesc
}

type UMServer struct {
	addr string
	dataClient *dataClient.DataCenterClient
}

var SMCG *rpc.SMClientGroup

func New(conf *UMServerConfig) *UMServer {
	dc := dataClient.New(&conf.DataCenterConf)
	dc.InitUsersData()

	return &UMServer{
		addr: conf.Addr,
		dataClient: dc,
	}
}

//必须在启动SMCG后才能调用该初始化
func (s *UMServer)Init() {
	allUsers, err := s.dataClient.GetAllUsers()
	if err != nil {
		return
	}

	logger.Printf("all users : %v\n", allUsers)

	client := SMCG.GetClient()
	defer SMCG.ReturnClient(client)

	var num int32
	num = 0
	insertUsers := make([]*comm.UserInfo, 0)
	for _, user := range allUsers {
		u := comm.UserInfo{}
		u = user
		if u.Abilities == nil {
			continue
		}
		logger.Printf("append user : %v\n", u)
		insertUsers = append(insertUsers, &u)

		num++
		if num == 10 {
			ir := &smrpc.InsertUserRecordsReq{
				UserRecordNum: num,
				Users: insertUsers,
			}
			resp, err := client.InsertUserRecords(context.Background(), ir)
			if resp.Comm.ErrorCode != comm.RetOK {
				return
			} else if err != nil {
				return
			}

			num = 0
			insertUsers = make([]*comm.UserInfo, 0)
		}
	}

	if num != 0 {
		ir := &smrpc.InsertUserRecordsReq{
			UserRecordNum: num,
			Users: insertUsers,
		}
		for _, tmpu := range ir.Users {
			logger.Printf("send user : %v\n", *tmpu)
		}
		resp, err := client.InsertUserRecords(context.Background(), ir)
		if resp.Comm.ErrorCode != comm.RetOK {
			return
		} else if err != nil {
			return
		}
	}

	return
}

func generateRespComm(errCode int32) *umrpc.UserManagerRespComm {
	return &umrpc.UserManagerRespComm{
		ErrorCode: errCode,
		ErrorMsg: comm.GetErrMsg(errCode),
	}
}

func generateRegisterUserResp(errCode int32, ui *comm.UserInfo) *umrpc.UserManagerRegisterResp {
	return &umrpc.UserManagerRegisterResp{
		Comm: generateRespComm(errCode),
		UserInfo: ui,
	}
}

func generateLoginUserResp(errCode int32, ui *comm.UserInfo) *umrpc.UserManagerLoginResp {
	return &umrpc.UserManagerLoginResp{
		Comm: generateRespComm(errCode),
		UserInfo: ui,
	}
}

func generateAddUserInfoResp(errCode int32, ui *comm.UserInfo) *umrpc.UserManagerAddUserInfoResp {
	return &umrpc.UserManagerAddUserInfoResp{
		Comm: generateRespComm(errCode),
		UserInfo: ui,
	}
}

func generateLogoutUserResp(errCode int32) *umrpc.UserManagerLogoutResp {
	return &umrpc.UserManagerLogoutResp{
		Comm: generateRespComm(errCode),
	}
}

func (s *UMServer)RegisterUser(ctx context.Context, rgReq *umrpc.UserManagerRegisterReq) (*umrpc.UserManagerRegisterResp, error) {
	logger.Infof("Receive a [User Register] request%v.\n", rgReq)

	//保存用户登录信息
	err := s.dataClient.PutUserCheckInfo(rgReq.UserCheckInfo)
	if err != nil {
		logger.Errorf("[DataClient]Put user check info failed, err:%s, req:%v.\n", err, rgReq)
		resp := generateRegisterUserResp(comm.RetPutUserCheckInfoErr, nil)
		return resp, nil
	}

	//生成该用户的用户信息
	err = s.dataClient.PutUserInfo(&comm.UserInfo{
		PhoneNumber: rgReq.UserCheckInfo.PhoneNumber,
		Name: rgReq.UserCheckInfo.Name,
		Status: comm.UserStatusOnline,
	})
	//生成用户信息失败同时删除用户登录信息，用户注册返回失败
	if err != nil {
		delErr := s.dataClient.DeleteUserCheckInfo(rgReq.UserCheckInfo)
		if delErr != nil {
			logger.Errorf("[DataClient]Delete user check info failed when put user info failed.err: %s, req:%v.\n", delErr, rgReq)
		}
		logger.Errorf("[DataClient]Put user info failed, err:%s, req:%v.\n", err, rgReq)
		resp := generateRegisterUserResp(comm.RetPutUserInfoErr, nil)
		return resp, nil
	}

	userInfo := comm.UserInfo{
		PhoneNumber: rgReq.UserCheckInfo.PhoneNumber,
	}

	//获取该注册用户的用户信息
	ui, err := s.dataClient.GetUserInfoByPhoneNumber(rgReq.UserCheckInfo.PhoneNumber)
	if err != nil {
		//获取该注册用户的用户信息失败，删除该用户登录信息和用户信息，并返回失败
		delErr := s.dataClient.DeleteUserCheckInfo(rgReq.UserCheckInfo)
		if delErr != nil {
			logger.Errorf("[DataClient]Delete user check info failed when get user info failed.err: %s, req:%v.\n", delErr, rgReq)
		}

		delErr = s.dataClient.DeleteUserInfo(&userInfo)
		if delErr != nil {
			logger.Errorf("[DataClient]Delete user info failed when get user info failed.err: %s, req:%v.\n", delErr, rgReq)
		}

		logger.Errorf("[DataClient]Get user info failed, err:%s, req:%v.\n", err, rgReq)

		resp := generateRegisterUserResp(comm.RetGetUserInfoErr, nil)
		return resp, nil
	}

	//用户注册成功，返回响应
	resp := generateRegisterUserResp(comm.RetOK, ui)

	return resp, nil
}

func (s *UMServer)LoginUser(ctx context.Context, liReq *umrpc.UserManagerLoginReq) (*umrpc.UserManagerLoginResp, error) {
	logger.Infof("Receive a [User Login] request%v.\n", liReq)

	ci := &comm.UserCheckInfo{
		PhoneNumber: liReq.UserCheckInfo.PhoneNumber,
	}

	err := s.dataClient.GetUserCheckInfo(ci)
	if err != nil {
		var retCode int32

		if err.Error() == dataClient.RetRecordNotFound {
			logger.Errorf("User not exist, req:%v.\n", liReq)
			retCode = comm.RetUserNotExist
		} else {
			logger.Errorf("[DataClient]Get user check info failed, err:%s, req:%v.\n", err, liReq)
			retCode = comm.RetGetUserCheckInfoErr
		}

		resp := generateLoginUserResp(retCode, nil)
		return resp, nil
	}

	if ci.Password != liReq.UserCheckInfo.Password {
		logger.Errorf("User Password not match, get[%s], expected[%s]. req:%v.\n", liReq.UserCheckInfo.Password, ci.Password)
		resp := generateLoginUserResp(comm.RetUserPasswordNotMatch, nil)
		return resp, nil
	}

	ui, err := s.dataClient.GetUserInfoByPhoneNumber(liReq.UserCheckInfo.PhoneNumber)
	if err != nil {
		logger.Errorf("[DataClient]Get user info failed, err:%s, req:%v.\n", liReq)
		resp := generateLoginUserResp(comm.RetGetUserInfoErr, nil)
		return resp, nil
	}

	if ui.Status == comm.UserStatusOnline {
		logger.Errorf("User[ID:%d] status is online, can not login.\n", ui.ID)
		resp := generateLoginUserResp(comm.RetUserOnlineLoginErr, nil)
		return resp, nil
	}

	err = s.dataClient.UpdateUserStatusByPhoneNumber(ui.PhoneNumber, comm.UserStatusOnline)
	if err != nil {
		logger.Errorf("User status update to online failed. err:%s, req:%v.\n", err, liReq)
		resp := generateLoginUserResp(comm.RetUpdateUserStatusErr, nil)
		return resp, nil
	}

	ui.Status = comm.UserStatusOnline
	resp := generateLoginUserResp(comm.RetOK, ui)
	return resp, nil
}

func (s *UMServer)LogoutUser(ctx context.Context, loReq *umrpc.UserManagerLogoutReq) (*umrpc.UserManagerLogoutResp, error) {
	err := s.dataClient.UpdateUserStatusByPhoneNumber(loReq.UserCheckInfo.PhoneNumber, comm.UserStatusOffline)
	if err != nil {
		logger.Errorf("[DataClient]update user status to offline failed, err:%s, req:%v.\n", err, loReq)
		resp := generateLogoutUserResp(comm.RetUpdateUserStatusErr)
		return resp, nil
	}

	resp := generateLogoutUserResp(comm.RetOK)
	return resp, nil

}

func (s *UMServer)AddUserInfo(ctx context.Context, auiReq *umrpc.UserManagerAddUserInfoReq) (*umrpc.UserManagerAddUserInfoResp, error) {
	err := s.dataClient.UpdateUserInfo(auiReq.NewUserInfo)
	if err != nil {
		logger.Errorf("[DataClient]update user info failed, err:%s, req:%v.\n", err, auiReq)
		resp := generateAddUserInfoResp(comm.RetUpdateUserInfoErr, nil)
		return resp, nil
	}

	ui, err := s.dataClient.GetUserInfoByID(auiReq.NewUserInfo.ID)
	if err != nil {
		logger.Errorf("[DataClient]get user info from data center failed, err:%s, req:%v.\n", err, auiReq)
		resp := generateAddUserInfoResp(comm.RetGetUserInfoErr, nil)
		return resp, nil
	}

	client := SMCG.GetClient()
	defer SMCG.ReturnClient(client)
	user := make([]*comm.UserInfo, 0)
	user = append(user, ui)
	sr := &smrpc.InsertUserRecordsReq{
		UserRecordNum: 1,
		Users: user,
	}
	sresp, err := client.InsertUserRecords(context.Background(), sr)
	if sresp.Comm.ErrorCode != comm.RetOK || err != nil {
		logger.Errorf("Insert User to search manager failed, err : %s, errmsg : %s, req : %v.",
		err, sresp.Comm.ErrorMsg, sr)
	}

	resp := generateAddUserInfoResp(comm.RetOK, ui)
	return resp, nil
}
