package server

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type cloudPoisonQuestServiceR81 struct {
	Service    string `json:"service"`
	Flag       uint32 `json:"flag"`
	Value      int    `json:"value"`
	ScaleBits  uint32 `json:"scale_bits32"`
	DamageBits uint32 `json:"damage_bits32"`
	CarryBits  uint32 `json:"carry_bits32"`
}
type cloudPoisonQuestCaseR81 struct {
	Mode  string `json:"mode"`
	Input struct {
		Name           string `json:"name"`
		Player         bool   `json:"player"`
		Quest          bool   `json:"quest"`
		Coop           bool   `json:"coop"`
		Raw            int32  `json:"raw"`
		ScaleBits      uint32 `json:"quest_scale_bits32"`
		CarryBits      uint32 `json:"carry_bits32"`
		AbsorptionBits uint32 `json:"absorption_bits32"`
		SourceProfile  string `json:"source_profile"`
		SourceClass    uint32 `json:"source_class"`
		CloudType      uint16 `json:"flame_type"`
		CloudClass     uint32 `json:"flame_class"`
		God            bool   `json:"god"`
	} `json:"input"`
	Services   []cloudPoisonQuestServiceR81 `json:"services"`
	CarryBits  uint32                       `json:"carry_bits32"`
	Marker     uint32                       `json:"marker"`
	MarkerType uint32                       `json:"marker_type"`
	Preserved  bool                         `json:"all_nonoutput_frame_bytes_preserved"`
}

func TestPlayerDamageCloudPoisonQuestPrecisionOriginal4E17B0Native(t *testing.T) {
	cloudPoisonQuestOriginalR81(t, false)
}
func TestPlayerDamageCloudPoisonQuestPartialPrecisionOriginal4E17B0Native(t *testing.T) {
	cloudPoisonQuestOriginalR81(t, true)
}

