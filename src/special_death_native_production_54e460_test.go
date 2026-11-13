package opennox

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"reflect"
	"runtime/debug"
	"strings"
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
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func TestSpecialDeathActualCProduction54E460(t *testing.T) {
	mode := os.Getenv("OPENNOX_SPECIAL_DEATH_PRODUCTION_CHILD")
	if mode == "" {
		for _, name := range []string{"marker", "marker-no-owner", "marker-duplicate", "boulder", "boulder-allocation-failure", "generator", "generator-quest-credit", "generator-allocation-failure", "generator-quest-no-owner"} {
			t.Run(name, func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.run=^TestSpecialDeathActualCProduction54E460$", "-test.v")
				cmd.Env = append(os.Environ(), "OPENNOX_SPECIAL_DEATH_PRODUCTION_CHILD="+name)
				out, err := cmd.CombinedOutput()
				t.Logf("actual default C %s %v\n%s", name, err, out)
				if err != nil || !bytes.Contains(out, []byte("SPECIAL_DEATH_ACTUAL_DEFAULT_C_PRODUCTION")) {
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
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeCoop)
	if strings.Contains(mode, "quest") {
		noxflags.SetGame(noxflags.GameModeQuest)
	}
	s.SetFrame(1400)
	s.SetTickRate(30)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if !s.Objs.InitHigh(128) {
		t.Fatal("actual native object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.InitHigh() == 0 {
		t.Fatal("actual wall grid")
	}
	t.Cleanup(s.Walls.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	wire := binary.LittleEndian.AppendUint32(nil, 2)
	for _, id := range []sound.ID{757, 1000} {
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
		t.Fatal("actual original AUD descriptors")
	}
	s.Audio.Reset()
	for _, def := range []*things.Thing{
		{Name: "Player", OnUpdate: &things.ProcFunc{Name: "PlayerUpdate"}},
		{Name: "Marker"}, {Name: "Boulder"},
		{Name: "MonsterGenerator", OnUpdate: &things.ProcFunc{Name: "MonsterGeneratorUpdate"}},
	} {
		if err := s.Types.ReadObjectType(def); err != nil {
			t.Fatal(err)
		}
	}
	if !strings.Contains(mode, "allocation-failure") {
		for _, name := range []string{"BigRock", "MediumRock", "SmallRock", "DestroyedGenerator"} {
			if err := s.Types.ReadObjectType(&things.Thing{Name: name, Mass: 10}); err != nil {
				t.Fatal(err)
			}
		}
	}
	entry, typeName := 0, "Marker"
	if strings.HasPrefix(mode, "boulder") {
		entry, typeName = 1, "Boulder"
	} else if strings.HasPrefix(mode, "generator") {
		entry, typeName = 2, "MonsterGenerator"
	}
	unit, parent := s.NewObjectByTypeID(typeName), s.NewObjectByTypeID("Player")
	if unit == nil || parent == nil || parent.UpdateData == nil || (entry == 2 && unit.UpdateData == nil) {
		t.Fatal("actual typed C constructor and complete update allocations")
	}
	unit.ObjFlags, unit.PosVec = object.FlagActive|object.FlagEnabled, types.Ptf(300, 300)
	unit.ObjClass = 0
	parent.ObjClass, parent.ObjFlags = object.ClassPlayer, object.FlagActive|object.FlagEnabled
	player, freePlayer := alloc.NewHigh(server.Player{})
	t.Cleanup(freePlayer)
	update := (*server.PlayerUpdateData)(parent.UpdateData)
	update.Player = player
	update.Field29 = [4]*server.Object{unit, nil, parent, nil}
	if mode == "marker-duplicate" {
		update.Field29[2] = unit
	}
	if entry == 0 && mode != "marker-no-owner" {
		unit.ObjOwner = parent
	}
	if entry == 2 {
		g := (*server.MonsterGenUpdateData)(unit.UpdateData)
		g.Field56, g.FuncInd60 = math.MaxUint32, 0
		if mode != "generator-quest-no-owner" {
			unit.Obj130 = parent
		}
	}
	sourceFrame := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit))))
	parentFrame := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(parent)), int(unsafe.Sizeof(*parent))))
	updateFrame := bytes.Clone(unsafe.Slice((*byte)(parent.UpdateData), int(unsafe.Sizeof(*update))))
	playerFrame := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(player)), int(unsafe.Sizeof(*player))))
	var generatorFrame []byte
	if entry == 2 {
		generatorFrame = bytes.Clone(unsafe.Slice((*byte)(unit.UpdateData), int(unsafe.Sizeof(server.MonsterGenUpdateData{}))))
	}
	index, divisor := legacy.BoulderDeathIndex54E4B0(), memmap.PtrUint32(0x587000, 290340)
	*index, *divisor = 11, 12
	sounds := 0
	s.Audio.OnSound(func(id sound.ID, kind int, source *server.Object, position types.Pointf) {
		want := sound.ID(757)
		if entry == 2 {
			want = 1000
		}
		if entry == 0 || id != want || kind != 0 || source != unit || position != unit.PosVec {
			t.Fatal("actual sound source/ID/position", id, kind, source, position)
		}
		sounds++
	})
	// The original owner service returns a terminal nonplayer object itself.
	// An unowned Marker with no update allocation therefore faults at its
	// marker-slot read. Preserve that raw fault, rather than inventing a nil
	// parent admission guard or treating this malformed fixture as a live case.
	wantFault := mode == "marker-no-owner"
	var fault any
	var faultStack string
	called := false
	func() {
		defer func() {
			if value := recover(); value != nil {
				fault, faultStack = value, string(debug.Stack())
			}
		}()
		called = legacy.SpecialDeathActualDefaultC54E460(entry, unit)
	}()
	if wantFault {
		if _, ok := fault.(interface{ RuntimeError() }); !ok || !strings.Contains(faultStack, ".specialDeathMarkerSlot54E460(") || !strings.Contains(faultStack, ".specialDeathOriginalC54E460(") {
			t.Fatalf("actual default-C original marker-slot fault: %v\n%s", fault, faultStack)
		}
	} else if fault != nil || !called {
		t.Fatalf("actual typed public C symbol with default services: %v\n%s", fault, faultStack)
	}
	pieces := 0
	table := []string{"BigRock", "MediumRock", "SmallRock", "MediumRock", "SmallRock", "SmallRock", "MediumRock", "SmallRock", "SmallRock", "SmallRock", "SmallRock", "SmallRock"}
	var created []*server.Object
	for o := s.Objs.Pending; o != nil; o = o.ObjNext {
		created = append(created, o)
	}
	pieces = len(created)
	for i, o := range created {
		if o.ObjOwner != nil || !o.Flags().Has(object.FlagActive|object.FlagPending) {
			t.Fatal("actual typed create ownership/flags")
		}
		name := o.ObjectTypeC().ID()
		if entry == 1 {
			want := table[(11+pieces-1-i)%12]
			mass := map[string]float32{"BigRock": 1, "MediumRock": 3, "SmallRock": 5}[want]
			if name != want || !o.Flags().Has(object.FlagBouncy) || o.Field29 != math.Float32bits(mass) || o.Field27 < -2 || o.Field27 > 0 || o.ZVal < 10 || o.ZVal > 70 {
				t.Fatal("actual packed name/mass sequence and native motion", i, name, want)
			}
		} else if entry != 2 || name != "DestroyedGenerator" || o.PosVec != unit.PosVec {
			t.Fatal("actual destroyed generator type/live source position", name, o.PosVec)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(o)) <= math.MaxUint32 {
			t.Fatal("actual constructed pointer must exceed4GiB")
		}
	}
	if entry == 1 && mode == "boulder" {
		if pieces < 20 || pieces > 30 || *index != uint32((11+pieces)%12) {
			t.Fatal("actual Boulder count/cache", pieces, *index)
		}
	} else {
		want := 0
		if entry == 2 && !strings.Contains(mode, "allocation-failure") {
			want = 1
		}
		if pieces != want || *index != 11 {
			t.Fatal("actual allocation/deletion branch and untouched cache", pieces, want, *index)
		}
	}
	deleted := mode != "boulder-allocation-failure" && !wantFault
	if unit.Flags().Has(object.FlagDestroyed) != deleted || (deleted && (s.Objs.DeletedList != unit || unit.DeletedAt != 1400)) {
		t.Fatal("actual delayed delete, including allocation failure contract")
	}
	if deleted {
		binary.LittleEndian.PutUint32(sourceFrame[unsafe.Offsetof(unit.ObjFlags):], uint32(unit.ObjFlags))
		binary.LittleEndian.PutUint32(sourceFrame[unsafe.Offsetof(unit.DeletedAt):], unit.DeletedAt)
	}
	if !bytes.Equal(sourceFrame, unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit)))) || !bytes.Equal(parentFrame, unsafe.Slice((*byte)(unsafe.Pointer(parent)), int(unsafe.Sizeof(*parent)))) {
		t.Fatal("complete actual source/parent native guards outside exact lifecycle writes")
	}
	if entry == 0 && mode != "marker-no-owner" {
		off := int(unsafe.Offsetof(update.Field29))
		clear(updateFrame[off : off+int(unsafe.Sizeof(uintptr(0)))])
	}
	if !bytes.Equal(updateFrame, unsafe.Slice((*byte)(parent.UpdateData), int(unsafe.Sizeof(*update)))) {
		t.Fatal("complete PlayerUpdate guard and first matching marker slot only")
	}
	credit := uint32(0)
	if mode == "generator-quest-credit" {
		credit = 1
		for name, value := range map[string]uint32{"field4668": 1, "field4692": 8} {
			field, ok := reflect.TypeOf(*player).FieldByName(name)
			if !ok {
				t.Fatal("independent actual player field declaration", name)
			}
			binary.LittleEndian.PutUint32(playerFrame[field.Offset:], value)
		}
	}
	words := legacy.SpecialDeathActualPlayerWords54E460(parent)
	if words != [4]uint32{credit, credit * 8, 0, 0} || !bytes.Equal(playerFrame, unsafe.Slice((*byte)(unsafe.Pointer(player)), int(unsafe.Sizeof(*player)))) {
		t.Fatal("actual typed credit helper, independent C getter and complete Player guard", words)
	}
	if entry == 2 && !bytes.Equal(generatorFrame, unsafe.Slice((*byte)(unit.UpdateData), len(generatorFrame))) {
		t.Fatal("complete generator update guard and real disabled script service")
	}
	if (entry == 0 && sounds != 0) || (entry != 0 && sounds != 1) {
		t.Fatal("actual sound count", sounds)
	}
	for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(parent), parent.UpdateData, unsafe.Pointer(player)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatal("all complete object/update/player allocations must exceed4GiB")
		}
	}
	t.Cleanup(func() {
		for _, o := range created {
			s.Objs.FreeObject(o)
		}
		s.Objs.Pending, s.Objs.DeletedList = nil, nil
		s.Objs.FreeObject(unit)
		s.Objs.FreeObject(parent)
	})
	t.Logf("SPECIAL_DEATH_ACTUAL_DEFAULT_C_PRODUCTION mode=%s pieces=%d sounds=%d cache=%d deleted=%v credit=%d source=%p original-marker-slot-fault=%v default-world=true", mode, pieces, sounds, *index, deleted, credit, unit, wantFault)
}
