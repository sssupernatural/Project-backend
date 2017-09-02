package types

type SearchResponse struct {
	// 搜索到的用户，已排序
	Users []ScoredUser

	// 搜索是否超时。超时的情况下也可能会返回部分结果
	Timeout bool

	// 搜索到的用户个数。注意这是全部用户中满足条件的个数，可能比返回的用户数要大
	NumUsers int
}

type ScoredUser struct {
	ID uint32

	// 用户的打分值
	// 搜索结果按照Scores的值排序，先按照第一个数排，如果相同则按照第二个数排序，依次类推。
	Scores []float32
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 为了方便排序

type ScoredUsers []ScoredUser

func (users ScoredUsers) Len() int {
	return len(users)
}
func (users ScoredUsers) Swap(i, j int) {
	users[i], users[j] = users[j], users[i]
}
func (users ScoredUsers) Less(i, j int) bool {
	// 为了从大到小排序，这实际上实现的是More的功能
	for iScore := 0; iScore < minInt(len(users[i].Scores), len(users[j].Scores)); iScore++ {
		if users[i].Scores[iScore] > users[j].Scores[iScore] {
			return true
		} else if users[i].Scores[iScore] < users[j].Scores[iScore] {
			return false
		}
	}
	return len(users[i].Scores) > len(users[j].Scores)
}