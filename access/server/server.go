package server

import (
	"fmt"
	"net/http"
	"access/rpc"
	umrpc "external/grpc/userManagerRPC"
	tmrpc "external/grpc/taskManagerRPC"
)

type AccessServer struct {
	umClient umrpc.UserManagerClient
	tmClient tmrpc.TaskManagerClient

	Addr        string
	UMAddr      string
	TMAddr      string
}

func New() (s *AccessServer) {
	s = &AccessServer{}
	return
}

func (s *AccessServer)Serve() error {
	var err error
	s.umClient, err = rpc.InitUserManagerClient(s.UMAddr)
	if err != nil {
		return err
	}

	s.tmClient, err = rpc.InitTaskManagerClient(s.TMAddr)
	if err != nil {
		return err
	}

	http.HandleFunc("/users", s.usersHandler)
	http.HandleFunc("/userinfo", s.userInfoHandler)
	http.HandleFunc("/tasks", s.tasksHandler)

	err = http.ListenAndServe(s.Addr, nil)
	if err != nil {
		fmt.Printf("Access Server Listen Failed! Error : %s\n", err)

		return err
	} else {
		fmt.Println("Start Serve!")
	}

	return nil
}

