package server

import (
	"reflect"
	"testing"

	"github.com/opennox/libs/object"
)

func TestEffectRegeneration4E01D0TimingAndGates(t *testing.T) {
	for _, tc := range []struct {
		name                                   string
		frame, last, fps                       uint32
		rate                                   int32
		cur, max                               uint16
		armor, special, dead, eliminated, heal bool
	}{
		{name: "cooldown", frame: 29, fps: 30, rate: 60, cur: 149, max: 150},
		{name: "non-interval", frame: 30, fps: 30, rate: 60, cur: 149, max: 150},
		{name: "heal", frame: 36, fps: 30, rate: 60, cur: 149, max: 150, heal: true},
		{name: "full", frame: 36, fps: 30, rate: 60, cur: 150, max: 150},
		{name: "overfull", frame: 36, fps: 30, rate: 60, cur: 151, max: 150},
		{name: "dead", frame: 36, fps: 30, rate: 60, cur: 149, max: 150, dead: true},
		{name: "eliminated", frame: 36, fps: 30, rate: 60, cur: 149, max: 150, eliminated: true},
		{name: "armor rate divided by three", frame: 32, fps: 30, rate: 60, cur: 149, max: 150, armor: true, special: true, heal: true},
		{name: "ordinary armor", frame: 32, fps: 30, rate: 60, cur: 149, max: 150, armor: true},
		{name: "non-armor never checks flags", frame: 32, fps: 30, rate: 60, cur: 149, max: 150, special: true},
		{name: "DWORD cooldown wrap", frame: 0, last: 0xffffffe2, fps: 30, rate: 60, cur: 149, max: 150, heal: true},
		{name: "DWORD product wrap", frame: 16, fps: 16, rate: 0x10000001, cur: 15, max: 16, heal: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			owner := &Object{HealthData: &HealthData{Cur: tc.cur, Max: tc.max}, Frame134: tc.last}
			if tc.dead {
				owner.ObjFlags |= object.FlagDead
			}
			if tc.eliminated {
				owner.ObjFlags |= 0x8000
			}
			item := &Object{InvHolder: owner, ObjClass: object.ClassWeapon}
			if tc.armor {
				item.ObjClass = object.ClassArmor
			}
			effect := &ModifierEff{Update100: ModifierEffFnc{Val: tc.rate}}
			var healed int
			r := regenerationRuntime4E01D0{
				frame: func() uint32 { return tc.frame }, fps: func() uint32 { return tc.fps },
				maxHP: UnitGetMaxHP4EE7A0, hp: UnitGetHP4EE780,
				armorFlags: func(got *Object) uint32 {
					if got != item || !tc.armor {
						t.Fatal("unexpected armor lookup")
					}
					if tc.special {
						return 0x4000
					}
					return 0
				},
				adjustHP: func(got *Object, delta int32) {
					if got != owner || delta != 1 {
						t.Fatal("wrong healing target/delta")
					}
					healed++
				},
			}
			effectRegeneration4E01D0(effect, item, r)
			if (healed == 1) != tc.heal || healed > 1 {
				t.Fatalf("heals=%d, want heal=%t", healed, tc.heal)
			}
		})
	}
}

func TestEffectRegeneration4E01D0CachedOwnerFreshRateInputs(t *testing.T) {
	owner := &Object{HealthData: &HealthData{Cur: 149, Max: 150}}
	item := &Object{InvHolder: owner, ObjClass: object.ClassArmor}
	effect := &ModifierEff{Update100: ModifierEffFnc{Val: 90}}
	var events []string
	frames, maxes, fpses := 0, 0, 0
	r := regenerationRuntime4E01D0{
		frame: func() uint32 {
			frames++
			events = append(events, "frame")
			if frames == 1 {
				return 40
			}
			return 42
		},
		fps: func() uint32 {
			fpses++
			events = append(events, "fps")
			if fpses == 1 {
				return 30
			}
			return 42
		},
		maxHP: func(got *Object) uint16 {
			if got != owner {
				t.Fatal("owner reloaded")
			}
			maxes++
			events = append(events, "max")
			if maxes == 1 {
				item.InvHolder = &Object{}
				return 150
			}
			return 180
		},
		hp: func(got *Object) uint16 {
			if got != owner {
				t.Fatal("HP owner reloaded")
			}
			events = append(events, "hp")
			return 149
		},
		armorFlags: func(got *Object) uint32 {
			if got != item {
				t.Fatal("armor item")
			}
			events = append(events, "armor")
			effect.Update100.Val = 900
			return 0x4000
		},
		adjustHP: func(got *Object, delta int32) {
			if got != owner || delta != 1 {
				t.Fatal("adjust args")
			}
			events = append(events, "adjust")
		},
	}
	// Cached rate 90 / 3, fresh FPS 42 / fresh max 180 => interval 7.
	effectRegeneration4E01D0(effect, item, r)
	if want := []string{"frame", "fps", "max", "hp", "armor", "max", "fps", "frame", "adjust"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events=%v, want %v", events, want)
	}
}

func TestEffectRegeneration4E01D0NullGatesAndArithmeticFaults(t *testing.T) {
	for _, item := range []*Object{nil, {}, {InvHolder: &Object{}}} {
		effectRegeneration4E01D0(nil, item, regenerationRuntime4E01D0{})
	}
	for _, tc := range []struct {
		name      string
		rate      int32
		secondMax uint16
	}{
		{"zero rate", 0, 150}, {"zero interval", 1, 150}, {"fresh zero maximum", 60, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("original arithmetic fault silently clamped")
				}
			}()
			owner := &Object{HealthData: &HealthData{Cur: 149, Max: 150}}
			maxCalls := 0
			effectRegeneration4E01D0(&ModifierEff{Update100: ModifierEffFnc{Val: tc.rate}}, &Object{InvHolder: owner}, regenerationRuntime4E01D0{
				frame: func() uint32 { return 36 }, fps: func() uint32 { return 30 }, hp: UnitGetHP4EE780,
				maxHP: func(*Object) uint16 {
					maxCalls++
					if maxCalls == 1 {
						return 150
					}
					return tc.secondMax
				},
			})
		})
	}
}
