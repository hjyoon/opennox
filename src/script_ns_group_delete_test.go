package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/noxscript/ns/asm"
	ns4 "github.com/opennox/noxscript/ns/v4"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// This compares the modern API with the unchanged SCRIPT03 GroupDelete
// builtin, not with a supplied deletion callback. Objects, updates, groups,
// and members are C-owned; all pointers must stay native-width. The fixture
// has no inventory, player, team, monitored creature, or active duration spell.
// Those services retain their own existing deletion regression tests.
type scriptNSGroupDeleteFixture struct {
	s        *Server
	units    []*server.Object
	groups   map[string]*server.MapGroup
	objects  [][]byte
	records  []scriptNSRoamFlagRecord515CB0
	sentinel *server.Object
}

func scriptNSGroupDeleteNew(t *testing.T, class object.Class, frame uint32) *scriptNSGroupDeleteFixture {
	t.Helper()
	f := &scriptNSGroupDeleteFixture{s: scriptNSRoamFlagServer515C80(t), groups: make(map[string]*server.MapGroup)}
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost)
	f.s.SetFrame(frame)
	if !f.s.Objs.Init(8) {
		t.Fatal("cannot initialize native object allocator")
	}
	t.Cleanup(f.s.Objs.FreeObjects)
	f.s.MapGroups.Init()
	t.Cleanup(f.s.MapGroups.Free)
	snapshot := func(name string, ptr unsafe.Pointer, size uintptr) {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("native %s pointer is not above 4 GiB: %p", name, ptr)
		}
		f.records = append(f.records, scriptNSRoamFlagRecord515CB0{name, ptr, size, scriptNSRoamFlagBytes515C80(ptr, size)})
	}
	for i := 0; i < 7; i++ {
		unit := f.s.Objs.NewObject(&server.ObjectType{})
		if unit == nil {
			t.Fatal("cannot allocate native object")
		}
		unit.ObjClass, unit.ObjFlags = object.ClassMonster, object.FlagEnabled
		unit.Extent, unit.NetCode = 0xe1234500+uint32(i), 201+uint32(i)
		unit.DeletedAt = 0x12345678
		update, free := alloc.New(server.MonsterUpdateData{})
		t.Cleanup(free)
		unit.UpdateData = unsafe.Pointer(update)
		update.AIStackInd, update.Field333, update.Field124 = -1, 0xdeadbeef, 0xabcdef01
		snapshot(fmt.Sprintf("update-%d", i), unsafe.Pointer(update), unsafe.Sizeof(*update))
		if i != 0 {
			f.units[i-1].ObjNext = unit
		}
		f.units = append(f.units, unit)
	}
	f.units[0].ObjClass = class
	f.units[0].ObjSubClass = object.SubClass(69778)
	f.units[1].ObjClass = object.ClassMissile
	f.units[2].ObjFlags |= object.FlagDead
	f.units[3].ObjFlags |= object.FlagDestroyed
	f.units[4].ObjClass = object.ClassSimple
	f.units[5].ObjFlags |= object.FlagNoUpdate
	f.s.Objs.List = f.units[0]
	f.s.Objs.DeletedList, f.s.Objs.Pending = nil, nil
	sentinel, freeSentinel := alloc.New(server.Object{})
	f.s.Objs.DeletedList = sentinel
	f.s.Objs.DeletedList.ObjFlags = object.FlagDestroyed
	f.s.Objs.DeletedList.DeletedAt = 17
	f.s.Objs.DeletedList.Extent = 0xe7777777
	f.sentinel = f.s.Objs.DeletedList
	t.Cleanup(freeSentinel)
	t.Cleanup(func() { f.s.Objs.List, f.s.Objs.DeletedList, f.s.Objs.Pending = nil, nil, nil })
	snapshot("sentinel", unsafe.Pointer(f.sentinel), unsafe.Sizeof(*f.sentinel))
	for i, spec := range []struct {
		name string
		kind server.MapGroupKind
	}{
		{"root", server.MapGroupGroups}, {"middle", server.MapGroupGroups},
		{"leaf-a", server.MapGroupObjects}, {"leaf-b", server.MapGroupObjects},
		{"empty", server.MapGroupObjects}, {"waypoints", server.MapGroupWaypoints},
		{"walls", server.MapGroupWalls},
	} {
		id := uint32(0xf1234e00) + uint32(i)
		if f.s.MapGroups.MapLoadAddGroup57C0C0(spec.name, id, byte(spec.kind)) != 1 {
			t.Fatal("cannot allocate native group")
		}
		g := f.s.MapGroups.GroupByInd(int(int32(id)))
		if g == nil {
			t.Fatal("native group lookup failed")
		}
		f.groups[spec.name] = g
	}
	add := func(name string, first uint32) {
		if f.s.MapGroups.Sub57C130([]uint32{first, 0}, f.groups[name].Index()) != 1 {
			t.Fatal("cannot allocate native group item")
		}
	}
	for _, i := range []int{0, 1, 0, 3, 2} {
		add("leaf-a", f.units[i].Extent)
	}
	add("leaf-a", 0xe2222222)
	for _, i := range []int{4, 5} {
		add("leaf-b", f.units[i].Extent)
	}
	add("leaf-b", 0xe3333333)
	add("middle", 0xf2222222)
	add("middle", f.groups["leaf-b"].Index())
	add("waypoints", f.units[6].Extent)
	add("walls", f.units[6].Extent)
	for _, name := range []string{"waypoints", "walls", "empty", "middle", "leaf-a"} {
		add("root", f.groups[name].Index())
	}
	for name, g := range f.groups {
		snapshot("group-"+name, g.C(), unsafe.Sizeof(*g))
		for i, it := 0, g.First(); it != nil; i, it = i+1, it.Next() {
			snapshot(fmt.Sprintf("item-%s-%d", name, i), it.C(), unsafe.Sizeof(*it))
		}
	}
	for _, unit := range f.units {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
			t.Fatal("native object is not above 4 GiB")
		}
		f.objects = append(f.objects, scriptNSRoamFlagBytes515C80(unsafe.Pointer(unit), unsafe.Sizeof(*unit)))
	}
	return f
}

