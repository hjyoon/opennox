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
	playerlib "github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Actual public C update composes with production balance/table, spatial
// index, enemy/team/trace, native mana/protection, sync/audio and C packet
// allocation. The read-only ray observer delegates to the production sender.
// Synthetic units, gameplay flag1 and private balance are controlled; no peer
// or campaign progression is claimed. Common owner cleanup is checked by E2E.
func TestSpellDrainManaPublicCNativeComposition52E210(t *testing.T) {
	const childEnv = "NOX_DRAIN_MANA_NATIVE_COMPOSITION_CHILD"
	mode := os.Getenv(childEnv)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"obelisk-debit", "quest-obelisk", "player-exhaustion", "monster-one", "blocked-team", "matched-team", "no-target-stop", "recent-guard", "moved-monster", "glyph-nil-player", "full-dword", "precision-balance"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestSpellDrainManaPublicCNativeComposition52E210$", "-test.v")
				cmd.Env = append(os.Environ(), childEnv+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual Drain Mana C composition: %v", err)
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
	oldFlags, oldGameplay := noxflags.GetGame(), noxflags.GetGamePlay()
	noxflags.UnsetGame(oldFlags)
	noxflags.SetGamePlay(noxflags.GameplayFlag1)
	if mode == "quest-obelisk" {
		noxflags.SetGame(noxflags.GameModeQuest)
	}
	t.Cleanup(func() {
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.SetGame(oldFlags)
		noxflags.UnsetGamePlay(noxflags.GetGamePlay())
		noxflags.SetGamePlay(oldGameplay)
	})
	s.spells.duration.Init(s)
	t.Cleanup(s.spells.duration.Free)
	s.SetFrame(150)
	s.SetTickRate(30)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.Init() == 0 {
		t.Fatal("native walls")
	}
	t.Cleanup(s.Walls.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	dir, oldData := t.TempDir(), datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	coefficient := "1.5"
	if mode == "full-dword" {
		coefficient = "65536"
	}
	if mode == "precision-balance" {
		coefficient = "1.99999999"
	}
	if err := os.WriteFile(filepath.Join(dir, "gamedata.yml"), []byte("ManaDrainRange: 240\nManaDrainCoeff: ["+coefficient+", "+coefficient+", "+coefficient+"]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	if s.Balance.Float("ManaDrainRange") != 240 {
		t.Fatal("production scalar balance")
	}
	if err := s.Types.ReadObjectType(&things.Thing{Name: "NativeDrainManaTarget"}); err != nil {
		t.Fatal(err)
	}
	if !s.Objs.Init(16) {
		t.Fatal("native objects")
	}
	t.Cleanup(s.Objs.FreeObjects)
	newObj := func(cls object.Class, code uint32) *server.Object {
		o := s.Objs.NewObject(s.Types.ByID("NativeDrainManaTarget"))
		o.ObjClass, o.ObjFlags, o.NetCode = cls, 0, code
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(o.CObj()) <= math.MaxUint32 {
			t.Fatal("C object below4GiB")
		}
		return o
	}
	source := newObj(object.ClassPlayer, 0x1234)
	target := newObj(object.ClassImmobile, 0x2345)
	target.Extent = 0x4567
	previous := newObj(object.ClassMonster, 0x3456)
	update := plasmaNativeAllocation531580[server.PlayerUpdateData](t)
	targetUpdate := plasmaNativeAllocation531580[server.PlayerUpdateData](t)
	info := plasmaNativeAllocation531580[server.Player](t)
	targetInfo := plasmaNativeAllocation531580[server.Player](t)
	monster := plasmaNativeAllocation531580[server.MonsterUpdateData](t)
	sourceMonster := plasmaNativeAllocation531580[server.MonsterUpdateData](t)
	obelisk := plasmaNativeAllocation531580[server.ObeliskUpdateData](t)
	health := plasmaNativeAllocation531580[server.HealthData](t)
	r := plasmaNativeAllocation531580[server.DurSpell](t)
	*info, *targetInfo = server.Player{PlayerInd: 7}, server.Player{PlayerInd: 8}
	info.Info().SetPlayerClass(playerlib.Warrior)
	targetInfo.Info().SetPlayerClass(playerlib.Wizard)
	*update = server.PlayerUpdateData{Player: info, ManaCur: 3, ManaPrev: 3, ManaMax: 100, State: server.PlayerState13}
	*targetUpdate = server.PlayerUpdateData{Player: targetInfo, ManaCur: 5, ManaPrev: 5, ManaMax: 10, State: server.PlayerState13}
	*obelisk = server.ObeliskUpdateData{Mana: 20}
	s.Players.ClassStatsMult(playerlib.Warrior).Mana = 1.5
	source.UpdateData, source.PosVec, source.PrevPos = unsafe.Pointer(update), types.Ptf(100, 200), types.Ptf(90, 190)
	target.UpdateData, target.ObjSubClass, target.PosVec = unsafe.Pointer(obelisk), object.SubClass(8), types.Ptf(200, 200)
	previous.PosVec = types.Ptf(220, 200)
	if mode == "player-exhaustion" || mode == "glyph-nil-player" {
		target.ObjClass = object.ClassPlayer
		target.ObjSubClass = 0
		target.UpdateData = unsafe.Pointer(targetUpdate)
	}
	if mode == "monster-one" {
		target.ObjClass = object.ClassMonster
		target.ObjSubClass = 0x10202
		target.UpdateData = unsafe.Pointer(monster)
		monster.StatusFlags = object.MonsterStatus(0x20)
	}
	if mode == "moved-monster" {
		source.ObjClass = object.ClassMonster
		source.UpdateData = unsafe.Pointer(sourceMonster)
	}
	if mode == "blocked-team" {
		source.TeamVal.ID, target.TeamVal.ID = 1, 2
	}
	if mode == "matched-team" {
		source.TeamVal.ID, target.TeamVal.ID = 1, 1
	}
	if mode == "full-dword" {
		obelisk.Mana = 100000
	}
	*r = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_DRAIN_MANA), Level: 3, Obj12: source, Caster16: source, Obj24: previous, Target48: previous, Pos: source.PosVec, Frame60: 150, Frame68: 0x87654321, Field72: int32(math.Float32bits(.25)), Flags88: 0x77556601}
	if mode == "full-dword" || mode == "precision-balance" {
		r.Field72 = 0
	}
	if mode == "recent-guard" {
		*health = server.HealthData{Cur: 200, Max: 200}
		source.HealthData = health
		source.Frame134 = 150
	}
	if mode == "moved-monster" {
		r.Pos.X -= 10
	}
	if mode == "glyph-nil-player" {
		r.Caster16 = nil
		r.Flag20 = 1
	}
	if mode != "no-target-stop" && mode != "recent-guard" && mode != "moved-monster" {
		target.NewPos = target.PosVec
		target.ObjFlags = object.FlagActive | object.FlagEnabled
		target.Shape.Kind = server.ShapeKindCircle
		target.Shape.Circle.R = 1
		target.Shape.Circle.R2 = 1
		s.Map.AddObjectToIndex(target)
		if !target.ObjFlags.Has(object.FlagPartitioned) {
			t.Fatal("real spatial insertion")
		}
		t.Cleanup(func() { s.Map.RemoveObjectFromIndex(target) })
	}
	target.Field38 = 0x11223344
	if mode == "player-exhaustion" && !s.IsEnemyTo(source, target) {
		t.Fatal("real player enemy prerequisite")
	}
	if mode == "monster-one" && !s.IsEnemyTo(source, target) {
		t.Fatal("real monster enemy prerequisite")
	}
	rt := s.spells.duration.drainManaRuntime52E210()
	if !rt.RoundTowardZero || rt.Balance64 == nil || rt.BalanceLevel64 == nil || rt.Selection == nil || rt.AddMana32 == nil {
		t.Fatal("production startup/native services")
	}
	if mode == "precision-balance" && math.Float64bits(rt.BalanceLevel64("ManaDrainCoeff", 2)) != math.Float64bits(1.99999999) {
		t.Fatal("binary64 coefficient")
	}
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
	call := func() int { return ccall.CallIntPtr(legacy.Get_nox_xxx_spellDrainMana_52E210(), r.C()) }
	wantTypes := []byte{}
	wantMana, wantSupply, wantTargetMana := uint16(3), int32(20), uint16(5)
	switch mode {
	case "obelisk-debit", "quest-obelisk", "matched-team":
		wantTypes = []byte{5}
		for i, amount := range []int32{1, 2, 1} {
			credit := amount
			if mode == "quest-obelisk" {
				credit = int32(float32(amount) * 1.5)
			} else {
				wantSupply -= amount
			}
			wantMana += uint16(credit)
			if got := call(); got != 0 {
				t.Fatalf("C tick%d return%d", i, got)
			}
			carry := float32(.75)
			if i == 1 {
				carry = .25
			}
			if update.ManaCur != wantMana || obelisk.Mana != wantSupply || r.Field72 != int32(math.Float32bits(carry)) {
				t.Fatalf("C tick%d mana%d/%d supply%d/%d carry%#x", i, update.ManaCur, wantMana, obelisk.Mana, wantSupply, uint32(r.Field72))
			}
		}
		wantSync := uint32(math.MaxUint32)
		if mode == "quest-obelisk" {
			wantSync = 0x11223344
		}
		if target.Field38 != wantSync || r.Target48 != target || s.spells.duration.durationRayTargets[r] != target {
			t.Fatal("sync/live ray target")
		}
	case "player-exhaustion":
		wantTypes = []byte{5, 12}
		for i, amount := range []uint16{1, 2, 1, 1} {
			wantMana += amount
			wantTargetMana -= amount
			if got := call(); got != 0 || update.ManaCur != wantMana || targetUpdate.ManaCur != wantTargetMana {
				t.Fatalf("player debit%d return%d source%d/%d target%d/%d", i, got, update.ManaCur, wantMana, targetUpdate.ManaCur, wantTargetMana)
			}
		}
		if got := call(); got != 1 || r.Target48 != nil || s.spells.duration.durationRayTargets[r] != nil {
			t.Fatal("exhausted player no-target stop", got)
		}
	case "monster-one":
		wantTypes = []byte{5}
		wantMana = 4
		if got := call(); got != 0 || update.ManaCur != 4 || monster.StatusFlags != object.MonsterStatus(0x20) {
			t.Fatal("monster one-mana source", got)
		}
	case "blocked-team":
		wantTypes = []byte{5}
		if got := call(); got != 0 || target.Field38 != 0x11223344 || s.spells.duration.durationRayTargets[r] != target {
			t.Fatal("opposing Obelisk team", got)
		}
	case "no-target-stop":
		wantTypes = []byte{12}
		s.spells.duration.durationRayTargets = map[*server.DurSpell]*server.Object{r: previous}
		if got := call(); got != 1 || r.Target48 != nil || s.spells.duration.durationRayTargets[r] != nil {
			t.Fatal("no target C stop", got)
		}
	case "recent-guard", "moved-monster":
		if got := call(); got != 1 {
			t.Fatal("native cancellation guard", got)
		}
	case "glyph-nil-player":
		wantTargetMana = 0
		if got := call(); got != 1 || targetUpdate.ManaCur != 0 || r.Target48 != previous || s.spells.duration.drainManaSelection.Target != target || s.spells.duration.durationRayTargets[r] != nil || r.Field72 != int32(math.Float32bits(.25)) {
			t.Fatal("native glyph player drain", got)
		}
	case "full-dword":
		wantTypes = []byte{5}
		wantSupply = 34464
		if got := call(); got != 0 || obelisk.Mana != 34464 || update.ManaCur != 3 || update.ManaPrev != 3 || r.Field72 != 0 {
			t.Fatal("native DWORD coefficient/word-wrapped mana core", got)
		}
	case "precision-balance":
		wantTypes = []byte{5}
		wantMana = 4
		wantSupply = 19
		if got := call(); got != 0 || update.ManaCur != 4 || obelisk.Mana != 19 || uint32(r.Field72) != 0x3f7ffffe {
			t.Fatalf("directed C coefficient ret%d mana%d supply%d carry%#x", got, update.ManaCur, obelisk.Mana, uint32(r.Field72))
		}
	default:
		t.Fatal("unknown fixture")
	}
	if update.ManaCur != wantMana || obelisk.Mana != wantSupply || targetUpdate.ManaCur != wantTargetMana {
		t.Fatalf("unexpected native mana source%d/%d supply%d/%d target%d/%d", update.ManaCur, wantMana, obelisk.Mana, wantSupply, targetUpdate.ManaCur, wantTargetMana)
	}
	if len(packets) != len(wantTypes) {
		t.Fatalf("packets%d want%d", len(packets), len(wantTypes))
	}
	for i, p := range packets {
		targetCode := uint16(target.NetCode)
		if target.ObjClass == object.ClassImmobile {
			targetCode = 0xc567
		}
		first, second := uint16(source.NetCode), targetCode
		if wantTypes[i] == 12 {
			first, second = targetCode, uint16(source.NetCode)
			if mode == "no-target-stop" {
				first = uint16(previous.NetCode)
			}
		}
		want := []byte{0x9e, wantTypes[i], 3, 0, 0, 0, 0}
		binary.LittleEndian.PutUint16(want[3:5], first)
		binary.LittleEndian.PutUint16(want[5:7], second)
		if !bytes.Equal(p.packet, want) || p.recipient != 255 || p.remove != 1 || p.sequence != 1 || p.result != 1 {
			t.Fatalf("allocated packet%+v want%x", p, want)
		}
	}
	if r.Obj12 != source || r.Obj24 != previous || r.Flags88 != 0x77556601 || r.Frame68 != 0x87654321 || update.Player != info || targetUpdate.Player != targetInfo {
		t.Fatal("native record/update links")
	}
	t.Logf("actual C DrainMana %s record=%p caster=%p target=%p mana3->%d supply->%d target-mana5->%d rays%d startup=true", mode, r, source, target, update.ManaCur, obelisk.Mana, targetUpdate.ManaCur, len(packets))
}
