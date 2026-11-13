package opennox

import (
	"bytes"
	"encoding/binary"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unsafe"
)

func TestGolemDeathActualCProduction549E90(t *testing.T) {
	mode := os.Getenv("OPENNOX_GOLEM_PRODUCTION_CHILD")
	if mode == "" {
		for _, name := range []string{"mech", "golem-coop", "golem-solo", "mech-allocation-failure", "golem-allocation-failure", "mech-empty-table"} {
			t.Run(name, func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.run=^TestGolemDeathActualCProduction549E90$", "-test.v")
				cmd.Env = append(os.Environ(), "OPENNOX_GOLEM_PRODUCTION_CHILD="+name)
				out, err := cmd.CombinedOutput()
				t.Logf("actual default C %s %v\n%s", name, err, out)
				if err != nil || !bytes.Contains(out, []byte("GOLEM_ACTUAL_DEFAULT_C_PRODUCTION")) {
					t.Fatal("actual default world composition", err)
				}
			})
		}
		return
	}
	legacy.InitBlobData()
	handles.InitHigh()
	t.Cleanup(handles.Release)
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldS, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldS, oldGet })
	oldFlags, oldGameplay := noxflags.GetGame(), noxflags.GetGamePlay()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	noxflags.SetGamePlay(noxflags.GameplayFlag1)
	if mode == "golem-solo" {
		noxflags.UnsetGame(noxflags.GameModeCoop)
	}
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
		noxflags.UnsetGamePlay(noxflags.GetGamePlay())
		noxflags.SetGamePlay(oldGameplay)
	})
	s.SetFrame(1400)
	s.SetTickRate(30)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if !s.Objs.InitHigh(64) {
		t.Fatal("actual object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.InitHigh() == 0 {
		t.Fatal("wall grid")
	}
	t.Cleanup(s.Walls.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	wire := binary.LittleEndian.AppendUint32(nil, 2)
	for _, id := range []sound.ID{487, 494} {
		name := id.String()
		wire = append(wire, byte(len(name)))
		wire = append(wire, name...)
		wire = binary.LittleEndian.AppendUint16(wire, 8)
		wire = append(wire, 0)
		wire = binary.LittleEndian.AppendUint16(wire, 40)
		wire = append(wire, 0, 0, 0, 0, 1, 's', 0)
	}
	file := luaCompatAudioFile{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(file) || file.Len() != 0 {
		t.Fatal("actual Golem death AUD")
	}
	s.Audio.Reset()
	index, count := legacy.GolemDeathIndex549E90(), memmap.PtrUint32(0x587000, 287344)
	oldIndex, oldCount := *index, *count
	*index, *count = 11, 12
	t.Cleanup(func() { *index, *count = oldIndex, oldCount })
	mech := strings.HasPrefix(mode, "mech")
	if mode == "mech-empty-table" {
		slot := memmap.PtrPtr(0x587000, 287976)
		old := *slot
		*slot = nil
		t.Cleanup(func() { *slot = old })
	}
	if !strings.Contains(mode, "allocation-failure") {
		for _, name := range []string{"MechGear", "MechHead", "MechArm", "MechLeg", "BigRock", "MediumRock", "SmallRock"} {
			if err := s.Types.ReadObjectType(&things.Thing{Name: name, Mass: 10}); err != nil {
				t.Fatal(err)
			}
		}
	}
	unit := s.Objs.NewObject(&server.ObjectType{})
	if unit == nil {
		t.Fatal("native source")
	}
	unit.ObjFlags = object.FlagActive | object.FlagEnabled
	unit.PosVec = types.Ptf(300, 300)
	sourceFrame := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit))))
	sounds := 0
	wantSound := sound.ID(487)
	if mech {
		wantSound = 494
	}
	s.Audio.OnSound(func(id sound.ID, kind int, source *server.Object, pos types.Pointf) {
		if id != wantSound || kind != 0 || source != unit || pos != unit.PosVec {
			t.Fatal("actual native source audio", id, kind, source, pos)
		}
		sounds++
	})
	value, ok := legacy.GolemDeathActualDefaultC549E90(mech, unit)
	if !ok || value != 1 {
		t.Fatal("typed public C/default services", value, ok)
	}
	names := []string{}
	for o := s.Objs.Pending; o != nil; o = o.ObjNext {
		if o.ObjOwner != nil || !o.Flags().Has(object.FlagActive|object.FlagPending|object.FlagBouncy) {
			t.Fatal("actual create flags/ownership")
		}
		name := o.ObjectTypeC().ID()
		names = append(names, name)
		mass := float32(2)
		if !mech {
			mass = map[string]float32{"BigRock": 1, "MediumRock": 3, "SmallRock": 5}[name]
		}
		if mass == 0 || o.Field29 != math.Float32bits(mass) || o.Field27 < -2 || o.Field27 > 0 || o.ZVal < 10 || o.ZVal > 70 {
			t.Fatal("actual native motion/mass/raise", name, o.Field29, o.Field27, o.ZVal)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(o)) <= math.MaxUint32 {
			t.Fatal("constructor pointer below4GiB")
		}
	}
	pieces := len(names)
	switch {
	case strings.Contains(mode, "allocation-failure") || mode == "mech-empty-table":
		if pieces != 0 || *index != 11 {
			t.Fatal("actual allocation/empty-table short circuit", pieces, *index)
		}
	case mech:
		if pieces != 8 || *index != 11 {
			t.Fatal("actual Mech count/untouched Golem cache", pieces, *index)
		}
		for i, name := range names {
			if name != []string{"MechGear", "MechHead", "MechArm", "MechGear", "MechLeg", "MechLeg", "MechGear", "MechGear"}[7-i] {
				t.Fatal("actual Mech table order", i, name)
			}
		}
	default:
		if mode == "golem-solo" {
			if pieces != 6 {
				t.Fatal("actual non-Coop count", pieces)
			}
		} else if pieces < 20 || pieces > 30 {
			t.Fatal("actual Coop count", pieces)
		}
		if *index != uint32((11+pieces)%12) {
			t.Fatal("actual C cache final value", *index, pieces)
		}
		table := []string{"BigRock", "MediumRock", "SmallRock", "MediumRock", "SmallRock", "SmallRock", "MediumRock", "SmallRock", "SmallRock", "SmallRock", "SmallRock", "SmallRock"}
		for i, name := range names {
			if name != table[(11+pieces-1-i)%12] {
				t.Fatal("actual shared C cache/table sequence", i, name)
			}
		}
	}
	if sounds != 1 {
		t.Fatal("actual source sound count", sounds)
	}
	if !bytes.Equal(sourceFrame, unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))) {
		t.Fatal("whole native source record")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatal("source pointer below4GiB")
	}
	t.Cleanup(func() {
		for o := s.Objs.Pending; o != nil; {
			next := o.ObjNext
			s.Objs.FreeObject(o)
			o = next
		}
		s.Objs.Pending = nil
		s.Objs.FreeObject(unit)
	})
	t.Logf("GOLEM_ACTUAL_DEFAULT_C_PRODUCTION mode=%s pieces=%d sounds=%d cache=%d real-C-constructor-point-force-create-raise-AUD-RNG=true", mode, pieces, sounds, *index)
}