func scriptNSGroupDeleteProgram(t *testing.T, handle int32) []byte {
	t.Helper()
	var data bytes.Buffer
	word := func(v uint32) {
		if err := binary.Write(&data, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	data.WriteString("SCRIPT03STRG")
	word(0)
	data.WriteString("CODE")
	word(3)
	for i, name := range []string{"GLOBAL", "GLOBAL", "test_group_delete"} {
		data.WriteString("FUNC")
		word(uint32(len(name)))
		data.WriteString(name)
		word(0)
		word(0)
		data.WriteString("SYMB")
		word(0)
		word(0)
		code := []uint32{uint32(asm.OpReturn0)}
		if i == 2 {
			code = []uint32{uint32(asm.OpPushInt), uint32(handle), uint32(asm.OpCallBuiltin), uint32(asm.BuiltinGroupDelete), uint32(asm.OpReturn0)}
		}
		data.WriteString("DATA")
		word(uint32(len(code) * 4))
		for _, v := range code {
			word(v)
		}
	}
	data.WriteString("DONE")
	return data.Bytes()
}

func (f *scriptNSGroupDeleteFixture) reset() {
	for i, unit := range f.units {
		copy(unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit))), f.objects[i])
	}
	f.s.Objs.DeletedList = f.sentinel
}

