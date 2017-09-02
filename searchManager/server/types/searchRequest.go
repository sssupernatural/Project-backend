package types

import "external/comm"

type SearchRequest struct {
	AbiHeap *comm.AbisHeap

	Locations []*comm.Location

	// 当不为nil时，仅从这些UserIDs包含的键中搜索（忽略值）
	UserIDs map[uint32]bool

	// 排序选项
	RankOptions *RankOptions

	// 超时，单位毫秒（千分之一秒）。此值小于等于零时不设超时。
	// 搜索超时的情况下仍有可能返回部分排序结果。
	Timeout int

	// 设为true时仅统计搜索到的文档个数，不返回具体的文档
	CountDocsOnly bool

	// 不排序，对于可在引擎外部（比如客户端）排序情况适用
	// 对返回文档很多的情况打开此选项可以有效节省时间
	Orderless bool

	Fields interface{}
}

type RankOptions struct {
	// 文档的评分规则，值为nil时使用Engine初始化时设定的规则
	ScoringCriteria ScoringCriteria

	// 默认情况下（ReverseOrder=false）按照分数从大到小排序，否则从小到大排序
	ReverseOrder bool

	// 从第几条结果开始输出
	OutputOffset int

	// 最大输出的搜索结果数，为0时无限制
	MaxOutputs int
}