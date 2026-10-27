package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

func TestPendingCollisionRadialRootNativeMapOriginal(t *testing.T) {
	blob, err := os.ReadFile("testdata/pending_radial_push_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Contexts   int `json:"contexts"`
		Executions int `json:"root_executions"`
		Rows       []struct {
			Mode  string `json:"mode"`
			Input struct {
				Order []int `json:"order"`
			} `json:"input"`
			Objects [][7]uint32 `json:"objects"`
			Origin  [2]uint32   `json:"origin"`
			Queue   struct {
				Head   int         `json:"head"`
				Tail   int         `json:"tail"`
				Status []uint32    `json:"status"`
				Links  []int       `json:"links"`
				Visits [][2]uint32 `json:"visits"`
			} `json:"queue"`
			Events []struct {
				Service string   `json:"service"`
				Args    []uint32 `json:"args"`
			} `json:"events"`
		} `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.Contexts != 2 || f.Executions != 4 || len(f.Rows) != 4 {
		t.Fatal("original combined root metadata")
	}
	rows := map[int]int{}
	for n, row := range f.Rows {
		if row.Mode == "actual-startup" {
			rows[row.Input.Order[1]] = n
		}
	}
	if len(rows) != 2 {
		t.Fatal("original insertion orders")
	}
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse-%t", reverse), func(t *testing.T) {
			srv := server.New(nil, nil, strman.New())
			t.Cleanup(srv.Close)
			srv.Map.Init()
			t.Cleanup(srv.Map.Free)
			if srv.Walls.Init() == 0 {
				t.Fatal("native wall index")
			}
			t.Cleanup(srv.Walls.Free)
			oldServer := GetServer
			GetServer = func() Server { return &powderCService53C9A0{srv: srv} }
			t.Cleanup(func() { GetServer = oldServer })
			oldLookup, oldPrepare := pendingProbeLookup537610, pendingProbePrepare537610
			defer func() { pendingProbeLookup537610, pendingProbePrepare537610 = oldLookup, oldPrepare }()
			pendingProbeLookup537610 = func(string) uint32 { t.Fatal("ordinary class must not look up types"); return 0 }
			pendingProbePrepare537610 = func() { t.Fatal("ordinary class must not reset AI paths") }
			runtime := pendingCollisionRuntime537610
			machine := pendingProbeMachine537580(runtime.Fist, runtime.UndeadKiller)
			t.Cleanup(machine.close)
			identityFrame := radialCFrameOriginal(t, 256)
			var objects [2]*server.Object
			var frames [2][]byte
			for n := range objects {
				frame := radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
				frames[n] = frame
				o := (*server.Object)(unsafe.Pointer(&frame[32]))
				objects[n] = o
				*o = server.Object{ObjFlags: object.FlagActive, Mass: 1, PosVec: types.Ptf(305.1+float32(n)*5, 300.1), Update: unsafe.Pointer(&identityFrame[32]), Field115: 0xffffffff, Field37: 0x55667788, Field38: 0x11223344}
				o.NewPos = o.PosVec
				o.PrevPos = o.PosVec
				o.Shape.Kind = server.ShapeKindCircle
				o.Shape.Circle.R = 1
				o.Shape.Circle.R2 = 1
			}
			order := []int{0, 1}
			if reverse {
				order = []int{1, 0}
			}
			for _, n := range order {
				srv.Map.AddObjectToIndex(objects[n])
			}
			defer func() {
				for _, o := range objects {
					srv.Map.RemoveObjectFromIndex(o)
				}
			}()
			before := [][]byte{bytes.Clone(frames[0]), bytes.Clone(frames[1])}
			pointFrame := radialCFrameOriginal(t, unsafe.Sizeof(types.Pointf{}))
			point := (*types.Pointf)(unsafe.Pointer(&pointFrame[32]))
			*point = types.Ptf(300.1, 300.1)
			pointBefore := bytes.Clone(pointFrame)
			_, probeSize := radialPushCSize52E0E0()
			probeFrame := radialCFrameOriginal(t, probeSize)
			probe := unsafe.Pointer(&probeFrame[32])
			radialPushCProbeInit52E0E0(probe)
			radialPushCProbeOrigin52E0E0(probe, point)
			probeBefore := bytes.Clone(probeFrame)
			radialPushCRoot52E040(point, 50, 10, 100, nil, radialPushCProbePointer52E0E0(), uintptr(probe))
			offsets := radialPushCProbeOffsets52E0E0()
			readWord := func(offset uintptr) uintptr {
				if unsafe.Sizeof(uintptr(0)) == 8 {
					return uintptr(binary.LittleEndian.Uint64(probeFrame[32+offset:]))
				}
				return uintptr(binary.LittleEndian.Uint32(probeFrame[32+offset:]))
			}
			putWord := func(frame []byte, offset, value uintptr) {
				if unsafe.Sizeof(value) == 8 {
					binary.LittleEndian.PutUint64(frame[32+offset:], uint64(value))
				} else {
					binary.LittleEndian.PutUint32(frame[32+offset:], uint32(value))
				}
			}
			last := -1
			for n, o := range objects {
				if readWord(offsets[1]) == uintptr(unsafe.Pointer(o)) {
					last = n
				}
			}
			if last < 0 {
				t.Fatal("foreign callback candidate")
			}
			row := f.Rows[rows[last]]
			if machine.head() != objects[row.Queue.Head] || machine.tail() != uintptr(unsafe.Pointer(objects[row.Queue.Tail])) {
				t.Fatal("original activation queue order")
			}
			for n, o := range objects {
				for slot, value := range row.Queue.Visits[n] {
					binary.LittleEndian.PutUint32(before[n][32+unsafe.Offsetof(o.Field62)+uintptr(slot)*4:], value)
				}
				binary.LittleEndian.PutUint32(before[n][32+unsafe.Offsetof(o.ForceVec):], row.Objects[n][2])
				binary.LittleEndian.PutUint32(before[n][36+unsafe.Offsetof(o.ForceVec):], row.Objects[n][3])
				expected := bytes.Clone(before[n])
				binary.LittleEndian.PutUint32(expected[32+unsafe.Offsetof(o.Field115):], 0)
				binary.LittleEndian.PutUint32(expected[32+unsafe.Offsetof(o.Field116):], row.Queue.Status[n])
				if !bytes.Equal(frames[n], expected) {
					for offset := range frames[n] {
						if frames[n][offset] != expected[offset] {
							t.Fatalf("native pushed object %d byte %d actual=%02x original=%02x", n, offset, frames[n][offset], expected[offset])
						}
					}
				}
				next := uintptr(0)
				if row.Queue.Links[n] >= 0 {
					next = uintptr(unsafe.Pointer(objects[row.Queue.Links[n]]))
				}
				if machine.next(o) != next {
					t.Fatal("native queue link differs from original")
				}
			}
			binary.LittleEndian.PutUint32(pointBefore[32:], row.Origin[0])
			binary.LittleEndian.PutUint32(pointBefore[36:], row.Origin[1])
			if !bytes.Equal(pointFrame, pointBefore) {
				t.Fatal("borrowed point record or canary")
			}
			binary.LittleEndian.PutUint32(probeBefore[32+offsets[0]:], 2)
			putWord(probeBefore, offsets[1], uintptr(unsafe.Pointer(objects[last])))
			putWord(probeBefore, offsets[3], uintptr(probe))
			distance := uint32(0)
			for _, event := range row.Events {
				if event.Service == "before:1" {
					distance = event.Args[1]
				}
			}
			binary.LittleEndian.PutUint32(probeBefore[32+offsets[2]:], distance)
			if !bytes.Equal(probeFrame, probeBefore) {
				t.Fatal("C callback userdata, raw distance or canary")
			}
			for _, index := range row.Input.Order {
				if machine.pop() != objects[index] {
					t.Fatal("original drain order")
				}
			}
			if machine.empty() != 1 || machine.compat() != [2]uint32{} {
				t.Fatal("drained compatibility state")
			}
			for n := range frames {
				if !bytes.Equal(frames[n], before[n]) {
					t.Fatal("drained object force, status or canary")
				}
			}
			if math.Float32bits(point.X) != row.Origin[0] || math.Float32bits(point.Y) != row.Origin[1] {
				t.Fatal("original live origin")
			}
		})
	}
}
