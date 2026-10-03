package server

import (
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func defaultDamagePreEffectFixture4E0B30() *ModifierEff {
	return &ModifierEff{AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
}

func TestDefaultDamageWorld4E0B30PreDamageAfterEarlierCallbacks(t *testing.T) {
	for _, stage := range []string{"protection", "buff off", "late defend"} {
		t.Run(stage, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
			stale, live := defaultDamagePreEffectFixture4E0B30(), defaultDamagePreEffectFixture4E0B30()
			weapon := &Object{TypeInd: 777, ObjClass: object.ClassWand,
				InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{stale}})}
			liveBase := &ModifierInitData{Modifiers: [4]*ModifierEff{3: live}}
			typ, input, wantPreDamage := object.DamageBlade, int32(8), int32(0)
			if stage != "late defend" {
				target.ObjClass, target.UpdateData = object.ClassObstacle, nil
				typ, wantPreDamage = object.DamageFlame, 4
			} else {
				target.InvFirstItem = &Object{ObjFlags: object.FlagEquipped,
					InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{2: defaultDamageLateEffectFixture4E0B30()}})}
			}
			target.Obj130, target.Field131, target.Frame134 = source, 99, 77
			var events []string
			r := damageMeleeWorldRuntime4E0B30(t)
			r.FireProtection = func(*Object) float64 {
				events = append(events, "protection")
				if stage == "protection" {
					weapon.InitData = unsafe.Pointer(liveBase)
				}
				return 0.5
			}
			r.BuffOff = func(*Object, EnchantID) {
				events = append(events, "buff")
				if stage == "buff off" {
					weapon.InitData = unsafe.Pointer(liveBase)
				}
			}
			r.CanApplyLateDefend = func(*ModifierEff) bool { return true }
			r.ApplyLateDefend = func(_ *ModifierEff, _, owner, w, attacker *Object, d int32, gotType object.DamageType) int32 {
				if owner != target || w != weapon || attacker != source || d != input || gotType != typ || target.Field131 != 99 || target.Frame134 != 77 {
					t.Fatal("late Defend prefix changed")
				}
				events = append(events, "defend")
				weapon.InitData = unsafe.Pointer(liveBase)
				return 0
			}
			r.CanApplyPreDamage = func(*ModifierEff) bool { return true }
			r.ApplyPreDamage = func(m *ModifierEff, w, attacker, owner *Object, d *int32) {
				if m != live || w != weapon || attacker != source || owner != target || *d != wantPreDamage ||
					target.Obj130 != weapon || target.Field131 != uint32(typ) || target.Frame134 != 1400 || target.Pos132 != source.PrevPos {
					t.Fatal("pre-Damage used an eager effect plan or ran before attribution")
				}
				if stage == "late defend" {
					ud := target.UpdateDataMonster()
					if ud.Field547 != 1 || ud.Field546 != 777 || !ud.StatusFlags.Has(object.MonStatusInjured) {
						t.Fatal("pre-Damage preceded injured latch")
					}
				}
				events = append(events, "live pre")
				*d = -7
			}
			r.DefaultDamageSound = func(*Object, *Object) { events = append(events, "sound") }
			r.DamageClear = func(owner *Object, d int32) {
				if owner != target || d != -7 {
					t.Fatal("raw pre-Damage word did not reach HP tail")
				}
				events = append(events, "HP")
			}
			if !DefaultDamageWorld4E0B30(target, source, weapon, input, typ, r) {
				t.Fatal("tail result")
			}
			want := []string{"protection", "buff", "live pre", "sound", "HP"}
			if stage == "late defend" {
				want = []string{"buff", "defend", "live pre", "sound", "HP"}
			}
			if !slices.Equal(events, want) {
				t.Fatalf("events=%v, want %v", events, want)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PreDamageLiveSlotsCachedBaseAndRawWord(t *testing.T) {
	target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
	first, stale, second, third, fourth := defaultDamagePreEffectFixture4E0B30(), defaultDamagePreEffectFixture4E0B30(), defaultDamagePreEffectFixture4E0B30(), defaultDamagePreEffectFixture4E0B30(), defaultDamagePreEffectFixture4E0B30()
	base := &ModifierInitData{Modifiers: [4]*ModifierEff{first, stale, nil, stale}}
	weapon := &Object{TypeInd: 777, ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(base)}
	var calls []*ModifierEff
	var originalAddress *int32
	r := damageMeleeWorldRuntime4E0B30(t)
	r.CanApplyPreDamage = func(*ModifierEff) bool { return true }
	r.ApplyPreDamage = func(m *ModifierEff, w, attacker, owner *Object, d *int32) {
		if w != weapon || attacker != source || owner != target || target.Obj130 != weapon || target.Frame134 != 1400 {
			t.Fatal("pre-Damage arguments/prefix")
		}
		if originalAddress == nil {
			originalAddress = d
		} else if originalAddress != d {
			t.Fatal("damage address changed")
		}
		calls = append(calls, m)
		switch m {
		case first:
			if *d != 8 {
				t.Fatal("first raw word")
			}
			base.Modifiers[1] = second
			weapon.InitData = unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{stale, stale, stale, stale}})
			*d = 0
		case second:
			if *d != 0 {
				t.Fatal("zero was clamped or stale base read")
			}
			base.Modifiers[2], *d = third, -7
		case third:
			if *d != -7 {
				t.Fatal("negative word changed")
			}
			base.Modifiers[3], *d = fourth, math.MinInt32
		case fourth:
			if *d != math.MinInt32 {
				t.Fatal("signed DWORD narrowed")
			}
			*d = math.MaxInt32
		default:
			t.Fatal("snapshot slot or replacement base executed")
		}
	}
	r.DamageClear = func(owner *Object, d int32) {
		if owner != target || d != math.MaxInt32 {
			t.Fatal("final pre-Damage DWORD changed")
		}
	}
	if !DefaultDamageWorld4E0B30(target, source, weapon, 8, object.DamageBlade, r) || !slices.Equal(calls, []*ModifierEff{first, second, third, fourth}) {
		t.Fatalf("live slots=%v", calls)
	}
}

