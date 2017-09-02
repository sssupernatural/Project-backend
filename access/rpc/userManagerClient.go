package rpc

import (
	"google.golang.org/grpc"
	umrpc "external/grpc/userManagerRPC"
	"errors"
)

const (
	ERROR   uint8 = 0
	SUCCESS uint8 = 1
)

type AccessUMClientGroup struct {
	UMAddr            string
	UMClientNum       int
	umClientGroupChan chan umrpc.UserManagerClient
}

func (cg *AccessUMClientGroup)GetClient() umrpc.UserManagerClient {
	curClient := <-cg.umClientGroupChan


	return curClient
}

func (cg *AccessUMClientGroup)ReturnClient(curClient umrpc.UserManagerClient) {
	cg.umClientGroupChan <- curClient

	return
}

func NewUMClientGroup(umAddr string, umNumber int) *AccessUMClientGroup {
	return &AccessUMClientGroup{
		UMAddr: umAddr,
		UMClientNum: umNumber,
		umClientGroupChan: make(chan umrpc.UserManagerClient, umNumber),
	}
}

func StartAccessUMClientGroup(cg *AccessUMClientGroup) error {
	err := initUserManagerClientGroup(cg)
	if err != nil {
		return err
	}

	return nil
}

func initUserManagerClient(cg *AccessUMClientGroup, errChan chan uint8) {
	conn, err := grpc.Dial(cg.UMAddr, grpc.WithInsecure())
	if err != nil {
		logger.Errorf("Initialize user manager client failed, err : %s.", err)
		errChan <- ERROR
		return
	}

	logger.Infof("client dial succeed,", conn)

	client := umrpc.NewUserManagerClient(conn)

	cg.umClientGroupChan <- client
	errChan<- SUCCESS

	return
}

func initUserManagerClientGroup(cg *AccessUMClientGroup) error {

	errChan := make(chan uint8, cg.UMClientNum)

	for i := 0; i < cg.UMClientNum; i++{
		go initUserManagerClient(cg, errChan)
	}

	totalFailed := 0

	for i := 0; i < cg.UMClientNum; i++ {
		errCode := <- errChan
		if errCode != SUCCESS {
			totalFailed++
		}
	}

	if totalFailed != 0 {
		logger.Errorf("Init user manager group failed, total failed client number is %d.", totalFailed)
		return errors.New("Init User Manager Client Group Failed.")
	}

	return nil
}