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

type sharedReflectServiceR82 struct {
	Service string `json:"service"`
	Phase   string `json:"phase"`
	Flag    uint32 `json:"flag"`
	Value   int    `json:"value"`
	Sound   int    `json:"sound"`
	Damage  int32  `json:"damage"`
}
type sharedReflectCaseR82 struct {
	Input struct {
		Name       string `json:"name"`
		Player     bool   `json:"player"`
		DamageType int32  `json:"damage_type"`
		Mode       string `json:"mode"`
		Subclass   uint32 `json:"subclass"`
		Raw        int32  `json:"raw"`
		Coop       bool   `json:"coop"`
	} `json:"input"`
	Services         []sharedReflectServiceR82 `json:"services"`
	Result           uint32                    `json:"result"`
	Marker           uint32                    `json:"marker"`
	MarkerType       uint32                    `json:"marker_type"`
	LiveMarker       uint32                    `json:"live_marker"`
	LiveMarkerType   uint32                    `json:"live_marker_type"`
	CloudClass       uint32                    `json:"cloud_class"`
	CloudSubclass    uint32                    `json:"cloud_subclass"`
	CloudType        uint16                    `json:"cloud_type"`
	CloudOwner       string                    `json:"cloud_owner"`
	Buffs            uint32                    `json:"buffs"`
	InfoFlags        uint32                    `json:"info_flags"`
	ObserverCleared  bool                      `json:"observer_cleared"`
	TargetLiveUpdate bool                      `json:"target_live_update"`
	Preserved        bool                      `json:"all_nonoutput_frame_bytes_preserved"`
}

func TestPlayerDamageSharedReflectLiveOwnerOriginal4E17B0Native(t *testing.T) {
	sharedReflectOriginalR82(t, false)
}
func TestPlayerDamageSharedReflectLiveOwnerPartialOriginal4E17B0Native(t *testing.T) {
	sharedReflectOriginalR82(t, true)
}

