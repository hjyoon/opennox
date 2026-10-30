package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type fightRunOriginalState531E20 struct {
	Cls         uint32
	Update      int
	Flags       [2]uint32
	SoundSets   [2]int `json:"sound_sets"`
	Enemies     [2]int
	Blocks      [2][2]uint32
	SoundIDs    [2]uint32 `json:"sound_ids"`
	Frame       uint32
	CopiedFrame uint32 `json:"copied_frame"`
}

type fightRunOriginalInput531E20 struct {
	fightRunOriginalState531E20
	Name      string
	Entry     uint32
	Mutations [3]string
}

type fightRunOriginalEvent531E20 struct {
	Kind  string
	Args  []int64
	State fightRunOriginalState531E20
}

func TestMonsterFightRunStateOriginalNative531E20(t *testing.T) {
	file := os.Getenv("NOX_MONSTER_FIGHT_RUN_ORIGINAL_FIXTURE")
	if file == "" {
		file = "testdata/monster_fight_run_state_native_original_531e20.json.gz"
	}
	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var fixture struct {
		OriginalSHA256 string `json:"original_sha256"`
		ControlWord    uint16 `json:"control_word"`
		Rows           []struct {
			Input  fightRunOriginalInput531E20
			Result struct {
				Events         []fightRunOriginalEvent531E20
				State          fightRunOriginalState531E20
				ReturnedStatus uint32 `json:"returned_status"`
			}
		}
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 5618 || fixture.ControlWord != 0xe3f && fixture.ControlWord != 0x23f {
		t.Fatal("unsealed fight/run original reference")
	}
	t.Logf("ORIGINAL FIGHT/RUN PROFILE: control=%#x contexts=%d; 13 roots and two unchanged helpers; script has three original arguments, adapter event13 checked separately", fixture.ControlWord, len(fixture.Rows))
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			c := row.Input
			unit, unitCheck := retreatOriginalAlloc545440[Object](t)
			var updates [2]*MonsterUpdateData
			var sounds [2]*[32]uint32
			var enemies [2]*Object
			checks := []func(){unitCheck}
			// Allocate every record before wiring links to the alternate records.
			for i := 0; i < 2; i++ {
				var check func()
				updates[i], check = retreatOriginalAlloc545440[MonsterUpdateData](t)
				checks = append(checks, check)
				sounds[i], check = retreatOriginalAlloc545440[[32]uint32](t)
				checks = append(checks, check)
				enemies[i], check = retreatOriginalAlloc545440[Object](t)
				checks = append(checks, check)
			}
			for i := 0; i < 2; i++ {
				updates[i].StatusFlags = object.MonsterStatus(c.Flags[i])
				updates[i].ScriptChangeFocus = ScriptCallback{Flags: c.Blocks[i][0], Func: int32(c.Blocks[i][1])}
				if ref := c.SoundSets[i]; ref >= 0 {
					updates[i].SoundSet122 = unsafe.Pointer(sounds[ref])
				}
				if ref := c.Enemies[i]; ref >= 0 {
					updates[i].CurrentEnemy = enemies[ref]
				}
				sounds[i][5] = c.SoundIDs[i]
				enemies[i].NetCode, enemies[i].PosVec = uint32(71+i), types.Ptf(float32(i+1), float32(i+8))
			}
			unit.ObjClass, unit.UpdateData = object.Class(c.Cls), unsafe.Pointer(updates[c.Update])
			unit.ObjSubClass, unit.NetCode = 0xdead, 0xc0dec0de
			unit.PosVec, unit.ForceVec = types.Ptf(100, 200), types.Ptf(0.75, -0.5)
			frame, copiedFrame := c.Frame, c.CopiedFrame
			state := func() fightRunOriginalState531E20 {
				v := fightRunOriginalState531E20{Cls: uint32(unit.ObjClass), Update: -1, Frame: frame, CopiedFrame: copiedFrame}
				for i, update := range updates {
					if unit.UpdateData == unsafe.Pointer(update) {
						v.Update = i
					}
					v.Flags[i], v.SoundSets[i], v.Enemies[i] = uint32(update.StatusFlags), -1, -1
					v.Blocks[i] = [2]uint32{update.ScriptChangeFocus.Flags, uint32(update.ScriptChangeFocus.Func)}
					v.SoundIDs[i] = sounds[i][5]
					for j := 0; j < 2; j++ {
						if update.SoundSet122 == unsafe.Pointer(sounds[j]) {
							v.SoundSets[i] = j
						}
						if update.CurrentEnemy == enemies[j] {
							v.Enemies[i] = j
						}
					}
				}
				return v
			}
			if state() != c.fightRunOriginalState531E20 {
				t.Fatalf("reference input differs: got=%+v original=%+v", state(), c.fightRunOriginalState531E20)
			}
			beforeUnit := *unit
			beforeUpdates := [2]MonsterUpdateData{*updates[0], *updates[1]}
			beforeSounds := [2][32]uint32{*sounds[0], *sounds[1]}
			beforeEnemies := [2]Object{*enemies[0], *enemies[1]}
			if unsafe.Sizeof(uintptr(0)) == 8 {
				for _, p := range []unsafe.Pointer{unit.CObj(), unsafe.Pointer(updates[0]), unsafe.Pointer(updates[1]), unsafe.Pointer(sounds[0]), unsafe.Pointer(sounds[1]), enemies[0].CObj(), enemies[1].CObj()} {
					if uintptr(p) <= 0xffffffff {
						t.Fatal("native fight/run records must exceed4GiB")
					}
				}
			}
			mutate := func(index int) {
				cached := c.Update
				switch c.Mutations[index] {
				case "none":
				case "swap-update":
					unit.UpdateData = unsafe.Pointer(updates[1-state().Update])
				case "class-clear":
					unit.ObjClass &^= object.ClassMonster
				case "set-flags":
					updates[cached].StatusFlags = 0x10000
					updates[1-cached].StatusFlags = 0x80008000
				case "set-enemy":
					if updates[cached].CurrentEnemy == enemies[1] {
						updates[cached].CurrentEnemy = nil
					} else {
						updates[cached].CurrentEnemy = enemies[1]
					}
				case "clear-sound":
					updates[cached].SoundSet122 = nil
				case "swap-sound":
					ref := 0
					if updates[cached].SoundSet122 == unsafe.Pointer(sounds[0]) {
						ref = 1
					}
					updates[cached].SoundSet122 = unsafe.Pointer(sounds[ref])
				case "frame":
					frame ^= 0x80000011
				case "set-block":
					updates[cached].ScriptChangeFocus.Flags ^= 0x100
					updates[cached].ScriptChangeFocus.Func = int32(uint32(updates[cached].ScriptChangeFocus.Func) ^ 0x7fffffff)
				default:
					t.Fatal("unknown original callback mutation")
				}
			}
			got := []fightRunOriginalEvent531E20{}
			observe := func(kind string, args ...int64) {
				if args == nil {
					args = []int64{}
				}
				got = append(got, fightRunOriginalEvent531E20{Kind: kind, Args: args, State: state()})
			}
			fault := false
			func() {
				defer func() {
					if recover() != nil {
						fault = true
					}
				}()
				var s Server
				switch c.Entry {
				case 0x531e20:
					s.MonsterActionFightStart531E20(unit, MonsterActionFightStartRuntime531E20{
						AudioEvent: func(id uint32, owner *Object) {
							if owner != unit {
								t.Fatal("audio owner changed")
							}
							observe("audio", int64(id), 0, 0, 0)
							mutate(0)
						},
						ScriptCallback: func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) {
							if block != &updates[c.Update].ScriptChangeFocus || trigger != unit || event != NoxEventMonsterFightStart {
								t.Fatal("cached script identity or separate event13 adapter contract changed")
							}
							ref := int64(-1)
							for i, enemy := range enemies {
								if caller == enemy {
									ref = int64(i)
								}
							}
							observe("script", int64(c.Update), ref, 0)
							mutate(1)
						},
						CopyFrameCounter: func() { observe("copy-frame"); copiedFrame = frame },
						UpdateSight: func(owner *Object) {
							if owner != unit {
								t.Fatal("sight owner changed")
							}
							observe("sight", 0)
							mutate(2)
						},
					})
				case 0x531e90:
					s.MonsterActionFightEnd531E90(unit)
				case 0x534750, 0x544740, 0x544920, 0x5456b0:
					s.MonsterActionRunStart534750(unit)
				case 0x534780, 0x544750, 0x544930, 0x544940, 0x5456c0, 0x546410, 0x546420:
					s.MonsterActionRunEnd534780(unit)
				default:
					t.Fatal("unknown original fight/run entry")
				}
			}()
			if fault || !reflect.DeepEqual(got, row.Result.Events) || state() != row.Result.State || uint32(updates[state().Update].StatusFlags) != row.Result.ReturnedStatus {
				t.Fatalf("original fight/run effects differ: fault=%t events-match=%t state-match=%t got=%+v original=%+v", fault, reflect.DeepEqual(got, row.Result.Events), state() == row.Result.State, state(), row.Result.State)
			}
			want := row.Result.State
			beforeUnit.ObjClass, beforeUnit.UpdateData = object.Class(want.Cls), unsafe.Pointer(updates[want.Update])
			for i := 0; i < 2; i++ {
				beforeUpdates[i].StatusFlags = object.MonsterStatus(want.Flags[i])
				beforeUpdates[i].ScriptChangeFocus = ScriptCallback{Flags: want.Blocks[i][0], Func: int32(want.Blocks[i][1])}
				beforeUpdates[i].SoundSet122, beforeUpdates[i].CurrentEnemy = nil, nil
				if ref := want.SoundSets[i]; ref >= 0 {
					beforeUpdates[i].SoundSet122 = unsafe.Pointer(sounds[ref])
				}
				if ref := want.Enemies[i]; ref >= 0 {
					beforeUpdates[i].CurrentEnemy = enemies[ref]
				}
			}
			compare := func(got, want unsafe.Pointer, n uintptr) {
				t.Helper()
				if !bytes.Equal(unsafe.Slice((*byte)(got), int(n)), unsafe.Slice((*byte)(want), int(n))) {
					t.Fatal("fight/run changed an unmodeled whole native record byte")
				}
			}
			compare(unit.CObj(), unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit))
			for i := 0; i < 2; i++ {
				compare(unsafe.Pointer(updates[i]), unsafe.Pointer(&beforeUpdates[i]), unsafe.Sizeof(beforeUpdates[i]))
				compare(unsafe.Pointer(sounds[i]), unsafe.Pointer(&beforeSounds[i]), unsafe.Sizeof(beforeSounds[i]))
				compare(enemies[i].CObj(), unsafe.Pointer(&beforeEnemies[i]), unsafe.Sizeof(beforeEnemies[i]))
			}
			for _, check := range checks {
				check()
			}
		})
	}
}
