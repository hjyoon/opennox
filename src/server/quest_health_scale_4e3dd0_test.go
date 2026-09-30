package server

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

type questHealthTestObject4E3DD0 struct {
	name         string
	class, flags uint32
	health       *HealthData
	typeID       uint16
	update       *questHealthTestUpdate4E3DD0
	next         *questHealthTestObject4E3DD0
}

type questHealthTestType4E3DD0 struct{ health *HealthData }

type questHealthTestUpdate4E3DD0 struct {
	definition *MonsterDef
	status     uint32
	history    [32]uint16
}

type questHealthTest4E3DD0 struct {
	trace                []string
	fault                int
	onEvent              func(string)
	ready                uint32
	cache                [4]float32
	balance              map[string]float32
	difficulty, factor   float64
	damageScale, hpScale float32
	first                *questHealthTestObject4E3DD0
	types                map[uint16]*questHealthTestType4E3DD0
}

func newQuestHealthTest4E3DD0(class uint32) *questHealthTest4E3DD0 {
	f := &questHealthTest4E3DD0{
		fault: -1, ready: 1, difficulty: 5, factor: 2.5,
		cache: [4]float32{2, 3, 0.25, 0.5},
		balance: map[string]float32{
			"GeneratorMaxHealth": 100, "PlayerDamageDiffInit": 2,
			"SystemHealthDiffInit": 3, "PlayerDamageDiffCoeff": 0.25,
			"SystemHealthDiffCoeff": 0.5,
		},
		types: map[uint16]*questHealthTestType4E3DD0{1: {health: &HealthData{Cur: 3, Max: 3}}},
	}
	update := &questHealthTestUpdate4E3DD0{}
	for i := range update.history {
		update.history[i] = 999
	}
	f.first = &questHealthTestObject4E3DD0{
		name: "one", class: class, health: &HealthData{Cur: 3, Max: 3}, typeID: 1, update: update,
	}
	return f
}

func (f *questHealthTest4E3DD0) record(event string) {
	f.trace = append(f.trace, event)
	if len(f.trace)-1 == f.fault {
		panic("test fault")
	}
	if f.onEvent != nil {
		f.onEvent(event)
	}
}

func (f *questHealthTest4E3DD0) hooks() questHealthScaleHooks4E3DD0[*questHealthTestObject4E3DD0, *MonsterDef, *HealthData, *questHealthTestType4E3DD0, *questHealthTestUpdate4E3DD0] {
	return questHealthScaleHooks4E3DD0[*questHealthTestObject4E3DD0, *MonsterDef, *HealthData, *questHealthTestType4E3DD0, *questHealthTestUpdate4E3DD0]{
		balanceFloat: func(key string) float32 { f.record("balance:" + key); return f.balance[key] },
		loadReady:    func() uint32 { f.record("ready"); return f.ready },
		storeReady:   func(value uint32) { f.record("ready:set"); f.ready = value },
		loadCache: func(index questHealthCache4E3DD0) float32 {
			f.record(fmt.Sprint("cache:", index))
			return f.cache[index]
		},
		storeCache: func(index questHealthCache4E3DD0, value float32) {
			f.record(fmt.Sprint("cache:set:", index))
			f.cache[index] = value
		},
		difficulty:       func() float64 { f.record("difficulty"); return f.difficulty },
		storeDamageScale: func(value float32) { f.record("damage:set"); f.damageScale = value },
		storeHealthScale: func(value float32) { f.record("scale:set"); f.hpScale = value },
		healthScale:      func() float64 { f.record("factor"); return f.factor },
		first:            func() *questHealthTestObject4E3DD0 { f.record("first"); return f.first },
		next: func(obj *questHealthTestObject4E3DD0) *questHealthTestObject4E3DD0 {
			f.record("next:" + obj.name)
			return obj.next
		},
		loadClass:         func(obj *questHealthTestObject4E3DD0) uint32 { f.record("class:" + obj.name); return obj.class },
		loadFlags:         func(obj *questHealthTestObject4E3DD0) uint32 { f.record("flags:" + obj.name); return obj.flags },
		loadHealth:        func(obj *questHealthTestObject4E3DD0) *HealthData { f.record("health:" + obj.name); return obj.health },
		healthPointerWord: func(*HealthData) uint16 { return 0xa1b2 },
		loadCurrent:       func(health *HealthData) uint16 { f.record("cur"); return health.Cur },
		loadMaximum:       func(health *HealthData) uint16 { f.record("max"); return health.Max },
		loadTypeID:        func(obj *questHealthTestObject4E3DD0) uint16 { f.record("type:" + obj.name); return obj.typeID },
		lookupType:        func(id uint16) *questHealthTestType4E3DD0 { f.record("lookup"); return f.types[id] },
		loadTypeHealth:    func(typ *questHealthTestType4E3DD0) *HealthData { f.record("type:health"); return typ.health },
		loadUpdate: func(obj *questHealthTestObject4E3DD0) *questHealthTestUpdate4E3DD0 {
			f.record("update:" + obj.name)
			return obj.update
		},
		loadDefinition: func(update *questHealthTestUpdate4E3DD0) *MonsterDef {
			f.record("definition")
			return update.definition
		},
		loadQuestHealth: func(definition *MonsterDef) uint16 {
			f.record("definition:hp")
			return uint16(definition.HealthQuest72)
		},
		loadStatusByte: func(update *questHealthTestUpdate4E3DD0) uint8 { f.record("status"); return uint8(update.status) },
		setHP: func(obj *questHealthTestObject4E3DD0, value uint16) int32 {
			f.record("set:" + obj.name)
			obj.health.Cur = value
			return 0x12345678
		},
		storeMaximum: func(health *HealthData, value uint16) { f.record("max:set"); health.Max = value },
		storeHistory: func(update *questHealthTestUpdate4E3DD0, index int, value uint16) {
			f.record(fmt.Sprintf("history:%02d", index))
			update.history[index] = value
		},
		historyEndWord: func(*questHealthTestUpdate4E3DD0) uint16 { return 0xc3d4 },
	}
}

