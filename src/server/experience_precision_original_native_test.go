package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type experiencePrecisionEventOriginal struct {
	Service    string `json:"service"`
	Experience uint32 `json:"experience_bits32"`
	Token      uint32 `json:"token"`
	Award      uint32 `json:"award_bits32"`
	Key        string `json:"key"`
	Path       string `json:"path"`
	Line       int    `json:"line"`
	Points     uint32 `json:"message_points"`
}

func TestExperiencePrecisionOriginalNative(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native 64-bit experience frame")
	}
	const path = "testdata/experience_precision_original.json"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Contexts int `json:"contexts"`
		Cases    []struct {
			Group      string                             `json:"group"`
			Name       string                             `json:"name"`
			Experience uint32                             `json:"experience_bits32"`
			Argument   uint32                             `json:"argument_bits32"`
			Expected   uint32                             `json:"expected_experience_bits32"`
			Return     *uint64                            `json:"expected_award_bits64"`
			Events     []experiencePrecisionEventOriginal `json:"services"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Contexts != 128 || len(fixture.Cases) != 128 {
		t.Fatal("incomplete original contexts")
	}
	if unsafe.Sizeof(Object{}) != 928 || unsafe.Offsetof(Object{}.Experience) != 32 || unsafe.Offsetof(Object{}.UpdateData) != 872 || unsafe.Sizeof(PlayerUpdateData{}) != 656 || unsafe.Offsetof(PlayerUpdateData{}.Player) != 336 || unsafe.Sizeof(Player{}) != 6160 || unsafe.Offsetof(Player{}.ProtUnitExperience) != 5908 {
		t.Fatal("native experience layouts changed")
	}
	for _, group := range []string{"direct", "scaled"} {
		t.Run(group, func(t *testing.T) {
			count := 0
			for _, c := range fixture.Cases {
				if c.Group != group {
					continue
				}
				count++
				t.Run(c.Name, func(t *testing.T) {
					unitBytes, freeUnit := alloc.Make([]byte(nil), 1056)
					defer freeUnit()
					updateBytes, freeUpdate := alloc.Make([]byte(nil), 784)
					defer freeUpdate()
					playerBytes, freePlayer := alloc.Make([]byte(nil), 6288)
					defer freePlayer()
					for _, frame := range [][]byte{unitBytes, updateBytes, playerBytes} {
						for i := range frame {
							frame[i] = 0xa5
						}
						if uintptr(unsafe.Pointer(&frame[64])) <= 0xffffffff {
							t.Fatal("native frame below4GiB")
						}
					}
					unit := (*Object)(unsafe.Pointer(&unitBytes[64]))
					update := (*PlayerUpdateData)(unsafe.Pointer(&updateBytes[64]))
					player := (*Player)(unsafe.Pointer(&playerBytes[64]))
					unit.Experience = math.Float32frombits(c.Experience)
					unit.UpdateData = unsafe.Pointer(update)
					update.Player = player
					player.ProtUnitExperience = 0x89abcdef
					beforeUnit := append([]byte(nil), unitBytes...)
					beforeUpdate := append([]byte(nil), updateBytes...)
					beforePlayer := append([]byte(nil), playerBytes...)
					events := make([]experiencePrecisionEventOriginal, 0, len(c.Events))
					appendEvent := func(e experiencePrecisionEventOriginal) {
						e.Experience = math.Float32bits(unit.Experience)
						events = append(events, e)
					}
					checkUnit := func(o *Object) {
						if o != unit {
							t.Fatal("object-bearing service pointer truncated or changed")
						}
					}
					protect := func(token uint32, award float32) {
						appendEvent(experiencePrecisionEventOriginal{Service: "protect", Token: token, Award: math.Float32bits(award)})
					}
					report := func(o *Object) { checkUnit(o); appendEvent(experiencePrecisionEventOriginal{Service: "report"}) }
					sync := func(o *Object) { checkUnit(o); appendEvent(experiencePrecisionEventOriginal{Service: "sync"}) }
					var result float64
					if group == "direct" {
						directExperienceGrantNative4EF3A0(unit, math.Float32frombits(c.Argument), directExperienceGrantNativeDeps4EF3A0{
							protectExperience: protect, reportExperience: report, syncLevel: sync,
							loadString: func(key, path string, line int) string {
								appendEvent(experiencePrecisionEventOriginal{Service: "string", Key: key, Path: path, Line: line})
								return "original-declared-message"
							},
							sendLineMessage: func(o *Object, message string, points uint32) {
								checkUnit(o)
								if message != "original-declared-message" {
									t.Fatal("message changed")
								}
								appendEvent(experiencePrecisionEventOriginal{Service: "line", Points: points})
							},
						})
						if c.Return != nil {
							t.Fatal("unexpected direct return fixture")
						}
					} else {
						result = unitGiveXPNative4EF270(unit, math.Float32frombits(c.Argument), unitGiveXPNativeDeps4EF270{protectExperience: protect, reportExperience: report, syncLevel: sync})
						if c.Return == nil {
							t.Fatal("missing scaled return fixture")
						}
					}
					for i, b := range unitBytes {
						if (i < 96 || i >= 100) && b != beforeUnit[i] {
							t.Fatalf("unit nonoutput/canary byte changed at%d", i)
						}
					}
					if !bytes.Equal(updateBytes, beforeUpdate) || !bytes.Equal(playerBytes, beforePlayer) {
						t.Fatal("update/player frame or canary changed")
					}
					got := binary.LittleEndian.Uint32(unitBytes[96:])
					same := len(events) == len(c.Events)
					if same {
						for i := range events {
							if !reflect.DeepEqual(events[i], c.Events[i]) {
								same = false
							}
						}
					}
					if got != c.Expected || (c.Return != nil && math.Float64bits(result) != *c.Return) || !same {
						t.Fatalf("actual native experience=%08x return=%016x services=%+v want original=%08x return=%v services=%+v", got, math.Float64bits(result), events, c.Expected, c.Return, c.Events)
					}
				})
			}
			if count != 64 {
				t.Fatal("incomplete group", count)
			}
		})
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(raw, after) {
		t.Fatal("original fixture changed", err)
	}
}
