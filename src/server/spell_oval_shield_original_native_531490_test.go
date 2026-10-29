package server

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
)

//go:embed testdata/spell_oval_shield_native_original_531490.json
var ovalShieldOriginalData531490 []byte

type ovalShieldOriginalState531490 struct {
	Target    int         `json:"target"`
	Level     uint32      `json:"level"`
	Expiry    uint32      `json:"expiry"`
	Position  [2]uint32   `json:"position"`
	FPS       uint32      `json:"fps"`
	Frame     uint32      `json:"frame"`
	Classes   []uint32    `json:"classes"`
	Flags     []uint32    `json:"flags"`
	Positions [][2]uint32 `json:"positions"`
}
type ovalShieldOriginalInput531490 struct {
	ovalShieldOriginalState531490
	Name           string `json:"name"`
	Kind           string `json:"kind"`
	TestResult     uint32 `json:"test_result"`
	PositionResult uint32 `json:"position_result"`
	Mutation       struct {
		Stage  string `json:"stage"`
		At     int    `json:"at"`
		Action string `json:"action"`
	} `json:"mutation"`
}
type ovalShieldOriginalEvent531490 struct {
	Kind  string                        `json:"kind"`
	Args  []int64                       `json:"args"`
	State ovalShieldOriginalState531490 `json:"state"`
}
type ovalShieldOriginalResult531490 struct {
	Ret    *uint32                         `json:"ret"`
	State  ovalShieldOriginalState531490   `json:"state"`
	Events []ovalShieldOriginalEvent531490 `json:"events"`
}

func cloneOvalShieldOriginalState531490(s ovalShieldOriginalState531490) ovalShieldOriginalState531490 {
	s.Classes = append([]uint32(nil), s.Classes...)
	s.Flags = append([]uint32(nil), s.Flags...)
	s.Positions = append([][2]uint32(nil), s.Positions...)
	return s
}

// The unchanged original executes every instruction of all three roots.
// Logical object identities keep its PE32 links distinct from native pointers;
// actual native public C entry composition is verified separately.
func TestSpellOvalShieldOriginalNative531490(t *testing.T) {
	digest := sha256.Sum256(ovalShieldOriginalData531490)
	if hex.EncodeToString(digest[:]) != "16c4b73fb7894b22e28a80f5bb87ef044dd15621a2f8d65deec3171ae9498ffd" {
		t.Fatal("sealed original fixture hash")
	}
	var fixture struct {
		OriginalSHA256 string `json:"original_sha256"`
		ControlWord    int    `json:"control_word"`
		Rows           []struct {
			Input  ovalShieldOriginalInput531490  `json:"input"`
			Result ovalShieldOriginalResult531490 `json:"result"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(ovalShieldOriginalData531490, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.ControlWord != 0xe3f || len(fixture.Rows) != 197 {
		t.Fatal("original identity/profile/contexts")
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			input := row.Input
			s := cloneOvalShieldOriginalState531490(input.ovalShieldOriginalState531490)
			events := make([]ovalShieldOriginalEvent531490, 0)
			calls := make(map[string]int)
			mutated := false
			event := func(kind string, args ...int64) {
				events = append(events, ovalShieldOriginalEvent531490{kind, args, cloneOvalShieldOriginalState531490(s)})
				at := calls[kind]
				calls[kind]++
				m := input.Mutation
				if mutated || m.Stage != kind || m.At != at {
					return
				}
				mutated = true
				switch m.Action {
				case "target-b":
					s.Target = 1
				case "target-monster":
					s.Target = 2
				case "target-nil":
					s.Target = -1
				case "level":
					s.Level = 0xfffffffe
				case "frame":
					s.Frame = 0xfffffff9
				case "fps":
					s.FPS = 0xffffffff
				case "target-flags":
					s.Flags[0] = 0x8020
				case "target-class":
					s.Classes[0] = 2
				case "record-position":
					s.Position = [2]uint32{0x4b800001, 0xcb800001}
				default:
					t.Fatalf("unknown declared mutation %q", m.Action)
				}
			}
			h := spellOvalShieldHooks531490[*ovalShieldOriginalState531490, int]{
				loadFPS:         func() uint32 { return s.FPS },
				loadLevel:       func(*ovalShieldOriginalState531490) uint32 { return s.Level },
				loadTarget:      func(*ovalShieldOriginalState531490) int { return s.Target + 1 },
				loadClass:       func(object int) uint8 { return uint8(s.Classes[object-1]) },
				loadFlags:       func(object int) uint32 { return s.Flags[object-1] },
				cancelOffensive: func(object int) { event("cancel", int64(object-1)) },
				applyBuff: func(object int, buff int32, duration int16, power int8) {
					event("apply", int64(object-1), int64(buff), int64(duration), int64(power))
				},
				loadFrame:  func() uint32 { return s.Frame },
				storeFrame: func(_ *ovalShieldOriginalState531490, frame uint32) { s.Expiry = frame },
				testBuff: func(object int, buff int32) int32 {
					event("test", int64(object-1), int64(buff))
					return int32(input.TestResult)
				},
				positionDelta: func(object int, _ *ovalShieldOriginalState531490) int32 {
					event("position", int64(object-1), int64(s.Position[0]), int64(s.Position[1]))
					return int32(input.PositionResult)
				},
				buffOff: func(object int, buff int32) { event("off", int64(object-1), int64(buff)) },
			}
			var ret *uint32
			switch input.Kind {
			case "create":
				value := uint32(spellOvalShieldCreate531490(&s, h))
				ret = &value
			case "update":
				value := uint32(spellOvalShieldUpdate5314F0(&s, h))
				ret = &value
			case "destroy":
				spellOvalShieldDestroy531560(&s, h)
			default:
				t.Fatalf("unknown root %q", input.Kind)
			}
			if input.Mutation.Stage != "" && !mutated {
				t.Fatal("declared mutation not reached")
			}
			got := ovalShieldOriginalResult531490{ret, cloneOvalShieldOriginalState531490(s), events}
			if !reflect.DeepEqual(got, row.Result) {
				actual, _ := json.Marshal(got)
				expected, _ := json.Marshal(row.Result)
				t.Fatalf("got %s\nwant %s", actual, expected)
			}
		})
	}
}
