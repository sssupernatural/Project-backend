package types

import (
	"reflect"
	"external/comm"
	"sort"
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
		logger.Info("[Ranker]type error.")
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


	/*logger.Printf("f.status : %s\n", comm.DescStatus(f.Status))
	if f.Status == comm.UserStatusOffline {
		return nil
	}
	*/


	score := make([]float32, 0)

	//排序第一级 主要能力分
	var mainAbi string
	if rf.Importance[0] != 0 {
		mainAbi = rf.Abis.ABIs[rf.Importance[0]].ABI
	}
	var firstAbi string
	if rf.Importance[1] != 0 {
		firstAbi = rf.Abis.ABIs[rf.Importance[1]].ABI
	}
	var secondAbi string
	if rf.Importance[2] != 0 {
		secondAbi = rf.Abis.ABIs[rf.Importance[2]].ABI
	}

	topAbisScore := 0
	for _, fAbi := range f.Abis.ABIs {
		if mainAbi != "" && mainAbi == fAbi.ABI {
			topAbisScore += 10
		} else if firstAbi != "" && firstAbi == fAbi.ABI {
			topAbisScore += 5
		} else if secondAbi != "" && secondAbi == fAbi.ABI {
			topAbisScore += 2
		}
	}

	score = append(score, float32(topAbisScore))

	//排序第二级 当前位置在响应位置中
	if rf.Locations != nil && len(rf.Locations) != 0 {
		rfLocOwner := GenerateLocOwner(rf.Locations)
		fLocOwner := GenerateLocOwner(f.Locations)

		var curLocOwner comm.Location
		scoreCurLoc := 0
		if f.CurLocation != nil {
			curLocOwner = GenerateOwnerLocation(f.CurLocation)
			fLocOwner = append(fLocOwner, curLocOwner)
			rfLocOwner = saveSortedDuplicateLocs(rfLocOwner, fLocOwner)
			for _, loc := range rfLocOwner {
				if loc == curLocOwner {
					scoreCurLoc = 1
				}
			}
		}
		score = append(score, float32(scoreCurLoc))

		//排序第三级 响应者的常用位置在响应位置的数量
		scoreLocs := len(rfLocOwner)
		score = append(score, float32(scoreLocs))
	}

	//排序第四级 其他能力得分
	scoreAbi := 0
	for _, rfAbi := range rf.Abis.ABIs {
		for _, fAbi := range f.Abis.ABIs {
			if rfAbi.ABI == fAbi.ABI {
				scoreAbi++
			}
		}
	}
	score = append(score, float32(scoreAbi))

	return score
}

type Locs []comm.Location

func (locs Locs) Len() int {
	return len(locs)
}

func (locs Locs) Swap(i, j int) {
	locs[i], locs[j] = locs[j], locs[i]
}

func (locs Locs) Less(i, j int) bool {
	if locs[i].Longitude == locs[j].Longitude {
		return locs[i].Latitude < locs[j].Latitude
	} else {
		return locs[i].Longitude < locs[j].Longitude
	}
}


func saveSortedDuplicateLocs(srcLocs Locs, compareLocs Locs) (srcResultLocs Locs) {
	compareCur := 0

	sort.Sort(compareLocs)
	sort.Sort(srcLocs)

	srcCur := sort.Search(len(srcLocs), func(i int) bool {
		return srcLocs[i].Longitude >= compareLocs[0].Longitude
	})

	srcHead := srcCur
	srcTop  := srcCur

	for srcCur < len(srcLocs) && compareCur < len(compareLocs) {
		if srcLocs[srcCur] == compareLocs[compareCur] {
			if srcTop != srcCur {
				srcLocs[srcTop] = srcLocs[srcCur]
			}
			srcTop++
			srcCur++
			compareCur++
		} else if srcLocs[srcCur].Longitude >= compareLocs[compareCur].Longitude {
			if srcLocs[srcCur].Latitude >= compareLocs[compareCur].Latitude {
				compareCur++
			} else {
				srcCur++
			}
		} else {
			srcCur++
		}
	}

	return srcLocs[srcHead:srcTop]
}


