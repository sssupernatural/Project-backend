package server

import (
	"golang.org/x/net/context"
	smrpc "external/grpc/searchManagerRPC"
	"external/comm"
	"dataCenter/dataClient"
	se "searchManager/server/engine"
	st "searchManager/server/types"
	"encoding/gob"
)

type SMServerConfig struct {
	Addr           string
    DataCenterConf dataClient.DataCenterDesc
}

type SMServer struct {
	addr string
	dataClient *dataClient.DataCenterClient

	Searcher *se.Engine
}

func New(conf *SMServerConfig) *SMServer {
	dc := dataClient.New(&conf.DataCenterConf)
	dc.InitTasksData()

	logger.Infof("[Server]Initialize Data Center Client succeed.")


	gob.Register(st.SMScoringField{})
	newSearcher := &se.Engine{}

	var newSearchInitOptions st.EngineInitOptions
	newSearchInitOptions.NumShards = 1
	newSearchInitOptions.IndexerInitOptions = &st.IndexerInitOptions{
		UserCacheSize: 10,
		SearchResultMax: 10,
	}
	newSearchInitOptions.DefaultRankOptions = &st.RankOptions{
		ScoringCriteria: &st.RankByInfo{},
		MaxOutputs: 10,
	}
	newSearcher.Init(newSearchInitOptions)
	logger.Infof("[Server]Initialize Search Engine succeed.")

	return &SMServer{
		addr: conf.Addr,
		dataClient: dc,
		Searcher: newSearcher,
	}
}

func generateSearchRespComm(errCode int32) *smrpc.SearchManagerRespComm {
	return &smrpc.SearchManagerRespComm{
		ErrorCode: errCode,
		ErrorMsg: comm.GetErrMsg(errCode),
	}
}
func generateSearchResponsersResp(errCode int32, ti *comm.TaskInfo) *smrpc.SearchResponsersResp {
	return &smrpc.SearchResponsersResp{
		Comm: generateSearchRespComm(errCode),
		Task: ti,
	}
}

func (s *SMServer)SearchResponsers(ctx context.Context, srReq *smrpc.SearchResponsersReq) (*smrpc.SearchResponsersResp, error) {
	logger.Info("[Server]Receive search request.")
	logger.Infof("[Server]Search request info : task id(%d), requester user id(%d).", srReq.Task.ID, srReq.Task.Requester)
	ti := srReq.Task
	ti.Responsers = make([]uint32, 0)

	sf := st.SMScoringField{
		Sex: ti.Desc.Sex,
		AgeMax: ti.Desc.AgeMax,
		AgeMin: ti.Desc.AgeMin,
		Abis:   ti.Desc.Abilities,
		Locations: ti.Desc.Locations,
		Importance: ti.Desc.ImportanceArray,
	}

	sr := st.SearchRequest{
		AbiHeap: srReq.Task.Desc.Abilities,
		Locations: srReq.Task.Desc.Locations,
		Fields: sf,
	}

	sresp := s.Searcher.Search(sr)
	for _, r := range sresp.Users {
		ti.Responsers = append(ti.Responsers, uint32(r.ID))
	}

	resp := generateSearchResponsersResp(comm.RetOK, ti)

	logger.Println("Handle Search Req success.")

	return resp, nil
}

func (s *SMServer)InsertUserRecords(ctx context.Context, irReq *smrpc.InsertUserRecordsReq) (*smrpc.InsertUserRecordsResp, error) {
	logger.Info("[Server]Receive insert request.")
	logger.Infof("[Server]Insert users num : %d.", irReq.UserRecordNum)
	for _, logU := range irReq.Users {
		logger.Infof("[Server]Insert User : ID(%d), status(%s), sex(%s), age(%d).",
			logU.ID, comm.DescStatus(logU.Status), comm.DescSex(logU.Sex), logU.Age)
	}

	if irReq.UserRecordNum == 0 {
		resp := &smrpc.InsertUserRecordsResp{
			Comm: &smrpc.SearchManagerRespComm{
				ErrorCode: comm.RetOK,
				ErrorMsg: comm.GetErrMsg(comm.RetOK),
			},
		}

		return resp, nil
	}

	for _, ui := range irReq.Users {
		curRecord := st.UserIndexData{
			Info: ui,
		}

		go s.Searcher.IndexUser(ui.ID, curRecord, false)
	}

	resp := &smrpc.InsertUserRecordsResp{
		Comm: &smrpc.SearchManagerRespComm{
			ErrorCode: comm.RetOK,
			ErrorMsg: comm.GetErrMsg(comm.RetOK),
		},
	}

	return resp, nil

}