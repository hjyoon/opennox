package server

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type elementalProtectionInputOriginal struct {
	Inventory   [][]uint32 `json:"inventory"`
	Unit        []uint32   `json:"unit"`
	Buff        bool       `json:"buff"`
	Power       uint8      `json:"power"`
	BalanceBits uint32     `json:"balance_bits"`
	UnitClass   uint32     `json:"unit_class"`
	ItemFlags   uint32     `json:"item_flags"`
	ItemClass   uint32     `json:"item_class"`
	Matching    bool       `json:"matching"`
	NullUnit    bool       `json:"null_unit"`
}

type elementalProtectionNativeFrameOriginal struct {
	Before  [4]uint64
	Objects [3]Object
	Attrs   [3]ModifierInitData
	Effects [12]ModifierEff
	Markers [2]byte
	After   [4]uint64
}

func TestElementalProtectionOriginalNative(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native 64-bit protection frame")
	}
	raw, err := os.ReadFile("testdata/elemental_protection_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		OriginalSHA256    string `json:"original_sha256"`
		OriginalUnchanged bool   `json:"original_unchanged"`
		Contexts          int    `json:"contexts"`
		RawRows           int    `json:"raw_rows"`
		Rows              []struct {
			Function, Name, Mode string
			Inputs               elementalProtectionInputOriginal
			ResultBits           uint64 `json:"result_bits"`
			BalanceCalls         []struct {
				Key   string
				Index int32
			} `json:"balance_calls"`
		}
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.OriginalUnchanged || fixture.Contexts != 96 || fixture.RawRows != 384 || len(fixture.Rows) != 384 {
		t.Fatal("incomplete original protection observations")
	}
	count := 0
	for _, fn := range []string{"FireProtection", "ElectricProtection", "PoisonProtection"} {
		t.Run(fn, func(t *testing.T) {
			leaves := 0
			for _, row := range fixture.Rows {
				if row.Function != fn || row.Mode != "actual-startup-inherited" {
					continue
				}
				leaves++
				count++
				t.Run(row.Name, func(t *testing.T) {
					frame, free := alloc.New(elementalProtectionNativeFrameOriginal{})
					defer free()
					for i := range frame.Before {
						frame.Before[i] = 0x123456789abcdef0
						frame.After[i] = 0xfedcba9876543210
					}
					unit := &frame.Objects[0]
					ctx := row.Inputs
					engage := unsafe.Pointer(&frame.Markers[0])
					other := unsafe.Pointer(&frame.Markers[1])
					unit.ObjClass = object.Class(ctx.UnitClass)
					var enchant EnchantID
					var key string
					switch fn {
					case "FireProtection":
						enchant = ENCHANT_PROTECT_FROM_FIRE
						key = "FireSpellProtection"
					case "ElectricProtection":
						enchant = ENCHANT_PROTECT_FROM_ELECTRICITY
						key = "ElectricitySpellProtection"
					case "PoisonProtection":
						enchant = ENCHANT_PROTECT_FROM_POISON
						key = "PoisonSpellProtection"
					}
					if ctx.Buff {
						unit.Buffs = uint32(1) << enchant
					}
					unit.BuffsPower[enchant] = ctx.Power
					unit.InitData = unsafe.Pointer(&frame.Attrs[2])
					if len(ctx.Inventory) > 0 {
						unit.InvFirstItem = &frame.Objects[1]
					}
					for i := 0; i < 2; i++ {
						item := &frame.Objects[i+1]
						item.ObjClass = object.Class(ctx.ItemClass)
						item.ObjFlags = object.Flags(ctx.ItemFlags)
						item.InitData = unsafe.Pointer(&frame.Attrs[i])
						if i == 0 && len(ctx.Inventory) > 1 {
							item.InvNextItem = &frame.Objects[2]
						}
					}
					addresses := []uintptr{uintptr(unsafe.Pointer(frame)), uintptr(unsafe.Pointer(unit)), uintptr(engage), uintptr(other)}
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
						attrs := &frame.Attrs[group]
						addresses = append(addresses, uintptr(unsafe.Pointer(attrs)))
						for i := 0; i < 4; i++ {
							modifier := &frame.Effects[4*group+i]
							addresses = append(addresses, uintptr(unsafe.Pointer(modifier)))
							modifier.Engage112 = engage
							if !ctx.Matching {
								modifier.Engage112 = other
							}
							if i < len(values) {
								modifier.EngageFloat120 = math.Float32frombits(values[i])
								attrs.Modifiers[i] = modifier
							}
						}
					}
					for _, address := range addresses {
						if address <= math.MaxUint32 {
							t.Fatal("native frame pointer not above UINT32_MAX")
						}
					}
					view := unsafe.Slice((*byte)(unsafe.Pointer(frame)), int(unsafe.Sizeof(*frame)))
					before := append([]byte(nil), view...)
					calls := 0
					loadBalance := func(actualKey string, index int32) float64 {
						if calls >= len(row.BalanceCalls) {
							t.Fatal("unexpected balance lookup")
						}
						want := row.BalanceCalls[calls]
						calls++
						if actualKey != key || actualKey != want.Key || index != want.Index {
							t.Fatalf("balance arguments=(%q,%d), original=(%q,%d)", actualKey, index, want.Key, want.Index)
						}
						return float64(math.Float32frombits(ctx.BalanceBits))
					}
					target := unit
					if ctx.NullUnit {
						target = nil
					}
					var result float64
					switch fn {
					case "FireProtection":
						result = fireProtectionNative4DFE40(target, engage, loadBalance)
					case "ElectricProtection":
						result = elementalProtectionNative4DFE40(target, engage, loadBalance, enchant, key)
					case "PoisonProtection":
						result = poisonProtectionNative4E0040(target, poisonProtectionNativeDeps4E0040{poisonProtectEngage: engage, loadBalance: loadBalance})
					}
					if !bytes.Equal(before, view) {
						t.Fatal("native protection frame nonoutput bytes changed")
					}
					if calls != len(row.BalanceCalls) {
						t.Fatalf("balance calls=%d original=%d", calls, len(row.BalanceCalls))
					}
					if got := math.Float64bits(result); got != row.ResultBits {
						t.Fatalf("actual native protection=%016x original=%016x", got, row.ResultBits)
					}
				})
			}
			if leaves != 32 {
				t.Fatalf("original contexts=%d", leaves)
			}
		})
	}
	if count != 96 {
		t.Fatalf("original startup contexts=%d", count)
	}
}
