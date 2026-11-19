package server

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/original_item_armor_damage_v1.json.gz
var itemArmorDamageOriginal []byte

type originalItemArmorDamageCase struct {
	Input struct {
		Name, Entry   string
		UnitClass     uint32 `json:"unit_class"`
		Flags         uint32
		Held          bool
		Damage        int32
		DamageType    int32 `json:"damage_type"`
		DefaultResult int32 `json:"default_result"`
		NullItem      bool  `json:"null_item"`
		NullSource    bool  `json:"null_source"`
		NullWeapon    bool  `json:"null_weapon"`
	}
	Observed struct {
		Result  int32
		Default *struct {
			Damage     int32
			DamageType int32 `json:"damage_type"`
		}
	}
}

func TestItemArmorDamageFrozenOriginal(t *testing.T) {
	reader, err := gzip.NewReader(bytes.NewReader(itemArmorDamageOriginal))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	const expectedSHA = "92ab60e9427cd87fc40e527346f51dbe1f1e8bea8b5a19ad2bcea5b97e220e3f"
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != expectedSHA {
		t.Fatalf("stock fixture digest %s, want %s", got, expectedSHA)
	}
	var original struct {
		Rows []originalItemArmorDamageCase
	}
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if len(original.Rows) != 245 {
		t.Fatalf("stock fixture has %d rows, want 245", len(original.Rows))
	}
	item, freeItem := alloc.NewHigh(Object{})
	t.Cleanup(freeItem)
	owner, freeOwner := alloc.NewHigh(Object{})
	t.Cleanup(freeOwner)
	source, freeSource := alloc.NewHigh(Object{})
	t.Cleanup(freeSource)
	weapon, freeWeapon := alloc.NewHigh(Object{})
	t.Cleanup(freeWeapon)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, arg := range []*Object{item, owner, source, weapon} {
			if uintptr(unsafe.Pointer(arg)) <= 0xffffffff {
				t.Fatalf("original comparison requires a native pointer above 4 GiB, got %p", arg)
			}
		}
	}
	for _, row := range original.Rows {
		input, expected := row.Input, row.Observed
		t.Run(input.Name, func(t *testing.T) {
			*item = Object{ObjClass: object.Class(input.UnitClass), Material: uint16(input.Flags)}
			if input.Held {
				item.InvHolder = owner
			}
			target, attack, effective := item, source, weapon
			if input.NullItem {
				target = nil
			}
			if input.NullSource {
				attack = nil
			}
			if input.NullWeapon {
				effective = nil
			}
			calls := 0
			applyDefault := func(obj, gotSource, gotWeapon *Object, amount int32, typ object.DamageType) bool {
				calls++
				want := expected.Default
				if want == nil || obj != target || gotSource != attack || gotWeapon != effective ||
					amount != want.Damage || int32(typ) != want.DamageType {
					t.Fatalf("forwarded %p/%p/%p damage=%d type=%d, want %p/%p/%p %+v",
						obj, gotSource, gotWeapon, amount, typ, target, attack, effective, want)
				}
				return input.DefaultResult != 0
			}
			before := *item
			typ := object.DamageType(input.DamageType)
			var result bool
			switch input.Entry {
			case "ball-damage":
				result = BallDamage4E14A0(target, attack, effective, input.Damage, typ)
			case "weapon-damage":
				result = WeaponDamage4E14B0(target, attack, effective, input.Damage, typ, applyDefault)
			case "armor-damage":
				result = ArmorDamage4E1500(target, attack, effective, input.Damage, typ, applyDefault)
			default:
				t.Fatalf("unknown damage entry %q", input.Entry)
			}
			if result != (expected.Result != 0) || (calls == 1) != (expected.Default != nil) || calls > 1 || *item != before {
				t.Fatalf("result=%t default calls=%d changed item=%t, want result=%d call=%+v",
					result, calls, *item != before, expected.Result, expected.Default)
			}
		})
	}
}
