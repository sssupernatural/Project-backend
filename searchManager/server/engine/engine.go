package engine

import (
	"runtime"
	"sync/atomic"
	"github.com/huichen/murmur"
	"fmt"
	"time"
	"sort"
	"searchManager/server/types"
	"searchManager/server/core"
)

const (
	NumNanosecondsInAMillisecond = 1000000
)

type Engine struct {
	// 计数器，用来统计有多少用户被索引等信息
	numUsersIndexed          uint64
	numUsersRemoved          uint64
	numUsersForceUpdated     uint64
	numIndexingRequests      uint64
	numRemovingRequests      uint64
	numForceUpdatingRequests uint64
	numAbiIndexAdded         uint64
	numLocIndexAdded         uint64
	numUsersStored           uint64

	// 记录初始化参数
	initOptions types.EngineInitOptions
	initialized bool

	indexers   []core.Indexer
	rankers    []core.Ranker

	// 建立索引器使用的通信通道
	indexerAddUserChannels    []chan indexerAddUserRequest
	indexerRemoveUserChannels []chan indexerRemoveUserRequest
	rankerAddUserChannels     []chan rankerAddUserRequest

	// 建立排序器使用的通信通道
	indexerLookupChannels    []chan indexerLookupRequest
	rankerRankChannels       []chan rankerRankRequest
	rankerRemoveUserChannels []chan rankerRemoveUserRequest
}

func (engine *Engine) Init(options types.EngineInitOptions) {
	// 将线程数设置为CPU数
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 初始化初始参数
	if engine.initialized {
		logger.Fatal("请勿重复初始化引擎")
	}
	options.Init()
	engine.initOptions = options
	engine.initialized = true

	// 初始化索引器和排序器
	for shard := 0; shard < options.NumShards; shard++ {
		engine.indexers = append(engine.indexers, core.Indexer{})
		engine.indexers[shard].Init(*options.IndexerInitOptions)

		engine.rankers = append(engine.rankers, core.Ranker{})
		engine.rankers[shard].Init()
	}

	// 初始化索引器通道
	engine.indexerAddUserChannels = make([]chan indexerAddUserRequest, options.NumShards)
	engine.indexerRemoveUserChannels = make([]chan indexerRemoveUserRequest, options.NumShards)
	engine.indexerLookupChannels = make([]chan indexerLookupRequest, options.NumShards)
	for shard := 0; shard < options.NumShards; shard++ {
		engine.indexerAddUserChannels[shard] = make(chan indexerAddUserRequest, options.IndexerBufferLength)
		engine.indexerRemoveUserChannels[shard] = make(chan indexerRemoveUserRequest, options.IndexerBufferLength)
		engine.indexerLookupChannels[shard] = make(chan indexerLookupRequest, options.IndexerBufferLength)
	}

	// 初始化排序器通道
	engine.rankerAddUserChannels = make([]chan rankerAddUserRequest, options.NumShards)
	engine.rankerRankChannels = make([]chan rankerRankRequest, options.NumShards)
	engine.rankerRemoveUserChannels = make([]chan rankerRemoveUserRequest, options.NumShards)
	for shard := 0; shard < options.NumShards; shard++ {
		engine.rankerAddUserChannels[shard] = make(chan rankerAddUserRequest, options.RankerBufferLength)
		engine.rankerRankChannels[shard] = make(chan rankerRankRequest, options.RankerBufferLength)
		engine.rankerRemoveUserChannels[shard] = make(chan rankerRemoveUserRequest, options.RankerBufferLength)
	}

	// 启动索引器和排序器
	for shard := 0; shard < options.NumShards; shard++ {
		go engine.indexerAddUserWorker(shard)
		go engine.indexerRemoveUserWorker(shard)
		go engine.rankerAddUserWorker(shard)
		go engine.rankerRemoveUserWorker(shard)

		for i := 0; i < options.NumIndexerThreadsPerShard; i++ {
			go engine.indexerLookupWorker(shard)
		}
		for i := 0; i < options.NumRankerThreadsPerShard; i++ {
			go engine.rankerRankWorker(shard)
		}
	}

	//启动周期强制刷新索引
	/*
	flushTicker := time.NewTicker(engine.initOptions.FlushIndexesPeriod)
	logger.Printf("Start period flush: period(%v).\n", engine.initOptions.FlushIndexesPeriod)
	go func() {
		for _ = range flushTicker.C {
			logger.Printf("Force flush, req num = %d.\n", engine.numIndexingRequests)
			engine.FlushIndex()
		}
	}()
	*/

	return
}

