package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type chainLightningTestWorld52F820 struct {
	objects  []*Object
	names    map[*Object]string
	weapons  map[*DurSpell]*Object
	events   []string
	frame    uint32
	levels   []uint32
	balances map[string]float32
}

func newChainLightningTestWorld52F820() *chainLightningTestWorld52F820 {
	return &chainLightningTestWorld52F820{
		names: make(map[*Object]string), weapons: make(map[*DurSpell]*Object),
		frame: 1000,
		balances: map[string]float32{
			"LightningRange": 200, "LightningDamage": 1.5,
			"LightningGlyphDamage": 5, "LightningSearchTime": 12,
		},
	}
}

func (w *chainLightningTestWorld52F820) object(name string, class object.Class, x, y float32) *Object {
	obj := &Object{ObjClass: class, PosVec: types.Ptf(x, y)}
	w.names[obj] = name
	w.objects = append(w.objects, obj)
	return obj
}

func (w *chainLightningTestWorld52F820) event(format string, args ...any) {
	w.events = append(w.events, fmt.Sprintf(format, args...))
}

func (w *chainLightningTestWorld52F820) runtime() SpellChainLightningRuntime52F820 {
	return SpellChainLightningRuntime52F820{
		Frame:    func() uint32 { return w.frame },
		TickRate: func() uint32 { return 30 },
		Balance:  func(key string) float32 { return w.balances[key] },
		TargetLimit: func(level uint32) uint32 {
			w.levels = append(w.levels, level)
			return level
		},
		ObjectsInCircle: func(center types.Pointf, radius float32, visit func(*Object) bool) {
			for _, obj := range w.objects {
				dx, dy := obj.PosVec.X-center.X, obj.PosVec.Y-center.Y
				if dx*dx+dy*dy <= radius*radius && !visit(obj) {
					return
				}
			}
		},
		CanInteract:   func(_, _ *Object) bool { return true },
		IsEnemy:       func(_, _ *Object) bool { return true },
		TraceRay:      func(_, _ types.Pointf, _ MapTraceFlags) bool { return true },
		PositionDelta: func(*Object, *types.Pointf) int32 { return 0 },
		CancelSpell:   func(id int32, caster *Object) { w.event("cancel:%d:%s", id, w.names[caster]) },
		FreeDuration:  func(ray *DurSpell) { w.event("free:%s", w.names[ray.Target48]) },
		NewLightningSub: func(record *DurSpell, from, to *Object) {
			ray := &DurSpell{Caster16: from, Target48: to, Spell: 7, Next: record.Sub108}
			record.Sub108 = ray
			w.event("ray:%s>%s", w.names[from], w.names[to])
		},
		StartRay:       func(ray *DurSpell) { w.event("start:%s>%s", w.names[ray.Caster16], w.names[ray.Target48]) },
		StopRay:        func(ray *DurSpell, target *Object) { w.event("stop:%s>%s", w.names[ray.Caster16], w.names[target]) },
		PointFX:        func(code uint8, pos types.Pointf) { w.event("fx:%d:%g,%g", code, pos.X, pos.Y) },
		Damage:         func(target, caster *Object, amount int32) { w.event("damage:%s:%d", w.names[target], amount) },
		Audio:          func(id uint16, target *Object) { w.event("audio:%d:%s", id, w.names[target]) },
		CastSound:      func() uint16 { return 78 },
		SetPlayerState: func(caster *Object, state PlayerState) { w.event("state:%s:%d", w.names[caster], state) },
		ManaSub:        func(caster *Object, amount int32) { w.event("mana:%s:%d", w.names[caster], amount) },
		ReportCharges: func(caster, wand *Object, charge, maxCharge uint8) {
			w.event("charge:%s:%s:%d/%d", w.names[caster], w.names[wand], charge, maxCharge)
		},
		LoadWeapon: func(record *DurSpell) *Object { return w.weapons[record] },
		StoreWeapon: func(record *DurSpell, wand *Object) {
			if wand == nil {
				delete(w.weapons, record)
			} else {
				w.weapons[record] = wand
			}
		},
	}
}

