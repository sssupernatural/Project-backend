package core

import (
	"sort"
	"sync"
	"searchManager/server/types"
)

type Ranker struct {
	lock struct {
		sync.RWMutex
		fields map[uint32]interface{}
		users   map[uint32]bool
	}
	initialized bool
}

func (ranker *Ranker) Init() {
	if ranker.initialized == true {
		logger.Fatal("排序器不能初始化两次")
	}
	ranker.initialized = true

	ranker.lock.fields = make(map[uint32]interface{})
	ranker.lock.users = make(map[uint32]bool)
}

// 给某个用户添加评分字段
func (ranker *Ranker) AddDoc(userID uint32, fields interface{}) {
	if ranker.initialized == false {
		logger.Fatal("排序器尚未初始化")
	}

	ranker.lock.Lock()
	ranker.lock.fields[userID] = fields
	ranker.lock.users[userID] = true
	ranker.lock.Unlock()
}

// 删除某个用户的评分字段
func (ranker *Ranker) RemoveDoc(userID uint32) {
	if ranker.initialized == false {
		logger.Fatal("排序器尚未初始化")
	}

	ranker.lock.Lock()
	delete(ranker.lock.fields, userID)
	delete(ranker.lock.users, userID)
	ranker.lock.Unlock()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 给文档评分并排序
func (ranker *Ranker) Rank(
	users []types.IndexedUser, options types.RankOptions, countDocsOnly bool, fields interface{}) (types.ScoredUsers, int) {
	if ranker.initialized == false {
		logger.Fatal("排序器尚未初始化")
	}

	// 对每个用户评分
	var outputUsers types.ScoredUsers
	numUsers := 0
	for _, u := range users {
		ranker.lock.RLock()
		// 判断doc是否存在
		if _, ok := ranker.lock.users[u.ID]; ok {
			fs := ranker.lock.fields[u.ID]
			ranker.lock.RUnlock()
			// 计算评分并剔除没有分值的用户
			scores := options.ScoringCriteria.Score(u, fs, fields)
			if len(scores) > 0 {
				if !countDocsOnly {
					outputUsers = append(outputUsers, types.ScoredUser{
						ID:     u.ID,
						Scores: scores,
					})
				}
				numUsers++
			}
		} else {
			ranker.lock.RUnlock()
		}
	}

	// 排序
	if !countDocsOnly {
		if options.ReverseOrder {
			sort.Sort(sort.Reverse(outputUsers))
		} else {
			sort.Sort(outputUsers)
		}
		// 当用户要求只返回部分结果时返回部分结果
		var start, end int
		if options.MaxOutputs != 0 {
			start = minInt(options.OutputOffset, len(outputUsers))
			end = minInt(options.OutputOffset+options.MaxOutputs, len(outputUsers))
		} else {
			start = minInt(options.OutputOffset, len(outputUsers))
			end = len(outputUsers)
		}
		return outputUsers[start:end], numUsers
	}
	return outputUsers, numUsers
}
