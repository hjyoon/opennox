package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
	"io"
	"os"
	"runtime/debug"
	"testing"
	"unsafe"
)

type roamLifecycleRawContract545790 interface {
	MonsterActionRoamStartNative545790() *server.MonsterUpdateData
	MonsterActionRoamCancelNative5457C0() *server.MonsterUpdateData
}

type roamLifecycleOriginalRow545790 struct {
	Input struct {
		Name, Entry string
		Index       int8
		Seed        uint32
		UnitClass   uint32 `json:"unit_class"`
		NilUnit     bool   `json:"nil_unit"`
		UpdateKind  string `json:"update_kind"`
	}
	Output struct {
		ReturnedUpdatePointer bool `json:"returned_update_pointer"`
		Writes                []struct {
			PC     string
			Offset int `json:"update_offset"`
			Bytes  int
			Value  uint32
		}
		Fault []json.RawMessage
		Guard bool `json:"whole_heap_and_guarded_caller_preserved"`
	}
}

func TestMonsterRoamWaypointLifecycleOriginal545790(t *testing.T) {
	TestMonsterRoamWaypointLifecycleNativeLayout545790(t)
	raw, err := os.ReadFile("../server/testdata/monster_roam_lifecycle_original_545790.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "d01324ba2502d77649e7892c9be0572044a05f08edc3404cdc8e4ef02fde6552" {
		t.Fatal("frozen original gzip changed")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err = z.Close(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(decoded)) != "63613298e0debb443b96e078362b444bbcd411302f6042b242754c0c801f8680" {
		t.Fatal("frozen decoded original changed")
	}
	var data struct {
		Rows []roamLifecycleOriginalRow545790
	}
	if err = json.Unmarshal(decoded, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Rows) != 1321 {
		t.Fatal("original context count", len(data.Rows))
	}
	layout := roamLifecycleOriginalLayout545790()
	for _, mode := range []string{"C", "Go"} {
		t.Run(mode, func(t *testing.T) {
			for _, row := range data.Rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					if !row.Output.Guard {
						t.Fatal("original guards not certified")
					}
					buf, free := alloc.MakeHigh([]byte{}, 0x10000)
					defer free()
					for i := 0; i < len(buf); i += 4 {
						binary.LittleEndian.PutUint32(buf[i:], row.Input.Seed)
					}
					unit, freeUnit := alloc.NewHigh(server.Object{})
					defer freeUnit()
					update := (*server.MonsterUpdateData)(unsafe.Pointer(&buf[0x6000]))
					unit.ObjClass, unit.UpdateData = object.Class(row.Input.UnitClass), unsafe.Pointer(update)
					update.AIStackInd = row.Input.Index
					// Untouched upper pointer bits must survive every other store.
					if unsafe.Sizeof(uintptr(0)) == 8 {
						high := uint64(0xface123400000000)
						for i := range update.Waypoints {
							// These are opaque poison bits, including unaligned values.
							// Prepare their exact bytes without manufacturing a Go pointer.
							offset := 0x6000 + int(layout[2]) + i*8
							binary.LittleEndian.PutUint64(buf[offset:], high|uint64(row.Input.Seed))
						}
						for i := range update.AIStack {
							for j := range update.AIStack[i].Args {
								update.AIStack[i].Args[j] = uintptr(high | uint64(row.Input.Seed))
							}
						}
					}
					invalidNative := originalInvalidNativePointer32(t)
					switch row.Input.UpdateKind {
					case "mapped":
					case "null":
						unit.UpdateData = nil
					case "invalid":
						if invalidNative != nil {
							unit.UpdateData = invalidNative
						} else {
							unit.UpdateData = unsafe.Pointer(uintptr(0x41000000))
						}
					default:
						t.Fatal("unknown original pointer mode", row.Input.UpdateKind)
					}
					beforeUnit := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(unit)), int(unsafe.Sizeof(*unit))))
					want := bytes.Clone(buf)
					for _, write := range row.Output.Writes {
						if write.Bytes != 4 || write.Value != 0 {
							t.Fatal("unexpected original write", write)
						}
						var offset, size int
						if row.Input.Entry == "545790" {
							switch write.Offset {
							case 300:
								offset, size = int(layout[2]), int(unsafe.Sizeof(uintptr(0)))
							case 364:
								offset, size = int(layout[3]), 4
							default:
								t.Fatal("unexpected original start write", write)
							}
						} else {
							if write.Offset != 556+24*int(row.Input.Index) {
								t.Fatal("original signed-index address", write, row.Input.Index)
							}
							offset, size = int(layout[5])+int(layout[6])*int(row.Input.Index)+int(layout[7]), int(unsafe.Sizeof(uintptr(0)))
						}
						clear(want[0x6000+offset : 0x6000+offset+size])
					}
					arg := unit
					if row.Input.NilUnit {
						arg = nil
					}
					var result unsafe.Pointer
					var fault any
					oldFault := debug.SetPanicOnFault(true)
					defer debug.SetPanicOnFault(oldFault)
					func() {
						defer func() { fault = recover() }()
						entry := 0
						if row.Input.Entry == "5457C0" {
							entry = 1
						}
						if mode == "C" {
							result = roamLifecycleOriginalC545790(entry, arg)
							return
						}
						native, ok := any(arg).(roamLifecycleRawContract545790)
						if !ok {
							t.Fatal("raw native lifecycle contract is not implemented")
						}
						if entry == 0 {
							result = unsafe.Pointer(native.MonsterActionRoamStartNative545790())
						} else {
							result = unsafe.Pointer(native.MonsterActionRoamCancelNative5457C0())
						}
					}()
					if (fault != nil) != (len(row.Output.Fault) != 0) {
						t.Fatal("original fault boundary", fault, row.Output.Fault)
					}
					if row.Output.ReturnedUpdatePointer && result != unsafe.Pointer(update) {
						t.Fatal("whole native update pointer return", result, update)
					}
					if !bytes.Equal(beforeUnit, unsafe.Slice((*byte)(unsafe.Pointer(unit)), len(beforeUnit))) {
						t.Fatal("complete native unit changed")
					}
					if !bytes.Equal(want, buf) {
						t.Fatal("complete 64 KiB record/canary bytes differ from original write history")
					}
				})
			}
		})
	}
}
