package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
)

type sharedCloudPrecisionRowR84 struct {
	Input struct {
		Name         string    `json:"name"`
		LifetimeBits uint32    `json:"lifetime_bits"`
		FPS          uint32    `json:"fps"`
		Mutation     string    `json:"mutation"`
		AllocFailed  bool      `json:"allocation_failure"`
		PositionBits [2]uint32 `json:"position_bits"`
		Intensity    byte      `json:"intensity"`
	} `json:"input"`
	OldDuration int32    `json:"old_duration"`
	NewDuration int32    `json:"replacement_duration"`
	Events      []string `json:"events"`
	Packet      []byte   `json:"packet"`
	Preserved   bool     `json:"guarded_bytes_unchanged"`
}

func sharedCloudPrecisionFixtureR84(t *testing.T) struct {
	Original string                       `json:"original_sha256"`
	Polyp    []sharedCloudPrecisionRowR84 `json:"polyp"`
	Packets  []sharedCloudPrecisionRowR84 `json:"packets"`
} {
	t.Helper()
	var f struct {
		Original string                       `json:"original_sha256"`
		Polyp    []sharedCloudPrecisionRowR84 `json:"polyp"`
		Packets  []sharedCloudPrecisionRowR84 `json:"packets"`
	}
	b, err := os.ReadFile("testdata/shared_cloud_callers_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if f.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(f.Polyp) != 13 || len(f.Packets) != 16 {
		t.Fatal("original shared-cloud fixture identity/count")
	}
	return f
}

