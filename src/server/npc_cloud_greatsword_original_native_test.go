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

type cloudGreatSwordServiceR77 struct {
	Service  string `json:"service"`
	Flag     uint32 `json:"flag"`
	Value    int    `json:"value"`
	Sound    int    `json:"sound"`
	Bits     uint32 `json:"bits"`
	Selected int    `json:"selected"`
	Amount   uint32 `json:"amount_bits32"`
	Damage   int32  `json:"damage"`
}

type cloudGreatSwordCaseR77 struct {
	Input struct {
		Name     string `json:"name"`
		Action   uint32 `json:"action"`
		Mode     string `json:"mode"`
		Subclass uint32 `json:"subclass"`
		Raw      int32  `json:"raw"`
		Balance  uint32 `json:"balance_bits32"`
		Coop     bool   `json:"coop"`
	} `json:"input"`
	Services         []cloudGreatSwordServiceR77 `json:"services"`
	Result           uint32                      `json:"result"`
	Marker           uint32                      `json:"marker"`
	MarkerType       uint32                      `json:"marker_type"`
	CloudClass       uint32                      `json:"cloud_class"`
	CloudSubclass    uint32                      `json:"cloud_subclass"`
	CloudOwner       string                      `json:"cloud_owner"`
	EntryAction      uint32                      `json:"entry_action"`
	LiveAction       uint32                      `json:"live_action"`
	EntryWeaponFlags uint32                      `json:"entry_weapon_flags"`
	TargetLiveUpdate bool                        `json:"target_live_update"`
	InventoryHead    int                         `json:"inventory_head"`
	ItemFlags        uint32                      `json:"item_flags"`
	SecondFlags      uint32                      `json:"second_flags"`
	Preserved        bool                        `json:"all_nonoutput_frame_bytes_preserved"`
}

func cloudGreatSwordAllocateR77[T comparable](t *testing.T) *T {
	t.Helper()
	p, free := alloc.New(*new(T))
	t.Cleanup(free)
	if uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
		t.Fatal("native C allocation is not above4GiB")
	}
	return p
}

// These inputs and complete output transcripts come from1782 independent
// unchanged GAME.EXE runs, sealed before the native production change. Mode,
// direction, projectile/owner/audio/action, integer balance lookup, durability
// and DefaultDamage are declared external services. The original executes
// readiness, inventory selection, balance scalar load/multiply/spill itself.
// The native frames are C-owned above4GiB; every nonoutput byte is checked.
func TestPlayerDamageNPCCloudGreatSwordOriginal4E17B0Native(t *testing.T) {
	cloudGreatSwordOriginalR77(t, false)
}

func TestPlayerDamageNPCCloudGreatSwordPartialOriginal4E17B0Native(t *testing.T) {
	cloudGreatSwordOriginalR77(t, true)
}

