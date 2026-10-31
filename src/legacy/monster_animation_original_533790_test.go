package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type animationInput533790 struct {
	Head        int8      `json:"head"`
	Action      uint32    `json:"action"`
	Status      uint32    `json:"status"`
	Enemy       bool      `json:"enemy"`
	OtherStatus uint32    `json:"other_status"`
	OtherEnemy  bool      `json:"other_enemy"`
	Class       uint32    `json:"class"`
	Flags       uint32    `json:"flags"`
	Type        uint16    `json:"type_index"`
	Caches      [4]uint32 `json:"caches"`
	Lookup      [4]uint16 `json:"lookup"`
	Found       [4]bool   `json:"found"`
	Repeat      int       `json:"repeat"`
	Mutation    *struct {
		Name       string `json:"name"`
		Occurrence int    `json:"occurrence"`
		Effects    []struct {
			Field string          `json:"field"`
			Value json.RawMessage `json:"value"`
		} `json:"effects"`
	} `json:"mutation"`
}
type animationPredicateEntry533790 struct {
	Name   string    `json:"name"`
	Call   int       `json:"call"`
	Caches [4]uint32 `json:"caches_before"`
	Type   uint16    `json:"unit_type"`
	Class  uint32    `json:"unit_class"`
	Status uint32    `json:"entry_status"`
	Live   string    `json:"live_update"`
}
type animationLookup533790 struct {
	Name       string    `json:"name"`
	Call       int       `json:"call"`
	Occurrence int       `json:"occurrence"`
	Before     [4]uint32 `json:"caches_before"`
	Found      bool      `json:"found"`
	Index      *uint16   `json:"index"`
	After      [4]uint32 `json:"caches_after_declared_mutations"`
}
type animationPublication533790 struct {
	Index int    `json:"index"`
	Value uint32 `json:"value"`
	Call  int    `json:"call"`
}
type animationMutation533790 struct {
	Field string          `json:"field"`
	Value json.RawMessage `json:"value"`
	Call  int             `json:"call"`
}
type animationRow533790 struct {
	Name   string               `json:"name"`
	Input  animationInput533790 `json:"input"`
	Output struct {
		Returns      []uint32                        `json:"returns"`
		Caches       [4]uint32                       `json:"caches"`
		Predicates   []animationPredicateEntry533790 `json:"predicate_entries"`
		Lookups      []animationLookup533790         `json:"lookups"`
		Publications []animationPublication533790    `json:"publications"`
		Mutations    []animationMutation533790       `json:"declared_mutations"`
		Type         uint16                          `json:"unit_type_after"`
		Class        uint32                          `json:"unit_class_after"`
		Live         string                          `json:"live_update_after"`
		Status       uint32                          `json:"entry_status_after"`
		OtherStatus  uint32                          `json:"other_status_after"`
		Enemy        bool                            `json:"entry_enemy_after"`
		OtherEnemy   bool                            `json:"other_enemy_after"`
		Action       *uint32                         `json:"entry_action_after"`
		Guards       bool                            `json:"complete_remaining_heap_stack_global_guards_code_tables_nonvolatile_and_FPU_preserved"`
	} `json:"output"`
}
type animationGuard533790[T any] struct {
	Before [64]byte
	Value  T
	After  [64]byte
}
type animationRegion533790 struct {
	Ptr  unsafe.Pointer
	Size uintptr
}
type animationFixture533790 struct {
	t            *testing.T
	srv          *server.Server
	unit         *animationGuard533790[server.Object]
	updates      [2]*animationGuard533790[server.MonsterUpdateData]
	target       *animationGuard533790[server.Object]
	cache        *animationGuard533790[[3]uint32]
	regions      []animationRegion533790
	row          *animationRow533790
	call         int
	lookupCounts map[string]int
	predicates   []animationPredicateEntry533790
	lookups      []animationLookup533790
	publications []animationPublication533790
	mutations    []animationMutation533790
	before       [][]byte
	mimicGuard   []byte
	registryKey  string
	recording    bool
	runtime      server.MonsterAnimationNativeRuntime533790
}

