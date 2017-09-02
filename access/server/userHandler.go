package server

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"external/comm"
	umrpc "external/grpc/userManagerRPC"
	"context"
	"errors"
)

const (
	UserActionRegister string = "REGISTER"
	UserActionLogin    string = "LOGIN"
	UserActionLogout   string = "LOGOUT"
)

func usersHandler(w http.ResponseWriter, req *http.Request) {
	logger.Infoln("Receive a users request")

	if req.Method != "POST" {
		logger.Errorf("Users request method must be POST, reccived req method is %s.\n", req.Method)
		http.Error(w, "User method is not POST!", 400)
		return
	}

	defer req.Body.Close()
	reqData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Errorf("Read request body failed, err: %s.\n", err)
		http.Error(w, "Server read request body failed!", 400)
		return
	}

	var curUserAction comm.UserAction

    err = json.Unmarshal(reqData, &curUserAction)
	if err != nil {
		logger.Errorf("Unmarshal request body failed, err: %s.\n", err)
		http.Error(w, "Server unmarshal request body failed!", 400)
		return
	}

	switch curUserAction.Action {
	case UserActionRegister : err = handleUserActionRegister(w, &curUserAction)
	case UserActionLogin    : err = handleUserActionLogin(w, &curUserAction)
	case UserActionLogout   : err = handleUserActionLogout(w, &curUserAction)
	default:
		logger.Errorf("Unexpected user action [%s]!.\n", curUserAction.Action)
		http.Error(w, "Unexpected user action!", 400)
		return
	}

	if err == nil {
		logger.Infoln("Handle users request succeed!")
	} else {
		w.Header().Set("ErrorMsg", err.Error())
		http.Error(w, "Server handle request error!", 500)
	}

	return
}

func handleUserActionRegister(w http.ResponseWriter, a *comm.UserAction) error {
	userRegisterReq := &umrpc.UserManagerRegisterReq {
		UserCheckInfo: &comm.UserCheckInfo{
			PhoneNumber: a.CheckInfo.PhoneNumber,
			Password:    a.CheckInfo.Password,
			Name:        a.CheckInfo.Name,
		},
	}

	umClient := UMCG.GetClient()
	defer UMCG.ReturnClient(umClient)
	resp, err := umClient.RegisterUser(context.Background(), userRegisterReq)
	if err != nil {
		logger.Errorf("Access RPC User Manager Failed, err[%s].\n", err)
		return err
	}

	if resp.Comm.ErrorCode != comm.RetOK {
		logger.Errorf("User Manager returns error[%s], register user failed. UserInfo[%s, %s, %s].\n",
		resp.Comm.ErrorMsg, a.CheckInfo.PhoneNumber, a.CheckInfo.Password, a.CheckInfo.Name)
		return errors.New(resp.Comm.ErrorMsg)
	}

	respData, err := json.Marshal(resp.UserInfo)
	if err != nil {
		logger.Errorf("Marshal response data failed! err : %s.\n", err)
		return err
	}

	w.Write(respData)

	return nil
}

func handleUserActionLogin(w http.ResponseWriter, a *comm.UserAction) error {
	userLoginReq := &umrpc.UserManagerLoginReq {
		UserCheckInfo: &comm.UserCheckInfo{
			PhoneNumber: a.CheckInfo.PhoneNumber,
			Password:    a.CheckInfo.Password,
		},
	}

	umClient := UMCG.GetClient()
	defer UMCG.ReturnClient(umClient)
	resp, err := umClient.LoginUser(context.Background(), userLoginReq)
	if err != nil {
		logger.Errorf("Access RPC User Manager Failed, err[%s].\n", err)
		return err
	}

	if resp.Comm.ErrorCode != comm.RetOK {
		logger.Errorf("User Manager returns error[%s], Login user failed. UserInfo[%s, %s, %s].\n",
			resp.Comm.ErrorMsg, a.CheckInfo.PhoneNumber, a.CheckInfo.Password, a.CheckInfo.Name)
		return errors.New(resp.Comm.ErrorMsg)
	}

	respData, err := json.Marshal(resp.UserInfo)
	if err != nil {
		logger.Errorf("Marshal response data failed! err : %s.\n", err)
		return err
	}

	w.Write(respData)

	return nil
}

func handleUserActionLogout(w http.ResponseWriter, a *comm.UserAction) error {
	userLogoutReq := &umrpc.UserManagerLogoutReq{
		UserCheckInfo: &comm.UserCheckInfo{
			PhoneNumber: a.CheckInfo.PhoneNumber,
		},
	}

	umClient := UMCG.GetClient()
	defer UMCG.ReturnClient(umClient)
	resp, err := umClient.LogoutUser(context.Background(), userLogoutReq)
	if err != nil {
		logger.Errorf("Access RPC User Manager Failed, err[%s].\n", err)
		return err
	}

	if resp.Comm.ErrorCode != comm.RetOK {
		logger.Errorf("User Manager returns error[%s], Logout user failed. phone_number[%s].\n",
			resp.Comm.ErrorMsg, a.CheckInfo.PhoneNumber)
		return errors.New(resp.Comm.ErrorMsg)
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func handleUserAddInfo(w http.ResponseWriter, u *comm.UserInfo) error {
	userAddInfoReq := &umrpc.UserManagerAddUserInfoReq{
		NewUserInfo: u,
	}

	logger.Printf("receive ui : %v.\n", u)

	umClient := UMCG.GetClient()
	defer UMCG.ReturnClient(umClient)
	resp, err := umClient.AddUserInfo(context.Background(), userAddInfoReq)
	if err != nil {
		logger.Errorf("Access RPC User Manager Failed, err[%s].\n", err)
		return err
	}

	if resp.Comm.ErrorCode != comm.RetOK {
		logger.Errorf("User Manager returns error[%s], Add user info failed. user[%v].\n",
			resp.Comm.ErrorMsg, u)
		return errors.New(resp.Comm.ErrorMsg)
	}

	respData, err := json.Marshal(resp.UserInfo)
	if err != nil {
		logger.Errorf("Marshal response data failed! err : %s.\n", err)
		return err
	}

	w.Write(respData)

	return nil
}

func userInfoHandler(w http.ResponseWriter, req *http.Request) {
	logger.Infoln("Receive a user info request")

	if req.Method != "POST" {
		logger.Errorf("User info request method must be POST, reccived req method is %s.\n", req.Method)
		http.Error(w, "User info method is not POST!", 400)
		return
	}

	defer req.Body.Close()
	reqData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Errorf("Read request body failed, err: %s.\n", err)
		http.Error(w, "Server read request body failed!", 400)
		return
	}

	var curUserInfo comm.UserInfo

	err = json.Unmarshal(reqData, &curUserInfo)
	if err != nil {
		logger.Errorf("Unmarshal request body failed, err: %s.\n", err)
		http.Error(w, "Server unmarshal request body failed!", 400)
		return
	}

	err = handleUserAddInfo(w, &curUserInfo)
	if err == nil {
		logger.Infoln("Handle user info request succeed!")
	} else {
		w.Header().Set("ErrorMsg", err.Error())
		http.Error(w, "Server handle request error!", 500)
	}

	return
}