func (f *scriptNSGroupDeleteFixture) check(t *testing.T, route string, members []int, frame uint32) {
	t.Helper()
	want := make([][]byte, len(f.units))
	for i := range f.units {
		want[i] = bytes.Clone(f.objects[i])
	}
	head := f.sentinel
	for _, i := range members {
		var obj server.Object
		copy(unsafe.Slice((*byte)(unsafe.Pointer(&obj)), int(unsafe.Sizeof(obj))), want[i])
		obj.ObjFlags |= object.FlagDestroyed
		obj.DeletedAt, obj.DeletedNext = frame, head
		want[i] = scriptNSRoamFlagBytes515C80(unsafe.Pointer(&obj), unsafe.Sizeof(obj))
		head = f.units[i]
	}
	if f.s.Objs.DeletedList != head || f.s.Objs.List != f.units[0] || f.s.Objs.Pending != nil || f.s.Frame() != frame {
		t.Errorf("%s changed deletion head, active/pending list, or game frame", route)
	}
	for i, unit := range f.units {
		if !bytes.Equal(want[i], scriptNSRoamFlagBytes515C80(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) {
			t.Errorf("%s object %d: flags=%x deletedAt=%x next=%p; unexpected native record bytes", route, i, unit.ObjFlags, unit.DeletedAt, unit.DeletedNext)
		}
	}
	for _, r := range f.records {
		if !bytes.Equal(r.before, scriptNSRoamFlagBytes515C80(r.ptr, r.size)) {
			t.Errorf("%s changed unrelated native record %s", route, r.name)
		}
	}
}

func TestScriptNSObjectGroupDeleteMatchesClassicBuiltin(t *testing.T) {
	for _, spec := range []struct {
		name    string
		members []int
	}{
		{"root", []int{2, 0, 1, 5, 4}}, {"middle", []int{5, 4}},
		{"leaf-a", []int{2, 0, 1}}, {"leaf-b", []int{5, 4}}, {"empty", nil},
	} {
		for _, class := range []object.Class{object.ClassMonster, object.ClassMonster | object.ClassSimple, object.ClassMissile, object.ClassWeapon, object.ClassSimple} {
			for _, frame := range []uint32{0, 64, math.MaxUint32} {
				t.Run(fmt.Sprintf("%s/class-%x/frame-%x", spec.name, class, frame), func(t *testing.T) {
					f := scriptNSGroupDeleteNew(t, class, frame)
					group := f.groups[spec.name]
					var api ns4.ObjGroup = f.s.noxScriptP().ObjectGroupByHandle(nsObjGroup{s: f.s, g: group})
					if api == nil {
						t.Fatal("public native-width group resolver returned nil")
					}
					var panicked any
					func() { defer func() { panicked = recover() }(); api.Delete() }()
					if panicked != nil {
						t.Errorf("modern group Delete panicked: %v", panicked)
					} else {
						f.check(t, "modern", spec.members, frame)
						api.Delete()
						f.check(t, "modern repeated", spec.members, frame)
					}
					f.reset()
					if err := f.s.Server.NoxScriptVM.ReadScript(bytes.NewReader(scriptNSGroupDeleteProgram(t, int32(group.Index())))); err != nil {
						t.Fatal(err)
					}
					for i := 0; i < 2; i++ {
						if err := f.s.Server.NoxScriptVM.CallByIndex(2, nil, nil); err != nil {
							t.Fatal(err)
						}
						f.check(t, fmt.Sprintf("classic-%d", i), spec.members, frame)
					}
				})
			}
		}
	}
}

func TestScriptNSObjectGroupDeleteEmptyAndNonObjectGroups(t *testing.T) {
	for _, name := range []string{"empty", "waypoints", "walls", "missing"} {
		t.Run(name, func(t *testing.T) {
			f := scriptNSGroupDeleteNew(t, object.ClassMonster, 123)
			var api ns4.ObjGroup = nsObjGroup{s: f.s, g: f.groups[name]}
			var panicked any
			func() { defer func() { panicked = recover() }(); api.Delete() }()
			if panicked != nil {
				t.Errorf("empty/non-object group Delete panicked: %v", panicked)
				return
			}
			f.check(t, "empty/non-object", nil, 123)
		})
	}
}
