package main

import (
	smrpc "external/grpc/searchManagerRPC"
	"net"
	"fmt"
	"google.golang.org/grpc"
	"flag"
	"searchManager/server"
	"github.com/BurntSushi/toml"
	"os"
	"dataCenter/dataClient"
	"searchManager/server/core"
	"searchManager/server/engine"
	"searchManager/server/abiTree"
	"searchManager/server/types"
)

var smsConfigPath string
var smConfig SMConfig

type SMConfig struct {
	SMLogPath  string
	SMAddr     string
	DCAddr     string
	DCUser     string
	DCPassword string
	DCDatabase string
}

func init() {
	flag.StringVar(&smsConfigPath, "conf", "./smServer.toml", "Search Manager Server Config Path")
}

func initSMLog(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}

	server.SetServerLogger(file)
	core.SetCoreLogger(file)
	engine.SetEngineLogger(file)
	abiTree.SetAbiTreeLogger(file)
	types.SetTypesLogger(file)
	return nil
}

func main() {

	fmt.Println("4444")
	flag.Parse()

	_, err := toml.DecodeFile(smsConfigPath, &smConfig)

	fmt.Printf("Decode SM config : lgpath(%s), smaddr(%s), dcaddr(%s), dcuser(%s), dcpw(%s), dcdb(%s).\n",
				smConfig.SMLogPath, smConfig.SMAddr, smConfig.DCAddr, smConfig.DCUser, smConfig.DCPassword,
				smConfig.DCDatabase)

	err = initSMLog(smConfig.SMLogPath)
	if err != nil {
		fmt.Println("Initialize Search Manager log failed.")
		return
	}

	smServer := server.New(&server.SMServerConfig{
		Addr: smConfig.SMAddr,
		DataCenterConf: dataClient.DataCenterDesc{
			Addr: smConfig.DCAddr,
			User: smConfig.DCUser,
			Password: smConfig.DCPassword,
			Database: smConfig.DCDatabase,
		},
	})

	lis, err := net.Listen("tcp", smConfig.SMAddr)
	if err != nil {
		fmt.Println("UM Listen Failed.")
	}

	s := grpc.NewServer()

	smrpc.RegisterSearchManagerServer(s, smServer)
	s.Serve(lis)
}