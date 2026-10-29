package opennox

import (
	"bytes"
	"image"
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
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// The actual C entries use the production binary64 balance reader, allocator,
// create-at/owner list, buff exports, native sidecar and delayed deletion.
// The visual definition is synthetic; the existing host E2E exercises stock
// Moonglow updates. Protection handle zero is the existing inactive handle.
func TestSpellMoonglowPublicCNativeComposition531A00(t *testing.T) {
	const childEnv = "NOX_MOONGLOW_NATIVE_COMPOSITION_CHILD"
	mode := os.Getenv(childEnv)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"player-center", "player-cursor", "nonplayer-light", "allocation-fail", "nil-target-cleanup", "player-missing-visual"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestSpellMoonglowPublicCNativeComposition531A00$", "-test.v")
				cmd.Env = append(os.Environ(), childEnv+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual native C Moonglow composition: %v", err)
				}
			})
		}
		return
	}
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(oldFlags)
	t.Cleanup(func() { noxflags.UnsetGame(noxflags.GetGame()); noxflags.SetGame(oldFlags) })
	s.spells.duration.Init(s)
	s.SetFrame(142)
	s.SetTickRate(30)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	dir, oldData := t.TempDir(), datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	if err := os.WriteFile(filepath.Join(dir, "gamedata.yml"), []byte("MoonglowEnchantmentDuration: 1.999999999999\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	if math.Float64bits(s.Balance.Float("MoonglowEnchantmentDuration")) != math.Float64bits(1.999999999999) {
		t.Fatal("actual binary64 balance")
	}
	if mode != "allocation-fail" {
		if err := s.Types.ReadObjectType(&things.Thing{Name: "Moonglow", Class: []things.Class{"SIMPLE"}}); err != nil {
			t.Fatal(err)
		}
	}
	type recordFrame struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	type objectFrame struct {
		Before [32]byte
		Object server.Object
		After  [32]byte
	}
	type dataFrame struct {
		Before [32]byte
		Data   server.PlayerUpdateData
		After  [32]byte
	}
	type playerFrame struct {
		Before [32]byte
		Player server.Player
		After  [32]byte
	}
	f, freeRecord := alloc.New(recordFrame{})
	t.Cleanup(freeRecord)
	*f = recordFrame{}
	source, freeSource := alloc.New(objectFrame{})
	t.Cleanup(freeSource)
	*source = objectFrame{}
	data, freeData := alloc.New(dataFrame{})
	t.Cleanup(freeData)
	*data = dataFrame{}
	info, freePlayer := alloc.New(playerFrame{})
	t.Cleanup(freePlayer)
	*info = playerFrame{}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
		source.Before[i], source.After[i] = byte(0x31+i), byte(0xc7-i)
		data.Before[i], data.After[i] = byte(0x71+i), byte(0xa7-i)
		info.Before[i], info.After[i] = byte(0x11+i), byte(0x97-i)
	}
	info.Player.CursorVec = image.Pt(123, 456)
	data.Data.Player = &info.Player
	source.Object = server.Object{ObjClass: object.ClassPlayer, TypeInd: 713, UpdateData: unsafe.Pointer(&data.Data), PosVec: types.Ptf(100, 200)}
	if mode == "player-cursor" {
		info.Player.Field3680 = 0x10
	}
	if mode == "nonplayer-light" {
		source.Object.ObjClass = object.ClassSimple | object.ClassClientPersist
		source.Object.UpdateData = nil
	}
	f.Record = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_MOONGLOW), Level: 0x103, Obj12: &source.Object, Caster16: &source.Object, Target48: &source.Object, Field72: 0x55aa, Pos: types.Ptf(11.5, -22.25), Frame68: 0x87654321, Flags88: 0x77556601}
	if mode == "nil-target-cleanup" {
		f.Record.Target48 = nil
	}
	if mode == "player-missing-visual" {
		source.Object.Buffs = 1 << server.ENCHANT_MOONGLOW
		source.Object.BuffsDur[server.ENCHANT_MOONGLOW] = 9
		source.Object.BuffsPower[server.ENCHANT_MOONGLOW] = 2
	}
	// Object cleanup must run while its C-owned owner frame remains alive.
	if !s.Objs.Init(8) {
		t.Fatal("native allocator initialization")
	}
	t.Cleanup(s.Objs.FreeObjects)
	raw := func(p unsafe.Pointer, n uintptr) []byte { return unsafe.Slice((*byte)(p), n) }
	recordBefore := append([]byte(nil), raw(unsafe.Pointer(f), unsafe.Sizeof(*f))...)
	dataBefore := append([]byte(nil), raw(unsafe.Pointer(data), unsafe.Sizeof(*data))...)
	playerBefore := append([]byte(nil), raw(unsafe.Pointer(info), unsafe.Sizeof(*info))...)
	sourceBefore := append([]byte(nil), raw(unsafe.Pointer(source), unsafe.Sizeof(*source))...)
	allowed := make([]bool, len(sourceBefore))
	allow := func(p unsafe.Pointer, n uintptr) {
		off := uintptr(p) - uintptr(unsafe.Pointer(source))
		for i := off; i < off+n; i++ {
			allowed[i] = true
		}
	}
	allow(unsafe.Pointer(&source.Object.Field38), unsafe.Sizeof(source.Object.Field38))
	allow(unsafe.Pointer(&source.Object.Field140), unsafe.Sizeof(source.Object.Field140))
	allow(unsafe.Pointer(&source.Object.Field129), unsafe.Sizeof(source.Object.Field129))
	allow(unsafe.Pointer(&source.Object.Buffs), unsafe.Sizeof(source.Object.Buffs))
	enchant := server.ENCHANT_MOONGLOW
	if mode == "nonplayer-light" {
		enchant = server.ENCHANT_LIGHT
	}
	allow(unsafe.Pointer(&source.Object.BuffsDur[enchant]), unsafe.Sizeof(source.Object.BuffsDur[enchant]))
	allow(unsafe.Pointer(&source.Object.BuffsPower[enchant]), unsafe.Sizeof(source.Object.BuffsPower[enchant]))
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{f.Record.C(), source.Object.CObj(), unsafe.Pointer(&data.Data), unsafe.Pointer(&info.Player)} {
			if uintptr(p) <= 0xffffffff {
				t.Fatal("native C-owned Moonglow links must be above4GiB")
			}
		}
	}
	var visual *server.Object
	switch mode {
	case "player-center", "player-cursor", "nonplayer-light", "allocation-fail":
		got := ccall.CallIntPtr(legacy.Get_nox_xxx_spellCreateMoonglow_531A00(), f.Record.C())
		want := 0
		if mode == "nonplayer-light" || mode == "allocation-fail" {
			want = 1
		}
		if got != want {
			t.Fatalf("actual C create=%d want=%d", got, want)
		}
		visual = s.spells.duration.moonglowVisuals[&f.Record]
		if mode == "allocation-fail" {
			if visual != nil || source.Object.Buffs != 0 || s.Objs.Pending != nil {
				t.Fatal("failed allocation must not create/buff")
			}
		} else {
			if source.Object.Buffs != 1<<enchant || source.Object.BuffsDur[enchant] != 1 || source.Object.BuffsPower[enchant] != 3 {
				t.Fatal("actual buff word/byte and directed duration conversion")
			}
			if mode == "nonplayer-light" {
				if visual != nil || s.Objs.Pending != nil {
					t.Fatal("nonplayer Light must not create a visual")
				}
			} else {
				point := types.Ptf(2944, 2944)
				if mode == "player-cursor" {
					point = types.Ptf(123, 456)
				}
				if visual == nil || visual.PosVec != point || visual.ObjOwner != &source.Object || source.Object.Field129 != visual || s.Objs.Pending != visual || visual.Field32 != 142 {
					t.Fatal("actual allocation/create-at/native ownership")
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(visual.CObj()) <= 0xffffffff {
					t.Fatal("native visual link must be above4GiB")
				}
			}
			ccall.CallVoidPtr(legacy.Get_sub_531AF0(), f.Record.C())
			if source.Object.Buffs != 0 || source.Object.BuffsDur[enchant] != 0 || source.Object.BuffsPower[enchant] != 0 || s.spells.duration.moonglowVisuals[&f.Record] != nil {
				t.Fatal("actual C destruction must clear buff and sidecar")
			}
			if visual != nil && (!visual.Flags().Has(object.FlagDestroyed) || s.Objs.DeletedList != visual) {
				t.Fatal("actual delayed deletion must own the visual")
			}
		}
	case "nil-target-cleanup":
		visual = s.Objs.NewObject(&server.ObjectType{})
		s.spells.duration.moonglowRuntime531A00().StoreVisual(&f.Record, visual)
		ccall.CallVoidPtr(legacy.Get_sub_531AF0(), f.Record.C())
		if s.spells.duration.moonglowVisuals[&f.Record] != nil || visual.Flags().Has(object.FlagDestroyed) || source.Object.Buffs != 0 {
			t.Fatal("retain ordinary nil-target lifetime cleanup without extra original-root work")
		}
	case "player-missing-visual":
		ccall.CallVoidPtr(legacy.Get_sub_531AF0(), f.Record.C())
		if source.Object.Buffs != 0 || source.Object.BuffsDur[enchant] != 0 || source.Object.BuffsPower[enchant] != 0 || s.Objs.DeletedList != nil {
			t.Fatal("missing visual must still remove the actual player buff")
		}
	default:
		t.Fatal("unknown composition mode", mode)
	}
	if !bytes.Equal(recordBefore, raw(unsafe.Pointer(f), unsafe.Sizeof(*f))) || !bytes.Equal(dataBefore, raw(unsafe.Pointer(data), unsafe.Sizeof(*data))) || !bytes.Equal(playerBefore, raw(unsafe.Pointer(info), unsafe.Sizeof(*info))) {
		t.Fatal("record/update/player native frame or canary changed")
	}
	for i, value := range raw(unsafe.Pointer(source), unsafe.Sizeof(*source)) {
		if value != sourceBefore[i] && !allowed[i] {
			t.Fatalf("nonoutput source canary changed at%d:before=%02x after=%02x", i, sourceBefore[i], value)
		}
	}
	t.Logf("MOONGLOW PUBLIC C COMPOSITION PASS:mode=%s record=%p target=%p visual=%p frame=%d", mode, &f.Record, &source.Object, visual, s.Frame())
}
