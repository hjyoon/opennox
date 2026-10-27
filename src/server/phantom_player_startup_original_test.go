package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type phantomInputOriginal struct {
	Name              string    `json:"name"`
	TargetPosition    [2]uint32 `json:"target_position"`
	Anchor            [2]uint32 `json:"anchor"`
	SourcePosition    [2]uint32 `json:"source_position"`
	InitialNew        [2]uint32 `json:"initial_new"`
	TargetDirection   uint16    `json:"target_direction"`
	SourceDirection   uint16    `json:"source_direction"`
	InitialDirection2 uint16    `json:"initial_direction2"`
	SourceFlags       uint32    `json:"source_flags"`
	AliasSource       bool      `json:"alias_source"`
	ZeroObjects       bool      `json:"zero_objects"`
}

type phantomDeleteEventOriginal struct {
	Service          string    `json:"service"`
	Object           string    `json:"object"`
	SourceNew        [2]uint32 `json:"source_new"`
	SourceDirection2 uint16    `json:"source_direction2"`
}

type phantomRowOriginal struct {
	Mode             string                       `json:"mode"`
	Input            phantomInputOriginal         `json:"input"`
	SourceNew        [2]uint32                    `json:"source_new"`
	SourceDirection2 uint16                       `json:"source_direction2"`
	Events           []phantomDeleteEventOriginal `json:"events"`
	Preserved        bool                         `json:"nonoutput_bytes_preserved"`
}

func TestPhantomPlayerStartupOriginal(t *testing.T) {
	blob, err := os.ReadFile("testdata/phantom_player_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA        string               `json:"original_sha256"`
		Unchanged  bool                 `json:"original_unchanged"`
		CW         uint16               `json:"actual_startup_control_word"`
		Contexts   int                  `json:"contexts"`
		Executions int                  `json:"root_executions"`
		Faults     []json.RawMessage    `json:"memory_faults"`
		Rows       []phantomRowOriginal `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !f.Unchanged || f.CW != 0xe3f || f.Contexts != 494 || f.Executions != 988 || len(f.Rows) != 988 || len(f.Faults) != 0 {
		t.Fatal("incomplete original PhantomPlayer fixture")
	}
	count := 0
	for _, row := range f.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) {
			testPhantomOriginal(t, row, func(p *Object, del func(*Object)) {
				new(Server).PhantomPlayerUpdate53B860(p, PhantomPlayerUpdateRuntime53B860{DelayedDelete: del})
			})
		})
	}
	if count != 494 {
		t.Fatalf("startup conditions=%d", count)
	}
}

func testPhantomOriginal(t *testing.T, row phantomRowOriginal, call func(*Object, func(*Object))) {
	t.Helper()
	i := row.Input
	frame := func(size uintptr) []byte {
		p, free := alloc.Make([]byte(nil), int(size)+64)
		t.Cleanup(free)
		for n := range p {
			p[n] = 0xa5
		}
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(&p[32])) <= math.MaxUint32 {
			t.Fatal("native frame below 4 GiB")
		}
		return p
	}
	sourceFrame, targetFrame, updateFrame := frame(unsafe.Sizeof(Object{})), frame(unsafe.Sizeof(Object{})), frame(unsafe.Sizeof(PhantomPlayerUpdateData{}))
	source, target := (*Object)(unsafe.Pointer(&sourceFrame[32])), (*Object)(unsafe.Pointer(&targetFrame[32]))
	update := (*PhantomPlayerUpdateData)(unsafe.Pointer(&updateFrame[32]))
	*source = Object{PosVec: rigidPointOriginal(i.SourcePosition), NewPos: rigidPointOriginal(i.InitialNew), Direction1: Dir16(i.SourceDirection), Direction2: Dir16(i.InitialDirection2), ObjFlags: object.Flags(i.SourceFlags)}
	*target = Object{PosVec: rigidPointOriginal(i.TargetPosition), Direction1: Dir16(i.TargetDirection)}
	if !i.ZeroObjects {
		source.Field29 = 0x13572468
		target.Field29 = 0x87654321
	}
	actualTarget := target
	if i.AliasSource {
		actualTarget = source
		source.PosVec = target.PosVec
		source.Direction1 = target.Direction1
	}
	*update = PhantomPlayerUpdateData{Target: actualTarget, Position: rigidPointOriginal(i.Anchor)}
	source.UpdateData = unsafe.Pointer(update)
	wantSource, wantTarget, wantUpdate := bytes.Clone(sourceFrame), bytes.Clone(targetFrame), bytes.Clone(updateFrame)
	binary.LittleEndian.PutUint32(wantSource[32+unsafe.Offsetof(source.NewPos):], row.SourceNew[0])
	binary.LittleEndian.PutUint32(wantSource[36+unsafe.Offsetof(source.NewPos):], row.SourceNew[1])
	binary.LittleEndian.PutUint16(wantSource[32+unsafe.Offsetof(source.Direction2):], row.SourceDirection2)
	events := make([]phantomDeleteEventOriginal, 0)
	call(source, func(p *Object) {
		if p != source {
			t.Fatal("delete target changed")
		}
		events = append(events, phantomDeleteEventOriginal{Service: "delete", Object: "source", SourceNew: [2]uint32{math.Float32bits(source.NewPos.X), math.Float32bits(source.NewPos.Y)}, SourceDirection2: uint16(source.Direction2)})
	})
	if !row.Preserved || !reflect.DeepEqual(events, row.Events) || !bytes.Equal(sourceFrame, wantSource) || !bytes.Equal(targetFrame, wantTarget) || !bytes.Equal(updateFrame, wantUpdate) {
		t.Fatalf("native new=%08x/%08x direction=%04x events=%+v; original new=%08x/%08x direction=%04x events=%+v; frames=%t/%t/%t", math.Float32bits(source.NewPos.X), math.Float32bits(source.NewPos.Y), uint16(source.Direction2), events, row.SourceNew[0], row.SourceNew[1], row.SourceDirection2, row.Events, bytes.Equal(sourceFrame, wantSource), bytes.Equal(targetFrame, wantTarget), bytes.Equal(updateFrame, wantUpdate))
	}
}
