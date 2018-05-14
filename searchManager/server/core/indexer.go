package core

import (
	"sync"
	"searchManager/server/types"
	"sort"
	"external/comm"
	"searchManager/server/abiTree"
)

type KeyAbiIndices struct {
	IDs []uint32
}

type KeyLocIndices struct {
	IDs []uint32
}

type Indexer struct {
	//从能力搜索键到用户的列表的反向索引
	tableLock struct{
		sync.RWMutex
		table      map[string]*KeyAbiIndices
		locTable   map[comm.Location]*KeyLocIndices
		usersState map[uint32]int   // nil: 表示无状态记录，0: 存在于索引中，1: 等待删除，2: 等待加入
	}

	addCacheLock struct{
		sync.RWMutex
		addCachePointer int
		addCache        types.UsersIndex
	}

	removeCacheLock struct{
		sync.RWMutex
		removeCachePointer int
		removeCache        types.UsersID
	}

	initOptions types.IndexerInitOptions
	initialized bool

	// 这实际上是总用户数的一个近似值
	numUsers uint64

	// 所有被索引用户的总能力键数量
	totalAbiLength float32
}

func (indexer *Indexer)constructAbiIndicesHeap(abisHeap *comm.AbisHeap) *abiTree.AbiIndicesHeap {

	abiIndicesHeap := &abiTree.AbiIndicesHeap{
		AbiIndices: make([]abiTree.AbiIndices, 0),
	}

	indexer.tableLock.RLock()
	for rangeIndex, abi := range abisHeap.ABIs {
		if _, ok := indexer.tableLock.table[abi.ABI]; ok && rangeIndex != 0 {
			abiIDs := make([]uint32, len(indexer.tableLock.table[abi.ABI].IDs))
			copy(abiIDs, indexer.tableLock.table[abi.ABI].IDs)
			abiIndices := abiTree.AbiIndices{
				Abi: abi.ABI,
				ParentIndex: abi.ParentIndex,
				IDs: abiIDs,
			}

			abiIndicesHeap.AbiIndices = append(abiIndicesHeap.AbiIndices, abiIndices)
		} else {
			abiIndices := abiTree.AbiIndices{
				Abi: abi.ABI,
				ParentIndex: abi.ParentIndex,
			}

			abiIndicesHeap.AbiIndices = append(abiIndicesHeap.AbiIndices, abiIndices)
		}
	}
	indexer.tableLock.RUnlock()

	return abiIndicesHeap
}

// 初始化索引器
func (indexer *Indexer) Init(options types.IndexerInitOptions) {
	if indexer.initialized == true {
		logger.Fatal("索引器不能初始化两次")
	}
	options.Init()
	indexer.initOptions = options
	indexer.initialized = true

	indexer.tableLock.table = make(map[string]*KeyAbiIndices)
	indexer.tableLock.locTable = make(map[comm.Location]*KeyLocIndices)
	indexer.tableLock.usersState = make(map[uint32]int)
	indexer.addCacheLock.addCache = make([]*types.UserIndex, indexer.initOptions.UserCacheSize)
	indexer.removeCacheLock.removeCache = make([]uint32, indexer.initOptions.UserCacheSize*2)
}

// 从KeyAbiIndices中得到第i个用户的ID
func (indexer *Indexer) getUserId(ti *KeyAbiIndices, i int) uint32 {
	return ti.IDs[i]
}

// 得到KeyAbiIndices中用户总数
func (indexer *Indexer) getIndexLength(ti *KeyAbiIndices) int {
	return len(ti.IDs)
}

// 从KeyLocIndices中得到第i个用户的ID
func (indexer *Indexer) getLocUserId(ti *KeyLocIndices, i int) uint32 {
	return ti.IDs[i]
}

// 得到keyLoc中用户总数
func (indexer *Indexer) getLocIndexLength(ti *KeyLocIndices) int {
	return len(ti.IDs)
}

