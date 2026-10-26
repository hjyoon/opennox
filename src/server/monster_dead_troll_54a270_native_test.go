package server

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type trollCloudGuard54A270 struct {
	Before [16]byte
	Data   ToxicCloudUpdateData
	After  [16]byte
}

func trollCloudAlloc54A270[T comparable](t *testing.T, value T) *T {
	t.Helper()
	ptr, free := alloc.New(value)
	t.Cleanup(free)
	*ptr = value
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(ptr)) <= math.MaxUint32 {
		t.Fatalf("C-owned native pointer is below 4 GiB: %p", ptr)
	}
	return ptr
}

func TestMonsterDeadTroll54A270OriginalNativeMatrix(t *testing.T) {
	// The historical nearest-control receipt remains unchanged. This receipt
	// replays those same inputs under the independently recorded startup CW.
	data, err := os.ReadFile("testdata/troll_dead_cloud_54a270_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle struct {
		OriginalSHA  string `json:"original_sha256"`
		CallbackSHA  string `json:"callback_sha256"`
		Float2IntSHA string `json:"float2int_sha256"`
		CW           uint32 `json:"caller_cw"`
		Rows         []struct {
			Name        string   `json:"name"`
			Lifetime    uint32   `json:"lifetime_bits"`
			FPS         uint32   `json:"fps"`
			Mutation    string   `json:"mutation"`
			AllocFailed bool     `json:"allocation_failure"`
			Result      uint32   `json:"returned"`
			OldDuration int32    `json:"old_duration"`
			NewDuration int32    `json:"replacement_duration"`
			Update      string   `json:"final_update"`
			FinalFPS    uint32   `json:"final_fps"`
			Placed      []uint32 `json:"placed_position"`
			Events      []string `json:"events"`
			Guarded     bool     `json:"guarded_bytes_unchanged"`
		} `json:"rows"`
		Entries map[string]int `json:"unhooked_entries"`
		Faults  []interface{}  `json:"memory_faults"`
		Sealed  bool           `json:"executable_bytes_unchanged"`
	}
	if err = json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	if oracle.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		oracle.CallbackSHA != "c21a25775cdac2ec64520866b21071063b02b89aaa5818ebfb511a685cc3f79e" ||
		oracle.Float2IntSHA != "10e536e2a987ea0fe14c91f007b2d5361cd87d4f0ca4cf5385fc82e53913d186" ||
		oracle.CW != 0x0e3f || len(oracle.Rows) != 30 || oracle.Entries["0x54a270"] != 30 ||
		oracle.Entries["0x419a70"] != 29 || len(oracle.Faults) != 0 || !oracle.Sealed {
		t.Fatal("bounded original callback receipt differs")
	}
	for _, row := range oracle.Rows {
		t.Run(row.Name, func(t *testing.T) {
			if !row.Guarded {
				t.Fatal("original guard bytes differ")
			}
			unit := trollCloudAlloc54A270(t, Object{PosVec: types.Ptf(7, 8), Field29: 0x12345678})
			guard := func(value int32) trollCloudGuard54A270 {
				g := trollCloudGuard54A270{Data: ToxicCloudUpdateData{Duration: value}}
				for i := range g.Before {
					g.Before[i], g.After[i] = 0x5a, 0x5a
				}
				return g
			}
			old, replacement := trollCloudAlloc54A270(t, guard(-123)), trollCloudAlloc54A270(t, guard(-456))
			cloud := trollCloudAlloc54A270(t, Object{UpdateData: unsafe.Pointer(&old.Data), Field29: 0x87654321})
			expectedUnit, expectedCloud := *unit, *cloud
			expectedOld, expectedReplacement := *old, *replacement
			unitBytes := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(unit)), unsafe.Sizeof(*unit)))
			cloudBytes := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(cloud)), unsafe.Sizeof(*cloud)))
			oldBytes := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(old)), unsafe.Sizeof(*old)))
			replacementBytes := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(replacement)), unsafe.Sizeof(*replacement)))
			t.Cleanup(func() {
				// Check all nonoutput bytes even when a numeric assertion fails.
				copy(unitBytes[unsafe.Offsetof(unit.PosVec):], unsafe.Slice((*byte)(unsafe.Pointer(&expectedUnit.PosVec)), unsafe.Sizeof(unit.PosVec)))
				copy(cloudBytes[unsafe.Offsetof(cloud.PosVec):], unsafe.Slice((*byte)(unsafe.Pointer(&expectedCloud.PosVec)), unsafe.Sizeof(cloud.PosVec)))
				copy(cloudBytes[unsafe.Offsetof(cloud.ObjOwner):], unsafe.Slice((*byte)(unsafe.Pointer(&expectedCloud.ObjOwner)), unsafe.Sizeof(cloud.ObjOwner)))
				copy(cloudBytes[unsafe.Offsetof(cloud.UpdateData):], unsafe.Slice((*byte)(unsafe.Pointer(&expectedCloud.UpdateData)), unsafe.Sizeof(cloud.UpdateData)))
				if !row.AllocFailed {
					copy(oldBytes[unsafe.Offsetof(old.Data):], unsafe.Slice((*byte)(unsafe.Pointer(&old.Data.Duration)), 4))
				}
				if !bytes.Equal(unitBytes, unsafe.Slice((*byte)(unsafe.Pointer(unit)), unsafe.Sizeof(*unit))) ||
					!bytes.Equal(cloudBytes, unsafe.Slice((*byte)(unsafe.Pointer(cloud)), unsafe.Sizeof(*cloud))) ||
					!bytes.Equal(oldBytes, unsafe.Slice((*byte)(unsafe.Pointer(old)), unsafe.Sizeof(*old))) ||
					!bytes.Equal(replacementBytes, unsafe.Slice((*byte)(unsafe.Pointer(replacement)), unsafe.Sizeof(*replacement))) {
					t.Error("Troll nonoutput native frame changed")
				} else {
					t.Log("all_native_Troll_nonoutput_frames_preserved:true")
				}
			})
			fps := row.FPS
			var events []string
			mutate := func(stage string) {
				if row.Mutation == stage+"-replace" {
					cloud.UpdateData = unsafe.Pointer(&replacement.Data)
					expectedCloud.UpdateData = cloud.UpdateData
				} else if row.Mutation == stage+"-nil" {
					cloud.UpdateData = nil
					expectedCloud.UpdateData = nil
				}
				if row.Mutation == "create-position" && stage == "create" {
					unit.PosVec = types.Ptf(99, 101)
					expectedUnit.PosVec = unit.PosVec
				}
				if row.Mutation == "balance-fps" && stage == "balance" {
					fps = 31
				}
			}
			rt := MonsterDeadCallbackRuntime549D80{
				NewObjectByTypeID: func(name string) *Object {
					events = append(events, "new:"+name)
					if name != "SmallToxicCloud" {
						t.Fatal("wrong constructor type")
					}
					if row.AllocFailed {
						return nil
					}
					return cloud
				},
				CreateObjectAt: func(got, owner *Object, pos types.Pointf) {
					events = append(events, "create")
					if got != cloud || owner != unit || pos != types.Ptf(7, 8) {
						t.Fatal("constructor/owner/position read order differs")
					}
					cloud.ObjOwner, cloud.PosVec = owner, pos
					expectedCloud.ObjOwner, expectedCloud.PosVec = owner, pos
					mutate("create")
				},
				Audio: func(id sound.ID, got *Object) {
					if id != sound.SoundTrollFlatus || got != unit {
						t.Fatal("sound or source identity differs")
					}
					events = append(events, "audio:644")
					mutate("audio")
				},
				BalanceFloat: func(key string) float64 {
					events = append(events, "balance:"+key)
					if key != "SmallToxicCloudLifetime" {
						t.Fatal("wrong lifetime key")
					}
					mutate("balance")
					return float64(math.Float32frombits(row.Lifetime))
				},
				TickRate: func() uint32 {
					events = append(events, "fps")
					return fps
				},
			}
			result := MonsterDeadCallbackNative549D80(unit, MonsterDeadCallbackTroll54A270, rt)
			if result != (row.Result == 1) || !reflect.DeepEqual(events, row.Events) || fps != row.FinalFPS {
				t.Fatalf("service order/result differs: %t/%q/fps-%d, want %d/%q/fps-%d", result, events, fps, row.Result, row.Events, row.FinalFPS)
			}
			expectedOld.Data.Duration, expectedReplacement.Data.Duration = row.OldDuration, row.NewDuration
			if *old != expectedOld || *replacement != expectedReplacement {
				t.Fatalf("cached lifetime/guard differs: old=%d new=%d, want %d/%d", old.Data.Duration, replacement.Data.Duration, row.OldDuration, row.NewDuration)
			}
			update := "original"
			if cloud.UpdateData == nil {
				update = "nil"
			} else if cloud.UpdateData == unsafe.Pointer(&replacement.Data) {
				update = "replacement"
			}
			if update != row.Update || *unit != expectedUnit || *cloud != expectedCloud {
				t.Fatal("callback changed object fields outside its cached duration store")
			}
			if !row.AllocFailed && (len(row.Placed) != 2 || math.Float32bits(cloud.PosVec.X) != row.Placed[0] || math.Float32bits(cloud.PosVec.Y) != row.Placed[1]) {
				t.Fatal("original placement bits differ")
			}
		})
	}
}

