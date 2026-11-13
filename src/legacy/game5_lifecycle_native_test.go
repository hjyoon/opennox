package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestGame5PublicCreateEntriesKeepNativeOwners(t *testing.T) {
	for entry, name := range []string{"obelisk", "animation", "trigger", "generator", "reward", "auto-spells", "weapon", "armor"} {
		t.Run(name, func(t *testing.T) {
			s := npcReflectLegacyServer4E17B0(t)
			if !s.Objs.InitHigh(4) {
				t.Fatal("object pool")
			}
			t.Cleanup(s.Objs.FreeObjects)
			s.FreeObjectTypes()
			t.Cleanup(s.FreeObjectTypes)
			if err := s.Types.ReadObjectType(&things.Thing{Name: "Game5NativeOwner"}); err != nil {
				t.Fatal(err)
			}
			obj := s.Objs.NewObject(s.Types.ByID("Game5NativeOwner"))
			if obj == nil {
				t.Fatal("native object")
			}
			t.Cleanup(func() { obj.UpdateData = nil; obj.InitData = nil; s.Objs.FreeObject(obj) })
			raw, free := alloc.MallocHigh(4096)
			t.Cleanup(free)
			buf := unsafe.Slice((*byte)(raw), 4096)
			for i := range buf {
				buf[i] = byte(i*29 + 7)
			}
			data := unsafe.Add(raw, 32)
			obj.UpdateData = data
			obj.InitData = data
			before := bytes.Clone(buf)
			want := bytes.Clone(buf)
			put := func(off uintptr, v uint32) { binary.LittleEndian.PutUint32(want[32+off:], v) }
			pointerResult := false
			switch entry {
			case 0:
				put(0, 50)
			case 1:
			case 2:
				at := 32 + unsafe.Offsetof(server.TriggerUpdateData{}.Colors)
				copy(want[at:at+6], []byte{90, 90, 90, 10, 10, 10})
				pointerResult = true
			case 3:
				u := server.MonsterGenUpdateData{}
				for _, off := range []uintptr{unsafe.Offsetof(u.FuncInd52), unsafe.Offsetof(u.FuncInd60), unsafe.Offsetof(u.ScriptCollision) + unsafe.Offsetof(u.ScriptCollision.Func), unsafe.Offsetof(u.FuncInd68)} {
					put(off, math.MaxUint32)
				}
				put(unsafe.Offsetof(u.Field92), 2)
				pointerResult = true
			case 4:
				u := server.RewardMarkerInitData{}
				put(unsafe.Offsetof(u.CategoryMask), 255)
				put(unsafe.Offsetof(u.ChanceMode), 0)
				pointerResult = true
			case 5:
				obj.ObjClass = object.ClassMonster
				at := 32 + unsafe.Offsetof(server.MonsterUpdateData{}.Field509)
				want[at] = 1
			}
			identity := game5ObjectIdentity(obj)
			if identity != uintptr(unsafe.Pointer(obj)) {
				t.Fatal("C owner identity")
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && (identity <= math.MaxUint32 || uintptr(data) <= math.MaxUint32) {
				t.Fatal("high native owners required")
			}
			got := game5LifecycleProbe(entry, obj)
			expected := uintptr(0)
			if pointerResult {
				expected = uintptr(data)
			} else if entry == 5 {
				expected = uintptr(obj.TypeInd)
			}
			if got != expected {
				t.Fatal("public C native return", got, expected)
			}
			if !bytes.Equal(buf, want) {
				t.Fatalf("whole4096-byte update/init owner changed outside declared stores; entry=%d before=%x", entry, before[:8])
			}
			if obj.UpdateData != data || obj.InitData != data {
				t.Fatal("native data pointer identity changed")
			}
		})
	}
}

func TestGame5PublicMonsterFunctionEntries(t *testing.T) {
	for entry, name := range []string{"OGRESTRIKE", "SWORDSMANDIE", "SKELETONDEAD"} {
		t.Run(fmt.Sprint(entry), func(t *testing.T) {
			def, free := alloc.NewHigh(server.MonsterDef{})
			defer free()
			frame := unsafe.Slice((*byte)(unsafe.Pointer(def)), int(unsafe.Sizeof(*def)))
			before := bytes.Clone(frame)
			cName, freeName := alloc.CStringHigh(name)
			defer freeName()
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(def)) <= math.MaxUint32 || uintptr(unsafe.Pointer(cName)) <= math.MaxUint32) {
				t.Fatal("high native definition and name required")
			}
			if got := game5DefinitionProbe(entry, def, unsafe.Pointer(cName)); got != 1 {
				t.Fatal("public native function loader", got)
			}
			offsets := []uintptr{unsafe.Offsetof(def.MeleeStrikeFunc236), unsafe.Offsetof(def.DieFunc228), unsafe.Offsetof(def.DeadFunc232)}
			at := offsets[entry]
			expected := bytes.Clone(before)
			ptr := game5DefinitionExpected(entry)
			if unsafe.Sizeof(uintptr(0)) == 8 {
				binary.LittleEndian.PutUint64(expected[at:], uint64(uintptr(ptr)))
			} else {
				binary.LittleEndian.PutUint32(expected[at:], uint32(uintptr(ptr)))
			}
			if !bytes.Equal(frame, expected) {
				t.Fatal("complete native MonsterDef or pointer slot differs")
			}
			unknown, freeUnknown := alloc.CStringHigh("UNKNOWN")
			defer freeUnknown()
			if got := game5DefinitionProbe(entry, def, unsafe.Pointer(unknown)); got != 0 || !bytes.Equal(frame, expected) {
				t.Fatal("unknown loader changed existing definition")
			}
			null, freeNull := alloc.CStringHigh("nUlL")
			defer freeNull()
			if got := game5DefinitionProbe(entry, def, unsafe.Pointer(null)); got != 1 {
				t.Fatal("NULL loader")
			}
			if !bytes.Equal(frame, before) {
				t.Fatal("NULL changed another field")
			}
			if game5DefinitionProbe(entry, nil, unsafe.Pointer(cName)) != 0 {
				t.Fatal("native nil owner policy")
			}
		})
	}
}

func TestGame5StrikePublicCEmptyNativeWorld(t *testing.T) {
	s := npcReflectLegacyServer4E17B0(t)
	s.Map.Init()
	defer s.Map.Free()
	obj := game5HighObject(t, 7)
	obj.ObjClass = object.ClassMonster
	update, freeUpdate := alloc.NewHigh(server.MonsterUpdateData{})
	defer freeUpdate()
	definition, freeDefinition := alloc.NewHigh(server.MonsterDef{})
	defer freeDefinition()
	update.MonsterDef = definition
	obj.UpdateData = unsafe.Pointer(update)
	defer func() { obj.UpdateData = nil; update.MonsterDef = nil }()
	for entry, want := range []int{0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 0} {
		if got := game5StrikeProbe(entry, obj); got != want {
			t.Fatal("actual C strike entry", entry, got, want)
		}
	}
}
