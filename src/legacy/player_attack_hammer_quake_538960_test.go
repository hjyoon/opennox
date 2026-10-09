package legacy

import (
	"bytes"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerAttackExport538960NPCWarHammerKeepsQuakePlayersNativeWidth(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width attack routing applies to 64-bit builds")
	}
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.Map.Init()
	t.Cleanup(srv.Map.Free)
	srv.NetList.Init()
	t.Cleanup(srv.NetList.Free)
	srv.SetFrame(2)
	bridge := &playerAttackLegacyServer538960{srv: srv}
	oldGetServer := GetServer
	GetServer = func() Server { return bridge }
	t.Cleanup(func() { GetServer = oldGetServer })
	oldAnimation := playerAnimFrames4F9F90
	playerAnimFrames4F9F90 = func(action int) (int, int) {
		if action != 39 {
			t.Fatalf("NPC hammer animation = %d, want 39", action)
		}
		return 4, 0
	}
	t.Cleanup(func() { playerAnimFrames4F9F90 = oldAnimation })
	directionX := memmap.PtrFloat32(0x587000, 194136)
	directionY := memmap.PtrFloat32(0x587000, 194140)
	oldX, oldY := *directionX, *directionY
	*directionX, *directionY = 1, 0
	t.Cleanup(func() { *directionX, *directionY = oldX, oldY })

	unit := &server.Object{
		ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterNPC),
		PosVec: types.Ptf(321, 654), NewPos: types.Ptf(321, 654),
	}
	unit.Shape.Kind = server.ShapeKindCircle
	unit.Shape.Circle.R = 5
	unit.Shape.Circle.R2 = 25
	update := &server.MonsterUpdateData{
		Field331: 37, Field481: 0x55667701,
		WeaponEquipFlags: 0x4000, Field516: 0xf6b88ee0, Field517: 0x99aabbcc,
	}
	weapon := &server.Object{
		TypeInd: 0x3212, ObjClass: object.ClassWeapon, ObjSubClass: 0x4000,
		ObjFlags: object.FlagEquipped, InvHolder: unit,
	}
	modifier := &server.Modifier{
		TypeInd: uint32(weapon.TypeInd), ReqStrength60: 20,
		DamageCoeffOrArmor64: 1.5, Range68: 40, DamageMin72: 10,
	}
	unit.UpdateData, unit.InvFirstItem = unsafe.Pointer(update), weapon
	srv.Modif.Dword_5d4594_251600 = modifier
	var pin runtime.Pinner
	defer pin.Unpin()
	pinPlayerAttackProjectilePointers538960(t, &pin,
		unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(weapon), unsafe.Pointer(modifier))

	// The real C quake callee must follow all three native pointers. Cached
	// camera position, not the unit position, determines the packet recipients.
	positions := []types.Pointf{
		unit.PosVec,
		types.Ptf(471, 654),
		types.Ptf(math.Nextafter32(621, float32(math.Inf(-1))), 654),
		types.Ptf(621, 654),
		types.Ptf(622, 654),
	}
	players := make([]*server.Player, len(positions))
	for i, pos := range positions {
		player := srv.Players.NewRaw(2001 + i)
		if player == nil {
			t.Fatal("cannot allocate quake recipient")
		}
		playerUpdate, freeUpdate := alloc.New(server.PlayerUpdateData{})
		t.Cleanup(freeUpdate)
		playerUnit, freeUnit := alloc.New(server.Object{})
		t.Cleanup(freeUnit)
		*playerUnit = server.Object{
			ObjClass: object.ClassPlayer, PosVec: types.Ptf(5000, 5000),
			UpdateData: unsafe.Pointer(playerUpdate),
		}
		playerUpdate.Player = player
		player.PlayerUnit = playerUnit
		player.Pos3632Vec = pos
		players[i] = player
		if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
			for _, pointer := range []unsafe.Pointer{
				unsafe.Pointer(playerUnit), unsafe.Pointer(playerUpdate), unsafe.Pointer(player),
			} {
				if uintptr(pointer) <= math.MaxUint32 {
					t.Fatalf("C-owned quake pointer = %p, want above ABI32", pointer)
				}
			}
		}
	}

	for _, tc := range []struct {
		name     string
		strength uint32
		want     [5]int
	}{
		// Multiples of ten isolate quake reporting from the caller's separate
		// strength-to-magnitude conversion. Original 004D915A loads a binary32
		// inverse of 90000: at distance 150, magnitude four truncates to TWO,
		// not three. Zero magnitude still emits a packet inside the boundary.
		{"magnitude four", 40, [5]int{4, 2, 0, -1, -1}},
		{"zero magnitude", 0, [5]int{0, 0, 0, -1, -1}},
		{"magnitude two", 20, [5]int{2, 1, 0, -1, -1}},
		{"magnitude twenty", 200, [5]int{20, 14, 0, -1, -1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			update.Field331, update.Field481 = tc.strength, 0x55667701
			update.Field120_1 = 1
			if got := playerAttackNativeEntry538960(unit); got != 1 {
				t.Fatalf("NPC hammer attack = %d, want active middle frame", got)
			}
			if update.Field120_1 != 2 || update.Field481 != 0x55667701 || update.Field517 != 0x99aabbcc {
				t.Fatalf("NPC hammer state = %d/%#x/%#x", update.Field120_1, update.Field481, update.Field517)
			}
			for i, player := range players {
				var want []byte
				if tc.want[i] >= 0 {
					want = []byte{0x97, byte(tc.want[i])}
				}
				got := srv.NetList.CopyPacketsA(ntype.PlayerInd(player.PlayerInd), netlist.Kind1)
				if !bytes.Equal(got, want) {
					t.Errorf("camera %v quake packet = % x, want % x", positions[i], got, want)
				}
			}
			// A repeated middle-frame update must not repeat either the area
			// attack or its quake packets.
			walls := bridge.wallDamageCalls
			if got := playerAttackNativeEntry538960(unit); got != 1 || bridge.wallDamageCalls != walls {
				t.Fatalf("repeated hammer frame = result:%d walls:%d/%d", got, bridge.wallDamageCalls, walls)
			}
			for _, player := range players {
				if got := srv.NetList.CopyPacketsA(ntype.PlayerInd(player.PlayerInd), netlist.Kind1); len(got) != 0 {
					t.Fatalf("repeated hammer frame emitted quake: % x", got)
				}
			}
		})
	}
	if bridge.wallDamageCalls != 4 || bridge.wallDamageAttacker != weapon {
		t.Fatalf("hammer wall calls = %d/%p, want 4/%p", bridge.wallDamageCalls, bridge.wallDamageAttacker, weapon)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(players)
}