func TestPolypCloudPrecisionOriginal54CB10Native(t *testing.T) {
	for _, want := range sharedCloudPrecisionFixtureR84(t).Polyp {
		t.Run(want.Input.Name, func(t *testing.T) {
			if !want.Preserved {
				t.Fatal("uncertified original frame")
			}
			source := trollCloudAlloc54A270(t, Object{PosVec: types.Pointf{X: math.Float32frombits(want.Input.PositionBits[0]), Y: math.Float32frombits(want.Input.PositionBits[1])}, Field29: 0x12345678})
			guard := func(v int32) trollCloudGuard54A270 {
				g := trollCloudGuard54A270{Data: ToxicCloudUpdateData{Duration: v}}
				for i := range g.Before {
					g.Before[i], g.After[i] = 0x5a, 0x5a
				}
				return g
			}
			old := trollCloudAlloc54A270(t, guard(-123))
			replacement := trollCloudAlloc54A270(t, guard(-456))
			cloud := trollCloudAlloc54A270(t, Object{UpdateData: unsafe.Pointer(&old.Data), Field29: 0x87654321})
			expectedCloud := *cloud
			sourceBytes := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(source)), unsafe.Sizeof(*source)))
			cloudBytes := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(cloud)), unsafe.Sizeof(*cloud)))
			oldBytes := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(old)), unsafe.Sizeof(*old)))
			replacementBytes := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(replacement)), unsafe.Sizeof(*replacement)))
			t.Cleanup(func() {
				copy(cloudBytes[unsafe.Offsetof(cloud.PosVec):], unsafe.Slice((*byte)(unsafe.Pointer(&expectedCloud.PosVec)), unsafe.Sizeof(cloud.PosVec)))
				copy(cloudBytes[unsafe.Offsetof(cloud.ObjOwner):], unsafe.Slice((*byte)(unsafe.Pointer(&expectedCloud.ObjOwner)), unsafe.Sizeof(cloud.ObjOwner)))
				copy(cloudBytes[unsafe.Offsetof(cloud.UpdateData):], unsafe.Slice((*byte)(unsafe.Pointer(&expectedCloud.UpdateData)), unsafe.Sizeof(cloud.UpdateData)))
				if !want.Input.AllocFailed {
					copy(oldBytes[unsafe.Offsetof(old.Data):], unsafe.Slice((*byte)(unsafe.Pointer(&old.Data.Duration)), 4))
				}
				if !bytes.Equal(sourceBytes, unsafe.Slice((*byte)(unsafe.Pointer(source)), unsafe.Sizeof(*source))) ||
					!bytes.Equal(cloudBytes, unsafe.Slice((*byte)(unsafe.Pointer(cloud)), unsafe.Sizeof(*cloud))) ||
					!bytes.Equal(oldBytes, unsafe.Slice((*byte)(unsafe.Pointer(old)), unsafe.Sizeof(*old))) ||
					!bytes.Equal(replacementBytes, unsafe.Slice((*byte)(unsafe.Pointer(replacement)), unsafe.Sizeof(*replacement))) {
					t.Error("Polyp nonoutput native frame changed")
				} else {
					t.Log("all_native_Polyp_nonoutput_frames_preserved:true")
				}
			})
			var events []string
			PolypDieNative54CB10(source, PolypDeathRuntime54CB10{
				NewObjectByTypeID: func(name string) *Object {
					if name != "ToxicCloud" {
						t.Fatal("allocation type")
					}
					events = append(events, "new:"+name)
					if want.Input.AllocFailed {
						return nil
					}
					return cloud
				},
				CreateObjectAt: func(c, owner *Object, pos types.Pointf) {
					if c != cloud || owner != nil || pos != source.PosVec {
						t.Fatal("CreateAt native identity/position")
					}
					events = append(events, "create")
					cloud.PosVec, cloud.ObjOwner = pos, owner
					expectedCloud.PosVec, expectedCloud.ObjOwner = pos, owner
					switch want.Input.Mutation {
					case "create-replace":
						cloud.UpdateData = unsafe.Pointer(&replacement.Data)
					case "create-nil":
						cloud.UpdateData = nil
					}
					expectedCloud.UpdateData = cloud.UpdateData
				},
				BalanceFloat: func(key string) float64 {
					if key != "ToxicCloudLifetime" {
						t.Fatal("balance key")
					}
					events = append(events, "balance:"+key)
					return float64(math.Float32frombits(want.Input.LifetimeBits))
				},
				TickRate: func() uint32 {
					events = append(events, "fps")
					return want.Input.FPS
				},
				Audio: func(id sound.ID, got *Object) {
					if id != sound.SoundPolypExplode || got != source {
						t.Fatal("audio native identity")
					}
					if old.Data.Duration != want.OldDuration || replacement.Data.Duration != want.NewDuration {
						t.Errorf("audio duration = %d/%d, want original %d/%d", old.Data.Duration, replacement.Data.Duration, want.OldDuration, want.NewDuration)
					}
					events = append(events, "audio:284")
				},
				DelayedDelete: func(got *Object) {
					if got != source {
						t.Fatal("delete native identity")
					}
					events = append(events, "delete")
				},
			})
			if !reflect.DeepEqual(events, want.Events) || old.Data.Duration != want.OldDuration || replacement.Data.Duration != want.NewDuration {
				t.Errorf("events/durations = %v/%d/%d, want original %v/%d/%d", events, old.Data.Duration, replacement.Data.Duration, want.Events, want.OldDuration, want.NewDuration)
			}
		})
	}
}

func TestMonsterGeneratorBreakFXPrecisionOriginal523200(t *testing.T) {
	for _, want := range sharedCloudPrecisionFixtureR84(t).Packets {
		t.Run(want.Input.Name, func(t *testing.T) {
			if !want.Preserved || len(want.Packet) != 7 {
				t.Fatal("uncertified original packet/point frame")
			}
			point := types.Pointf{X: math.Float32frombits(want.Input.PositionBits[0]), Y: math.Float32frombits(want.Input.PositionBits[1])}
			got := monsterGeneratorBreakFXPacket523200(point, want.Input.Intensity)
			if !bytes.Equal(got[:], want.Packet) {
				t.Fatal(fmt.Sprintf("packet = %x, want original %x", got, want.Packet))
			}
		})
	}
}
