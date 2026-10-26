package server

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type playerCloudGreatSwordServiceR78 struct {
	Service  string `json:"service"`
	Flag     uint32 `json:"flag"`
	Value    int    `json:"value"`
	Sound    int    `json:"sound"`
	Bits     uint32 `json:"bits"`
	Selected int    `json:"selected"`
	Amount   uint32 `json:"amount_bits32"`
	Damage   int32  `json:"damage"`
}

type playerCloudGreatSwordCaseR78 struct {
	Input struct {
		Name     string `json:"name"`
		State    uint32 `json:"state"`
		Mode     string `json:"mode"`
		Subclass uint32 `json:"subclass"`
		Raw      int32  `json:"raw"`
		Balance  uint32 `json:"balance_bits32"`
		Coop     bool   `json:"coop"`
	} `json:"input"`
	Services         []playerCloudGreatSwordServiceR78 `json:"services"`
	Result           uint32                            `json:"result"`
	Marker           uint32                            `json:"marker"`
	MarkerType       uint32                            `json:"marker_type"`
	CloudClass       uint32                            `json:"cloud_class"`
	CloudSubclass    uint32                            `json:"cloud_subclass"`
	CloudOwner       string                            `json:"cloud_owner"`
	EntryState       uint32                            `json:"entry_state"`
	LiveState        uint32                            `json:"live_state"`
	EntryWeaponFlags uint32                            `json:"entry_weapon_flags"`
	TargetLiveUpdate bool                              `json:"target_live_update"`
	InventoryHead    int                               `json:"inventory_head"`
	ItemFlags        uint32                            `json:"item_flags"`
	SecondFlags      uint32                            `json:"second_flags"`
	Preserved        bool                              `json:"all_nonoutput_frame_bytes_preserved"`
}

func playerCloudGreatSwordAllocateR78[T comparable](t *testing.T) *T {
	t.Helper()
	p, free := alloc.New(*new(T))
	t.Cleanup(free)
	if uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
		t.Fatal("native C allocation is not above4GiB")
	}
	return p
}

// These inputs and complete output transcripts come from1872 independent
// unchanged GAME.EXE runs, sealed before the native production change. Mode,
// direction, projectile/owner/audio/action, integer balance lookup, durability
// and DefaultDamage are declared external services. The original executes
// cached player stance, inventory selection and balance scalar load/product/spill itself.
// The native frames are C-owned above4GiB; every nonoutput byte is checked.
func TestPlayerDamagePlayerCloudGreatSwordOriginal4E17B0Native(t *testing.T) {
	playerCloudGreatSwordOriginalR78(t, false)
}

func TestPlayerDamagePlayerCloudGreatSwordPartialOriginal4E17B0Native(t *testing.T) {
	playerCloudGreatSwordOriginalR78(t, true)
}

