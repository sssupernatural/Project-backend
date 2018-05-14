package types

import "external/comm"

const (
	locationDivideUnit float64 = 1
)

type UserIndex struct {
	ID      uint32
	KeyAbis []KeyAbiIndex
	KeyLocs []KeyLocIndex
}

type KeyAbiIndex struct {
	Abi        string
	Experience int32
}

type KeyLocIndex struct {
	Location comm.Location
}

type IndexedUser struct {
	ID uint32
}

type UsersIndex []*UserIndex

func GenerateOwnerLocation(srcLoc *comm.Location) comm.Location {
	longitude := float64(int64(srcLoc.Longitude / locationDivideUnit)) * locationDivideUnit
	latitude := float64(int64(srcLoc.Latitude / locationDivideUnit)) * locationDivideUnit
	return comm.Location{Longitude: longitude, Latitude:latitude}
}

func GenerateLocOwner(locs []*comm.Location) []comm.Location {
	out := make([]comm.Location, 0)

	if locs == nil || len(locs) == 0 {
		return out
	}

	for _, l := range locs {
		out = append(out, GenerateOwnerLocation(l))
	}

	return out
}

func (users UsersIndex) Len() int {
	return len(users)
}

func (users UsersIndex) Swap(i, j int) {
	users[i], users[j] = users[j], users[i]
}

func (users UsersIndex) Less(i, j int) bool {
	return users[i].ID < users[j].ID
}

type UsersID []uint32

func (users UsersID) Len() int {
	return len(users)
}

func (users UsersID) Swap(i, j int) {
	users[i], users[j] = users[j], users[i]
}

func (users UsersID) Less(i, j int) bool {
	return users[i] < users[j]
}
