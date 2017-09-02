package server

import (
	"fmt"
	"net/http"
	"access/rpc"
)

var UMCG *rpc.AccessUMClientGroup
var TMCG *rpc.AccessTMClientGroup

type AccessServer struct {
	Addr        string
	UMAddr      string
	UMClientNum int
	TMAddr      string
	TMClientNum int
}

func New() (s *AccessServer) {
	s = &AccessServer{}
	return
}

func (s *AccessServer)Serve() error {

	UMCG = rpc.NewUMClientGroup(s.UMAddr, s.UMClientNum)

	err := rpc.StartAccessUMClientGroup(UMCG)
	if err != nil {
		fmt.Printf("Init Access Server UMCG failed! Error : %s\n", err)
		return err
	}

	TMCG = rpc.NewTMClientGroup(s.TMAddr, s.TMClientNum)

	err = rpc.StartAccessTMClientGroup(TMCG)
	if err != nil {
		fmt.Printf("Init Access Server TMCG failed! Error : %s\n", err)
		return err
	}

	http.HandleFunc("/users", usersHandler)
	http.HandleFunc("/userinfo", userInfoHandler)
	http.HandleFunc("/tasks", tasksHandler)

	err = http.ListenAndServe(s.Addr, nil)
	if err != nil {
		fmt.Printf("Access Server Listen Failed! Error : %s\n", err)

		return err
	} else {
		fmt.Println("Start Serve!")
	}

	return nil
}

