package server

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type monsterDeathUpdateView544EC0 struct {
	Waypoints      uint32               `json:"waypoints"`
	WaypointIDs    [16]uint32           `json:"waypoint_ids"`
	Seen           uint8                `json:"seen"`
	EnemyIDs       [16]uint32           `json:"enemy_ids"`
	CurrentEnemy   uint32               `json:"current_enemy"`
	Field91        uint32               `json:"field91"`
	Electric       uint8                `json:"electric"`
	ElectricCanary uint8                `json:"electric_canary"`
	StaminaCanary  [3]uint8             `json:"stamina_canaries"`
	Status         object.MonsterStatus `json:"status"`
	Definition     string               `json:"definition"`
}

type monsterDeathUpdateSnapshot544EC0 struct {
	At          string                       `json:"at"`
	UpdateNil   bool                         `json:"update_nil"`
	Entry       monsterDeathUpdateView544EC0 `json:"entry"`
	Replacement monsterDeathUpdateView544EC0 `json:"replacement"`
}

// GAME.EXE executes all three sealed routines. Only the explicitly named
// external services are replaced; their ordered mutations expose cached and
// live loads, counted pointer stores and the original fault boundaries.
func TestMonsterDeathUpdateOriginal544EC0(t *testing.T) {
	raw, err := os.ReadFile("testdata/monster_death_update_544ec0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"code_bytes_unchanged"`
		Rows      []struct {
			Input struct {
				Name    string               `json:"name"`
				Entry   string               `json:"entry"`
				Zombie  bool                 `json:"zombie"`
				Status  object.MonsterStatus `json:"status"`
				Actions []struct {
					At   string `json:"at"`
					Kind string `json:"kind"`
				} `json:"actions"`
				Waypoints  uint32               `json:"waypoints"`
				Seen       uint8                `json:"seen"`
				Delete     object.MonsterStatus `json:"delete"`
				Definition bool                 `json:"definition"`
				Frame      uint32               `json:"frame"`
				DeadFrame  uint32               `json:"dead_frame"`
				Duration   uint32               `json:"duration"`
				Enemy      bool                 `json:"enemy"`
			} `json:"input"`
			Events      []string                           `json:"events"`
			Entry       monsterDeathUpdateView544EC0       `json:"entry"`
			Replacement monsterDeathUpdateView544EC0       `json:"replacement"`
			Update      string                             `json:"current_update"`
			UpdateNil   bool                               `json:"update_nil"`
			Flags       object.Flags                       `json:"flags"`
			Snapshots   []monsterDeathUpdateSnapshot544EC0 `json:"snapshots"`
			Faults      []struct {
				Address uint32 `json:"address"`
			} `json:"faults"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || len(fixture.Rows) != 58 {
		t.Fatal("unchanged original death update fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			unit := monsterActionTestObject50A910(t)
			entry, replacement := unit.UpdateDataMonster(), new(MonsterUpdateData)
			a, b := &MonsterDef{StatusFlags92: row.Input.Delete}, &MonsterDef{StatusFlags92: 1}
			entry.MonsterDef, replacement.MonsterDef = a, b
			unit.ObjFlags = 0x12008014
			unit.PosVec = types.Ptf(321, 654)
			unit.Update = unsafe.Pointer(unit)
			waypointIDs, enemyIDs := map[*Waypoint]uint32{}, map[*Object]uint32{}
			current := &Object{}
			enemyIDs[current] = 0x44000001
			for n, ud := range []*MonsterUpdateData{entry, replacement} {
				if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(ud)) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(ud.MonsterDef)) <= uintptr(^uint32(0))) {
					t.Fatal("test update/definition must use native pointers above 4 GiB")
				}
				ud.Field74, ud.Field282_1 = row.Input.Waypoints, row.Input.Seen
				if n == 1 {
					ud.Field74, ud.Field282_1 = 1, 2
				}
				ud.Field91 = 0x12345678 + uint32(n)
				ud.Field523_2, ud.Field523_3 = 0x34+uint8(n), 0xa5+uint8(n)
				ud.Stamina, ud.Field282_2 = 0xa5+uint8(n), 0xa5a5+uint16(n)*0x101
				ud.Field137, ud.Field123 = row.Input.DeadFrame, row.Input.Duration
				if row.Input.Enemy {
					ud.CurrentEnemy = current
				}
				for i := range ud.Waypoints {
					wp, enemy := &Waypoint{}, &Object{}
					waypointIDs[wp] = 0x41000000 + uint32(n)*0x100 + uint32(i) + 1
					enemyIDs[enemy] = 0x42000000 + uint32(n)*0x100 + uint32(i) + 1
					ud.Waypoints[i], ud.SeenEnemies[i] = wp, enemy
				}
			}
			entry.StatusFlags = row.Input.Status
			if !row.Input.Definition {
				entry.MonsterDef = nil
			}
			mutate := func(at string) {
				for _, m := range row.Input.Actions {
					if m.At != at {
						continue
					}
					if strings.HasPrefix(m.Kind, "swap-update") {
						unit.UpdateData = unsafe.Pointer(replacement)
					}
					switch m.Kind {
					case "entry-fire", "swap-update-entry-fire":
						entry.StatusFlags = object.MonStatusOnFire
					case "entry-clear", "swap-update-entry-clear":
						entry.StatusFlags = 0
					case "entry-delete", "swap-update-entry-delete":
						entry.MonsterDef.StatusFlags92 = 1
					case "entry-retain":
						entry.MonsterDef.StatusFlags92 = 0
					case "swap-update-replacement-delete":
						b.StatusFlags92 = 0
					case "definition-B":
						entry.MonsterDef = b
					case "nil-definition":
						entry.MonsterDef = nil
					case "nil-update":
						unit.UpdateData = nil
					case "change-counts":
						entry.Field74, entry.Field282_1 = 1, 1
					case "swap-update":
					default:
						t.Fatal("unreviewed mutation", m.Kind)
					}
				}
			}
			view := func(ud *MonsterUpdateData) monsterDeathUpdateView544EC0 {
				out := monsterDeathUpdateView544EC0{Waypoints: ud.Field74, Seen: ud.Field282_1, CurrentEnemy: enemyIDs[ud.CurrentEnemy], Field91: ud.Field91, Electric: ud.Field523_2, ElectricCanary: ud.Field523_3, StaminaCanary: [3]uint8{ud.Stamina, uint8(ud.Field282_2), uint8(ud.Field282_2 >> 8)}, Status: ud.StatusFlags, Definition: "nil"}
				if ud.MonsterDef == a {
					out.Definition = "A"
				} else if ud.MonsterDef == b {
					out.Definition = "B"
				}
				for i := range ud.Waypoints {
					out.WaypointIDs[i], out.EnemyIDs[i] = waypointIDs[ud.Waypoints[i]], enemyIDs[ud.SeenEnemies[i]]
				}
				return out
			}
			events, snapshots := make([]string, 0), make([]monsterDeathUpdateSnapshot544EC0, 0)
			capture := func(at string) {
				snapshots = append(snapshots, monsterDeathUpdateSnapshot544EC0{At: at, UpdateNil: unit.Update == nil, Entry: view(entry), Replacement: view(replacement)})
			}
			identity := func(got *Object) {
				if got != unit {
					t.Fatal("service object identity")
				}
			}
			zombie := func(got *Object) bool {
				identity(got)
				events = append(events, fmt.Sprintf("zombie:%t", row.Input.Zombie))
				mutate("zombie")
				return row.Input.Zombie
			}
			reason := ""
			unsupported := func(s string, got *Object) { identity(got); reason = s }
			handled, panicked := false, false
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
						t.Errorf("native death update panic: %v", r)
					}
				}()
				if row.Input.Entry == "end" {
					handled = new(Server).MonsterActionDyingEnd544CA0(unit, MonsterActionDyingRuntime544C40{IsZombie: zombie, ZombieBurnDelete: func(got *Object) { identity(got); events = append(events, "burn-delete") }, Unsupported: unsupported})
				} else {
					s := new(Server)
					s.SetFrame(row.Input.Frame)
					handled = s.MonsterActionDeadUpdate544EC0(unit, MonsterActionDeadRuntime544D80{
						IsZombie: zombie,
						RemoveUpdatable: func(got *Object) {
							identity(got)
							events = append(events, "remove")
							capture("remove")
							mutate("remove")
						},
						DelayedDelete: func(got *Object) { identity(got); events = append(events, "delete"); capture("delete") },
						SparkExplosion: func(p types.Pointf, strength byte) {
							if p != unit.PosVec || strength != 100 {
								t.Fatal("spark arguments")
							}
							events = append(events, "spark")
							mutate("spark")
						},
						ZombieBurnDelete: func(got *Object) { identity(got); events = append(events, "burn-delete") },
						RaiseZombie:      func(got *Object) { identity(got); events = append(events, "raise") },
						Unsupported:      unsupported,
					})
				}
			}()
			if panicked {
				return
			}
			update := "nil"
			if unit.UpdateData == unsafe.Pointer(entry) {
				update = "entry"
			} else if unit.UpdateData == unsafe.Pointer(replacement) {
				update = "replacement"
			}
			if handled != (len(row.Faults) == 0) || !reflect.DeepEqual(events, row.Events) || view(entry) != row.Entry || view(replacement) != row.Replacement || update != row.Update || (unit.Update == nil) != row.UpdateNil || unit.ObjFlags != row.Flags || !reflect.DeepEqual(snapshots, row.Snapshots) {
				t.Fatalf("handled=%t reason=%q events=%v/%v entry=%+v/%+v replacement=%+v/%+v update=%s/%s nil=%t/%t snapshots=%+v/%+v", handled, reason, events, row.Events, view(entry), row.Entry, view(replacement), row.Replacement, update, row.Update, unit.Update == nil, row.UpdateNil, snapshots, row.Snapshots)
			}
			if len(row.Faults) == 0 && reason != "" {
				t.Fatal(reason)
			}
			if len(row.Faults) != 0 {
				want := "missing live monster definition"
				if row.Faults[0].Address == 296 {
					want = "missing live monster dead update"
				}
				if reason != want {
					t.Fatalf("fault boundary reason=%q want=%q", reason, want)
				}
			}
		})
	}
}

func TestMonsterDeathUpdateNativeBoundaries544EC0(t *testing.T) {
	for _, name := range []string{"waypoint-count", "seen-count", "live-delete-service", "missing-remove", "dying-burn-service"} {
		t.Run(name, func(t *testing.T) {
			unit := monsterActionTestObject50A910(t)
			ud := unit.UpdateDataMonster()
			ud.MonsterDef = &MonsterDef{}
			ud.Field523_2 = 7
			ud.Field523_3 = 0xa5
			ud.Field91 = 9
			ud.Field74 = 1
			ud.Field282_1 = 1
			ud.Waypoints[0] = &Waypoint{}
			ud.SeenEnemies[0] = &Object{}
			ud.CurrentEnemy = &Object{}
			unit.Update = unsafe.Pointer(unit)
			before := *ud
			reason := ""
			removed := 0
			r := MonsterActionDeadRuntime544D80{IsZombie: func(*Object) bool { return false }, RemoveUpdatable: func(*Object) {
				removed++
				if name == "live-delete-service" {
					ud.MonsterDef.StatusFlags92 = 1
				}
			}, Unsupported: func(s string, got *Object) {
				if got != unit {
					t.Fatal("identity")
				}
				reason = s
			}}
			if name == "waypoint-count" {
				ud.Field74 = 17
			}
			if name == "seen-count" {
				ud.Field282_1 = 17
			}
			if name == "missing-remove" {
				r.RemoveUpdatable = nil
			}
			handled := false
			if name == "dying-burn-service" {
				ud.StatusFlags = object.MonStatusOnFire
				handled = new(Server).MonsterActionDyingEnd544CA0(unit, MonsterActionDyingRuntime544C40{IsZombie: func(*Object) bool { return true }, Unsupported: r.Unsupported})
				if reason != "burning zombie deletion" || unit.Update == nil || ud.Field523_2 != 7 {
					t.Fatal("dying use boundary", reason)
				}
			} else {
				handled = new(Server).MonsterActionDeadUpdate544EC0(unit, r)
				switch name {
				case "waypoint-count":
					if reason != "monster dead waypoint count" || removed != 1 || unit.Update != nil || ud.Field523_2 != 0 || ud.Field74 != 17 || ud.Field91 != 9 || ud.Waypoints != before.Waypoints {
						t.Fatal("waypoint boundary", reason)
					}
				case "seen-count":
					if reason != "monster dead seen-enemy count" || removed != 1 || unit.Update != nil || ud.Field523_2 != 0 || ud.Field74 != 0 || ud.Field91 != 0 || ud.Waypoints[0] != nil || ud.Field282_1 != 17 || ud.SeenEnemies != before.SeenEnemies {
						t.Fatal("seen boundary", reason)
					}
				case "live-delete-service":
					if reason != "dead monster deletion" || removed != 1 || unit.Update != nil || ud.Field523_2 != 0 || ud.CurrentEnemy != nil {
						t.Fatal("late delete boundary", reason)
					}
				case "missing-remove":
					if reason != "" || removed != 0 || unit.Update == nil || *ud != before {
						t.Fatal("missing removal preflight", reason)
					}
				}
			}
			if handled || ud.Field523_3 != 0xa5 {
				t.Fatal("unsupported native boundary or adjacent byte changed")
			}
		})
	}
}
