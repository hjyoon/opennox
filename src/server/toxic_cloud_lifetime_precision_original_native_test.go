package server

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

type toxicCloudPrecisionCaseR84 struct {
	Input struct {
		Name          string `json:"name"`
		LifetimeBits  uint32 `json:"lifetime_bits"`
		FPS           uint32 `json:"fps"`
		Target        bool   `json:"target"`
		Allow         int32  `json:"allow"`
		Allocate      bool   `json:"allocate"`
		ReplaceUpdate bool   `json:"replace_update"`
	} `json:"input"`
	ScaledBits *uint32  `json:"scaled_bits"`
	Data0      int32    `json:"data0"`
	Data1      int32    `json:"data1"`
	Services   []string `json:"services"`
	Preserved  bool     `json:"all_nonoutput_frame_bytes_preserved"`
}

type toxicCloudPrecisionFixtureR84 struct {
	Original   string                       `json:"original_sha256"`
	Cases      []toxicCloudPrecisionCaseR84 `json:"cases"`
	Historical []struct {
		InputBits uint32 `json:"input_bits"`
		Result    int32  `json:"result"`
	} `json:"historical_conversions"`
}

func loadToxicCloudPrecisionR84(t *testing.T) toxicCloudPrecisionFixtureR84 {
	t.Helper()
	b, err := os.ReadFile("testdata/toxic_cloud_lifetime_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f toxicCloudPrecisionFixtureR84
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if f.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(f.Cases) != 219 || len(f.Historical) != 13 {
		t.Fatal("original toxic-cloud precision fixture identity/count")
	}
	return f
}

// Only gate/allocation/lifecycle/balance/audio services are declared inputs.
// Actual native object reads, cached UpdateData, multiply, spill and conversion
// execute together. Six C-owned records above 4 GiB retain their nonoutput
// bytes even when the old numeric result fails an assertion.
func TestPoisonGasTrapPrecisionOriginal4EB910Native(t *testing.T) {
	for _, want := range loadToxicCloudPrecisionR84(t).Cases {
		t.Run(want.Input.Name, func(t *testing.T) {
			if !want.Preserved {
				t.Fatal("uncertified original frame")
			}
			source := playerCloudGreatSwordAllocateR78[Object](t)
			target := playerCloudGreatSwordAllocateR78[Object](t)
			owner := playerCloudGreatSwordAllocateR78[Object](t)
			cloud := playerCloudGreatSwordAllocateR78[Object](t)
			data0 := playerCloudGreatSwordAllocateR78[ToxicCloudUpdateData](t)
			data1 := playerCloudGreatSwordAllocateR78[ToxicCloudUpdateData](t)
			*source = Object{PosVec: types.Pointf{X: 12.5, Y: -4.25}, ObjOwner: owner}
			*cloud = Object{UpdateData: unsafe.Pointer(data0)}
			*data0 = ToxicCloudUpdateData{Duration: 0x13572468}
			*data1 = ToxicCloudUpdateData{Duration: 0x13572468}
			type frame struct {
				name string
				ptr  unsafe.Pointer
				size uintptr
			}
			frames := []frame{{"source", unsafe.Pointer(source), unsafe.Sizeof(*source)}, {"target", unsafe.Pointer(target), unsafe.Sizeof(*target)}, {"owner", unsafe.Pointer(owner), unsafe.Sizeof(*owner)}, {"cloud", unsafe.Pointer(cloud), unsafe.Sizeof(*cloud)}, {"data0", unsafe.Pointer(data0), unsafe.Sizeof(*data0)}, {"data1", unsafe.Pointer(data1), unsafe.Sizeof(*data1)}}
			before := make([][]byte, len(frames))
			for i, f := range frames {
				before[i] = append([]byte(nil), unsafe.Slice((*byte)(f.ptr), f.size)...)
			}
			success := want.Input.Target && want.Input.Allow != 0 && want.Input.Allocate
			t.Cleanup(func() {
				preserved := true
				for i, f := range frames {
					expected := append([]byte(nil), before[i]...)
					actual := unsafe.Slice((*byte)(f.ptr), f.size)
					if success && i == 3 && want.Input.ReplaceUpdate {
						ptr := unsafe.Pointer(data1)
						copy(expected[unsafe.Offsetof(cloud.UpdateData):], unsafe.Slice((*byte)(unsafe.Pointer(&ptr)), unsafe.Sizeof(ptr)))
					}
					if success && (i == 4 && !want.Input.ReplaceUpdate || i == 5 && want.Input.ReplaceUpdate) {
						copy(expected[:4], actual[:4]) // The selected duration is the original output.
					}
					if !reflect.DeepEqual(actual, expected) {
						preserved = false
						t.Errorf("nonoutput %s frame changed", f.name)
					}
				}
				if preserved {
					t.Log("all_native_nonoutput_frames_preserved:true")
				}
			})
			var events []string
			var scaled *uint32
			dep := poisonGasTrapCollideNativeDeps4EB910{
				allowed: func(s, v *Object) int32 {
					if s != source || v != target {
						t.Fatal("gate native identity")
					}
					events = append(events, "gate")
					return want.Input.Allow
				},
				newObject: func(name string) *Object {
					if name != "ToxicCloud" {
						t.Fatal("allocation type")
					}
					events = append(events, "new")
					if !want.Input.Allocate {
						return nil
					}
					return cloud
				},
				createAt: func(c, o *Object, pos types.Pointf, reserved uint32) {
					if c != cloud || o != owner || pos != source.PosVec || reserved != 0 {
						t.Fatal("CreateAt native argument/frame identity")
					}
					events = append(events, "create")
					if want.Input.ReplaceUpdate {
						cloud.UpdateData = unsafe.Pointer(data1)
					}
				},
				loadLifetime: func(key string) float32 {
					if key != "ToxicCloudLifetime" {
						t.Fatal("balance key")
					}
					events = append(events, "balance")
					return math.Float32frombits(want.Input.LifetimeBits)
				},
				loadFPS: func() uint32 { return want.Input.FPS },
				multiply: func(value float32, fps uint32) float32 {
					v := poisonGasTrapMultiply4EB910(value, fps)
					b := math.Float32bits(v)
					scaled = &b
					return v
				},
				floatToInt: poisonGasTrapRound4EB910,
				audio: func(id uint32, s *Object, kind int32, code uint32) {
					if id != 847 || s != source || kind != 0 || code != 0 {
						t.Fatal("audio native arguments")
					}
					selected := data0
					if want.Input.ReplaceUpdate {
						selected = data1
					}
					if selected.Duration == 0x13572468 {
						t.Error("audio preceded duration store")
					}
					events = append(events, "audio")
				},
				delayedDelete: func(s *Object) {
					if s != source {
						t.Fatal("delete native identity")
					}
					events = append(events, "delete")
				},
			}
			arg := target
			if !want.Input.Target {
				arg = nil
			}
			poisonGasTrapCollideNative4EB910(source, arg, nil, dep)
			if !reflect.DeepEqual(events, want.Services) && !(len(events) == 0 && len(want.Services) == 0) {
				t.Errorf("services = %v, want %v", events, want.Services)
			}
			if !reflect.DeepEqual(scaled, want.ScaledBits) || data0.Duration != want.Data0 || data1.Duration != want.Data1 {
				t.Errorf("scaled=%v duration=(%d,%d), want scaled=%v duration=(%d,%d)", scaled, data0.Duration, data1.Duration, want.ScaledBits, want.Data0, want.Data1)
			}
		})
	}
}

func TestPoisonGasTrapConversionOriginal419A70Native(t *testing.T) {
	for _, want := range loadToxicCloudPrecisionR84(t).Historical {
		t.Run(fmt.Sprintf("bits%08x", want.InputBits), func(t *testing.T) {
			if got := poisonGasTrapRound4EB910(math.Float32frombits(want.InputBits)); got != want.Result {
				t.Fatalf("conversion = %d, want original %d", got, want.Result)
			}
		})
	}
}
