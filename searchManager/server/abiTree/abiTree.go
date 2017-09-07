package abiTree

import (
	"sort"
)

const DefaultMaxCacheResponsersNum = 50

//AbisHeap是由客户端的能力树通过广度优先遍历生成的能力堆

type AbiIndices struct {
	Abi         string
	ParentIndex int32
	IDs         []uint32
}

type AbiIndicesHeap struct {
	AbiIndices []AbiIndices
}

type AbiBranchNode struct {
	indicesNum int
	heapIndex  []int32
	abiIndices []*AbiIndices
}

type AbiBranch struct {
	nodes []AbiBranchNode
}

type AbiTree struct {
	branches []AbiBranch
}

type AbiLevel struct {
	splitNum int
	result   []int
	SplitResults [][]int
}

type SplitPermutation struct {
	split          []int
	result         [][]int
	resultInterger []int
}

func (h *AbiIndicesHeap)ConstructAbiTree() *AbiTree {
	//查找能力堆的叶子能力节点（没有子能力的节点）
	ends := h.constructAbiEnds()

	abiTree := &AbiTree{
		branches: make([]AbiBranch, 0),
	}

	//遍历叶子能力节点，构造能力树的各个能力枝
	for i := len(ends) - 1; i >= 0; i-- {
		//构造能力枝
		newAbiBranch := AbiBranch{
			nodes : make([]AbiBranchNode, 0),
		}

		newAbiBranchNode := AbiBranchNode{
			indicesNum: 1,
			heapIndex: make([]int32, 0),
			abiIndices: make([]*AbiIndices, 0),
		}
		newAbiBranchNode.heapIndex = append(newAbiBranchNode.heapIndex, ends[i])
		newAbiBranchNode.abiIndices = append(newAbiBranchNode.abiIndices, &h.AbiIndices[ends[i]])
		//同父能力节点的叶子能力节点形成一个能力枝的末端枝节点
		for j := i-1; j >= 0; j-- {
			if h.AbiIndices[ends[j]].ParentIndex == h.AbiIndices[ends[i]].ParentIndex && h.AbiIndices[ends[j]].ParentIndex != 0 {
				newAbiBranchNode.heapIndex = append(newAbiBranchNode.heapIndex, ends[j])
				newAbiBranchNode.abiIndices = append(newAbiBranchNode.abiIndices, &h.AbiIndices[ends[j]])
				newAbiBranchNode.indicesNum++
				if j == 0 {
					i = j
					break
				}
			} else {
				i = j+1
				break
			}
		}
		//添加新能力枝的末端枝节点
		newAbiBranch.nodes = append(newAbiBranch.nodes, newAbiBranchNode)
		parentIndex := h.AbiIndices[ends[i]].ParentIndex
		//末端枝节点向上遍历能力节点，形成能力枝
		for parentIndex > 0 {
			newAbiBranchNode := AbiBranchNode{
				indicesNum: 1,
				heapIndex: make([]int32, 0),
				abiIndices: make([]*AbiIndices, 0),
			}
			newAbiBranchNode.heapIndex = append(newAbiBranchNode.heapIndex, parentIndex)
			newAbiBranchNode.abiIndices = append(newAbiBranchNode.abiIndices, &h.AbiIndices[parentIndex])
			newAbiBranch.nodes = append(newAbiBranch.nodes, newAbiBranchNode)
			parentIndex = h.AbiIndices[parentIndex].ParentIndex
		}
		abiTree.branches = append(abiTree.branches, newAbiBranch)
	}

	return abiTree
}

func (l *AbiLevel)splitLevel(level int, splitIndex int) {
	restLevel := 0

	if level == 0 {
		newResult := make([]int, l.splitNum)
		l.SplitResults = append(l.SplitResults, newResult)
		return
	}

	for i := 1; i <= level; i++ {
		if i >= l.result[splitIndex-1] {
			l.result[splitIndex] = i
			restLevel = level - i

			if restLevel == 0 && splitIndex >= 1 {
				newResult := make([]int, l.splitNum)
				copy(newResult, l.result[1:])
				l.SplitResults = append(l.SplitResults, newResult)
			} else if splitIndex+1 < len(l.result) {
				l.splitLevel(restLevel, splitIndex+1)
			}
			l.result[splitIndex] = 0
		}
	}
}

