package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

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

func TestPlayerAttackExport538960RestoresWarcryWithNativePointers(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.Map.Init()
	t.Cleanup(srv.Map.Free)
	srv.SetFrame(3)
	oldGetServer := GetServer
	GetServer = func() Server { return &playerAttackLegacyServer538960{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	unit := &server.Object{
		ObjClass: object.ClassPlayer,
		PosVec:   types.Ptf(400, 500),
		NewPos:   types.Ptf(400, 500),
		Field34:  0,
	}
	update := &server.PlayerUpdateData{Field59_0: 2}
	player := &server.Player{Field8: 23}
	unit.UpdateData = unsafe.Pointer(update)
	update.Player = player
	record := &server.ExecAbilityClass{
		Unit: unit, Abil: server.AbilityWarcry, Active: 1,
	}
	srv.Abils.SetExecHead(record)

	var pin runtime.Pinner
	pin.Pin(unit)
	pin.Pin(update)
	pin.Pin(player)
	defer pin.Unpin()
	for name, pointer := range map[string]uintptr{
		"unit": uintptr(unsafe.Pointer(unit)), "update": uintptr(unsafe.Pointer(update)),
		"player": uintptr(unsafe.Pointer(player)),
	} {
		if pointer <= math.MaxUint32 {
			t.Fatalf("%s pointer = %#x, want address above the ABI32 range", name, pointer)
		}
	}

	oldCounterSpell := Nox_xxx_castCounterSpell_52BBB0
	t.Cleanup(func() { Nox_xxx_castCounterSpell_52BBB0 = oldCounterSpell })
	var calls int
	var gotSpell spell.ID
	var gotUnit [3]*server.Object
	Nox_xxx_castCounterSpell_52BBB0 = func(
		id spell.ID, a2, a3, a4 *server.Object, _ *server.SpellAcceptArg, _ int,
	) int {
		calls++
		gotSpell = id
		gotUnit = [3]*server.Object{a2, a3, a4}
		return 1
	}

	// GAME.EXE fires Warcry exactly on the 2 -> 3 animation transition. The
	// empty animation table intentionally yields a zero frame count after the
	// effect, which also exercises the original deactivation and frame clamp.
	if got := playerAttackNativeEntry538960(unit); got != 0 {
		t.Fatalf("warcry attack result = %d, want completed animation", got)
	}
	if calls != 1 || gotSpell != spell.SPELL_COUNTERSPELL ||
		gotUnit != [3]*server.Object{unit, unit, unit} {
		t.Fatalf("counterspell = calls:%d spell:%d units:%p/%p/%p, want 1/%d/%p/%p/%p",
			calls, gotSpell, gotUnit[0], gotUnit[1], gotUnit[2],
			spell.SPELL_COUNTERSPELL, unit, unit, unit)
	}
	if record.Active != 0 {
		t.Fatalf("warcry active value = %d, want deactivated", record.Active)
	}
	if update.Field59_0 != math.MaxUint8 {
		t.Fatalf("warcry stored frame = %d, want original zero-count clamp 255", update.Field59_0)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestPlayerAttackExport538960RestoresBerserkWithNativePointers(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.SetFrame(7)
	oldGetServer := GetServer
	GetServer = func() Server { return &playerAttackLegacyServer538960{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	unit := &server.Object{
		ObjClass:   object.ClassPlayer,
		Field34:    0,
		Direction1: 64,
		ForceVec:   types.Ptf(1, -2),
		SpeedCur:   -1,
		SpeedBase:  2.5,
	}
	update := &server.PlayerUpdateData{Field59_0: 6}
	player := &server.Player{Field8: 23}
	unit.UpdateData = unsafe.Pointer(update)
	update.Player = player
	record := &server.ExecAbilityClass{
		Unit: unit, Abil: server.AbilityBerserk, Active: 1,
	}
	srv.Abils.SetExecHead(record)

	var pin runtime.Pinner
	pin.Pin(unit)
	pin.Pin(update)
	pin.Pin(player)
	defer pin.Unpin()
	if pointer := uintptr(unsafe.Pointer(unit)); pointer <= math.MaxUint32 {
		t.Fatalf("unit pointer = %#x, want address above the ABI32 range", pointer)
	}

	cosine, sine := server.SinCosDir(byte(unit.Direction1))
	wantSpeed := float32(15)
	wantForce := types.Ptf(1+wantSpeed*cosine, -2+wantSpeed*sine)
	if got := playerAttackNativeEntry538960(unit); got != 0 {
		t.Fatalf("berserk attack result = %d, want completed zero-count animation", got)
	}
	if unit.SpeedCur != wantSpeed || unit.ForceVec != wantForce {
		t.Fatalf("berserk motion = speed:%g force:%+v, want %g/%+v",
			unit.SpeedCur, unit.ForceVec, wantSpeed, wantForce)
	}
	if update.Field59_0 != math.MaxUint8 || record.Active != 1 {
		t.Fatalf("berserk state = frame:%d active:%d, want 255/1", update.Field59_0, record.Active)
	}

	unit.Buffs = uint32(1) << server.ENCHANT_HELD
	unit.SpeedCur = -7
	unit.ForceVec = types.Ptf(9, 11)
	update.Field59_0 = 8
	if got := playerAttackNativeEntry538960(unit); got != 0 {
		t.Fatalf("held berserk attack result = %d, want blocked", got)
	}
	if unit.SpeedCur != -7 || unit.ForceVec != types.Ptf(9, 11) || update.Field59_0 != 8 {
		t.Fatalf("held berserk mutated state = speed:%g force:%+v frame:%d",
			unit.SpeedCur, unit.ForceVec, update.Field59_0)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestPlayerAttackWarcryStunnable538960MatchesOriginalMask(t *testing.T) {
	eligible := &server.Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(object.MonsterWarcryStun),
	}
	if !playerAttackWarcryStunnable538960(eligible) {
		t.Fatal("eligible Warcry monster was rejected")
	}
	for name, target := range map[string]*server.Object{
		"nil":            nil,
		"player":         {ObjClass: object.ClassPlayer, ObjSubClass: eligible.ObjSubClass},
		"wrong-subclass": {ObjClass: object.ClassMonster},
		"dead":           {ObjClass: eligible.ObjClass, ObjSubClass: eligible.ObjSubClass, ObjFlags: object.FlagDead},
		"destroyed":      {ObjClass: eligible.ObjClass, ObjSubClass: eligible.ObjSubClass, ObjFlags: object.FlagDestroyed},
		"dead-destroyed": {ObjClass: eligible.ObjClass, ObjSubClass: eligible.ObjSubClass, ObjFlags: object.FlagDead | object.FlagDestroyed},
	} {
		if playerAttackWarcryStunnable538960(target) {
			t.Fatalf("%s target passed the original Warcry mask", name)
		}
	}
}
