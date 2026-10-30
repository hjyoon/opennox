package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

// The original keeps its entry update pointer, then reloads its definition and
// function at use. Native event7 is the existing OpenNox adapter extension;
// the original script call has only block/caller/trigger arguments.
func TestMonsterDeathLiveOriginalNearestControl544C40(t *testing.T) {
	raw, err := os.ReadFile("testdata/monster_death_live_544c40.json")
	if err != nil {
		t.Fatal(err)
	}
	type mutation struct {
		At   string `json:"at"`
		Kind string `json:"kind"`
	}
	var fixture struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"code_bytes_unchanged"`
		Rows      []struct {
			Input struct {
				Name          string     `json:"name"`
				Entry         string     `json:"entry"`
				Sound         bool       `json:"sound"`
				Initial       string     `json:"initial"`
				Actions       []mutation `json:"actions"`
				Soul          bool       `json:"soul"`
				Zombie        []bool     `json:"zombie"`
				Duration      []float32  `json:"duration"`
				Random        int        `json:"random_return"`
				Flags         *uint32    `json:"flags"`
				EntryDuration *uint32    `json:"entry_duration"`
			} `json:"input"`
			Events              []string     `json:"events"`
			Called              []string     `json:"called"`
			Flags               object.Flags `json:"flags"`
			EntryDuration       uint32       `json:"entry_duration"`
			ReplacementDuration uint32       `json:"replacement_duration"`
			Update              string       `json:"current_update"`
			Motion              []uint32     `json:"motion_bits"`
			ZombieQueries       int          `json:"zombie_queries"`
			RandomArgs          []int        `json:"random_args"`
			Faults              []struct {
				Address uint32 `json:"address"`
			} `json:"faults"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || len(fixture.Rows) != 38 {
		t.Fatal("unchanged original death fixture")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			unit := monsterActionTestObject50A910(t)
			entry := unit.UpdateDataMonster()
			replacement := new(MonsterUpdateData)
			var marks [3]byte
			fn := map[string]unsafe.Pointer{"none": nil, "A": unsafe.Pointer(&marks[0]), "B": unsafe.Pointer(&marks[1]), "C": unsafe.Pointer(&marks[2])}
			a, b, c := &MonsterDef{}, &MonsterDef{}, &MonsterDef{}
			setFn := func(d *MonsterDef, p unsafe.Pointer) {
				if row.Input.Entry == "dying" {
					d.DieFunc228 = p
				} else {
					d.DeadFunc232 = p
				}
			}
			setFn(a, fn[row.Input.Initial])
			setFn(b, fn["B"])
			setFn(c, fn["C"])
			entry.MonsterDef, replacement.MonsterDef = a, c
			entry.Field123, replacement.Field123 = 0xcccc1111, 0xdddd2222
			if row.Input.EntryDuration != nil {
				entry.Field123 = *row.Input.EntryDuration
			}
			unit.ObjFlags = object.Flags(0x12000004)
			if row.Input.Flags != nil {
				unit.ObjFlags = object.Flags(*row.Input.Flags)
			}
			unit.VelVec, unit.ForceVec, unit.Pos24 = types.Ptf(1, 2), types.Ptf(3, 4), types.Ptf(5, 6)
			entry.ScriptDeath = ScriptCallback{Flags: 0xa5, Func: 17}
			replacement.ScriptDeath = ScriptCallback{Flags: 0x5a, Func: 19}
			var sounds [16]uint32
			sounds[15] = 715
			if row.Input.Sound {
				entry.SoundSet122 = unsafe.Pointer(&sounds[0])
			}
			if row.Input.Soul {
				unit.Field131 = 14
				unit.ObjSubClass = 0x10000
			}
			mutate := func(at string) {
				for _, m := range row.Input.Actions {
					if m.At != at {
						continue
					}
					switch m.Kind {
					case "function-B":
						setFn(entry.MonsterDef, fn["B"])
					case "definition-B":
						entry.MonsterDef = b
					case "clear-function":
						setFn(entry.MonsterDef, nil)
					case "nil-definition":
						entry.MonsterDef = nil
					case "swap-update":
						unit.UpdateData = unsafe.Pointer(replacement)
					case "replacement-function-B":
						setFn(replacement.MonsterDef, fn["B"])
					case "flags-B":
						unit.ObjFlags = object.Flags(0x53000004)
					default:
						t.Fatal("unreviewed mutation", m.Kind)
					}
				}
			}
			var events, called []string
			var randomArgs []int
			zombieCalls := 0
			reason := ""
			accepted := func(p unsafe.Pointer) bool { return p == fn["A"] || p == fn["B"] || p == fn["C"] }
			callback := func(p unsafe.Pointer, got *Object) {
				if got != unit {
					t.Fatal("callback unit identity")
				}
				name := ""
				for n, v := range fn {
					if v == p {
						name = n
					}
				}
				if name == "" || name == "none" {
					t.Fatal("unregistered callback")
				}
				if row.Input.Entry == "dead" && (unit.VelVec != (types.Pointf{}) || unit.ForceVec != (types.Pointf{}) || unit.Pos24 != (types.Pointf{})) {
					t.Fatal("dead callback precedes zero motion")
				}
				events = append(events, "callback:"+name)
				called = append(called, name)
				if row.Input.Entry == "dead" {
					mutate("dead")
				} else {
					mutate("die")
				}
			}
			unsupported := func(s string, got *Object) {
				if got != unit {
					t.Fatal("unsupported identity")
				}
				reason = s
			}
			handled := false
			if row.Input.Entry == "dying" {
				handled = new(Server).MonsterActionDyingStart544C40(unit, MonsterActionDyingRuntime544C40{
					AudioEvent: func(id uint32, got *Object) {
						if id != 715 || got != unit {
							t.Fatal("audio identity")
						}
						events = append(events, "audio:715")
						mutate("audio")
					},
					ScriptCallback: func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) {
						if block != &entry.ScriptDeath || caller != nil || trigger != unit || event != NoxEventMonsterDead {
							t.Fatal("cached script block/native extension identity")
						}
						events = append(events, "script:entry")
						mutate("script")
					},
					CanDieFunc: accepted, DieFunc: callback, Unsupported: unsupported,
				})
			} else {
				handled = new(Server).MonsterActionDeadStart544D80(unit, MonsterActionDeadRuntime544D80{
					// The unchanged 38-row producer explicitly uses FPCW 0x037f.
					ConvertDuration: playerCollideRound4E8460,
					CreateReleasedSoul: func(got *Object) {
						if got != unit {
							t.Fatal("soul identity")
						}
						events = append(events, "soul")
						mutate("soul")
					},
					CanDeadFunc: accepted, DeadFunc: callback,
					IsZombie: func(got *Object) bool {
						if got != unit || zombieCalls >= len(row.Input.Zombie) {
							t.Fatal("zombie query identity/count")
						}
						answer := row.Input.Zombie[zombieCalls]
						events = append(events, fmt.Sprintf("zombie:%t", answer))
						mutate(fmt.Sprintf("zombie%d", zombieCalls))
						zombieCalls++
						return answer
					},
					ZombieDeadDuration: func(index int) float32 {
						if index < 0 || index >= len(row.Input.Duration) {
							t.Fatal("duration index")
						}
						events = append(events, fmt.Sprintf("duration:%d", index))
						return row.Input.Duration[index]
					},
					RandomInt: func(minimum, maximum int) int {
						randomArgs = append(randomArgs, minimum, maximum)
						events = append(events, fmt.Sprintf("random:%d:%d", minimum, maximum))
						mutate("random")
						return row.Input.Random
					},
					Unsupported: unsupported,
				})
			}
			current := "entry"
			if unit.UpdateData == unsafe.Pointer(replacement) {
				current = "replacement"
			}
			motion := []uint32{math.Float32bits(unit.VelVec.X), math.Float32bits(unit.VelVec.Y), math.Float32bits(unit.ForceVec.X), math.Float32bits(unit.ForceVec.Y), math.Float32bits(unit.Pos24.X), math.Float32bits(unit.Pos24.Y)}
			sameStrings := func(a, b []string) bool {
				if len(a) != len(b) {
					return false
				}
				for i := range a {
					if a[i] != b[i] {
						return false
					}
				}
				return true
			}
			sameInts := len(randomArgs) == len(row.RandomArgs)
			if sameInts {
				for i := range randomArgs {
					if randomArgs[i] != row.RandomArgs[i] {
						sameInts = false
					}
				}
			}
			if handled != (len(row.Faults) == 0) || !sameStrings(events, row.Events) || !sameStrings(called, row.Called) || unit.ObjFlags != row.Flags || entry.Field123 != row.EntryDuration || replacement.Field123 != row.ReplacementDuration || current != row.Update || !reflect.DeepEqual(motion, row.Motion) || zombieCalls != row.ZombieQueries || !sameInts {
				t.Fatalf("handled=%t faults=%v reason=%q events=%v/%v called=%v/%v flags=%x/%x durations=%x/%x,%x/%x update=%s/%s motion=%v/%v queries=%d/%d random=%v/%v", handled, row.Faults, reason, events, row.Events, called, row.Called, unit.ObjFlags, row.Flags, entry.Field123, row.EntryDuration, replacement.Field123, row.ReplacementDuration, current, row.Update, motion, row.Motion, zombieCalls, row.ZombieQueries, randomArgs, row.RandomArgs)
			}
			if len(row.Faults) == 0 {
				if reason != "" {
					t.Fatal(reason)
				}
			} else if !strings.Contains(reason, "live monster definition") {
				t.Fatal("original live nil-definition fault was not reported", reason)
			}
		})
	}
}

