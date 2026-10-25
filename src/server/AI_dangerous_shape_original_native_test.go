package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestAIPathDangerousShapeOriginal50B2C0Native(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native 64-bit AI dangerous shape frame")
	}
	const path = "testdata/AI_dangerous_shape_original.json"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Contexts int `json:"contexts"`
		Cases    []struct {
			Group    string   `json:"group"`
			Name     string   `json:"name"`
			Kind     uint32   `json:"kind"`
			Values   []uint32 `json:"values_bits32"`
			Expected []uint32 `json:"expected_expanded_shape_words32"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Contexts != 182 || len(fixture.Cases) != 182 || unsafe.Sizeof(Shape{}) != 52 {
		t.Fatal("incomplete original contexts or changed native shape")
	}
	for _, group := range []string{"stock_dangerous", "finite_circle", "finite_box", "kind_controls"} {
		t.Run(group, func(t *testing.T) {
			count := 0
			for _, c := range fixture.Cases {
				if c.Group != group {
					continue
				}
				count++
				t.Run(c.Name, func(t *testing.T) {
					if len(c.Expected) != 13 {
						t.Fatal("incomplete full shape words")
					}
					frame, free := alloc.Make([]byte(nil), 188)
					defer free()
					for i := range frame {
						frame[i] = 0xa5
					}
					if uintptr(unsafe.Pointer(&frame[64])) <= 0xffffffff {
						t.Fatal("native frame below4GiB")
					}
					binary.LittleEndian.PutUint32(frame[64:], c.Kind)
					offset := 12
					if c.Kind == 2 {
						offset = 4
					}
					for i, value := range c.Values {
						binary.LittleEndian.PutUint32(frame[64+offset+4*i:], value)
					}
					before := append([]byte(nil), frame...)
					shape := (*Shape)(unsafe.Pointer(&frame[64]))
					aiPathExpandDangerousShape50B2C0(shape)
					for i, b := range frame {
						if (i < 64 || i >= 116) && b != before[i] {
							t.Fatalf("native Z/nonoutput/canary byte changed at%d", i)
						}
					}
					got := make([]uint32, 13)
					same := true
					for i := range got {
						got[i] = binary.LittleEndian.Uint32(frame[64+4*i:])
						if got[i] != c.Expected[i] {
							same = false
						}
					}
					if !same {
						t.Fatalf("actual native AI dangerous shape=%08x want original=%08x", got, c.Expected)
					}
				})
			}
			want := 64
			if group == "stock_dangerous" {
				want = 50
			}
			if group == "kind_controls" {
				want = 4
			}
			if count != want {
				t.Fatal("incomplete group", count, want)
			}
		})
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(raw, after) {
		t.Fatal("original fixture changed", err)
	}
}
