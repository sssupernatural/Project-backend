package types

import (
	"external/comm"
)

type UserIndexData struct {
	Info *comm.UserInfo
}

type SMScoringField struct {
	Sex         int32
	Age         int32
	AgeMin      int32
	AgeMax      int32
	Status      int32
	Abis        *comm.AbisHeap
	CurLocation *comm.Location
	Locations   []*comm.Location
}

func (d *UserIndexData)GetScoringField() *SMScoringField {
	sf := &SMScoringField{
		Sex:         d.Info.Sex,
		Age:         d.Info.Age,
		Status:      d.Info.Status,
		Abis:        d.Info.Abilities,
		CurLocation: d.Info.CurLocation,
		Locations:   d.Info.Locations,
	}

	return sf
}

func (d *UserIndexData)GetID() uint32 {
	return d.Info.ID
}

func (d *UserIndexData)GetKeyAbiIndexes() []KeyAbiIndex {
	out := make([]KeyAbiIndex, 0)
	for index, abi := range d.Info.Abilities.ABIs {
		if index == 0 {
			continue
		}

		out = append(out, KeyAbiIndex{
			Abi: abi.ABI,
			Experience: abi.Experience,
		})
	}

	return out
}

func (d *UserIndexData)GetKeyLocIndexes() []KeyLocIndex {
	out := make([]KeyLocIndex, 0)
	for _, loc := range d.Info.Locations {
		out = append(out, KeyLocIndex{
			Location: comm.Location{
				Longitude: loc.Longitude,
				Latitude:  loc.Latitude,
			},
		})
	}

	out = append(out, KeyLocIndex{
		Location: comm.Location{
			Longitude: d.Info.CurLocation.Longitude,
			Latitude:  d.Info.CurLocation.Latitude,
		},
	})

	return out
}

