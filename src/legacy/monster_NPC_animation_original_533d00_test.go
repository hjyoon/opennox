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
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type NPCAnimationReturn533D00 struct {
	Role   string  `json:"role"`
	Offset uintptr `json:"offset"`
}
type NPCAnimationHelper533D00 struct {
	Entry    string `json:"entry"`
	Argument uint32 `json:"argument"`
	Call     int    `json:"call"`
}
type animationInput533D00 struct {
	Subclass     uint32    `json:"subclass"`
	Animation    string    `json:"animation"`
	Equip        uint32    `json:"equip"`
	Attack       uint8     `json:"attack"`
	TableProfile int       `json:"table_profile"`
	Head         int8      `json:"head"`
	Action       uint32    `json:"action"`
	Status       uint32    `json:"status"`
	Enemy        bool      `json:"enemy"`
	OtherStatus  uint32    `json:"other_status"`
	OtherEnemy   bool      `json:"other_enemy"`
	Class        uint32    `json:"class"`
	Flags        uint32    `json:"flags"`
	Type         uint16    `json:"type_index"`
	Caches       [4]uint32 `json:"caches"`
	Lookup       [4]uint16 `json:"lookup"`
	Found        [4]bool   `json:"found"`
	Repeat       int       `json:"repeat"`
	Mutation     *struct {
		Name       string `json:"name"`
		Occurrence int    `json:"occurrence"`
		Effects    []struct {
			Field string          `json:"field"`
			Value json.RawMessage `json:"value"`
		} `json:"effects"`
	} `json:"mutation"`
}
type animationPredicateEntry533D00 struct {
	Name   string    `json:"name"`
	Call   int       `json:"call"`
	Caches [4]uint32 `json:"caches_before"`
	Type   uint16    `json:"unit_type"`
	Class  uint32    `json:"unit_class"`
	Status uint32    `json:"entry_status"`
	Live   string    `json:"live_update"`
}
type animationLookup533D00 struct {
	Name       string    `json:"name"`
	Call       int       `json:"call"`
	Occurrence int       `json:"occurrence"`
	Before     [4]uint32 `json:"caches_before"`
	Found      bool      `json:"found"`
	Index      *uint16   `json:"index"`
	After      [4]uint32 `json:"caches_after_declared_mutations"`
}
type animationPublication533D00 struct {
	Index int    `json:"index"`
	Value uint32 `json:"value"`
	Call  int    `json:"call"`
}
type animationMutation533D00 struct {
	Field string          `json:"field"`
	Value json.RawMessage `json:"value"`
	Call  int             `json:"call"`
}
type animationRow533D00 struct {
	Name   string               `json:"name"`
	Input  animationInput533D00 `json:"input"`
	Output struct {
		ScratchAfter   string                          `json:"scratch_after"`
		AnimationAfter string                          `json:"entry_animation_after"`
		Helpers        []NPCAnimationHelper533D00      `json:"helper_entries"`
		Returns        []NPCAnimationReturn533D00      `json:"returns"`
		Caches         [4]uint32                       `json:"caches"`
		Predicates     []animationPredicateEntry533D00 `json:"predicate_entries"`
		Lookups        []animationLookup533D00         `json:"lookups"`
		Publications   []animationPublication533D00    `json:"publications"`
		Mutations      []animationMutation533D00       `json:"declared_mutations"`
		Type           uint16                          `json:"unit_type_after"`
		Class          uint32                          `json:"unit_class_after"`
		Live           string                          `json:"live_update_after"`
		Status         uint32                          `json:"entry_status_after"`
		OtherStatus    uint32                          `json:"other_status_after"`
		Enemy          bool                            `json:"entry_enemy_after"`
		OtherEnemy     bool                            `json:"other_enemy_after"`
		Action         *uint32                         `json:"entry_action_after"`
		Guards         bool                            `json:"complete_remaining_heap_stack_global_guards_code_tables_nonvolatile_and_FPU_preserved"`
	} `json:"output"`
}
type animationGuard533D00[T any] struct {
	Before [64]byte
	Value  T
	After  [64]byte
}
type animationRegion533D00 struct {
	Ptr  unsafe.Pointer
	Size uintptr
}
type animationFixture533D00 struct {
	t             *testing.T
	srv           *server.Server
	unit          *animationGuard533D00[server.Object]
	updates       [2]*animationGuard533D00[server.MonsterUpdateData]
	target        *animationGuard533D00[server.Object]
	cache         *animationGuard533D00[[3]uint32]
	regions       []animationRegion533D00
	row           *animationRow533D00
	call          int
	lookupCounts  map[string]int
	predicates    []animationPredicateEntry533D00
	lookups       []animationLookup533D00
	publications  []animationPublication533D00
	mutations     []animationMutation533D00
	before        [][]byte
	mimicGuard    []byte
	registryKey   string
	recording     bool
	runtime       server.MonsterAnimationNativeRuntime533790
	parent        server.MonsterNPCAnimationNativeRuntime533D00
	arrays        [2]*animationGuard533D00[[16]server.MonsterAnim]
	frames        [53][2]int32
	helpers       []NPCAnimationHelper533D00
	scratchBefore []byte
	projection    bool
}

