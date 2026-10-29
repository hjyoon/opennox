package opennox

import (
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) SpellFirewalkUpdate52ED40(record *server.DurSpell) int32 {
	world := s.S()
	return world.SpellFirewalkUpdate52ED40(record, server.SpellFirewalkRuntime52ED40{
		RoundTowardZero: true,
		LoadTypeID: func(i int) uint32 {
			return *memmap.PtrUint32(0x5D4594, 2487888+4*uintptr(i))
		},
		StoreTypeID: func(i int, id uint32) {
			*memmap.PtrUint32(0x5D4594, 2487888+4*uintptr(i)) = id
		},
		LookupType: func(name string) uint32 { return uint32(world.Types.IndByID(name)) },
		NewObject:  func(id uint32) *server.Object { return world.NewObjectByTypeInd(int(int32(id))) },
		CreateAt: func(flame *server.Object, position types.Pointf) {
			s.CreateObjectAt(flame, nil, position)
		},
		Audio:    func(position types.Pointf) { s.Audio.EventPos(sound.ID(46), position, 0, 0) },
		Decay:    func(flame *server.Object, delay uint32) { world.DecaySetTime511660(flame, delay) },
		TickRate: world.TickRate,
	})
}
