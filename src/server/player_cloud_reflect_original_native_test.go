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

type cloudReflectServiceR79 struct {
	Service string `json:"service"`
	Phase   string `json:"phase"`
	Flag    uint32 `json:"flag"`
	Value   int    `json:"value"`
	Sound   int    `json:"sound"`
	Damage  int32  `json:"damage"`
}
type cloudReflectCaseR79 struct {
	Input struct {
		Name     string `json:"name"`
		Mode     string `json:"mode"`
		Subclass uint32 `json:"subclass"`
		Raw      int32  `json:"raw"`
		Coop     bool   `json:"coop"`
	} `json:"input"`
	Services         []cloudReflectServiceR79 `json:"services"`
	Result           uint32                   `json:"result"`
	Marker           uint32                   `json:"marker"`
	MarkerType       uint32                   `json:"marker_type"`
	LiveMarker       uint32                   `json:"live_marker"`
	LiveMarkerType   uint32                   `json:"live_marker_type"`
	CloudClass       uint32                   `json:"cloud_class"`
	CloudSubclass    uint32                   `json:"cloud_subclass"`
	CloudType        uint16                   `json:"cloud_type"`
	CloudOwner       string                   `json:"cloud_owner"`
	Buffs            uint32                   `json:"buffs"`
	InfoFlags        uint32                   `json:"info_flags"`
	ObserverCleared  bool                     `json:"observer_cleared"`
	TargetLiveUpdate bool                     `json:"target_live_update"`
	Preserved        bool                     `json:"all_nonoutput_frame_bytes_preserved"`
}

func TestPlayerDamagePlayerCloudReflectOriginal4E17B0Native(t *testing.T) {
	cloudReflectOriginalR79(t, false)
}
func TestPlayerDamagePlayerCloudReflectPartialOriginal4E17B0Native(t *testing.T) {
	cloudReflectOriginalR79(t, true)
}

