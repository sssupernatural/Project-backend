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
	fmt.Print("请输入该用户密码：")
	fmt.Scanf("%s", &curUserAction.CheckInfo.Password)
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

	if len(u.Abilities.ABIs) != 0 {
		fmt.Print("| | 能力 : ")
		for _, abi := range u.Abilities.ABIs {
			fmt.Printf("[%s(%d级)] | ", abi.ABI, getAbiLevel(abi.Experience))
		}
		fmt.Println()
	}
	fmt.Println("| -------------------------UserDesc-------------------------")

	return
}

func (u *User)printUserAbilities()  {
	fmt.Println("当前用户的能力列表：")
	for i, a := range u.info.Abilities.ABIs {
		fmt.Printf("%d-[%s(%d级)] | ", i, a.ABI, getAbiLevel(a.Experience))
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

func printSysAbisSports() {
	fmt.Println("-------------------------------------------------------------------")
	fmt.Println()
	fmt.Println("                     [00000:运动]")
	fmt.Println("                           ┃")
	fmt.Println("            ┏━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━┓")
	fmt.Println("      [01000:球类运动]                     [02000:器械类运动]")
	fmt.Println("            ┃                                  ┃   ")
	fmt.Println("      ┏━━━━━┻━━━━━━┓            ┏━━━━━━━━━━━━━━╋━━━━━━━━━━━━┓")
	fmt.Println("[01100:足球]  [01200:篮球]  [02100:自行车]  [02200:健身]  [02300:滑板] ")
	fmt.Println()
	fmt.Println("--------------------------------------------------------------------")
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

func printSysAbisLiving() {
	fmt.Println("---------------------------------------------------------------------------------------")
	fmt.Println()
	fmt.Println("                                            [20000:生活]")
	fmt.Println("                                                  ┃")
	fmt.Println("             ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━━┳━━━━━━━━━━━━┳━━━━━━━━━━━━┳━━━━━━━━━━━┓")
	fmt.Println("       [21000:美食]                               [22000:交通]  [23000:住宿]  [24000:衣装] [25000:宠物]")
	fmt.Println("             ┃                                         ┃")
	fmt.Println("      ┏━━━━━━┻━━━━━┓            ┏━━━━━━━━━━━━━━┳━━━━━━━┻━━━━━━┳━━━━━━━━━━━━┓")
	fmt.Println("[21100:火锅]  [21200:炒菜]  [21300:短途车]  [21400:长途车]  [21500:火车]  [21600:飞机]")
	fmt.Println()
	fmt.Println("---------------------------------------------------------------------------------------")
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
	fmt.Println("[31100:通讯IT]  [31200:游戏IT]  [31300:社交IT]  [31400:前端]  [31500:后端]  [32100:销售]  [32200:管理]  [32300:生产]")
	fmt.Println()
	fmt.Println("---------------------------------------------------------------------------------------------------------")
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

func printSysAbiAll()  {
	printSysAbisSports()
	printSysAbisEntertainment()
	printSysAbisLiving()
	printSysAbiWork()
	printSysAbiTravel()
	printSysAbiHelp()
}


func (u *User)EditAbilities() {
	if u.ID == InvalidUserID {
		fmt.Println("用户未登录，无法修改用户信息。")
		return
	}

	var cmd string
	var anum string
	var abi string

	fmt.Println("能力图如下：")
	printSysAbiAll()

	for {
		u.printUserAbilities()
	    fmt.Print("删除能力请输入d，添加能力请输入a，完成请输入f：")
		fmt.Scanf("%s", &cmd)
		if cmd == "f" {
			break
		}

		if cmd == "d" {
			for {
				fmt.Print("请输入要删除的一个能力的编号(删除结束请输入f):")
				fmt.Scanf("%s", &anum)
				if anum == "f" {
					break
				} else {
					index, _ := strconv.Atoi(anum)
					if index >= 0 && index < len(u.info.Abilities) {
						u.info.Abilities = append(u.info.Abilities[:index], u.info.Abilities[index+1:]...)
					}
					u.printUserAbilities()
				}
			}
		}

		if cmd == "a" {
			for {
				fmt.Print("请输入要添加的一个能力(结束请输入f):")
				fmt.Scanf("%s", &abi)
				if abi == "f" {
					break
				} else {
					u.info.Abilities = append(u.info.Abilities, abi)
					u.printUserAbilities()
				}
			}
		}
    }

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

	tmpAbisIndexes := make([]int, 0)
	fmt.Println("请添加能力项，能力图如下")
	for {
		var ability string
		fmt.Print("输入要添加的能力项编号(添加完成请输入字母'f') : ")
		fmt.Scanf("%s", &ability)
		if ability == "f" {
			break
		} else {
			index, _ := strconv.Atoi(ability)
			tmpAbisIndexes = append(tmpAbisIndexes, index)
		}
	}

	newUserInfo.Abilities = generateAbiHeapByAbiIndexes(tmpAbisIndexes)

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
