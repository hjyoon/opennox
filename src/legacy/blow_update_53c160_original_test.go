package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"image"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

type blowOriginalState53C160 struct {
	Objects    [3][7]uint32 `json:"objects"`
	Directions [3]uint16    `json:"directions"`
}
type blowOriginalInput53C160 struct {
	Name             string `json:"name"`
	Entry            string `json:"entry"`
	Candidates       []int  `json:"candidates"`
	CanInteract      bool   `json:"can_interact"`
	ScanMutation     string `json:"scan_mutation"`
	InteractMutation string `json:"interact_mutation"`
	blowOriginalState53C160
}
type blowOriginalEvent53C160 struct {
	Service   string   `json:"service"`
	Index     int32    `json:"index,omitempty"`
	Candidate int      `json:"candidate,omitempty"`
	Rect      []uint32 `json:"rect_bits32,omitempty"`
	blowOriginalState53C160
}
type blowOriginalFixture53C160 struct {
	OriginalSHA string `json:"original_sha256"`
	Contexts    int    `json:"contexts"`
	Executions  int    `json:"executions"`
	Tables      []struct {
		Address string `json:"address"`
		Hex     string `json:"hex"`
	} `json:"tables"`
	Raw []struct {
		Address string `json:"address"`
		Bits    uint32 `json:"bits"`
	} `json:"raw_table_words"`
	Rows []struct {
		Mode      string                    `json:"mode"`
		Input     blowOriginalInput53C160   `json:"input"`
		Output    blowOriginalState53C160   `json:"output"`
		Events    []blowOriginalEvent53C160 `json:"events"`
		Preserved bool                      `json:"nonoutput_bytes_preserved"`
	} `json:"rows"`
}