// 向 ADDCACHE 中加入一个用户
func (indexer *Indexer) AddUserToCache(user *types.UserIndex, forceUpdate bool) {
	if indexer.initialized == false {
		logger.Fatal("索引器尚未初始化")
	}

	indexer.addCacheLock.Lock()
	//SWT : 将用户添加到"等待加入索引的用户缓存区"
	if user != nil {
		indexer.addCacheLock.addCache[indexer.addCacheLock.addCachePointer] = user
		indexer.addCacheLock.addCachePointer++
		logger.Infof("[Indexer]Add user to indexer cache succeed, user id(%d).", user.ID)
	}
	//缓存满或者强制更新时，将"等待加入索引的用户缓存区"中的用户添加到索引表中
	if indexer.addCacheLock.addCachePointer >= indexer.initOptions.UserCacheSize || forceUpdate {
		indexer.tableLock.Lock()
		position := 0
		for i := 0; i < indexer.addCacheLock.addCachePointer; i++ {
			userIndex := indexer.addCacheLock.addCache[i]
			if userState, ok := indexer.tableLock.usersState[userIndex.ID]; ok && userState <= 1 {
				// ok && userState == 0 表示存在于索引中，需先删除再添加
				// ok && userState == 1 表示不一定存在于索引中，等待删除，需先删除再添加
				//将先删除再添加的用户放置到"等待加入索引的用户缓存区"的前端
				if position != i {
					indexer.addCacheLock.addCache[position], indexer.addCacheLock.addCache[i] =
						indexer.addCacheLock.addCache[i], indexer.addCacheLock.addCache[position]
				}
				//将"等待加入索引的用户缓存区"中那些存在于索引中的用户添加到"等待从索引中删除的用户缓存区"
				if userState == 0 {
					indexer.removeCacheLock.Lock()
					indexer.removeCacheLock.removeCache[indexer.removeCacheLock.removeCachePointer] =
						userIndex.ID
					indexer.removeCacheLock.removeCachePointer++
					indexer.removeCacheLock.Unlock()
					indexer.tableLock.usersState[userIndex.ID] = 1
					indexer.numUsers--
				}
				position++
			} else if !ok {
				//将不存在于索引中的待添加用户在索引的用户状态置为"待添加"
				indexer.tableLock.usersState[userIndex.ID] = 2
			}
		}

		indexer.tableLock.Unlock()
		if indexer.RemoveUserToCache(0, forceUpdate) {
			// 只有当存在于索引表中的用户已被删除，其才可以重新加入到索引表中
			position = 0
		}

		addCachedUsers := indexer.addCacheLock.addCache[position:indexer.addCacheLock.addCachePointer]
		indexer.addCacheLock.addCachePointer = position
		indexer.addCacheLock.Unlock()
		sort.Sort(addCachedUsers)
		indexer.AddUsers(&addCachedUsers)
	} else {
		indexer.addCacheLock.Unlock()
	}

}

