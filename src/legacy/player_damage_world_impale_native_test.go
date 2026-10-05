package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Actual C collision entry -> native DamageCollide -> registered PlayerDamage ->
// DefaultDamage -> UnitSetHP. No damage/HP callback is replaced. These scalar
// records model the six stock declarations; headless tests load thing.bin itself.
func TestPlayerDamageSpikesCollisionCRecords4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	if !srv.Objs.Init(64) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	callback, size, ok := server.ObjectCollideHandler("DamageCollide")
	if !ok || callback == nil || size != 8 {
		t.Fatal("stock DamageCollide registration")
	}
	for _, hazard := range []struct {
		name   string
		class  object.Class
		damage uint8
		hp     uint16
	}{
		{"Spike", object.ClassDangerous | object.ClassImmobile | object.ClassVisibleEnable, 2, 59},
		{"PeriodicSpike", object.ClassDangerous | object.ClassImmobile | object.ClassVisibleEnable, 2, 59},
		{"SpikeBlock", object.ClassObstacle | object.ClassDangerous | object.ClassSimple, 3, 59},
		{"SpikeBlockImmobile", object.ClassObstacle | object.ClassDangerous | object.ClassImmobile, 3, 59},
		{"RotatingSpikes", object.ClassObstacle | object.ClassDangerous | object.ClassSimple | object.ClassVisibleEnable, 8, 57},
		{"RotatingSpikesImmobile", object.ClassObstacle | object.ClassDangerous | object.ClassImmobile | object.ClassVisibleEnable, 8, 57},
	} {
		for _, owner := range []string{"self", "player", "NPC"} {
			t.Run(fmt.Sprintf("%s/%s", hazard.name, owner), func(t *testing.T) {
				v, a, w := spellMissileImpactCRecords4E17B0(t, srv, owner)
				data, freeData := alloc.New(server.DamageCollideData{})
				t.Cleanup(freeData)
				*data = server.DamageCollideData{Damage: hazard.damage, Reserved: [3]uint8{4, 5, 6}, DamageType: 3}
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(data)) <= math.MaxUint32 {
					t.Fatal("collide data not above 4 GiB")
				}
				w.ObjClass, w.ObjSubClass, w.CollideData = hazard.class, 0, unsafe.Pointer(data)
				v.Damage = playerDamageMeleeCallbackNative4E17B0()
				ud := v.UpdateDataPlayer()
				ud.Field57 = math.Float32bits(.25)
				before := *w
				ccall.CallVoidPtr3(callback, unsafe.Pointer(w), unsafe.Pointer(v), nil)
				if v.HealthData.Cur != hazard.hp || v.Obj130 != w || v.Field131 != 3 || v.Pos132 != w.PrevPos ||
					v.Frame134 != 1400 || *w != before || data.Damage != hazard.damage || data.Reserved != [3]uint8{4, 5, 6} {
					t.Fatalf("actual C collision: player=%p parent=%p hazard=%p HP=%d want=%d raw-byte=%d", v, a, w, v.HealthData.Cur, hazard.hp, hazard.damage)
				}
				if (a == w && (ud.Field76 != 2 || ud.Field75 != 3)) ||
					(a != w && (ud.Field76 != 1 || ud.Field75 != uint32(w.TypeInd))) {
					t.Fatal("collision attribution marker")
				}
				t.Logf("C-owned >4 GiB %s source=%s actual C collision HP=60->%d byte=%d IMPALE=3", hazard.name, owner, v.HealthData.Cur, data.Damage)
			})
		}
	}
}
