package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/monster_direction_original_533cc0_v1.json.gz
var monsterDirectionOriginal []byte

type monsterDirectionCase struct {
	Input struct {
		Name       string    `json:"name"`
		Position   [2]uint32 `json:"position"`
		Target     [2]uint32 `json:"target"`
		Velocity   [2]uint32 `json:"velocity"`
		Directions [2]uint16 `json:"directions"`
		Neighbor   uint32    `json:"neighbor"`
		Alias      string    `json:"alias"`
		NilUnit    bool      `json:"nil_unit"`
		NilPoint   bool      `json:"nil_point"`
	} `json:"input"`
	TargetBefore   *[2]uint32 `json:"target_before"`
	DirectionAfter uint16     `json:"direction_after"`
}

func TestMonsterDirectionNativeLayout(t *testing.T) {
	// C ends after ScriptPickup; Go appends two private native handles.
	sharedSize := unsafe.Offsetof(server.Object{}.ScriptPickup) + unsafe.Sizeof(server.Object{}.ScriptPickup)
	want := [9]uintptr{sharedSize, unsafe.Offsetof(server.Object{}.PosVec), unsafe.Offsetof(server.Object{}.PosVec) + 4,
		unsafe.Offsetof(server.Object{}.VelVec), unsafe.Offsetof(server.Object{}.VelVec) + 4,
		unsafe.Offsetof(server.Object{}.Direction1), unsafe.Offsetof(server.Object{}.Direction2), unsafe.Sizeof(types.Pointf{}), unsafe.Offsetof(types.Pointf{}.Y)}
	if got := monsterDirectionLayoutC(); got != want || got[0] > unsafe.Sizeof(server.Object{}) {
		t.Fatalf("actual C/native Go owner layout: got %v want %v", got, want)
	}
}

func TestMonsterDirectionNativeOriginal(t *testing.T) {
	var rows []monsterDirectionCase
	npcVoiceOriginalDecode(t, monsterDirectionOriginal, "36649c07124d17176f54c8e9ca1223f5ecf0d0d40238099fa32b5e2f51c5b0da", "e269a0ca09e7b2a0209ccad0f6954db41a6e38b9c4a2a126b8f547d48b70a712", 623, &rows)
	layout := monsterDirectionLayoutC()
	for route, name := range []string{"public-C", "typed-C-callback", "production-Go-provider"} {
		for _, placement := range []int{0, 64, 256} {
			t.Run(fmt.Sprintf("%s/placement%d", name, placement), func(t *testing.T) {
				raw, free := alloc.MakeHigh([]byte{}, 0x10000)
				t.Cleanup(free)
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
					t.Fatal("native foreign owner must be above 4 GiB")
				}
				objectOffset := 0x1000 + placement
				separateOffset := 0x3000 + placement
				for _, c := range rows {
					t.Run(c.Input.Name, func(t *testing.T) {
						pointOffset := separateOffset
						for i := range raw {
							raw[i] = byte(i*31 + 19)
						}
						put32 := func(offset int, v uint32) { binary.LittleEndian.PutUint32(raw[offset:offset+4], v) }
						put32(objectOffset+int(layout[1]), c.Input.Position[0])
						put32(objectOffset+int(layout[2]), c.Input.Position[1])
						put32(objectOffset+int(layout[3]), c.Input.Velocity[0])
						put32(objectOffset+int(layout[4]), c.Input.Velocity[1])
						binary.LittleEndian.PutUint16(raw[objectOffset+int(layout[5]):], c.Input.Directions[0])
						binary.LittleEndian.PutUint16(raw[objectOffset+int(layout[6]):], c.Input.Directions[1])
						put32(objectOffset+int(layout[6])+2, c.Input.Neighbor)
						put32(pointOffset, c.Input.Target[0])
						put32(pointOffset+4, c.Input.Target[1])
						switch c.Input.Alias {
						case "position":
							pointOffset = objectOffset + int(layout[1])
						case "velocity":
							pointOffset = objectOffset + int(layout[3])
						case "direction-word":
							pointOffset = objectOffset + int(layout[5])
						}
						unit := (*server.Object)(unsafe.Pointer(&raw[objectOffset]))
						point := (*types.Pointf)(unsafe.Pointer(&raw[pointOffset]))
						if c.Input.NilUnit {
							unit = nil
						}
						if c.Input.NilPoint {
							point = nil
						} else if got := [2]uint32{math.Float32bits(point.X), math.Float32bits(point.Y)}; c.TargetBefore == nil || got != *c.TargetBefore {
							t.Fatalf("native alias input: got %v want %v", got, c.TargetBefore)
						}
						want := append([]byte(nil), raw...)
						binary.LittleEndian.PutUint16(want[objectOffset+int(layout[6]):], c.DirectionAfter)
						monsterDirectionActualCall(route, unit, point)
						if !bytes.Equal(raw, want) {
							for i := range raw {
								if raw[i] != want[i] {
									t.Fatalf("native arena byte %d: got %02x want %02x; direction %d want %d", i, raw[i], want[i], (*server.Object)(unsafe.Pointer(&raw[objectOffset])).Direction2, c.DirectionAfter)
								}
							}
						}
						runtime.KeepAlive(raw)
					})
				}
			})
		}
	}
}