// Independent unchanged-original160contexts for admitted bare missiles.
// Constant facing, callback effects and final frames are compared; direction
// callee reads and marker timing inside that callee are outside this fixture. Callback services are declared; every native C frame is above4GiB.
// Cleanup validates nonoutput bytes even at the old unsupported boundary.
func sharedReflectOriginalR82(t *testing.T, partial bool) {
	b, err := os.ReadFile("testdata/shared_reflect_live_owner_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Original string                 `json:"original_sha256"`
		Cases    []sharedReflectCaseR82 `json:"cases"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if f.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(f.Cases) != 160 {
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
			npc := playerCloudGreatSwordAllocateR78[MonsterUpdateData](t)
			*npc = MonsterUpdateData{Field547: 99, Field546: 77}
			*target = Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(entry), Buffs: 1 << 27}
			if !want.Input.Player {
				*target = Object{ObjClass: object.ClassMonster, ObjSubClass: 16, UpdateData: unsafe.Pointer(npc), Buffs: 1 << 27}
			}
			*source = Object{ObjClass: object.ClassMonster, ObjSubClass: 0x10, UpdateData: unsafe.Pointer(su)}
			*cloud = Object{TypeInd: 1174, ObjClass: object.ClassMissile, ObjSubClass: object.SubClass(want.Input.Subclass), PosVec: types.Ptf(19, 11), PrevPos: types.Ptf(17, 9), ObjOwner: source}
			*entry = PlayerUpdateData{Player: info, State: PlayerState13, Field76: 99, Field75: 77}
			*live = PlayerUpdateData{Player: info, Field76: 31, Field75: 33}
			info.Field3680 = 0
			info.CameraFollowObj = nil
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
			watch(unsafe.Pointer(npc), unsafe.Sizeof(*npc))
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
			permit(unsafe.Pointer(&npc.Field547), 4)
			permit(unsafe.Pointer(&npc.Field546), 4)
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
			var services []sharedReflectServiceR82
			add := func(e sharedReflectServiceR82) { services = append(services, e) }
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
					add(sharedReflectServiceR82{Service: "mode", Flag: 2048, Value: v})
					return want.Input.Coop
				},
				ObserveClear:        func(*Object) { t.Fatal("unpossessed Reflect entry invoked observer clear") },
				BlockSourceExcluded: func(*Object) bool { t.Fatal("reflected hit entered ordinary exclusion"); return false },
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
					add(sharedReflectServiceR82{Service: "direction", Phase: phase, Value: v})
					return v != 0
				},
				ProjectileReflect: func(got, by *Object) {
					if got != cloud || by != target {
						t.Fatal("reflect identity")
					}
					add(sharedReflectServiceR82{Service: "reflect"})
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
					add(sharedReflectServiceR82{Service: "clear-owner"})
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
					add(sharedReflectServiceR82{Service: "set-owner"})
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
					add(sharedReflectServiceR82{Service: "change-owner"})
				},
				PointFX: func(sound int, pos types.Pointf) {
					if sound != 132 || pos != target.PosVec {
						t.Fatal("original Reflect point arguments")
					}
					add(sharedReflectServiceR82{Service: "point-fx", Sound: sound})
				},
				Audio: func(sound int, got *Object) {
					if sound != 122 || got != target {
						t.Fatal("Reflect audio")
					}
					add(sharedReflectServiceR82{Service: "audio", Sound: sound})
				},
				QuestMode: func() bool { add(sharedReflectServiceR82{Service: "mode", Flag: 4096}); return false },
				DefaultDamage: func(got, by, w *Object, raw int32, typ object.DamageType) bool {
					if got != target || by != source || w != cloud || raw != want.Input.Raw || int32(typ) != want.Input.DamageType {
						t.Fatal("raw default identity/amount")
					}
					add(sharedReflectServiceR82{Service: "default", Damage: raw})
					return true
				},
			}
			if partial {
				runtime.CoopMode = nil
			}
			var reason string
			runtime.Unsupported = func(s string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = s }
			handled, result := PlayerDamageNative4E17B0(target, source, cloud, want.Input.Raw, object.DamageType(want.Input.DamageType), runtime)
			if !handled || result != (want.Result != 0) || reason != "" || !slices.Equal(services, want.Services) {
				t.Fatalf("original Reflect mismatch handled/result=%t/%t reason=%q services=%+v want=%+v", handled, result, reason, services, want.Services)
			}
			got := want
			got.Services = services
			got.Marker = entry.Field76
			got.MarkerType = entry.Field75
			if !want.Input.Player {
				got.Marker, got.MarkerType = npc.Field547, npc.Field546
			}
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

// A partial runtime can discover a newly required owner service only after
// reflection. Keep its completed prefix/effects and reject at that live use.
func TestPlayerDamageSharedReflectLiveOwnerMissingLateService(t *testing.T) {
	for _, missing := range []string{"clear", "set", "change"} {
		t.Run(missing, func(t *testing.T) {
			target := playerCloudGreatSwordAllocateR78[Object](t)
			source := playerCloudGreatSwordAllocateR78[Object](t)
			attack := playerCloudGreatSwordAllocateR78[Object](t)
			update := playerCloudGreatSwordAllocateR78[PlayerUpdateData](t)
			*target = Object{ObjClass: object.ClassPlayer, Buffs: 1 << 27, UpdateData: unsafe.Pointer(update)}
			*source = Object{ObjClass: object.ClassMonster}
			*attack = Object{ObjClass: object.ClassMissile, ObjSubClass: 64, ObjOwner: source}
			if missing == "change" {
				attack.ObjSubClass = 0
			}
			var events []string
			var reason string
			r := PlayerDamageRuntime4E17B0{
				playerPrefix: &playerDamagePrefix4E17B0{update: update},
				BlockDirection: func(*Object, types.Pointf) bool {
					events = append(events, "direction")
					return true
				},
				ProjectileReflect: func(got, by *Object) {
					if got != attack || by != target {
						t.Fatal("reflection identity")
					}
					events = append(events, "reflect")
					attack.ObjSubClass = 0
					if missing == "change" {
						attack.ObjSubClass = 2
					}
				},
				ClearOwner: func(got *Object) {
					events = append(events, "clear")
					got.ObjOwner = nil
				},
				SetOwner: func(owner, got *Object) {
					events = append(events, "set")
					got.ObjOwner = owner
				},
				Audio: func(int, *Object) { t.Fatal("missing live service reached audio") },
				Unsupported: func(why string, got, src, weapon *Object, amount int32, typ object.DamageType) {
					if got != target || src != source || weapon != attack || amount != 7 || typ != object.DamageImpact {
						t.Fatal("unsupported live reflection identity")
					}
					reason = why
					events = append(events, "unsupported")
				},
			}
			want := []string{"direction", "reflect", "unsupported"}
			wantReason := "missing live Reflect Shield owner service"
			if missing == "clear" {
				r.ClearOwner = nil
			} else if missing == "set" {
				r.SetOwner = nil
			} else {
				want = []string{"direction", "reflect", "clear", "set", "unsupported"}
				wantReason = "missing live Reflect Shield change-owner service"
			}
			applicable, handled, result := playerDamageReflectShield4E17B0(target, source, attack, 7, object.DamageImpact, r)
			if !applicable || handled || result || reason != wantReason || !slices.Equal(events, want) {
				t.Fatalf("live service boundary=%t/%t/%t reason=%q events=%v", applicable, handled, result, reason, events)
			}
		})
	}
}