func playerCloudGreatSwordOriginalR78(t *testing.T, partial bool) {
	b, err := os.ReadFile("testdata/player_cloud_greatsword_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Original string                         `json:"original_sha256"`
		Cases    []playerCloudGreatSwordCaseR78 `json:"cases"`
	}
	if err := json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Cases) != 1872 {
		t.Fatal("original fixture identity/count")
	}
	for _, want := range fixture.Cases {
		// The historical partial runtime supplies no Coop query. Its remaining
		// callback transcript and complete native outputs still match the
		// original; the full runtime additionally checks the actual query.
		if partial {
			if len(want.Services) == 0 || want.Services[0].Service != "mode" || want.Services[0].Flag != 2048 {
				t.Fatal("original Coop prefix identity")
			}
			want.Services = slices.Clone(want.Services[1:])
		}
		t.Run(want.Input.Name, func(t *testing.T) {
			target, source, cloud := playerCloudGreatSwordAllocateR78[Object](t), playerCloudGreatSwordAllocateR78[Object](t), playerCloudGreatSwordAllocateR78[Object](t)
			entry, live := playerCloudGreatSwordAllocateR78[PlayerUpdateData](t), playerCloudGreatSwordAllocateR78[PlayerUpdateData](t)
			sourceUpdate := playerCloudGreatSwordAllocateR78[MonsterUpdateData](t)
			info := playerCloudGreatSwordAllocateR78[Player](t)
			item, second := playerCloudGreatSwordAllocateR78[Object](t), playerCloudGreatSwordAllocateR78[Object](t)
			hp1, hp2 := playerCloudGreatSwordAllocateR78[HealthData](t), playerCloudGreatSwordAllocateR78[HealthData](t)
			*target = Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(entry), InvFirstItem: item}
			*source = Object{ObjClass: object.ClassMonster, ObjSubClass: 0x10, UpdateData: unsafe.Pointer(sourceUpdate)}
			*cloud = Object{TypeInd: 1174, ObjClass: object.Class(0x110000), PrevPos: types.Ptf(17, 9), ObjOwner: source}
			*entry = PlayerUpdateData{Field57: math.Float32bits(.5), Field21: math.Float32bits(.25), Field76: 99, Field75: 77, State: PlayerState(want.Input.State), Player: info}
			*live = PlayerUpdateData{Field76: 31, Field75: 33, State: PlayerState0, Player: info}
			info.WeaponEquip = 0x400
			*hp1 = HealthData{Cur: 50, Max: 50}
			*hp2 = *hp1
			*item = Object{ObjClass: object.ClassWeapon, ObjSubClass: 0x400, ObjFlags: object.FlagEquipped, InvNextItem: second, HealthData: hp1}
			*second = Object{ObjClass: object.ClassWeapon, ObjSubClass: 0x400, ObjFlags: object.FlagEquipped, HealthData: hp2}
			type frame struct {
				p       unsafe.Pointer
				before  []byte
				allowed []bool
			}
			var frames []*frame
			watch := func(p unsafe.Pointer, n uintptr) {
				frames = append(frames, &frame{p: p, before: slices.Clone(unsafe.Slice((*byte)(p), int(n))), allowed: make([]bool, int(n))})
			}
			for _, p := range []*Object{target, source, cloud, item, second} {
				watch(unsafe.Pointer(p), unsafe.Sizeof(*p))
			}
			for _, p := range []*PlayerUpdateData{entry, live} {
				watch(unsafe.Pointer(p), unsafe.Sizeof(*p))
			}
			watch(unsafe.Pointer(sourceUpdate), unsafe.Sizeof(*sourceUpdate))
			watch(unsafe.Pointer(info), unsafe.Sizeof(*info))
			for _, p := range []*HealthData{hp1, hp2} {
				watch(unsafe.Pointer(p), unsafe.Sizeof(*p))
			}
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
				t.Fatal("mutable output outside owned C frames")
			}
			permit(unsafe.Pointer(&entry.Field76), 4)
			permit(unsafe.Pointer(&entry.Field75), 4)
			var services []playerCloudGreatSwordServiceR78
			add := func(s playerCloudGreatSwordServiceR78) { services = append(services, s) }
			mode := want.Input.Mode
			runtime := PlayerDamageRuntime4E17B0{
				CoopMode: func() bool {
					value := 0
					if want.Input.Coop {
						value = 1
					}
					add(playerCloudGreatSwordServiceR78{Service: "mode", Flag: 2048, Value: value})
					return want.Input.Coop
				},
				BlockSourceExcluded: func(got *Object) bool {
					if got != cloud || entry.Field76 != 0 || entry.Field75 != 77 {
						t.Fatal("exclusion identity/prefix")
					}
					return mode == "excluded-type"
				},
				BlockDirection: func(got *Object, pos types.Pointf) bool {
					if got != target || pos != types.Ptf(17, 9) {
						t.Fatal("facing identity/cached point")
					}
					if mode != "nonmissile" {
						permit(unsafe.Pointer(&cloud.ObjClass), unsafe.Sizeof(cloud.ObjClass))
						cloud.ObjClass |= object.ClassMissile
					}
					permit(unsafe.Pointer(&cloud.ObjSubClass), unsafe.Sizeof(cloud.ObjSubClass))
					cloud.ObjSubClass = object.SubClass(want.Input.Subclass)
					if mode == "live-update" {
						permit(unsafe.Pointer(&target.UpdateData), unsafe.Sizeof(target.UpdateData))
						target.UpdateData = unsafe.Pointer(live)
					}
					if mode == "drop-cached-equipment" {
						permit(unsafe.Pointer(&info.WeaponEquip), 4)
						info.WeaponEquip = 0
					}
					if mode == "state-ready" || mode == "state-unready" {
						permit(unsafe.Pointer(&entry.State), unsafe.Sizeof(entry.State))
						entry.State = PlayerState0
						if mode == "state-ready" {
							entry.State = PlayerState13
						}
					}

					value := 0
					if mode != "rear" {
						value = 1
					}
					add(playerCloudGreatSwordServiceR78{Service: "direction", Value: value})
					return value != 0
				},
				ProjectileReflect: func(got, by *Object) {
					if got != cloud || by != target {
						t.Fatal("reflect identity")
					}
					add(playerCloudGreatSwordServiceR78{Service: "reflect"})
					if mode == "reflect-clears-missile" {
						cloud.ObjClass &^= object.ClassMissile
					}
					if mode == "reflect-subclass2" {
						cloud.ObjSubClass = 2
					}
				},
				ClearOwner: func(got *Object) {
					if got != cloud {
						t.Fatal("clear owner identity")
					}
					add(playerCloudGreatSwordServiceR78{Service: "clear-owner"})
					permit(unsafe.Pointer(&cloud.ObjOwner), unsafe.Sizeof(cloud.ObjOwner))
					cloud.ObjOwner = nil
				},
				SetOwner: func(owner, got *Object) {
					if owner != target || got != cloud {
						t.Fatal("set owner identity")
					}
					add(playerCloudGreatSwordServiceR78{Service: "set-owner"})
					permit(unsafe.Pointer(&cloud.ObjOwner), unsafe.Sizeof(cloud.ObjOwner))
					cloud.ObjOwner = target
				},
				Audio: func(sound int, got *Object) {
					if sound != 890 || got != target {
						t.Fatal("block sound identity")
					}
					add(playerCloudGreatSwordServiceR78{Service: "audio", Sound: sound})
				},
				BlockDamagePercent: func() float64 {
					add(playerCloudGreatSwordServiceR78{Service: "balance-lookup", Bits: want.Input.Balance})
					if mode == "balance-replaces-inventory" {
						permit(unsafe.Pointer(&target.InvFirstItem), unsafe.Sizeof(target.InvFirstItem))
						target.InvFirstItem = second
					}
					return float64(math.Float32frombits(want.Input.Balance))
				},
				QuestMode: func() bool { add(playerCloudGreatSwordServiceR78{Service: "mode", Flag: 4096}); return false },
				DefaultDamage: func(got, attacker, weapon *Object, damage int32, typ object.DamageType) bool {
					if got != target || attacker != source || weapon != cloud || damage != want.Input.Raw || typ != object.DamagePoison {
						t.Fatal("raw poison default identity/amount")
					}
					add(playerCloudGreatSwordServiceR78{Service: "default", Damage: damage})
					return true
				},
			}

			runtime.Melee.RandomInt = func(low, high int) int {
				if low != 18 || high != 20 {
					t.Fatal("GreatSword RNG range")
				}
				add(playerCloudGreatSwordServiceR78{Service: "random-state", Value: 18})
				return 18
			}
			runtime.PlayerSetState = func(got *Object, state PlayerState) bool {
				if got != target || (state != PlayerState13 && state != PlayerState18) {
					t.Fatal("player state identity/range")
				}
				add(playerCloudGreatSwordServiceR78{Service: "player-state", Value: int(state)})
				u := got.UpdateDataPlayer()
				permit(unsafe.Pointer(&u.State), unsafe.Sizeof(u.State))
				u.State = state
				return true
			}
			runtime.Melee.CanDamageBlockWeapon = func(got *Object) bool { return got == item || got == second }
			runtime.Melee.DamageBlockWeapon = func(got, by, attacker, weapon *Object, amount float32, typ object.DamageType) bool {
				if (got != item && got != second) || by != target || attacker != source || weapon != cloud || typ != object.DamagePoison {
					t.Fatal("weapon durability identity")
				}
				selected := 1
				if got == second {
					selected = 2
				}
				add(playerCloudGreatSwordServiceR78{Service: "weapon-wear", Selected: selected, Amount: math.Float32bits(amount)})
				if mode == "wear-destroy" || mode == "wear-relinks-inventory" {
					permit(unsafe.Pointer(&got.ObjFlags), unsafe.Sizeof(got.ObjFlags))
					got.ObjFlags |= object.FlagDestroyed
					if mode == "wear-relinks-inventory" {
						permit(unsafe.Pointer(&target.InvFirstItem), unsafe.Sizeof(target.InvFirstItem))
						target.InvFirstItem = second
					}
				}
				return true
			}
			if partial {
				runtime.CoopMode = nil
			}
			t.Cleanup(func() {
				for _, f := range frames {
					actual := slices.Clone(unsafe.Slice((*byte)(f.p), len(f.before)))
					expected := slices.Clone(f.before)
					for i, ok := range f.allowed {
						if ok {
							expected[i] = actual[i]
						}
					}
					if !bytes.Equal(actual, expected) {
						t.Fatal("native nonoutput C-frame bytes changed")
					}
				}
				t.Log("all_native_nonoutput_frames_preserved:true")
			})
			var reason string
			runtime.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			handled, result := PlayerDamageNative4E17B0(target, source, cloud, want.Input.Raw, object.DamagePoison, runtime)
			if !handled || result != (want.Result != 0) || reason != "" || !slices.Equal(services, want.Services) {
				t.Fatalf("original branch handled/result=%t/%t reason=%q services=%+v want=%+v", handled, result, reason, services, want.Services)
			}
			got := want
			got.Services = services
			got.Marker = entry.Field76
			got.MarkerType = entry.Field75
			got.CloudClass = uint32(cloud.ObjClass)
			got.CloudSubclass = uint32(cloud.ObjSubClass)
			got.CloudOwner = "source"
			if cloud.ObjOwner == target {
				got.CloudOwner = "target"
			} else if cloud.ObjOwner != source {
				t.Fatal("unexpected owner")
			}
			got.EntryState = uint32(entry.State)
			got.LiveState = uint32(live.State)
			got.EntryWeaponFlags = info.WeaponEquip
			got.TargetLiveUpdate = target.UpdateData == unsafe.Pointer(live)
			got.InventoryHead = 1
			if target.InvFirstItem == second {
				got.InventoryHead = 2
			} else if target.InvFirstItem != item {
				t.Fatal("unexpected inventory")
			}
			got.ItemFlags = uint32(item.ObjFlags)
			got.SecondFlags = uint32(second.ObjFlags)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("original frame outputs got=%+v want=%+v", got, want)
			}

		})
	}
}