// Independent unchanged-original384contexts, afterObserveClear introducesa
// missile. Callback services are declared; every native C frame is above4GiB.
// Cleanup validates nonoutput bytes even at the old unsupported boundary.
func cloudReflectOriginalR79(t *testing.T, partial bool) {
	b, err := os.ReadFile("testdata/player_cloud_reflect_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Original string                `json:"original_sha256"`
		Cases    []cloudReflectCaseR79 `json:"cases"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if f.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(f.Cases) != 384 {
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
			target, source, cloud := playerCloudGreatSwordAllocateR78[Object](t), playerCloudGreatSwordAllocateR78[Object](t), playerCloudGreatSwordAllocateR78[Object](t)
			entry, live := playerCloudGreatSwordAllocateR78[PlayerUpdateData](t), playerCloudGreatSwordAllocateR78[PlayerUpdateData](t)
			info := playerCloudGreatSwordAllocateR78[Player](t)
			su := playerCloudGreatSwordAllocateR78[MonsterUpdateData](t)
			*target = Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(entry), Buffs: 1 << 27}
			*source = Object{ObjClass: object.ClassMonster, ObjSubClass: 0x10, UpdateData: unsafe.Pointer(su)}
			*cloud = Object{TypeInd: 1174, ObjClass: object.Class(0x110000), PosVec: types.Ptf(19, 11), PrevPos: types.Ptf(17, 9), ObjOwner: source}
			*entry = PlayerUpdateData{Player: info, State: PlayerState13, Field76: 99, Field75: 77}
			*live = PlayerUpdateData{Player: info, Field76: 31, Field75: 33}
			info.Field3680 = 2
			info.CameraFollowObj = source
			type frame struct {
				p       unsafe.Pointer
				before  []byte
				allowed []bool
			}
			var frames []*frame
			watch := func(p unsafe.Pointer, n uintptr) {
				if p == nil || unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
					t.Fatal("C frame below4GiB")
				}
				frames = append(frames, &frame{p, slices.Clone(unsafe.Slice((*byte)(p), int(n))), make([]bool, int(n))})
			}
			for _, p := range []*Object{target, source, cloud} {
				watch(unsafe.Pointer(p), unsafe.Sizeof(*p))
			}
			for _, p := range []*PlayerUpdateData{entry, live} {
				watch(unsafe.Pointer(p), unsafe.Sizeof(*p))
			}
			watch(unsafe.Pointer(info), unsafe.Sizeof(*info))
			watch(unsafe.Pointer(su), unsafe.Sizeof(*su))
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
			permit(unsafe.Pointer(&entry.Field76), 4)
			permit(unsafe.Pointer(&entry.Field75), 4)
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
			var services []cloudReflectServiceR79
			add := func(e cloudReflectServiceR79) { services = append(services, e) }
			mode := want.Input.Mode
			setClass := func(v object.Class) {
				permit(unsafe.Pointer(&cloud.ObjClass), unsafe.Sizeof(cloud.ObjClass))
				cloud.ObjClass = v
			}
			setSub := func(v uint32) {
				permit(unsafe.Pointer(&cloud.ObjSubClass), unsafe.Sizeof(cloud.ObjSubClass))
				cloud.ObjSubClass = object.SubClass(v)
			}
			setOwner := func(v *Object) {
				permit(unsafe.Pointer(&cloud.ObjOwner), unsafe.Sizeof(cloud.ObjOwner))
				cloud.ObjOwner = v
			}
			runtime := PlayerDamageRuntime4E17B0{
				CoopMode: func() bool {
					v := 0
					if want.Input.Coop {
						v = 1
					}
					add(cloudReflectServiceR79{Service: "mode", Flag: 2048, Value: v})
					return want.Input.Coop
				},
				ObserveClear: func(got *Object) {
					if got != target || entry.Field76 != 0 || entry.Field75 != 77 || info.ObserveTarget() != source {
						t.Fatal("observer prefix/cached marker")
					}
					add(cloudReflectServiceR79{Service: "observe-clear"})
					permit(unsafe.Pointer(&info.Field3680), 4)
					permit(unsafe.Pointer(&info.CameraFollowObj), unsafe.Sizeof(info.CameraFollowObj))
					info.Field3680 = 0
					info.CameraFollowObj = nil
					if mode != "observe-nonmissile" {
						setClass(cloud.ObjClass | object.ClassMissile)
					}
					setSub(want.Input.Subclass)
					if mode == "observe-clears-reflect" {
						permit(unsafe.Pointer(&target.Buffs), unsafe.Sizeof(target.Buffs))
						target.Buffs = 0
					}
					if mode == "observe-live-update" {
						permit(unsafe.Pointer(&target.UpdateData), unsafe.Sizeof(target.UpdateData))
						target.UpdateData = unsafe.Pointer(live)
					}
					if mode == "observe-type-change" {
						permit(unsafe.Pointer(&cloud.TypeInd), unsafe.Sizeof(cloud.TypeInd))
						cloud.TypeInd = 1222
					}
				},
				BlockSourceExcluded: func(got *Object) bool {
					if got != cloud || entry.Field76 != 0 || entry.Field75 != 77 {
						t.Fatal("ordinary exclusion prefix")
					}
					return false
				},
				BlockDirection: func(got *Object, pos types.Pointf) bool {
					if got != target {
						t.Fatal("direction identity")
					}
					phase := "ordinary"
					switch pos {
					case types.Ptf(19, 11):
						phase = "reflect"
					case types.Ptf(17, 9):
					default:
						t.Fatal("wrong current/previous direction point")
					}
					if phase == "reflect" {
						if mode == "direction-clears-missile" {
							setClass(cloud.ObjClass &^ object.ClassMissile)
						}
						if mode == "direction-sub40" {
							setSub(64)
						}
					}
					v := 1
					if phase == "reflect" && mode == "reflect-rear" {
						v = 0
					}
					add(cloudReflectServiceR79{Service: "direction", Phase: phase, Value: v})
					return v != 0
				},
				ProjectileReflect: func(got, by *Object) {
					if got != cloud || by != target {
						t.Fatal("reflect identity")
					}
					add(cloudReflectServiceR79{Service: "reflect"})
					switch mode {
					case "reflect-sub40":
						setSub(64)
					case "reflect-sub0":
						setSub(0)
					case "reflect-sub2":
						setSub(2)
					case "reflect-clears-missile":
						setClass(cloud.ObjClass &^ object.ClassMissile)
					}
				},
				ClearOwner: func(got *Object) {
					if got != cloud {
						t.Fatal("clear owner identity")
					}
					add(cloudReflectServiceR79{Service: "clear-owner"})
					setOwner(nil)
					if mode == "clear-sub2" {
						setSub(2)
					}
					if mode == "clear-clears-missile" {
						setClass(cloud.ObjClass &^ object.ClassMissile)
					}
				},
				SetOwner: func(by, got *Object) {
					if by != target || got != cloud {
						t.Fatal("set owner identity")
					}
					add(cloudReflectServiceR79{Service: "set-owner"})
					setOwner(target)
					if mode == "set-sub2" {
						setSub(2)
					}
					if mode == "set-clears-missile" {
						setClass(cloud.ObjClass &^ object.ClassMissile)
					}
				},
				ChangeOwner: func(got, by *Object) {
					if got != cloud || by != target {
						t.Fatal("change owner identity")
					}
					add(cloudReflectServiceR79{Service: "change-owner"})
				},
				Audio: func(sound int, got *Object) {
					if sound != 122 || got != target {
						t.Fatal("Reflect audio")
					}
					add(cloudReflectServiceR79{Service: "audio", Sound: sound})
				},
				QuestMode: func() bool { add(cloudReflectServiceR79{Service: "mode", Flag: 4096}); return false },
				DefaultDamage: func(got, by, w *Object, raw int32, typ object.DamageType) bool {
					if got != target || by != source || w != cloud || raw != want.Input.Raw || typ != object.DamagePoison {
						t.Fatal("raw default identity/amount")
					}
					add(cloudReflectServiceR79{Service: "default", Damage: raw})
					return true
				},
			}
			if partial {
				runtime.CoopMode = nil
			}
			var reason string
			runtime.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
			handled, result := PlayerDamageNative4E17B0(target, source, cloud, want.Input.Raw, object.DamagePoison, runtime)
			if !handled || result != (want.Result != 0) || reason != "" || !slices.Equal(services, want.Services) {
				t.Fatalf("original Reflect mismatch handled/result=%t/%t reason=%q services=%+v want=%+v", handled, result, reason, services, want.Services)
			}
			got := want
			got.Services = services
			got.Marker = entry.Field76
			got.MarkerType = entry.Field75
			got.LiveMarker = live.Field76
			got.LiveMarkerType = live.Field75
			got.CloudClass = uint32(cloud.ObjClass)
			got.CloudSubclass = uint32(cloud.ObjSubClass)
			got.CloudType = cloud.TypeInd
			got.CloudOwner = "source"
			if cloud.ObjOwner == target {
				got.CloudOwner = "target"
			} else if cloud.ObjOwner != source {
				t.Fatal("unexpected owner")
			}
			got.Buffs = uint32(target.Buffs)
			got.InfoFlags = info.Field3680
			got.ObserverCleared = info.CameraFollowObj == nil
			got.TargetLiveUpdate = target.UpdateData == unsafe.Pointer(live)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("original Reflect output got=%+v want=%+v", got, want)
			}
		})
	}
}
