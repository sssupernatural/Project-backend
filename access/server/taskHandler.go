package server

import (
	"net/http"
	"io/ioutil"
	"external/comm"
	"encoding/json"
	tmrpc "external/grpc/taskManagerRPC"
	"context"
	"errors"
)

const (
	TaskActionQuery    string = "QUERY"
	TaskActionAccept   string = "ACCEPT"
	TaskActionChoose   string = "CHOOSE"
	TaskActionFulfil   string = "FULFIL"
	TaskActionEvaluate string = "EVALUATE"
)

func (s *AccessServer)tasksHandler(w http.ResponseWriter, req *http.Request) {
	logger.Infoln("Receive a tasks request")

	defer req.Body.Close()
	reqData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Errorf("Read request body failed, err: %s.\n", err)
		http.Error(w, "Server read request body failed!", 400)
		return
	}

	if req.Method == "POST" {
		var createInfo comm.TaskCreateInfo

		err = json.Unmarshal(reqData, &createInfo)
		if err != nil {
			logger.Errorf("Unmarshal request body failed, err: %s.\n", err)
			http.Error(w, "Server unmarshal request body failed!", 400)
			return
		}

		err = s.handleTaskCreate(w, &createInfo)
	} else {
		var curTaskAction comm.TaskAction

		err = json.Unmarshal(reqData, &curTaskAction)
		if err != nil {
			logger.Errorf("Unmarshal request body failed, err: %s.\n", err)
			http.Error(w, "Server unmarshal request body failed!", 400)
			return
		}

		switch curTaskAction.Action {
		case TaskActionQuery : err = s.handleTaskActionQuery(w, &curTaskAction)
		case TaskActionAccept    : err = s.handleTaskActionAccept(w, &curTaskAction)
		case TaskActionChoose   : err = s.handleTaskActionChoose(w, &curTaskAction)
		case TaskActionFulfil    : err = s.handleTaskActionFulfil(w, &curTaskAction)
		case TaskActionEvaluate   : err = s.handleTaskActionEvaluate(w, &curTaskAction)
		default:
			logger.Errorf("Unexpected task action [%s]!.\n", curTaskAction.Action)
			http.Error(w, "Unexpected task action!", 400)
			return
		}
	}

	if err == nil {
		logger.Infoln("Handle tasks request succeed!")
	} else {
		w.Header().Set("ErrorMsg", err.Error())
		http.Error(w, "Server handle request error!", 500)
	}

	return

}

func (s *AccessServer)handleTaskCreate(w http.ResponseWriter, tc *comm.TaskCreateInfo) error {
	taskCreateReq := &tmrpc.CreateTaskReq{
		CreateInfo: tc,
	}

	resp, err := s.tmClient.CreateTask(context.Background(), taskCreateReq)
	if err != nil {
		logger.Errorf("Access RPC Task Manager Failed, err[%s].\n", err)
		return err
	}

	if resp.Comm.ErrorCode != comm.RetOK {
		logger.Errorf("Task Manager returns error[%s], Create Task failed. req[%v].\n",
			resp.Comm.ErrorMsg, tc)
		return errors.New(resp.Comm.ErrorMsg)
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func (s *AccessServer)handleTaskActionQuery(w http.ResponseWriter, a *comm.TaskAction) error {
	queryTasksReq := &tmrpc.QueryUserTasksReq {
		UserID: a.UserID,
	}

	resp, err := s.tmClient.QueryUserTasks(context.Background(), queryTasksReq)
	if err != nil {
		logger.Errorf("Access RPC task Manager Failed, err[%s].\n", err)
		return err
	}

	if resp.Comm.ErrorCode != comm.RetOK {
		logger.Errorf("Task Manager returns error[%s], query user's tasks failed. req:%v.\n",
			resp.Comm.ErrorMsg, a)
		return errors.New(resp.Comm.ErrorMsg)
	}

	respData, err := json.Marshal(resp.Tasks)
	if err != nil {
		logger.Errorf("Marshal response data failed! err : %s.\n", err)
		return err
	}

	w.Write(respData)

	return nil
}

func (s *AccessServer)handleTaskActionAccept(w http.ResponseWriter, a *comm.TaskAction) error {
	accepttasksReq := &tmrpc.AcceptTaskReq {
		TaskID: a.TaskID,
		ResponserID: a.UserID,
		Decision: a.Decision,
	}

	resp, err := s.tmClient.AcceptTask(context.Background(), accepttasksReq)
	if err != nil {
		logger.Errorf("Access RPC task Manager Failed, err[%s].\n", err)
		return err
	}

	if resp.Comm.ErrorCode != comm.RetOK {
		logger.Errorf("Task Manager returns error[%s], accept failed. req:%v.\n",
			resp.Comm.ErrorMsg, a)
		return errors.New(resp.Comm.ErrorMsg)
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func (s *AccessServer)handleTaskActionChoose(w http.ResponseWriter, a *comm.TaskAction) error {
	chooseResponserReq := &tmrpc.ChooseTaskResponserReq {
		TaskID: a.TaskID,
		ChoseResponsersIDs: a.ChosenResponserIDs,
	}

	resp, err := s.tmClient.ChooseTaskResponser(context.Background(), chooseResponserReq)
	if err != nil {
		logger.Errorf("Access RPC task Manager Failed, err[%s].\n", err)
		return err
	}

	if resp.Comm.ErrorCode != comm.RetOK {
		logger.Errorf("Task Manager returns error[%s], choose task responser failed. req:%v.\n",
			resp.Comm.ErrorMsg, a)
		return errors.New(resp.Comm.ErrorMsg)
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func (s *AccessServer)handleTaskActionFulfil(w http.ResponseWriter, a *comm.TaskAction) error {
	fulfilTaskReq := &tmrpc.FulfilTaskReq {
		TaskID: a.TaskID,
		ResponserID: a.UserID,
	}

	resp, err := s.tmClient.FulfilTask(context.Background(), fulfilTaskReq)
	if err != nil {
		logger.Errorf("Access RPC task Manager Failed, err[%s].\n", err)
		return err
	}

	if resp.Comm.ErrorCode != comm.RetOK {
		logger.Errorf("Task Manager returns error[%s], fulfil task failed. req:%v.\n",
			resp.Comm.ErrorMsg, a)
		return errors.New(resp.Comm.ErrorMsg)
	}

	w.WriteHeader(http.StatusOK)

	return nil
}

func (s *AccessServer)handleTaskActionEvaluate(w http.ResponseWriter, a *comm.TaskAction) error {
	evaluateTaskReq := &tmrpc.EvaluateAndFinishTaskReq {
		TaskID: a.TaskID,
		RequesterID: a.UserID,
	}

	resp, err := s.tmClient.EvaluateAndFinishTask(context.Background(), evaluateTaskReq)
	if err != nil {
		logger.Errorf("Access RPC task Manager Failed, err[%s].\n", err)
		return err
	}

	if resp.Comm.ErrorCode != comm.RetOK {
		logger.Errorf("Task Manager returns error[%s], evaluate task failed. req:%v.\n",
			resp.Comm.ErrorMsg, a)
		return errors.New(resp.Comm.ErrorMsg)
	}

	w.WriteHeader(http.StatusOK)

	return nil
}