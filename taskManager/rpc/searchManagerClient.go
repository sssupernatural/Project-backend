package rpc

import (
	"google.golang.org/grpc"
	smrpc "external/grpc/searchManagerRPC"
)

func InitSearchManagerClient(smAddr string) (smrpc.SearchManagerClient) {
	conn, err := grpc.Dial(smAddr, grpc.WithInsecure())
	if err != nil {
		logger.Errorf("Initialize search manager client failed, err : %s.", err)
		return nil
	}

	logger.Info("search manager client dial succeed.")

	client := smrpc.NewSearchManagerClient(conn)

	return client
}