func TestSpellChainLightningCreateDestroy52F820KeepsNativeWandPointer(t *testing.T) {
	w := newChainLightningTestWorld52F820()
	caster := w.object("caster", object.ClassPlayer, 10, 20)
	wand := w.object("wand", object.ClassWeapon, 0, 0)
	wand.ObjSubClass = object.SubClass(0x40000)
	data := &WandUseData{Flags: 4, Charge: 3, MaxCharge: 4}
	wand.UseData.Ptr = unsafe.Pointer(data)
	caster.UpdateData = unsafe.Pointer(&PlayerUpdateData{EquippedWeapon: wand})
	record := &DurSpell{Caster16: caster, Spell: 43, Pos: types.Ptf(11, 22)}
	if got := SpellChainLightningCreate52F820(record, w.runtime()); got != 0 {
		t.Fatalf("create = %d", got)
	}
	if w.weapons[record] != wand || record.Flags88&2 == 0 || record.Field72 != 0 {
		t.Fatalf("wand sidecar/flags = %p/%#x/%#x", w.weapons[record], record.Flags88, record.Field72)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(wand)) <= math.MaxUint32 {
		t.Fatalf("test wand was not allocated above the PE32 address range: %p", wand)
	}
	SpellChainLightningDestroy530100(record, w.runtime())
	if w.weapons[record] != nil || data.Flags&4 != 0 || record.Sub104 != nil || record.Sub108 != nil {
		t.Fatalf("destroy sidecar/flags/lists = %p/%#x/%p/%p", w.weapons[record], data.Flags, record.Sub104, record.Sub108)
	}
	want := []string{"cancel:24:caster", "fx:129:11,22"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestSpellChainLightningUpdate52F8A0BounceAndRayReuse(t *testing.T) {
	w := newChainLightningTestWorld52F820()
	caster := w.object("caster", object.ClassMonster, 0, 0)
	w.object("first", object.ClassMonster, 100, 0)
	w.object("second", object.ClassMonster, 130, 0)
	w.object("third", object.ClassMonster, 160, 0)
	record := &DurSpell{Caster16: caster, Spell: 43, Level: 3, Pos: caster.PosVec, Frame60: 1000}
	if got := SpellChainLightningUpdate52F8A0(record, w.runtime()); got != 0 {
		t.Fatalf("first update = %d", got)
	}
	if record.Frame68 != 1012 || math.Float32frombits(uint32(record.Field76)) != -0.5 {
		t.Fatalf("frame/carry = %d/%g", record.Frame68, math.Float32frombits(uint32(record.Field76)))
	}
	firstEvents := append([]string(nil), w.events...)
	wantFirst := []string{
		"ray:caster>first", "ray:first>second", "ray:first>third",
		"start:first>third", "damage:third:2",
		"start:first>second", "damage:second:2",
		"start:caster>first", "damage:first:2",
		"audio:78:caster", "audio:78:first",
	}
	if !reflect.DeepEqual(firstEvents, wantFirst) {
		t.Fatalf("first events = %v, want %v", firstEvents, wantFirst)
	}
	w.events = nil
	w.frame = 1001
	if got := SpellChainLightningUpdate52F8A0(record, w.runtime()); got != 0 {
		t.Fatalf("second update = %d", got)
	}
	for _, event := range w.events {
		if len(event) >= 6 && (event[:6] == "start:" || event[:5] == "stop:") {
			t.Fatalf("unchanged ray emitted network event: %v", w.events)
		}
	}
	if math.Float32frombits(uint32(record.Field76)) != 0 {
		t.Fatalf("second carry = %g", math.Float32frombits(uint32(record.Field76)))
	}
	SpellChainLightningDestroy530100(record, w.runtime())
	if record.Sub104 != nil || record.Sub108 != nil {
		t.Fatal("ray generations retained after destroy")
	}
}

func TestSpellChainLightningUpdate52F8A0LimitsTargetsByRecordLevel(t *testing.T) {
	for level := uint32(1); level <= 5; level++ {
		t.Run(fmt.Sprintf("level_%d", level), func(t *testing.T) {
			w := newChainLightningTestWorld52F820()
			caster := w.object("caster", object.ClassMonster, 0, 0)
			for i := 1; i <= 5; i++ {
				w.object(fmt.Sprintf("target%d", i), object.ClassMonster, float32(i*20), 0)
			}
			record := &DurSpell{
				Caster16: caster,
				Spell:    43,
				Level:    level,
				Pos:      caster.PosVec,
				Frame60:  w.frame,
			}
			if got := SpellChainLightningUpdate52F8A0(record, w.runtime()); got != 0 {
				t.Fatalf("update = %d", got)
			}
			if !reflect.DeepEqual(w.levels, []uint32{level}) {
				t.Fatalf("target-limit lookup = %v, want record level %d", w.levels, level)
			}
			var rays, damages int
			for _, event := range w.events {
				if len(event) >= 4 && event[:4] == "ray:" {
					rays++
				}
				if len(event) >= 7 && event[:7] == "damage:" {
					damages++
				}
			}
			if rays != int(level) || damages != int(level) {
				t.Fatalf("level %d emitted %d rays and %d damage calls; events = %v", level, rays, damages, w.events)
			}
		})
	}
}

func TestSpellChainLightningTrap52F8A0DamagesVisibleTargets(t *testing.T) {
	w := newChainLightningTestWorld52F820()
	w.object("target", object.ClassMonster, 20, 0)
	w.object("outside", object.ClassMonster, 300, 0)
	record := &DurSpell{Spell: 43, Flag20: 1, Pos: types.Ptf(0, 0)}
	if got := SpellChainLightningUpdate52F8A0(record, w.runtime()); got != 1 {
		t.Fatalf("trap update = %d", got)
	}
	want := []string{"damage:target:5", "fx:129:20,0", "audio:78:target"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}
