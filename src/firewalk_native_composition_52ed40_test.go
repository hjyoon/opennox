package opennox

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

//go:embed server/testdata/spell_firewalk_native_original_startup_52ed40.json
var firewalkCOriginal52ED40 []byte

type firewalkCState52ED40 struct {
	Target    int          `json:"target"`
	Level     uint32       `json:"level"`
	Frame60   uint32       `json:"frame60"`
	Frame64   uint32       `json:"frame64"`
	Previous  [2]uint32    `json:"previous"`
	Anchor    [2]uint32    `json:"anchor"`
	Positions [2][2]uint32 `json:"positions"`
	Flags     [2]uint32    `json:"flags"`
	Radii     [2]uint32    `json:"radii"`
	Cache     [3]uint32    `json:"cache"`
	FPS       uint32       `json:"fps"`
}

// The public C symbol composes real type lookup, object allocation/create,
// parsed AUD events and native decay. Selected original states/positions and
// decay arguments are read from the same sealed machine-code fixture.
// OnSound mutations are declared callbacks. Actual decay deadlines retain the
// original DWORD delay and live FPS; synthetic types supply no earned campaign
// credit. The normal object-creation pending flag remains intact.
func TestSpellFirewalkPublicCNativeComposition52ED40(t *testing.T) {
	const child = "NOX_FIREWALK_NATIVE_COMPOSITION_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"default", "cold-first", "first-wrap", "cold-later", "alloc-[False, False]", "level-2147483648", "distance-nan", "fps-2147483648", "mutate-audio-fps", "mutate-audio-target-b", "mutate-audio-anchor", "mutate-audio-previous", "nil", "flags-32800", "distance-15"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestSpellFirewalkPublicCNativeComposition52ED40$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual Firewalk C composition: %v", err)
				}
			})
		}
		return
	}
	var fixture struct {
		Cases []struct {
			Input struct {
				firewalkCState52ED40
				Name     string `json:"name"`
				Mutation struct {
					Stage  string `json:"stage"`
					At     int    `json:"at"`
					Action string `json:"action"`
				} `json:"mutation"`
			} `json:"input"`
			Result struct {
				Return int                  `json:"return_value"`
				State  firewalkCState52ED40 `json:"state"`
				Events []struct {
					Kind string            `json:"kind"`
					Args []json.RawMessage `json:"args"`
				} `json:"events"`
			} `json:"result"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(firewalkCOriginal52ED40, &fixture); err != nil {
		t.Fatal(err)
	}
	selected := -1
	for i, c := range fixture.Cases {
		if c.Input.Name == mode {
			selected = i
		}
	}
	if selected < 0 {
		t.Fatal("sealed native composition case")
	}
	c := fixture.Cases[selected]
	in, want := c.Input, c.Result
	point := func(v [2]uint32) types.Pointf {
		return types.Ptf(math.Float32frombits(v[0]), math.Float32frombits(v[1]))
	}
	bits := func(v types.Pointf) [2]uint32 { return [2]uint32{math.Float32bits(v.X), math.Float32bits(v.Y)} }
	u32 := func(v json.RawMessage) uint32 {
		var x uint32
		if err := json.Unmarshal(v, &x); err != nil {
			t.Fatal(err)
		}
		return x
	}
	var expectedPositions [][2]uint32
	var expectedDelays []uint32
	wrotePosition := in.Target >= 0 && in.Flags[in.Target]&0x8020 == 0 && in.Frame60 == in.Frame64
	for _, e := range want.Events {
		switch e.Kind {
		case "create":
			expectedPositions = append(expectedPositions, [2]uint32{u32(e.Args[1]), u32(e.Args[2])})
		case "decay":
			expectedDelays = append(expectedDelays, u32(e.Args[1]))
		case "random":
			wrotePosition = true
		}
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
	s.spells.duration.Init(s)
	t.Cleanup(s.spells.duration.Free)
	s.SetFrame(150)
	s.SetTickRate(in.FPS)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	var typeIDs [3]uint32
	for i, name := range [3]string{"SmallFlame", "MediumFlame", "Flame"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
		typeIDs[i] = uint32(s.Types.IndByID(name))
		if typeIDs[i] == 0 {
			t.Fatal("production type lookup")
		}
	}
	if !s.Objs.Init(16) {
		t.Fatal("native object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	t.Cleanup(s.DecayDestroy5117B0)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	name := sound.ID(46).String()
	wire := binary.LittleEndian.AppendUint32(nil, 1)
	wire = append(wire, byte(len(name)))
	wire = append(wire, name...)
	wire = binary.LittleEndian.AppendUint16(wire, 8)
	wire = append(wire, 0)
	wire = binary.LittleEndian.AppendUint16(wire, 40)
	wire = append(wire, 0, 0, 0, 0, 1, 's', 0)
	file := luaCompatAudioFile{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(file) || file.Len() != 0 || s.Audio.Field12(46) != 1 {
		t.Fatal("real AUD descriptor")
	}
	s.Audio.Reset()
	cacheIDs := typeIDs
	if mode == "alloc-[False, False]" {
		cacheIDs = [3]uint32{0xffffffff, 0xfffffffe, 0xfffffffd}
	}
	mapID := func(v uint32) uint32 {
		for i, id := range [3]uint32{11, 22, 33} {
			if v == id {
				return cacheIDs[i]
			}
		}
		return v
	}
	var oldCache [3]uint32
	for i := range oldCache {
		p := memmap.PtrUint32(0x5D4594, 2487888+4*uintptr(i))
		oldCache[i] = *p
		*p = mapID(in.Cache[i])
	}
	t.Cleanup(func() {
		for i, v := range oldCache {
			*memmap.PtrUint32(0x5D4594, 2487888+4*uintptr(i)) = v
		}
	})
	objects := [2]*server.Object{plasmaNativeAllocation531580[server.Object](t), plasmaNativeAllocation531580[server.Object](t)}
	for i, o := range objects {
		*o = server.Object{ObjClass: object.ClassMonster, ObjFlags: object.Flags(in.Flags[i]), PosVec: point(in.Positions[i])}
		o.Shape.Circle.R = math.Float32frombits(in.Radii[i])
	}
	type guarded struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	f := plasmaNativeAllocation531580[guarded](t)
	for i := range f.Before {
		f.Before[i] = byte(0x51 + i)
		f.After[i] = byte(0xe7 - i)
	}
	r := &f.Record
	*r = server.DurSpell{ID: 0xabcd, Spell: 28, Level: in.Level, Obj12: objects[0], Caster16: objects[1], Obj24: objects[0], Pos: types.Ptf(100, 200), Frame60: in.Frame60, Frame64: in.Frame64, Frame68: 0x87654321, Field72: int32(in.Previous[0]), Field76: uintptr(in.Previous[1]), Field80: in.Anchor[0], Field84: in.Anchor[1], Flags88: 0x77556601}
	if in.Target >= 0 {
		r.Target48 = objects[in.Target]
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		hi := uint64(0xa55a000000000000)
		r.Field76 |= uintptr(hi)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(r.C()) <= math.MaxUint32 {
		t.Fatal("native C-owned duration below4GiB")
	}
	expected := *f
	expected.Record.Frame64 = want.State.Frame64
	expected.Record.Field72 = int32(want.State.Previous[0])
	if wrotePosition {
		expected.Record.Field76 = uintptr(want.State.Previous[1])
	}
	expected.Record.Field80, expected.Record.Field84 = want.State.Anchor[0], want.State.Anchor[1]
	if want.State.Target < 0 {
		expected.Record.Target48 = nil
	} else {
		expected.Record.Target48 = objects[want.State.Target]
	}
	var flames []*server.Object
	var positions [][2]uint32
	s.Audio.OnSound(func(id sound.ID, kind int, obj *server.Object, p types.Pointf) {
		if id != 46 || kind != 0 || obj != nil || s.Objs.Pending == nil {
			t.Fatal("real positional audio event")
		}
		flame := s.Objs.Pending
		flames = append(flames, flame)
		positions = append(positions, bits(p))
		m := in.Mutation
		if m.Stage != "audio" || len(flames)-1 != m.At {
			return
		}
		switch m.Action {
		case "fps":
			s.SetTickRate(60)
		case "target-b":
			r.Target48 = objects[1]
		case "anchor":
			r.Field80 = math.Float32bits(7.25)
			r.Field84 = math.Float32bits(-8.5)
		case "previous":
			r.Field72 = int32(math.Float32bits(3.25))
			r.Field76 = uintptr(math.Float32bits(-4.5))
		default:
			t.Fatal("native composition callback")
		}
	})
	got := ccall.CallIntPtr(legacy.Get_nox_xxx_firewalkTick_52ED40(), r.C())
	if got != want.Return {
		t.Fatalf("public C return=%d original=%d", got, want.Return)
	}
	raw := func(p *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p))) }
	if !bytes.Equal(raw(f), raw(&expected)) {
		t.Fatalf("native C duration/canary differs from original state: native=%+v original=%+v", *r, expected.Record)
	}
	if !reflect.DeepEqual(positions, expectedPositions) {
		t.Fatalf("created/audio positions=%v original=%v", positions, expectedPositions)
	}
	if s.TickRate() != want.State.FPS {
		t.Fatal("live FPS callback state")
	}
	for i, flame := range flames {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(flame.CObj()) <= math.MaxUint32 {
			t.Fatal("allocated native flame below4GiB")
		}
		if bits(flame.PosVec) != expectedPositions[i] || bits(flame.PrevPos) != expectedPositions[i] || bits(flame.NewPos) != expectedPositions[i] || flame.ObjOwner != nil {
			t.Fatal("actual CreateObjectAt state")
		}
		if int32(in.Level) < 2 && uint32(flame.TypeInd) != typeIDs[0] {
			t.Fatal("signed original level must select SmallFlame")
		}
		if flame.Field34 != 150+expectedDelays[i] || uint32(flame.ObjFlags)&0x00400000 == 0 {
			t.Fatalf("actual decay deadline=%d original delay=%d", flame.Field34, expectedDelays[i])
		}
		if !flame.ObjFlags.Has(object.FlagPending) {
			t.Fatal("ordinary CreateObjectAt pending flag was changed")
		}
	}
	audioCount := 0
	s.Audio.EachEvent(func(e *server.AudioEvent) {
		audioCount++
		if e.Sound != 46 || e.Obj != nil || e.Kind != 0 || e.Code != 0 {
			t.Fatal("actual audio allocation")
		}
	})
	if audioCount != len(expectedPositions) {
		t.Fatal("actual audio event count")
	}
	for i, v := range want.State.Cache {
		if *memmap.PtrUint32(0x5D4594, 2487888+4*uintptr(i)) != mapID(v) {
			t.Fatal("original cache store/order")
		}
	}
	t.Logf("original case=%s native C record=%p target=%p flames=%d positional events=%d FPS=%d; real create/audio/decay and full record canaries verified", mode, r, r.Target48, len(flames), audioCount, s.TickRate())
}
