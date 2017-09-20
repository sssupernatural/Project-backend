package main

import (
	umrpc "external/grpc/userManagerRPC"
	"net"
	"fmt"
	"google.golang.org/grpc"
	"flag"
	"userManager/server"
	"github.com/BurntSushi/toml"
	"os"
	"dataCenter/dataClient"
	"userManager/rpc"
)

var umsConfigPath string
var umConfig UMConfig

type UMConfig struct {
	UMLogPath   string
	UMAddr      string
	DCAddr      string
	DCUser      string
	DCPassword  string
	DCDatabase  string
	SMAddr      string
}

func init() {
	flag.StringVar(&umsConfigPath, "conf", "./umServer.toml", "User Manager Server Config Path")
}

func initUMLog(path string) error {
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

	_, err := toml.DecodeFile(umsConfigPath, &umConfig)

	fmt.Printf("Decode UM config : lgpath(%s), umaddr(%s), dcaddr(%s), dcuser(%s), dcpw(%s), dcdb(%s), smaddr(%s).\n",
				umConfig.UMLogPath, umConfig.UMAddr, umConfig.DCAddr, umConfig.DCUser, umConfig.DCPassword,
				umConfig.DCDatabase, umConfig.SMAddr)

	err = initUMLog(umConfig.UMLogPath)
	if err != nil {
		fmt.Println("Initialize User Manager log failed.")
		return
	}

	umServer := server.New(&server.UMServerConfig{
		Addr: umConfig.UMAddr,
		SMAddr: umConfig.SMAddr,
		DataCenterConf: dataClient.DataCenterDesc{
			Addr: umConfig.DCAddr,
			User: umConfig.DCUser,
			Password: umConfig.DCPassword,
			Database: umConfig.DCDatabase,
		},
	})

	go umServer.Init()

	lis, err := net.Listen("tcp", umConfig.UMAddr)
	if err != nil {
		fmt.Println("UM Listen Failed.")
	}

	s := grpc.NewServer()

	umrpc.RegisterUserManagerServer(s, umServer)

	s.Serve(lis)
}