package rpc

import (
	"google.golang.org/grpc"
	umrpc "external/grpc/userManagerRPC"
)

func InitUserManagerClient(umAddr string) (umrpc.UserManagerClient, error) {
	conn, err := grpc.Dial(umAddr, grpc.WithInsecure())
	if err != nil {
		logger.Errorf("Initialize user manager client failed, err : %s.", err)
		return nil, err
	}

	logger.Info("um client dial succeed.")

	client := umrpc.NewUserManagerClient(conn)

	return client, nil
}
