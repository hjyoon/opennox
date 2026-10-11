package legacy

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Real C attack entry -> frames -> actual spatial index -> outer circle
// filter -> target selection -> ray -> registered damage receiver. Four
// animation frames and the receiver are explicit fixture services, not
// Windows wholeworld/HP/death certification.
func TestPlayerAttackCircle518040IndexedCFrameCycles(t *testing.T) {
	b, err := os.ReadFile("../server/testdata/player_attack_circle_518040.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle struct {
		Rows []struct {
			Input struct {
				Name       string
				Radius     uint32
				Center     [2]uint32
				Candidates []struct {
					Position [2]uint32
					Radius   uint32
				}
			} `json:"case"`
			Events []struct{ Kind string }
		}
	}
	if err := json.Unmarshal(b, &oracle); err != nil {
		t.Fatal(err)
	}
	cases := 0
	for _, row := range oracle.Rows {
		if !strings.HasPrefix(row.Input.Name, "circle-retained-edge-") {
			continue
		}
		cases++
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
					t.Fatal("changed LongSword animation", action)
				}
				return 4, 0
			}
			t.Cleanup(func() { GetServer, playerAnimFrames4F9F90 = oldServer, oldFrames })
			if memmap.Float32(0x587000, 194136) != 1 || memmap.Float32(0x587000, 194140) != 0 {
				t.Fatal("unchanged original direction zero required")
			}
			unit, freeUnit := alloc.New(server.Object{})
			update, freeUpdate := alloc.New(server.PlayerUpdateData{})
			player, freePlayer := alloc.New(server.Player{})
			weapon, freeWeapon := alloc.New(server.Object{})
			modifier, freeModifier := alloc.New(server.Modifier{})
			candidate, freeCandidate := alloc.New(server.Object{})
			data, freeData := alloc.New(server.MonsterUpdateData{})
			for _, free := range []func(){freeUnit, freeUpdate, freePlayer, freeWeapon, freeModifier, freeCandidate, freeData} {
				t.Cleanup(free)
			}
			for _, ptr := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(player), unsafe.Pointer(weapon), unsafe.Pointer(modifier), unsafe.Pointer(candidate), unsafe.Pointer(data)} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
					t.Fatal("entire attack graph must be C-owned >4 GiB")
				}
			}
			unit.ObjClass, unit.Direction1 = object.ClassPlayer, 0
			unit.PosVec = types.Ptf(math.Float32frombits(row.Input.Center[0]), math.Float32frombits(row.Input.Center[1]))
			unit.NewPos, unit.Shape.Kind = unit.PosVec, server.ShapeKindCircle
			unit.UpdateData, update.Player, update.EquippedWeapon = unsafe.Pointer(update), player, weapon
			player.WeaponEquip, player.Field8 = 0x200, 23
			player.Info().SetField2239(37)
			weapon.TypeInd, weapon.ObjClass, weapon.ObjSubClass = 0x1234, object.ClassWeapon, object.SubClass(object.WeaponLongSword)
			modifier.TypeInd, modifier.ReqStrength60, modifier.DamageCoeffOrArmor64 = uint32(weapon.TypeInd), 20, 1.5
			modifier.Range68, modifier.DamageMin72 = math.Float32frombits(row.Input.Radius), 10
			s.Modif.Dword_5d4594_251600 = modifier
			if len(row.Input.Candidates) != 1 {
				t.Fatal("original boundary candidate count changed")
			}
			input := row.Input.Candidates[0]
			candidate.TypeInd, candidate.ObjClass, candidate.ObjFlags = math.MaxUint16, object.ClassMonster, object.FlagActive
			candidate.PosVec = types.Ptf(math.Float32frombits(input.Position[0]), math.Float32frombits(input.Position[1]))
			candidate.NewPos, candidate.Shape.Kind, candidate.Shape.Circle.R = candidate.PosVec, server.ShapeKindCircle, math.Float32frombits(input.Radius)
			candidate.Shape.Circle.R2, candidate.UpdateData = candidate.Shape.Circle.R*candidate.Shape.Circle.R, unsafe.Pointer(data)
			hits := 0
			pointer := objectDamageNativeProbePtr()
			server.RegisterObjectDamageGo(fmt.Sprintf("PlayerCirclePrecision%d", objectDamageNativeTestSequence.Add(1)), pointer,
				func(target, source, item *server.Object, damage int32, typ object.DamageType) bool {
					if target != candidate || source != unit || item != weapon || damage != 36 || typ != object.DamageBlade {
						t.Fatalf("changed damage/attribution/type: %p/%p/%p/%d/%d", target, source, item, damage, typ)
					}
					hits++
					return true
				})
			candidate.Damage = pointer
			s.Map.AddObjectToIndex(candidate)
			if !s.IsEnemyTo(unit, candidate) || !s.CanInteract(unit, candidate, 1) {
				t.Fatal("original edge target is not eligible")
			}
			want := 0
			for _, event := range row.Events {
				if event.Kind == "visit" {
					want++
				}
			}
			logic, other := s.Rand.Logic.Index(), s.Rand.Other.Index()
			for cycle := uint32(0); cycle < 3; cycle++ {
				unit.Field34, update.Field59_0, update.Field0 = 100+cycle*100, 0, 0
				before := hits
				for tick := uint32(0); tick <= 4; tick++ {
					s.SetFrame(unit.Field34 + tick)
					for repeat := 0; repeat < 2; repeat++ {
						result := playerAttackNativeEntry538960(unit)
						active := 1
						if tick == 4 {
							active = 0
						}
						if result != active || update.Field59_0 != uint8(min(tick, 3)) || update.Field0 != unit.Field34+4 {
							t.Fatalf("cycle=%d tick=%d repeat=%d result/frame/deadline=%d/%d/%d", cycle, tick, repeat, result, update.Field59_0, update.Field0)
						}
					}
				}
				if hits-before != want {
					t.Errorf("C indexed attack cycle=%d damage calls=%d original circle visits=%d radius=%08x", cycle, hits-before, want, row.Input.Radius)
				}
			}
			if s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other {
				t.Fatal("outer circle filter changed gameplay RNG indexes")
			}
			t.Logf("C-owned >4 GiB indexed C attack; three cycles original visits=%d damage calls=%d damage=36 BLADE; RNG unchanged", want, hits)
		})
	}
	if cases != 9 {
		t.Fatal("lost original nine circle boundary cases", cases)
	}
}