// Unchanged original PlayerDamage, scalar getter and signed conversion run
// independently under startup and three FP controls. Native C frames above
// 4 GiB are checked even when the old numeric result fails.
func cloudPoisonQuestOriginalR81(t *testing.T, partial bool) {
	b, err := os.ReadFile("testdata/cloud_poison_quest_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Original string                    `json:"original_sha256"`
		Cases    []cloudPoisonQuestCaseR81 `json:"cases"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if f.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(f.Cases) != 1456 {
		t.Fatal("original fixture identity/count")
	}
	for _, want := range f.Cases {
		if partial {
			if len(want.Services) == 0 || want.Services[0].Service != "mode" || want.Services[0].Flag != 2048 {
				t.Fatal("original Coop prefix")
			}
			want.Services = slices.Clone(want.Services[1:])
		}
		t.Run(want.Input.Name, func(t *testing.T) {
			if want.Mode != "actual-startup-inherited" || !want.Preserved || want.Input.SourceProfile != "monster" || want.Input.SourceClass != 2 || want.Input.CloudType != 1174 || want.Input.CloudClass != 0x110000 || want.Input.God {
				t.Fatal("unexpected declared original input")
			}
			target, source, cloud := playerCloudGreatSwordAllocateR78[Object](t), playerCloudGreatSwordAllocateR78[Object](t), playerCloudGreatSwordAllocateR78[Object](t)
			playerData := playerCloudGreatSwordAllocateR78[PlayerUpdateData](t)
			npcData := playerCloudGreatSwordAllocateR78[MonsterUpdateData](t)
			sourceData := playerCloudGreatSwordAllocateR78[MonsterUpdateData](t)
			info := playerCloudGreatSwordAllocateR78[Player](t)
			*playerData = PlayerUpdateData{Player: info, Field21: want.Input.CarryBits, Field57: want.Input.AbsorptionBits, Field76: 99, Field75: 77}
			*npcData = MonsterUpdateData{Field1: want.Input.CarryBits, Field518: want.Input.AbsorptionBits, Field547: 99, Field546: 77}
			*target = Object{ObjClass: object.ClassMonster, ObjSubClass: 0x10, UpdateData: unsafe.Pointer(npcData)}
			marker, kind, carry := &npcData.Field547, &npcData.Field546, &npcData.Field1
			if want.Input.Player {
				*target = Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(playerData)}
				marker, kind, carry = &playerData.Field76, &playerData.Field75, &playerData.Field21
			}
			*source = Object{TypeInd: 0x6701, ObjClass: object.ClassMonster, ObjSubClass: 0x10, UpdateData: unsafe.Pointer(sourceData)}
			*cloud = Object{TypeInd: 1174, ObjClass: object.Class(0x110000), PosVec: types.Ptf(100, 0), PrevPos: types.Ptf(100, 0)}
			type frame struct {
				p       unsafe.Pointer
				before  []byte
				allowed []bool
			}
			var frames []*frame
			watch := func(p unsafe.Pointer, n uintptr) {
				if uintptr(p) <= math.MaxUint32 {
					t.Fatal("C frame below4GiB")
				}
				frames = append(frames, &frame{p, slices.Clone(unsafe.Slice((*byte)(p), int(n))), make([]bool, int(n))})
			}
			for _, p := range []*Object{target, source, cloud} {
				watch(unsafe.Pointer(p), unsafe.Sizeof(*p))
			}
			watch(unsafe.Pointer(playerData), unsafe.Sizeof(*playerData))
			watch(unsafe.Pointer(npcData), unsafe.Sizeof(*npcData))
			watch(unsafe.Pointer(sourceData), unsafe.Sizeof(*sourceData))
			watch(unsafe.Pointer(info), unsafe.Sizeof(*info))
			permit := func(p unsafe.Pointer, n uintptr) {
				for _, f := range frames {
					lo, at := uintptr(f.p), uintptr(p)
					if at >= lo && at+n <= lo+uintptr(len(f.before)) {
						for i := at - lo; i < at-lo+n; i++ {
							f.allowed[i] = true
						}
						return
					}
				}
				t.Fatal("mutable field outside C frames")
			}
			permit(unsafe.Pointer(marker), 4)
			permit(unsafe.Pointer(kind), 4)
			t.Cleanup(func() {
				for _, f := range frames {
					after := unsafe.Slice((*byte)(f.p), len(f.before))
					for i, v := range after {
						if !f.allowed[i] && f.before[i] != v {
							t.Fatalf("native nonoutput frame changed at%d", i)
						}
					}
				}
				t.Log("all_native_nonoutput_frames_preserved:true")
			})
			var services []cloudPoisonQuestServiceR81
			add := func(e cloudPoisonQuestServiceR81) { services = append(services, e) }
			attributed := func() {
				if *marker != 1 || *kind != 1174 || *carry != want.Input.CarryBits {
					t.Fatal("cached marker/live carry order")
				}
			}
			runtime := PlayerDamageRuntime4E17B0{
				CoopMode: func() bool {
					if *marker != 99 || *kind != 77 {
						t.Fatal("Coop query after marker mutation")
					}
					v := 0
					if want.Input.Coop {
						v = 1
					}
					add(cloudPoisonQuestServiceR81{Service: "mode", Flag: 2048, Value: v})
					return want.Input.Coop
				},
				BlockSourceExcluded: func(got *Object) bool {
					if got != cloud || *marker != 0 || *kind != 77 {
						t.Fatal("cloud exclusion prefix")
					}
					return false
				},
				BlockDirection: func(got *Object, pos types.Pointf) bool {
					if got != target || pos != types.Ptf(100, 0) {
						t.Fatal("cloud ordinary direction arguments")
					}
					attributed()
					add(cloudPoisonQuestServiceR81{Service: "direction"})
					return false
				},
				QuestMode: func() bool {
					attributed()
					v := 0
					if want.Input.Quest {
						v = 1
					}
					add(cloudPoisonQuestServiceR81{Service: "mode", Flag: 4096, Value: v})
					return want.Input.Quest
				},
				QuestDamageScale: func() float32 {
					attributed()
					add(cloudPoisonQuestServiceR81{Service: "quest-scale", ScaleBits: want.Input.ScaleBits})
					return math.Float32frombits(want.Input.ScaleBits)
				},
				DefaultDamage: func(got, by, w *Object, raw int32, typ object.DamageType) bool {
					if got != target || by != source || w != cloud || typ != object.DamagePoison {
						t.Fatal("cloud default identity/type")
					}
					attributed()
					add(cloudPoisonQuestServiceR81{Service: "default", DamageBits: uint32(raw), CarryBits: *carry})
					return true
				},
			}
			if partial {
				runtime.CoopMode = nil
			}
			var reason string
			runtime.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
			handled, result := PlayerDamageNative4E17B0(target, source, cloud, want.Input.Raw, object.DamagePoison, runtime)
			if !handled || !result || reason != "" {
				t.Fatalf("original cloud POISON admission mismatch handled/result=%t/%t reason=%q", handled, result, reason)
			}
			if !slices.Equal(services, want.Services) {
				t.Fatalf("original cloud POISON Quest result mismatch services=%+v want=%+v", services, want.Services)
			}
			got := want
			got.Services = services
			got.CarryBits = *carry
			got.Marker = *marker
			got.MarkerType = *kind
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("original cloud POISON output got=%+v want=%+v", got, want)
			}
		})
	}
}
