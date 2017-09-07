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
		logger.Infof("[Ranker]Handle ranker add user req: user id(%d).", request.userID)
		engine.rankers[shard].AddUserRankField(request.userID, request.fields)
	}
}

func (engine *Engine) rankerRankWorker(shard int) {
	for {
		request := <-engine.rankerRankChannels[shard]
		logger.Infof("[Ranker]Receive rank request, req : %v.", request)
		logger.Infof("[Ranker]Rank option : %v.", request.options)
		if request.options.MaxOutputs != 0 {
			request.options.MaxOutputs += request.options.OutputOffset
		}
		request.options.OutputOffset = 0
		outputUsers, numUsers := engine.rankers[shard].Rank(request.users, request.options, request.countDocsOnly, request.fields)
		logger.Infof("[Ranker]output Users after rank : %v.", outputUsers)
		request.rankerReturnChannel <- rankerReturnRequest{users: outputUsers, numUsers: numUsers}
	}
}

func (engine *Engine) rankerRemoveUserWorker(shard int) {
	for {
		request := <-engine.rankerRemoveUserChannels[shard]
		engine.rankers[shard].RemoveUserRankField(request.userID)
	}
}