/*
func (engine *Engine)ForceFlushIndexesPeriod() {
	logger.Println("Start period flush Indexes.")
	tick := time.NewTicker(engine.initOptions.FlushIndexesPeriod)

	for range tick.C {
		logger.Printf("Force flush, req num = %d.\n", engine.numIndexingRequests)
		if engine.numIndexingRequests > 0 {
			engine.FlushIndex()
		}
	}
}
*/

// 将用户加入索引
//
// 输入参数：
//  userId	      标识用户编号，必须唯一，userId == 0 表示非法用户（用于强制刷新索引），[1, +oo) 表示合法用户
//  data	      见UserIndexData注释
//  forceUpdate   是否强制刷新 cache，如果设为 true，则尽快添加到索引，否则等待 cache 满之后一次全量添加
//
// 注意：
//      1. 这个函数是线程安全的，请尽可能并发调用以提高索引速度
//      2. 这个函数调用是非同步的，也就是说在函数返回时有可能用户还没有加入索引中，因此
//         如果立刻调用Search可能无法查询到这个用户。强制刷新索引请调用FlushIndex函数。
func (engine *Engine) IndexUser(userId uint32, data types.UserIndexData, forceUpdate bool) {
	engine.internalIndexUser(userId, data, forceUpdate)
}

func (engine *Engine) internalIndexUser(
	userId uint32, data types.UserIndexData, forceUpdate bool) {
	if !engine.initialized {
		logger.Fatal("必须先初始化引擎")
	}

	if userId != 0 {
		atomic.AddUint64(&engine.numIndexingRequests, 1)
	}
	if forceUpdate {
		atomic.AddUint64(&engine.numForceUpdatingRequests, 1)
	}

	hash := murmur.Murmur3([]byte(fmt.Sprint("%d%v", userId, data)))
	shard := engine.getShard(hash)

	if userId != 0 {
		indexerRequest := indexerAddUserRequest{
			user: &types.UserIndex{
				ID:      data.GetID(),
				KeyAbis: data.GetKeyAbiIndexes(),
				KeyLocs: data.GetKeyLocIndexes(),
			},
			forceUpdate: forceUpdate,
		}

		engine.indexerAddUserChannels[shard] <- indexerRequest
	}

	if forceUpdate {
		for i := 0; i < engine.initOptions.NumShards; i++ {
			if i == shard {
				continue
			}
			engine.indexerAddUserChannels[i] <- indexerAddUserRequest{forceUpdate: true}
		}
	}

	if userId != 0 {
		rankerRequest := rankerAddUserRequest{
			userID: userId,
			fields: data.GetScoringField(),
		}

		engine.rankerAddUserChannels[shard] <- rankerRequest
	}
}

