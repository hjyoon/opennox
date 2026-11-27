package legacy

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestRawDefaultDamageRegisteredCallbacksKeepGoOwned(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	target := &server.Object{InvFirstItem: &server.Object{}}
	source := &server.Object{HealthData: &server.HealthData{Cur: 7}}
	weapon := &server.Object{InvHolder: source}
	objectArg := func(unit *server.Object) rawDefaultDamageNativeArgument4E0B30 {
		return rawDefaultDamageNativeArgument4E0B30{Kind: "object", Pointer: unsafe.Pointer(unit)}
	}

	t.Run("damage", func(t *testing.T) {
		damage := int32(-17)
		const kind = object.DamageType(0x10203040)
		calls := 0
		callback := objectDamageNativeProbePtr()
		server.RegisterObjectDamageGo(fmt.Sprintf("RawGoOwnedDamage%d", objectDamageNativeTestSequence.Add(1)), callback,
			func(a, b, c *server.Object, amount int32, typ object.DamageType) bool {
				calls++
				if a != target || b != source || c != weapon || amount != damage || typ != kind {
					t.Fatalf("damage arguments = %p/%p/%p/%d/%#x", a, b, c, amount, typ)
				}
				return true
			})
		reply := rawDefaultDamageInvokeService4E0B30(rawDefaultDamageNativeService4E0B30{
			Indirect: true, Site: 0x4e0d05, Callback: callback,
			Arguments: []rawDefaultDamageNativeArgument4E0B30{
				objectArg(target), objectArg(source), objectArg(weapon),
				{Kind: "word", Word: uint32(damage)}, {Kind: "word", Word: uint32(kind)},
			},
		})
		if calls != 1 || reply.Word != 0 {
			t.Fatalf("damage callback ran %d times and exposed return word %d", calls, reply.Word)
		}
	})

	runtime := defaultDamageWorldRuntime4E0B30(srv)
	for _, tc := range []struct {
		name     string
		site     uint32
		callback unsafe.Pointer
		bind     *func(*server.Object, *server.Object) int
		source   *server.Object
	}{
		{"default-sound", 0x4e1049, runtime.DefaultDamageSoundC, &Nox_xxx_soundDefaultDamageSound_532E20, source},
		{"player-sound", 0x4e104f, runtime.PlayerDamageSoundC, &Nox_xxx_soundPlayerDamageSound_5328B0, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			old := *tc.bind
			t.Cleanup(func() { *tc.bind = old })
			calls := 0
			*tc.bind = func(a, b *server.Object) int {
				calls++
				if a != target || b != tc.source {
					t.Fatalf("sound arguments = %p/%p", a, b)
				}
				return 0x12345678
			}
			reply := rawDefaultDamageInvokeService4E0B30(rawDefaultDamageNativeService4E0B30{
				Indirect: true, Site: tc.site, Callback: tc.callback,
				Arguments: []rawDefaultDamageNativeArgument4E0B30{objectArg(target), objectArg(tc.source)},
			})
			if calls != 1 || reply.Word != 0 {
				t.Fatalf("sound callback ran %d times and exposed return word %d", calls, reply.Word)
			}
		})
	}
}
