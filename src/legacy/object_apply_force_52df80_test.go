package legacy

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestApplyForceExport52DF80MovableGateBeforeOrigin(t *testing.T) {
	old := GetServer
	GetServer = func() Server { t.Fatal("blocked force consulted the server before returning"); return nil }
	t.Cleanup(func() { GetServer = old })
	for _, flag := range []object.Flags{8, 32, 64, 0x8000, 0x8068} {
		t.Run(fmt.Sprintf("flags-%x", uint32(flag)), func(t *testing.T) {
			unit := selfWeaponElectricNativeAlloc(t, server.Object{ObjFlags: flag})
			before := *unit
			applyForceExport52DF80(unit, nil, math.MaxFloat32)
			if *unit != before {
				t.Fatal("blocked force changed object")
			}
		})
	}
	for _, class := range []object.Class{0x400000, 0x400001, 0x400004} {
		t.Run(fmt.Sprintf("class-%x", uint32(class)), func(t *testing.T) {
			unit := selfWeaponElectricNativeAlloc(t, server.Object{ObjClass: class})
			before := *unit
			applyForceExport52DF80(unit, nil, math.MaxFloat32)
			if *unit != before {
				t.Fatal("immobile force changed object")
			}
		})
	}
}
