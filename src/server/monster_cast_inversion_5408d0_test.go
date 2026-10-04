package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

type monsterInversionTestWorld5408D0 struct {
	events        []string
	unit          *Object
	update        *MonsterUpdateData
	flag          uint32
	frame         uint32
	missiles      []*Object
	candidates    []spell.ID
	selected      int
	cast          func()
	afterCooldown func()
	randomCalls   int
}

func newMonsterInversionTestWorld5408D0(t *testing.T) *monsterInversionTestWorld5408D0 {
	t.Helper()
	unit, update := monsterFightSpellTestUnit540B90(t)
	unit.PosVec = types.Ptf(12.5, -3.75)
	update.Field363 = 90
	update.Field362_0, update.Field362_2 = 11, 13
	monsterFightSetSpellFlag540B90(update, 1, monsterInversionSpellMask5408D0)
	monsterFightSetSpellFlag540B90(update, 3, monsterInversionSpellMask5408D0)
	monsterFightSetSpellFlag540B90(update, 136, monsterInversionSpellMask5408D0)
	return &monsterInversionTestWorld5408D0{
		unit: unit, update: update, flag: 77, frame: 100, selected: 1,
		missiles: []*Object{{ObjClass: object.ClassMissile, ObjSubClass: object.SubClass(object.MissileMagic),
			UpdateData: unsafe.Pointer(&MissileUpdateData{Target: unit})}},
	}
}

func (w *monsterInversionTestWorld5408D0) hooks(t *testing.T) monsterCastInversionHooks5408D0 {
	t.Helper()
	return monsterCastInversionHooks5408D0{
		frame: func() uint32 { w.events = append(w.events, "frame"); return w.frame },
		balance: func(key string) float64 {
			w.events = append(w.events, "balance:"+key)
			return 100.000001
		},
		eachMissile: func(pos types.Pointf, radius float32, visit func(*Object) bool) {
			w.events = append(w.events, "query")
			if pos != w.unit.PosVec || math.Float32bits(radius) != math.Float32bits(50) {
				t.Fatal("wrong query position or binary32 range spill")
			}
			for _, missile := range w.missiles {
				if !visit(missile) {
					t.Fatal("targeted match stopped original full enumeration")
				}
			}
		},
		loadThreat:  func() uint32 { w.events = append(w.events, "load-threat"); return w.flag },
		storeThreat: func(flag uint32) { w.events = append(w.events, fmt.Sprint("store-threat:", flag)); w.flag = flag },
		spellAllowed: func(id spell.ID) bool {
			w.events = append(w.events, fmt.Sprintf("allowed:%d", int(id)))
			if id == 3 {
				return false
			}
			w.candidates = append(w.candidates, id)
			return true
		},
		random: func(min, max int) int {
			w.events = append(w.events, fmt.Sprintf("rng:%d:%d", min, max))
			w.randomCalls++
			if w.randomCalls == 1 {
				return w.selected
			}
			if w.afterCooldown != nil {
				w.afterCooldown()
			}
			return max
		},
		cast: func(unit *Object, id spell.ID, target *Object) {
			w.events = append(w.events, fmt.Sprintf("cast:%d", int(id)))
			if unit != w.unit || target != w.unit {
				t.Fatal("inversion is not a native self-cast")
			}
			if w.cast != nil {
				w.cast()
			}
		},
	}
}

func TestMonsterCastInversion5408D0ExactCallOrderAndFullSpellSpan(t *testing.T) {
	w := newMonsterInversionTestWorld5408D0(t)
	if !monsterCastInversion5408D0(w.unit, w.hooks(t)) {
		t.Fatal("eligible spell was not selected")
	}
	want := []string{"frame", "store-threat:0", "balance:InversionRange", "query", "store-threat:1", "load-threat",
		"allowed:1", "allowed:3", "allowed:136", "rng:0:1", "cast:136", "rng:11:13", "frame"}
	if !reflect.DeepEqual(w.events, want) || !reflect.DeepEqual(w.candidates, []spell.ID{1, 136}) || w.update.Field363 != 113 {
		t.Fatalf("events/candidates/deadline = %q/%v/%d", w.events, w.candidates, w.update.Field363)
	}
}

func TestMonsterCastInversion5408D0CachedUpdateLiveBoundsAndPostRNGFrame(t *testing.T) {
	w := newMonsterInversionTestWorld5408D0(t)
	replacement := new(MonsterUpdateData)
	w.cast = func() {
		w.unit.UpdateData = unsafe.Pointer(replacement)
		w.update.Field362_0, w.update.Field362_2 = 0xfffe, 0xffff
		w.frame = 1000
	}
	w.afterCooldown = func() { w.frame = math.MaxUint32 - 7 }
	if !monsterCastInversion5408D0(w.unit, w.hooks(t)) {
		t.Fatal("cast was not selected")
	}
	if w.update.Field363 != 65527 || replacement.Field363 != 0 ||
		!reflect.DeepEqual(w.events[len(w.events)-2:], []string{"rng:65534:65535", "frame"}) {
		t.Fatalf("cached deadline/replacement/events = %d/%d/%q", w.update.Field363, replacement.Field363, w.events)
	}
	w.unit.UpdateData = unsafe.Pointer(w.update)
}

