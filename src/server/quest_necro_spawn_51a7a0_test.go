package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

type questNecroTest51A7A0 struct {
	trace                []string
	fault                int
	onEvent              func(string)
	unit, marker, reward *Object
	update               *MonsterUpdateData
	scale                float64
	position             types.Pointf
	stage                uint32
}

func newQuestNecroTest51A7A0() *questNecroTest51A7A0 {
	u := &MonsterUpdateData{MonsterDef: &MonsterDef{HealthQuest72: 25}}
	return &questNecroTest51A7A0{
		fault: -1, update: u, scale: 1.5, stage: 5,
		position: types.Pointf{X: 17, Y: 23},
		unit:     &Object{TypeInd: 9, UpdateData: unsafe.Pointer(u), HealthData: &HealthData{Cur: 7, Max: 7}},
		marker:   new(Object), reward: new(Object),
	}
}

func (f *questNecroTest51A7A0) record(event string) {
	f.trace = append(f.trace, event)
	if len(f.trace)-1 == f.fault {
		panic("test fault")
	}
	if f.onEvent != nil {
		f.onEvent(event)
	}
}

func (f *questNecroTest51A7A0) hooks() questNecroSpawnHooks51A7A0[*Object, *MonsterDef, *MonsterUpdateData, *HealthData, *ObjectType, types.Pointf] {
	return questNecroSpawnHooks51A7A0[*Object, *MonsterDef, *MonsterUpdateData, *HealthData, *ObjectType, types.Pointf]{
		newObject: func(name string) *Object {
			f.record("new:" + name)
			if name == "Necromancer" {
				return f.unit
			}
			if name == "RewardMarker" {
				return f.marker
			}
			panic(name)
		},
		healthScale:     func() float64 { f.record("scale"); return f.scale },
		loadUpdate:      func(o *Object) *MonsterUpdateData { f.record("update"); return (*MonsterUpdateData)(o.UpdateData) },
		loadDefinition:  func(u *MonsterUpdateData) *MonsterDef { f.record("definition"); return u.MonsterDef },
		loadQuestHealth: func(d *MonsterDef) int32 { f.record("quest-health"); return int32(d.HealthQuest72) },
		loadType:        func(o *Object) uint16 { f.record("type"); return o.TypeInd },
		lookupType: func(id uint16) *ObjectType {
			f.record(fmt.Sprint("lookup:", id))
			return &ObjectType{health: &HealthData{Max: 20}}
		},
		loadTypeHealth: func(o *ObjectType) *HealthData { f.record("type-health"); return o.Health() },
		loadHealth:     func(o *Object) *HealthData { f.record("health"); return o.HealthData },
		loadCurrent:    func(h *HealthData) uint16 { f.record("current"); return h.Cur },
		loadMaximum:    func(h *HealthData) uint16 { f.record("maximum"); return h.Max },
		setHP:          func(o *Object, v uint16) { f.record(fmt.Sprint("hp:", v)); o.HealthData.Cur = v },
		storeMaximum:   func(h *HealthData, v uint16) { f.record(fmt.Sprint("max:", v)); h.Max = v },
		storeAI: func(u *MonsterUpdateData, field, value uint32) {
			f.record(fmt.Sprintf("ai:%d:%08x", field, value))
			if u != f.update {
				panic("lost cached update")
			}
		},
		loadPosition: func() types.Pointf { f.record("position"); return f.position },
		createAt: func(o *Object, p types.Pointf) {
			f.record(fmt.Sprintf("create:%g:%g", p.X, p.Y))
			if o != f.unit {
				panic("wrong unit")
			}
		},
		stage: func() uint32 { f.record("stage"); return f.stage },
		activateReward: func(marker *Object, stage uint32) *Object {
			f.record(fmt.Sprint("reward:", stage))
			if marker != f.marker {
				panic("wrong marker")
			}
			return f.reward
		},
		inventoryPut: func(owner, item *Object, report int32) {
			f.record(fmt.Sprint("inventory:", report))
			if owner != f.unit || item != f.reward {
				panic("wrong inventory identity")
			}
		},
		freeObject: func(marker *Object) {
			f.record("free")
			if marker != f.marker {
				panic("wrong free")
			}
		},
	}
}

