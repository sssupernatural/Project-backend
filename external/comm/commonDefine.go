package comm

const (
	UserStatusOnline  int32 = 10000
	UserStatusOffline int32 = 10001

	UserSexMale       int32 = 10003
	UserSexFemale     int32 = 10004
	UserSexUnknown    int32 = 10005

	TaskStatusCreating      int32  = 11000
	TaskStatusWaitingAccept int32  = 11001
	TaskStatusWaitingChoose int32  = 11002
	TaskStatusProcessing    int32  = 11003
	TaskStatusFulfilled     int32  = 11004
	TasKStatusSearchResponserFailed = 11005
	TasKStatusSearchResponserNone = 11006


	TaskDecisionAccept int32 = 12001
	TaskDecisionRefuse int32 = 12002

	TaskFulfilStatusFinished int32 = 13001
	TaskFulfilStatusDoing    int32 = 13002

)

const 	(
	MsgUserStatusOnline  string = "在线"
	MsgUserStatusOffline string = "离线"
	MsgUserStatusUnknown string = "?"

	MsgUserSexMale       string = "男"
	MsgUserSexFemale     string = "女"
	MsgUserSexUnknown    string = "男OR女"

	MsgTaskStatusCreating        string  = "创建中"
	MsgTaskStatusWaitingAccept string  = "等待响应者"
	MsgTaskStatusWaitingChoose string  = "等待请求者选择响应者"
	MsgTaskStatusProcessing    string  = "进行中"
	MsgTaskStatusFulfilled     string  = "已完成"
	MsgTaskStatusUnknown       string  = "异常"
	MsgTasKStatusSearchResponserFailed string = "搜索备选响应者失败"
	MsgTasKStatusSearchResponserNone string = "没有找到合适的响应者"

	MsgTaskFulfilStatusFinished string = "任务已完成"
	MsgTaskFulfilStatusDoing    string = "任务进行中"

)

func DescStatus(status int32) string {
	switch status {
	default: return MsgUserStatusUnknown
	case UserStatusOnline : return MsgUserStatusOnline
	case UserStatusOffline: return MsgUserStatusOffline
	}
}

func DescTaskStatus(status int32) string {
	switch status {
	default: return MsgTaskStatusUnknown
	case TaskStatusCreating : return MsgTaskStatusCreating
	case TaskStatusWaitingAccept: return MsgTaskStatusWaitingAccept
	case TaskStatusWaitingChoose: return MsgTaskStatusWaitingChoose
	case TaskStatusProcessing: return MsgTaskStatusProcessing
	case TaskStatusFulfilled: return MsgTaskStatusFulfilled
	case TasKStatusSearchResponserFailed: return MsgTasKStatusSearchResponserFailed
	case TasKStatusSearchResponserNone: return MsgTasKStatusSearchResponserNone
	}
}

func DescSex(sex int32) string {
	switch sex {
	default: return MsgUserSexUnknown
	case UserSexMale : return MsgUserSexMale
	case UserSexFemale: return MsgUserSexFemale
	}
}

func DescFulfilStatus(fulfilStatus int32) string {
	switch fulfilStatus {
	default: return MsgTaskFulfilStatusDoing
	case TaskFulfilStatusFinished : return MsgTaskFulfilStatusFinished
	case TaskFulfilStatusDoing: return MsgTaskFulfilStatusDoing
	}
}

