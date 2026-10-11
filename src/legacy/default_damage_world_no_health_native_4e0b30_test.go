package legacy

import (
	"bytes"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Both reported metadata sets traverse the real C dispatcher and registered
// DefaultDamage/UnitDamageClear callbacks. No HP/damage callback is replaced.
func TestDefaultDamageWorldNoHealthNative4E0B30(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	const id = "NativeWorldNoHealthDefault"
	if err := srv.Types.ReadObjectType(&things.Thing{Name: id, OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !srv.Objs.Init(96) {
		t.Fatal("native world no-health allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	oldLog := srv.Log
	t.Cleanup(func() { srv.Log = oldLog })
	for _, meta := range []struct {
		name  string
		typ   uint16
		class object.Class
		flags object.Flags
	}{{"PlayerStart", 714, 0, 16781893}, {"InvisibleTeleportPentagram", 717, 4195328, 150995469}} {
		for _, health := range []string{"nil", "zero-max", "ordinary"} {
			for _, weapon := range []bool{false, true} {
				for _, damage := range []int32{1, 50} {
					t.Run(fmt.Sprintf("%s/%s/weapon-%t/raw-%d", meta.name, health, weapon, damage), func(t *testing.T) {
						var diagnostics bytes.Buffer
						srv.Log = slog.New(slog.NewTextHandler(&diagnostics, nil))
						v := srv.Objs.NewObject(&server.ObjectType{})
						a := srv.Objs.NewObject(&server.ObjectType{})
						h, freeH := alloc.New(server.HealthData{})
						defer freeH()
						*h = server.HealthData{Cur: 120, Max: 120, Field2: 119}
						v.TypeInd, v.ObjClass, v.ObjSubClass, v.ObjFlags = meta.typ, meta.class, 0, meta.flags
						v.Damage, v.Material = srv.Types.ByID(id).Damage, 0x4000
						v.Frame134, v.Field131, v.Pos132 = 77, 19, types.Ptf(31, 47)
						if health != "nil" {
							v.HealthData = h
						}
						if health == "zero-max" {
							h.Max = 0
						}
						a.TypeInd, a.ObjClass, a.ObjSubClass, a.ObjFlags = 706, 1572865, 1, 553665028
						a.PrevPos = types.Ptf(17, 19)
						w := (*server.Object)(nil)
						if weapon {
							w = a
						}
						for _, p := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(a), unsafe.Pointer(h)} {
							if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
								t.Fatal("C records are not above 4 GiB")
							}
						}
						beforeV, beforeA, beforeH := *v, *a, *h
						beforeU := v.UpdateData
						hp := []uint16{h.Cur}
						for hit := 1; hit <= 2; hit++ {
							result := objectDamageDispatchCallNative(v, a, w, damage, object.DamageCrush)
							want := beforeV
							want.Obj130, want.Field131, want.Frame134, want.Pos132 = a, 2, 1400, a.PrevPos
							wantHP := beforeH.Cur
							if health == "ordinary" {
								wantHP -= uint16(hit) * uint16(damage)
								want.Field38 = math.MaxUint32
							}
							hp = append(hp, h.Cur)
							if strings.Contains(diagnostics.String(), "native branch is not ported") {
								t.Fatalf("reported world no-health shape rejected: %s", diagnostics.String())
							}
							if !result || *v != want || *a != beforeA || v.UpdateData != beforeU ||
								h.Cur != wantHP || h.Max != beforeH.Max || h.Field2 != beforeH.Field2 {
								t.Fatalf("C world no-health hit=%d result=%t HP=%d/%d source=%p/%p frame=%d", hit, result, h.Cur, wantHP, v.Obj130, a, v.Frame134)
							}
						}
						t.Logf("WORLD NO-HEALTH C-owned >4 GiB C->registered DefaultDamage->UnitDamageClear: target=%s type=%d class=%d flags=%d source=DeathBall type=706 class=1572865 flags=553665028 weapon=%p raw=%d/type2 health=%s HP=%v no-unsupported-warning", meta.name, meta.typ, uint32(meta.class), uint32(meta.flags), w, damage, health, hp)
					})
				}
			}
		}
	}
}