func TestMonsterCastInversion5408D0ExactEarlyGates(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*monsterInversionTestWorld5408D0)
		want  []string
	}{
		{name: "disabled", setup: func(w *monsterInversionTestWorld5408D0) { w.unit.ObjFlags = 0 }},
		{name: "cannot cast", setup: func(w *monsterInversionTestWorld5408D0) { w.update.StatusFlags = 0 }},
		{name: "future deadline", setup: func(w *monsterInversionTestWorld5408D0) { w.update.Field363 = 101 }, want: []string{"frame"}},
		{name: "unsigned deadline", setup: func(w *monsterInversionTestWorld5408D0) { w.update.Field363 = math.MaxUint32 }, want: []string{"frame"}},
		{name: "cast object", setup: func(w *monsterInversionTestWorld5408D0) {
			w.update.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
		}, want: []string{"frame"}},
		{name: "cast location", setup: func(w *monsterInversionTestWorld5408D0) {
			w.update.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_LOCATION)
		}, want: []string{"frame"}},
		{name: "cast duration", setup: func(w *monsterInversionTestWorld5408D0) {
			w.update.AIStack[0].Action = uint32(ai.ACTION_CAST_DURATION_SPELL)
		}, want: []string{"frame"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newMonsterInversionTestWorld5408D0(t)
			tc.setup(w)
			before := *w.update
			if monsterCastInversion5408D0(w.unit, w.hooks(t)) || !reflect.DeepEqual(w.events, tc.want) || w.flag != 77 || *w.update != before {
				t.Fatalf("early gate changed state: events=%q flag=%d", w.events, w.flag)
			}
		})
	}
}

func TestMonsterCastInversion5408D0InertThreatAndSpellSetsDoNotConsumeRNG(t *testing.T) {
	for _, tc := range []string{"no missile", "non magic", "not missile", "other target", "no spell mask", "spell denied"} {
		t.Run(tc, func(t *testing.T) {
			w := newMonsterInversionTestWorld5408D0(t)
			h := w.hooks(t)
			wantFlag := uint32(0)
			switch tc {
			case "no missile":
				w.missiles = nil
			case "non magic":
				w.missiles = []*Object{{ObjClass: object.ClassMissile}}
			case "not missile":
				w.missiles = []*Object{{ObjSubClass: object.SubClass(object.MissileMagic)}}
			case "other target":
				w.missiles[0].UpdateDataMissile().Target = new(Object)
			case "no spell mask":
				for id := spell.ID(1); id < SpellsMax; id++ {
					monsterFightSetSpellFlag540B90(w.update, id, 0x10000000)
				}
				wantFlag = 1
			case "spell denied":
				h.spellAllowed = func(spell.ID) bool { return false }
				wantFlag = 1
			}
			if monsterCastInversion5408D0(w.unit, h) || w.randomCalls != 0 || w.flag != wantFlag || w.update.Field363 != 90 {
				t.Fatalf("inert selector consumed work: events=%q flag=%d", w.events, w.flag)
			}
		})
	}
}

func TestMonsterCastInversion5408D0DoesNotOwnAntiMagicOrWholeStackGate(t *testing.T) {
	for _, tc := range []string{"anti magic belongs to caller", "cast below non cast head", "deadline equality"} {
		t.Run(tc, func(t *testing.T) {
			w := newMonsterInversionTestWorld5408D0(t)
			switch tc {
			case "anti magic belongs to caller":
				w.unit.Buffs = 1 << ENCHANT_ANTI_MAGIC
			case "cast below non cast head":
				w.update.AIStack[0].Action = uint32(ai.ACTION_CAST_SPELL_ON_OBJECT)
				w.update.AIStackInd = 1
				w.update.AIStack[1].Action = uint32(ai.ACTION_WAIT)
			case "deadline equality":
				w.update.Field363 = w.frame
			}
			if !monsterCastInversion5408D0(w.unit, w.hooks(t)) {
				t.Fatal("selector applied a gate absent from 005408D0")
			}
		})
	}
}

func TestMonsterUnitIsMagicMissile540B60ExactResidualAndFaultGates(t *testing.T) {
	for _, tc := range []string{"non missile", "non magic", "other target", "targeted", "nil target matches", "missing update faults"} {
		t.Run(tc, func(t *testing.T) {
			unit, target := new(Object), new(Object)
			var want *Object
			flag := uint32(77)
			switch tc {
			case "non missile":
				unit.ObjSubClass = object.SubClass(object.MissileMagic)
				want = unit
			case "non magic":
				unit.ObjClass = object.ClassMissile
				want = unit
			default:
				unit.ObjClass = object.ClassMissile
				unit.ObjSubClass = object.SubClass(object.MissileMagic)
				data := &MissileUpdateData{Target: new(Object)}
				unit.UpdateData = unsafe.Pointer(data)
				want = target
				if tc == "targeted" {
					data.Target = target
				}
				if tc == "nil target matches" {
					target, data.Target, want = nil, nil, nil
				}
				if tc == "missing update faults" {
					unit.UpdateData = nil
				}
			}
			var got *Object
			var fault any
			func() {
				defer func() { fault = recover() }()
				got = monsterUnitIsMagicMissile540B60(unit, target, func(value uint32) { flag = value })
			}()
			if tc == "missing update faults" {
				if fault == nil || flag != 77 {
					t.Fatal("missing update was swallowed or wrote the scan flag")
				}
				return
			}
			wantFlag := uint32(77)
			if tc == "targeted" || tc == "nil target matches" {
				wantFlag = 1
			}
			if fault != nil || got != want || flag != wantFlag {
				t.Fatalf("callback result/flag/fault = %p/%d/%v", got, flag, fault)
			}
		})
	}
}

func TestMonsterCastInversion5408D0ContinuesEnumerationAfterMatchingMissile(t *testing.T) {
	w := newMonsterInversionTestWorld5408D0(t)
	w.missiles = append(w.missiles, w.missiles[0], &Object{ObjClass: object.ClassMissile})
	if !monsterCastInversion5408D0(w.unit, w.hooks(t)) {
		t.Fatal("targeted missiles were not found")
	}
	count := 0
	for _, event := range w.events {
		if event == "store-threat:1" {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("matching stores = %d, want 2 (no early break)", count)
	}
}