// 向反向索引表中加入 ADDCACHE 中所有用户
func (indexer *Indexer) AddUsers(users *types.UsersIndex) {
	if indexer.initialized == false {
		logger.Fatal("索引器尚未初始化")
	}

	indexer.tableLock.Lock()
	defer indexer.tableLock.Unlock()
	indexPointers := make(map[string]int, len(indexer.tableLock.table))
	indexLocPointers := make(map[comm.Location]int, len(indexer.tableLock.table))

	// UserID 递增顺序遍历插入用户保证索引移动次数最少
	for i, user := range *users {
		logger.Infof("[Indexer]Add user to index : user id(%d).", user.ID)
		if i < len(*users)-1 && (*users)[i].ID == (*users)[i+1].ID {
			// 如果有重复用户加入，因为稳定排序，只加入最后一个
			continue
		}
		if userState, ok := indexer.tableLock.usersState[user.ID]; ok && userState == 1 {
			// 如果此时 userState 仍为 1，说明该用户需被删除
			// userState 合法状态为 nil & 2，保证一定不会插入已经在索引表中的用户
			continue
		}

		userIdIsNew := true
		for _, keyAbi := range user.KeyAbis {
			indices, foundKeyAbi := indexer.tableLock.table[keyAbi.Abi]
			if !foundKeyAbi {
				// 如果没找到该搜索键则加入
				ti := KeyAbiIndices{}
				ti.IDs = []uint32{user.ID}
				indexer.tableLock.table[keyAbi.Abi] = &ti
				continue
			}

			// 查找应该插入的位置，且索引一定不存在
			position, _ := indexer.searchIndex(
				indices, indexPointers[keyAbi.Abi], indexer.getIndexLength(indices)-1, user.ID)
			indexPointers[keyAbi.Abi] = position
			indices.IDs = append(indices.IDs, 0)
			copy(indices.IDs[position+1:], indices.IDs[position:])
			indices.IDs[position] = user.ID
		}

		for _, keyLoc := range user.KeyLocs {
			ownerLoc := types.GenerateOwnerLocation(&keyLoc.Location)
			indices, foundKeyLoc := indexer.tableLock.locTable[ownerLoc]
			if !foundKeyLoc {
				// 如果没找到该搜索键则加入
				ti := KeyLocIndices{}
				ti.IDs = []uint32{user.ID}
				indexer.tableLock.locTable[ownerLoc] = &ti
				continue
			}

			// 查找应该插入的位置，且索引一定不存在
			position, _ := indexer.searchLocIndex(
				indices, indexLocPointers[ownerLoc], indexer.getLocIndexLength(indices)-1, user.ID)
			indexLocPointers[ownerLoc] = position
			indices.IDs = append(indices.IDs, 0)
			copy(indices.IDs[position+1:], indices.IDs[position:])
			indices.IDs[position] = user.ID
		}

		// 更新用户状态和总数
		if userIdIsNew {
			indexer.tableLock.usersState[user.ID] = 0
			indexer.numUsers++
		}

		/*
		logger.Println("after add users abimap :")
		for abimkey, abim := range indexer.tableLock.table {
			logger.Printf("[%s|%v] - ", abimkey, *abim)
		}
		logger.Println()
		logger.Println("after add users locmap :")
		for locmkey, locm := range indexer.tableLock.locTable {
			logger.Printf("[%v|%v] - ", locmkey, *locm)
		}
		logger.Println()
		*/
	}
}

// 向 REMOVECACHE 中加入一个待删除用户
// 返回值表示用户是否在索引表中被删除
func (indexer *Indexer) RemoveUserToCache(userID uint32, forceUpdate bool) bool {
	if indexer.initialized == false {
		logger.Fatal("索引器尚未初始化")
	}

	indexer.removeCacheLock.Lock()
	if userID != 0 {
		indexer.tableLock.Lock()
		if userState, ok := indexer.tableLock.usersState[userID]; ok && userState == 0 {
			indexer.removeCacheLock.removeCache[indexer.removeCacheLock.removeCachePointer] = userID
			indexer.removeCacheLock.removeCachePointer++
			indexer.tableLock.usersState[userID] = 1
			indexer.numUsers--
			logger.Infof("[Indexer]Remove user from indexer cache succeed, user id(%d), user cache status(0->1).", userID)
		} else if ok && userState == 2 {
			// 删除一个等待加入的用户
			indexer.tableLock.usersState[userID] = 1
			logger.Infof("[Indexer]Remove user from indexer cache succeed, user id(%d), user cache status(2->1).", userID)
		} else if !ok {
			// 若用户不存在，则无法判断其是否在 addCache 中，需避免这样的操作
			logger.Errorf("[Indexer]Remove user from indexer cache failed, user id(%d).", userID)
		}
		indexer.tableLock.Unlock()
	}

	if indexer.removeCacheLock.removeCachePointer > 0 &&
		(indexer.removeCacheLock.removeCachePointer >= indexer.initOptions.UserCacheSize ||
			forceUpdate) {
		removeCachedUsers := indexer.removeCacheLock.removeCache[:indexer.removeCacheLock.removeCachePointer]
		indexer.removeCacheLock.removeCachePointer = 0
		indexer.removeCacheLock.Unlock()
		sort.Sort(removeCachedUsers)
		indexer.RemoveUsers(&removeCachedUsers)
		return true
	}
	indexer.removeCacheLock.Unlock()
	return false
}