func TestMonsterDeadTroll54A270RetainsNativeAdmissionGuards(t *testing.T) {
	for _, missing := range []string{"new", "create", "audio", "balance", "fps", "nil-update"} {
		t.Run(missing, func(t *testing.T) {
			unit, cloud := trollCloudAlloc54A270(t, Object{}), trollCloudAlloc54A270(t, Object{})
			var events []byte
			rt := MonsterDeadCallbackRuntime549D80{
				NewObjectByTypeID: func(string) *Object { events = append(events, 'n'); return cloud },
				CreateObjectAt:    func(*Object, *Object, types.Pointf) { events = append(events, 'c') },
				Audio:             func(sound.ID, *Object) { events = append(events, 'a') },
				BalanceFloat:      func(string) float64 { events = append(events, 'b'); return 1.25 },
				TickRate:          func() uint32 { events = append(events, 'f'); return 30 },
			}
			switch missing {
			case "new":
				rt.NewObjectByTypeID = nil
			case "create":
				rt.CreateObjectAt = nil
			case "audio":
				rt.Audio = nil
			case "balance":
				rt.BalanceFloat = nil
			case "fps":
				rt.TickRate = nil
			}
			ok := MonsterDeadCallbackNative549D80(unit, MonsterDeadCallbackTroll54A270, rt)
			want := missing == "nil-update"
			if ok != want || want && !bytes.Equal(events, []byte("ncabf")) || !want && len(events) != 0 {
				t.Fatalf("native admission: result=%t events=%s", ok, events)
			}
		})
	}
}
