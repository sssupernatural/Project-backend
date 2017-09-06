package abiTree

import (
	"sort"
	"fmt"
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
	ends := h.constructAbiEnds()

	abiTree := &AbiTree{
		branches: make([]AbiBranch, 0),
	}

	for i := len(ends) - 1; i >= 0; i-- {
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
		newAbiBranch.nodes = append(newAbiBranch.nodes, newAbiBranchNode)
		parentIndex := h.AbiIndices[ends[i]].ParentIndex
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


func (t *AbiTree)searchIDsByLevelMode(mode []int) []uint32{
	resultIDs := make([]uint32, 0)
	for index, branch := range t.branches {
		if len(branch.nodes) <= mode[index] {
			return resultIDs
		}
	}

	branchEndSort := make([]branchEndIndicesCounter, 0)
	for index, branch := range t.branches {
		branchEndSort = append(branchEndSort, branchEndIndicesCounter{
			branchIndex: index,
			endIndicesNum: len(branch.nodes[mode[index]].abiIndices),
		})
	}
	sort.Sort(branchEndIndicesSort(branchEndSort))
	fmt.Printf("branch sort:%v\n", branchEndSort)

	idExistMap := make(map[uint32]bool)
	for i := 0; i < len(t.branches); i++ {
		searchIDs := make([]uint32, 0)
		startBranchIndex := branchEndSort[i].branchIndex

		for _, indices := range t.branches[startBranchIndex].nodes[mode[startBranchIndex]].abiIndices {
			if indices.IDs != nil && len(indices.IDs) != 0 {
				searchIDs = append(searchIDs, indices.IDs...)
			}
		}

		sort.Sort(uint32Slice(searchIDs))
		searchIDs = uint32Slice(searchIDs).RemoveDuplicate()

		fmt.Printf("\nsrcIDs : %v | ", searchIDs)

		for branchIndex, m := range mode {
			if branchIndex == startBranchIndex {
				continue
			}

			for _, indices := range t.branches[branchIndex].nodes[m].abiIndices {
				fmt.Printf("compIDs : %v | ", indices.IDs)
				if indices.IDs == nil || len(indices.IDs) == 0 {
					searchIDs = searchIDs[0:0]
				} else {
					searchIDs = saveSortedDuplicateIds(searchIDs, indices.IDs)
				}
				fmt.Printf("srcIDs : %v | ", searchIDs)
			}
		}

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
	branchesHeightSum := t.getBranchesHeightSum()
	for level := 0; level < branchesHeightSum; level++ {
		abiLevel := &AbiLevel{
			splitNum: len(t.branches),
			result: make([]int, len(t.branches)+1),
			SplitResults: make([][]int, 0),
		}

		fmt.Printf("split level : %d\n", level)
		abiLevel.splitLevel(level, 1)
		for _, split := range abiLevel.SplitResults {
			sp := &SplitPermutation{
				split: split,
				result: make([][]int, 0),
				resultInterger: make([]int, 0),
			}
			sp.perm(0, len(sp.split))

			for _, perm := range sp.result {
				fmt.Printf("perm : %v | ", perm)
				modeResult := t.searchIDsByLevelMode(perm)
				fmt.Printf("modeResult : %v\n", modeResult)
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