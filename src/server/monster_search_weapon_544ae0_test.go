package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/types"
)

func TestMonsterSearchWeapon544AE0OriginalSelection(t *testing.T) {
	for _, mode := range []string{"nearest", "tie", "binary32-best-spill", "NaN-last", "NaN-middle", "distance-limit", "infinite"} {
		t.Run(mode, func(t *testing.T) {
			unit := &Object{}
			first := &Object{ObjClass: object.ClassWeapon, PosVec: types.Ptf(30, 40)}
			second := &Object{ObjClass: object.ClassWeapon, PosVec: types.Ptf(10, 20)}
			items, want := []*Object{first, second}, second
			switch mode {
			case "tie":
				second.PosVec = types.Ptf(40, 30)
				want = first
			case "binary32-best-spill":
				first.PosVec = types.Ptf(1, 0.0003)
				second.PosVec = types.Ptf(1, 0.00034)
			case "NaN-last":
				second.PosVec.X = float32(math.NaN())
			case "NaN-middle":
				first.PosVec.X = float32(math.NaN())
			case "distance-limit":
				first.PosVec = types.Ptf(1000, 3000)
				second.PosVec = types.Ptf(3000, 1000)
				want = nil
			case "infinite":
				first.PosVec.X = float32(math.Inf(1))
				second.PosVec.X = float32(math.Inf(-1))
				want = nil
			}
			calls := []string{}
			got := monsterSearchWeapon544AE0(unit, 75, monsterSearchWeaponHooks544AE0{
				eachInCircle: func(pos types.Pointf, radius float32, each func(*Object) bool) {
					if pos != unit.PosVec || radius != 75 {
						t.Fatal("enumeration arguments")
					}
					for _, item := range items {
						if !each(item) {
							t.Fatal("search stopped early")
						}
					}
				},
				classCanUse: func(item *Object, class player.Class) bool {
					if class != player.Warrior {
						t.Fatal("non-Warrior eligibility query")
					}
					calls = append(calls, "class")
					return true
				},
				canInteract: func(owner, item *Object, flags int) bool {
					if owner != unit || flags != 0 {
						t.Fatal("interaction arguments")
					}
					calls = append(calls, "interact")
					return true
				},
			})
			if got != want || !reflect.DeepEqual(calls, []string{"class", "interact", "class", "interact"}) {
				t.Fatalf("winner=%p, want %p; calls=%v", got, want, calls)
			}
		})
	}
}

func TestMonsterSearchWeapon544AE0FiltersAndLivePositions(t *testing.T) {
	unit := &Object{PosVec: types.Ptf(100, 200)}
	wrongClass := &Object{ObjClass: object.ClassWand}
	wrongWarrior := &Object{ObjClass: object.ClassWeapon}
	blocked := &Object{ObjClass: object.ClassWeapon}
	winner := &Object{ObjClass: object.ClassWeapon, PosVec: types.Ptf(10000, 10000)}
	classes, interactions := []*Object{}, []*Object{}
	got := monsterSearchWeapon544AE0(unit, 75, monsterSearchWeaponHooks544AE0{
		eachInCircle: func(_ types.Pointf, _ float32, each func(*Object) bool) {
			for _, item := range []*Object{nil, wrongClass, wrongWarrior, blocked, winner} {
				each(item)
			}
		},
		classCanUse: func(item *Object, _ player.Class) bool { classes = append(classes, item); return item != wrongWarrior },
		canInteract: func(owner, item *Object, flags int) bool {
			interactions = append(interactions, item)
			if item == blocked {
				return false
			}
			// Class is not rechecked, and distance is read after this callback.
			winner.ObjClass = 0
			winner.PosVec, unit.PosVec = types.Ptf(301, 400), types.Ptf(300, 400)
			return true
		},
	})
	if got != winner || !reflect.DeepEqual(classes, []*Object{wrongWarrior, blocked, winner}) || !reflect.DeepEqual(interactions, []*Object{blocked, winner}) {
		t.Fatal("filter order or live distance lost")
	}
}

