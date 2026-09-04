package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

type monsterMainLegacyServer547210 struct {
	Server
	srv *server.Server
}

func (s *monsterMainLegacyServer547210) S() *server.Server {
	return s.srv
}

func TestMonsterMainExport547210RejectsPE32FallbackOn64Bit(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width fallback regression applies to 64-bit builds")
	}

	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldGetServer := GetServer
	GetServer = func() Server { return &monsterMainLegacyServer547210{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	unit := &server.Object{
		ObjClass: object.ClassMonster,
		ObjFlags: object.FlagEnabled,
	}
	update := &server.MonsterUpdateData{AIStackInd: 0}
	update.AIStack[0].Action = uint32(ai.ACTION_FIGHT)
	unit.UpdateData = unsafe.Pointer(update)

	var pin runtime.Pinner
	pin.Pin(unit)
	pin.Pin(update)
	defer pin.Unpin()
	for name, pointer := range map[string]uintptr{
		"unit":   uintptr(unsafe.Pointer(unit)),
		"update": uintptr(unsafe.Pointer(update)),
	} {
		if pointer <= math.MaxUint32 {
			t.Fatalf("%s pointer = %#x, want address above the ABI32 range", name, pointer)
		}
	}

	beforeUnit := *unit
	beforeUpdate := *update
	Nox_xxx_monsterMainAIFn_547210(unit)
	if *unit != beforeUnit || *update != beforeUpdate {
		t.Fatal("unsupported native monster-main state was mutated")
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
}
