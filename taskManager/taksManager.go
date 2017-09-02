package main

import (
	tmrpc "external/grpc/taskManagerRPC"
	"net"
	"fmt"
	"google.golang.org/grpc"
	"flag"
	"taskManager/server"
	"github.com/BurntSushi/toml"
	"os"
	"dataCenter/dataClient"
	"taskManager/rpc"
)

var tmsConfigPath string
var tmConfig TMConfig

type TMConfig struct {
	TMLogPath   string
	TMAddr      string
	DCAddr      string
	DCUser      string
	DCPassword  string
	TasksDCDatabase  string
	UsersDCDatabase  string
	SMAddr      string
	SMClientNum int
}

func init() {
	flag.StringVar(&tmsConfigPath, "conf", "./tmServer.toml", "Task Manager Server Config Path")
}

func initTMLog(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}

	server.SetServerLogger(file)
	rpc.SetRPCLogger(file)

	return nil
}

func main() {

	flag.Parse()

	_, err := toml.DecodeFile(tmsConfigPath, &tmConfig)

	fmt.Printf("Decode UM config : lgpath(%s), umaddr(%s), dcaddr(%s), dcuser(%s), dcpw(%s), tasksdcdb(%s), usersdcdb(%s).\n",
				tmConfig.TMLogPath, tmConfig.TMAddr, tmConfig.DCAddr, tmConfig.DCUser, tmConfig.DCPassword,
				tmConfig.TasksDCDatabase, tmConfig.UsersDCDatabase)

	err = initTMLog(tmConfig.TMLogPath)
	if err != nil {
		fmt.Println("Initialize Task Manager log failed.")
		return
	}

	tmServer := server.New(&server.TMServerConfig{
		Addr: tmConfig.TMAddr,
		TasksDataCenterConf: dataClient.DataCenterDesc{
			Addr: tmConfig.DCAddr,
			User: tmConfig.DCUser,
			Password: tmConfig.DCPassword,
			Database: tmConfig.TasksDCDatabase,
		},
		UsersDataCenterConf: dataClient.DataCenterDesc{
			Addr: tmConfig.DCAddr,
			User: tmConfig.DCUser,
			Password: tmConfig.DCPassword,
			Database: tmConfig.UsersDCDatabase,
		},
	})

	server.SMCG = rpc.NewSMClientGroup(tmConfig.SMAddr, tmConfig.SMClientNum)

	err = rpc.StartSMClientGroup(server.SMCG)
	if err != nil {
		fmt.Printf("Init Task Manager Server SMCG failed! Error : %s\n", err)
		return
	}

	lis, err := net.Listen("tcp", tmConfig.TMAddr)
	if err != nil {
		fmt.Println("UM Listen Failed.")
	}

	s := grpc.NewServer()

	tmrpc.RegisterTaskManagerServer(s, tmServer)
	s.Serve(lis)
}