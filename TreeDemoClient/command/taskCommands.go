package command

import (
	"bytes"
	"net/http"
	"fmt"
	"external/comm"
	"encoding/json"
	"time"
	"io/ioutil"
	"errors"
	"strconv"
)

const (
	URLTasks    string = "http://localhost:8080/tasks"
)

func printTaskInfo(t *comm.TaskInfoWithUsers) {
	fmt.Println("|-----------------------------------------TaskInfo-----------------------------------------")
	fmt.Printf("| 任务编号 : %d\n", t.ID)
	fmt.Printf("| 任务状态 : %s\n", comm.DescTaskStatus(t.Status))
	fmt.Println("| 任务信息 :")
	printTaskCreateInfo(t.Desc)
	fmt.Println("| 任务请求者信息 :")
	printUserInfo(t.Requester)
	fmt.Println("| 任务响应者信息 :")
	if t.ChosenResponser == nil || len(t.ChosenResponser) == 0{
		fmt.Println("| |任务还没有被选择的响应者")
	} else {
		for index, r := range t.ChosenResponser {
			if t.FulfilStatus[index] == comm.TaskFulfilStatusDoing {
				fmt.Println("| |正在进行任务的响应者 : ")
			} else if t.FulfilStatus[index] == comm.TaskFulfilStatusFinished {
				fmt.Println("| |已经完成任务的响应者 : ")
			}
			printUserInfo(r)
		}
	}
	fmt.Println("| 任务备选响应者信息 :")
	if t.Responsers == nil || len(t.Responsers) == 0 {
		fmt.Println("| |任务还没有备选响应者")
	} else {
		for _, r := range t.Responsers {
			printUserInfo(r)
		}
	}
	fmt.Println("|-----------------------------------------TaskInfo-----------------------------------------")

}

