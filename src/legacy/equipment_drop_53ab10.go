package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func equipmentDropServerSubFlag53AB10(flag uint32) int32 {
	if Nox_xxx_getServerSubFlags_409E60()&flag != 0 {
		return 1
	}
	return 0
}

var weaponDropCall53AB10 = func(
	owner, item *server.Object,
	point *types.Pointf,
) int32 {
	outer := GetServer()
	return outer.S().WeaponDrop53AB10(owner, item, point, server.WeaponDropRuntime53AB10{
		DefaultDrop:   defaultDropCall4ED290,
		ServerSubFlag: equipmentDropServerSubFlag53AB10,
	})
}

var armorDropCall53EB70 = func(
	owner, item *server.Object,
	point *types.Pointf,
) int32 {
	outer := GetServer()
	return outer.S().ArmorDrop53EB70(owner, item, point, server.ArmorDropRuntime53EB70{
		DefaultDrop:   defaultDropCall4ED290,
		ServerSubFlag: equipmentDropServerSubFlag53AB10,
	})
}
