package server

import (
	"math"
	"reflect"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
)

func TestRewardSpellBookNativeLayout4F09F0(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantInitData := uintptr(692)
	wantUseData := uintptr(736)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantInitData = 760
		wantUseData = 848
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantInitData},
		{"Object.UseData", unsafe.Offsetof(Object{}.UseData), wantUseData},
		{"RewardMarkerInitData size", unsafe.Sizeof(RewardMarkerInitData{}), 220},
		{"RewardMarkerInitData.RewardFlags", unsafe.Offsetof(RewardMarkerInitData{}.RewardFlags), 4},
		{"RewardMarkerInitData.Spells", unsafe.Offsetof(RewardMarkerInitData{}.Spells), 8},
		{"RewardMarkerInitData.ChanceMode", unsafe.Offsetof(RewardMarkerInitData{}.ChanceMode), 212},
		{"SpellRewardUseData size", unsafe.Sizeof(SpellRewardUseData{}), 1},
		{"SpellRewardUseData.Spell", unsafe.Offsetof(SpellRewardUseData{}.Spell), 0},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
	if len(RewardMarkerInitData{}.Spells) != rewardSpellBookSpellCount4F09F0 {
		t.Fatalf("explicit spell count = %d, want %d", len(RewardMarkerInitData{}.Spells), rewardSpellBookSpellCount4F09F0)
	}
}

func TestRewardSpellBookNativeUsesNativePointersAndCachedInitData4F09F0(t *testing.T) {
	entry := &RewardMarkerInitData{RewardFlags: 1}
	entry.Spells[136] = 1
	replacement := &RewardMarkerInitData{}
	marker := &Object{InitData: unsafe.Pointer(entry)}
	useData := &SpellRewardUseData{}
	created := &Object{}
	created.UseData.SetPtr(unsafe.Pointer(useData))
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(marker)) <= math.MaxUint32 || uintptr(unsafe.Pointer(created)) <= math.MaxUint32) {
		t.Fatalf("test pointers do not exercise native high addresses: marker=%p created=%p", marker, created)
	}

	var events []string
	got := rewardSpellBookNative4F09F0(marker, math.MaxUint32, rewardSpellBookNativeDeps4F09F0{
		pickSlots: func(uint32) uint32 {
			t.Fatal("explicit native path called slot selector")
			return 0
		},
		randomInt: func(minimum, maximum int32) int32 {
			events = append(events, "rng")
			if minimum != 0 || maximum != 0 {
				t.Fatalf("RNG bounds = %d..%d, want 0..0", minimum, maximum)
			}
			marker.InitData = unsafe.Pointer(replacement)
			return 0
		},
		checkSpellClass: func(class, spellID uint32) int32 {
			events = append(events, "class")
			if spellID != 136 || (class != 1 && class != 2) {
				t.Fatalf("class/spell = %d/%d", class, spellID)
			}
			return 0
		},
		createObjectByType: func(typeName string) *Object {
			events = append(events, "create:"+typeName)
			return created
		},
	})
	if got != created || useData.Spell != 136 {
		t.Fatalf("result/use spell = %p/%d, want %p/136", got, useData.Spell, created)
	}
	wantEvents := []string{"rng", "class", "class", "create:CommonSpellBook"}
	if !reflect.DeepEqual(events, wantEvents) || marker.InitData != unsafe.Pointer(replacement) || entry.Spells[136] != 1 {
		t.Fatalf("events/live marker/entry = %v/%p/%d, want %v/replacement/1", events, marker.InitData, entry.Spells[136], wantEvents)
	}
}

func TestRewardSpellBookNativeAutomaticForwardsStageAndFullSpellID4F09F0(t *testing.T) {
	marker := &Object{InitData: unsafe.Pointer(&RewardMarkerInitData{})}
	useData := &SpellRewardUseData{}
	created := &Object{}
	created.UseData.SetPtr(unsafe.Pointer(useData))
	var classes []uint32
	got := rewardSpellBookNative4F09F0(marker, 7, rewardSpellBookNativeDeps4F09F0{
		pickSlots: func(stage uint32) uint32 {
			if stage != 7 {
				t.Fatalf("slot stage = %d, want 7", stage)
			}
			return 1
		},
		randomInt: func(minimum, maximum int32) int32 {
			if minimum != 0 || maximum != 47 {
				t.Fatalf("slot-one spell RNG = %d..%d, want 0..47", minimum, maximum)
			}
			return 0
		},
		checkSpellClass: func(class, spellID uint32) int32 {
			if spellID != 0x47 {
				t.Fatalf("full spell ID = %#x, want 0x47", spellID)
			}
			classes = append(classes, class)
			if class == 2 {
				return 9
			}
			return 0
		},
		createObjectByType: func(typeName string) *Object {
			if typeName != rewardSpellBookWizardType4F09F0 {
				t.Fatalf("created type = %q, want WizardSpellBook", typeName)
			}
			return created
		},
	})
	if got != created || useData.Spell != 0x47 || !reflect.DeepEqual(classes, []uint32{1, 2, 1}) {
		t.Fatalf("result/spell/classes = %p/%#x/%v", got, useData.Spell, classes)
	}
}

