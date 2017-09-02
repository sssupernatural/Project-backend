package rpc

import (
	"google.golang.org/grpc"
	smrpc "external/grpc/searchManagerRPC"
	"errors"
)

const (
	ERROR   uint8 = 0
	SUCCESS uint8 = 1
)

type SMClientGroup struct {
	SMAddr            string
	SMClientNum       int
	smClientGroupChan chan smrpc.SearchManagerClient
}

func (cg *SMClientGroup)GetClient() smrpc.SearchManagerClient {
	curClient := <-cg.smClientGroupChan

	return curClient
}

func (cg *SMClientGroup)ReturnClient(curClient smrpc.SearchManagerClient) {
	cg.smClientGroupChan <- curClient

	return
}

func NewSMClientGroup(smAddr string, smNumber int) *SMClientGroup {
	return &SMClientGroup{
		SMAddr: smAddr,
		SMClientNum: smNumber,
		smClientGroupChan: make(chan smrpc.SearchManagerClient, smNumber),
	}
}

func StartSMClientGroup(cg *SMClientGroup) error {
	err := initSearchManagerClientGroup(cg)
	if err != nil {
		return err
	}

	return nil
}

func initSearchManagerClient(cg *SMClientGroup, errChan chan uint8) {
	conn, err := grpc.Dial(cg.SMAddr, grpc.WithInsecure())
	if err != nil {
		logger.Errorf("Initialize search manager client failed, err : %s.", err)
		errChan <- ERROR
		return
	}

	logger.Infof("client dial succeed,", conn)

	client := smrpc.NewSearchManagerClient(conn)

	cg.smClientGroupChan <- client
	errChan<- SUCCESS

	return
}

func initSearchManagerClientGroup(cg *SMClientGroup) error {

	errChan := make(chan uint8, cg.SMClientNum)

	for i := 0; i < cg.SMClientNum; i++{
		go initSearchManagerClient(cg, errChan)
	}

	totalFailed := 0

	for i := 0; i < cg.SMClientNum; i++ {
		errCode := <- errChan
		if errCode != SUCCESS {
			totalFailed++
		}
	}

	if totalFailed != 0 {
		logger.Errorf("Init search manager group failed, total failed client number is %d.", totalFailed)
		return errors.New("Init Search Manager Client Group Failed.")
	}

	return nil
}