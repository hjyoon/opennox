package opennox

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/noxscript/ns/asm"
	ns4 "github.com/opennox/noxscript/ns/v4"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

type scriptNSRoamFlagRecord515CB0 struct {
	name   string
	ptr    unsafe.Pointer
	size   uintptr
	before []byte
}

type scriptNSRoamFlagGroupFixture515CB0 struct {
	s       *Server
	units   []*server.Object
	updates []*server.MonsterUpdateData
	groups  map[string]*server.MapGroup
	records []scriptNSRoamFlagRecord515CB0
}

func scriptNSRoamFlagGroups515CB0(t *testing.T) *scriptNSRoamFlagGroupFixture515CB0 {
	t.Helper()
	f := &scriptNSRoamFlagGroupFixture515CB0{s: scriptNSRoamFlagServer515C80(t), groups: make(map[string]*server.MapGroup)}
	noxflags.SetGame(noxflags.GameHost)
	f.s.MapGroups.Init()
	for i := 0; i < 6; i++ {
		unit, update := scriptNSRoamFlagFixture515C80(t)
		unit.Extent = 0xe1234500 + uint32(i)
		if i != 0 {
			f.units[i-1].ObjNext = unit
		}
		f.units = append(f.units, unit)
		f.updates = append(f.updates, update)
	}
	f.units[1].ObjClass = object.ClassPlayer
	f.units[3].ObjFlags |= object.FlagDestroyed
	f.units[5].UpdateData = nil
	f.s.Objs.List = f.units[0]
	t.Cleanup(func() { f.s.Objs.List = nil; f.s.MapGroups.Free() })
	for i, spec := range []struct {
		name string
		kind server.MapGroupKind
	}{
		{"root", server.MapGroupGroups}, {"leaf-a", server.MapGroupObjects},
		{"leaf-b", server.MapGroupObjects}, {"middle", server.MapGroupGroups},
		{"empty", server.MapGroupObjects}, {"waypoints", server.MapGroupWaypoints},
		{"walls", server.MapGroupWalls},
	} {
		id := uint32(0xf1234e00) + uint32(i)
		if f.s.MapGroups.MapLoadAddGroup57C0C0(spec.name, id, byte(spec.kind)) != 1 {
			t.Fatal("native group allocation failed")
		}
		g := f.s.MapGroups.GroupByInd(int(int32(id)))
		if g == nil || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(g.C()) <= math.MaxUint32) {
			t.Fatalf("native group lookup/allocation failed: id=%08x group=%p", id, g)
		}
		f.groups[spec.name] = g
	}
	add := func(name string, first, second uint32) {
		if f.s.MapGroups.Sub57C130([]uint32{first, second}, f.groups[name].Index()) != 1 {
			t.Fatal("native group item allocation failed")
		}
	}
	for _, i := range []int{0, 1, 0, 5} {
		add("leaf-a", f.units[i].Extent, 0)
	}
	for _, id := range []uint32{f.units[2].Extent, f.units[3].Extent, 0xe2220000} {
		add("leaf-b", id, 0)
	}
	add("waypoints", f.units[0].Extent, 0)
	add("walls", f.groups["leaf-a"].Index(), 1)
	for _, name := range []string{"leaf-b", "waypoints", "walls"} {
		add("middle", f.groups[name].Index(), 0)
	}
	add("middle", 0xf2220000, 0)
	for _, name := range []string{"leaf-a", "middle", "empty"} {
		add("root", f.groups[name].Index(), 0)
	}
	add("root", 0xf3330000, 0)
	snapshot := func(name string, ptr unsafe.Pointer, size uintptr) {
		f.records = append(f.records, scriptNSRoamFlagRecord515CB0{name, ptr, size, scriptNSRoamFlagBytes515C80(ptr, size)})
	}
	for i, unit := range f.units {
		snapshot(fmt.Sprintf("unit-%d", i), unsafe.Pointer(unit), unsafe.Sizeof(*unit))
	}
	for name, g := range f.groups {
		snapshot("group-"+name, g.C(), unsafe.Sizeof(*g))
		for i, it := 0, g.First(); it != nil; i, it = i+1, it.Next() {
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(it.C()) <= math.MaxUint32 {
				t.Fatalf("native item is not above 4 GiB: %p", it)
			}
			snapshot(fmt.Sprintf("item-%s-%d", name, i), it.C(), unsafe.Sizeof(*it))
		}
	}
	return f
}