func TestDefaultDamageWorld4E0B30PreDamageLiveWeaponClassGate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		entry, live object.Class
		want        bool
	}{
		{"gains Weapon", 0, object.ClassWeapon, true}, {"gains Wand", 0, object.ClassWand, true},
		{"loses Weapon", object.ClassWeapon, 0, false}, {"loses Wand", object.ClassWand, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source := &Object{ObjClass: object.ClassObstacle}, damageMeleeUnitFixture4E17B0(t, true)
			m := defaultDamagePreEffectFixture4E0B30()
			weapon := &Object{ObjClass: tc.entry, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{m}})}
			r := damageMeleeWorldRuntime4E0B30(t)
			r.FireProtection = func(*Object) float64 {
				weapon.ObjClass = tc.live
				if !tc.want {
					weapon.InitData = nil
				} // A skipped helper must not load this base.
				return 0
			}
			r.CanApplyPreDamage = func(*ModifierEff) bool { return true }
			calls := 0
			r.ApplyPreDamage = func(got *ModifierEff, _, _, _ *Object, d *int32) {
				if !tc.want || got != m || *d != 8 {
					t.Fatal("entry-time class gate used")
				}
				calls++
			}
			r.DamageClear = func(*Object, int32) {}
			if !DefaultDamageWorld4E0B30(target, source, weapon, 8, object.DamageFlame, r) || calls != map[bool]int{false: 0, true: 1}[tc.want] {
				t.Fatalf("calls=%d want=%t", calls, tc.want)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PreDamageNilInitFaultAtUse(t *testing.T) {
	for _, skip := range []bool{false, true} {
		t.Run(map[bool]string{false: "eligible faults after attribution", true: "excluded class skips helper"}[skip], func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
			weapon := &Object{TypeInd: 777, ObjClass: object.ClassWand}
			r := damageMeleeWorldRuntime4E0B30(t)
			var events []string
			r.BuffOff = func(*Object, EnchantID) {
				events = append(events, "buff")
				if skip {
					weapon.ObjClass = 0
				}
			}
			r.DefaultDamageSound = func(*Object, *Object) { events = append(events, "sound") }
			r.DamageClear = func(*Object, int32) { events = append(events, "HP") }
			var caught any
			func() {
				defer func() { caught = recover() }()
				DefaultDamageWorld4E0B30(target, source, weapon, 3, object.DamageBlade, r)
			}()
			want := []string{"buff"}
			if skip {
				want = []string{"buff", "sound", "HP"}
			}
			ud := target.UpdateDataMonster()
			if (caught == nil) != skip || !slices.Equal(events, want) || target.Obj130 != weapon || target.Frame134 != 1400 || target.Field131 != uint32(object.DamageBlade) ||
				ud.Field547 != 1 || ud.Field546 != 777 || !ud.StatusFlags.Has(object.MonStatusInjured) || target.HealthData.Cur != 200 {
				t.Fatalf("at-use prefix: caught=%v events=%v", caught, events)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PreDamageIntroducedUnknownContinues(t *testing.T) {
	target, source := &Object{ObjClass: object.ClassObstacle}, damageMeleeUnitFixture4E17B0(t, true)
	first, unknown, last := defaultDamagePreEffectFixture4E0B30(), defaultDamagePreEffectFixture4E0B30(), defaultDamagePreEffectFixture4E0B30()
	base := &ModifierInitData{Modifiers: [4]*ModifierEff{first, nil, nil, last}}
	weapon := &Object{ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(base)}
	r := damageMeleeWorldRuntime4E0B30(t)
	var events []string
	r.CanApplyPreDamage = func(m *ModifierEff) bool { return m != unknown }
	r.ApplyPreDamage = func(m *ModifierEff, _, _, _ *Object, d *int32) {
		switch m {
		case first:
			events = append(events, "first")
			base.Modifiers[1], *d = unknown, -7
		case last:
			if *d != -7 {
				t.Fatal("unknown effect changed raw word")
			}
			events = append(events, "last")
			*d = 1
		default:
			t.Fatal("unknown callback executed")
		}
	}
	r.Unsupported = func(reason string, owner, attacker, w *Object, d int32, typ object.DamageType) {
		if reason != "unsupported live weapon pre-damage effect" || owner != target || attacker != source || w != weapon || d != -7 || typ != object.DamageBlade {
			t.Fatalf("at-use rejection=%q damage=%d", reason, d)
		}
		events = append(events, "reject")
	}
	r.DamageClear = func(*Object, int32) { events = append(events, "HP") }
	if !DefaultDamageWorld4E0B30(target, source, weapon, 8, object.DamageBlade, r) || !slices.Equal(events, []string{"first", "reject", "last", "HP"}) {
		t.Fatalf("continuation=%v", events)
	}
}

func TestDefaultDamageWorld4E0B30PreDamageIntroducedMissingService(t *testing.T) {
	for _, missing := range []string{"availability", "callback", "unsupported effect"} {
		t.Run(missing, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, false), damageMeleeUnitFixture4E17B0(t, true)
			base := &ModifierInitData{}
			weapon := &Object{TypeInd: 777, ObjClass: object.ClassWand, InitData: unsafe.Pointer(base)}
			m := defaultDamagePreEffectFixture4E0B30()
			r := damageMeleeWorldRuntime4E0B30(t)
			var events []string
			r.BuffOff = func(*Object, EnchantID) {
				base.Modifiers[2] = m
				events = append(events, "buff")
			}
			if missing != "availability" {
				r.CanApplyPreDamage = func(got *ModifierEff) bool {
					if got != m || target.Obj130 != weapon || target.Frame134 != 1400 {
						t.Fatal("availability did not observe the at-use prefix")
					}
					return missing != "unsupported effect"
				}
			}
			if missing != "callback" {
				r.ApplyPreDamage = func(*ModifierEff, *Object, *Object, *Object, *int32) {
					t.Fatal("unavailable effect was called")
				}
			}
			r.Unsupported = func(reason string, owner, attacker, w *Object, d int32, typ object.DamageType) {
				if reason != "unsupported live weapon pre-damage effect" || owner != target || attacker != source || w != weapon || d != 5 || typ != object.DamageBlade {
					t.Fatalf("live unavailable callback rejection=%q, damage=%d", reason, d)
				}
				events = append(events, "reject")
			}
			r.DefaultDamageSound = func(*Object, *Object) { events = append(events, "sound") }
			r.DamageClear = func(owner *Object, d int32) {
				if owner != target || d != 5 {
					t.Fatal("unavailable effect changed the damage word")
				}
				events = append(events, "HP")
			}
			if !DefaultDamageWorld4E0B30(target, source, weapon, 5, object.DamageBlade, r) ||
				!slices.Equal(events, []string{"buff", "reject", "sound", "HP"}) || target.Obj130 != weapon || target.Field131 != uint32(object.DamageBlade) ||
				target.Frame134 != 1400 || target.UpdateDataMonster().Field547 != 1 {
				t.Fatalf("unavailable live callback prefix/tail=%v", events)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PreDamageLiveGameBallGuard(t *testing.T) {
	for _, tc := range []struct {
		name     string
		ballType uint16
		drop     bool
		reason   string
	}{
		{"missing type", 0, false, "unsupported live GameBall type"},
		{"missing drop", 77, false, "unsupported live GameBall drop"},
		{"supported", 77, true, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target, source := damageMeleeUnitFixture4E17B0(t, true), damageMeleeUnitFixture4E17B0(t, true)
			weapon := &Object{ObjClass: object.ClassWand, InitData: unsafe.Pointer(&ModifierInitData{})}
			m := defaultDamagePreEffectFixture4E0B30()
			r := damageMeleeWorldRuntime4E0B30(t)
			r.GameBallType = tc.ballType
			r.BuffOff = func(*Object, EnchantID) {
				weapon.InitData = unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{m}})
			}
			r.CanApplyPreDamage = func(*ModifierEff) bool { return true }
			var events []string
			r.ApplyPreDamage = func(got *ModifierEff, _, _, _ *Object, d *int32) {
				if got != m || *d != 3 {
					t.Fatal("live pre-Damage arguments")
				}
				target.Field129, *d = &Object{TypeInd: 77}, 30
				events = append(events, "pre")
			}
			r.DefaultDamageSound = func(*Object, *Object) { events = append(events, "sound") }
			var reason string
			r.Unsupported = func(s string, _, _, _ *Object, d int32, _ object.DamageType) {
				if d != 30 {
					t.Fatal("guard lost live damage")
				}
				reason = s
				events = append(events, "reject")
			}
			if tc.drop {
				r.GameBallOnDamage = func(attacker, owner *Object, d int32) {
					if attacker != source || owner != target || d != 30 {
						t.Fatal("drop arguments")
					}
					events = append(events, "drop")
				}
			}
			r.PlayerSetState = func(owner *Object, state PlayerState) bool {
				if owner != target || state != PlayerState30 || tc.reason != "" {
					t.Fatal("unsupported hurt tail")
				}
				events = append(events, "hurt")
				return true
			}
			r.DamageClear = func(owner *Object, d int32) {
				if owner != target || d != 30 || tc.reason != "" {
					t.Fatal("unsupported HP tail")
				}
				events = append(events, "HP")
			}
			if !DefaultDamageWorld4E0B30(target, source, weapon, 3, object.DamageBlade, r) || reason != tc.reason {
				t.Fatalf("guard=%q want=%q", reason, tc.reason)
			}
			want := []string{"pre", "sound", "reject"}
			if tc.reason == "" {
				want = []string{"pre", "sound", "drop", "hurt", "HP"}
			}
			if !slices.Equal(events, want) || target.Obj130 != weapon || target.Frame134 != 1400 || target.HealthData.Cur != 200 {
				t.Fatalf("tail order=%v want=%v", events, want)
			}
		})
	}
}
