package command

import (
	"fmt"
	"net/http"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"external/comm"
	"strconv"
)

const (
	URLUsers    string = "http://localhost:8080/users"
    URLUserInfo string = "http://localhost:8080/userinfo"
)

const (
	UserStateOnline  uint8 = 0
	UserStateOffline uint8 = 1
)

const (
	UserSexMale   uint8 = 0
	UserSexFeMale uint8 = 0
)

type User struct {
	ID          uint32
	Name        string
	PhoneNumber string
	info        *comm.UserInfo
	TIs         []*comm.TaskInfoWithUsers
}

var ClientIDToAbiMap map[int]string
var ClientAbiToIDMap map[string]int

func (u *User)Clear()  {
	u.ID = InvalidUserID
	u.Name = ""
	u.PhoneNumber = ""
	u.info = nil
}

const InvalidUserID uint32 = 0

func RegisterUser(u *User) {
	if u.ID != InvalidUserID {
		fmt.Printf("用户已经登录，无法注册，当前用户名为[%s],手机号为[%s]。\n", u.Name, u.PhoneNumber)
		return
	}

	var curUserAction comm.UserAction

	curUserAction.CheckInfo = &comm.UserCheckInfo{}

	fmt.Print("请输入注册用户名：")
	fmt.Scanf("%s", &curUserAction.CheckInfo.Name)
	fmt.Print("请输入注册手机号：")
	fmt.Scanf("%s", &curUserAction.CheckInfo.PhoneNumber)
	for i := 0; len(curUserAction.CheckInfo.Password) < 6 ; i++ {
		if i != 0 {
			fmt.Println("密码至少为6位(数字或者字母)，请重新设置")
		}
		fmt.Print("请输入该用户密码：")
		fmt.Scanf("%s", &curUserAction.CheckInfo.Password)
	}
	fmt.Println("注册中...")

	curUserAction.Action = "REGISTER"

	data, err := json.Marshal(curUserAction)
	if err != nil {
		fmt.Printf("注册用户时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest("POST", URLUsers, dataReader)
	if err != nil {
		fmt.Printf("注册用户时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("注册用户时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("注册用户时返回了错误的结果，请重试。错误信息：%s.\n", resp.Status)
		return
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("注册用户时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	var userInfo comm.UserInfo
	err = json.Unmarshal(respBody, &userInfo)
	if err != nil {
		fmt.Printf("注册用户时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	fmt.Println("注册成功，当前用户信息：")
	printUserInfo(&userInfo)

	u.ID = userInfo.ID
	u.Name = userInfo.Name
	u.PhoneNumber = userInfo.PhoneNumber
	u.info = &userInfo

	return
}

func LoginUser(u *User) {
	if u.ID != InvalidUserID {
		fmt.Printf("用户已经登录，当前用户名为[%s],手机号为[%s]。\n", u.Name, u.PhoneNumber)
		return
	}

	var curUserAction comm.UserAction

	curUserAction.CheckInfo = &comm.UserCheckInfo{}

	fmt.Print("请输入手机号：")
	fmt.Scanf("%s", &curUserAction.CheckInfo.PhoneNumber)
	fmt.Print("请输入密码：")
	fmt.Scanf("%s", &curUserAction.CheckInfo.Password)
	fmt.Println("登录中...")

	curUserAction.Action = "LOGIN"

	data, err := json.Marshal(curUserAction)
	if err != nil {
		fmt.Printf("用户登录时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest("POST", URLUsers, dataReader)
	if err != nil {
		fmt.Printf("用户登录时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("用户登录时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("用户登录时返回了错误的结果，请重试。错误信息：%s, 具体信息: %s.\n",
			resp.Status, resp.Header.Get("ErrorMsg"))
		return
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("用户登录时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	var userInfo comm.UserInfo
	err = json.Unmarshal(respBody, &userInfo)
	if err != nil {
		fmt.Printf("用户登录时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	fmt.Println("登录成功, 当前用户信息：")
	printUserInfo(&userInfo)

	u.ID = userInfo.ID
	u.Name = userInfo.Name
	u.PhoneNumber = userInfo.PhoneNumber
	u.info = &userInfo

	return
}

func (u *User)Logout() {
	if u.ID == InvalidUserID {
		fmt.Println("用户未登录，无法退出。")
		return
	}

	var curUserAction comm.UserAction

	curUserAction.CheckInfo = &comm.UserCheckInfo{}

	curUserAction.Action = "LOGOUT"
	curUserAction.CheckInfo.PhoneNumber = u.PhoneNumber

	data, err := json.Marshal(curUserAction)
	if err != nil {
		fmt.Printf("用户退出时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest("POST", URLUsers, dataReader)
	if err != nil {
		fmt.Printf("用户退出时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("用户退出时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("用户退出时返回了错误的结果，请重试。错误信息：%s, 具体信息: %s.\n",
			resp.Status, resp.Header.Get("ErrorMsg"))
		return
	}

	fmt.Printf("用户号 :%d\n", u.ID)
	fmt.Printf("用户名 :%s\n", u.Name)
	fmt.Printf("手机号 :%s\n", u.PhoneNumber)
	fmt.Println("用户退出成功！")

	u.Clear()

	return

}

func getAbiLevel(exp int32) int32 {
	return exp/5
}

func printUserInfo(u *comm.UserInfo) {
	fmt.Println("| -------------------------UserDesc-------------------------")
	fmt.Printf("| | 编号 : %d\n", u.ID)
	fmt.Printf("| | 手机号 : %s\n", u.PhoneNumber)
	fmt.Printf("| | 用户名 : %s\n", u.Name)
	fmt.Printf("| | 状态 : %s\n", comm.DescStatus(u.Status))
	fmt.Printf("| | 性别 : %s\n", comm.DescSex(u.Sex))
	fmt.Printf("| | 年龄 : %d\n", u.Age)
	if u.CurLocation != nil && (u.CurLocation.Latitude != 0 || u.CurLocation.Longitude != 0) {
		fmt.Printf("| | 当前位置 : [%f,%f]\n", u.CurLocation.Longitude, u.CurLocation.Latitude)
	}
	if len(u.Locations) != 0 {
		fmt.Print("| | 位置 : ")
		for _, pos := range u.Locations {
			fmt.Printf("[%f,%f] | ", pos.Longitude, pos.Latitude)
		}
		fmt.Println()
	}

	if u.Abilities != nil {
		fmt.Print("| | 能力 : ")
		for index, abi := range u.Abilities.ABIs {
			if index != 0 {
				fmt.Printf("[%s(%d级)] | ", abi.ABI, getAbiLevel(abi.Experience))
			}
		}
		fmt.Println()
	}
	fmt.Println("| -------------------------UserDesc-------------------------")

	return
}

func (u *User)printUserAbilities()  {
	fmt.Println("当前用户的能力列表：")
	for _, a := range u.info.Abilities.ABIs {
		fmt.Printf("%d-[%s(%d级)] | ", ClientAbiToIDMap[a.ABI], a.ABI, getAbiLevel(a.Experience))
	}
	fmt.Println()
}

func (u *User)printUserLocations()  {
	fmt.Println("当前用户的位置列表：")
	for i, l := range u.info.Locations {
		fmt.Printf("%d-[%v] | ", i, l)
	}
	fmt.Println()
}

func printSysAbisEntertainment() {
	fmt.Println("--------------------------------------------------------------------------------------------------------")
	fmt.Println()
	fmt.Println("                                                  [10000:娱乐]")
	fmt.Println("                                                        ┃")
	fmt.Println("                 ┏━━━━━━━━━━━━┳━━━━━━━━━━━━┳━━━━━━━━━━━━┻━━━┳━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━┓")
	fmt.Println("           [11000:唱歌]  [12000:桌游]  [13000:密室逃脱]  [14000:真人CS]  [15000:电脑游戏]  [16000:掌上游戏]")
	fmt.Println("                              ┃                                            ┃")
	fmt.Println("      ┏━━━━━━━━━━━━━━┳━━━━━━━━┻━━━━━━┳━━━━━━━━━━━━━┓               ┏━━━━━━━┻━━━━━━┓ ")
	fmt.Println("[12100:狼人杀]   [12200:三国杀]   [12300:麻将]   [12400:扑克]    [15100:网络游戏] [15200:单机游戏]")
	fmt.Println("                                           ┏━━━━━━━━━━━━━━━━━━━━━┳━┻━━━━━━━━━━━━┳━━━━━━━━━━━━━━┓       ")
	fmt.Println("                                     [15110:绝地求生大逃杀]   [15120:DOTA]   [15130:CSGO]   [15140:H1Z1] ")
	fmt.Println()
	fmt.Println("---------------------------------------------------------------------------------------------------------")
}

var sysAbisEntertainment = []sysAbi{{10000, "娱乐"}, {11000, "唱歌"}, {12000, "桌游"}, {13000, "密室逃脱"},
	                                {14000, "真人CS"}, {15000, "电脑游戏"}, {16000, "掌上游戏"}, {12100, "狼人杀"},
	                                {12200, "三国杀"},{12300, "麻将"},{12400, "扑克"},{15100, "网络游戏"},
	                                {15200, "单机游戏"},{15110, "绝地求生大逃杀"},{15120, "DOTA"},{15130, "CSGO"},
	                                {15140, "H1Z1"},
}

func printSysAbisLiving() {
	fmt.Println("---------------------------------------------------------------------------------------")
	fmt.Println()
	fmt.Println("                                            [20000:生活]")
	fmt.Println("                                                  ┃")
	fmt.Println("             ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━━┳━━━━━━━━━━━━┳━━━━━━━━━━━━┳━━━━━━━━━━━┓")
	fmt.Println("       [21000:美食]                               [22000:交通]  [23000:住宿]  [24000:衣装] [25000:宠物]")
	fmt.Println("             ┃                                         ┃")
	fmt.Println("      ┏━━━━━━┻━━━━━┓            ┏━━━━━━━━━━━━━━┳━━━━━━━┻━━━━━━┳━━━━━━━━━━━━┓")
	fmt.Println("[21100:火锅]  [21200:炒菜]  [22100:短途车]  [22200:长途车]  [22300:火车]  [22400:飞机]")
	fmt.Println()
	fmt.Println("---------------------------------------------------------------------------------------")
}

var sysAbisLiving = []sysAbi{{20000, "生活"}, {21000, "美食"}, {22000, "交通"}, {23000, "住宿"},
	{24000, "衣装"}, {25000, "宠物"}, {21100, "火锅"}, {21200, "炒菜"},
	{22100, "短途车"},{22200, "长途车"},{22300, "火车"},{22400, "飞机"},
}

func printSysAbiWork() {
	fmt.Println("---------------------------------------------------------------------------------------------------------")
	fmt.Println()
	fmt.Println("                                                         [30000:职业]")
	fmt.Println("                                                               ┃")
	fmt.Println("                                ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓")
	fmt.Println("                              [31000:IT]                                              [32000:医药]")
	fmt.Println("                                    ┃                                                       ┃")
	fmt.Println("      ┏━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━┳━━━━━━━━━━━━━┓             ┏━━━━━━━━━━━━╋━━━━━━━━━━━━┓")
	fmt.Println("[31100:通讯IT]  [31200:游戏IT]  [31300:社交IT]  [31400:前端]  [31500:后端]  [32100:医药销售]  [32200:医药管理]  [32300:医药生产]")
	fmt.Println()
	fmt.Println("---------------------------------------------------------------------------------------------------------")
}

var sysAbisWork = []sysAbi{{30000, "职业"}, {31000, "IT"}, {32000, "医药"}, {31100, "通讯IT"},
	{31200, "游戏IT"}, {31300, "社交IT"}, {31400, "前端"}, {31500, "后端"},
	{32100, "医药销售"},{32200, "医药管理"},{32300, "医药生产"},
}

func printSysAbiTravel()  {
	fmt.Println("-------------------------------")
	fmt.Println()
	fmt.Println("        [40000:旅游]")
	fmt.Println("              ┃")
	fmt.Println("      ┏━━━━━━━┻━━━━━━┓")
	fmt.Println("[41000:国内旅游] [42000:国外旅游]")
	fmt.Println()
	fmt.Println("-------------------------------")
}

var sysAbisTravel = []sysAbi{{40000, "旅游"}, {41000, "国内旅游"}, {42000, "国外旅游"},
}

func printSysAbiHelp() {
	fmt.Println("-------------------------------")
	fmt.Println()
	fmt.Println("        [50000:帮忙]")
	fmt.Println("              ┃")
	fmt.Println("      ┏━━━━━━━┻━━━━━━┓")
	fmt.Println("[51000:帮小忙] [52000:帮大忙]")
	fmt.Println()
	fmt.Println("-------------------------------")
}

var sysAbisHelp = []sysAbi{{50000, "帮忙"}, {51000, "帮小忙"}, {52000, "帮大忙"},
}

func printSysAbisSports() {
	fmt.Println("-------------------------------------------------------------------")
	fmt.Println()
	fmt.Println("                     [60000:运动]")
	fmt.Println("                           ┃")
	fmt.Println("            ┏━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━┓")
	fmt.Println("      [61000:球类运动]                     [62000:器械类运动]")
	fmt.Println("            ┃                                  ┃   ")
	fmt.Println("      ┏━━━━━┻━━━━━━┓            ┏━━━━━━━━━━━━━━╋━━━━━━━━━━━━┓")
	fmt.Println("[61100:足球]  [61200:篮球]  [62100:自行车]  [62200:健身]  [62300:滑板] ")
	fmt.Println()
	fmt.Println("--------------------------------------------------------------------")
}

var sysAbisSports = []sysAbi{{60000, "运动"}, {61000, "球类运动"}, {62000, "器械类运动"}, {61100, "足球"},
	{61200, "篮球"}, {62100, "自行车"}, {62200, "健身"}, {62300, "滑板"},
}

func printSysAbiAll()  {
	printSysAbisEntertainment()
	printSysAbisLiving()
	printSysAbiWork()
	printSysAbiTravel()
	printSysAbiHelp()
	printSysAbisSports()
}

type sysAbi struct {
	id  int
	abi string
}

func GenerateSysIDToAbiMap() map[int]string {
	sysAbis := make([]sysAbi, 0)
	sysAbis = append(sysAbis, sysAbisEntertainment...)
	sysAbis = append(sysAbis, sysAbisLiving...)
	sysAbis = append(sysAbis, sysAbisWork...)
	sysAbis = append(sysAbis, sysAbisTravel...)
	sysAbis = append(sysAbis, sysAbisHelp...)
	sysAbis = append(sysAbis, sysAbisSports...)

	IDToAbiMap := make(map[int]string)

	for _, abi := range sysAbis {
		IDToAbiMap[abi.id] = abi.abi
	}

	return IDToAbiMap
}

func GenerateSysAbiToIDMap() map[string]int {
	sysAbis := make([]sysAbi, 0)
	sysAbis = append(sysAbis, sysAbisEntertainment...)
	sysAbis = append(sysAbis, sysAbisLiving...)
	sysAbis = append(sysAbis, sysAbisWork...)
	sysAbis = append(sysAbis, sysAbisTravel...)
	sysAbis = append(sysAbis, sysAbisHelp...)
	sysAbis = append(sysAbis, sysAbisSports...)

	AbiToIDMap := make(map[string]int)

	for _, abi := range sysAbis {
		AbiToIDMap[abi.abi] = abi.id
	}

	return AbiToIDMap
}

func printUserAbis(abiIDs []int) {
	fmt.Println("当前用户的能力列表：")
	for _, id := range abiIDs {
		fmt.Printf("%d-[%s] | ", id, ClientIDToAbiMap[id])
	}
	fmt.Println()
}

func (u *User)EditAbilities() {
	if u.ID == InvalidUserID {
		fmt.Println("用户未登录，无法修改用户信息。")
		return
	}

	var cmd string
	var anum string
	var abiID string

	fmt.Println("系统能力图如下：")
	printSysAbiAll()

	curAbiIDs := make([]int, 0)
	for _, abi := range u.info.Abilities.ABIs {
		curAbiIDs = append(curAbiIDs, ClientAbiToIDMap[abi.ABI])
	}

	u.printUserAbilities()

	for {
	    fmt.Print("删除能力请输入d，添加能力请输入a，完成请输入f：")
		fmt.Scanf("%s", &cmd)
		if cmd == "f" {
			break
		}

		if cmd == "d" {
			for {
				fmt.Print("请输入要删除的一个能力的编号(结束请输入f):")
				fmt.Scanf("%s", &anum)
				if anum == "f" {
					break
				} else {
					id, _ := strconv.Atoi(anum)
					if _, ok := ClientIDToAbiMap[id]; ok {
						for index, curID := range curAbiIDs {
							if curID == id {
								curAbiIDs = append(curAbiIDs[:index], curAbiIDs[index+1:]...)
							}
						}
					}
					printUserAbis(curAbiIDs)
				}
			}
		}

		if cmd == "a" {
			for {
				fmt.Print("请输入要添加的一个能力的编号(结束请输入f):")
				fmt.Scanf("%s", &abiID)
				if abiID == "f" {
					break
				} else {
					id, _ := strconv.Atoi(abiID)
					if _, ok := ClientIDToAbiMap[id]; ok {
						curAbiIDs = append(curAbiIDs, id)
					}
					printUserAbis(curAbiIDs)
				}
			}
		}
    }

	u.info.Abilities = u.generateAbiHeapByAbiIndexes(curAbiIDs)

	fmt.Println("正在修改用户能力...")

	data, err := json.Marshal(u.info)
	if err != nil {
		fmt.Printf("修改用户能力时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest("POST", URLUserInfo, dataReader)
	if err != nil {
		fmt.Printf("修改用户能力时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("修改用户能力时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("修改用户能力时返回了错误的结果，请重试。错误信息：%s, 具体信息: %s.\n",
			resp.Status, resp.Header.Get("ErrorMsg"))
		return
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("修改用户能力时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	var userInfo comm.UserInfo
	err = json.Unmarshal(respBody, &userInfo)
	if err != nil {
		fmt.Printf("修改用户能力时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	fmt.Println("修改用户能力成功，当前用户信息：")
	printUserInfo(&userInfo)

	u.info = &userInfo

	return
}

func (u *User)EditLocations() {
	if u.ID == InvalidUserID {
		fmt.Println("用户未登录，无法修改位置信息。")
		return
	}

	var cmd string
	var lnum string
	var loc comm.Location

	for {
		u.printUserLocations()
		fmt.Print("删除位置请输入d，添加位置请输入a，完成请输入f：")
		fmt.Scanf("%s", &cmd)
		if cmd == "f" {
			break
		}

		if cmd == "d" {
			for {
				fmt.Print("请输入要删除的一个位置的编号(结束请输入f):")
				fmt.Scanf("%s", &lnum)
				if lnum == "f" {
					break
				} else {
					index, _ := strconv.Atoi(lnum)
					if index >= 0 && index < len(u.info.Locations) {
						u.info.Locations = append(u.info.Locations[:index], u.info.Locations[index+1:]...)
					}
					u.printUserLocations()
				}
			}
		}

		if cmd == "a" {
			for {
				fmt.Print("请输入要添加的一个位置:")
				fmt.Scanf("%f,%f", &loc.Longitude, &loc.Latitude)
				if loc.Longitude == 0 && loc.Latitude == 0 {
					break
				} else {
					u.info.Locations = append(u.info.Locations, &loc)
					u.printUserLocations()
				}
			}
		}
	}

	fmt.Println("正在修改用户位置...")

	data, err := json.Marshal(u.info)
	if err != nil {
		fmt.Printf("修改用户位置时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest("POST", URLUserInfo, dataReader)
	if err != nil {
		fmt.Printf("修改用户位置时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("修改用户位置时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("修改用户位置时返回了错误的结果，请重试。错误信息：%s, 具体信息: %s.\n",
			resp.Status, resp.Header.Get("ErrorMsg"))
		return
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("修改用户位置时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	var userInfo comm.UserInfo
	err = json.Unmarshal(respBody, &userInfo)
	if err != nil {
		fmt.Printf("修改用户位置时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	fmt.Println("修改用户位置成功，当前用户信息：")
	printUserInfo(&userInfo)

	u.info = &userInfo

	return
}

func (u *User)generateAbiHeapByAbiIndexes(abiIDs []int) *comm.AbisHeap {
	out := &comm.AbisHeap{
		ABIs: make([]*comm.AbiNode, 0),
	}
	out.ABIs = append(out.ABIs, &comm.AbiNode{})
	outIndex := 1

	abiIndexMap := make(map[string]int)

	divideNumbers := []int{10000, 1000, 100, 10, 1}

	for _, divideNum := range divideNumbers {
		for _, id := range abiIDs {
			curID := (id/divideNum)*divideNum
			curAbi := ClientIDToAbiMap[curID]
			_, ok := abiIndexMap[curAbi]
			if !ok {
				parentIndex := 0
				if divideNum == 10000 {
					parentIndex = 0
				} else {
					parentAbi := ClientIDToAbiMap[(id/(divideNum*10))*(divideNum*10)]
					parentIndex = abiIndexMap[parentAbi]
				}
				var exp int32 = 0
				if u.info.Abilities != nil {
					for _, uAbi := range u.info.Abilities.ABIs {
						if curAbi == uAbi.ABI {
							exp = uAbi.Experience
						}
					}
				}
				out.ABIs = append(out.ABIs, &comm.AbiNode{
					ABI: curAbi,
					Experience: exp,
					ParentIndex: int32(parentIndex),
				})
				abiIndexMap[curAbi] = outIndex
				outIndex++
			}
		}
	}

	return out
}

func (u *User)AddInfo() {
	if u.ID == InvalidUserID {
		fmt.Println("用户未登录，无法编辑用户能力。")
		return
	}

	newUserInfo := &comm.UserInfo{
		ID: u.ID,
		PhoneNumber: u.PhoneNumber,
		Name: u.Name,
		Status: comm.UserStatusOnline,
		CurLocation: &comm.Location{},
		Locations: make([]*comm.Location, 0),
	}

	var sex string
	fmt.Print("性别 : ")
	fmt.Scanf("%s", &sex)
	if sex == "男" {
		newUserInfo.Sex = comm.UserSexMale
	} else if sex == "女" {
		newUserInfo.Sex = comm.UserSexFemale
	} else {
		newUserInfo.Sex = comm.UserSexUnknown
	}

	fmt.Print("年龄 : ")
	fmt.Scanf("%d", &newUserInfo.Age)

	tmpAbisIDs := make([]int, 0)
	fmt.Println("请添加能力项，能力图如下")
	printSysAbiAll()
	for {
		var ability string
		fmt.Print("输入要添加的能力项编号(添加完成请输入字母'f') : ")
		fmt.Scanf("%s", &ability)
		if ability == "f" {
			break
		} else {
			id, _ := strconv.Atoi(ability)
			if _, ok := ClientIDToAbiMap[id]; ok {
				tmpAbisIDs = append(tmpAbisIDs, id)
			}
		}
	}

	if u.info.Abilities != nil {
		for _, curAbi := range u.info.Abilities.ABIs {
			tmpAbisIDs = append(tmpAbisIDs, ClientAbiToIDMap[curAbi.ABI])
		}
	}

	newUserInfo.Abilities = u.generateAbiHeapByAbiIndexes(tmpAbisIDs)

	fmt.Print("设置当前位置坐标 : ")
	fmt.Scanf("%f,%f", &newUserInfo.CurLocation.Longitude, &newUserInfo.CurLocation.Latitude)

	var pos *comm.Location
	for {
		pos = &comm.Location{}

		fmt.Print("添加常用位置坐标，添加完成请输入坐标(0,0): ")
		fmt.Scanf("%f,%f", &pos.Longitude, &pos.Latitude)
		if pos.Longitude == 0 && pos.Latitude == 0 {
			break
		} else {
			newUserInfo.Locations = append(newUserInfo.Locations, pos)
		}
	}

	fmt.Println("正在修改用户信息...")

	data, err := json.Marshal(newUserInfo)
	if err != nil {
		fmt.Printf("修改用户信息时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	dataReader := bytes.NewReader(data)

	req, err := http.NewRequest("POST", URLUserInfo, dataReader)
	if err != nil {
		fmt.Printf("修改用户信息时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("修改用户信息时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("修改用户信息时返回了错误的结果，请重试。错误信息：%s, 具体信息: %s.\n",
			resp.Status, resp.Header.Get("ErrorMsg"))
		return
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("修改用户信息时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	var userInfo comm.UserInfo
	err = json.Unmarshal(respBody, &userInfo)
	if err != nil {
		fmt.Printf("修改用户信息时出现异常错误，请重试。错误信息: %s.\n", err)
		return
	}

	fmt.Println("修改用户信息成功，当前用户信息：")
	printUserInfo(&userInfo)

	u.info = &userInfo

	return
}
