package opennox

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	ns4 "github.com/opennox/noxscript/ns/v4"
	"github.com/opennox/noxscript/ns/v4/subclass"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Type records are the real Go-owned registry metadata; only the instance
// control is C-owned. Do not put the metadata's Go pointers/strings in C memory.
// The expected masks below come from the ordered name tables independently of
// ParseSubClass. Overlapping class namespaces intentionally remain bit-based.
func TestScriptNSObjectTypeHasSubclassRegistryAndInstanceParity(t *testing.T) {
	s := scriptNSRoamFlagServer515C80(t)
	s.Types.Free()
	previous := ns4.Runtime()
	ns4.SetRuntime(s.noxScriptP())
	t.Cleanup(func() { ns4.SetRuntime(previous) })
	masks := make(map[string]uint32)
	for _, table := range object.SubClassNames {
		for bit, name := range table {
			if name == "" || name == "NULL" {
				continue
			}
			if _, ok := masks[name]; !ok {
				masks[name] = uint32(1) << bit
			}
		}
	}
	names := make([]string, 0, len(masks))
	for name := range masks {
		names = append(names, name)
	}
	sort.Strings(names)
	queries := []string{
		"", "NULL", "null", "NO_SUCH_SUBCLASS", "\x00", " SHOPKEEPER", "SHOPKEEPER ",
		"NPC+SHOPKEEPER", "NPC | SHOPKEEPER", "0x8", "2147483648", "상점", "NPC\x00ignored",
	}
	for _, name := range names {
		queries = append(queries, name, strings.ToLower(name))
	}
	var fixtures []struct {
		name  string
		class things.Class
		sub   []things.SubClass
	}
	fixtures = append(fixtures, struct {
		name  string
		class things.Class
		sub   []things.SubClass
	}{"empty", "SIMPLE", nil})
	for i, table := range object.SubClassNames {
		var sub []things.SubClass
		for _, name := range table {
			if name != "" && name != "NULL" {
				sub = append(sub, things.SubClass(name))
			}
		}
		fixtures = append(fixtures, struct {
			name  string
			class things.Class
			sub   []things.SubClass
		}{fmt.Sprintf("namespace-%d", i), "SIMPLE", sub})
	}
	for _, class := range []things.Class{"SIMPLE", "PLAYER", "WEAPON", "ARMOR", "MONSTER"} {
		fixtures = append(fixtures, struct {
			name  string
			class things.Class
			sub   []things.SubClass
		}{"overlap-" + string(class), class, []things.SubClass{"FLAG", "SHOPKEEPER", "NPC"}})
	}
	for fi, spec := range fixtures {
		name := fmt.Sprintf("ScriptSubclassFixture%d", fi)
		definition := &things.Thing{Name: name, Class: []things.Class{spec.class}, SubClass: spec.sub}
		if spec.class == "MONSTER" {
			definition.OnUpdate = &things.ProcFunc{Name: "MonsterUpdate"}
		}
		if err := s.Types.ReadObjectType(definition); err != nil {
			t.Fatal(err)
		}
		typ := s.Types.ByID(name)
		if typ == nil {
			t.Fatal("real registry type was not created")
		}
		wantMask := uint32(0)
		for _, sub := range spec.sub {
			wantMask |= masks[string(sub)]
		}
		if uint32(typ.SubClass()) != wantMask {
			t.Fatalf("metadata mask %s = %x, want independently assembled %x", spec.name, typ.SubClass(), wantMask)
		}
		unit, free := alloc.New(server.Object{})
		t.Cleanup(free)
		*unit = server.Object{ObjClass: typ.Class(), ObjSubClass: typ.SubClass(), TypeInd: uint16(typ.Ind()), ObjFlags: object.FlagEnabled, Extent: 0xf1234500 + uint32(fi)}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(typ)) <= math.MaxUint32 || uintptr(unsafe.Pointer(unit)) <= math.MaxUint32) {
			t.Fatalf("native type/instance addresses must exceed 4 GiB: %p/%p", typ, unit)
		}
		typeBefore := scriptNSRoamFlagBytes515C80(unsafe.Pointer(typ), unsafe.Sizeof(*typ))
		unitBefore := scriptNSRoamFlagBytes515C80(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
		var records [][]byte
		var pointers []unsafe.Pointer
		var sizes []uintptr
		if h := typ.Health(); h != nil {
			pointers = append(pointers, h.C())
			sizes = append(sizes, unsafe.Sizeof(*h))
		}
		if typ.UpdateData != nil {
			pointers = append(pointers, typ.UpdateData)
			sizes = append(sizes, typ.UpdateDataSize)
		}
		for i, ptr := range pointers {
			records = append(records, scriptNSRoamFlagBytes515C80(ptr, sizes[i]))
		}
		var direct ns4.ObjType = nsObjType{s: s, t: typ}
		byName := ns4.ObjectType(strings.ToUpper(name))
		byIndex := ns4.ObjectTypeByInd(typ.Ind())
		if byName == nil || byIndex == nil || byName.Index() != typ.Ind() || byIndex.Name() != name {
			t.Fatal("public type registry lookup failed")
		}
		instance := nsObj{Object: asObjectS(unit)}
		for qi, query := range queries {
			t.Run(fmt.Sprintf("%s/query-%03d-%q", spec.name, qi, query), func(t *testing.T) {
				mask := masks[strings.ToUpper(query)]
				want := wantMask&mask != 0
				if got := instance.HasSubclass(subclass.SubClass(query)); got != want {
					t.Fatalf("existing instance control = %t, want %t", got, want)
				}
				for _, route := range []struct {
					name string
					typ  ns4.ObjType
				}{{"direct", direct}, {"public-name", byName}, {"public-index", byIndex}} {
					var got bool
					var panicked any
					func() {
						defer func() { panicked = recover() }()
						got = route.typ.HasSubclass(subclass.SubClass(query))
					}()
					if panicked != nil || got != want {
						t.Errorf("%s type subclass query = %t, want %t; panic=%v", route.name, got, want, panicked)
					}
					if !bytes.Equal(typeBefore, scriptNSRoamFlagBytes515C80(unsafe.Pointer(typ), unsafe.Sizeof(*typ))) ||
						!bytes.Equal(unitBefore, scriptNSRoamFlagBytes515C80(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) {
						t.Errorf("%s subclass query mutated type/instance metadata", route.name)
					}
					for i, ptr := range pointers {
						if !bytes.Equal(records[i], scriptNSRoamFlagBytes515C80(ptr, sizes[i])) {
							t.Errorf("%s subclass query mutated auxiliary type record %d", route.name, i)
						}
					}
				}
			})
		}
	}
}

func TestScriptNSObjectTypeHasSubclassInvalidNameBeforeTypeRead(t *testing.T) {
	var typ ns4.ObjType = nsObjType{}
	for _, name := range []subclass.SubClass{"NO_SUCH_SUBCLASS", " SHOPKEEPER", "NPC+SHOPKEEPER", "0x8", "NPC\x00ignored"} {
		t.Run(string(name), func(t *testing.T) {
			var got bool
			var panicked any
			func() { defer func() { panicked = recover() }(); got = typ.HasSubclass(name) }()
			if panicked != nil || got {
				t.Fatalf("invalid name did not return false before metadata access: result=%t panic=%v", got, panicked)
			}
		})
	}
}