func TestQuestHealthScale4E3DD0CacheAndLiveCurveOrder(t *testing.T) {
	f := newQuestHealthTest4E3DD0(0)
	f.ready, f.first = 0, nil
	f.onEvent = func(event string) {
		if event == "damage:set" {
			f.difficulty = 3
			f.cache[questHealthHealthCoeff4E3DD0] = 2
			f.cache[questHealthHealthInit4E3DD0] = 1
		}
	}
	if got := questHealthScale4E3DD0(f.hooks()); got != 0 {
		t.Fatalf("empty return = %d", got)
	}
	want := []string{
		"balance:GeneratorMaxHealth", "ready", "balance:PlayerDamageDiffInit", "cache:set:0",
		"balance:SystemHealthDiffInit", "cache:set:1", "balance:PlayerDamageDiffCoeff", "cache:set:2",
		"balance:SystemHealthDiffCoeff", "cache:set:3", "ready:set", "difficulty", "cache:2", "cache:0",
		"damage:set", "difficulty", "cache:3", "cache:1", "scale:set", "first",
	}
	if !reflect.DeepEqual(f.trace, want) || f.ready != 1 || f.damageScale != 3 || f.hpScale != 5 {
		t.Fatalf("trace=%v ready=%d scales=%v/%v", f.trace, f.ready, f.damageScale, f.hpScale)
	}
	for _, ready := range []uint32{1, 2, math.MaxUint32} {
		f := newQuestHealthTest4E3DD0(0)
		f.ready, f.first = ready, nil
		questHealthScale4E3DD0(f.hooks())
		for _, event := range f.trace[1:] {
			if strings.HasPrefix(event, "balance:") || strings.Contains(event, "cache:set") || event == "ready:set" {
				t.Fatalf("ready=%d unexpectedly reinitialized: %v", ready, f.trace)
			}
		}
	}
}

func TestQuestHealthScale4E3DD0FilterPriorityAndShortReturns(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		class, flags, status uint32
		cur, max             uint16
		want                 uint16
		set, health          bool
	}{
		{"other", 0x1234, 0, 0, 3, 3, 0x1234, false, false},
		{"destroyed_generator", 0x20000, 0x8000, 0, 3, 3, 0, false, false},
		{"destroyed_monster", 2, 0x8007, 0, 3, 3, 0x8007, false, false},
		{"destroyed_both", 0x20002, 0x8009, 0, 3, 3, 0x8009, false, false},
		{"zero_max", 2, 0, 0, 3, 0, 0xa1b2, false, true},
		{"zero_cur", 2, 0, 0, 0, 3, 0, false, true},
		{"injured_monster", 2, 0, 0, 2, 3, 2, false, true},
		{"injured_generator", 0x20000, 0, 0, 2, 3, 2, false, true},
		{"status_excluded", 2, 0, 0x80, 3, 3, 3, false, true},
		{"other_status_bits", 2, 0, 0xffffff7f, 3, 3, 0xc3d4, true, true},
		{"generator_priority", 0x20002, 0, 0x80, 3, 3, 0x5678, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newQuestHealthTest4E3DD0(tc.class)
			f.first.flags, f.first.update.status = tc.flags, tc.status
			f.first.health.Cur, f.first.health.Max = tc.cur, tc.max
			if !tc.health {
				f.first.health = nil
			}
			got := questHealthScale4E3DD0(f.hooks())
			set := false
			for _, event := range f.trace {
				set = set || event == "set:one"
			}
			if uint16(got) != tc.want || set != tc.set {
				t.Fatalf("return=%04x set=%t trace=%v", uint16(got), set, f.trace)
			}
			if !tc.set && tc.health && (f.first.health.Cur != tc.cur || f.first.health.Max != tc.max) {
				t.Fatal("excluded object changed health")
			}
			if tc.class == 0x20002 && tc.flags == 0 {
				for _, event := range f.trace {
					if event == "update:one" || event == "status" {
						t.Fatal("generator fell through to monster")
					}
				}
			}
		})
	}
}