var animationNames533D00 = [4]string{"Mimic", "CarnivorousPlant", "Zombie", "VileZombie"}

func animationSnapshot533D00(p unsafe.Pointer, n uintptr) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(p), int(n))...)
}
func animationAllocate533D00[T comparable](t *testing.T, f *animationFixture533D00) *animationGuard533D00[T] {
	p, free := alloc.New(animationGuard533D00[T]{})
	t.Cleanup(free)
	f.regions = append(f.regions, animationRegion533D00{unsafe.Pointer(p), unsafe.Sizeof(*p)})
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
		t.Fatal("all actual native animation allocations must exceed4GiB")
	}
	return p
}
func animationNewFixture533D00(t *testing.T) *animationFixture533D00 {
	f := &animationFixture533D00{t: t, srv: npcReflectLegacyServer4E17B0(t)}
	f.unit = animationAllocate533D00[server.Object](t, f)
	for i := range f.updates {
		f.updates[i] = animationAllocate533D00[server.MonsterUpdateData](t, f)
	}
	f.target = animationAllocate533D00[server.Object](t, f)
	for i := range f.arrays {
		f.arrays[i] = animationAllocate533D00[[16]server.MonsterAnim](t, f)
	}
	weaponPointer := memmap.PtrOff(0x587000, 215824)
	weaponBefore := animationSnapshot533D00(weaponPointer, 108)
	t.Cleanup(func() { copy(unsafe.Slice((*byte)(weaponPointer), 108), weaponBefore) })
	scratchPointer := memmap.PtrOff(0x5D4594, 2487964)
	scratchBefore := animationSnapshot533D00(unsafe.Add(scratchPointer, -32), 80)
	t.Cleanup(func() { copy(unsafe.Slice((*byte)(unsafe.Add(scratchPointer, -32)), 80), scratchBefore) })
	oldFrames := playerAnimFrames4F9F90
	playerAnimFrames4F9F90 = func(index int) (int, int) {
		if index < 0 || index >= len(f.frames) {
			t.Fatal("frame input outside original table")
		}
		v := f.frames[index]
		return int(v[0]), int(v[1])
	}
	t.Cleanup(func() { playerAnimFrames4F9F90 = oldFrames })
	f.cache = animationAllocate533D00[[3]uint32](t, f)
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
				f.publications = append(f.publications, animationPublication533D00{i + 1, v, f.call})
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
			f.predicates = append(f.predicates, animationPredicateEntry533D00{name, f.call, f.caches(), unit.TypeInd, uint32(unit.ObjClass), uint32(f.updates[0].Value.StatusFlags), live})
			return animationPredicatePublicC533790(index, unit)
		}
	}
	f.runtime = server.MonsterAnimationNativeRuntime533790{Mimic: wrap(0, "Mimic"), Plant: wrap(1, "CarnivorousPlant"), Zombie: wrap(2, "ZombiePair")}
	oldFactory := monsterAnimationRuntimeNative533790
	monsterAnimationRuntimeNative533790 = func() server.MonsterAnimationNativeRuntime533790 { return f.runtime }
	t.Cleanup(func() { monsterAnimationRuntimeNative533790 = oldFactory })
	f.parent = server.MonsterNPCAnimationNativeRuntime533D00{
		WeaponAnimation: func(value uint32) int32 {
			f.helpers = append(f.helpers, NPCAnimationHelper533D00{"0x4fa280", value, f.call})
			return NPCWeaponAnimationActualC533D00(value)
		},
		Frames: func(index int32) (int32, int32) {
			f.helpers = append(f.helpers, NPCAnimationHelper533D00{"0x4f9f90", uint32(index), f.call})
			return NPCFramesActualC533D00(index)
		},
		Selector: func(unit *server.Object) int32 {
			return int32(ccall.CallIntPtr(animationSelectorPublicC533790(), unit.CObj()))
		},
		Scratch: func() *server.MonsterAnim { return (*server.MonsterAnim)(memmap.PtrOff(0x5D4594, 2487964)) },
	}
	oldParent := monsterNPCAnimationRuntimeNative533D00
	monsterNPCAnimationRuntimeNative533D00 = func() server.MonsterNPCAnimationNativeRuntime533D00 { return f.parent }
	t.Cleanup(func() { monsterNPCAnimationRuntimeNative533D00 = oldParent })
	t.Cleanup(f.srv.FreeObjectTypes)
	return f
}
func (f *animationFixture533D00) caches() [4]uint32 {
	return [4]uint32{*(*uint32)(statePredicateOriginalMimicCachePointer534840()), f.cache.Value[0], f.cache.Value[1], f.cache.Value[2]}
}
func (f *animationFixture533D00) setCache(i int, v uint32) {
	if i == 0 {
		statePredicateOriginalSwapMimicCache534840(v)
	} else {
		f.cache.Value[i-1] = v
	}
}
func (f *animationFixture533D00) lookup(name string) {
	f.t.Helper()
	i := -1
	for n, v := range animationNames533D00 {
		if v == name {
			i = n
			break
		}
	}
	if i < 0 {
		f.t.Fatal("unknown dictionary name", name)
	}
	f.lookupCounts[name]++
	e := animationLookup533D00{Name: name, Call: f.call, Occurrence: f.lookupCounts[name], Before: f.caches(), Found: f.row.Input.Found[i]}
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
			case "entry_animation":
				if err := json.Unmarshal(effect.Value, &text); err != nil {
					f.t.Fatal(err)
				}
				f.updates[0].Value.Field119 = f.animationPointer(text)
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
			f.mutations = append(f.mutations, animationMutation533D00{effect.Field, effect.Value, f.call})
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
func (f *animationFixture533D00) animationPointer(role string) *[16]server.MonsterAnim {
	switch role {
	case "entry":
		return &f.arrays[0].Value
	case "other":
		return &f.arrays[1].Value
	case "nil":
		return nil
	default:
		f.t.Fatal("unknown animation role", role)
		return nil
	}
}
func (f *animationFixture533D00) prepare(row *animationRow533D00) {
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
					name = animationNames533D00[n]
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
	f.unit.Value = server.Object{TypeInd: c.Type, ObjClass: object.Class(c.Class), ObjSubClass: object.SubClass(c.Subclass), ObjFlags: object.Flags(c.Flags), UpdateData: unsafe.Pointer(&f.updates[0].Value)}
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
		f.updates[i].Value.WeaponEquipFlags = c.Equip
		f.updates[i].Value.Field517 = uint32(c.Attack)
		f.updates[i].Value.Field119 = f.animationPointer(c.Animation)
		item := (*server.AIStackItem)(unsafe.Add(unsafe.Pointer(&f.updates[i].Value.AIStack[0]), int(c.Head)*int(unsafe.Sizeof(f.updates[i].Value.AIStack[0]))))
		item.Action = c.Action
		if enemy {
			f.updates[i].Value.CurrentEnemy = &f.target.Value
		}
	}
	for i := range f.arrays {
		b := unsafe.Slice((*byte)(unsafe.Pointer(&f.arrays[i].Value)), 256)
		for n := range b {
			b[n] = byte(n + 7*i)
		}
	}
	for i := range f.frames {
		if c.TableProfile == 0 {
			f.frames[i] = [2]int32{int32(i*257 + 3), int32(i*259 + 5)}
		} else {
			f.frames[i] = [2]int32{[]int32{-1, -2147483648, 0, 256}[i%4], []int32{-2147483647, -1, 511, 1}[i%4]}
		}
	}
	for i := 0; i < 27; i++ {
		*memmap.PtrUint32(0x587000, uintptr(215824+4*i)) = uint32((i*7 + 3) % 53)
	}
	for i := range unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 2487964)), 16) {
		unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 2487964)), 16)[i] = 0x6d
	}
	f.scratchBefore = animationSnapshot533D00(unsafe.Add(memmap.PtrOff(0x5D4594, 2487964), -32), 80)
	f.helpers = []NPCAnimationHelper533D00{}
	for i, v := range c.Caches {
		f.setCache(i, v)
	}
	f.lookupCounts = map[string]int{}
	f.predicates = []animationPredicateEntry533D00{}
	f.lookups = []animationLookup533D00{}
	f.publications = []animationPublication533D00{}
	f.mutations = []animationMutation533D00{}
	f.before = make([][]byte, len(f.regions))
	for i, r := range f.regions {
		f.before[i] = animationSnapshot533D00(r.Ptr, r.Size)
	}
	f.mimicGuard = animationSnapshot533D00(unsafe.Add(statePredicateOriginalMimicCachePointer534840(), -64), 132)
}
func animationPatchValue533D00[T any](t *testing.T, expected [][]byte, regions []animationRegion533D00, pointer *T, value T) {
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
func (f *animationFixture533D00) check(control bool) {
	f.t.Helper()
	r := f.row
	o := r.Output
	if !o.Guards || len(o.Returns) != r.Input.Repeat {
		f.t.Fatal("invalid original guard/call contract")
	}
	if !control {
		if !f.projection && !reflect.DeepEqual(f.helpers, o.Helpers) {
			f.t.Errorf("actual weapon/frame helper entries=%v original=%v", f.helpers, o.Helpers)
		}
		scratch, err := hex.DecodeString(o.ScratchAfter)
		if err != nil || len(scratch) != 16 {
			f.t.Fatal("invalid original scratch record")
		}
		copy(f.scratchBefore[32:48], scratch)
		animationPatchValue533D00(f.t, f.before, f.regions, &f.updates[0].Value.Field119, f.animationPointer(o.AnimationAfter))
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
		pub := []animationPublication533D00{}
		for _, p := range o.Publications {
			if p.Index != 0 {
				pub = append(pub, p)
			}
		}
		if !reflect.DeepEqual(f.publications, pub) {
			f.t.Errorf("actual Plant/Zombie publications=%v original=%v", f.publications, pub)
		}
		animationPatchValue533D00(f.t, f.before, f.regions, &f.unit.Value.TypeInd, o.Type)
		animationPatchValue533D00(f.t, f.before, f.regions, &f.unit.Value.ObjClass, object.Class(o.Class))
		link := unsafe.Pointer(&f.updates[0].Value)
		if o.Live == "other" {
			link = unsafe.Pointer(&f.updates[1].Value)
		} else if o.Live != "entry" {
			f.t.Fatal("unknown original live-update role")
		}
		animationPatchValue533D00(f.t, f.before, f.regions, &f.unit.Value.UpdateData, link)
		for i := range f.updates {
			status, enemy := o.Status, o.Enemy
			if i == 1 {
				status, enemy = o.OtherStatus, o.OtherEnemy
			}
			animationPatchValue533D00(f.t, f.before, f.regions, &f.updates[i].Value.StatusFlags, object.MonsterStatus(status))
			var ptr *server.Object
			if enemy {
				ptr = &f.target.Value
			}
			animationPatchValue533D00(f.t, f.before, f.regions, &f.updates[i].Value.CurrentEnemy, ptr)
		}
		if o.Action != nil {
			if r.Input.Head < 0 {
				f.t.Fatal("unexpected original action output")
			}
			animationPatchValue533D00(f.t, f.before, f.regions, &f.updates[0].Value.AIStack[r.Input.Head].Action, *o.Action)
		}
		for i := range f.cache.Value {
			animationPatchValue533D00(f.t, f.before, f.regions, &f.cache.Value[i], o.Caches[i+1])
		}
		*(*uint32)(unsafe.Pointer(&f.mimicGuard[64])) = o.Caches[0]
	}
	if !bytes.Equal(f.scratchBefore, animationSnapshot533D00(unsafe.Add(memmap.PtrOff(0x5D4594, 2487964), -32), 80)) {
		f.t.Error("actual scratch bytes or neighbor guards differ from original")
	}
	for i, region := range f.regions {
		if !bytes.Equal(f.before[i], unsafe.Slice((*byte)(region.Ptr), int(region.Size))) {
			f.t.Errorf("native guarded unit/update/other/target/cache region%d differs beyond original outputs", i)
		}
	}
	if !bytes.Equal(f.mimicGuard, animationSnapshot533D00(unsafe.Add(statePredicateOriginalMimicCachePointer534840(), -64), 132)) {
		f.t.Error("native Mimic cache or neighbor guards differ")
	}
}
func animationRunOriginal533D00(t *testing.T, kind string, invoke func(*animationFixture533D00) unsafe.Pointer) {
	path := os.Getenv("NOX_NPC_ANIMATION_ORIGINAL_REFERENCE")
	if path == "" {
		path = "../server/testdata/monster_NPC_animation_original_533D00.json.gz"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != "849441cc80e93c8d8b9a904b7cbcd4fe96d046e6ace6e0bd5a195ccbc3eccdab" {
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
		Rows     []animationRow533D00 `json:"rows"`
	}
	if err = json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if ref.Schema != "original-monster-NPC-animation-533D00-v1" || ref.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || ref.Control != 0xe3f || ref.Entry != 0x533D00 || len(ref.Rows) != 19668 {
		t.Fatal("original animation provenance/dimensions")
	}
	f := animationNewFixture533D00(t)
	f.projection = kind == "existing private Go projection"
	control := os.Getenv("NOX_NPC_ANIMATION_CONTROL_ONLY") == "true"
	calls := 0
	for _, row := range ref.Rows {
		if f.projection && row.Input.Subclass&0x10 == 0 {
			continue
		}
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
					var base uintptr
					switch want.Role {
					case "scratch":
						base = uintptr(memmap.PtrOff(0x5D4594, 2487964))
					case "entry", "other":
						base = uintptr(unsafe.Pointer(f.animationPointer(want.Role)))
					case "nil":
					default:
						t.Fatal("unknown original return role", want.Role)
					}
					if uintptr(got) != base+want.Offset {
						t.Errorf("native parent pointer call%d=%p original role=%s offset=%d", call, got, want.Role, want.Offset)
					}
				}
			}
			f.recording = false
			f.check(control)
		})
	}
	if !control && !f.projection && calls != 20004 {
		t.Fatal("actual root call count", calls)
	}
	t.Logf("original %s NPC-animation contexts=19668 calls=%d control-only=%t actual-native-predicates-real-dictionary-cached-state-and-guards=true", kind, calls, control)
}
func TestNPCAnimationOriginalActualC533D00(t *testing.T) {
	animationRunOriginal533D00(t, "actual C", func(f *animationFixture533D00) unsafe.Pointer {
		return ccall.CallPtrPtr(NPCAnimationPublicC533D00(), f.unit.Value.CObj())
	})
}
func TestNPCAnimationOriginalGo533D00(t *testing.T) {
	animationRunOriginal533D00(t, "Go", func(f *animationFixture533D00) unsafe.Pointer {
		return unsafe.Pointer(server.MonsterNPCAnimationNative533D00(&f.unit.Value, f.parent))
	})
}
