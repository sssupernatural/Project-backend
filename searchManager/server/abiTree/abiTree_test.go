package abiTree

import (
	"testing"
	"fmt"
	"time"
)

func TestSaveSortedDuplicate(t *testing.T)  {
	src := []uint32{3,4,5,8}
	compare := []uint32{3,4,5,8}

	src = saveSortedDuplicateIds(src, compare)

	fmt.Printf("%v\n", src)
}

func TestRemoveSortedDuplicate(t *testing.T)  {
	srcArray := []uint32{3,4,5,8}
	compareArray := []uint32{3,4,5,8}

	src := srcArray[:]
	compare := compareArray[:]

	src = removeSortedDuplicateIDs(src, compare)

	fmt.Printf("%v\n", src)
}

func TestRemoveDuplicateAbiIndicesIDs(t *testing.T) {
	h := AbiIndicesHeap{
		AbiIndices: make([]AbiIndices, 3),
	}

	h.AbiIndices[0] = AbiIndices{
		Abi: "Root",
		ParentIndex: 0,
	}

	/*
	ids1 := []uint32{4,5,6,7,8,10}
	h.AbiIndices[1] = AbiIndices{
		Abi: "运动",
		ParentIndex: 0,
		IDs: ids1,
	}
	*/


	ids2 := []uint32{3,5,8,9,10}
	h.AbiIndices[1] = AbiIndices{
		Abi: "艺术",
		ParentIndex: 0,
		IDs: ids2,
	}

	/*
	ids3 := []uint32{5,6}
	h.AbiIndices[3] = AbiIndices{
		Abi: "足球",
		ParentIndex: 1,
		IDs: ids3,
	}

	ids4 := []uint32{5,7,11}
	h.AbiIndices[4] = AbiIndices{
		Abi: "羽毛球",
		ParentIndex: 1,
		IDs: ids4,
	}
	*/

	ids5 := []uint32{5,9,10,11}
	h.AbiIndices[2] = AbiIndices{
		Abi: "文学",
		ParentIndex: 2,
		IDs: ids5,
	}


	//locationIds := []uint32{3,5,8,9,10}

	h.FilterIDsByAbisIndices()
	fmt.Printf("heap : %v\n", h)
	abiTree := h.ConstructAbiTree()
	fmt.Printf("TREE: %v\n", abiTree)
	searchIDs := abiTree.SearchIDs(10)

	fmt.Printf("\nsearch : %v\n", searchIDs)

}

func TestSplitLevel(t *testing.T) {

	for i := 0; i <= 5; i++ {
		Level := &AbiLevel{
			splitNum: 2,
			result: make([]int, 3),
			SplitResults: make([][]int, 0),
		}

		Level.splitLevel(i, 1)

		fmt.Printf("\n%d, %v\n", i, Level)
	}
}

func TestSplitPerm(t *testing.T) {
	split := []int{1,2,0,0,0}

	sp := &SplitPermutation{
		split: split,
		result: make([][]int, 0),
	}

	sp.perm(0, len(sp.split))

	fmt.Printf("\n%v\n", sp.result)

	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for _ = range ticker.C {
			fmt.Println("!!!")
		}
	}()

	time.Sleep(time.Second*50)
}