func TestQuestHealthScale4E3DD0WordConversionAndGeneratorCap(t *testing.T) {
	for _, tc := range []struct {
		name             string
		factor           float64
		cur, max         uint16
		cap              float32
		wantCur, wantMax uint16
	}{
		{"ties_even", 2.5, 1, 3, 100, 2, 8},
		{"negative_fabs", -2.5, 3, 1, 100, 8, 2},
		{"independent_cap", 2.5, 3, 1, 4, 4, 2},
		{"cap_round_even", 2.5, 3, 3, 3.5, 4, 4},
		{"cap_zero_after_minimum", 0, 3, 3, 0, 0, 0},
		{"zero_product", 0, 3, 3, 100, 1, 1},
		{"cap_low_word", 2.5, 3, 3, 65536, 0, 0},
		{"negative_cap_word", 2.5, 3, 3, -1, 8, 8},
		{"unsigned_word", 1, 65535, 65535, -1, 65535, 65535},
		{"wrapped_product", 2, 65535, 65535, -1, 65534, 65534},
		{"invalid_product", math.Inf(1), 3, 3, 100, 1, 1},
		{"nan_product", math.NaN(), 3, 3, 100, 1, 1},
		{"invalid_cap", 2.5, 3, 3, float32(math.Inf(1)), 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newQuestHealthTest4E3DD0(0x20000)
			f.factor, f.balance["GeneratorMaxHealth"] = tc.factor, tc.cap
			f.types[1].health.Cur, f.types[1].health.Max = tc.cur, tc.max
			questHealthScale4E3DD0(f.hooks())
			if got := f.first.health; got.Cur != tc.wantCur || got.Max != tc.wantMax {
				t.Fatalf("health=%d/%d want=%d/%d", got.Cur, got.Max, tc.wantCur, tc.wantMax)
			}
		})
	}
	for _, tc := range []struct {
		value float32
		want  uint16
	}{
		{0.5, 0}, {1.5, 2}, {2.5, 2}, {-3.5, 4}, {65536, 0},
		{2147483648, 0}, {float32(math.NaN()), 0}, {float32(math.Inf(-1)), 0},
	} {
		if got := questHealthWord4E3DD0(tc.value); got != tc.want {
			t.Fatalf("word(%v)=%d want=%d", tc.value, got, tc.want)
		}
	}
}

func TestQuestHealthScale4E3DD0GeneratorCachesTypeAndNextButReloadsHealth(t *testing.T) {
	f := newQuestHealthTest4E3DD0(0x20000)
	obj, typ := f.first, f.types[1]
	other := &questHealthTestObject4E3DD0{name: "two", class: 0x7654}
	obj.next = other
	replacement := &HealthData{Cur: 77, Max: 99}
	factors := 0
	f.onEvent = func(event string) {
		switch event {
		case "factor":
			factors++
			if factors == 1 {
				typ.health = &HealthData{Cur: 1, Max: 3}
			} else {
				typ.health = &HealthData{Cur: 5, Max: 9}
				f.factor = 1.5
			}
			obj.typeID = 2 // lookup result stays cached despite live object changes
		case "set:one":
			obj.next, obj.health = nil, replacement
		}
	}
	if got := questHealthScale4E3DD0(f.hooks()); uint16(got) != 0x7654 {
		t.Fatalf("cached successor return=%04x", uint16(got))
	}
	if replacement.Cur != 8 || replacement.Max != 8 {
		t.Fatalf("reloaded health=%+v", replacement)
	}
	wantTail := []string{"set:one", "health:one", "max:set", "next:two", "class:two"}
	if !reflect.DeepEqual(f.trace[len(f.trace)-len(wantTail):], wantTail) {
		t.Fatalf("tail=%v", f.trace)
	}
	if factors != 2 {
		t.Fatalf("scale reads=%d", factors)
	}
}

