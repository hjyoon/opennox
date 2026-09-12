package opennox

import (
	"github.com/opennox/opennox/v1/server"
	"github.com/opennox/opennox/v1/server/noxscript"
)

// nsSetRoamFlag515C40 follows GAME.EXE 00515C40's pop order: flag, object.
func nsSetRoamFlag515C40(vm noxscript.VM) int {
	s := vm.(*noxScript)
	flag := uint8(s.PopI32())
	obj := s.PopObject()
	obj.SetRoamFlag515C80(flag)
	return 0
}

// nsGroupSetRoamFlag515CB0 follows GAME.EXE 00515CB0's pop order: flag,
// group. The callback receives native-width object pointers from 00502670.
func nsGroupSetRoamFlag515CB0(vm noxscript.VM) int {
	s := vm.(*noxScript)
	flag := uint8(s.PopI32())
	group := s.PopGroup()
	s.s.Server.ForEachGroup502670(group, server.MapGroupObjects, func(value any) {
		value.(*server.Object).SetRoamFlag515C80(flag)
	})
	return 0
}