// 向反向索引表中删除 REMOVECACHE 中所有用户
func (indexer *Indexer) RemoveUsers(users *types.UsersID) {
	if indexer.initialized == false {
		logger.Fatal("索引器尚未初始化")
	}

	indexer.tableLock.Lock()
	defer indexer.tableLock.Unlock()

	if users != nil {
		logger.Info("[Indexer]Remove users from index, users id : %v.", users)
	}

	//删除用户状态
	for _, userID := range *users {
		delete(indexer.tableLock.usersState, userID)
	}

	for keyAbi, indices := range indexer.tableLock.table {
		indicesTop, indicesPointer := 0, 0
		usersPointer := sort.Search(
			len(*users), func(i int) bool { return (*users)[i] >= indices.IDs[0] })
		// 双指针扫描，进行批量删除操作
		for usersPointer < len(*users) && indicesPointer < indexer.getIndexLength(indices) {
			if indices.IDs[indicesPointer] < (*users)[usersPointer] {
				if indicesTop != indicesPointer {
					indices.IDs[indicesTop] = indices.IDs[indicesPointer]
				}
				indicesTop++
				indicesPointer++
			} else if indices.IDs[indicesPointer] == (*users)[usersPointer] {
				indicesPointer++
				usersPointer++
			} else {
				usersPointer++
			}
		}
		if indicesTop != indicesPointer {
			indices.IDs = append(
				indices.IDs[:indicesTop], indices.IDs[indicesPointer:]...)
		}
		if len(indices.IDs) == 0 {
			delete(indexer.tableLock.table, keyAbi)
		}
	}

	for keyLoc, indices := range indexer.tableLock.locTable {
		indicesTop, indicesPointer := 0, 0
		usersPointer := sort.Search(
			len(*users), func(i int) bool { return (*users)[i] >= indices.IDs[0] })
		// 双指针扫描，进行批量删除操作
		for usersPointer < len(*users) && indicesPointer < indexer.getLocIndexLength(indices) {
			if indices.IDs[indicesPointer] < (*users)[usersPointer] {
				if indicesTop != indicesPointer {
					indices.IDs[indicesTop] = indices.IDs[indicesPointer]
				}
				indicesTop++
				indicesPointer++
			} else if indices.IDs[indicesPointer] == (*users)[usersPointer] {
				indicesPointer++
				usersPointer++
			} else {
				usersPointer++
			}
		}
		if indicesTop != indicesPointer {
			indices.IDs = append(
				indices.IDs[:indicesTop], indices.IDs[indicesPointer:]...)
		}
		if len(indices.IDs) == 0 {
			delete(indexer.tableLock.locTable, keyLoc)
		}
	}

	logger.Info("[Indexer]Remove users from index finished.")
}

func (indexer *Indexer)getLocationOwnersIDs(locationOwners []comm.Location) []uint32 {
	locationsIDs := make([]uint32, 0)
	indexer.tableLock.RLock()
	for _, loc := range locationOwners {
		if _, ok := indexer.tableLock.locTable[loc]; ok {
			locationsIDs = append(locationsIDs, indexer.tableLock.locTable[loc].IDs...)
		}
	}
	indexer.tableLock.RUnlock()

	return locationsIDs
}