func TestQuestHealthScale4E3DD0MonsterCachedDefinitionUpdateAndLiveHistory(t *testing.T) {
	for _, useDefinition := range []bool{false, true} {
		t.Run(fmt.Sprint(useDefinition), func(t *testing.T) {
			f := newQuestHealthTest4E3DD0(2)
			f.balance["GeneratorMaxHealth"] = 0 // monster health is never capped
			obj, update := f.first, f.first.update
			if useDefinition {
				update.definition = &MonsterDef{HealthQuest72: 0x12340003}
			}
			replacementUpdate := &questHealthTestUpdate4E3DD0{}
			replacementHP, thirdHP := &HealthData{Cur: 77, Max: 99}, &HealthData{Cur: 23, Max: 25}
			factors := 0
			f.onEvent = func(event string) {
				switch event {
				case "factor":
					factors++
					f.types[1].health.Max = 100
					if update.definition != nil {
						update.definition.HealthQuest72 = 200
					}
					if factors == 2 {
						f.factor = 1.5
					}
				case "set:one":
					obj.health, obj.update = replacementHP, replacementUpdate
				case "history:00":
					obj.health = thirdHP // next sample must reload, first value was already read
				}
			}
			if got := questHealthScale4E3DD0(f.hooks()); uint16(got) != 0xc3d4 {
				t.Fatalf("history return=%04x", uint16(got))
			}
			if replacementHP.Cur != 4 || replacementHP.Max != 8 || thirdHP.Max != 25 {
				t.Fatalf("hp=%+v third=%+v", replacementHP, thirdHP)
			}
			for i, got := range update.history {
				want := uint16(23)
				if i == 0 {
					want = 4
				}
				if got != want {
					t.Fatalf("history[%d]=%d want=%d", i, got, want)
				}
				if replacementUpdate.history[i] != 0 {
					t.Fatal("reloaded update instead of cached update")
				}
			}
			if factors != 2 {
				t.Fatalf("scale reads=%d", factors)
			}
			joined := strings.Join(f.trace, ",")
			if !strings.Contains(joined, "lookup,update:one,definition,") || !strings.Contains(joined, "set:one,health:one,max:set,health:one,cur,history:00,health:one,cur,history:01") {
				t.Fatalf("callback/reload order=%s", joined)
			}
		})
	}
}

func TestQuestHealthScale4E3DD0ObservableFaultPrefixes(t *testing.T) {
	for _, class := range []uint32{0x20000, 2} {
		base := newQuestHealthTest4E3DD0(class)
		base.ready, base.cache = 0, [4]float32{}
		questHealthScale4E3DD0(base.hooks())
		for fault := range base.trace {
			t.Run(fmt.Sprintf("%x/%02d", class, fault), func(t *testing.T) {
				f := newQuestHealthTest4E3DD0(class)
				f.ready, f.cache, f.fault = 0, [4]float32{}, fault
				var recovered any
				func() { defer func() { recovered = recover() }(); questHealthScale4E3DD0(f.hooks()) }()
				if recovered != "test fault" || !reflect.DeepEqual(f.trace, base.trace[:fault+1]) {
					t.Fatalf("fault=%v trace=%v want=%v", recovered, f.trace, base.trace[:fault+1])
				}
				completed := make(map[string]bool)
				for _, event := range f.trace[:fault] {
					completed[event] = true
				}
				wantCur, wantMax := uint16(3), uint16(3)
				if completed["set:one"] {
					wantCur = 8
				}
				if completed["max:set"] {
					wantMax = 8
				}
				if f.first.health.Cur != wantCur || f.first.health.Max != wantMax {
					t.Fatalf("partial hp=%+v", f.first.health)
				}
				if (f.ready == 1) != completed["ready:set"] {
					t.Fatalf("ready=%d before initialization completed", f.ready)
				}
				for i, value := range f.cache {
					want := float32(0)
					if completed[fmt.Sprint("cache:set:", i)] {
						want = base.cache[i]
					}
					if value != want {
						t.Fatalf("partial cache[%d]=%v want=%v", i, value, want)
					}
				}
				for i, got := range f.first.update.history {
					want := uint16(999)
					if completed[fmt.Sprintf("history:%02d", i)] {
						want = 8
					}
					if got != want {
						t.Fatalf("partial history[%d]=%d want=%d", i, got, want)
					}
				}
			})
		}
	}
}