func getSplitInteger(split []int) int {
	integer := 0

	for _, i := range split {
		integer = integer * 10 + i
	}

	return integer
}

func (sp *SplitPermutation)resultNotExist() bool {
	integer := getSplitInteger(sp.split)
	for _, i := range sp.resultInterger {
		if i == integer {
			return false
		}
	}

	sp.resultInterger = append(sp.resultInterger, integer)

	return true
}

func (sp *SplitPermutation)perm(index1 int, index2 int) {
	if index1 == index2 {
		if sp.resultNotExist() {
			newResult := make([]int, len(sp.split))
			copy(newResult, sp.split)
			sp.result = append(sp.result, newResult)
		}

		return
	}

	for i := index1; i < index2; i++ {
		sp.split[index1], sp.split[i] = sp.split[i], sp.split[index1]
		sp.perm(index1+1, index2)
		sp.split[index1], sp.split[i] = sp.split[i], sp.split[index1]
	}
}

type uint32Slice []uint32

func (s uint32Slice) Len() int { return len(s) }
func (s uint32Slice) Swap(i, j int){ s[i], s[j] = s[j], s[i] }
func (s uint32Slice) Less(i, j int) bool { return s[i] < s[j] }
func (s uint32Slice) RemoveDuplicate() []uint32 {
	out := make([]uint32, 0)
	for i := 0; i < len(s)-1; i++ {
		if s[i] != s[i+1] {
			out = append(out, s[i])
		}
	}

	if len(s) > 1 && s[len(s)-1] != s[len(s)-2] {
		out = append(out, s[len(s)-1])
	} else if len(s) == 1 {
		out = append(out, s[0])
	}

	return out
}

type branchEndIndicesCounter struct {
	branchIndex int
	endIndicesNum int
}

type branchEndIndicesSort []branchEndIndicesCounter

func (s branchEndIndicesSort) Len() int { return len(s) }
func (s branchEndIndicesSort) Swap(i, j int){ s[i], s[j] = s[j], s[i] }
func (s branchEndIndicesSort) Less(i, j int) bool { return s[i].endIndicesNum < s[j].endIndicesNum }

//针对一个搜索模式的能力搜索的方法为：
//能力枝的搜索枝节点由搜索模式决定，先对起始能力枝的搜索枝节点中的所有能力节点的用户求并集，再由该用户集与其他能力枝节点的各个能力节点的用户求交集
//该搜索方法优先满足能力广度，即该方法优先返回能力项满足更多的用户
func (t *AbiTree)searchIDsByLevelMode(mode []int) []uint32{
	resultIDs := make([]uint32, 0)
	//某个能力枝的高度达不到搜索级别时直接返回
	for index, branch := range t.branches {
		if len(branch.nodes) <= mode[index] {
			return resultIDs
		}
	}

	//以能力节点数量从小到大对能力枝的枝末端节点进行排序
	//以搜索枝节点的能力节点数量少的能力枝开始搜索，可以先获得满足能力数量更多的用户
	//虽然在能力评分中也会计算能力匹配分，但是在搜索用户数有限的情况下，该方法能优先返回更符合要求的用户
	branchEndSort := make([]branchEndIndicesCounter, 0)
	for index, branch := range t.branches {
		branchEndSort = append(branchEndSort, branchEndIndicesCounter{
			branchIndex: index,
			endIndicesNum: len(branch.nodes[mode[index]].abiIndices),
		})
	}
	sort.Sort(branchEndIndicesSort(branchEndSort))
	logger.Infof("[AbiTree][ModeSearch] : Branch end sort result : %v.", branchEndSort)

	idExistMap := make(map[uint32]bool)
	for i := 0; i < len(t.branches); i++ {
		searchIDs := make([]uint32, 0)
		//获得起始能力枝的编号
		startBranchIndex := branchEndSort[i].branchIndex

		//起始能力枝的搜索枝节点的各个能力节点用户求并集
		for _, indices := range t.branches[startBranchIndex].nodes[mode[startBranchIndex]].abiIndices {
			if indices.IDs != nil && len(indices.IDs) != 0 {
				searchIDs = append(searchIDs, indices.IDs...)
			}
		}
		//对源用户集排序
		sort.Sort(uint32Slice(searchIDs))
		searchIDs = uint32Slice(searchIDs).RemoveDuplicate()

		logger.Infof("[AbiTree][ModeSearch][%d] : start branch index(%d), source IDs(%v).", i, startBranchIndex, searchIDs)

		//遍历其他能力枝求用户交集
		for branchIndex, m := range mode {
			if branchIndex == startBranchIndex {
				continue
			}

			//遍历其他能力枝的搜索直接点的能力节点求用户交集
			for _, indices := range t.branches[branchIndex].nodes[m].abiIndices {
				logger.Infof("[AbiTree][ModeSearch][%d] : compare IDs(%v).", i, indices.IDs)
				if indices.IDs == nil || len(indices.IDs) == 0 {
					searchIDs = searchIDs[0:0]
				} else {
					searchIDs = saveSortedDuplicateIds(searchIDs, indices.IDs)
				}
				logger.Infof("[AbiTree][ModeSearch][%d] : source IDs(%v).", i, searchIDs)
			}

			logger.Infof("[AbiTree][ModeSearch][%d] : result IDs(%v).", i, searchIDs)
		}

		//将搜索到的用户加入搜索结果
		for _, id := range searchIDs {
			_, ok := idExistMap[id]
			if !ok {
				idExistMap[id] = true
				resultIDs = append(resultIDs, id)
			}
		}
	}

	return resultIDs
}

