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


	gob.Register(st.SMScoringField{})
	newSearcher := &se.Engine{}
	logger.Println("Init Search Engine Success!")

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
	newSearchInitOptions.Init()
	newSearcher.Init(newSearchInitOptions)

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
	logger.Println("Receive a Search Req.")
	ti := srReq.Task
	ti.Responsers = make([]uint32, 0)

	sf := st.SMScoringField{
		Sex: ti.Desc.Sex,
		AgeMax: ti.Desc.AgeMax,
		AgeMin: ti.Desc.AgeMin,
		Abis:   ti.Desc.Abilities,
		Locations: ti.Desc.Locations,
	}

	sr := st.SearchRequest{
		AbiHeap: srReq.Task.Desc.Abilities,
		Locations: srReq.Task.Desc.Locations,
		Fields: sf,
	}

	if sr.RankOptions == nil {
		logger.Info("[Server]Rank option is nil.")
	} else {
		logger.Infof("[Server]Rank option : %v.", sr.RankOptions)
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
	logger.Printf("Receive a Insert Users Req, user number : %d.\n",irReq.UserRecordNum)
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

		logger.Printf("interface index user : id(%d), user(%v).\n", ui.ID, *curRecord.Info)

		go s.Searcher.IndexUser(ui.ID, curRecord, false)
	}

	resp := &smrpc.InsertUserRecordsResp{
		Comm: &smrpc.SearchManagerRespComm{
			ErrorCode: comm.RetOK,
			ErrorMsg: comm.GetErrMsg(comm.RetOK),
		},
	}
	//s.Searcher.FlushIndex()

	logger.Printf("Handle Insert Users Req Succeed, user number : %d.\n",irReq.UserRecordNum)

	return resp, nil

}