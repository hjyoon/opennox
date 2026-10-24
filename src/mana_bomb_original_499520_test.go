package opennox

import (
	"encoding/json"
	"image"
	"math"
	"os"
	"reflect"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
)

type manaBombOriginalState499520 struct {
	Pos      [2]int    `json:"pos"`
	Effect   [5]uint32 `json:"effect"`
	Z        int16     `json:"z"`
	VelZ     int8      `json:"vel_z"`
	Aux      byte      `json:"aux"`
	Field127 uint32    `json:"field127"`
	Update   uint32    `json:"update"`
	Field8   uint32    `json:"field8"`
	Field9   uint32    `json:"field9"`
}
type manaBombOriginalSpawn499520 struct {
	Type    int    `json:"type"`
	Pos     [2]int `json:"pos"`
	Success bool   `json:"success"`
	ID      int    `json:"id"`
}
type manaBombOriginalRandom499520 struct {
	Lo    int `json:"lo"`
	Hi    int `json:"hi"`
	Value int `json:"value"`
}
type manaBombOriginalActivation499520 struct {
	ID    int                         `json:"id"`
	State manaBombOriginalState499520 `json:"state"`
}

type manaBombOriginalTrace499520 struct {
	Service string `json:"service"`
	Lo      int    `json:"lo"`
	Hi      int    `json:"hi"`
	Value   int64  `json:"value"`
	Type    int    `json:"type"`
	Pos     [2]int `json:"pos"`
	Success bool   `json:"success"`
	ID      int    `json:"id"`
}

