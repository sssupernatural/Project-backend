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

	searcher *se.Engine
}

func New(conf *SMServerConfig) *SMServer {
	dc := dataClient.New(&conf.DataCenterConf)
	dc.InitTasksData()


	gob.Register(st.SMScoringField{})
	newSearcher := &se.Engine{}
	var newSearchInitOptions st.EngineInitOptions
	newSearchInitOptions.NumShards = 1
	newSearchInitOptions.IndexerInitOptions = &st.IndexerInitOptions{
		UserCacheSize: 10,
		SearchResultMax: 10,
	}
	newSearchInitOptions.DefaultRankOptions = &st.RankOptions{
		ScoringCriteria: st.RankByInfo{},
		MaxOutputs: 10,
	}
	newSearchInitOptions.Init()
	newSearcher.Init(newSearchInitOptions)
	logger.Println("Init Search Engine Success!")

	return &SMServer{
		addr: conf.Addr,
		dataClient: dc,
		searcher: newSearcher,
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

	ti := srReq.Task
	ti.Responsers = make([]uint32, 0)

	sf := &st.SMScoringField{
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

	sresp := s.searcher.Search(sr)
	for _, r := range sresp.Users {
		ti.Responsers = append(ti.Responsers, uint32(r.ID))
	}

	resp := generateSearchResponsersResp(comm.RetOK, ti)

	return resp, nil


	return nil, nil
}

func (s *SMServer)InsertUserRecords(ctx context.Context, irReq *smrpc.InsertUserRecordsReq) (*smrpc.InsertUserRecordsResp, error) {
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

		go s.searcher.IndexUser(ui.ID, curRecord, false)
	}

	resp := &smrpc.InsertUserRecordsResp{
		Comm: &smrpc.SearchManagerRespComm{
			ErrorCode: comm.RetOK,
			ErrorMsg: comm.GetErrMsg(comm.RetOK),
		},
	}

	return resp, nil

}