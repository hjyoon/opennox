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
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// The modern interface and the existing classic builtin both reach the live
// native-width object resolver and real monster action-stack implementation.
// Only the inputs below are synthetic: no lookup, stack, or roam callback is
// replaced, and no stock map, script, or expected image is changed.
type scriptNSGroupWanderFixture512930 struct {
	s       *Server
	units   []*server.Object
	updates []*server.MonsterUpdateData
	groups  map[string]*server.MapGroup
	records []scriptNSRoamFlagRecord515CB0
}

func scriptNSGroupWanderFixtureNew512930(t *testing.T, flag uint32, initial ai.ActionType) *scriptNSGroupWanderFixture512930 {
	t.Helper()
	f := &scriptNSGroupWanderFixture512930{s: scriptNSRoamFlagServer515C80(t), groups: make(map[string]*server.MapGroup)}
	noxflags.SetGame(noxflags.GameHost)
	oldEngine := noxflags.GetEngine() & noxflags.EngineShowAI
	noxflags.UnsetEngine(noxflags.EngineShowAI)
	t.Cleanup(func() { noxflags.SetEngine(oldEngine) })
	f.s.SetFrame(0xffffffff)
	if !f.s.Objs.Init(8) {
		t.Fatal("cannot initialize native object allocator")
	}
	t.Cleanup(f.s.Objs.FreeObjects)
	f.s.MapGroups.Init()
	t.Cleanup(f.s.MapGroups.Free)
	waypoint, freeWaypoint := alloc.New(server.Waypoint{})
	t.Cleanup(freeWaypoint)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(waypoint)) <= math.MaxUint32 {
		t.Fatal("native reused waypoint is not above 4 GiB")
	}
	for i := 0; i < 8; i++ {
		unit := f.s.Objs.NewObject(&server.ObjectType{})
		update, free := alloc.New(server.MonsterUpdateData{})
		t.Cleanup(free)
		unit.ObjClass, unit.ObjFlags = object.ClassMonster, object.FlagEnabled
		unit.Extent = 0xe1234500 + uint32(i)
		unit.UpdateData = unsafe.Pointer(update)
		*update = server.MonsterUpdateData{
			AIStackInd: 0, Field333: flag, Field124: 51, Field137: 93,
			Aggression: .01, RetreatLevel: .98, ResumeLevel: 1,
			Field2: 17, Field67: 18, Field74: 19, Field91: 20,
			Field120_1: 21, Field120_2: 22, Field120_3: 23,
		}
		update.AIStack[0] = server.AIStackItem{Action: uint32(initial), Args: [4]uintptr{101, 102, 103, 104}}
		update.AIStack[1] = server.AIStackItem{Action: uint32(ai.ACTION_ROAM), Args: [4]uintptr{uintptr(unsafe.Pointer(waypoint)), 202, 0xabcdef00, 204}, Field5: 77}
		if unsafe.Sizeof(uintptr(0)) == 8 {
			update.AIStack[1].Args[2] |= uintptr(uint64(0x12345678) << 32)
			if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unsafe.Pointer(update)) <= math.MaxUint32 {
				t.Fatalf("native records are not above 4 GiB: unit=%p update=%p", unit, update)
			}
		}
		if i != 0 {
			f.units[i-1].ObjNext = unit
		}
		f.units, f.updates = append(f.units, unit), append(f.updates, update)
	}
	f.units[1].ObjClass, f.units[1].UpdateData = object.ClassPlayer, nil
	f.units[2].ObjFlags |= object.FlagDead
	f.units[3].ObjFlags |= object.FlagDestroyed
	f.units[4].ObjFlags |= object.FlagNoUpdate // This is not the original dead gate.
	f.units[5].ObjClass |= object.ClassPlayer
	f.units[6].ObjClass, f.units[6].UpdateData = 0, nil
	f.s.Objs.List = f.units[0]
	t.Cleanup(func() { f.s.Objs.List = nil })
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
		if g == nil || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(g.C()) <= math.MaxUint32) {
			t.Fatal("native group lookup/address failed")
		}
		f.groups[spec.name] = g
	}
	add := func(name string, first, second uint32) {
		if f.s.MapGroups.Sub57C130([]uint32{first, second}, f.groups[name].Index()) != 1 {
			t.Fatal("cannot allocate native group item")
		}
	}
	for _, i := range []int{0, 1, 0, 4, 2, 6} {
		add("leaf-a", f.units[i].Extent, 0)
	}
	for _, id := range []uint32{f.units[5].Extent, f.units[3].Extent, 0xe2220000} {
		add("leaf-b", id, 0)
	}
	add("waypoints", f.units[7].Extent, 0)
	add("walls", f.groups["leaf-b"].Index(), 1)
	add("middle", f.groups["leaf-b"].Index(), 0)
	add("middle", 0xf2220000, 0)
	for _, name := range []string{"leaf-a", "middle", "empty", "waypoints", "walls"} {
		add("root", f.groups[name].Index(), 0)
	}
	add("root", 0xf3330000, 0)
	// Sub57C130 prepends entries. Preserve that malformed first-child input
	// separately: MapGroupType faults before either command can run. A second
	// group gives the public handle resolver a valid first child while retaining
	// the nested missing IDs, duplicate objects, and non-object group members.
	f.groups["unresolvable"] = f.groups["root"]
	validRootID := uint32(0xf5555555)
	if f.s.MapGroups.MapLoadAddGroup57C0C0("valid-root", validRootID, byte(server.MapGroupGroups)) != 1 {
		t.Fatal("cannot allocate valid native root group")
	}
	f.groups["root"] = f.s.MapGroups.GroupByInd(int(int32(validRootID)))
	if f.groups["root"] == nil || (unsafe.Sizeof(uintptr(0)) == 8 && uintptr(f.groups["root"].C()) <= math.MaxUint32) {
		t.Fatal("valid root native address/lookup failed")
	}
	add("root", 0xf3330000, 0)
	for _, name := range []string{"walls", "waypoints", "empty", "middle", "leaf-a"} {
		add("root", f.groups[name].Index(), 0)
	}
	snapshot := func(name string, ptr unsafe.Pointer, size uintptr) {
		f.records = append(f.records, scriptNSRoamFlagRecord515CB0{name, ptr, size, scriptNSRoamFlagBytes515C80(ptr, size)})
	}
	for i, unit := range f.units {
		snapshot(fmt.Sprintf("unit-%d", i), unsafe.Pointer(unit), unsafe.Sizeof(*unit))
	}
	snapshot("waypoint", unsafe.Pointer(waypoint), unsafe.Sizeof(*waypoint))
	for name, g := range f.groups {
		snapshot("group-"+name, g.C(), unsafe.Sizeof(*g))
		for i, it := 0, g.First(); it != nil; i, it = i+1, it.Next() {
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(it.C()) <= math.MaxUint32 {
				t.Fatal("native group item is not above 4 GiB")
			}
			snapshot(fmt.Sprintf("item-%s-%d", name, i), it.C(), unsafe.Sizeof(*it))
		}
	}
	return f
}