var animationNames533790 = [4]string{"Mimic", "CarnivorousPlant", "Zombie", "VileZombie"}

func animationSnapshot533790(p unsafe.Pointer, n uintptr) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(p), int(n))...)
}
func animationAllocate533790[T comparable](t *testing.T, f *animationFixture533790) *animationGuard533790[T] {
	p, free := alloc.New(animationGuard533790[T]{})
	t.Cleanup(free)
	f.regions = append(f.regions, animationRegion533790{unsafe.Pointer(p), unsafe.Sizeof(*p)})
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
		t.Fatal("all actual native animation allocations must exceed4GiB")
	}
	return p
}
func animationNewFixture533790(t *testing.T) *animationFixture533790 {
	f := &animationFixture533790{t: t, srv: npcReflectLegacyServer4E17B0(t)}
	f.unit = animationAllocate533790[server.Object](t, f)
	for i := range f.updates {
		f.updates[i] = animationAllocate533790[server.MonsterUpdateData](t, f)
	}
	f.target = animationAllocate533790[server.Object](t, f)
	f.cache = animationAllocate533790[[3]uint32](t, f)
	oldCache := statePredicateOriginalSwapMimicCache534840(0)
	t.Cleanup(func() { statePredicateOriginalSwapMimicCache534840(oldCache) })
	owner := GetServer
	GetServer = func() Server {
		if f.recording {
			f.lookup("Mimic")
		}
		return owner()
	}
	t.Cleanup(func() { GetServer = owner })
	oldTypes := zombieEntryTypes534A10
	zombieEntryTypes534A10 = func() server.TypePredicateRuntime534A10 {
		return server.TypePredicateRuntime534A10{
			Read: func(i int) uint32 { return f.cache.Value[i] },
			Write: func(i int, v uint32) {
				f.cache.Value[i] = v
				f.publications = append(f.publications, animationPublication533790{i + 1, v, f.call})
			},
			Lookup: func(name string) uint32 { f.lookup(name); return uint32(f.srv.Types.IndByID(name)) },
		}
	}
	t.Cleanup(func() { zombieEntryTypes534A10 = oldTypes })
	wrap := func(index int, name string) func(*server.Object) uint32 {
		return func(unit *server.Object) uint32 {
			if unit != &f.unit.Value {
				t.Fatal("actual predicate received wrong native object")
			}
			live := "entry"
			if unit.UpdateData == unsafe.Pointer(&f.updates[1].Value) {
				live = "other"
			}
			f.predicates = append(f.predicates, animationPredicateEntry533790{name, f.call, f.caches(), unit.TypeInd, uint32(unit.ObjClass), uint32(f.updates[0].Value.StatusFlags), live})
			return animationPredicatePublicC533790(index, unit)
		}
	}
	f.runtime = server.MonsterAnimationNativeRuntime533790{Mimic: wrap(0, "Mimic"), Plant: wrap(1, "CarnivorousPlant"), Zombie: wrap(2, "ZombiePair")}
	oldFactory := monsterAnimationRuntimeNative533790
	monsterAnimationRuntimeNative533790 = func() server.MonsterAnimationNativeRuntime533790 { return f.runtime }
	t.Cleanup(func() { monsterAnimationRuntimeNative533790 = oldFactory })
	t.Cleanup(f.srv.FreeObjectTypes)
	return f
}
func (f *animationFixture533790) caches() [4]uint32 {
	return [4]uint32{*(*uint32)(statePredicateOriginalMimicCachePointer534840()), f.cache.Value[0], f.cache.Value[1], f.cache.Value[2]}
}
func (f *animationFixture533790) setCache(i int, v uint32) {
	if i == 0 {
		statePredicateOriginalSwapMimicCache534840(v)
	} else {
		f.cache.Value[i-1] = v
	}
}
func (f *animationFixture533790) lookup(name string) {
	f.t.Helper()
	i := -1
	for n, v := range animationNames533790 {
		if v == name {
			i = n
			break
		}
	}
	if i < 0 {
		f.t.Fatal("unknown dictionary name", name)
	}
	f.lookupCounts[name]++
	e := animationLookup533790{Name: name, Call: f.call, Occurrence: f.lookupCounts[name], Before: f.caches(), Found: f.row.Input.Found[i]}
	m := f.row.Input.Mutation
	if m != nil && m.Name == name && (m.Occurrence == 0 && e.Occurrence == 1 || m.Occurrence == e.Occurrence) {
		for _, effect := range m.Effects {
			var number uint32
			var flag bool
			var text string
			switch effect.Field {
			case "unit_type", "unit_class", "entry_status", "other_status", "entry_action", "cache_0", "cache_1", "cache_2", "cache_3":
				if err := json.Unmarshal(effect.Value, &number); err != nil {
					f.t.Fatal(err)
				}
				switch effect.Field {
				case "unit_type":
					f.unit.Value.TypeInd = uint16(number)
				case "unit_class":
					f.unit.Value.ObjClass = object.Class(number)
				case "entry_status":
					f.updates[0].Value.StatusFlags = object.MonsterStatus(number)
				case "other_status":
					f.updates[1].Value.StatusFlags = object.MonsterStatus(number)
				case "entry_action":
					if f.row.Input.Head < 0 {
						f.t.Fatal("mutation outside original valid head")
					}
					f.updates[0].Value.AIStack[f.row.Input.Head].Action = number
				default:
					f.setCache(int(effect.Field[len(effect.Field)-1]-'0'), number)
				}
			case "live_update":
				if err := json.Unmarshal(effect.Value, &text); err != nil {
					f.t.Fatal(err)
				}
				if text == "other" {
					f.unit.Value.UpdateData = unsafe.Pointer(&f.updates[1].Value)
				} else if text == "entry" {
					f.unit.Value.UpdateData = unsafe.Pointer(&f.updates[0].Value)
				} else {
					f.t.Fatal("unknown update role")
				}
			case "entry_enemy", "other_enemy":
				if err := json.Unmarshal(effect.Value, &flag); err != nil {
					f.t.Fatal(err)
				}
				index := 0
				if effect.Field == "other_enemy" {
					index = 1
				}
				f.updates[index].Value.CurrentEnemy = nil
				if flag {
					f.updates[index].Value.CurrentEnemy = &f.target.Value
				}
			default:
				f.t.Fatal("unknown declared mutation", effect.Field)
			}
			f.mutations = append(f.mutations, animationMutation533790{effect.Field, effect.Value, f.call})
		}
	}
	e.After = f.caches()
	if e.Found {
		index := f.row.Input.Lookup[i]
		e.Index = &index
	}
	f.lookups = append(f.lookups, e)
	want := 0
	if e.Found {
		want = int(f.row.Input.Lookup[i])
	}
	if actual := f.srv.Types.IndByID(name); actual != want {
		f.t.Fatalf("real native dictionary %s=%d input WORD=%d", name, actual, want)
	}
}
func (f *animationFixture533790) prepare(row *animationRow533790) {
	f.t.Helper()
	f.row = row
	c := row.Input
	key := fmt.Sprint(c.Lookup, c.Found)
	if key != f.registryKey {
		f.srv.FreeObjectTypes()
		maximum := 0
		for _, v := range c.Lookup {
			if int(v) > maximum {
				maximum = int(v)
			}
		}
		for index := 1; index <= maximum; index++ {
			name := fmt.Sprintf("AnimationType%d", index)
			for n, value := range c.Lookup {
				if c.Found[n] && int(value) == index {
					name = animationNames533790[n]
				}
			}
			if err := f.srv.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
				f.t.Fatal(err)
			}
		}
		f.registryKey = key
	}
	for _, r := range f.regions {
		for i := range unsafe.Slice((*byte)(r.Ptr), int(r.Size)) {
			unsafe.Slice((*byte)(r.Ptr), int(r.Size))[i] = byte(i*31 + 19)
		}
	}
	f.unit.Value = server.Object{TypeInd: c.Type, ObjClass: object.Class(c.Class), ObjFlags: object.Flags(c.Flags), UpdateData: unsafe.Pointer(&f.updates[0].Value)}
	f.target.Value = server.Object{}
	for i := range f.updates {
		status, enemy := c.Status, c.Enemy
		if i == 1 {
			status, enemy = c.OtherStatus, c.OtherEnemy
		}
		f.updates[i].Value = server.MonsterUpdateData{AIStackInd: c.Head, StatusFlags: object.MonsterStatus(status)}
		if c.Head >= 0 {
			if int(c.Head) >= len(f.updates[i].Value.AIStack) {
				f.t.Fatal("original head outside pinned matrix")
			}
			f.updates[i].Value.AIStack[c.Head].Action = c.Action
		}
		if enemy {
			f.updates[i].Value.CurrentEnemy = &f.target.Value
		}
	}
	for i, v := range c.Caches {
		f.setCache(i, v)
	}
	f.lookupCounts = map[string]int{}
	f.predicates = []animationPredicateEntry533790{}
	f.lookups = []animationLookup533790{}
	f.publications = []animationPublication533790{}
	f.mutations = []animationMutation533790{}
	f.before = make([][]byte, len(f.regions))
	for i, r := range f.regions {
		f.before[i] = animationSnapshot533790(r.Ptr, r.Size)
	}
	f.mimicGuard = animationSnapshot533790(unsafe.Add(statePredicateOriginalMimicCachePointer534840(), -64), 132)
}
func animationPatchValue533790[T any](t *testing.T, expected [][]byte, regions []animationRegion533790, pointer *T, value T) {
	t.Helper()
	a := uintptr(unsafe.Pointer(pointer))
	n := unsafe.Sizeof(value)
	for i, r := range regions {
		base := uintptr(r.Ptr)
		if base <= a && a+n <= base+r.Size {
			offset := a - base
			copy(expected[i][offset:offset+n], unsafe.Slice((*byte)(unsafe.Pointer(&value)), int(n)))
			return
		}
	}
	t.Fatal("expected patch outside guarded native records")
}
func (f *animationFixture533790) check(control bool) {
	f.t.Helper()
	r := f.row
	o := r.Output
	if !o.Guards || len(o.Returns) != r.Input.Repeat {
		f.t.Fatal("invalid original guard/call contract")
	}
	if !control {
		if !reflect.DeepEqual(f.predicates, o.Predicates) {
			f.t.Errorf("actual predicate entry/order/state differs: got=%v original=%v", f.predicates, o.Predicates)
		}
		if !reflect.DeepEqual(f.lookups, o.Lookups) {
			f.t.Errorf("actual real-dictionary call/order/mutation differs: got=%v original=%v", f.lookups, o.Lookups)
		}
		if !reflect.DeepEqual(f.mutations, o.Mutations) {
			f.t.Errorf("declared dictionary effects differ: got=%v original=%v", f.mutations, o.Mutations)
		}
		if f.caches() != o.Caches {
			f.t.Errorf("actual caches=%08x original=%08x", f.caches(), o.Caches)
		}
		// The unchanged native C Mimic store is checked by its separately
		// qualified primitive, actual cache snapshots and lookup observations.
		// Observe every actual Go Plant/Zombie publication, without inventing
		// a C store-count observation that this fixture cannot obtain.
		pub := []animationPublication533790{}
		for _, p := range o.Publications {
			if p.Index != 0 {
				pub = append(pub, p)
			}
		}
		if !reflect.DeepEqual(f.publications, pub) {
			f.t.Errorf("actual Plant/Zombie publications=%v original=%v", f.publications, pub)
		}
		animationPatchValue533790(f.t, f.before, f.regions, &f.unit.Value.TypeInd, o.Type)
		animationPatchValue533790(f.t, f.before, f.regions, &f.unit.Value.ObjClass, object.Class(o.Class))
		link := unsafe.Pointer(&f.updates[0].Value)
		if o.Live == "other" {
			link = unsafe.Pointer(&f.updates[1].Value)
		} else if o.Live != "entry" {
			f.t.Fatal("unknown original live-update role")
		}
		animationPatchValue533790(f.t, f.before, f.regions, &f.unit.Value.UpdateData, link)
		for i := range f.updates {
			status, enemy := o.Status, o.Enemy
			if i == 1 {
				status, enemy = o.OtherStatus, o.OtherEnemy
			}
			animationPatchValue533790(f.t, f.before, f.regions, &f.updates[i].Value.StatusFlags, object.MonsterStatus(status))
			var ptr *server.Object
			if enemy {
				ptr = &f.target.Value
			}
			animationPatchValue533790(f.t, f.before, f.regions, &f.updates[i].Value.CurrentEnemy, ptr)
		}
		if o.Action != nil {
			if r.Input.Head < 0 {
				f.t.Fatal("unexpected original action output")
			}
			animationPatchValue533790(f.t, f.before, f.regions, &f.updates[0].Value.AIStack[r.Input.Head].Action, *o.Action)
		}
		for i := range f.cache.Value {
			animationPatchValue533790(f.t, f.before, f.regions, &f.cache.Value[i], o.Caches[i+1])
		}
		*(*uint32)(unsafe.Pointer(&f.mimicGuard[64])) = o.Caches[0]
	}
	for i, region := range f.regions {
		if !bytes.Equal(f.before[i], unsafe.Slice((*byte)(region.Ptr), int(region.Size))) {
			f.t.Errorf("native guarded unit/update/other/target/cache region%d differs beyond original outputs", i)
		}
	}
	if !bytes.Equal(f.mimicGuard, animationSnapshot533790(unsafe.Add(statePredicateOriginalMimicCachePointer534840(), -64), 132)) {
		f.t.Error("native Mimic cache or neighbor guards differ")
	}
}
func animationRunOriginal533790(t *testing.T, kind string, invoke func(*animationFixture533790) uint32) {
	path := os.Getenv("NOX_ANIMATION_ORIGINAL_REFERENCE")
	if path == "" {
		path = "../server/testdata/monster_animation_selector_original_533790.json.gz"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != "1138dde79ead136120908feb14740281833366b7ef95bd30b4ada4b73e1298a9" {
		t.Fatal("unsealed original animation reference")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Schema   string               `json:"schema"`
		Original string               `json:"original_sha256"`
		Control  uint16               `json:"control_word"`
		Entry    uint32               `json:"entry"`
		Rows     []animationRow533790 `json:"rows"`
	}
	if err = json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if ref.Schema != "original-monster-animation-selector-533790-v1" || ref.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || ref.Control != 0xe3f || ref.Entry != 0x533790 || len(ref.Rows) != 14840 {
		t.Fatal("original animation provenance/dimensions")
	}
	f := animationNewFixture533790(t)
	control := os.Getenv("NOX_ANIMATION_CONTROL_ONLY") == "true"
	calls := 0
	for _, row := range ref.Rows {
		t.Run(row.Name, func(t *testing.T) {
			f.t = t
			f.prepare(&row)
			if !control {
				f.recording = true
				defer func() { f.recording = false }()
				for call, want := range row.Output.Returns {
					f.call = call
					got := invoke(f)
					calls++
					if got != want {
						t.Errorf("native animation DWORD call%d=%#x original=%#x", call, got, want)
					}
				}
			}
			f.recording = false
			f.check(control)
		})
	}
	if !control && calls != 15064 {
		t.Fatal("actual root call count", calls)
	}
	t.Logf("original %s animation contexts=14840 calls=%d control-only=%t actual-native-predicates-real-dictionary-cached-state-and-guards=true", kind, calls, control)
}
func TestAnimationSelectorOriginalActualC533790(t *testing.T) {
	animationRunOriginal533790(t, "actual C", func(f *animationFixture533790) uint32 {
		return uint32(ccall.CallIntPtr(animationSelectorPublicC533790(), f.unit.Value.CObj()))
	})
}
