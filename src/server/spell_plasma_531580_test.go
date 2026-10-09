package server

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type plasmaTestWorld531580 struct {
	hecubah, orb  uint32
	frame, fps    uint32
	objects       []*Object
	names         map[*Object]string
	weapons, rays map[*DurSpell]*Object
	events        []string
	facing        int
	balances      map[string]float32
	damageHook    func(*Object, *Object)
	stateHook     func(*Object)
}

func newPlasmaTestWorld531580() *plasmaTestWorld531580 {
	return &plasmaTestWorld531580{
		frame: 1000, fps: 30, names: make(map[*Object]string),
		weapons: make(map[*DurSpell]*Object), rays: make(map[*DurSpell]*Object),
		balances: map[string]float32{"PlasmaDamage": 2.75, "PlasmaDamageHecubah": 7.75, "PlasmaSearchTime": 12.75},
	}
}

func (w *plasmaTestWorld531580) unit(name string, class object.Class, x, y float32) *Object {
	u := &Object{ObjClass: class, PosVec: types.Ptf(x, y)}
	w.names[u] = name
	w.objects = append(w.objects, u)
	if class.Has(object.ClassPlayer) {
		u.UpdateData = unsafe.Pointer(&PlayerUpdateData{})
	}
	return u
}

func (w *plasmaTestWorld531580) staff(caster *Object, charge uint8) (*Object, *WandUseData) {
	wand := &Object{ObjClass: object.ClassWeapon, ObjSubClass: object.SubClass(0x4000000)}
	data := &WandUseData{Flags: 0x87654324, Charge: charge, MaxCharge: 4}
	wand.UseData.Ptr = unsafe.Pointer(data)
	caster.UpdateDataPlayer().EquippedWeapon = wand
	w.names[wand] = "wand"
	return wand, data
}

func (w *plasmaTestWorld531580) event(format string, args ...any) {
	w.events = append(w.events, fmt.Sprintf(format, args...))
}

func (w *plasmaTestWorld531580) runtime() SpellPlasmaRuntime531580 {
	return SpellPlasmaRuntime531580{
		HecubahType: &w.hecubah, HecubahOrbType: &w.orb,
		LookupType: func(name string) uint32 {
			w.event("lookup:%s", name)
			if name == "Hecubah" {
				return 700
			}
			return 701
		},
		Frame: func() uint32 { return w.frame }, TickRate: func() uint32 { return w.fps },
		Balance: func(key string) float32 { w.event("balance:%s", key); return w.balances[key] },
		ObjectsInCircle: func(pos types.Pointf, radius float32, visit func(*Object) bool) {
			w.event("circle:%g", radius)
			for _, u := range w.objects {
				if !visit(u) {
					break
				}
			}
		},
		IsEnemy:       func(a, b *Object) bool { w.event("enemy:%s", w.names[b]); return a != b },
		CanInteract:   func(a, b *Object) bool { w.event("los:%s", w.names[b]); return true },
		Facing:        func(a, b *Object) int { w.event("facing:%s", w.names[b]); return w.facing },
		Distance:      ObjectDistance4E6C00,
		PositionDelta: func(*Object, *types.Pointf) int32 { w.event("position"); return 0 },
		StartRay:      func(r *DurSpell) { w.event("start:%s", w.names[r.Target48]) },
		StopRay:       func(r *DurSpell, target *Object) { w.event("stop:%s", w.names[target]) },
		PointFX:       func(code uint8, pos types.Pointf) { w.event("fx:%d:%g,%g", code, pos.X, pos.Y) },
		Damage: func(target, caster *Object, amount int32) {
			w.event("damage:%s:%d", w.names[target], amount)
			if w.damageHook != nil {
				w.damageHook(target, caster)
			}
		},
		Audio: func(id uint16, target *Object) { w.event("audio:%d:%s", id, w.names[target]) },
		SetPlayerState: func(caster *Object, state PlayerState) {
			w.event("state:%s:%d", w.names[caster], state)
			if w.stateHook != nil {
				w.stateHook(caster)
			}
		},
		ReportCharges: func(update *PlayerUpdateData, wand *Object, charge, max uint8) {
			w.event("charge:%s:%d/%d", w.names[wand], charge, max)
		},
		LoadWeapon: func(r *DurSpell) *Object { return w.weapons[r] },
		StoreWeapon: func(r *DurSpell, wand *Object) {
			if wand == nil {
				delete(w.weapons, r)
			} else {
				w.weapons[r] = wand
			}
		},
		LoadRayTarget: func(r *DurSpell) *Object { return w.rays[r] },
		StoreRayTarget: func(r *DurSpell, target *Object) {
			if target == nil {
				delete(w.rays, r)
			} else {
				w.rays[r] = target
			}
		},
	}
}