func TestMonsterMainWeaponTail547210OriginalGates(t *testing.T) {
	for _, mode := range []string{"unarmed-warrior", "nonbot", "armed-warrior", "conjurer", "wizard", "not-due", "no-weapon", "missing-player-record", "missing-player", "missing-search", "missing-inventory"} {
		t.Run(mode, func(t *testing.T) {
			s, unit, update := monsterMainHealthRetreatFixture547210(t)
			s.SetFrame(160)
			owner := &Player{}
			owner.Info().SetPlayerClass(player.Warrior)
			pud := &PlayerUpdateData{Player: owner, Field73: update}
			update.StatusFlags, update.Field545 = object.MonStatusBot, pud
			weapon := &Object{ObjClass: object.ClassWeapon}
			searches, places, wantSearch, wantPlace, wantHandled := 0, 0, 1, 1, true
			runtime := MonsterMainRuntime547210{
				SearchWeapon: func(got *Object, radius float32) *Object {
					if got != unit || radius != 75 {
						t.Fatal("weapon search arguments")
					}
					searches++
					if mode == "no-weapon" {
						return nil
					}
					return weapon
				},
				PlaceInventory: func(got, item *Object, a, b int) bool {
					if got != unit || item != weapon || a != 1 || b != 1 || unit.UpdateData != unsafe.Pointer(pud) || !unit.ObjClass.Has(object.ClassPlayer) || unit.ObjClass.Has(object.ClassMonster) || unit.ObjSubClass != 0 {
						t.Fatal("inventory must run between normal native morphs")
					}
					places++
					return false
				},
			}
			switch mode {
			case "nonbot":
				update.StatusFlags = 0
				wantSearch, wantPlace = 0, 0
			case "armed-warrior":
				owner.WeaponEquip = 1
				wantSearch, wantPlace = 0, 0
			case "conjurer":
				owner.Info().SetPlayerClass(player.Conjurer)
				wantSearch, wantPlace = 0, 0
			case "wizard":
				owner.Info().SetPlayerClass(player.Wizard)
				wantSearch, wantPlace = 0, 0
			case "not-due":
				s.SetFrame(161)
				wantSearch, wantPlace = 0, 0
			case "no-weapon":
				wantPlace = 0
			case "missing-player-record":
				update.Field545 = nil
				wantHandled, wantSearch, wantPlace = false, 0, 0
			case "missing-player":
				pud.Player = nil
				wantHandled, wantSearch, wantPlace = false, 0, 0
			case "missing-search":
				runtime.SearchWeapon = nil
				wantHandled, wantSearch, wantPlace = false, 0, 0
			case "missing-inventory":
				runtime.PlaceInventory = nil
				wantHandled, wantSearch, wantPlace = false, 0, 0
			}
			if s.monsterMainPickupWeapon547210(unit, update, runtime) != wantHandled || searches != wantSearch || places != wantPlace || unit.UpdateData != unsafe.Pointer(update) || !unit.ObjClass.Has(object.ClassMonster) {
				t.Fatalf("weapon tail: searches=%d places=%d class=%v", searches, places, unit.ObjClass)
			}
		})
	}
}

func TestMonsterMainWeaponTail547210CachedAdmissionLiveMorphLinks(t *testing.T) {
	s, unit, cached := monsterMainHealthRetreatFixture547210(t)
	s.SetFrame(160)
	owner := &Player{}
	cached.Field545 = &PlayerUpdateData{Player: owner}
	cached.StatusFlags = object.MonStatusBot
	live, reverse := &MonsterUpdateData{}, &MonsterUpdateData{}
	pud := &PlayerUpdateData{Field73: live}
	live.Field545 = pud
	unit.UpdateData = unsafe.Pointer(live)
	weapon := &Object{ObjClass: object.ClassWeapon}
	events := []string{}
	if !s.monsterMainPickupWeapon547210(unit, cached, MonsterMainRuntime547210{
		SearchWeapon: func(*Object, float32) *Object {
			events = append(events, "search")
			cached.StatusFlags = 0
			owner.WeaponEquip = 1
			s.SetFrame(161)
			return weapon
		},
		PlaceInventory: func(owner, item *Object, a, b int) bool {
			if unit.UpdateData != unsafe.Pointer(pud) || !unit.ObjClass.Has(object.ClassPlayer) {
				t.Fatal("morph used cached link instead of live unit")
			}
			events = append(events, "place")
			pud.Field73 = reverse
			return false
		},
	}) || unit.UpdateData != unsafe.Pointer(reverse) || unit.ObjSubClass != 16 || !unit.ObjClass.Has(object.ClassMonster) || !reflect.DeepEqual(events, []string{"search", "place"}) {
		t.Fatal("original cached admission/live reverse link lost")
	}
}