func TestRewardSpellClassCheckMatches57AEA0ForKnownClasses4F09F0(t *testing.T) {
	tests := []struct {
		flags things.SpellFlags
		class uint32
		want  int32
	}{
		{0, 1, 9},
		{0, 2, 9},
		{things.SpellClassWizard, 1, 0},
		{things.SpellClassWizard, 2, 9},
		{things.SpellClassConjurer, 1, 9},
		{things.SpellClassConjurer, 2, 0},
		{things.SpellClassAny, 1, 0},
		{things.SpellClassAny, 2, 0},
		{things.SpellClassWizard | things.SpellClassConjurer, 1, 0},
		{things.SpellClassWizard | things.SpellClassConjurer, 2, 0},
		{things.SpellClassAny, 0, 9},
		{things.SpellClassAny, 3, 9},
		{things.SpellClassAny, math.MaxUint32, 9},
	}
	for _, test := range tests {
		if got := rewardSpellClassCheck4F09F0(test.flags, test.class); got != test.want {
			t.Errorf("flags/class %#x/%d = %d, want %d", test.flags, test.class, got, test.want)
		}
	}
}

func TestRewardSpellBookServerUsesLogicRNGSpellRegistryAndObjectFactory4F09F0(t *testing.T) {
	srv := new(Server)
	srv.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(srv.handle, srv)
	t.Cleanup(func() { servers.Delete(srv.handle) })
	srv.Rand.Logic = prand.New(2011)
	srv.Objs.init(srv.handle)
	if !srv.Objs.Init(2) {
		t.Fatal("object allocator initialization failed")
	}
	t.Cleanup(srv.Objs.FreeObjects)

	const spellID = spell.ID(7)
	srv.Spells.byID = map[spell.ID]*SpellDef{
		spellID: {ID: spellID, Def: things.Spell{Flags: things.SpellClassAny}},
	}
	templateUse := &SpellRewardUseData{}
	bookType := &ObjectType{
		s:           &srv.Types,
		ind:         1,
		id:          rewardSpellBookCommonType4F09F0,
		flags:       object.FlagNoCollide,
		UseData:     UseDataPtr{Ptr: unsafe.Pointer(templateUse)},
		UseDataSize: unsafe.Sizeof(*templateUse),
	}
	srv.Types.byID = map[string]*ObjectType{"commonspellbook": bookType}
	srv.Types.byInd = []*ObjectType{nil, bookType}
	data := &RewardMarkerInitData{RewardFlags: 1}
	data.Spells[spellID] = 1
	marker := &Object{InitData: unsafe.Pointer(data)}

	beforeRNG := srv.Rand.Logic.Index()
	got := srv.RewardSpellBook4F09F0(marker, math.MaxUint32)
	if got == nil {
		t.Fatal("server object factory returned nil")
	}
	if got.TypeInd != 1 || got.UseDataSpellReward().Spell != uint8(spellID) {
		t.Fatalf("server result type/spell = %d/%d, want 1/%d", got.TypeInd, got.UseDataSpellReward().Spell, spellID)
	}
	if index, want := srv.Rand.Logic.Index(), (beforeRNG+1)%4096; index != want {
		t.Fatalf("logic RNG index = %d, want %d after one explicit-selection draw", index, want)
	}
	got.UseData.Free()
}

func TestRewardSpellBookNativeFaultBoundaries4F09F0(t *testing.T) {
	t.Run("nil marker faults before dependencies", func(t *testing.T) {
		calls := 0
		defer func() {
			fault := recover()
			if fault == nil || calls != 0 {
				t.Fatalf("nil marker recover/calls = %v/%d, want panic/0", fault, calls)
			}
		}()
		rewardSpellBookNative4F09F0(nil, 0, rewardSpellBookNativeDeps4F09F0{
			pickSlots: func(uint32) uint32 { calls++; return 1 },
		})
	})

	t.Run("nil InitData faults before dependencies", func(t *testing.T) {
		calls := 0
		defer func() {
			fault := recover()
			if fault == nil || calls != 0 {
				t.Fatalf("nil InitData recover/calls = %v/%d, want panic/0", fault, calls)
			}
		}()
		rewardSpellBookNative4F09F0(&Object{}, 0, rewardSpellBookNativeDeps4F09F0{
			pickSlots: func(uint32) uint32 { calls++; return 1 },
		})
	})

	t.Run("nil UseData faults after create", func(t *testing.T) {
		data := &RewardMarkerInitData{RewardFlags: 1}
		data.Spells[7] = 1
		created := &Object{}
		createdCalls := 0
		defer func() {
			fault := recover()
			if fault == nil || createdCalls != 1 {
				t.Fatalf("nil UseData recover/create calls = %v/%d, want panic/1", fault, createdCalls)
			}
		}()
		rewardSpellBookNative4F09F0(&Object{InitData: unsafe.Pointer(data)}, 0, rewardSpellBookNativeDeps4F09F0{
			randomInt:       func(int32, int32) int32 { return 0 },
			checkSpellClass: func(uint32, uint32) int32 { return 0 },
			createObjectByType: func(string) *Object {
				createdCalls++
				return created
			},
		})
	})
}