func TestSpellPlasmaCreate531580NativeWeaponAndLiveCaster(t *testing.T) {
	w := newPlasmaTestWorld531580()
	caster := w.unit("caster", object.ClassPlayer, 10, 20)
	wand, data := w.staff(caster, 3)
	r := &DurSpell{Field72: -1, Field76: 99, Flags88: 0x11220001, Pos: types.Ptf(11, 22), Target48: caster}
	rt := w.runtime()
	fx := rt.PointFX
	rt.PointFX = func(code uint8, pos types.Pointf) {
		if r.Field72 != 0 || r.Field76 != 0 || r.Target48 != nil || w.weapons[r] != nil {
			t.Fatal("initialization must precede FX")
		}
		fx(code, pos)
		r.Caster16 = caster
	}
	if got := SpellPlasmaCreate531580(r, rt); got != 0 {
		t.Fatalf("create = %d", got)
	}
	if r.Flags88 != 0x11220003 || r.Field72 != 0 || w.weapons[r] != wand || data.Flags != 0x87654324 {
		t.Fatalf("record/wand = %#v/%p/%#x", r, w.weapons[r], data.Flags)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(wand)) <= math.MaxUint32 {
		t.Fatalf("fixture wand must exceed PE32: %p", wand)
	}
	if !reflect.DeepEqual(w.events, []string{"fx:131:11,22"}) {
		t.Fatal(w.events)
	}
	SpellPlasmaDestroy5319E0(r, rt)
	if w.weapons[r] != nil || data.Flags != 0x87654320 {
		t.Fatalf("destroy = %p/%#x", w.weapons[r], data.Flags)
	}
}

func TestSpellPlasmaCreate531580Branches(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		class                  object.Class
		weapon, plasma, active bool
		want                   int32
	}{
		{"orphan", 0, false, false, false, 0},
		{"monster", object.ClassMonster, false, false, false, 0},
		{"no-weapon", object.ClassPlayer, false, false, false, 1},
		{"wrong-weapon", object.ClassPlayer, true, false, true, 1},
		{"inactive-staff", object.ClassPlayer, true, true, false, 0},
		{"active-staff", object.ClassPlayer, true, true, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newPlasmaTestWorld531580()
			r := &DurSpell{Field72: 123, Field76: 456}
			if tc.class != 0 {
				r.Caster16 = w.unit("caster", tc.class, 0, 0)
			}
			if tc.weapon {
				wand, data := w.staff(r.Caster16, 3)
				if !tc.plasma {
					wand.ObjSubClass = 0
				}
				if !tc.active {
					data.Flags &^= 4
				}
			}
			if got := SpellPlasmaCreate531580(r, w.runtime()); got != tc.want {
				t.Fatalf("create = %d, want %d", got, tc.want)
			}
			if (w.weapons[r] != nil) != (tc.weapon && tc.plasma && tc.active) {
				t.Fatal("incorrect sidecar")
			}
		})
	}
}

func TestSpellPlasmaUpdate531600InitialCacheAndCancellation(t *testing.T) {
	for _, branch := range []string{"orphan", "cancel", "monster-moved", "high-flag-byte"} {
		t.Run(branch, func(t *testing.T) {
			w := newPlasmaTestWorld531580()
			r := &DurSpell{}
			rt := w.runtime()
			want := int32(1)
			if branch != "orphan" {
				r.Caster16 = w.unit("caster", object.ClassPlayer, 0, 0)
			}
			if branch == "cancel" {
				r.Flags88 = 0x20
			}
			if branch == "monster-moved" {
				r.Caster16.ObjClass = object.ClassMonster
				rt.PositionDelta = func(*Object, *types.Pointf) int32 { w.event("moved"); return 1 }
			}
			if branch == "high-flag-byte" {
				r.Flags88 = 0x2000
				w.objects = nil
				want = 0
			}
			if got := SpellPlasmaUpdate531600(r, rt); got != want {
				t.Fatalf("update = %d", got)
			}
			if w.hecubah != 700 || w.orb != 701 || len(w.events) < 2 || w.events[0] != "lookup:Hecubah" || w.events[1] != "lookup:HecubahWithOrb" {
				t.Fatal(w.events)
			}
			w.events = nil
			SpellPlasmaUpdate531600(r, rt)
			for _, e := range w.events {
				if e == "lookup:Hecubah" {
					t.Fatal("cache initialized twice")
				}
			}
		})
	}
}