func scriptNSGroupWanderProgram512930(t *testing.T, handle int32) []byte {
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
	for i, name := range []string{"GLOBAL", "GLOBAL", "test_group_wander"} {
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
			code = []uint32{uint32(asm.OpPushInt), uint32(handle), uint32(asm.OpCallBuiltin), uint32(asm.BuiltinGroupWander), uint32(asm.OpReturn0)}
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

func TestScriptNSObjectGroupWander512930MatchesClassicBuiltin(t *testing.T) {
	for _, spec := range []struct {
		name    string
		members []int
	}{
		{"root", []int{0, 4, 5}}, {"leaf-a", []int{0, 4}}, {"leaf-b", []int{5}},
		{"empty", nil}, {"waypoints", nil}, {"walls", nil}, {"missing", nil},
	} {
		for _, flag := range []uint32{0, 0x80, 0xdeadbea5, 0xffffffff} {
			for _, initial := range []ai.ActionType{ai.ACTION_IDLE, ai.ACTION_WAIT, ai.ACTION_ROAM} {
				t.Run(fmt.Sprintf("%s/flag-%08x/initial-%d", spec.name, flag, initial), func(t *testing.T) {
					f := scriptNSGroupWanderFixtureNew512930(t, flag, initial)
					baseline, want := make([][]byte, len(f.updates)), make([][]byte, len(f.updates))
					for i, update := range f.updates {
						baseline[i] = scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))
						expected := *update
						for _, member := range spec.members {
							if i != member {
								continue
							}
							expected.AIStackInd = 1
							expected.AIStack[0].Action, expected.AIStack[0].Field5, expected.AIStack[0].Args[0] = uint32(ai.ACTION_REPORT), 0, 10
							expected.AIStack[1].Action, expected.AIStack[1].Field5, expected.AIStack[1].Args[0] = uint32(ai.ACTION_ROAM), 0, 0
							expected.AIStack[1].Args[2] = expected.AIStack[1].Args[2]&^uintptr(0xff) | uintptr(uint8(flag))
							expected.Field2, expected.Field67, expected.Field74, expected.Field91 = 0, 0, 0, 0
							expected.Field120_1, expected.Field120_2, expected.Field120_3 = 0, 0, 0
							expected.Field124, expected.Field137 = 0xffffffff, 0xffffffff
						}
						want[i] = scriptNSRoamFlagBytes515C80(unsafe.Pointer(&expected), unsafe.Sizeof(expected))
					}
					check := func(route string) {
						for i, update := range f.updates {
							if !bytes.Equal(want[i], scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))) {
								t.Errorf("%s update bytes differ: unit=%d stack=%+v", route, i, update.AIStack[:2])
							}
						}
						for _, r := range f.records {
							if !bytes.Equal(r.before, scriptNSRoamFlagBytes515C80(r.ptr, r.size)) {
								t.Errorf("%s changed object, group, item, or waypoint bytes: %s", route, r.name)
							}
						}
						if f.s.AI.StackChanged != (len(spec.members) != 0) {
							t.Errorf("%s stack-change flag = %v", route, f.s.AI.StackChanged)
						}
					}
					g := f.groups[spec.name]
					var api ns4.ObjGroup = nsObjGroup{s: f.s, g: g}
					if g != nil && (spec.name == "root" || spec.name == "leaf-a" || spec.name == "leaf-b" || spec.name == "empty") {
						api = f.s.noxScriptP().ObjectGroupByHandle(api)
						if api == nil {
							t.Fatal("public group handle lookup failed")
						}
					}
					var panicked any
					func() { defer func() { panicked = recover() }(); api.Wander() }()
					if panicked != nil {
						t.Errorf("modern group Wander panicked: %v", panicked)
					} else {
						check("modern")
					}
					for i, update := range f.updates {
						copy(unsafe.Slice((*byte)(unsafe.Pointer(update)), int(unsafe.Sizeof(*update))), baseline[i])
					}
					f.s.AI.StackChanged = false
					handle := int32(-1)
					if g != nil {
						handle = int32(g.Index())
					}
					if err := f.s.Server.NoxScriptVM.ReadScript(bytes.NewReader(scriptNSGroupWanderProgram512930(t, handle))); err != nil {
						t.Fatal(err)
					}
					if err := f.s.Server.NoxScriptVM.CallByIndex(2, nil, nil); err != nil {
						t.Fatal(err)
					}
					check("classic")
				})
			}
		}
	}
}