func TestManaBombOriginal499520(t *testing.T) {
	raw, err := os.ReadFile("testdata/mana_bomb_original_499520.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA       string `json:"original_sha256"`
		Unchanged bool   `json:"original_unchanged"`
		FPCW      uint32 `json:"fpcw"`
		Rows      []struct {
			Input struct {
				Name      string  `json:"name"`
				Kind      string  `json:"kind"`
				Path      [4]int  `json:"path"`
				Pos       []int   `json:"pos"`
				Angle     uint16  `json:"angle"`
				Reverse   byte    `json:"reverse"`
				Fill      byte    `json:"fill"`
				Mode      byte    `json:"mode"`
				Frame     uint32  `json:"frame"`
				Start     *uint32 `json:"start"`
				Field80   *uint32 `json:"field80"`
				Distance  *uint16 `json:"distance"`
				Radius    int     `json:"radius"`
				FrameStep uint32  `json:"frame_step"`
				Fail      bool    `json:"fail"`
				Failure   string  `json:"failure"`
			} `json:"input"`
			Spawns      []manaBombOriginalSpawn499520      `json:"spawns"`
			Randoms     []manaBombOriginalRandom499520     `json:"randoms"`
			Frames      []uint32                           `json:"frames"`
			Trace       []manaBombOriginalTrace499520      `json:"trace"`
			Activations []manaBombOriginalActivation499520 `json:"activations"`
			Events      []struct {
				Service string `json:"service"`
				Pos     [2]int `json:"pos"`
			} `json:"events"`
			Source manaBombOriginalState499520 `json:"source_state"`
		} `json:"rows"`
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.FPCW != 0xe7f || len(fixture.Rows) != 754 {
		t.Fatal("unchanged original ManaBomb fixture")
	}
	count := 0
	for _, row := range fixture.Rows {
		if row.Input.Kind == "payload" || row.Input.Kind == "spawn" && row.Input.Angle > 255 {
			continue
		}
		count++
		t.Run(row.Input.Name, func(t *testing.T) {
			in := row.Input
			pattern := uint32(in.Fill) * 0x01010101
			marker := new(byte)
			makeDr := func() *client.Drawable {
				d := &client.Drawable{ZVal: uint16(pattern), VelZ: int8(in.Fill), Field_74_4: in.Fill, Field_127: pattern, Field_8: pattern, Field_9: pattern, AnimStart: in.Frame}
				e := d.UnionEffect()
				e.Field_108, e.Field_109, e.Field_110, e.Field_111, e.Field_112 = pattern, pattern, pattern, pattern, pattern
				if in.Fill != 0 {
					d.ClientUpdateFuncPtr = unsafe.Pointer(marker)
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(d)) <= math.MaxUint32 {
					t.Fatal("native drawable address must exceed4GiB")
				}
				return d
			}
			state := func(d *client.Drawable) manaBombOriginalState499520 {
				update := uint32(0)
				switch d.ClientUpdateFuncPtr {
				case nil:
				case legacy.Get_sub_4CA720():
					update = 0x4ca720
				case unsafe.Pointer(marker):
					update = pattern
				default:
					t.Fatal("unexpected update pointer")
				}
				e := d.UnionEffect()
				return manaBombOriginalState499520{Pos: [2]int{d.PosVec.X, d.PosVec.Y}, Effect: [5]uint32{e.Field_108, e.Field_109, e.Field_110, e.Field_111, e.Field_112}, Z: int16(d.ZVal), VelZ: d.VelZ, Aux: d.Field_74_4, Field127: d.Field_127, Update: update, Field8: d.Field_8, Field9: d.Field_9}
			}
			source := makeDr()
			source.PosVec = image.Pt(100, 200)
			if len(in.Pos) == 2 {
				source.PosVec = image.Pt(in.Pos[0], in.Pos[1])
			}
			source.AnimStart = 500
			if in.Start != nil {
				source.AnimStart = *in.Start
			}
			source.Field_80 = 500
			if in.Field80 != nil {
				source.Field_80 = *in.Field80
			}
			e := source.UnionEffect()
			e.Field_108 = uint32(uint16(in.Path[0])) | uint32(uint16(in.Path[1]))<<16
			e.Field_109 = uint32(uint16(in.Path[2])) | uint32(uint16(in.Path[3]))<<16
			distance := uint16(30)
			if in.Distance != nil {
				distance = *in.Distance
			}
			e.Field_110 = uint32(distance) | uint32(byte(in.Angle))<<16 | uint32(in.Reverse)<<24
			before := *source
			var spawns []manaBombOriginalSpawn499520
			var randoms []manaBombOriginalRandom499520
			var frames []uint32
			var trace []manaBombOriginalTrace499520
			var activations []manaBombOriginalActivation499520
			ids := map[*client.Drawable]int{}
			spawnIndex, randomIndex := 0, 0
			hooks := manaBombDrawableHooks4CCAC0{
				radius: func() int { return in.Radius }, typeID: func(name string) int {
					switch name {
					case "VioletSpark":
						return 1
					case "ManaBombOrb":
						return 2
					default:
						t.Fatal(name)
						return 0
					}
				},
				frame: func() uint32 {
					v := in.Frame + uint32(len(frames))*in.FrameStep
					frames = append(frames, v)
					trace = append(trace, manaBombOriginalTrace499520{Service: "frame", Value: int64(v)})
					return v
				},
				random: func(lo, hi int) int {
					randomIndex++
					v := 0
					switch {
					case lo == 0 && hi == 255:
						v = (randomIndex*37 + 11) % 256
					case lo == 0:
						if hi >= 0 {
							v = []int{0, hi, hi / 2}[randomIndex%3]
						}
					case lo == 10 && hi == 30:
						v = 20
					case lo == 2 && hi == 8:
						v = 4
					case lo == 3 && hi == 10:
						v = 5
					default:
						t.Fatalf("unexpected RNG range%d..%d", lo, hi)
					}
					randoms = append(randoms, manaBombOriginalRandom499520{lo, hi, v})
					trace = append(trace, manaBombOriginalTrace499520{Service: "random", Lo: lo, Hi: hi, Value: int64(v)})
					return v
				},
				spawn: func(typ int, pos image.Point) *client.Drawable {
					spawnIndex++
					fail := in.Fail || in.Failure == "all" || in.Failure == "orbs" && typ == 2 || in.Failure == "sparks" && typ == 1 || in.Failure == "every5" && spawnIndex%5 == 0
					spawns = append(spawns, manaBombOriginalSpawn499520{typ, [2]int{pos.X, pos.Y}, !fail, spawnIndex})
					trace = append(trace, manaBombOriginalTrace499520{Service: "spawn", Type: typ, Pos: [2]int{pos.X, pos.Y}, Success: !fail, ID: spawnIndex})
					if fail {
						return nil
					}
					d := makeDr()
					d.PosVec = pos
					d.TypeIDVal = uint32(typ)
					ids[d] = spawnIndex
					return d
				},
				activate: func(d *client.Drawable) {
					activations = append(activations, manaBombOriginalActivation499520{ids[d], state(d)})
					trace = append(trace, manaBombOriginalTrace499520{Service: "activate", ID: ids[d]})
				},
			}
			switch in.Kind {
			case "spawn":
				spawnManaBombOrb4CA720(2, image.Pt(in.Path[0], in.Path[1]), image.Pt(in.Path[2], in.Path[3]), byte(in.Angle), in.Reverse != 0, hooks)
			case "charge":
				if updateManaBombCharge4CCAC0(source, hooks) != 1 {
					t.Fatal("charge result")
				}
			case "step":
				next, done := manaBombOrbStep4CA720(source, in.Frame)
				if len(row.Events) != 1 {
					t.Fatal("original step service")
				}
				event := row.Events[0]
				wantDone := event.Service == "delete"
				if done != wantDone || !done && next != image.Pt(event.Pos[0], event.Pos[1]) {
					t.Fatalf("next=%v done=%v want=%v/%v", next, done, event.Pos, wantDone)
				}
				if !done && (row.Source.Field8 != uint32(event.Pos[0]) || row.Source.Field9 != uint32(event.Pos[1])) {
					t.Fatal("original post-position marker")
				}
			default:
				t.Fatal(in.Kind)
			}
			if !reflect.DeepEqual(*source, before) {
				t.Fatal("source/input canary changed")
			}
			if in.Kind != "step" && !slices.Equal(trace, row.Trace) {
				t.Fatalf("service trace=%v want=%v", trace, row.Trace)
			}
			if in.Kind != "step" && (!slices.Equal(spawns, row.Spawns) || !slices.Equal(randoms, row.Randoms) || !slices.Equal(frames, row.Frames) || !slices.Equal(activations, row.Activations)) {
				t.Fatalf("spawns=%v/%v random=%v/%v frames=%v/%v activations=%v/%v", spawns, row.Spawns, randoms, row.Randoms, frames, row.Frames, activations, row.Activations)
			}
		})
	}
	if count != 712 {
		t.Fatal("original native vector count", count)
	}
}
