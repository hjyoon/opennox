package server

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type deadUpdateOriginalView544EC0 struct {
	Waypoints       uint32               `json:"waypoints"`
	WaypointIDs     [16]uint32           `json:"waypoint_ids"`
	Seen            uint8                `json:"seen"`
	EnemyIDs        [16]uint32           `json:"enemy_ids"`
	Enemy           bool                 `json:"enemy"`
	Field91         uint32               `json:"field91"`
	Electric        uint8                `json:"electric"`
	ElectricCanary  uint8                `json:"electric_canary"`
	StaminaCanaries [3]uint8             `json:"stamina_canaries"`
	Status          object.MonsterStatus `json:"status"`
	Definition      int                  `json:"definition"`
	DeadFrame       uint32               `json:"dead_frame"`
	Duration        uint32               `json:"duration"`
}

type deadUpdateOriginalState544EC0 struct {
	Update      int                             `json:"update"`
	UpdateNil   bool                            `json:"update_nil"`
	Class       uint32                          `json:"cls"`
	Subclass    uint32                          `json:"subclass"`
	Flags       uint32                          `json:"flags"`
	Position    [2]float32                      `json:"position"`
	Frame       uint32                          `json:"frame"`
	Updates     [2]deadUpdateOriginalView544EC0 `json:"updates"`
	Definitions [2]object.MonsterStatus         `json:"definitions"`
}

type deadUpdateOriginalEvent544EC0 struct {
	Kind  string                        `json:"kind"`
	Unit  int                           `json:"unit"`
	State deadUpdateOriginalState544EC0 `json:"state"`
}

