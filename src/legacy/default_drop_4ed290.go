package legacy

import (
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func defaultDropRuntime4ED290(outer Server) server.DefaultDropRuntime4ED290 {
	s := outer.S()
	buffRuntime := flagPickupBuffPurgeRuntime4EA7A0(s)
	return server.DefaultDropRuntime4ED290{
		ItemIsDroppable: s.ItemIsDroppable53EBF0,
		ItemDropMask:    s.ItemDropMask53EC80,
		DetachInventory: Sub_4ED0C0,
		CreateAt: func(obj, owner *server.Object, point types.Pointf) {
			outer.CreateObjectAt(obj, owner, point)
		},
		DelayedDelete: outer.DelayedDelete,
		TeamFlagStatus: func(teamID, status, material uint8, carrier uint16) int32 {
			return Sub_4E82C0(teamID, status, material, carrier)
		},
		GameFlag: func(flag uint32) uint32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		SetDecayTime: func(obj *server.Object, frames uint32) {
			// 00511660 remains the sole ABI32 dependency in the decay scheduler;
			// preserve the original signed C-int bit pattern at this boundary.
			Nox_xxx_unitSetDecayTime_511660(obj, int(int32(frames)))
		},
		BuffOff: func(obj *server.Object, enchant uint32) {
			buffRuntime.BuffOff(obj, server.EnchantID(enchant))
		},
	}
}

func defaultDropCall4ED290(
	owner, item *server.Object,
	point *types.Pointf,
) int32 {
	outer := GetServer()
	return outer.S().DefaultDrop4ED290(owner, item, point, defaultDropRuntime4ED290(outer))
}
