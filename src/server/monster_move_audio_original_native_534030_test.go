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

type moveAudioOriginalInput534030 struct {
	Name                                     string
	Cls, Subclass, Net, Frame, Frames, Delay uint32
	Update, Definition                       int
	SoundSet                                 int `json:"sound_set"`
	Animation                                [2]uint8
	Steps                                    [2]uint32
	SoundIDs                                 [2]uint32 `json:"sound_ids"`
}

type moveAudioOriginalState534030 struct {
	Cls, Subclass, Net   uint32
	Update               int
	Definitions          [2]int
	SoundSets            [2]int `json:"sound_sets"`
	Animation            [2][2]uint8
	Steps                [2][2]uint32
	SoundIDs             [2]uint32 `json:"sound_ids"`
	Frame, Frames, Delay uint32
}

type moveAudioOriginalEvent534030 struct {
	Kind  string
	Args  []uint32
	State moveAudioOriginalState534030
}

func TestMonsterMoveAudioOriginalNative534030(t *testing.T) {
	file := os.Getenv("NOX_MONSTER_MOVE_AUDIO_ORIGINAL_FIXTURE")
	if file == "" {
		file = "testdata/monster_move_audio_native_original_534030.json.gz"
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
			Input  moveAudioOriginalInput534030
			Result struct {
				Events []moveAudioOriginalEvent534030
				State  moveAudioOriginalState534030
			}
		}
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 5876 || fixture.ControlWord != 0xe3f && fixture.ControlWord != 0x23f {
		t.Fatal("unsealed MoveAudio original reference")
	}
	t.Logf("ORIGINAL MOVE AUDIO PROFILE: control=%#x primary contexts=%d; only AudioEvent declared, getters execute unchanged", fixture.ControlWord, len(fixture.Rows))
	// The normal outside-audio-loop path records every actual submitted event,
	// including sound zero, before later descriptor/device filtering occurs.
	var s Server
	s.Types.playerAnimFrames = make([][2]int, 5)
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			c := row.Input
			unit, unitCheck := retreatOriginalAlloc545440[Object](t)
			var updates [2]*MonsterUpdateData
			var defs [2]*MonsterDef
			var sounds [2]*[32]uint32
			checks := []func(){unitCheck}
			for i := 0; i < 2; i++ {
				var check func()
				updates[i], check = retreatOriginalAlloc545440[MonsterUpdateData](t)
				checks = append(checks, check)
				defs[i], check = retreatOriginalAlloc545440[MonsterDef](t)
				checks = append(checks, check)
				sounds[i], check = retreatOriginalAlloc545440[[32]uint32](t)
				checks = append(checks, check)
				sounds[i][18] = c.SoundIDs[i]
				defs[i].MoveSndFrameA100, defs[i].MoveSndFrameB104 = 2, 255
			}
			// Allocate every target before wiring either native record.
			for i := 0; i < 2; i++ {
				updates[i].MonsterDef = defs[1-c.Definition]
				if i == c.Update {
					updates[i].MonsterDef = defs[c.Definition]
				}
				if c.SoundSet >= 0 {
					idx := 1 - c.SoundSet
					if i == c.Update {
						idx = c.SoundSet
					}
					updates[i].SoundSet122 = unsafe.Pointer(sounds[idx])
				}
				updates[i].Field120_1, updates[i].Field120_2 = 255, 1
				if i == c.Update {
					updates[i].Field120_1, updates[i].Field120_2 = c.Animation[0], c.Animation[1]
				}
				updates[i].Field137, updates[i].Field74 = 0x12345678, 74
			}
			defs[c.Definition].MoveSndFrameA100, defs[c.Definition].MoveSndFrameB104 = c.Steps[0], c.Steps[1]
			unit.ObjClass, unit.ObjSubClass, unit.NetCode = object.Class(c.Cls), object.SubClass(c.Subclass), c.Net
			unit.UpdateData = unsafe.Pointer(updates[c.Update])
			unit.PosVec, unit.ForceVec = types.Ptf(100, 200), types.Ptf(0.75, -0.5)
			unit.Direction1, unit.Direction2 = 0xabcd, 0xbabe
			s.SetFrame(c.Frame)
			s.Types.playerAnimFrames[4] = [2]int{int(c.Frames), int(c.Delay)}
			s.Audio.delayedObj = nil
			s.Audio.inAudio = false
			if unsafe.Sizeof(uintptr(0)) == 8 {
				for _, p := range []unsafe.Pointer{unit.CObj(), unsafe.Pointer(updates[c.Update]), unsafe.Pointer(defs[c.Definition]), unsafe.Pointer(sounds[0])} {
					if uintptr(p) <= 0xffffffff {
						t.Fatal("native MoveAudio records must exceed4GiB")
					}
				}
			}
			state := func() moveAudioOriginalState534030 {
				v := moveAudioOriginalState534030{Cls: uint32(unit.ObjClass), Subclass: uint32(unit.ObjSubClass), Net: unit.NetCode, Update: -1, Frame: s.Frame(), Frames: uint32(s.Types.playerAnimFrames[4][0]), Delay: uint32(s.Types.playerAnimFrames[4][1])}
				for i, u := range updates {
					if unit.UpdateData == unsafe.Pointer(u) {
						v.Update = i
					}
					v.Definitions[i], v.SoundSets[i] = -1, -1
					for j := 0; j < 2; j++ {
						if u.MonsterDef == defs[j] {
							v.Definitions[i] = j
						}
						if u.SoundSet122 == unsafe.Pointer(sounds[j]) {
							v.SoundSets[i] = j
						}
					}
					v.Animation[i] = [2]uint8{u.Field120_1, u.Field120_2}
					v.Steps[i] = [2]uint32{defs[i].MoveSndFrameA100, defs[i].MoveSndFrameB104}
					v.SoundIDs[i] = sounds[i][18]
				}
				return v
			}
			if state() != row.Result.State {
				t.Fatalf("reference input setup differs: got=%+v original=%+v", state(), row.Result.State)
			}
			beforeUnit := *unit
			var beforeUpdates [2]MonsterUpdateData
			var beforeDefs [2]MonsterDef
			var beforeSounds [2][32]uint32
			for i := 0; i < 2; i++ {
				beforeUpdates[i], beforeDefs[i], beforeSounds[i] = *updates[i], *defs[i], *sounds[i]
			}
			fault := false
			func() {
				defer func() {
					if recover() != nil {
						fault = true
					}
				}()
				s.monsterMoveAudio534030(unit)
			}()
			got := []moveAudioOriginalEvent534030{}
			for _, event := range s.Audio.delayedObj {
				if event.Obj != unit {
					t.Fatal("native audio lost its original owner")
				}
				got = append(got, moveAudioOriginalEvent534030{Kind: "audio", Args: []uint32{uint32(event.ID), 0, uint32(event.Kind), event.Code}, State: state()})
			}
			if fault || !reflect.DeepEqual(got, row.Result.Events) || state() != row.Result.State {
				t.Fatalf("original MoveAudio effects differ: fault=%t submitted=%v original=%v state-match=%t", fault, got, row.Result.Events, state() == row.Result.State)
			}
			compare := func(got, want unsafe.Pointer, n uintptr) {
				t.Helper()
				if !bytes.Equal(unsafe.Slice((*byte)(got), int(n)), unsafe.Slice((*byte)(want), int(n))) {
					t.Fatal("MoveAudio changed a whole native record")
				}
			}
			compare(unit.CObj(), unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit))
			for i := 0; i < 2; i++ {
				compare(unsafe.Pointer(updates[i]), unsafe.Pointer(&beforeUpdates[i]), unsafe.Sizeof(beforeUpdates[i]))
				compare(unsafe.Pointer(defs[i]), unsafe.Pointer(&beforeDefs[i]), unsafe.Sizeof(beforeDefs[i]))
				compare(unsafe.Pointer(sounds[i]), unsafe.Pointer(&beforeSounds[i]), unsafe.Sizeof(beforeSounds[i]))
			}
			for _, check := range checks {
				check()
			}
			clear(s.Audio.delayedObj)
		})
	}
}