func (t *AbiTree)getBranchesHeightSum() int {
	sum := 0
	for _, branch := range t.branches {
		sum += len(branch.nodes)
	}

	return sum
}

func (t *AbiTree)SearchIDs(requiredIDsNum int) []uint32 {
	curIDsNum := 0
	resultIDs := make([]uint32, 0)
	logger.Info("[AbiTree]Start search.")
	//计算能力树的能力枝的总高度（各个能力枝的枝节点数量总和）
	branchesHeightSum := t.getBranchesHeightSum()
	logger.Infof("[AbiTree]Branches height sum : %d.", branchesHeightSum)
	//从0向上提高搜索级别，级别越低，搜索到的用户能力匹配越高
	// 一个搜索级别代表能力树中某一个能力枝在搜索时的用户来自该能力枝的枝末端节点向上提高一个能力枝节点。搜索级别不高于能力枝总高度，
	for level := 0; level < branchesHeightSum; level++ {
		logger.Infof("[AbiTree]Search level : %d.", level)
		//将搜索级别拆分按能力枝的数量进行拆分
		abiLevel := &AbiLevel{
			splitNum: len(t.branches),
			result: make([]int, len(t.branches)+1),
			SplitResults: make([][]int, 0),
		}
		abiLevel.splitLevel(level, 1)

		//遍历搜索级别的拆分结果
		for _, split := range abiLevel.SplitResults {
			logger.Infof("[AbiTree]Search level split : %v.", split)
			//获取该拆分结果的全排列
			sp := &SplitPermutation{
				split: split,
				result: make([][]int, 0),
				resultInterger: make([]int, 0),
			}
			sp.perm(0, len(sp.split))

			//遍历该拆分结果的全排列
			for _, perm := range sp.result {
				logger.Infof("[AbiTree]Search split perm : %v.", perm)
				//对一个能力级别的一个拆分结果的一个排列进行能力搜索
				modeResult := t.searchIDsByLevelMode(perm)
				logger.Infof("[AbiTree]Search split perm result IDs : %v.", modeResult)
				if len(modeResult) + curIDsNum >= requiredIDsNum {
					resultIDs = append(resultIDs, modeResult[:requiredIDsNum-curIDsNum]...)
					return resultIDs
				} else {
					resultIDs = append(resultIDs, modeResult...)
					curIDsNum += len(modeResult)
				}
			}
		}
	}

	return resultIDs
}

