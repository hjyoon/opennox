package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

type playerAttackLegacyServer538960 struct {
	Server
	srv *server.Server
}

func (s *playerAttackLegacyServer538960) S() *server.Server {
	return s.srv
}

func TestPlayerAttackExport538960KeepsNativePointerWidth(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	// Keep this object in Go-managed memory. On 64-bit Linux the legacy C heap
	// may sit below 4 GiB because the test binary is non-PIE, while the Go heap
	// still exercises the native high half on every supported 64-bit host. All
	// pointer fields remain nil, so passing it for the duration of this CGo call
	// does not retain a Go pointer in C memory.
	unit := new(server.Object)
	if pointer := uintptr(unsafe.Pointer(unit)); pointer <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want address above the ABI32 range", unit)
	}

	// An object without player update data is deliberately used here. The
	// native entry rejects it safely; the decompiled ABI32 body would truncate
	// the pointer before reading the original +748 update-data slot.
	if got := Nox_xxx_playerAttack_538960(unit); got != 0 {
		t.Fatalf("player attack result = %d, want 0 for missing update data", got)
	}
	runtime.KeepAlive(unit)
}

func TestPlayerAttackExport538960KeepsNestedPlayerPointersNativeWidth(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldGetServer := GetServer
	GetServer = func() Server { return &playerAttackLegacyServer538960{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	unit := &server.Object{ObjClass: object.ClassPlayer}
	update := &server.PlayerUpdateData{}
	player := &server.Player{Field8: 23}
	weapon := &server.Object{TypeInd: math.MaxUint16}
	unit.UpdateData = unsafe.Pointer(update)
	update.Player = player
	update.EquippedWeapon = weapon
	player.Info().SetField2239(37)

	// Pin the complete graph for the duration of the C call. The regression
	// deliberately uses the Go heap because it reliably resides above 4 GiB
	// even when a Linux C allocator serves the legacy heap below that boundary.
	var pin runtime.Pinner
	pin.Pin(unit)
	pin.Pin(update)
	pin.Pin(player)
	pin.Pin(weapon)
	defer pin.Unpin()
	for name, pointer := range map[string]uintptr{
		"unit":   uintptr(unsafe.Pointer(unit)),
		"update": uintptr(unsafe.Pointer(update)),
		"player": uintptr(unsafe.Pointer(player)),
		"weapon": uintptr(unsafe.Pointer(weapon)),
	} {
		if pointer <= math.MaxUint32 {
			t.Fatalf("%s pointer = %#x, want address above the ABI32 range", name, pointer)
		}
	}

	// A deliberately unknown weapon type exits immediately after the native C
	// body has loaded every nested pointer and the Go Strength callback has
	// followed unit -> update -> player. Any PE32 truncation in that path faults
	// before this zero result can be returned.
	if got := Nox_xxx_playerAttack_538960(unit); got != 0 {
		t.Fatalf("player attack result = %d, want 0 for an unknown weapon type", got)
	}
	if got := playerAttackNativeEntry538960(unit); got != 0 {
		t.Fatalf("C player attack entry result = %d, want 0 for an unknown weapon type", got)
	}
	if got := player.Info().Field2239(); got != 37 {
		t.Fatalf("player strength = %d, want 37 after native traversal", got)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
	runtime.KeepAlive(weapon)
}
