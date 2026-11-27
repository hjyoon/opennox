package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestDefaultDamageWorld4E0B30UsesLiveRegisteredDamageSound(t *testing.T) {
	defaultSound, playerSound := unsafe.Pointer(new(byte)), unsafe.Pointer(new(byte))
	for _, tc := range []struct {
		name          string
		before, after unsafe.Pointer
		want          string
	}{
		{"player-sound-on-armor", playerSound, playerSound, "player sound"},
		{"protection-selects-player-sound", defaultSound, playerSound, "player sound"},
		{"protection-selects-default-sound", playerSound, defaultSound, "default sound"},
		{"protection-clears-callback", playerSound, nil, "default sound"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := &Object{ObjClass: object.ClassArmor, DamageSound: tc.before}
			var events []string
			sound := func(name string) func(*Object, *Object) {
				return func(gotTarget, gotSource *Object) {
					if gotTarget != target || gotSource != nil {
						t.Fatalf("sound objects = %p/%p, want %p/nil", gotTarget, gotSource, target)
					}
					events = append(events, name)
				}
			}
			runtime := DefaultDamageWorldRuntime4E0B30{
				FireProtection: func(got *Object) float64 {
					if got != target {
						t.Fatalf("protection owner = %p, want %p", got, target)
					}
					got.DamageSound = tc.after
					events = append(events, "fire protection")
					return 0
				},
				BuffOff:             func(*Object, EnchantID) {},
				DefaultDamageSound:  sound("default sound"),
				PlayerDamageSound:   sound("player sound"),
				DefaultDamageSoundC: defaultSound,
				PlayerDamageSoundC:  playerSound,
				DamageClear: func(got *Object, damage int32) {
					if got != target || damage != 6 {
						t.Fatalf("damage = %p/%d, want %p/6", got, damage, target)
					}
					events = append(events, "damage")
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("registered damage sound rejected: %s", reason)
				},
			}
			if !DefaultDamageWorld4E0B30(target, nil, nil, 6, object.DamageLava, runtime) {
				t.Fatal("source-less lava returned false")
			}
			want := []string{"fire protection", tc.want, "damage"}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}
