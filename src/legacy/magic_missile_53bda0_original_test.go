package legacy

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

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

type magicMissileState53BDA0 struct {
	Objects    [4][10]uint32 `json:"objects"`
	Directions [4]uint16     `json:"directions"`
	Owners     [4]int64      `json:"owners"`
	Updates    [2][7]int64   `json:"updates"`
	UpdateRef  int           `json:"update_ref"`
	Frame      uint32        `json:"frame"`
	FPS        uint32        `json:"fps"`
}
type magicMissileEvent53BDA0 struct {
	Service string  `json:"service"`
	Args    []int64 `json:"args"`
	magicMissileState53BDA0
}
type magicMissileInput53BDA0 struct {
	Name          string `json:"name"`
	Mutation      string `json:"mutation"`
	SearchResult  int64  `json:"search_result"`
	CollideResult uint32 `json:"collide_result"`
	magicMissileState53BDA0
}

func TestMagicMissileOriginalNative53BDA0(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/magic_missile_53bda0_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		OriginalSHA string `json:"original_sha256"`
		Contexts    int    `json:"contexts"`
		Executions  int    `json:"root_executions"`
		Tables      []struct {
			Address string `json:"address"`
			Hex     string `json:"hex"`
		} `json:"tables"`
		Raw []struct {
			Address string `json:"address"`
			Bits    uint32 `json:"bits"`
		} `json:"raw_direction_words"`
		Rows []struct {
			Mode      string                    `json:"mode"`
			Input     magicMissileInput53BDA0   `json:"input"`
			Output    magicMissileState53BDA0   `json:"output"`
			Events    []magicMissileEvent53BDA0 `json:"events"`
			Result    uint32                    `json:"result"`
			Preserved bool                      `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err = json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.Contexts != 1009 || f.Executions != 2018 || len(f.Rows) != 2018 {
		t.Fatal("original provenance")
	}
	var table []byte
	if err = json.Unmarshal([]byte(`"`+f.Tables[0].Hex+`"`), new(string)); err != nil {
		t.Fatal(err)
	}
	table = make([]byte, len(f.Tables[0].Hex)/2)
	for n := range table {
		var v uint64
		v, err = strconv.ParseUint(f.Tables[0].Hex[n*2:n*2+2], 16, 8)
		if err != nil {
			t.Fatal(err)
		}
		table[n] = byte(v)
	}
	words := map[int32]uint32{}
	for n := 0; n < 512; n++ {
		words[int32(n)] = binary.LittleEndian.Uint32(table[n*4:])
	}
	for _, w := range f.Raw {
		a, e := strconv.ParseInt(w.Address, 0, 64)
		if e != nil {
			t.Fatal(e)
		}
		words[int32((a-0x5b6658)/4)] = w.Bits
	}
	for _, lane := range []string{"go-root", "actual-c"} {
		t.Run(lane, func(t *testing.T) {
			seen := 0
			for _, row := range f.Rows {
				if row.Mode != "actual-startup" {
					continue
				}
				seen++
				t.Run(row.Input.Name, func(t *testing.T) {
					if !row.Preserved {
						t.Fatal("original canary violation")
					}
					i := row.Input
					var objects [4]*server.Object
					var frames, before [4][]byte
					for n, w := range i.Objects {
						frames[n] = radialCFrameOriginal(t, unsafe.Sizeof(server.Object{}))
						objects[n] = (*server.Object)(unsafe.Pointer(&frames[n][32]))
						*objects[n] = server.Object{PosVec: types.Ptf(math.Float32frombits(w[0]), math.Float32frombits(w[1])), VelVec: types.Ptf(math.Float32frombits(w[2]), math.Float32frombits(w[3])), ForceVec: types.Ptf(math.Float32frombits(w[4]), math.Float32frombits(w[5])), Float28: math.Float32frombits(w[6]), Field32: w[7], SpeedCur: math.Float32frombits(w[8]), ObjFlags: object.Flags(w[9]), Direction2: server.Dir16(i.Directions[n])}
					}
					pointer := func(n int64) *server.Object {
						if n < 0 {
							return nil
						}
						return objects[n]
					}
					index := func(o *server.Object) int64 {
						if o == nil {
							return -1
						}
						for n, p := range objects {
							if o == p {
								return int64(n)
							}
						}
						t.Fatal("foreign native object pointer")
						return -2
					}
					for n, o := range objects {
						o.ObjOwner = pointer(i.Owners[n])
					}
					var updates [2]*server.MissileUpdateData
					var uframes, ubefore [2][]byte
					for n, w := range i.Updates {
						uframes[n] = radialCFrameOriginal(t, unsafe.Sizeof(server.MissileUpdateData{}))
						updates[n] = (*server.MissileUpdateData)(unsafe.Pointer(&uframes[n][32]))
						*updates[n] = server.MissileUpdateData{Owner: pointer(w[0]), Target: pointer(w[1]), Field_8: uint32(w[2]), SpellID: int32(w[3]), Field_16: uint32(w[4]), Field_20: uint32(w[5]), Field_24: uint32(w[6])}
						ubefore[n] = bytes.Clone(uframes[n])
					}
					objects[0].UpdateData = unsafe.Pointer(updates[i.UpdateRef])
					for n := range frames {
						before[n] = bytes.Clone(frames[n])
					}
					frame, fps := i.Frame, i.FPS
					state := func() magicMissileState53BDA0 {
						s := magicMissileState53BDA0{Frame: frame, FPS: fps}
						for n, o := range objects {
							s.Objects[n] = [10]uint32{math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), math.Float32bits(o.VelVec.X), math.Float32bits(o.VelVec.Y), math.Float32bits(o.ForceVec.X), math.Float32bits(o.ForceVec.Y), math.Float32bits(o.Float28), o.Field32, math.Float32bits(o.SpeedCur), uint32(o.ObjFlags)}
							s.Directions[n] = uint16(o.Direction2)
							s.Owners[n] = index(o.ObjOwner)
						}
						for n, u := range updates {
							s.Updates[n] = [7]int64{index(u.Owner), index(u.Target), int64(u.Field_8), int64(uint32(u.SpellID)), int64(u.Field_16), int64(u.Field_20), int64(u.Field_24)}
						}
						if objects[0].UpdateData == unsafe.Pointer(updates[0]) {
							s.UpdateRef = 0
						} else if objects[0].UpdateData == unsafe.Pointer(updates[1]) {
							s.UpdateRef = 1
						} else {
							t.Fatal("foreign update pointer")
						}
						return s
					}
					events := make([]magicMissileEvent53BDA0, 0)
					collision := func(m *server.Object) int32 {
						if m != objects[0] {
							t.Fatal("collision pointer")
						}
						events = append(events, magicMissileEvent53BDA0{Service: "collide", Args: []int64{0, 0, 0}, magicMissileState53BDA0: state()})
						return int32(i.CollideResult)
					}
					rt := server.MagicMissileUpdateRuntime53BDA0{Frame: func() uint32 { return frame }, FPS: func() uint32 { return fps }, CollideResult: collision, Collide: func(m *server.Object) { collision(m) }, DirectionFloat: func(k int32) float32 {
						w, ok := words[k]
						if !ok {
							t.Fatalf("unsealed direction index%d", k)
						}
						return math.Float32frombits(w)
					}}
					rt.SearchTarget = func(m, owner *server.Object) *server.Object {
						if m != objects[0] || owner != objects[1] {
							t.Fatal("cached search arguments")
						}
						events = append(events, magicMissileEvent53BDA0{Service: "search", Args: []int64{0, 1, 32, int64(math.Float32bits(600))}, magicMissileState53BDA0: state()})
						switch i.Mutation {
						case "":
						case "replace-update":
							objects[0].UpdateData = unsafe.Pointer(updates[1])
						case "position":
							objects[0].PosVec = types.Ptf(-500.1, 1e20)
						case "target-position":
							objects[2].PosVec = types.Ptf(1e-20, -500.3)
						case "owner":
							objects[0].ObjOwner = objects[2]
							updates[0].Owner = objects[3]
						case "speed":
							objects[0].SpeedCur = -10.3
						case "direction":
							objects[0].Direction2 = 250
						case "frame":
							frame, fps = math.MaxUint32, 0
						case "target-flags":
							objects[2].ObjFlags = 0x8020
						case "preassign-target":
							updates[0].Target = objects[3]
						default:
							t.Fatal("unknown mutation")
						}
						return pointer(i.SearchResult)
					}
					if lane == "go-root" {
						server.MagicMissileUpdate53BDA0(objects[0], rt)
					} else {
						old := magicMissileUpdateRuntime53BDA0
						magicMissileUpdateRuntime53BDA0 = func() server.MagicMissileUpdateRuntime53BDA0 { return rt }
						defer func() { magicMissileUpdateRuntime53BDA0 = old }()
						got := magicMissileCProbe53BDA0(objects[0])
						if uint32(got) != row.Result {
							t.Errorf("result %#x want %#x", uint32(got), row.Result)
						}
					}
					if got := state(); !reflect.DeepEqual(got, row.Output) {
						t.Errorf("state got %+v want %+v", got, row.Output)
					}
					if !reflect.DeepEqual(events, row.Events) {
						t.Errorf("events got %+v want %+v", events, row.Events)
					}
					var o server.Object
					ranges := [][2]uintptr{{unsafe.Offsetof(o.PosVec), 8}, {unsafe.Offsetof(o.ForceVec), 8}, {unsafe.Offsetof(o.Float28), 4}, {unsafe.Offsetof(o.Direction2), 2}, {unsafe.Offsetof(o.ObjFlags), 4}, {unsafe.Offsetof(o.ObjOwner), unsafe.Sizeof(o.ObjOwner)}, {unsafe.Offsetof(o.SpeedCur), 4}, {unsafe.Offsetof(o.UpdateData), unsafe.Sizeof(o.UpdateData)}}
					for n, b := range before {
						for _, p := range ranges {
							a, z := 32+int(p[0]), int(p[1])
							copy(b[a:a+z], frames[n][a:a+z])
						}
						if !bytes.Equal(b, frames[n]) {
							t.Fatal("nonoutput object bytes or canaries changed")
						}
					}
					for n, b := range ubefore {
						z := int(2 * unsafe.Sizeof(uintptr(0)))
						copy(b[32:32+z], uframes[n][32:32+z])
						if !bytes.Equal(b, uframes[n]) {
							t.Fatal("nonoutput update bytes or canaries changed")
						}
					}
				})
			}
			if seen != 1009 {
				t.Fatal("empty or incomplete original selection", seen)
			}
		})
	}
}