func TestQuestNecroSpawn51A7A0OrderAndEveryFaultPrefix(t *testing.T) {
	want := []string{
		"new:Necromancer", "scale", "update", "definition", "quest-health", "hp:37",
		"health", "max:38", "health", "current", "health", "maximum",
		"ai:340:00000004", "ai:411:10000000", "ai:423:10000000", "ai:326:3f547ae1",
		"ai:510:00000001", "ai:410:08000000", "ai:444:20000000", "ai:415:40000000",
		"position", "create:17:23", "new:RewardMarker", "stage", "reward:7", "inventory:0", "free",
	}
	f := newQuestNecroTest51A7A0()
	questNecroSpawn51A7A0(f.hooks())
	if !reflect.DeepEqual(f.trace, want) || f.unit.HealthData.Cur != 37 || f.unit.HealthData.Max != 38 {
		t.Fatalf("trace=%v health=%+v", f.trace, f.unit.HealthData)
	}
	for i := range want {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			f := newQuestNecroTest51A7A0()
			f.fault = i
			func() {
				defer func() {
					if recover() != "test fault" {
						t.Fatal("missing injected fault")
					}
				}()
				questNecroSpawn51A7A0(f.hooks())
			}()
			if !reflect.DeepEqual(f.trace, want[:i+1]) {
				t.Fatalf("fault prefix=%v", f.trace)
			}
		})
	}
}

func TestQuestNecroSpawn51A7A0BranchesAndLiveReloads(t *testing.T) {
	t.Run("failed allocation still reads scale", func(t *testing.T) {
		f := newQuestNecroTest51A7A0()
		f.unit = nil
		questNecroSpawn51A7A0(f.hooks())
		if !reflect.DeepEqual(f.trace, []string{"new:Necromancer", "scale"}) {
			t.Fatal(f.trace)
		}
	})
	t.Run("type fallback", func(t *testing.T) {
		f := newQuestNecroTest51A7A0()
		f.update.MonsterDef = nil
		questNecroSpawn51A7A0(f.hooks())
		if !reflect.DeepEqual(f.trace[4:9], []string{"type", "lookup:9", "type-health", "maximum", "hp:30"}) || f.unit.HealthData.Max != 30 {
			t.Fatal(f.trace)
		}
	})
	for _, scale := range []float64{-2, 0, math.NaN(), math.Inf(-1)} {
		f := newQuestNecroTest51A7A0()
		f.scale = scale
		questNecroSpawn51A7A0(f.hooks())
		if f.unit.HealthData.Cur != 25 || f.unit.HealthData.Max != 25 {
			t.Fatalf("scale=%v health=%+v", scale, f.unit.HealthData)
		}
	}
	t.Run("zero repair and live pointers", func(t *testing.T) {
		f := newQuestNecroTest51A7A0()
		f.update.MonsterDef.HealthQuest72 = 0
		original := f.unit.HealthData
		first, second := new(HealthData), new(HealthData)
		f.onEvent = func(e string) {
			if e == "hp:0" {
				f.unit.HealthData = first
				f.unit.UpdateData = unsafe.Pointer(new(MonsterUpdateData))
			}
			if e == "hp:1" {
				f.unit.HealthData = second
				f.position = types.Pointf{X: 41, Y: 43}
			}
		}
		questNecroSpawn51A7A0(f.hooks())
		if original.Cur != 7 || original.Max != 7 || first.Max != 0 || second.Cur != 1 || second.Max != 1 || f.trace[len(f.trace)-6] != "create:41:43" {
			t.Fatalf("health=%+v/%+v/%+v trace=%v", original, first, second, f.trace)
		}
	})
	t.Run("stage wraps and nil reward still frees", func(t *testing.T) {
		f := newQuestNecroTest51A7A0()
		f.stage, f.reward = math.MaxUint32, nil
		questNecroSpawn51A7A0(f.hooks())
		if !reflect.DeepEqual(f.trace[len(f.trace)-3:], []string{"stage", "reward:1", "free"}) {
			t.Fatal(f.trace)
		}
	})
	t.Run("nil marker ends before stage", func(t *testing.T) {
		f := newQuestNecroTest51A7A0()
		f.marker = nil
		questNecroSpawn51A7A0(f.hooks())
		if f.trace[len(f.trace)-1] != "new:RewardMarker" {
			t.Fatal(f.trace)
		}
	})
}

func TestQuestMinionHealth51A7A0X87ProductAndSeparateSpill(t *testing.T) {
	for _, tc := range []struct {
		name             string
		base             int32
		scale            float32
		current, maximum uint16
	}{
		{"nearest even up", 25, 1.5, 37, 38},
		{"nearest even down", 23, 1.5, 34, 34},
		{"signed base", -25, 1.5, 65536 - 37, 65536 - 38},
		{"zero", 0, 2, 0, 0},
		{"word wrap", 65536, 1, 0, 0},
		{"qword product beyond binary64 precision", math.MaxInt32, 16777218, 65534, 0},
		{"invalid qword", math.MaxInt32, math.MaxFloat32, 0, 0},
		{"invalid signed dword spill", math.MaxInt32, 1, 65535, 0},
		{"positive infinity", 25, float32(math.Inf(1)), 0, 0},
		{"nan", 25, float32(math.NaN()), 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, m := questMinionHealth51A7A0(tc.base, tc.scale)
			if c != tc.current || m != tc.maximum {
				t.Fatalf("got %d/%d, want %d/%d", c, m, tc.current, tc.maximum)
			}
		})
	}
}
