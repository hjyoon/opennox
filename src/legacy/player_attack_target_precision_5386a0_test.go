package legacy

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Real C entry -> frame advance -> indexed scan -> candidate selection ->
// directional ray/damage callback. The four animation frames and supplied
// damage receiver are explicit fixture services, not original world combat.
func TestPlayerAttackTarget5386A0IndexedCFrameCycles(t *testing.T) {
	b, err := os.ReadFile("../server/testdata/player_attack_target_5386a0.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle struct {
		Rows []struct {
			Input struct {
				Name       string
				Range      uint32
				Direction  uint16
				Position   [2]uint32
				Candidates []struct {
					Name     string
					Class    uint32 `json:"cls"`
					Position [2]uint32
					Radius   uint32
				}
			} `json:"case"`
			Selected string
		}
	}
	if err := json.Unmarshal(b, &oracle); err != nil {
		t.Fatal(err)
	}
	for _, row := range oracle.Rows {
		switch row.Input.Name {
		case "cone--1", "cone-+0", "closer-object-after-retained-edge-store", "sticky-first-monster":
		default:
			continue
		}
		t.Run(row.Input.Name, func(t *testing.T) {
			s := server.New(nil, nil, strman.New())
			t.Cleanup(s.Close)
			s.Map.Init()
			t.Cleanup(s.Map.Free)
			if s.Walls.Init() == 0 {
				t.Fatal("cannot initialize native wall grid")
			}
			t.Cleanup(s.Walls.Free)
			bridge := &playerAttackLegacyServer538960{srv: s}
			oldServer, oldFrames := GetServer, playerAnimFrames4F9F90
			GetServer = func() Server { return bridge }
			playerAnimFrames4F9F90 = func(action int) (int, int) {
				if action != 28 {
					t.Fatalf("animation=%d want original LongSword action 28", action)
				}
				return 4, 0
			}
			t.Cleanup(func() { GetServer, playerAnimFrames4F9F90 = oldServer, oldFrames })
			if memmap.Float32(0x587000, 194136) != 1 || memmap.Float32(0x587000, 194140) != 0 {
				t.Fatal("original direction zero was changed")
			}
			unit, freeUnit := alloc.New(server.Object{})
			t.Cleanup(freeUnit)
			update, freeUpdate := alloc.New(server.PlayerUpdateData{})
			t.Cleanup(freeUpdate)
			player, freePlayer := alloc.New(server.Player{})
			t.Cleanup(freePlayer)
			weapon, freeWeapon := alloc.New(server.Object{})
			t.Cleanup(freeWeapon)
			modifier, freeModifier := alloc.New(server.Modifier{})
			t.Cleanup(freeModifier)
			unit.ObjClass, unit.Direction1 = object.ClassPlayer, server.Dir16(row.Input.Direction)
			unit.PosVec = types.Ptf(math.Float32frombits(row.Input.Position[0]), math.Float32frombits(row.Input.Position[1]))
			unit.NewPos, unit.Shape.Kind = unit.PosVec, server.ShapeKindCircle
			unit.UpdateData, update.Player, update.EquippedWeapon = unsafe.Pointer(update), player, weapon
			player.WeaponEquip, player.Field8 = 0x200, 23
			player.Info().SetField2239(37)
			weapon.TypeInd, weapon.ObjClass, weapon.ObjSubClass = 0x1234, object.ClassWeapon, object.SubClass(object.WeaponLongSword)
			modifier.TypeInd, modifier.ReqStrength60, modifier.DamageCoeffOrArmor64 = uint32(weapon.TypeInd), 20, 1.5
			modifier.Range68, modifier.DamageMin72 = math.Float32frombits(row.Input.Range), 10
			s.Modif.Dword_5d4594_251600 = modifier
			for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(player), unsafe.Pointer(weapon), unsafe.Pointer(modifier)} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
					t.Fatal("attack graph must be C-owned >4 GiB")
				}
			}
			names := make(map[*server.Object]string)
			var candidates []*server.Object
			for _, input := range row.Input.Candidates {
				candidate, free := alloc.New(server.Object{})
				t.Cleanup(free)
				candidate.TypeInd, candidate.ObjClass, candidate.ObjFlags = math.MaxUint16, object.Class(input.Class), object.FlagActive
				candidate.PosVec = types.Ptf(math.Float32frombits(input.Position[0]), math.Float32frombits(input.Position[1]))
				candidate.NewPos, candidate.Shape.Kind, candidate.Shape.Circle.R = candidate.PosVec, server.ShapeKindCircle, math.Float32frombits(input.Radius)
				candidate.Shape.Circle.R2 = candidate.Shape.Circle.R * candidate.Shape.Circle.R
				ud, freeUD := alloc.New(server.MonsterUpdateData{})
				t.Cleanup(freeUD)
				candidate.UpdateData = unsafe.Pointer(ud)
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(candidate)) <= math.MaxUint32 {
					t.Fatal("target must retain its C-owned >4 GiB pointer")
				}
				names[candidate], candidates = input.Name, append(candidates, candidate)
			}
			var hits []string
			pointer := objectDamageNativeProbePtr()
			server.RegisterObjectDamageGo(fmt.Sprintf("PlayerTargetPrecision%d", objectDamageNativeTestSequence.Add(1)), pointer,
				func(target, source, item *server.Object, damage int32, typ object.DamageType) bool {
					if source != unit || item != weapon || damage != 36 || typ != object.DamageBlade {
						t.Fatalf("changed damage attribution/value/type: %p/%p/%d/%d", source, item, damage, typ)
					}
					hits = append(hits, names[target])
					return true
				})
			// Buckets prepend: preserve the original declared candidate order.
			for i := len(candidates) - 1; i >= 0; i-- {
				candidate := candidates[i]
				candidate.Damage = pointer
				s.Map.AddObjectToIndex(candidate)
				if !s.IsEnemyTo(unit, candidate) || !s.CanInteract(unit, candidate, 1) {
					t.Fatal("indexed fixture candidate is not eligible")
				}
			}
			logic, other := s.Rand.Logic.Index(), s.Rand.Other.Index()
			for cycle := uint32(0); cycle < 3; cycle++ {
				unit.Field34, update.Field59_0, update.Field0 = 100+cycle*100, 0, 0
				before := len(hits)
				for tick := uint32(0); tick <= 4; tick++ {
					s.SetFrame(unit.Field34 + tick)
					for repeat := 0; repeat < 2; repeat++ {
						result := playerAttackNativeEntry538960(unit)
						wantResult := 1
						if tick == 4 {
							wantResult = 0
						}
						if result != wantResult || update.Field59_0 != uint8(min(tick, 3)) || update.Field0 != unit.Field34+4 {
							t.Fatalf("cycle=%d tick=%d repeat=%d result/frame/deadline=%d/%d/%d", cycle, tick, repeat, result, update.Field59_0, update.Field0)
						}
					}
				}
				wantHits := 1
				if row.Selected == "" {
					wantHits = 0
				}
				if len(hits)-before != wantHits || wantHits != 0 && hits[before] != row.Selected {
					t.Fatalf("C indexed attack cycle=%d hits=%v original target=%q", cycle, hits[before:], row.Selected)
				}
			}
			if s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other {
				t.Fatal("candidate selection changed gameplay RNG indexes")
			}
			t.Logf("C-owned >4 GiB; three unchanged frame cycles, original target=%q hits=%v", row.Selected, hits)
		})
	}
}
