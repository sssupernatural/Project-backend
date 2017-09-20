package rpc

import (
	"google.golang.org/grpc"
	tmrpc "external/grpc/taskManagerRPC"
)

func InitTaskManagerClient(tmAddr string) (tmrpc.TaskManagerClient, error) {
	conn, err := grpc.Dial(tmAddr, grpc.WithInsecure())
	if err != nil {
		logger.Errorf("Initialize task manager client failed, err : %s.", err)
		return nil, err
	}

	logger.Info("tm client dial succeed.")

	client := tmrpc.NewTaskManagerClient(conn)

	return client, nil
}