func TestScriptNSObjectGroupWander512930UnresolvableFirstChildPreservesLookupFault(t *testing.T) {
	f := scriptNSGroupWanderFixtureNew512930(t, 0xdeadbea5, ai.ACTION_WAIT)
	g := f.groups["unresolvable"]
	var lookupPanic any
	func() {
		defer func() { lookupPanic = recover() }()
		f.s.noxScriptP().ObjectGroupByHandle(nsObjGroup{s: f.s, g: g})
	}()
	if lookupPanic == nil {
		t.Fatal("missing first child no longer preserves the existing public group lookup fault")
	}
	if err := f.s.Server.NoxScriptVM.ReadScript(bytes.NewReader(scriptNSGroupWanderProgram512930(t, int32(g.Index())))); err != nil {
		t.Fatal(err)
	}
	if err := f.s.Server.NoxScriptVM.CallByIndex(2, nil, nil); err == nil {
		t.Fatal("classic group resolver no longer reports the malformed first-child fault")
	}
	for _, r := range f.records {
		if !bytes.Equal(r.before, scriptNSRoamFlagBytes515C80(r.ptr, r.size)) {
			t.Fatalf("lookup changed native record: %s", r.name)
		}
	}
	if f.s.AI.StackChanged {
		t.Fatal("malformed lookup reached an action-stack callback")
	}
}
