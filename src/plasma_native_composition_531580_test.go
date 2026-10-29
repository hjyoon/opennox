package opennox

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type plasmaNativeFrame531580[T comparable] struct {
	Before [32]byte
	Value  T
	After  [32]byte
}

func plasmaNativeAllocation531580[T comparable](t *testing.T) *T {
	t.Helper()
	f, free := alloc.New(plasmaNativeFrame531580[T]{})
	*f = plasmaNativeFrame531580[T]{}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	t.Cleanup(func() {
		defer free()
		for i := range f.Before {
			if f.Before[i] != byte(0x51+i) || f.After[i] != byte(0xe7-i) {
				t.Error("native allocation canary")
			}
		}
	})
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&f.Value)) <= math.MaxUint32 {
		t.Fatal("native C-owned allocation below 4 GiB")
	}
	return &f.Value
}

// Synthetic C-owned duration/unit records exercise the real public C exports,
// binary64 balance loader, facing/distance/LOS, registered DefaultDamage and
// UnitSetHP, state/audio, important-packet sender, charge report and destruction.
// The read-only ray packet observer delegates to production allocation. No
// gameplay service is replaced, and no peer connection is opened.
func TestSpellPlasmaPublicCNativeComposition531580(t *testing.T) {
	const childEnv = "NOX_PLASMA_NATIVE_COMPOSITION_CHILD"
	mode := os.Getenv(childEnv)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"player-retained", "hecubah-retained", "ray-switch", "grace-expiry", "cursor-acquire", "no-target", "monster-moved", "cancel-flag", "player-no-weapon"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestSpellPlasmaPublicCNativeComposition531580$", "-test.v")
				cmd.Env = append(os.Environ(), childEnv+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual Plasma C composition: %v", err)
				}
			})
		}
		return
	}
	handles.Init()
	t.Cleanup(handles.Release)
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(oldFlags)
	t.Cleanup(func() { noxflags.UnsetGame(noxflags.GetGame()); noxflags.SetGame(oldFlags) })
	oldGameplay := noxflags.GetGamePlay()
	noxflags.SetGamePlay(noxflags.GameplayFlag1)
	t.Cleanup(func() { noxflags.UnsetGamePlay(noxflags.GetGamePlay()); noxflags.SetGamePlay(oldGameplay) })
	s.spells.duration.Init(s)
	s.SetFrame(140)
	s.SetTickRate(30)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.Init() == 0 {
		t.Fatal("native wall grid")
	}
	t.Cleanup(s.Walls.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	dir, oldData := t.TempDir(), datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	if err := os.WriteFile(filepath.Join(dir, "gamedata.yml"), []byte("PlasmaDamage: 1.99999999\nPlasmaDamageHecubah: 7.75\nPlasmaSearchTime: 1.99999999\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"NativePlasmaVictim", "Hecubah", "HecubahWithOrb"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: name, OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
			t.Fatal(err)
		}
	}
	if math.Float64bits(s.Balance.Float("PlasmaDamage")) != math.Float64bits(1.99999999) {
		t.Fatal("production binary64 balance")
	}
	rt := s.spells.duration.plasmaRuntime531580()
	oldHec, oldOrb := *rt.HecubahType, *rt.HecubahOrbType
	*rt.HecubahType, *rt.HecubahOrbType = 0, 0
	t.Cleanup(func() { *rt.HecubahType, *rt.HecubahOrbType = oldHec, oldOrb })
	if !s.Objs.Init(16) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	newObj := func(typ *server.ObjectType, cls object.Class, code uint32) *server.Object {
		o := s.Objs.NewObject(typ)
		o.ObjClass = cls
		o.ObjFlags = 0
		o.NetCode = code
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(o.CObj()) <= math.MaxUint32 {
			t.Fatal("native object below 4 GiB")
		}
		return o
	}
	source := newObj(&server.ObjectType{}, object.ClassPlayer, 0x1234)
	target := newObj(s.Types.ByID("NativePlasmaVictim"), object.ClassMonster, 0x2345)
	previous := newObj(&server.ObjectType{}, object.ClassMonster, 0x3456)
	wand := newObj(&server.ObjectType{}, object.ClassWeapon, 0x4567)
	update := plasmaNativeAllocation531580[server.PlayerUpdateData](t)
	info := plasmaNativeAllocation531580[server.Player](t)
	monster := plasmaNativeAllocation531580[server.MonsterUpdateData](t)
	sourceMonster := plasmaNativeAllocation531580[server.MonsterUpdateData](t)
	health := plasmaNativeAllocation531580[server.HealthData](t)
	data := plasmaNativeAllocation531580[server.WandUseData](t)
	r := plasmaNativeAllocation531580[server.DurSpell](t)
	*info = server.Player{PlayerInd: 7}
	*update = server.PlayerUpdateData{Player: info, EquippedWeapon: wand, State: server.PlayerState13}
	*health = server.HealthData{Cur: 200, Max: 200, Field2: 200}
	*data = server.WandUseData{Flags: 0x87654324, Charge: 3, MaxCharge: 4, Progress: 75}
	source.UpdateData = unsafe.Pointer(update)
	source.PosVec = types.Ptf(100, 200)
	source.PrevPos = types.Ptf(90, 190)
	source.Direction1 = 0
	target.UpdateData = unsafe.Pointer(monster)
	target.HealthData = health
	target.ObjSubClass = 0x10202
	target.PosVec = types.Ptf(200, 200)
	previous.PosVec = types.Ptf(220, 200)
	wand.ObjSubClass = 0x4000000
	wand.UseData.Ptr = unsafe.Pointer(data)
	*r = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_PLASMA), Level: 3, Obj12: source, Caster16: source, Obj24: previous, Pos: source.PosVec, Frame68: 0x87654321, Flags88: 0x77556601, Field76: 99}
	type sent struct {
		packet                              []byte
		recipient, remove, sequence, result int
	}
	var packets []sent
	productionSend := s.NetSendPacketXxx
	s.NetSendPacketXxx = func(recipient int, p []byte, related *server.Object, remove, sequence int) int {
		result := productionSend(recipient, p, related, remove, sequence)
		if len(p) == 7 && p[0] == 0x9e {
			packets = append(packets, sent{append([]byte(nil), p...), recipient, remove, sequence, result})
		}
		return result
	}
	t.Cleanup(func() { s.NetSendPacketXxx = productionSend })
	if mode == "player-no-weapon" {
		update.EquippedWeapon = nil
	}
	if mode == "monster-moved" {
		source.ObjClass = object.ClassMonster
		source.UpdateData = unsafe.Pointer(sourceMonster)
	}
	wantCreate := 0
	if mode == "player-no-weapon" {
		wantCreate = 1
	}
	if got := ccall.CallIntPtr(legacy.Get_nox_xxx_plasmaSmth_531580(), r.C()); got != wantCreate {
		t.Fatalf("actual C create=%d", got)
	}
	if r.Target48 != nil || r.Field76 != 0 || r.Field72 != 0 {
		t.Fatal("create record initialization")
	}
	if mode == "player-no-weapon" || mode == "monster-moved" {
		if rt.LoadWeapon(r) != nil {
			t.Fatal("nonwand caster retained weapon")
		}
	} else if rt.LoadWeapon(r) != wand || r.Flags88 != 0x77556603 {
		t.Fatal("native equipment/flags")
	}
	wantHP := uint16(200)
	wantCharge := uint8(3)
	wantRayTypes := []byte{}
	damage := uint16(1)
	r.Target48 = target
	switch mode {
	case "hecubah-retained":
		target.TypeInd = uint16(s.Types.IndByID("Hecubah"))
		damage = 7
	case "ray-switch":
		rt.StoreRayTarget(r, previous)
		wantRayTypes = append(wantRayTypes, 8)
	case "cursor-acquire":
		r.Target48 = nil
		update.CursorObj = target
		if !s.IsEnemyTo(source, target) {
			t.Fatal("production cursor enemy prerequisite")
		}
	case "grace-expiry":
		target.PosVec = types.Ptf(0, 200)
		r.Field76 = 1
		rt.StoreRayTarget(r, target)
		if legacy.Nox_server_testTwoPointsAndDirection_4E6E50(source.PosVec, int16(source.Direction1), target.PosVec)&2 == 0 {
			t.Fatal("production facing grace prerequisite")
		}
	case "no-target":
		r.Target48 = nil
		rt.StoreRayTarget(r, previous)
	case "monster-moved":
		r.Pos.X -= 10
	case "cancel-flag":
		r.Flags88 |= 0x20
	}
	switch mode {
	case "player-retained", "hecubah-retained", "ray-switch", "cursor-acquire":
		wantRayTypes = append(wantRayTypes, 1)
		for hit := 0; hit < 3; hit++ {
			want := 0
			if hit == 2 {
				want = 1
			}
			if got := ccall.CallIntPtr(legacy.Get_nox_xxx_plasmaShot_531600(), r.C()); got != want {
				t.Fatalf("actual C hit%d return%d want%d", hit, got, want)
			}
			wantHP -= damage
			wantCharge--
			if health.Cur != wantHP || data.Charge != wantCharge || data.Progress != 100*uint32(wantCharge)/4 || rt.LoadRayTarget(r) != target || r.Frame68 != 141 || update.State != 22 {
				t.Fatalf("actual damage/charge/ray/state: HP%d charge%d progress%d ray%p expiry%d state%d", health.Cur, data.Charge, data.Progress, rt.LoadRayTarget(r), r.Frame68, update.State)
			}
			if target.Obj130 != source || target.Field131 != 14 || target.Frame134 != 140 || monster.Field546 != 14 || monster.Field547 != 2 || !monster.StatusFlags.Has(object.MonStatusOnFire|object.MonStatusInjured) {
				t.Fatal("registered native Plasma damage attribution/monster state")
			}
		}
	case "grace-expiry", "no-target":
		if got := ccall.CallIntPtr(legacy.Get_nox_xxx_plasmaShot_531600(), r.C()); got != 0 || r.Target48 != nil || rt.LoadRayTarget(r) != nil {
			t.Fatal("actual no-target stop")
		}
		wantRayTypes = []byte{8}
	case "monster-moved", "cancel-flag":
		if got := ccall.CallIntPtr(legacy.Get_nox_xxx_plasmaShot_531600(), r.C()); got != 1 {
			t.Fatal("actual cancellation gate")
		}
	case "player-no-weapon":
	default:
		t.Fatal("unknown mode")
	}
	if health.Cur != wantHP || data.Charge != wantCharge {
		t.Fatal("unexpected damage/charge")
	}
	if len(packets) != len(wantRayTypes) {
		t.Fatalf("actual ray packets %d want%d: %v", len(packets), len(wantRayTypes), packets)
	}
	for i, p := range packets {
		first, second := uint16(source.NetCode), uint16(target.NetCode)
		if wantRayTypes[i] == 8 {
			first, second = uint16(target.NetCode), uint16(source.NetCode)
			if mode == "ray-switch" || mode == "no-target" {
				first = uint16(previous.NetCode)
			}
		}
		want := []byte{0x9e, wantRayTypes[i], byte(source.Direction1), 0, 0, 0, 0}
		binary.LittleEndian.PutUint16(want[3:5], first)
		binary.LittleEndian.PutUint16(want[5:7], second)
		if !bytes.Equal(p.packet, want) || p.recipient != 255 || p.remove != 1 || p.sequence != 1 || p.result != 1 {
			t.Fatalf("actual production C packet allocation/payload: %+v want%x", p, want)
		}
	}
	before := *r
	ccall.CallVoidPtr(legacy.Get_sub_5319E0(), r.C())
	if *r != before || rt.LoadWeapon(r) != nil {
		t.Fatal("destroy changed native record or retained dying weapon sidecar")
	}
	wantFlags := uint32(0x87654320)
	if mode == "player-no-weapon" || mode == "monster-moved" {
		wantFlags |= 4
	}
	if data.Flags != wantFlags || data.Charge != wantCharge || data.MaxCharge != 4 || r.Obj12 != source || r.Obj24 != previous || update.Player != info || wand.UseData.Ptr != unsafe.Pointer(data) {
		t.Fatal("destroy flags/charge/native links")
	}
	t.Logf("actual C Plasma %s record=%p caster=%p target=%p wand=%p HP200->%d charge3->%d rays%d native lifetime cleanup verified", mode, r, source, target, wand, health.Cur, data.Charge, len(packets))
}
