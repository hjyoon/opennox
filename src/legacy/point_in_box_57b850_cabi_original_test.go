package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestElevatorPointInBoxCEntryOriginal(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/elevator_shaft_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Contexts int `json:"contexts"`
		Rows     []struct {
			Mode  string `json:"mode"`
			Input struct {
				Kind    string    `json:"kind"`
				Name    string    `json:"name"`
				Source  [2]uint32 `json:"source"`
				Target  [2]uint32 `json:"target"`
				Size    [2]uint32 `json:"size"`
				Corners [8]uint32 `json:"corners"`
			} `json:"input"`
			Return int `json:"return_int"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Contexts != 1061 {
		t.Fatal("incomplete original lift/point fixture")
	}
	selected := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup" || row.Input.Kind != "point" {
			continue
		}
		selected++
		t.Run(row.Input.Name, func(t *testing.T) {
			frame, free := alloc.Make([]byte(nil), 192)
			t.Cleanup(free)
			for i := range frame {
				frame[i] = 0xa5
			}
			rigidNativeAddress(t, unsafe.Pointer(&frame[32]))
			source := (*types.Pointf)(unsafe.Pointer(&frame[16]))
			shape := (*server.Shape)(unsafe.Pointer(&frame[32]))
			target := (*types.Pointf)(unsafe.Pointer(&frame[128]))
			*source = rigidNativePoint(row.Input.Source)
			*target = rigidNativePoint(row.Input.Target)
			words := []uint32{3, 0xa5a5a5a5, 0xa5a5a5a5, row.Input.Size[0], row.Input.Size[1]}
			words = append(words, row.Input.Corners[:]...)
			for i, w := range words {
				binary.LittleEndian.PutUint32(frame[32+4*i:], w)
			}
			before := bytes.Clone(frame)
			if got := pointInBoxCABI57B850(source, shape, target); got != (row.Return != 0) || !bytes.Equal(frame, before) {
				t.Fatalf("actual C entry=%t original=%d bytes preserved=%t", got, row.Return, bytes.Equal(frame, before))
			}
		})
	}
	if selected != 476 {
		t.Fatalf("original point conditions=%d", selected)
	}
}
