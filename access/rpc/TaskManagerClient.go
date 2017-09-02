package rpc

import (
	"google.golang.org/grpc"
	tmrpc "external/grpc/taskManagerRPC"
	"errors"
)

type AccessTMClientGroup struct {
	TMAddr            string
	TMClientNum       int
	tmClientGroupChan chan tmrpc.TaskManagerClient
}

func (cg *AccessTMClientGroup)GetClient() tmrpc.TaskManagerClient {
	curClient := <-cg.tmClientGroupChan


	return curClient
}

func (cg *AccessTMClientGroup)ReturnClient(curClient tmrpc.TaskManagerClient) {
	cg.tmClientGroupChan <- curClient

	return
}

func NewTMClientGroup(tmAddr string, tmNumber int) *AccessTMClientGroup {
	return &AccessTMClientGroup{
		TMAddr: tmAddr,
		TMClientNum: tmNumber,
		tmClientGroupChan: make(chan tmrpc.TaskManagerClient, tmNumber),
	}
}

func StartAccessTMClientGroup(cg *AccessTMClientGroup) error {
	err := initTaskManagerClientGroup(cg)
	if err != nil {
		return err
	}

	return nil
}

func initTaskManagerClient(cg *AccessTMClientGroup, errChan chan uint8) {
	conn, err := grpc.Dial(cg.TMAddr, grpc.WithInsecure())
	if err != nil {
		logger.Errorf("Initialize task manager client failed, err : %s.", err)
		errChan <- ERROR
		return
	}

	logger.Infof("client dial succeed,", conn)

	client := tmrpc.NewTaskManagerClient(conn)

	cg.tmClientGroupChan <- client
	errChan<- SUCCESS

	return
}

func initTaskManagerClientGroup(cg *AccessTMClientGroup) error {

	errChan := make(chan uint8, cg.TMClientNum)

	for i := 0; i < cg.TMClientNum; i++{
		go initTaskManagerClient(cg, errChan)
	}

	totalFailed := 0

	for i := 0; i < cg.TMClientNum; i++ {
		errCode := <- errChan
		if errCode != SUCCESS {
			totalFailed++
		}
	}

	if totalFailed != 0 {
		logger.Errorf("Init task manager group failed, total failed client number is %d.", totalFailed)
		return errors.New("Init task Manager Client Group Failed.")
	}

	return nil
}