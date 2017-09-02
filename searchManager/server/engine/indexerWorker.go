package engine

import (
	"sync/atomic"
	"searchManager/server/types"
	"external/comm"
)

type indexerAddUserRequest struct {
	user        *types.UserIndex
	forceUpdate bool
}

type indexerLookupRequest struct {
	countDocsOnly       bool
	abisHeap            *comm.AbisHeap
    locationOwners      []comm.Location
	userIds             map[uint32]bool
	options             types.RankOptions
	rankerReturnChannel chan rankerReturnRequest
	orderless           bool
	fields              interface{}
}

type indexerRemoveUserRequest struct {
	userId      uint32
	forceUpdate bool
}

func (engine *Engine) indexerAddUserWorker(shard int) {
	for {
		request := <-engine.indexerAddUserChannels[shard]
		engine.indexers[shard].AddUserToCache(request.user, request.forceUpdate)
		if request.user != nil {
			atomic.AddUint64(&engine.numAbiIndexAdded,
				uint64(len(request.user.KeyAbis)))
			atomic.AddUint64(&engine.numLocIndexAdded,
				uint64(len(request.user.KeyLocs)))
			atomic.AddUint64(&engine.numUsersIndexed, 1)
		}
		if request.forceUpdate {
			atomic.AddUint64(&engine.numUsersForceUpdated, 1)
		}
	}
}

func (engine *Engine) indexerRemoveUserWorker(shard int) {
	for {
		request := <-engine.indexerRemoveUserChannels[shard]
		engine.indexers[shard].RemoveUserToCache(request.userId, request.forceUpdate)
		if request.userId != 0 {
			atomic.AddUint64(&engine.numUsersRemoved, 1)
		}
		if request.forceUpdate {
			atomic.AddUint64(&engine.numUsersForceUpdated, 1)
		}
	}
}

func (engine *Engine) indexerLookupWorker(shard int) {
	for {
		request := <-engine.indexerLookupChannels[shard]

		var users []types.IndexedUser
		var numUsers int
		if request.userIds == nil {
			users, numUsers = engine.indexers[shard].Lookup(request.abisHeap, request.locationOwners, nil, request.countDocsOnly)
		} else {
			users, numUsers = engine.indexers[shard].Lookup(request.abisHeap, request.locationOwners, request.userIds, request.countDocsOnly)
		}

		if request.countDocsOnly {
			request.rankerReturnChannel <- rankerReturnRequest{numUsers: numUsers}
			continue
		}

		if len(users) == 0 {
			request.rankerReturnChannel <- rankerReturnRequest{}
			continue
		}

		if request.orderless {
			var outputUsers []types.ScoredUser
			for _, u := range users {
				outputUsers = append(outputUsers, types.ScoredUser{ID: u.ID})
			}
			request.rankerReturnChannel <- rankerReturnRequest{
				users:    outputUsers,
				numUsers: len(outputUsers),
			}
			continue
		}

		rankerRequest := rankerRankRequest{
			countDocsOnly:       request.countDocsOnly,
			users:               users,
			options:             request.options,
			rankerReturnChannel: request.rankerReturnChannel,
			fields:              request.fields,
		}

		engine.rankerRankChannels[shard] <- rankerRequest
	}
}