func TestMonsterDeathLiveUnsupportedCallbacks544C40(t *testing.T) {
	for _, name := range []string{"dying-audio-unknown", "dying-script-unknown", "dying-script-missing-service", "dead-soul-unknown", "dead-soul-missing-service", "dead-late-zombie-missing-duration"} {
		t.Run(name, func(t *testing.T) {
			unit := monsterActionTestObject50A910(t)
			entry := unit.UpdateDataMonster()
			var good, bad byte
			accepted := unsafe.Pointer(&good)
			unknown := unsafe.Pointer(&bad)
			entry.MonsterDef = &MonsterDef{}
			unit.VelVec = types.Ptf(1, 2)
			var events []string
			reason := ""
			report := func(s string, _ *Object) { reason = s }
			if strings.HasPrefix(name, "dying") {
				entry.MonsterDef.DieFunc228 = accepted
				missing := strings.HasSuffix(name, "missing-service")
				if missing {
					entry.MonsterDef.DieFunc228 = nil
				}
				var sound [16]uint32
				entry.SoundSet122 = unsafe.Pointer(&sound[0])
				r := MonsterActionDyingRuntime544C40{CanDieFunc: func(p unsafe.Pointer) bool { return p == accepted }, DieFunc: func(unsafe.Pointer, *Object) { t.Fatal("unported live callback called") }, Unsupported: report}
				r.AudioEvent = func(uint32, *Object) {
					events = append(events, "audio")
					if name == "dying-audio-unknown" {
						entry.MonsterDef.DieFunc228 = unknown
					}
				}
				r.ScriptCallback = func(*ScriptCallback, *Object, *Object, ScriptEventType) {
					events = append(events, "script")
					if name == "dying-script-unknown" {
						entry.MonsterDef.DieFunc228 = unknown
					}
					if missing {
						entry.MonsterDef.DieFunc228 = accepted
					}
				}
				if missing {
					r.DieFunc = nil
				}
				if new(Server).MonsterActionDyingStart544C40(unit, r) || !strings.Contains(reason, "live monster die function") || !reflect.DeepEqual(events, []string{"audio", "script"}) {
					t.Fatalf("live dying boundary reason=%q events=%v", reason, events)
				}
			} else {
				entry.MonsterDef.DeadFunc232 = accepted
				unit.Field131 = 14
				unit.ObjSubClass = 0x10000
				missing := name == "dead-soul-missing-service"
				if missing {
					entry.MonsterDef.DeadFunc232 = nil
				}
				late := name == "dead-late-zombie-missing-duration"
				zombie := false
				r := MonsterActionDeadRuntime544D80{CanDeadFunc: func(p unsafe.Pointer) bool { return p == accepted }, DeadFunc: func(unsafe.Pointer, *Object) {
					if !late {
						t.Fatal("unported live dead callback called")
					}
					events = append(events, "dead")
					zombie = true
				}, IsZombie: func(*Object) bool { return zombie }, ZombieDeadDuration: func(int) float32 { return 1 }, RandomInt: func(int, int) int { return 1 }, Unsupported: report}
				r.CreateReleasedSoul = func(*Object) {
					events = append(events, "soul")
					if missing {
						entry.MonsterDef.DeadFunc232 = accepted
					} else if !late {
						entry.MonsterDef.DeadFunc232 = unknown
					}
				}
				if missing {
					r.DeadFunc = nil
				}
				if late {
					r.ZombieDeadDuration = nil
					r.RandomInt = nil
				}
				handled := new(Server).MonsterActionDeadStart544D80(unit, r)
				wantReason := "live monster dead function"
				wantEvents := []string{"soul"}
				if late {
					wantReason = "live zombie dead duration"
					wantEvents = append(wantEvents, "dead")
				}
				if handled || !strings.Contains(reason, wantReason) || !reflect.DeepEqual(events, wantEvents) || unit.VelVec != (types.Pointf{}) {
					t.Fatalf("live dead boundary handled=%t reason=%q events=%v", handled, reason, events)
				}
			}
		})
	}
}