func cloudGreatSwordOriginalR77(t *testing.T, partial bool) {
	b, err := os.ReadFile("testdata/npc_cloud_greatsword_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Original string                   `json:"original_sha256"`
		Cases    []cloudGreatSwordCaseR77 `json:"cases"`
	}
	if err := json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || len(fixture.Cases) != 1782 {
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
			target, source, cloud := cloudGreatSwordAllocateR77[Object](t), cloudGreatSwordAllocateR77[Object](t), cloudGreatSwordAllocateR77[Object](t)
			entry, live, sourceUpdate := cloudGreatSwordAllocateR77[MonsterUpdateData](t), cloudGreatSwordAllocateR77[MonsterUpdateData](t), cloudGreatSwordAllocateR77[MonsterUpdateData](t)
			item, second := cloudGreatSwordAllocateR77[Object](t), cloudGreatSwordAllocateR77[Object](t)
			hp1, hp2 := cloudGreatSwordAllocateR77[HealthData](t), cloudGreatSwordAllocateR77[HealthData](t)
			*target = Object{ObjClass: object.ClassMonster, ObjSubClass: 0x10, UpdateData: unsafe.Pointer(entry), InvFirstItem: item}
			*source = Object{ObjClass: object.ClassMonster, ObjSubClass: 0x10, UpdateData: unsafe.Pointer(sourceUpdate)}
			*cloud = Object{TypeInd: 1174, ObjClass: object.Class(0x110000), PrevPos: types.Ptf(17, 9), ObjOwner: source}
			*entry = MonsterUpdateData{Field518: math.Float32bits(.5), Field1: math.Float32bits(.25), Field547: 99, Field546: 77, WeaponEquipFlags: 0x400}
			entry.AIStack[0].Action = want.Input.Action
			*live = MonsterUpdateData{Field547: 31, Field546: 33}
			live.AIStack[0].Action = 3
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
			for _, p := range []*MonsterUpdateData{entry, live, sourceUpdate} {
				watch(unsafe.Pointer(p), unsafe.Sizeof(*p))
			}
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
			permit(unsafe.Pointer(&entry.Field547), 4)
			permit(unsafe.Pointer(&entry.Field546), 4)
			var services []cloudGreatSwordServiceR77
			add := func(s cloudGreatSwordServiceR77) { services = append(services, s) }
			mode := want.Input.Mode
			runtime := PlayerDamageRuntime4E17B0{
				CoopMode: func() bool {
					value := 0
					if want.Input.Coop {
						value = 1
					}
					add(cloudGreatSwordServiceR77{Service: "mode", Flag: 2048, Value: value})
					return want.Input.Coop
				},
				BlockSourceExcluded: func(got *Object) bool {
					if got != cloud || entry.Field547 != 0 || entry.Field546 != 77 {
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
						permit(unsafe.Pointer(&entry.WeaponEquipFlags), 4)
						entry.WeaponEquipFlags = 0
					}
					value := 0
					if mode != "rear" {
						value = 1
					}
					add(cloudGreatSwordServiceR77{Service: "direction", Value: value})
					return value != 0
				},
				ProjectileReflect: func(got, by *Object) {
					if got != cloud || by != target {
						t.Fatal("reflect identity")
					}
					add(cloudGreatSwordServiceR77{Service: "reflect"})
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
					add(cloudGreatSwordServiceR77{Service: "clear-owner"})
					permit(unsafe.Pointer(&cloud.ObjOwner), unsafe.Sizeof(cloud.ObjOwner))
					cloud.ObjOwner = nil
				},
				SetOwner: func(owner, got *Object) {
					if owner != target || got != cloud {
						t.Fatal("set owner identity")
					}
					add(cloudGreatSwordServiceR77{Service: "set-owner"})
					permit(unsafe.Pointer(&cloud.ObjOwner), unsafe.Sizeof(cloud.ObjOwner))
					cloud.ObjOwner = target
				},
				Audio: func(sound int, got *Object) {
					if sound != 890 || got != target {
						t.Fatal("block sound identity")
					}
					add(cloudGreatSwordServiceR77{Service: "audio", Sound: sound})
				},
				BlockDamagePercent: func() float64 {
					add(cloudGreatSwordServiceR77{Service: "balance-lookup", Bits: want.Input.Balance})
					if mode == "balance-replaces-inventory" {
						permit(unsafe.Pointer(&target.InvFirstItem), unsafe.Sizeof(target.InvFirstItem))
						target.InvFirstItem = second
					}
					return float64(math.Float32frombits(want.Input.Balance))
				},
				QuestMode: func() bool { add(cloudGreatSwordServiceR77{Service: "mode", Flag: 4096}); return false },
				DefaultDamage: func(got, attacker, weapon *Object, damage int32, typ object.DamageType) bool {
					if got != target || attacker != source || weapon != cloud || damage != want.Input.Raw || typ != object.DamagePoison {
						t.Fatal("raw poison default identity/amount")
					}
					add(cloudGreatSwordServiceR77{Service: "default", Damage: damage})
					return true
				},
			}
			runtime.Melee.MonsterBlockAction = func(got *Object) {
				if got != target {
					t.Fatal("block action identity")
				}
				add(cloudGreatSwordServiceR77{Service: "block-action"})
				u := got.UpdateDataMonster()
				permit(unsafe.Pointer(&u.AIStack[0].Action), 4)
				u.AIStack[0].Action = 23
			}
			runtime.Melee.MonsterPopBlockAction = func(got *Object) {
				if got != target {
					t.Fatal("pop action identity")
				}
				add(cloudGreatSwordServiceR77{Service: "pop-action"})
				u := got.UpdateDataMonster()
				permit(unsafe.Pointer(&u.AIStack[0].Action), 4)
				u.AIStack[0].Action = want.Input.Action
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
				add(cloudGreatSwordServiceR77{Service: "weapon-wear", Selected: selected, Amount: math.Float32bits(amount)})
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
			var reason string
			runtime.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
			handled, result := PlayerDamageNative4E17B0(target, source, cloud, want.Input.Raw, object.DamagePoison, runtime)
			if !handled || result != (want.Result != 0) || reason != "" || !slices.Equal(services, want.Services) {
				t.Fatalf("original branch handled/result=%t/%t reason=%q services=%+v want=%+v", handled, result, reason, services, want.Services)
			}
			got := want
			got.Services = services
			got.Marker = entry.Field547
			got.MarkerType = entry.Field546
			got.CloudClass = uint32(cloud.ObjClass)
			got.CloudSubclass = uint32(cloud.ObjSubClass)
			got.CloudOwner = "source"
			if cloud.ObjOwner == target {
				got.CloudOwner = "target"
			} else if cloud.ObjOwner != source {
				t.Fatal("unexpected owner")
			}
			got.EntryAction = entry.AIStack[0].Action
			got.LiveAction = live.AIStack[0].Action
			got.EntryWeaponFlags = entry.WeaponEquipFlags
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
		})
	}
}