// 将用户从索引中删除
//
// 输入参数：
//  docId	      标识用户编号，必须唯一，docId == 0 表示非法用户（用于强制刷新索引），[1, +oo) 表示合法用户
//  forceUpdate 是否强制刷新 cache，如果设为 true，则尽快删除索引，否则等待 cache 满之后一次全量删除
//
// 注意：
//      1. 这个函数是线程安全的，请尽可能并发调用以提高索引速度
//      2. 这个函数调用是非同步的，也就是说在函数返回时有可能用户还没有加入索引中，因此
//         如果立刻调用Search可能无法查询到这个用户。强制刷新索引请调用FlushIndex函数。
func (engine *Engine) RemoveUser(userID uint32, forceUpdate bool) {
	if !engine.initialized {
		logger.Fatal("必须先初始化引擎")
	}

	if userID != 0 {
		atomic.AddUint64(&engine.numRemovingRequests, 1)
	}
	if forceUpdate {
		atomic.AddUint64(&engine.numForceUpdatingRequests, 1)
	}
	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		engine.indexerRemoveUserChannels[shard] <- indexerRemoveUserRequest{userId: userID, forceUpdate: forceUpdate}
		if userID == 0 {
			continue
		}
		engine.rankerRemoveUserChannels[shard] <- rankerRemoveUserRequest{userID: userID}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 查找满足搜索条件的文档，此函数线程安全
func (engine *Engine) Search(request types.SearchRequest) (output types.SearchResponse) {
	if !engine.initialized {
		logger.Fatal("必须先初始化引擎")
	}

	var rankOptions types.RankOptions
	if request.RankOptions == nil {
		rankOptions = *engine.initOptions.DefaultRankOptions
	} else {
		rankOptions = *request.RankOptions
	}
	if rankOptions.ScoringCriteria == nil {
		rankOptions.ScoringCriteria = engine.initOptions.DefaultRankOptions.ScoringCriteria
	}

	// 建立排序器返回的通信通道
	rankerReturnChannel := make(
	chan rankerReturnRequest, engine.initOptions.NumShards)

	locOwners := types.GenerateLocOwner(request.Locations)

	logger.Printf("search req abis : %v.\n", request.AbiHeap)
	logger.Printf("search req locs : %v.\n", locOwners)

	logger.Infof("[Engine]Rank option : %v.", rankOptions)

	// 生成查找请求
	lookupRequest := indexerLookupRequest{
		countDocsOnly:       request.CountDocsOnly,
		abisHeap:            request.AbiHeap,
		locationOwners:      locOwners,
		userIds:             request.UserIDs,
		options:             rankOptions,
		rankerReturnChannel: rankerReturnChannel,
		orderless:           request.Orderless,
		fields:              request.Fields,
	}

	// 向索引器发送查找请求
	for shard := 0; shard < engine.initOptions.NumShards; shard++ {
		engine.indexerLookupChannels[shard] <- lookupRequest
	}

	// 从通信通道读取排序器的输出
	numUsers := 0
	rankOutput := types.ScoredUsers{}
	timeout := request.Timeout
	isTimeout := false
	if timeout <= 0 {
		// 不设置超时
		for shard := 0; shard < engine.initOptions.NumShards; shard++ {
			rankerOutput := <-rankerReturnChannel
			if !request.CountDocsOnly {
				for _, user := range rankerOutput.users {
					rankOutput = append(rankOutput, user)
				}
			}
			numUsers += rankerOutput.numUsers
		}
	} else {
		// 设置超时
		deadline := time.Now().Add(time.Nanosecond * time.Duration(NumNanosecondsInAMillisecond*request.Timeout))
		for shard := 0; shard < engine.initOptions.NumShards; shard++ {
			select {
			case rankerOutput := <-rankerReturnChannel:
				if !request.CountDocsOnly {
					for _, user := range rankerOutput.users {
						rankOutput = append(rankOutput, user)
					}
				}
				numUsers += rankerOutput.numUsers
			case <-time.After(deadline.Sub(time.Now())):
				isTimeout = true
				break
			}
		}
	}

	// 再排序
	if !request.CountDocsOnly && !request.Orderless {
		if rankOptions.ReverseOrder {
			sort.Sort(sort.Reverse(rankOutput))
		} else {
			sort.Sort(rankOutput)
		}
	}

	logger.Infof("[Engine]Ranker output : %v.", rankOutput)

	// 准备输出
	// 仅当CountDocsOnly为false时才充填output.Docs
	if !request.CountDocsOnly {
		if request.Orderless {
			// 无序状态无需对Offset截断
			output.Users = rankOutput
		} else {
			var start, end int
			if rankOptions.MaxOutputs == 0 {
				start = minInt(rankOptions.OutputOffset, len(rankOutput))
				end = len(rankOutput)
			} else {
				start = minInt(rankOptions.OutputOffset, len(rankOutput))
				end = minInt(start+rankOptions.MaxOutputs, len(rankOutput))
			}
			output.Users = rankOutput[start:end]
		}
	}
	output.NumUsers = numUsers
	output.Timeout = isTimeout
	logger.Infof("[Engine]Search output : %v.", output)
	return
}

// 阻塞等待直到所有索引添加完毕
func (engine *Engine) FlushIndex() {
	logger.Println("flush index!")
	for {
		runtime.Gosched()
		if engine.numIndexingRequests == engine.numUsersIndexed &&
			engine.numRemovingRequests*uint64(engine.initOptions.NumShards) == engine.numUsersRemoved {
			// 保证 CHANNEL 中 REQUESTS 全部被执行完
			break
		}
	}
	// 强制更新，保证其为最后的请求
	engine.IndexUser(0, types.UserIndexData{}, true)
	for {
		runtime.Gosched()
		if engine.numForceUpdatingRequests*uint64(engine.initOptions.NumShards) == engine.numUsersForceUpdated {
			return
		}
	}
}

// 关闭引擎
func (engine *Engine) Close() {
	engine.FlushIndex()
}

// 从文本hash得到要分配到的shard
func (engine *Engine) getShard(hash uint32) int {
	return int(hash - hash/uint32(engine.initOptions.NumShards)*uint32(engine.initOptions.NumShards))
}