func (f *scriptNSRoamFlagGroupFixture515CB0) checkRecords(t *testing.T) {
	t.Helper()
	for _, r := range f.records {
		if !bytes.Equal(r.before, scriptNSRoamFlagBytes515C80(r.ptr, r.size)) {
			t.Fatalf("SetRoamFlag mutated object/group/list bytes: %s", r.name)
		}
	}
}

func TestScriptNSObjectGroupRoamFlag515CB0MatchesClassicBuiltin(t *testing.T) {
	f := scriptNSRoamFlagGroups515CB0(t)
	baseline := make([][]byte, len(f.updates))
	for i, update := range f.updates {
		baseline[i] = scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))
	}
	reset := func() {
		for i, update := range f.updates {
			copy(unsafe.Slice((*byte)(unsafe.Pointer(update)), int(unsafe.Sizeof(*update))), baseline[i])
		}
	}
	for _, spec := range []struct {
		name    string
		members []int
	}{
		{"root", []int{0, 2}}, {"leaf-a", []int{0}}, {"leaf-b", []int{2}},
		{"empty", nil}, {"waypoints", nil}, {"walls", nil}, {"missing", nil},
	} {
		for _, flag := range []int32{0, 0x80, 0x128, -1, math.MinInt32, math.MaxInt32} {
			t.Run(fmt.Sprintf("%s/flag-%d", spec.name, flag), func(t *testing.T) {
				reset()
				g := f.groups[spec.name]
				var api ns4.ObjGroup = nsObjGroup{s: f.s, g: g}
				api.SetRoamFlag(int(flag))
				f.checkRecords(t)
				want := make([][]byte, len(f.updates))
				for i, update := range f.updates {
					want[i] = bytes.Clone(baseline[i])
					for _, member := range spec.members {
						if i == member {
							want[i][unsafe.Offsetof(update.Field333)] = uint8(flag)
						}
					}
					if !bytes.Equal(want[i], scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
						t.Fatalf("modern group route changed wrong update bytes: unit=%d Field333=%08x", i, update.Field333)
					}
				}
				reset()
				id := uint32(0xf4440000)
				if g != nil {
					id = g.Index()
				}
				program := scriptNSRoamFlagProgram515C80(t, asm.BuiltinGroupSetRoamFlag, int32(id), flag)
				if err := f.s.Server.NoxScriptVM.ReadScript(bytes.NewReader(program)); err != nil {
					t.Fatal(err)
				}
				if err := f.s.Server.NoxScriptVM.CallByIndex(2, nil, nil); err != nil {
					t.Fatal(err)
				}
				for i, update := range f.updates {
					if !bytes.Equal(want[i], scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
						t.Fatalf("classic group route differs: unit=%d Field333=%08x", i, update.Field333)
					}
				}
				f.checkRecords(t)
			})
		}
	}
}

func TestScriptNSObjectGroupRoamFlag515CB0NativeWideArgument(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-wide int case requires a 64-bit host")
	}
	f := scriptNSRoamFlagGroups515CB0(t)
	var api ns4.ObjGroup = nsObjGroup{s: f.s, g: f.groups["root"]}
	api.SetRoamFlag(int(int64(1)<<40 | 0x1a5))
	for i, update := range f.updates {
		want := uint32(0xdeadbeef)
		if i == 0 || i == 2 {
			want = 0xdeadbea5
		}
		if update.Field333 != want {
			t.Fatalf("native-wide flag/group membership differs: unit=%d got=%08x want=%08x", i, update.Field333, want)
		}
	}
	f.checkRecords(t)
}
