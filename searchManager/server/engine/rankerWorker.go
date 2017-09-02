package engine

import (
	"searchManager/server/types"
)

type rankerAddUserRequest struct {
	userID uint32
	fields interface{}
}

type rankerRankRequest struct {
	users               []types.IndexedUser
	options             types.RankOptions
	rankerReturnChannel chan rankerReturnRequest
	countDocsOnly       bool
	fields              interface{}
}

type rankerReturnRequest struct {
	users    types.ScoredUsers
	numUsers int
}

type rankerRemoveUserRequest struct {
	userID uint32
}

func (engine *Engine) rankerAddUserWorker(shard int) {
	for {
		request := <-engine.rankerAddUserChannels[shard]
		engine.rankers[shard].AddDoc(request.userID, request.fields)
	}
}

func (engine *Engine) rankerRankWorker(shard int) {
	for {
		request := <-engine.rankerRankChannels[shard]
		if request.options.MaxOutputs != 0 {
			request.options.MaxOutputs += request.options.OutputOffset
		}
		request.options.OutputOffset = 0
		outputUsers, numUsers := engine.rankers[shard].Rank(request.users, request.options, request.countDocsOnly, request.fields)
		request.rankerReturnChannel <- rankerReturnRequest{users: outputUsers, numUsers: numUsers}
	}
}

func (engine *Engine) rankerRemoveUserWorker(shard int) {
	for {
		request := <-engine.rankerRemoveUserChannels[shard]
		engine.rankers[shard].RemoveDoc(request.userID)
	}
}