func TestSpellPlasmaUpdate531600DamageRayChargeOrder(t *testing.T) {
	w := newPlasmaTestWorld531580()
	caster := w.unit("caster", object.ClassPlayer, 0, 0)
	target := w.unit("target", object.ClassMonster, 100, 0)
	wand, data := w.staff(caster, 2)
	caster.UpdateDataPlayer().CursorObj = target
	r := &DurSpell{Caster16: caster, Field36: 0x12345678, Spell: 59, Frame68: 444}
	rt := w.runtime()
	SpellPlasmaCreate531580(r, rt)
	w.events = nil
	if got := SpellPlasmaUpdate531600(r, rt); got != 0 {
		t.Fatal(got)
	}
	want := []string{"lookup:Hecubah", "lookup:HecubahWithOrb", "enemy:target", "start:target", "balance:PlasmaDamage", "damage:target:2", "state:caster:22", "audio:98:caster", "audio:98:target", "balance:PlasmaSearchTime", "state:caster:22", "charge:wand:1/4"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if r.Target48 != target || w.rays[r] != target || w.weapons[r] != wand || r.Frame68 != 1012 || data.Charge != 1 || data.Progress != 25 || r.Field36 != 0x12345678 {
		t.Fatalf("record/charge = %#v/%#v", r, data)
	}
	w.events = nil
	w.frame++
	if got := SpellPlasmaUpdate531600(r, rt); got != 1 {
		t.Fatalf("last charge result = %d", got)
	}
	if data.Charge != 0 || data.Progress != 0 {
		t.Fatal(data)
	}
	for _, e := range w.events {
		if e == "start:target" || e == "audio:98:target" {
			t.Fatalf("retained ray/audio cadence = %v", w.events)
		}
	}
	SpellPlasmaDestroy5319E0(r, rt)
	if w.weapons[r] != nil || data.Flags&4 != 0 {
		t.Fatal("active wand not released")
	}
}

func TestSpellPlasmaUpdate531600NoTargetDoesNotConsumeCharge(t *testing.T) {
	w := newPlasmaTestWorld531580()
	caster := w.unit("caster", object.ClassPlayer, 0, 0)
	_, data := w.staff(caster, 3)
	r := &DurSpell{Caster16: caster, Frame68: 77}
	rt := w.runtime()
	SpellPlasmaCreate531580(r, rt)
	w.rays[r] = w.unit("previous", object.ClassMonster, 10, 0)
	w.objects = []*Object{caster}
	if got := SpellPlasmaUpdate531600(r, rt); got != 0 {
		t.Fatal(got)
	}
	if data.Charge != 3 || w.rays[r] != nil || r.Frame68 != 77 {
		t.Fatal("idle ray consumed charge or refreshed expiry")
	}
	if w.events[len(w.events)-1] != "stop:previous" {
		t.Fatal(w.events)
	}
}

func TestSpellPlasmaUpdate531600GracePeriodAndReacquisition(t *testing.T) {
	w := newPlasmaTestWorld531580()
	caster := w.unit("caster", object.ClassPlayer, 0, 0)
	target := w.unit("target", object.ClassMonster, 100, 0)
	r := &DurSpell{Caster16: caster, Target48: target, Frame68: 9999}
	w.rays[r] = target
	w.facing = 2
	rt := w.runtime()
	for i := 0; i < 90; i++ {
		w.frame++
		if got := SpellPlasmaUpdate531600(r, rt); got != 0 {
			t.Fatal(got)
		}
		want := uintptr(89 - i)
		if r.Field76 != want || (r.Target48 != nil) != (i < 89) || r.Frame68 != 9999 {
			t.Fatalf("grace tick %d: %d/%p/%d", i, r.Field76, r.Target48, r.Frame68)
		}
	}
	if w.rays[r] != nil || w.events[len(w.events)-1] != "stop:target" {
		t.Fatal("grace did not stop old ray")
	}
	SpellPlasmaUpdate531600(r, rt)
	if r.Target48 != target || r.Field76 != 0 || r.Frame68 != w.frame+12 {
		t.Fatal("no reacquisition after grace expiry")
	}
	w.facing = 0
	r.Field76 = 25
	SpellPlasmaUpdate531600(r, rt)
	if r.Field76 != 0 {
		t.Fatal("facing restored without resetting grace")
	}
}

func TestSpellPlasmaUpdate531600CursorUsesSurfaceDistanceWithoutLOS(t *testing.T) {
	w := newPlasmaTestWorld531580()
	caster := w.unit("caster", object.ClassPlayer, 0, 0)
	target := w.unit("large", object.ClassMonster, 410, 0)
	caster.UpdateDataPlayer().CursorObj = target
	r := &DurSpell{Caster16: caster}
	rt := w.runtime()
	rt.Distance = func(a, b *Object) float64 {
		if a != caster || b != target {
			t.Fatal("wrong cursor distance operands")
		}
		return 400
	}
	rt.Facing = func(*Object, *Object) int { t.Fatal("cursor acquisition checked facing"); return 0 }
	rt.CanInteract = func(*Object, *Object) bool { t.Fatal("cursor acquisition checked LOS"); return false }
	if got := SpellPlasmaUpdate531600(r, rt); got != 0 || r.Target48 != target {
		t.Fatal("cursor's inclusive surface-distance boundary rejected")
	}
}

func TestSpellPlasmaClosest531920EligibilityAndOriginalFacing(t *testing.T) {
	for _, facing := range []int{0, 1, 2, 4, 8, 12, 13, 15} {
		t.Run(fmt.Sprint(facing), func(t *testing.T) {
			w := newPlasmaTestWorld531580()
			caster := w.unit("caster", object.ClassPlayer, 0, 0)
			w.unit("item", object.ClassWeapon, 1, 0)
			dead := w.unit("dead", object.ClassMonster, 2, 0)
			dead.ObjFlags = object.Flags(0x20)
			excluded := w.unit("excluded", object.ClassMonster, 3, 0)
			excluded.ObjSubClass = object.SubClass(0x8000)
			w.unit("blocked", object.ClassMonster, 4, 0)
			w.unit("ally", object.ClassMonster, 5, 0)
			want := w.unit("nearest", object.ClassPlayer, 20, 0)
			w.unit("farther", object.ClassMonster, 30, 0)
			w.facing = facing
			rt := w.runtime()
			rt.IsEnemy = func(a, b *Object) bool { return w.names[b] != "ally" }
			rt.CanInteract = func(a, b *Object) bool { return w.names[b] != "blocked" }
			if got := plasmaClosest531920(caster, rt); got != want {
				t.Fatalf("closest = %s", w.names[got])
			}
		})
	}
}

func TestSpellPlasmaClosest531920StrictBoundaryTieAndUnordered(t *testing.T) {
	w := newPlasmaTestWorld531580()
	caster := w.unit("caster", object.ClassPlayer, 0, 0)
	boundary := w.unit("boundary", object.ClassMonster, 400, 0)
	if got := plasmaClosest531920(caster, w.runtime()); got != nil {
		t.Fatalf("strict 400^2 bound selected %p", got)
	}
	boundary.PosVec.X = 20
	tie := w.unit("tie", object.ClassMonster, -20, 0)
	if got := plasmaClosest531920(caster, w.runtime()); got != boundary {
		t.Fatal("equal distance replaced first target", got, tie)
	}
	nan := w.unit("unordered", object.ClassMonster, float32(math.NaN()), 0)
	if got := plasmaClosest531920(caster, w.runtime()); got != nan {
		t.Fatal("FCOM unordered branch not preserved")
	}
}

func TestSpellPlasmaUpdate531600HecubahDamageAndDeathEffect(t *testing.T) {
	for _, typ := range []uint16{700, 701, 702} {
		t.Run(fmt.Sprint(typ), func(t *testing.T) {
			w := newPlasmaTestWorld531580()
			caster := w.unit("caster", object.ClassPlayer, 0, 0)
			target := w.unit("target", object.ClassMonster, 100, 0)
			target.TypeInd = typ
			r := &DurSpell{Caster16: caster, Target48: target}
			w.damageHook = func(u, _ *Object) { u.ObjFlags |= object.Flags(0x8000) }
			SpellPlasmaUpdate531600(r, w.runtime())
			want := "damage:target:7"
			if typ == 702 {
				want = "damage:target:2"
			}
			seen := false
			for i, e := range w.events {
				if e == want {
					seen = true
					if w.events[i+1] != "fx:131:100,0" {
						t.Fatal(w.events)
					}
				}
			}
			if !seen {
				t.Fatal(w.events)
			}
		})
	}
}

func TestSpellPlasmaUpdate531600LiveReloadsAfterCallbacks(t *testing.T) {
	w := newPlasmaTestWorld531580()
	caster := w.unit("caster", object.ClassPlayer, 0, 0)
	nextCaster := w.unit("next-caster", object.ClassPlayer, 0, 0)
	target := w.unit("target", object.ClassMonster, 100, 0)
	replacement := w.unit("replacement", object.ClassMonster, 120, 0)
	_, data := w.staff(caster, 2)
	nextWand, _ := w.staff(nextCaster, 4)
	w.names[nextWand] = "next-wand"
	r := &DurSpell{Caster16: caster, Target48: target}
	rt := w.runtime()
	SpellPlasmaCreate531580(r, rt)
	r.Target48 = target
	w.damageHook = func(*Object, *Object) { r.Caster16 = nextCaster; r.Target48 = replacement }
	stateCalls := 0
	w.stateHook = func(*Object) {
		stateCalls++
		if stateCalls == 2 {
			r.Caster16 = caster
			nextCaster.UpdateData = nil
			w.weapons[r] = nextWand
		}
	}
	updateSnapshot := nextCaster.UpdateDataPlayer()
	rt.ReportCharges = func(update *PlayerUpdateData, wand *Object, charge, max uint8) {
		if update != updateSnapshot || wand != nextWand || charge != 1 || max != 4 {
			t.Fatal("charge report did not retain update/data but reload weapon")
		}
	}
	if got := SpellPlasmaUpdate531600(r, rt); got != 0 {
		t.Fatal(got)
	}
	if w.rays[r] != replacement || data.Charge != 1 {
		t.Fatal("damage callback target/caster reload lost")
	}
}

func TestPlasmaBalanceInt531600ChopAndIndefinite(t *testing.T) {
	for _, tc := range []struct {
		value float32
		want  int32
	}{
		{2.75, 2}, {-2.75, -2}, {0.99, 0}, {-0.99, 0},
		{2147483520, 2147483520}, {-2147483648, math.MinInt32},
		{2147483648, math.MinInt32}, {float32(math.Inf(1)), math.MinInt32},
		{float32(math.Inf(-1)), math.MinInt32}, {float32(math.NaN()), math.MinInt32},
	} {
		if got := plasmaBalanceInt531600(tc.value); got != tc.want {
			t.Fatalf("chop(%g) = %d, want %d", tc.value, got, tc.want)
		}
	}
}

func TestSpellPlasmaUpdate531600DepletedWandStillStrikesBeforeCancel(t *testing.T) {
	w := newPlasmaTestWorld531580()
	caster := w.unit("caster", object.ClassPlayer, 0, 0)
	target := w.unit("target", object.ClassMonster, 100, 0)
	_, data := w.staff(caster, 0)
	r := &DurSpell{Caster16: caster}
	rt := w.runtime()
	SpellPlasmaCreate531580(r, rt)
	r.Target48 = target
	if got := SpellPlasmaUpdate531600(r, rt); got != 1 {
		t.Fatal(got)
	}
	if data.Charge != 0 {
		t.Fatal("empty charge wrapped")
	}
	seen := false
	for _, e := range w.events {
		if e == "damage:target:2" {
			seen = true
		}
		if strings.HasPrefix(e, "charge:") {
			t.Fatal("depleted wand reported extra consumption")
		}
	}
	if !seen {
		t.Fatal("original pre-charge damage order changed")
	}
}

func TestSpellPlasmaUpdate531600ExpiryUsesDWORDWrap(t *testing.T) {
	w := newPlasmaTestWorld531580()
	w.frame = math.MaxUint32 - 4
	caster := w.unit("caster", object.ClassPlayer, 0, 0)
	target := w.unit("target", object.ClassMonster, 100, 0)
	r := &DurSpell{Caster16: caster, Target48: target}
	if got := SpellPlasmaUpdate531600(r, w.runtime()); got != 0 || r.Frame68 != 7 {
		t.Fatalf("expiry = %d, result=%d", r.Frame68, got)
	}
}