// The existing component APIs return handled bools. Their contract compares
// original state, arguments and ordered callbacks, excluding raw EAX. The
// separate native C composition checks Cleanup's actual pointer return.
func TestMonsterDeadUpdateOriginalNative544EC0(t *testing.T) {
	path := "testdata/monster_dead_update_native_original_544ec0.json"
	if p := os.Getenv("NOX_MONSTER_DEAD_UPDATE_ORIGINAL_FIXTURE"); p != "" {
		path = p
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA     string `json:"original_sha256"`
		Control uint16 `json:"control_word"`
		Rows    []struct {
			Input struct {
				Name               string               `json:"name"`
				Entry              string               `json:"entry"`
				Update             int                  `json:"update"`
				Zombie             uint32               `json:"zombie"`
				Status             object.MonsterStatus `json:"status"`
				AlternateStatus    object.MonsterStatus `json:"alternate_status"`
				Frame              uint32               `json:"frame"`
				DeadFrame          uint32               `json:"dead_frame"`
				Duration           uint32               `json:"duration"`
				Enemy              bool                 `json:"enemy"`
				Waypoints          uint32               `json:"waypoints"`
				Seen               uint8                `json:"seen"`
				AlternateWaypoints uint32               `json:"alternate_waypoints"`
				AlternateSeen      uint8                `json:"alternate_seen"`
				Delete             object.MonsterStatus `json:"delete"`
				AlternateDelete    object.MonsterStatus `json:"alternate_delete"`
				Mutation           struct {
					Stage  string `json:"stage"`
					Action string `json:"action"`
					Value  uint32 `json:"value"`
				} `json:"mutation"`
			} `json:"input"`
			Result struct {
				Events []deadUpdateOriginalEvent544EC0 `json:"events"`
				State  deadUpdateOriginalState544EC0   `json:"state"`
			} `json:"result"`
			CleanupReturn int `json:"cleanup_return_update"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.Control != 0xe3f || len(f.Rows) != 694 {
		t.Fatal("original DeadUpdate/Cleanup identity/profile/count")
	}
	for _, row := range f.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			defer func() {
				if v := recover(); v != nil {
					t.Errorf("valid original native context panic: %v", v)
				}
			}()
			type guarded struct {
				Before      [32]byte
				Unit        Object
				Updates     [2]MonsterUpdateData
				Definitions [2]MonsterDef
				Waypoints   [2][16]Waypoint
				Enemies     [2][16]Object
				Current     Object
				After       [32]byte
			}
			p, free := alloc.New(guarded{})
			defer free()
			*p = guarded{}
			for i := range p.Before {
				p.Before[i], p.After[i] = byte(0x21+i), byte(0xed-i)
			}
			unit := &p.Unit
			unit.ObjClass, unit.ObjSubClass, unit.ObjFlags = object.ClassMonster, 0xa5, 0x12008014
			unit.PosVec = types.Ptf(321, 654)
			unit.Update, unit.UpdateData = unsafe.Pointer(&p.Before[0]), unsafe.Pointer(&p.Updates[row.Input.Update])
			waypointIDs, enemyIDs := map[*Waypoint]uint32{}, map[*Object]uint32{}
			waypoints, enemies := map[uint32]*Waypoint{}, map[uint32]*Object{}
			for n := range p.Updates {
				ud := &p.Updates[n]
				ud.MonsterDef = &p.Definitions[n]
				ud.Field74, ud.Field282_1, ud.StatusFlags = row.Input.Waypoints, row.Input.Seen, row.Input.Status
				if n == 1 {
					ud.Field74, ud.Field282_1, ud.StatusFlags = row.Input.AlternateWaypoints, row.Input.AlternateSeen, row.Input.AlternateStatus
				}
				ud.Field91, ud.Field137, ud.Field123 = 0x12345678+uint32(n), row.Input.DeadFrame, row.Input.Duration
				ud.Stamina, ud.Field282_2 = 0xa5+uint8(n), 0xa5a5+uint16(n)*0x101
				ud.Field523_2, ud.Field523_3 = 0x34+uint8(n), 0xd5+uint8(n)
				if row.Input.Enemy {
					ud.CurrentEnemy = &p.Current
				}
				for i := range ud.Waypoints {
					wp, enemy := &p.Waypoints[n][i], &p.Enemies[n][i]
					wi, ei := 0x41000000+uint32(n)*0x100+uint32(i)+1, 0x42000000+uint32(n)*0x100+uint32(i)+1
					waypointIDs[wp], enemyIDs[enemy], waypoints[wi], enemies[ei] = wi, ei, wp, enemy
					ud.Waypoints[i], ud.SeenEnemies[i] = wp, enemy
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(unsafe.Pointer(ud)) <= 0xffffffff || uintptr(unsafe.Pointer(ud.MonsterDef)) <= 0xffffffff || uintptr(unsafe.Pointer(&p.Waypoints[n][0])) <= 0xffffffff || uintptr(unsafe.Pointer(&p.Enemies[n][0])) <= 0xffffffff) {
					t.Fatal("native links must exceed 4 GiB")
				}
			}
			p.Definitions[0].StatusFlags92, p.Definitions[1].StatusFlags92 = row.Input.Delete, row.Input.AlternateDelete
			expected := *p
			s := new(Server)
			s.SetFrame(row.Input.Frame)
			view := func(ud *MonsterUpdateData) deadUpdateOriginalView544EC0 {
				definition := -1
				for i := range p.Definitions {
					if ud.MonsterDef == &p.Definitions[i] {
						definition = i
					}
				}
				if definition < 0 {
					t.Fatal("unexpected definition identity")
				}
				if ud.CurrentEnemy != nil && ud.CurrentEnemy != &p.Current {
					t.Fatal("unexpected current enemy")
				}
				v := deadUpdateOriginalView544EC0{Waypoints: ud.Field74, Seen: ud.Field282_1, Enemy: ud.CurrentEnemy != nil, Field91: ud.Field91, Electric: ud.Field523_2, ElectricCanary: ud.Field523_3, StaminaCanaries: [3]uint8{ud.Stamina, uint8(ud.Field282_2), uint8(ud.Field282_2 >> 8)}, Status: ud.StatusFlags, Definition: definition, DeadFrame: ud.Field137, Duration: ud.Field123}
				for i := range ud.Waypoints {
					var ok bool
					v.WaypointIDs[i], ok = waypointIDs[ud.Waypoints[i]]
					if !ok && ud.Waypoints[i] != nil {
						t.Fatal("unexpected waypoint identity")
					}
					v.EnemyIDs[i], ok = enemyIDs[ud.SeenEnemies[i]]
					if !ok && ud.SeenEnemies[i] != nil {
						t.Fatal("unexpected seen-enemy identity")
					}
				}
				return v
			}
			state := func() deadUpdateOriginalState544EC0 {
				index := -1
				for i := range p.Updates {
					if unit.UpdateData == unsafe.Pointer(&p.Updates[i]) {
						index = i
					}
				}
				if index < 0 {
					t.Fatal("unexpected current update identity")
				}
				return deadUpdateOriginalState544EC0{Update: index, UpdateNil: unit.Update == nil, Class: uint32(unit.ObjClass), Subclass: uint32(unit.ObjSubClass), Flags: uint32(unit.ObjFlags), Position: [2]float32{unit.PosVec.X, unit.PosVec.Y}, Frame: s.Frame(), Updates: [2]deadUpdateOriginalView544EC0{view(&p.Updates[0]), view(&p.Updates[1])}, Definitions: [2]object.MonsterStatus{p.Definitions[0].StatusFlags92, p.Definitions[1].StatusFlags92}}
			}
			mutate := func(stage string) {
				m := row.Input.Mutation
				if m.Stage != stage {
					return
				}
				a, value := m.Action, m.Value
				switch a {
				case "swap-update", "primary-status", "alternate-status", "definition-alt", "primary-delete", "alternate-delete", "frame", "counts", "enemy", "all":
				default:
					t.Fatal("unknown declared mutation", a)
				}
				if a == "swap-update" || a == "all" {
					unit.UpdateData = unsafe.Pointer(&p.Updates[1])
				}
				if a == "primary-status" || a == "all" {
					p.Updates[0].StatusFlags = object.MonsterStatus(value)
				}
				if a == "alternate-status" || a == "all" {
					p.Updates[1].StatusFlags = object.MonsterStatus(value ^ 0x80000)
				}
				if a == "definition-alt" || a == "all" {
					p.Updates[0].MonsterDef = &p.Definitions[1]
				}
				if a == "primary-delete" || a == "all" {
					p.Definitions[0].StatusFlags92 = object.MonsterStatus(value)
				}
				if a == "alternate-delete" || a == "all" {
					p.Definitions[1].StatusFlags92 = object.MonsterStatus(value)
				}
				if a == "frame" || a == "all" {
					s.SetFrame(value)
				}
				if a == "counts" || a == "all" {
					p.Updates[0].Field74, p.Updates[0].Field282_1 = 1, 1
					p.Updates[1].Field74, p.Updates[1].Field282_1 = 16, 16
				}
				if a == "enemy" || a == "all" {
					p.Updates[0].CurrentEnemy = nil
					p.Updates[1].CurrentEnemy = &p.Current
				}
				if a == "all" {
					unit.ObjClass = 0x400000
					unit.ObjSubClass = 0xfedcba98
					unit.ObjFlags = 0x87654321
					unit.PosVec = types.Ptf(-37.5, 19.25)
				}
			}
			events := make([]deadUpdateOriginalEvent544EC0, 0)
			emit := func(kind string, got *Object) {
				if got != unit {
					t.Fatal("native service unit identity")
				}
				events = append(events, deadUpdateOriginalEvent544EC0{Kind: kind, Unit: 0, State: state()})
				mutate(kind)
			}
			runtime := MonsterActionDeadRuntime544D80{
				IsZombie:        func(got *Object) bool { emit("zombie", got); return row.Input.Zombie != 0 },
				RemoveUpdatable: func(got *Object) { emit("remove", got) },
				SparkExplosion: func(pos types.Pointf, strength byte) {
					if pos != unit.PosVec || strength != 100 {
						t.Fatal("native spark arguments")
					}
					emit("spark", unit)
				},
				ZombieBurnDelete: func(got *Object) { emit("burn", got) },
				RaiseZombie:      func(got *Object) { emit("raise", got) },
				DelayedDelete:    func(got *Object) { emit("delete", got) },
				Unsupported:      func(reason string, _ *Object) { t.Fatal("valid original context rejected", reason) },
			}
			handled := false
			if row.Input.Entry == "cleanup" {
				handled = monsterDeadCleanup544F70(unit, runtime)
				if row.CleanupReturn != row.Input.Update {
					t.Fatal("original cleanup return identity")
				}
			} else if row.Input.Entry == "update" {
				handled = s.MonsterActionDeadUpdate544EC0(unit, runtime)
			} else {
				t.Fatal("unknown entry")
			}
			if !handled || !reflect.DeepEqual(events, row.Result.Events) || state() != row.Result.State {
				t.Fatalf("handled=%t events=%+v want=%+v state=%+v want=%+v", handled, events, row.Result.Events, state(), row.Result.State)
			}
			want := row.Result.State
			expected.Unit.ObjClass, expected.Unit.ObjSubClass, expected.Unit.ObjFlags = object.Class(want.Class), object.SubClass(want.Subclass), object.Flags(want.Flags)
			expected.Unit.PosVec = types.Ptf(want.Position[0], want.Position[1])
			expected.Unit.UpdateData = unsafe.Pointer(&p.Updates[want.Update])
			if want.UpdateNil {
				expected.Unit.Update = nil
			}
			for n, v := range want.Updates {
				ud := &expected.Updates[n]
				ud.Field74, ud.Field282_1, ud.Field91 = v.Waypoints, v.Seen, v.Field91
				ud.Field523_2, ud.Field523_3 = v.Electric, v.ElectricCanary
				ud.Stamina, ud.Field282_2 = v.StaminaCanaries[0], uint16(v.StaminaCanaries[1])|uint16(v.StaminaCanaries[2])<<8
				ud.StatusFlags, ud.MonsterDef, ud.Field137, ud.Field123 = v.Status, &p.Definitions[v.Definition], v.DeadFrame, v.Duration
				ud.CurrentEnemy = nil
				if v.Enemy {
					ud.CurrentEnemy = &p.Current
				}
				for i := range ud.Waypoints {
					ud.Waypoints[i], ud.SeenEnemies[i] = waypoints[v.WaypointIDs[i]], enemies[v.EnemyIDs[i]]
				}
			}
			for i, v := range want.Definitions {
				expected.Definitions[i].StatusFlags92 = v
			}
			raw := func(v *guarded) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
			if !bytes.Equal(raw(p), raw(&expected)) {
				t.Fatal("complete native unit/two-updates/definitions/waypoints/enemies/canaries comparison")
			}
		})
	}
}