func TestPlayerAttackExport538960WarHammerRoundsAllStrengthBytes(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width attack routing applies to 64-bit builds")
	}
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.Map.Init()
	t.Cleanup(srv.Map.Free)
	srv.NetList.Init()
	t.Cleanup(srv.NetList.Free)
	srv.SetFrame(2)
	bridge := &playerAttackLegacyServer538960{srv: srv}
	oldGetServer := GetServer
	GetServer = func() Server { return bridge }
	t.Cleanup(func() { GetServer = oldGetServer })
	oldAnimation := playerAnimFrames4F9F90
	playerAnimFrames4F9F90 = func(action int) (int, int) {
		if action != 39 {
			t.Fatalf("hammer animation = %d, want 39", action)
		}
		return 4, 0
	}
	t.Cleanup(func() { playerAnimFrames4F9F90 = oldAnimation })
	unit := &server.Object{
		ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterNPC),
		PosVec: types.Ptf(321, 654),
	}
	update := &server.MonsterUpdateData{WeaponEquipFlags: uint32(object.WeaponHammer)}
	weapon := &server.Object{
		TypeInd: 0x3213, ObjClass: object.ClassWeapon,
		ObjSubClass: object.SubClass(object.WeaponHammer),
		ObjFlags:    object.FlagEquipped, InvHolder: unit,
	}
	modifier := &server.Modifier{TypeInd: uint32(weapon.TypeInd), Range68: 40}
	unit.UpdateData, unit.InvFirstItem = unsafe.Pointer(update), weapon
	srv.Modif.Dword_5d4594_251600 = modifier
	var pin runtime.Pinner
	defer pin.Unpin()
	pinPlayerAttackProjectilePointers538960(t, &pin,
		unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(weapon), unsafe.Pointer(modifier))
	player := srv.Players.NewRaw(2101)
	if player == nil {
		t.Fatal("cannot allocate quake recipient")
	}
	playerUpdate, freeUpdate := alloc.New(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	playerUnit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	playerUnit.ObjClass, playerUnit.UpdateData = object.ClassPlayer, unsafe.Pointer(playerUpdate)
	playerUpdate.Player, player.PlayerUnit = player, playerUnit
	player.Pos3632Vec = unit.PosVec
	for strength := 0; strength <= math.MaxUint8; strength++ {
		update.Field331, update.Field481 = uint32(strength), 0x55667701
		update.Field120_1 = 1
		if got := playerAttackNativeEntry538960(unit); got != 1 {
			t.Fatalf("strength %d attack = %d, want active middle frame", strength, got)
		}
		// Original 0053976E FMULS -> FSTPS -> 00419A70 FISTPL. The
		// multiplication is spilled as binary32 before default ties-to-even;
		// the expected value does not call any production conversion helper.
		scaled := float32(strength) * float32(0.1)
		want := []byte{0x97, byte(math.RoundToEven(float64(scaled)))}
		got := srv.NetList.CopyPacketsA(ntype.PlayerInd(player.PlayerInd), netlist.Kind1)
		if !bytes.Equal(got, want) {
			t.Errorf("strength %d quake packet = % x, want % x", strength, got, want)
		}
		if update.Field120_1 != 2 || update.Field481 != 0x55667701 {
			t.Fatalf("strength %d frame/unrelated = %d/%#x", strength, update.Field120_1, update.Field481)
		}
	}
	if bridge.wallDamageCalls != 256 || bridge.wallDamageAttacker != weapon {
		t.Fatalf("hammer wall calls = %d/%p, want 256/%p", bridge.wallDamageCalls, bridge.wallDamageAttacker, weapon)
	}
	runtime.KeepAlive(unit)
}
