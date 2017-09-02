package types

import (
	"reflect"
	"external/comm"
)

// 评分规则通用接口
type ScoringCriteria interface {
	// 给一个文档评分，文档排序时先用第一个分值比较，如果
	// 分值相同则转移到第二个分值，以此类推。
	// 返回空切片表明该文档应该从最终排序结果中剔除。
	Score(user IndexedUser, fields interface{}, requestFields interface{}) []float32
}

// 一个简单的评分规则，文档分数为BM25
type RankByAbi struct {
}

func (rule RankByAbi) Score(user IndexedUser, fields interface{}, requestFields interface{}) []float32 {
	return []float32{0}
}

// 一个简单的评分规则，文档分数为BM25
type RankByInfo struct {
}

func (rule RankByInfo) Score(user IndexedUser, fields interface{}, requestFields interface{}) []float32 {
	if reflect.TypeOf(fields) != reflect.TypeOf(SMScoringField{}) ||
		reflect.TypeOf(requestFields) != reflect.TypeOf(SMScoringField{}) {
		return nil
	}

	f := fields.(SMScoringField)
	rf := requestFields.(SMScoringField)

	if rf.Sex != comm.UserSexUnknown {
		if rf.Sex != f.Sex {
			return nil
		}
	}

	if rf.AgeMin > f.Age || rf.AgeMax < f.Age {
		return nil
	}

	if f.Status == comm.UserStatusOffline {
		return nil
	}

	return []float32{0}
}


