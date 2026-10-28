package server

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/opennox/v1/common/sound"
)

// These 15 required-pointer/early-return outcomes were measured by executing
// unchanged GAME.EXE entries (SHA 0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a).
// Public adapters keep their separate, existing guarded contracts.
func TestLifetimeEffectRawPointerContracts53CC90(t *testing.T) {
	type entry struct {
		kind       LifetimeEffectKind53CC90
		name       string
		nilSource  bool
		nilOwner   bool
		class      object.Class
		flags      object.Flags
		withUpdate bool
		fault      bool
		services   []string
	}
	var cases []entry
	for k := LifetimeEffectSelfDestruct53CC90; k <= LifetimeEffectFlameCleanse53D510; k++ {
		cases = append(cases, entry{kind: k, name: "nil-source", nilSource: true, fault: true})
	}
	cases = append(cases,
		entry{kind: LifetimeEffectMoonglow53D270, name: "nil-owner-safe", nilOwner: true, services: []string{"delete"}},
		entry{kind: LifetimeEffectMoonglow53D270, name: "nonplayer-owner-safe"},
		entry{kind: LifetimeEffectMoonglow53D270, name: "destroyed-nil-update-safe", class: 4, flags: 0x20, services: []string{"delete", "buff"}},
		entry{kind: LifetimeEffectMoonglow53D270, name: "active-nil-update", class: 4, fault: true},
		entry{kind: LifetimeEffectMoonglow53D270, name: "active-nil-player", class: 4, withUpdate: true, fault: true},
		entry{kind: LifetimeEffectTelekinesis53D330, name: "nil-owner", nilOwner: true, fault: true},
		entry{kind: LifetimeEffectTelekinesis53D330, name: "destroyed-nil-update-safe", flags: 0x20, services: []string{"sound", "audio", "delete", "buff"}},
		entry{kind: LifetimeEffectTelekinesis53D330, name: "active-nil-update", fault: true},
		entry{kind: LifetimeEffectTelekinesis53D330, name: "active-nil-player", withUpdate: true, fault: true},
	)
	for _, c := range cases {
		t.Run(fmt.Sprintf("%d/%s", c.kind, c.name), func(t *testing.T) {
			owner := &Object{ObjClass: c.class, ObjFlags: c.flags}
			if c.withUpdate {
				owner.UpdateData = unsafe.Pointer(&PlayerUpdateData{})
			}
			source := &Object{ObjOwner: owner, Field34: 100}
			if c.nilOwner {
				source.ObjOwner = nil
			}
			if c.nilSource {
				source = nil
			}
			var services []string
			h := LifetimeEffectRuntime53CC90{
				Frame: func() uint32 { return 0 }, FPS: func() uint32 { return 30 },
				DelayedDelete: func(o *Object) {
					if o != source {
						t.Fatal("delete object")
					}
					services = append(services, "delete")
				},
				BuffOff: func(o *Object, id EnchantID) {
					want := ENCHANT_MOONGLOW
					if c.kind == LifetimeEffectTelekinesis53D330 {
						want = ENCHANT_TELEKINESIS
					}
					if o != owner || id != want {
						t.Fatal("buff arguments")
					}
					services = append(services, "buff")
				},
				SpellSound: func(id spell.ID, kind int) sound.ID {
					if id != spell.SPELL_TELEKINESIS || kind != 2 {
						t.Fatal("sound arguments")
					}
					services = append(services, "sound")
					return 0x1234abcd
				},
				Audio: func(id sound.ID, o *Object) {
					if id != 0x1234abcd || o != owner {
						t.Fatal("audio arguments")
					}
					services = append(services, "audio")
				},
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				LifetimeEffectUpdateRaw53CC90(c.kind, source, h)
			}()
			if (recovered != nil) != c.fault {
				t.Fatalf("fault %v, want %v", recovered, c.fault)
			}
			if recovered != nil {
				if _, ok := recovered.(runtime.Error); !ok {
					t.Fatalf("unexpected non-memory panic: %v", recovered)
				}
			}
			if !reflect.DeepEqual(services, c.services) {
				t.Fatalf("services %v, want %v", services, c.services)
			}
		})
	}
}