func printTaskCreateInfo(t *comm.TaskCreateInfo)  {
	fmt.Println("| -------------------------TaskDesc-------------------------")
	fmt.Printf("| | 任务简述 : %s\n", t.Brief)
	fmt.Printf("| | 响应者性别 : %s\n", comm.DescSex(t.Sex))
	fmt.Printf("| | 响应者年龄区间 : %d-%d\n", t.AgeMin, t.AgeMax)
	fmt.Print("| | 响应者位置 : ")
	for _, l := range t.Locations {
		fmt.Printf("[%f,%f] |", l.Longitude, l.Latitude)
	}
	fmt.Println()
	fmt.Print("| | 响应者能力 : ")
	for _, a := range t.Abilities.ABIs {
		fmt.Printf("[%s] |", a.ABI)
	}
	fmt.Println()
	fmt.Println("| -------------------------TaskDesc-------------------------")
}
func (u *User)CreateTask() {
	if u.ID == InvalidUserID {
		fmt.Println("用户未登录，无法创建任务。")
		return
	}

	createTaskInfo := &comm.TaskCreateInfo {
		RequesterID: u.ID,
		Locations: make([]*comm.Location, 0),
	}

	var option int32
	fmt.Print("输入任务描述：")
	fmt.Scanf("%s", &createTaskInfo.Brief)
	fmt.Print("选择响应者性别序号：1-[男] | 2-[女] | 3-[任意]: ")
	fmt.Scanf("%d", &option)
	if option == 1 {
		createTaskInfo.Sex = comm.UserSexMale
	} else if option == 2{
		createTaskInfo.Sex = comm.UserSexFemale
	} else {
		createTaskInfo.Sex = comm.UserSexUnknown
	}

	fmt.Print("输入响应者年龄区间(eg:18-30): ")
	fmt.Scanf("%d-%d", &createTaskInfo.AgeMin, &createTaskInfo.AgeMax)

	for {
		fmt.Print("添加响应者位置坐标(结束请输入(0,0)): ")
		curloc := &comm.Location{}
		fmt.Scanf("%f,%f", &curloc.Longitude, &curloc.Latitude)
		if curloc.Latitude == 0 && curloc.Longitude == 0 {
			break
		} else {
			createTaskInfo.Locations = append(createTaskInfo.Locations, curloc)
		}
	}

	fmt.Println("系统能力图如下：")
	printSysAbiAll()
	taskAbis := make([]int, 0)
	for {
		fmt.Print("添加响应者能力(输入能力的系统能力编号,结束请输入'f'): ")
		var curabi string
		fmt.Scanf("%s", &curabi)
		if curabi == "f" {
			break
		} else {
			id, _ := strconv.Atoi(curabi)
			if _, ok := ClientIDToAbiMap[id]; ok {
				taskAbis = append(taskAbis, id)
			}
		}
	}

	createTaskInfo.Abilities = u.generateAbiHeapByAbiIndexes(taskAbis)

	printTaskCreateInfo(createTaskInfo)
	fmt.Println("创建中...")

	data, err := json.Marshal(createTaskInfo)
	if err != nil {
		fmt.Printf("创建任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest("POST", URLTasks, dataReader)
	if err != nil {
		fmt.Printf("创建任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	transport := http.Transport {
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("创建任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("创建任务时返回了错误的结果，请重试。错误信息：%s.\n", resp.Status)
		return
	}

	fmt.Println("任务创建成功！")

	return
}

func (u *User)QueryTaskGetTaskInfo() []*comm.TaskInfoWithUsers {
	if u.ID == InvalidUserID {
		fmt.Println("用户未登录，无法查询任务。")
		return nil
	}

	var curTaskAction comm.TaskAction

	curTaskAction.Action = "QUERY"
	curTaskAction.UserID = u.ID

	data, err := json.Marshal(curTaskAction)
	if err != nil {
		fmt.Printf("查询用户任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return nil
	}

	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest("GET", URLTasks, dataReader)
	if err != nil {
		fmt.Printf("查询用户任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return nil
	}

	transport := http.Transport {
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: &transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("查询用户任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return nil
	}

	var tis [20]*comm.TaskInfoWithUsers

	/*
	for i:= 0; i < 10; i++ {
		tis[i] = &comm.TaskInfoWithUsers{}
	}
	*/

	//tis[0] = &comm.TaskInfoWithUsers{}

	/*

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	for {
		curTI := &comm.TaskInfoWithUsers{}
		if err := dec.Decode(curTI); err != nil && err == io.EOF {
			break
		} else if err == nil {
			fmt.Printf("n%v\n", curTI)
			tis = append(tis, curTI)
		} else {
			fmt.Printf("err : %s\n", err)
		}
	}
	*/

	if resp.StatusCode != 200 {
		if resp.Header.Get("ErrorMsg") == comm.GetErrMsg(comm.RetUserHasNoTask) {
			return nil
		} else {
			fmt.Printf("查询用户任务时返回了错误的结果，请重试。错误信息：%s, 具体信息: %s.\n",
				resp.Status, resp.Header.Get("ErrorMsg"))
			return nil
		}
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("查询用户任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return nil
	}

	err = json.Unmarshal(respBody, &tis)
	if err != nil {
		fmt.Printf("查询用户任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return nil
	}

	outTis := make([]*comm.TaskInfoWithUsers, 0)

	for _, ti := range tis {
		if ti != nil {
			outTis = append(outTis, ti)
		}
	}

	return outTis
}

func (u *User)QueryTask() {
	fmt.Println("查询用户相关任务中...")
	tis := u.QueryTaskGetTaskInfo()
	if tis == nil || len(tis) == 0 {
		fmt.Println("用户当前没有相关任务。")
		return
	}

	fmt.Println("任务查询成功，当前用户的相关任务如下：")
	for _, t := range tis {
		if t.Requester.ID == u.ID {
			fmt.Println("你创建的任务：")
		} else {
			if t.ChosenResponser != nil && len(t.ChosenResponser) != 0 {
				crTag := false
				for _, cr := range t.ChosenResponser {
					if cr.ID == u.ID {
						crTag = true
						break
					}
				}

				if crTag {
					fmt.Println("你作为响应者的任务: ")
				} else {
					fmt.Println("你作为备选响应者的任务: ")
				}
			}
		}

		printTaskInfo(t)
	}
	fmt.Println()
}

func (u *User)CheckAcceptTask(a *comm.TaskAction) error {
	for _, ti := range u.TIs {
		if ti.ID == a.TaskID {
			return nil
		}
	}

	return errors.New("输入了错误的任务编号")
}


func (u *User)AcceptTask() {
	if u.ID == InvalidUserID {
		fmt.Println("用户未登录，无法接受任务。")
		return
	}

	var curTaskAction comm.TaskAction

	curTaskAction.Action = "ACCEPT"
	curTaskAction.UserID = u.ID
	var decision int32
	fmt.Printf("请输入要选择的任务编号 : ")
	fmt.Scanf("%d", &curTaskAction.TaskID)
	fmt.Printf("请输入你对任务的选择的编号 ：1-[接受] | 2-[拒绝] ：")
	fmt.Scanf("%d", &decision)
	if decision == 1 {
		curTaskAction.Decision = comm.TaskDecisionAccept
	} else {
		curTaskAction.Decision = comm.TaskDecisionRefuse
	}

	err := u.CheckAcceptTask(&curTaskAction)
	if err != nil {
		fmt.Printf("用户接受任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	fmt.Println("接受任务中...")
	data, err := json.Marshal(curTaskAction)
	if err != nil {
		fmt.Printf("用户接受任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest("GET", URLTasks, dataReader)
	if err != nil {
		fmt.Printf("用户接受任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	transport := http.Transport {
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: &transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("用户接受任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("用户接受任务时返回了错误的结果，请重试。错误信息：%s, 具体信息: %s.\n",
			resp.Status, resp.Header.Get("ErrorMsg"))
		return
	}

	fmt.Printf("接受任务成功，任务编号 : %d\n", curTaskAction.TaskID)

	return
}

func (u *User)CheckChooseResponser(a *comm.TaskAction) error {
	findTaskTag := false
	for _, ti := range u.TIs {
		if ti.ID == a.TaskID {
			findTaskTag = true
			findTag := false
			for _, cr := range a.ChosenResponserIDs {
				for _, r := range ti.Responsers {
					if cr == r.ID {
						findTag = true
						break
					}
				}
				if !findTag {
					return errors.New("你选择的响应者不在备选响应者中")
				} else {
					findTag = false
				}
			}
		}
	}

	if !findTaskTag {
		errors.New("输入了错误的任务编号")
	}

	return nil
}

func (u *User)ChooseResponser() {
	if u.ID == InvalidUserID {
		fmt.Println("用户未登录，无法选择任务响应者。")
		return
	}

	var curTaskAction comm.TaskAction

	curTaskAction.ChosenResponserIDs = make([]uint32, 0)

	curTaskAction.Action = "CHOOSE"
	fmt.Printf("请输入要选择响应者的任务编号 : ")
	fmt.Scanf("%d", &curTaskAction.TaskID)

	var tmpUID uint32 = 0
	for {
		fmt.Printf("请输入要选择的响应者编号(结束请输入0) : ")
		fmt.Scanf("%d", &tmpUID)
		if tmpUID == 0 {
			break
		} else {
			newResponserUID := tmpUID
			curTaskAction.ChosenResponserIDs = append(curTaskAction.ChosenResponserIDs, newResponserUID)
		}
	}

	err := u.CheckChooseResponser(&curTaskAction)
	if err != nil {
		fmt.Printf("选择响应者时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	fmt.Println("选择响应者中...")
	data, err := json.Marshal(curTaskAction)
	if err != nil {
		fmt.Printf("选择响应者时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest("GET", URLTasks, dataReader)
	if err != nil {
		fmt.Printf("选择响应者时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	transport := http.Transport {
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: &transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("选择响应者时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("选择响应者时返回了错误的结果，请重试。错误信息：%s, 具体信息: %s.\n",
			resp.Status, resp.Header.Get("ErrorMsg"))
		return
	}

	fmt.Printf("选择响应者成功，任务编号 : %d, 选择的响应者编号 ： %d\n", curTaskAction.TaskID, curTaskAction.UserID)

	return
}

func (u *User)CheckFulfilTask(a *comm.TaskAction) error {
	for _, ti := range u.TIs {
		if ti.ID == a.TaskID {
			for _, cr := range ti.ChosenResponser {
				if cr.ID == a.UserID {
					return nil
				}
			}
		}
	}

	return errors.New("输入了错误的任务编号,或者该任务你不是响应者,无法完成")
}

func (u *User)FulfilTask() {
	if u.ID == InvalidUserID {
		fmt.Println("用户未登录，无法完成任务。")
		return
	}

	var curTaskAction comm.TaskAction

	curTaskAction.Action = "FULFIL"
	curTaskAction.UserID = u.ID
	fmt.Printf("请输入已经完成的任务编号 : ")
	fmt.Scanf("%d", &curTaskAction.TaskID)

	err := u.CheckFulfilTask(&curTaskAction)
	if err != nil {
		fmt.Printf("通知系统任务已完成时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	fmt.Println("通知系统任务已完成...")
	data, err := json.Marshal(curTaskAction)
	if err != nil {
		fmt.Printf("通知系统任务已完成时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest("GET", URLTasks, dataReader)
	if err != nil {
		fmt.Printf("通知系统任务已完成时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	transport := http.Transport {
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("通知系统任务已完成时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("通知系统任务已完成时返回了错误的结果，请重试。错误信息：%s, 具体信息: %s.\n",
			resp.Status, resp.Header.Get("ErrorMsg"))
		return
	}

	fmt.Printf("通知系统任务已完成成功，任务编号 : %d\n", curTaskAction.TaskID)

	return
}

func (u *User)CheckEvaluateTask(a *comm.TaskAction) error {
	for _, ti := range u.TIs {
		if ti.ID == a.TaskID {
			if ti.Requester.ID == a.UserID {
				return nil
			}
		}
	}

	return errors.New("输入了错误的任务编号,或者该任务你不是请求者,无法结束")
}

func (u *User)EvaluateTask() {
	if u.ID == InvalidUserID {
		fmt.Println("用户未登录，无法评价并结束任务。")
		return
	}

	var curTaskAction comm.TaskAction

	curTaskAction.Action = "EVALUATE"
	curTaskAction.UserID = u.ID
	fmt.Printf("请输入评价并结束的任务编号 : ")
	fmt.Scanf("%d", &curTaskAction.TaskID)

	err := u.CheckEvaluateTask(&curTaskAction)
	if err != nil {
		fmt.Printf("评价并结束任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	fmt.Println("评价并结束任务中...")
	data, err := json.Marshal(curTaskAction)
	if err != nil {
		fmt.Printf("评价并结束任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest("DELETE", URLTasks, dataReader)
	if err != nil {
		fmt.Printf("评价并结束任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	transport := http.Transport {
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("评价并结束任务时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("评价并结束任务时返回了错误的结果，请重试。错误信息：%s, 具体信息: %s.\n",
			resp.Status, resp.Header.Get("ErrorMsg"))
		return
	}

	fmt.Printf("评价并结束任务成功，任务编号 : %d\n", curTaskAction.TaskID)

	return
}

func (u *User)QueryTaskAuto()  {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		if u.ID == InvalidUserID {
			continue
		}

		tis := u.QueryTaskGetTaskInfo()
		if tis == nil || len(tis) == 0 {
			continue
		}

		printCommandTag := false

		newTIs := make([]*comm.TaskInfoWithUsers, 0)

		found := false
		var foundTI *comm.TaskInfoWithUsers
		var foundTIIndex int
		for _, t := range tis {
			for index, ut := range u.TIs {
				if t.ID == ut.ID {
					foundTIIndex = index
					foundTI = ut
					found = true
					break
				}
			}

			if found {
				if foundTI.Status != t.Status {
					fmt.Println()
					if t.Requester.ID == u.ID {
						fmt.Printf("[请求者任务]任务状态更新：[%s] -> [%s], 任务信息: \n", comm.DescTaskStatus(foundTI.Status), comm.DescTaskStatus(t.Status))
					} else {
						fmt.Printf("[响应者任务]任务状态更新：[%s] -> [%s], 任务信息: \n", comm.DescTaskStatus(foundTI.Status), comm.DescTaskStatus(t.Status))
					}
					printTaskInfo(t)
					printCommandTag = true
				} else if len(foundTI.Responsers) != len(t.Responsers) {
					fmt.Println()
					if t.Requester.ID == u.ID {
						fmt.Println("[请求者任务]任务备选响应者更新，任务信息: ")
					} else {
						fmt.Println("[响应者任务]任务备选响应者更新，任务信息: ")
					}
					printTaskInfo(t)
					printCommandTag = true
				}

				u.TIs = append(u.TIs[:foundTIIndex], u.TIs[foundTIIndex+1:]...)
				u.TIs = append(u.TIs, t)
			} else {
				fmt.Println()
				if t.Requester.ID == u.ID {
					fmt.Println("[请求者任务]接收到新的任务信息: ")
				} else {
					fmt.Println("[响应者任务]接收到新的任务信息: ")
				}
				printTaskInfo(t)
				printCommandTag = true
				newTIs = append(newTIs, t)
			}

			foundTIIndex = 0
			foundTI = nil
			found = false
		}

		for _, ti := range newTIs {
			u.TIs = append(u.TIs, ti)
		}

		if printCommandTag {
			fmt.Print("Command : ")
		}
	}
}