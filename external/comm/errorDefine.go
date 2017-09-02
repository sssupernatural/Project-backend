package comm

const (
	RetOK                int32 = 0	//成功
	RetInvalidParameters int32 = -10000	//非法参数

	//Data operation Error
	RetPutUserCheckInfoErr int32 = -11000	//向Data Center写入UserCheckInfo失败
	RetGetUserCheckInfoErr int32 = -11001	//从Data Center读取UserCheckInfo失败
	RetPutUserInfoErr      int32 = -11002	//向Data Center写入UserInfo失败
	RetGetUserInfoErr      int32 = -11003   //从Data Center读取UserInfo失败
	RetUpdateUserStatusErr int32 = -11004   //向Data Center更新UserInfo的用户状态失败
	RetUpdateUserInfoErr   int32 = -11005   //向Data Center更新UserInfo失败

	RetPutTaskInfoErr      int32 = -11006	//向Data Center写入TaskInfo失败
	RetUpdateTaskInfoErr   int32 = -11007   //向Data Center更新TaskInfo失败

	//User Manager Error
	RetUserNotExist         int32 = -12000	//访问的用户不存在
	RetUserPasswordNotMatch int32 = -12001	//用户密码错误
	RetUserOnlineLoginErr   int32 = -12002	//用户重复登录错误

	//Task Manager Error
	RetUserHasNoTask        int32 = -13000  //用户当前无相关任务
	RetNoSuchTaks           int32 = -13001  //用户当前无相关任务

)

const (
	MsgRetOK string = "OK"
	MsgRetInvalidParameters string = "Invalid Parameters"

	//Data operation Center
	MsgRetPutUserCheckInfoErr string = "Put user check info to data center failed."
	MsgRetGetUserCheckInfoErr string = "Get user check info from data center failed."
	MsgRetPutUserInfoErr string = "Put user info to data center failed."
	MsgRetGetUserInfoErr string = "Get user info from data center failed."
	MsgRetUpdateUserStatusErr string = "Update user status to data center failed."
    MsgRetUpdateUserInfoErr string = "Update user info to data center failed."

	MsgRetPutTaskInfoErr string = "Put task info to data center failed."
	MsgRetUpdateTaskInfoErr string = "Update task info to data center failed."

	//User Manager Error
	MsgRetUserNotExist string = "User not exist."
	MsgRetUserPasswordNotMatch string = "User's password not match."
	MsgRetUserOnlineLoginErr string = "User is now online, can not login."

	//Task Manager Error
	MsgRetUserHasNoTask        string = "User has no task."
	MsgRetNoSuchTaks        string = "No such task."
)

func GetErrMsg(code int32) string {
	switch code {
	case RetOK : return MsgRetOK
	case RetInvalidParameters : return MsgRetInvalidParameters

	//DataClientErr
	case RetPutUserCheckInfoErr: return MsgRetPutUserCheckInfoErr
	case RetGetUserCheckInfoErr: return MsgRetGetUserCheckInfoErr
	case RetPutUserInfoErr: return MsgRetPutUserInfoErr
	case RetGetUserInfoErr: return MsgRetGetUserInfoErr
	case RetUpdateUserStatusErr: return MsgRetUpdateUserStatusErr
	case RetUpdateUserInfoErr: return MsgRetUpdateUserInfoErr
	case RetPutTaskInfoErr: return MsgRetPutTaskInfoErr
	case RetUpdateTaskInfoErr: return MsgRetUpdateTaskInfoErr

	//User Manager Error
	case RetUserNotExist: return MsgRetUserNotExist
	case RetUserPasswordNotMatch: return MsgRetUserPasswordNotMatch
	case RetUserOnlineLoginErr: return MsgRetUserOnlineLoginErr

	//Task Manager Error
	case RetUserHasNoTask: return MsgRetUserHasNoTask
	case RetNoSuchTaks: return MsgRetNoSuchTaks

	}

	return MsgRetOK
}