// 查找符合请求要求的用户
// 当userIds不为nil时仅从userIds指定的用户中查找 : 未实现
func (indexer *Indexer) Lookup(
	abisHeap *comm.AbisHeap, locationOwners []comm.Location, userIDs map[uint32]bool, countDocsOnly bool) (users []types.IndexedUser, numUsers int) {
	if indexer.initialized == false {
		logger.Fatal("索引器尚未初始化")
	}

	numUsers = 0

	//根据索引表和能力堆构造能力索引堆
	abisIndicesHeap := indexer.constructAbiIndicesHeap(abisHeap)
	logger.Infof("[Indexer]Construct abi indices heap : %v.", abisIndicesHeap)
	if len(locationOwners) == 0 {
		//过滤能力堆中的用户，规则是父能力节点不包含子能力节点的用户
		logger.Infof("[Indexer]Filter abi indices heap IDs.")
		abisIndicesHeap.FilterIDsByAbisIndices()
	} else {
		//过滤能力堆中的用户，规则是父能力节点不包含子能力节点的用户，且所有能力节点都只包含在归属坐标组中的用户
		locationOwnersIds := indexer.getLocationOwnersIDs(locationOwners)
		logger.Infof("[Indexer]Filter abi indices heap IDs with location owner IDs(%v).", locationOwnersIds)
		if len(locationOwnersIds) != 0 {
			abisIndicesHeap.FilterIDsByAbisIndicesAndLocationIndices(locationOwnersIds)
		} else {
			abisIndicesHeap.FilterIDsByAbisIndices()
		}
	}
	logger.Infof("[Indexer]Filtered abi indices heap : %v", abisIndicesHeap)

	//根据能力索引堆构建能力索引树
	newAbiTree := abisIndicesHeap.ConstructAbiTree()

	findUsers := newAbiTree.SearchIDs(indexer.initOptions.SearchResultMax)
	logger.Infof("[Indexer]Abis tree search find users : %v.", findUsers)

	resultUsers := make([]types.IndexedUser, 0)
	for _, id := range findUsers {
		newIndexedUser := types.IndexedUser{
			ID: id,
		}
		resultUsers = append(resultUsers, newIndexedUser)
	}

	return resultUsers, len(resultUsers)
}

// 二分法查找indices中某用户的索引项
// 第一个返回参数为找到的位置或需要插入的位置
// 第二个返回参数标明是否找到
func (indexer *Indexer) searchIndex(
	indices *KeyAbiIndices, start int, end int, userId uint32) (int, bool) {
	// 特殊情况
	if indexer.getIndexLength(indices) == start {
		return start, false
	}
	if userId < indexer.getUserId(indices, start) {
		return start, false
	} else if userId == indexer.getUserId(indices, start) {
		return start, true
	}
	if userId > indexer.getUserId(indices, end) {
		return end + 1, false
	} else if userId == indexer.getUserId(indices, end) {
		return end, true
	}

	// 二分
	var middle int
	for end-start > 1 {
		middle = (start + end) / 2
		if userId == indexer.getUserId(indices, middle) {
			return middle, true
		} else if userId > indexer.getUserId(indices, middle) {
			start = middle
		} else {
			end = middle
		}
	}
	return end, false
}

// 二分法查找indices中某用户的索引项
// 第一个返回参数为找到的位置或需要插入的位置
// 第二个返回参数标明是否找到
func (indexer *Indexer) searchLocIndex(
	indices *KeyLocIndices, start int, end int, userId uint32) (int, bool) {
	// 特殊情况
	if indexer.getLocIndexLength(indices) == start {
		return start, false
	}
	if userId < indexer.getLocUserId(indices, start) {
		return start, false
	} else if userId == indexer.getLocUserId(indices, start) {
		return start, true
	}
	if userId > indexer.getLocUserId(indices, end) {
		return end + 1, false
	} else if userId == indexer.getLocUserId(indices, end) {
		return end, true
	}

	// 二分
	var middle int
	for end-start > 1 {
		middle = (start + end) / 2
		if userId == indexer.getLocUserId(indices, middle) {
			return middle, true
		} else if userId > indexer.getLocUserId(indices, middle) {
			start = middle
		} else {
			end = middle
		}
	}
	return end, false
}


