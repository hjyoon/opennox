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
	"github.com/opennox/opennox/v1/common/sound"
)

type hurtSoundOriginalInput532800 struct {
	Name, Mutation  string
	Cls, Frame, FPS uint32
	Update          int
	Deadlines       [2]uint32
	SoundSets       [2]int    `json:"sound_sets"`
	SoundIDs        [2]uint32 `json:"sound_ids"`
	RandomReturn    uint32    `json:"random_return"`
}

type hurtSoundOriginalState532800 struct {
	Cls, Frame, FPS uint32
	Update          int
	Deadlines       [2]uint32
	SoundSets       [2]int    `json:"sound_sets"`
	SoundIDs        [2]uint32 `json:"sound_ids"`
}

type hurtSoundOriginalEvent532800 struct {
	Kind  string
	Args  []uint32
	State hurtSoundOriginalState532800
}

func TestMonsterHurtSoundOriginalNative532800(t *testing.T) {
	file := os.Getenv("NOX_MONSTER_HURT_SOUND_ORIGINAL_FIXTURE")
	if file == "" {
		file = "testdata/monster_hurt_sound_native_original_532800.json.gz"
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
			Input  hurtSoundOriginalInput532800
			Result struct {
				Events []hurtSoundOriginalEvent532800
				State  hurtSoundOriginalState532800
			}
		}
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Rows) != 4000 || fixture.ControlWord != 0xe3f && fixture.ControlWord != 0x23f {
		t.Fatal("unsealed HurtSound original reference")
	}
	t.Logf("ORIGINAL HURT SOUND PROFILE: control=%#x primary contexts=%d; unchanged root/getter, declared RNG/audio, read-only frame/FPS", fixture.ControlWord, len(fixture.Rows))
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			c := row.Input
			unit, unitCheck := retreatOriginalAlloc545440[Object](t)
			var updates [2]*MonsterUpdateData
			var sounds [2]*[32]uint32
			checks := []func(){unitCheck}
			for i := 0; i < 2; i++ {
				var check func()
				updates[i], check = retreatOriginalAlloc545440[MonsterUpdateData](t)
				checks = append(checks, check)
				sounds[i], check = retreatOriginalAlloc545440[[32]uint32](t)
				checks = append(checks, check)
			}
			for i := 0; i < 2; i++ {
				updates[i].Field133, updates[i].Field137 = c.Deadlines[i], 0x12345678
				if ref := c.SoundSets[i]; ref >= 0 {
					updates[i].SoundSet122 = unsafe.Pointer(sounds[ref])
				}
				sounds[i][2] = c.SoundIDs[i]
			}
			unit.ObjClass, unit.UpdateData = object.Class(c.Cls), unsafe.Pointer(updates[c.Update])
			unit.ObjSubClass, unit.NetCode = 0xdead, 0xc0dec0de
			unit.PosVec, unit.ForceVec = types.Ptf(100, 200), types.Ptf(0.75, -0.5)
			unit.Direction1, unit.Direction2 = 0xabcd, 0xbabe
			frame, fps := c.Frame, c.FPS
			state := func() hurtSoundOriginalState532800 {
				v := hurtSoundOriginalState532800{Cls: uint32(unit.ObjClass), Frame: frame, FPS: fps, Update: -1}
				for i, update := range updates {
					if unit.UpdateData == unsafe.Pointer(update) {
						v.Update = i
					}
					v.Deadlines[i], v.SoundSets[i], v.SoundIDs[i] = update.Field133, -1, sounds[i][2]
					for j, set := range sounds {
						if update.SoundSet122 == unsafe.Pointer(set) {
							v.SoundSets[i] = j
						}
					}
				}
				return v
			}
			initial := row.Result.State
			if len(row.Result.Events) != 0 {
				initial = row.Result.Events[0].State
			}
			if state() != initial {
				t.Fatalf("reference input setup differs: got=%+v original=%+v", state(), initial)
			}
			beforeUnit := *unit
			beforeUpdates := [2]MonsterUpdateData{*updates[0], *updates[1]}
			beforeSounds := [2][32]uint32{*sounds[0], *sounds[1]}
			if unsafe.Sizeof(uintptr(0)) == 8 {
				for _, p := range []unsafe.Pointer{unit.CObj(), unsafe.Pointer(updates[0]), unsafe.Pointer(updates[1]), unsafe.Pointer(sounds[0]), unsafe.Pointer(sounds[1])} {
					if uintptr(p) <= 0xffffffff {
						t.Fatal("native HurtSound records must exceed4GiB")
					}
				}
			}
			mutate := func() {
				mode, cached, live := c.Mutation, c.Update, state().Update
				if mode == "frame" || mode == "frame-update" {
					frame ^= 0x80000011
				}
				if mode == "fps" {
					fps ^= 0x80000007
				}
				if mode == "update" || mode == "frame-update" || mode == "live-set" || mode == "clear-live" {
					live = 1 - cached
					unit.UpdateData = unsafe.Pointer(updates[live])
				}
				if mode == "cached-set" || mode == "live-set" {
					owner := cached
					if mode == "live-set" {
						owner = live
					}
					ref := 0
					if updates[owner].SoundSet122 == unsafe.Pointer(sounds[0]) {
						ref = 1
					}
					updates[owner].SoundSet122 = unsafe.Pointer(sounds[ref])
				}
				if mode == "clear-cached" {
					updates[cached].SoundSet122 = nil
				}
				if mode == "clear-live" {
					updates[live].SoundSet122 = nil
				}
				if mode == "class-remove" {
					unit.ObjClass &^= object.ClassMonster
				}
				if mode == "deadlines" {
					updates[0].Field133, updates[1].Field133 = 0x11223344, 0xaabbccdd
				}
				if mode == "ids" {
					sounds[0][2], sounds[1][2] = 0, 777
				}
			}
			got := []hurtSoundOriginalEvent532800{}
			observe := func(kind string, args ...uint32) {
				got = append(got, hurtSoundOriginalEvent532800{Kind: kind, Args: args, State: state()})
			}
			boundsMatch, fault := true, false
			func() {
				defer func() {
					if recover() != nil {
						fault = true
					}
				}()
				monsterHurtSound532800(unit, monsterHurtSoundHooks532800{
					frame:    func() uint32 { observe("frame", frame); return frame },
					tickRate: func() uint32 { observe("fps", fps); return fps },
					random: func(minimum, maximum int) int {
						// Original RNG arguments are signed DWORDs. Equality of
						// their low bits alone must not hide a widened Go int.
						boundsMatch = boundsMatch && minimum == int(int32(uint32(minimum))) && maximum == int(int32(uint32(maximum)))
						observe("random", uint32(minimum), uint32(maximum))
						mutate()
						return int(int32(c.RandomReturn))
					},
					play: func(id sound.ID, owner *Object) {
						if owner != unit {
							t.Fatal("native HurtSound lost its original owner")
						}
						observe("audio", uint32(id), 0, 0, 0)
					},
				})
			}()
			if fault || !boundsMatch || !reflect.DeepEqual(got, row.Result.Events) || state() != row.Result.State {
				t.Fatalf("original HurtSound effects differ: fault=%t signed-bounds=%t submitted=%v original=%v state-match=%t", fault, boundsMatch, got, row.Result.Events, state() == row.Result.State)
			}
			want := row.Result.State
			beforeUnit.ObjClass, beforeUnit.UpdateData = object.Class(want.Cls), unsafe.Pointer(updates[want.Update])
			for i := 0; i < 2; i++ {
				beforeUpdates[i].Field133 = want.Deadlines[i]
				beforeUpdates[i].SoundSet122 = nil
				if ref := want.SoundSets[i]; ref >= 0 {
					beforeUpdates[i].SoundSet122 = unsafe.Pointer(sounds[ref])
				}
				beforeSounds[i][2] = want.SoundIDs[i]
			}
			compare := func(got, want unsafe.Pointer, n uintptr) {
				t.Helper()
				if !bytes.Equal(unsafe.Slice((*byte)(got), int(n)), unsafe.Slice((*byte)(want), int(n))) {
					t.Fatal("HurtSound changed an unmodeled whole native record byte")
				}
			}
			compare(unit.CObj(), unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit))
			for i := 0; i < 2; i++ {
				compare(unsafe.Pointer(updates[i]), unsafe.Pointer(&beforeUpdates[i]), unsafe.Sizeof(beforeUpdates[i]))
				compare(unsafe.Pointer(sounds[i]), unsafe.Pointer(&beforeSounds[i]), unsafe.Sizeof(beforeSounds[i]))
			}
			for _, check := range checks {
				check()
			}
		})
	}
}
