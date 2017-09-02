package types

import (
	"runtime"
	"time"
)

var (
	// EngineInitOptions的默认值
	defaultNumShards                 = 2
	defaultIndexerBufferLength       = runtime.NumCPU()
	defaultNumIndexerThreadsPerShard = runtime.NumCPU()
	defaultRankerBufferLength        = runtime.NumCPU()
	defaultNumRankerThreadsPerShard  = runtime.NumCPU()
	defaultDefaultRankOptions        = RankOptions{
		ScoringCriteria: RankByInfo{},
		MaxOutputs: defaultSearchResultMax,
	}
	defaultIndexerInitOptions = IndexerInitOptions{
		UserCacheSize: defaultUserCacheSize,
	}

	defaultFlushIndexesPeriod = time.Second * 10
)

type EngineInitOptions struct {
	// 索引器和排序器的shard数目
	// 被检索/排序的文档会被均匀分配到各个shard中
	NumShards int

	// 索引器的信道缓冲长度
	IndexerBufferLength int

	// 索引器每个shard分配的线程数
	NumIndexerThreadsPerShard int

	// 排序器的信道缓冲长度
	RankerBufferLength int

	// 排序器每个shard分配的线程数
	NumRankerThreadsPerShard int

	// 索引器初始化选项
	IndexerInitOptions *IndexerInitOptions

	// 默认的搜索选项
	DefaultRankOptions *RankOptions

	//强制刷新索引周期
	FlushIndexesPeriod time.Duration
}

// 初始化EngineInitOptions，当用户未设定某个选项的值时用默认值取代
func (options *EngineInitOptions) Init() {
	if options.NumShards == 0 {
		options.NumShards = defaultNumShards
	}

	if options.IndexerBufferLength == 0 {
		options.IndexerBufferLength = defaultIndexerBufferLength
	}

	if options.NumIndexerThreadsPerShard == 0 {
		options.NumIndexerThreadsPerShard = defaultNumIndexerThreadsPerShard
	}

	if options.RankerBufferLength == 0 {
		options.RankerBufferLength = defaultRankerBufferLength
	}

	if options.NumRankerThreadsPerShard == 0 {
		options.NumRankerThreadsPerShard = defaultNumRankerThreadsPerShard
	}

	if options.IndexerInitOptions == nil {
		options.IndexerInitOptions = &defaultIndexerInitOptions
	}

	if options.DefaultRankOptions == nil {
		options.DefaultRankOptions = &defaultDefaultRankOptions
	}

	if options.DefaultRankOptions.ScoringCriteria == nil {
		options.DefaultRankOptions.ScoringCriteria = defaultDefaultRankOptions.ScoringCriteria
	}

	if options.FlushIndexesPeriod == 0 {
		options.FlushIndexesPeriod = defaultFlushIndexesPeriod
	}
}
