package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type defaultDamageElementalFrameOriginal struct {
	Before  [4]uint64
	Objects [4]Object
	Attrs   [3]ModifierInitData
	Effects [12]ModifierEff
	Markers [2]byte
	After   [4]uint64
}

func TestDefaultDamageElementalStartupOriginal4E0B30Native(t *testing.T) {
	raw, err := os.ReadFile("testdata/default_damage_elemental_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != "319d65161677bfa85ef8eb4f8c6df0202364340b03a9feb07d1b155b2e0d1475" {
		t.Fatal("changed whole-original elemental damage fixture")
	}
	var fixture struct {
		Original      string `json:"original_sha256"`
		Unchanged     bool   `json:"original_unchanged"`
		Contexts      int    `json:"contexts"`
		CompleteCalls int    `json:"original_complete_calls"`
		Substitutions int    `json:"numeric_or_protection_return_substitutions"`
		Cases         []struct {
			Name       string                           `json:"name"`
			Function   string                           `json:"function"`
			Input      elementalProtectionInputOriginal `json:"input"`
			RawDamage  int32                            `json:"damage"`
			DamageType object.DamageType                `json:"damage_type"`
			Damage     int32                            `json:"clear_damage"`
			Buffs      uint32                           `json:"buffs"`
			Attributed bool                             `json:"attributed"`
			FinalType  uint32                           `json:"damage_type_after"`
			Frame      uint32                           `json:"frame"`
			Position   [2]uint32                        `json:"position_bits"`
			Result     int                              `json:"result"`
			Events     []string                         `json:"events"`
			Preserved  bool                             `json:"all_nonoutput_bytes_preserved"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		!fixture.Unchanged || fixture.Contexts != 622 || fixture.CompleteCalls != 2488 ||
		fixture.Substitutions != 0 || len(fixture.Cases) != 622 {
		t.Fatal("incomplete whole-original elemental damage observations")
	}
	for _, row := range fixture.Cases {
		t.Run(row.Name, func(t *testing.T) {
			frame, free := alloc.New(defaultDamageElementalFrameOriginal{})
			t.Cleanup(free)
			*frame = defaultDamageElementalFrameOriginal{}
			for i := range frame.Before {
				frame.Before[i], frame.After[i] = 0x123456789abcdef0, 0xfedcba9876543210
			}
			v, a := &frame.Objects[0], &frame.Objects[1]
			ctx := row.Input
			engage, other := unsafe.Pointer(&frame.Markers[0]), unsafe.Pointer(&frame.Markers[1])
			enchant, key := ENCHANT_PROTECT_FROM_FIRE, "FireSpellProtection"
			if row.Function == "ElectricProtection" {
				enchant, key = ENCHANT_PROTECT_FROM_ELECTRICITY, "ElectricitySpellProtection"
			} else if row.Function != "FireProtection" {
				t.Fatal("unknown original protection root")
			}
			*v = Object{TypeInd: 714, ObjClass: object.Class(ctx.UnitClass), ObjFlags: 0x1000000,
				Buffs: 1, Field131: 19, Frame134: 77, Pos132: types.Ptf(31, 47), InitData: unsafe.Pointer(&frame.Attrs[2])}
			if ctx.Buff {
				v.Buffs |= uint32(1) << enchant
			}
			v.BuffsPower[enchant] = ctx.Power
			*a = Object{TypeInd: 706, ObjClass: 1572865, ObjSubClass: 1, ObjFlags: 553665028, PrevPos: types.Ptf(17, 19)}
			if len(ctx.Inventory) > 0 {
				v.InvFirstItem = &frame.Objects[2]
			}
			for i := 0; i < 2; i++ {
				item := &frame.Objects[i+2]
				*item = Object{ObjClass: object.Class(ctx.ItemClass), ObjFlags: object.Flags(ctx.ItemFlags), InitData: unsafe.Pointer(&frame.Attrs[i])}
				if i == 0 && len(ctx.Inventory) > 1 {
					item.InvNextItem = &frame.Objects[3]
				}
			}
			for group := 0; group < 3; group++ {
				var values []uint32
				if group == 2 {
					values = ctx.Unit
				} else if group < len(ctx.Inventory) {
					values = ctx.Inventory[group]
				}
				if len(values) > 4 {
					t.Fatal("more than four modifier slots")
				}
				for i, value := range values {
					effect := &frame.Effects[4*group+i]
					effect.Engage112, effect.EngageFloat120 = engage, math.Float32frombits(value)
					if !ctx.Matching {
						effect.Engage112 = other
					}
					frame.Attrs[group].Modifiers[i] = effect
				}
			}
			if unsafe.Sizeof(uintptr(0)) == 8 {
				for _, p := range []unsafe.Pointer{unsafe.Pointer(frame), unsafe.Pointer(v), unsafe.Pointer(a), v.InitData, engage, other} {
					if uintptr(p) <= math.MaxUint32 {
						t.Fatal("C-owned native frame pointer not above 4 GiB")
					}
				}
			}
			view := unsafe.Slice((*byte)(unsafe.Pointer(frame)), int(unsafe.Sizeof(*frame)))
			before := append([]byte(nil), view...)
			t.Cleanup(func() {
				expected := append([]byte(nil), before...)
				objectOffset := unsafe.Offsetof(frame.Objects)
				for _, field := range [][2]uintptr{
					{unsafe.Offsetof(v.Buffs), unsafe.Sizeof(v.Buffs)},
					{unsafe.Offsetof(v.Obj130), unsafe.Sizeof(v.Obj130)},
					{unsafe.Offsetof(v.Field131), unsafe.Sizeof(v.Field131)},
					{unsafe.Offsetof(v.Pos132), unsafe.Sizeof(v.Pos132)},
					{unsafe.Offsetof(v.Frame134), unsafe.Sizeof(v.Frame134)},
				} {
					start, size := objectOffset+field[0], field[1]
					copy(expected[start:start+size], view[start:start+size])
				}
				if !row.Preserved || !bytes.Equal(expected, view) {
					t.Error("native nonoutput object/modifier/guard frame changed")
				} else {
					t.Log("all_native_nonoutput_frames_preserved:true")
				}
			})
			var events []string
			var damages []int32
			protect := func(got *Object) float64 {
				if got != v {
					t.Fatal("wrong protection target")
				}
				events = append(events, "protection")
				return elementalProtectionNative4DFE40(got, engage, func(gotKey string, index int32) float64 {
					if gotKey != key || index != int32(uint32(ctx.Power)-1) {
						t.Fatal("wrong balance key/index")
					}
					events = append(events, fmt.Sprintf("balance:%s:%d", gotKey, index))
					if index == -1 {
						return 0 // Declared zero guard preceding the original table.
					}
					return float64(math.Float32frombits(ctx.BalanceBits))
				}, enchant, key)
			}
			r := DefaultDamageWorldRuntime4E0B30{
				Frame: func() uint32 { return 1400 }, GameplayFlag1: func() bool { return true }, QuestMode: func() bool { return false },
				IsEnemy: func(target, source *Object) bool {
					if target != v || source != a {
						t.Fatal("wrong enemy query")
					}
					events = append(events, "enemy")
					return false
				},
				FireProtection: protect, ElectricProtection: protect,
				Audio: func(id int, target *Object) {
					want := 104
					if row.Function == "ElectricProtection" {
						want = 108
					}
					if target != v || id != want {
						t.Fatal("wrong protection sound")
					}
					events = append(events, "audio")
				},
				BuffOff: func(target *Object, enchant EnchantID) {
					if target != v || enchant != 0 {
						t.Fatal("wrong invisibility removal")
					}
					v.Buffs &^= 1
					events = append(events, "buff")
				},
				MonsterHasHitSound: func(*Object) bool { t.Fatal("unexpected monster sound lookup"); return false },
				DefaultDamageSound: func(target, source *Object) {
					if target != v || source != a {
						t.Fatal("wrong damage sound")
					}
					events = append(events, "sound")
				},
				DamageClear: func(target *Object, damage int32) {
					if target != v {
						t.Fatal("wrong clear target")
					}
					damages = append(damages, damage)
					events = append(events, "clear")
					unitDamageClearNative4EE5E0(target, damage, unitDamageClearNativeDeps4EE5E0{
						engineFlag: func(uint32) int32 { t.Fatal("nil health reached engine flags"); return 0 },
						setHP:      func(*Object, uint16) { t.Fatal("nil health wrote HP") },
					})
				},
				Unsupported: func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Error(why) },
			}
			result := DefaultDamageWorld4E0B30(v, a, a, row.RawDamage, row.DamageType, r)
			if result != (row.Result != 0) || !reflect.DeepEqual(damages, []int32{row.Damage}) || !reflect.DeepEqual(events, row.Events) {
				t.Errorf("native result/damage/events=%t/%v/%v, original=%d/%d/%v", result, damages, events, row.Result, row.Damage, row.Events)
			}
			wantSource := (*Object)(nil)
			if row.Attributed {
				wantSource = a
			}
			if v.Obj130 != wantSource || v.Buffs != row.Buffs || v.Field131 != row.FinalType || v.Frame134 != row.Frame ||
				(math.Float32bits(v.Pos132.X) != row.Position[0] || math.Float32bits(v.Pos132.Y) != row.Position[1]) {
				t.Error("native elemental damage metadata differs from original")
			}
		})
	}
}