func (h *AbiIndicesHeap)GetAbiIndicesParent(indicesIndex int, deep int) (int, AbiIndices, int32) {
	curIndices := h.AbiIndices[indicesIndex]
	parentIndex := curIndices.ParentIndex
	getDeep := 1
	for i := 0; i < deep ; i++ {
		if parentIndex == 0 {
			return getDeep, h.AbiIndices[parentIndex], parentIndex
		}

		curIndices = h.AbiIndices[parentIndex]
		parentIndex = curIndices.ParentIndex
		getDeep++
	}

	return getDeep, h.AbiIndices[parentIndex], parentIndex
}

func (h *AbiIndicesHeap)constructAbiEnds() []int32 {
	parents := make([]bool, len(h.AbiIndices))
	parents[0] = true
	for i := 1; i < len(h.AbiIndices); i++ {
		parents[h.AbiIndices[i].ParentIndex] = true
	}

	AbiEnds := make([]int32, 0)
	for index, isParent := range parents {
		if !isParent {
			AbiEnds = append(AbiEnds, int32(index))
		}
	}

	return AbiEnds
}

func removeSortedDuplicateIDs(srcIds []uint32, removeIds []uint32) (srcResultIds []uint32) {
	srcTop, srcCur := 0, 0

	removeCur := sort.Search(len(removeIds), func(i int) bool {return removeIds[i] >= srcIds[0]})

	for srcCur < len(srcIds) && removeCur < len(removeIds) {
		if removeIds[removeCur] > srcIds[srcCur] {
			if srcTop != srcCur {
				srcIds[srcTop] = srcIds[srcCur]
			}
			srcTop++
			srcCur++
		} else if removeIds[removeCur] == srcIds[srcCur] {
			removeCur++
			srcCur++
		} else {
			removeCur++
		}
	}

	return append(srcIds[:srcTop], srcIds[srcCur:]...)
}

func saveSortedDuplicateIds(srcIDs []uint32, compareIds []uint32) (srcResultIds []uint32) {
	compareCur := 0

	srcCur := sort.Search(len(srcIDs), func(i int) bool {
		return srcIDs[i] >= compareIds[0]
	})

	srcHead := srcCur
	srcTop  := srcCur

	for srcCur < len(srcIDs) && compareCur < len(compareIds) {
		if srcIDs[srcCur] == compareIds[compareCur] {
			if srcTop != srcCur {
				srcIDs[srcTop] = srcIDs[srcCur]
			}
			srcTop++
			srcCur++
			compareCur++
		} else if srcIDs[srcCur] > compareIds[compareCur] {
			compareCur++
		} else {
			srcCur++
		}
	}

	return srcIDs[srcHead:srcTop]
}

func (h *AbiIndicesHeap)removeDuplicateAbiIndicesIDs() {
	ends := h.constructAbiEnds()

	var parentIndex int32 = 0
	var childIndex int32 = 0
	var end int32 = 0

	for _, end = range ends {
		childIndex = end
		for h.AbiIndices[childIndex].ParentIndex != 0 {
			parentIndex = h.AbiIndices[childIndex].ParentIndex
			for parentIndex != 0 {
				h.AbiIndices[parentIndex].IDs = removeSortedDuplicateIDs(h.AbiIndices[parentIndex].IDs, h.AbiIndices[childIndex].IDs)
				parentIndex = h.AbiIndices[parentIndex].ParentIndex
			}
			childIndex = h.AbiIndices[childIndex].ParentIndex
		}
	}
}

func (h *AbiIndicesHeap)saveDuplicateAbiIndicesIDs(compareIds []uint32) {
	for index, abiIndices := range h.AbiIndices {
		h.AbiIndices[index].IDs = saveSortedDuplicateIds(abiIndices.IDs, compareIds)
	}
}

func (h *AbiIndicesHeap)FilterIDsByAbisIndices() {
	h.removeDuplicateAbiIndicesIDs()
}

func (h *AbiIndicesHeap)FilterIDsByAbisIndicesAndLocationIndices(locationIds []uint32) {
	h.removeDuplicateAbiIndicesIDs()
	h.saveDuplicateAbiIndicesIDs(locationIds)
}