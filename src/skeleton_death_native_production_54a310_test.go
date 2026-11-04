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
	"testing"
	"unsafe"
)

func TestSkeletonDeathActualCProduction54A310(t *testing.T) {
	mode := os.Getenv("OPENNOX_SKELETON_PRODUCTION_CHILD")
	if mode == "" {
		for _, m := range []string{"skeleton", "lord", "debris", "drop", "drop-flags-off", "allocation-failure"} {
			t.Run(m, func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.run=^TestSkeletonDeathActualCProduction54A310$", "-test.v")
				cmd.Env = append(os.Environ(), "OPENNOX_SKELETON_PRODUCTION_CHILD="+m)
				out, err := cmd.CombinedOutput()
				t.Logf("actual default C %s %v\n%s", m, err, out)
				if err != nil || !bytes.Contains(out, []byte("SKELETON_ACTUAL_DEFAULT_C_PRODUCTION")) {
					t.Fatal("actual default world composition", err)
				}
			})
		}
		return
	}
	// Fresh test children do not run game startup. Initialize the real packed
	// data and native pointer slots, including the stock bone-name table.
	legacy.InitBlobData()
	handles.Init()
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
	if !s.Objs.Init(64) {
		t.Fatal("real object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.Init() == 0 {
		t.Fatal("wall grid")
	}
	t.Cleanup(s.Walls.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	name := sound.ID(356).String()
	wire := binary.LittleEndian.AppendUint32(nil, 1)
	wire = append(wire, byte(len(name)))
	wire = append(wire, name...)
	wire = binary.LittleEndian.AppendUint16(wire, 8)
	wire = append(wire, 0)
	wire = binary.LittleEndian.AppendUint16(wire, 40)
	wire = append(wire, 0, 0, 0, 0, 1, 's', 0)
	file := luaCompatAudioFile{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(file) || file.Len() != 0 {
		t.Fatal("actual skeleton death AUD")
	}
	s.Audio.Reset()
	index, count := legacy.SkeletonDeathIndex54A310(), memmap.PtrUint32(0x587000, 287348)
	blobIndex := memmap.PtrUint32(0x5D4594, 2491588)
	if index == blobIndex {
		t.Fatal("shared C cache aliases retired blob storage")
	}
	oldIndex, oldCount, oldBlobIndex := *index, *count, *blobIndex
	*index, *count, *blobIndex = 0, 2, 0x7fffffff
	t.Cleanup(func() { *index, *count, *blobIndex = oldIndex, oldCount, oldBlobIndex })
	entry := map[string]int{"skeleton": 0, "lord": 3, "debris": 2, "drop": 1, "drop-flags-off": 1, "allocation-failure": 2}[mode]
	if mode != "allocation-failure" {
		for _, n := range []string{"Skull", "ArmBone", "LegBone", "Sword", "WoodenShield", "SteelShield", "TestItem"} {
			if err := s.Types.ReadObjectType(&things.Thing{Name: n, Mass: 10}); err != nil {
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
	s.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, pos types.Pointf) {
		if id != 356 || kind != 0 || owner != unit || pos != unit.PosVec {
			t.Fatal("real source audio", id, kind, owner, pos)
		}
		sounds++
	})
	source := unit
	if mode == "drop-flags-off" {
		noxflags.UnsetGame(noxflags.GameModeCoop)
		source = nil
	}
	value, ok := legacy.SkeletonDeathActualDefaultC54A310(entry, source, "TestItem", [4]string{}, 0)
	if !ok || value != 1 {
		t.Fatal("typed public C and default runtime", value, ok)
	}
	if *blobIndex != 0x7fffffff {
		t.Fatal("raw native entry changed retired blob cache")
	}
	pieces, bones, skulls, items := 0, 0, 0, 0
	for obj := s.Objs.Pending; obj != nil; obj = obj.ObjNext {
		pieces++
		if obj.ObjOwner != nil || !obj.Flags().Has(object.FlagActive|object.FlagPending) {
			t.Fatal("real create ownership/state")
		}
		switch obj.ObjectTypeC().ID() {
		case "Skull":
			skulls++
			if obj.ZVal != 40 {
				t.Fatal("skull raise")
			}
		case "ArmBone", "LegBone":
			bones++
			if obj.ZVal < 10 || obj.ZVal > 35 {
				t.Fatal("bone raise")
			}
		default:
			items++
		}
		if obj.ObjectTypeC().ID() == "Skull" || obj.ObjectTypeC().ID() == "ArmBone" || obj.ObjectTypeC().ID() == "LegBone" {
			if !obj.Flags().Has(object.FlagBouncy) || obj.Field29 != math.Float32bits(4) || obj.Field27 < -2 || obj.Field27 > 0 {
				t.Fatal("native debris motion")
			}
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
			t.Fatal("constructor pointer below4GiB")
		}
	}
	switch mode {
	case "drop":
		if items != 1 || pieces != 1 || sounds != 0 {
			t.Fatal("actual drop", items, pieces, sounds)
		}
	case "drop-flags-off":
		if pieces != 0 || sounds != 0 {
			t.Fatal("raw coop gate nil admission")
		}
	case "allocation-failure":
		if pieces != 0 || sounds != 1 {
			t.Fatal("raw debris allocation prefix")
		}
	default:
		if skulls != 1 || bones < 10 || bones > 20 || items > 1 || sounds != 1 || *index != uint32(bones%2) {
			t.Fatal("actual debris/count/cache/audio", skulls, bones, items, sounds, *index)
		}
	}
	if !bytes.Equal(sourceFrame, unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))) {
		t.Fatal("whole source record")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatal("source pointer below4GiB")
	}
	t.Cleanup(func() {
		for obj := s.Objs.Pending; obj != nil; {
			next := obj.ObjNext
			s.Objs.FreeObject(obj)
			obj = next
		}
		s.Objs.Pending = nil
		s.Objs.FreeObject(unit)
	})
	t.Logf("SKELETON_ACTUAL_DEFAULT_C_PRODUCTION mode=%s entry=%d pieces=%d bones=%d sounds=%d cache=%d real-root-constructor-point-force-create-raise-AUD-RNG=true", mode, entry, pieces, bones, sounds, *index)
}