func TestBlowOriginalNativeC53C160(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/blow_update_53c160_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture blowOriginalFixture53C160
	if err = json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Contexts != 1292 || len(fixture.Rows) != 1292 || fixture.Executions != 2584 {
		t.Fatal("original provenance")
	}
	words := map[int64]uint32{}
	for _, table := range fixture.Tables {
		addr, e := strconv.ParseInt(table.Address, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		raw, e := hex.DecodeString(table.Hex)
		if e != nil {
			t.Fatal(e)
		}
		for i := 0; i < len(raw); i += 4 {
			words[addr+int64(i)] = binary.LittleEndian.Uint32(raw[i:])
		}
	}
	for _, raw := range fixture.Raw {
		addr, e := strconv.ParseInt(raw.Address, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		words[addr] = raw.Bits
	}
	read := func(addr int64) uint32 {
		v, ok := words[addr]
		if !ok {
			t.Fatalf("unsealed table address %#x", addr)
		}
		return v
	}
	var report *os.File
	if path := os.Getenv("OPENNOX_BLOW_STATE_REPORT"); path != "" {
		report, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer report.Close()
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			if !row.Preserved {
				t.Fatal("original frame canary")
			}
			input := row.Input
			var objs [3]*server.Object
			var frames [3][]byte
			var beforeFrames [3][]byte
			for n, v := range input.Objects {
				frames[n] = radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
				objs[n] = (*server.Object)(unsafe.Pointer(&frames[n][32]))
				*objs[n] = server.Object{ObjFlags: object.Flags(v[0]), ObjClass: object.Class(v[1]), PosVec: types.Ptf(math.Float32frombits(v[2]), math.Float32frombits(v[3])), ForceVec: types.Ptf(math.Float32frombits(v[4]), math.Float32frombits(v[5])), Mass: math.Float32frombits(v[6]), Direction1: server.Dir16(input.Directions[n])}
				beforeFrames[n] = append([]byte(nil), frames[n]...)
			}
			snapshot := func() blowOriginalState53C160 {
				var out blowOriginalState53C160
				for n, o := range objs {
					out.Objects[n] = [7]uint32{uint32(o.ObjFlags), uint32(o.ObjClass), math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), math.Float32bits(o.ForceVec.X), math.Float32bits(o.ForceVec.Y), math.Float32bits(o.Mass)}
					out.Directions[n] = uint16(o.Direction1)
				}
				return out
			}
			events := make([]blowOriginalEvent53C160, 0)
			mutate := func(kind string) {
				switch kind {
				case "":
				case "source-position":
					objs[0].PosVec = types.Ptf(-300.1, 987.3)
				case "candidate-position":
					objs[1].PosVec = types.Ptf(-1e10, 5000.3)
				case "mass-zero":
					objs[1].Mass = 0
				case "mass-negative":
					objs[1].Mass = -2.3
				case "force":
					objs[1].ForceVec = types.Ptf(123.1, -1e6)
				case "source-disabled":
					objs[0].ObjFlags = 0
				case "candidate-destroyed":
					objs[1].ObjFlags = 0x20
				case "candidate-immobile":
					objs[1].ObjClass = 0x400000
				case "wake-second":
					objs[2].ObjFlags = 0
					objs[2].ObjClass = 0
					objs[2].PosVec = types.Ptf(120.1, 201.3)
				default:
					const prefix = "direction-"
					if len(kind) < len(prefix) || kind[:len(prefix)] != prefix {
						t.Fatal("unknown mutation", kind)
					}
					d, e := strconv.ParseUint(kind[len(prefix):], 10, 16)
					if e != nil {
						t.Fatal(e)
					}
					objs[0].Direction1 = server.Dir16(d)
				}
			}
			indexed := func(direction int16) image.Point {
				events = append(events, blowOriginalEvent53C160{Service: "direction", Index: int32(direction), blowOriginalState53C160: snapshot()})
				classify := func(raw uint32) int {
					v := int32(raw)
					threshold := int32(read(0x5bf2cc))
					if v > threshold {
						return 1
					}
					if v >= -threshold {
						return 0
					}
					return -1
				}
				return image.Pt(classify(read(0x5b5e58+int64(direction)*8)), classify(read(0x5b5e5c+int64(direction)*8)))
			}
			floatWord := func(index int32) float32 {
				events = append(events, blowOriginalEvent53C160{Service: "float-word", Index: index, blowOriginalState53C160: snapshot()})
				return math.Float32frombits(read(0x5b6658 + int64(index)*4))
			}
			runtime := server.BlowUpdateRuntime53C160{
				IndexedDirection: indexed,
				EachInRect: func(rect types.Rectf, fn func(*server.Object) bool) {
					events = append(events, blowOriginalEvent53C160{Service: "scan", Rect: []uint32{math.Float32bits(rect.Min.X), math.Float32bits(rect.Min.Y), math.Float32bits(rect.Max.X), math.Float32bits(rect.Max.Y)}, blowOriginalState53C160: snapshot()})
					mutate(input.ScanMutation)
					for _, n := range input.Candidates {
						if !fn(objs[n]) {
							t.Fatal("early scan stop")
						}
					}
				},
				CanInteract: func(source, candidate *server.Object) bool {
					if source != objs[0] {
						t.Fatal("source identity")
					}
					idx := -1
					for n, o := range objs {
						if candidate == o {
							idx = n
						}
					}
					if idx < 0 {
						t.Fatal("candidate identity")
					}
					events = append(events, blowOriginalEvent53C160{Service: "interact", Candidate: idx, blowOriginalState53C160: snapshot()})
					mutate(input.InteractMutation)
					return input.CanInteract
				},
				DirectionFloat: floatWord,
			}
			oldRuntime := blowUpdateRuntime53C160
			blowUpdateRuntime53C160 = func() server.BlowUpdateRuntime53C160 { return runtime }
			defer func() { blowUpdateRuntime53C160 = oldRuntime }()
			blowCProbe53C160(objs[1], objs[0], input.Entry == "candidate")
			output := snapshot()
			if report != nil {
				if e := json.NewEncoder(report).Encode(map[string]any{"name": input.Name, "output": output, "events": events, "output_matches": reflect.DeepEqual(output, row.Output), "events_match": reflect.DeepEqual(events, row.Events)}); e != nil {
					t.Fatal(e)
				}
			}
			if !reflect.DeepEqual(output, row.Output) {
				t.Errorf("native output differs from unchanged original: got%v want%v", output, row.Output)
			}
			for n, originalFrame := range beforeFrames {
				want := append([]byte(nil), originalFrame...)
				offsets := []uintptr{unsafe.Offsetof(server.Object{}.ObjFlags), unsafe.Offsetof(server.Object{}.ObjClass), unsafe.Offsetof(server.Object{}.PosVec), unsafe.Offsetof(server.Object{}.PosVec) + 4, unsafe.Offsetof(server.Object{}.ForceVec), unsafe.Offsetof(server.Object{}.ForceVec) + 4, unsafe.Offsetof(server.Object{}.Mass)}
				for k, off := range offsets {
					binary.LittleEndian.PutUint32(want[32+int(off):], row.Output.Objects[n][k])
				}
				binary.LittleEndian.PutUint16(want[32+int(unsafe.Offsetof(server.Object{}.Direction1)):], row.Output.Directions[n])
				if !bytes.Equal(want, frames[n]) {
					t.Fatalf("native frame/canary %d differs from original outputs", n)
				}
			}
			if !reflect.DeepEqual(events, row.Events) {
				t.Errorf("direction/world/read trace differs: got%d events want%d", len(events), len(row.Events))
			}
		})
	}
}
