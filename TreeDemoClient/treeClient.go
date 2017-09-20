package main

import (
	"fmt"
	"strings"
	"TreeDemoClient/command"
	"external/comm"
)

var curUser command.User

const (
	CommandQuit            string = "q"
	CommandUserRegister    string = "ru"
	CommandUserLogin       string = "iu"
	CommandUserLogout      string = "ou"
	CommandUserAddInfo     string = "au"
	CommandUserEditAbis    string = "ea"
	CommandUserEditLocs    string = "el"

	CommandCreateTask      string = "ct"
	CommandQueryTask       string = "qt"
	CommandAcceptTask      string = "at"
	CommandChooseResponser string = "cr"
	CommandFulfilTask      string = "ft"
	CommandEvaluateTask    string = "et"

)

var comd string

func showCommandHelp() {
	fmt.Println("q  -> 退出")
	fmt.Println("ru -> 注册")
	fmt.Println("iu -> 登录")
	fmt.Println("ou -> 登出")
	fmt.Println("au -> 添加信息")
	fmt.Println("ea -> 编辑能力")
	fmt.Println("el -> 编辑位置")
	fmt.Println("ct -> 发起任务")
	fmt.Println("qt -> 查询任务")
	fmt.Println("at -> 接受任务")
	fmt.Println("cr -> 选择响应者")
	fmt.Println("ft -> 完成任务")
	fmt.Println("et -> 评价并结束任务")
}

func init() {
	curUser.ID = command.InvalidUserID
	curUser.TIs = make([]*comm.TaskInfoWithUsers, 0)
}

func main() {

	go curUser.QueryTaskAuto()

	command.ClientIDToAbiMap = command.GenerateSysIDToAbiMap()
	command.ClientAbiToIDMap = command.GenerateSysAbiToIDMap()

	for {
		fmt.Print("Command : ")
		fmt.Scanln(&comd)
		strings.TrimSuffix(comd, "\n")

		switch comd {
		default                     : showCommandHelp()
		case CommandQuit            : {
			if curUser.ID != command.InvalidUserID {
				curUser.Logout()
			}
			return
		}

		case CommandUserRegister    : command.RegisterUser(&curUser)
		case CommandUserLogin       : command.LoginUser(&curUser)
		case CommandUserLogout      : curUser.Logout()
		case CommandUserAddInfo     : curUser.AddInfo()
		case CommandUserEditAbis    : curUser.EditAbilities()
		case CommandUserEditLocs    : curUser.EditLocations()
		case CommandCreateTask      : curUser.CreateTask()
		case CommandQueryTask       : curUser.QueryTask()
		case CommandAcceptTask      : curUser.AcceptTask()
		case CommandChooseResponser : curUser.ChooseResponser()
		case CommandFulfilTask      : curUser.FulfilTask()
		case CommandEvaluateTask    : curUser.EvaluateTask()
		}

	}
}
