//go:build darwin && arm64

package legacy

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func armorCoefficientBytes415BD0[T any](value *T) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(value)), int(unsafe.Sizeof(*value)))...)
}

func TestArmorCoefficientLayout415BD0(t *testing.T) {
	class, typ := armorCoefficientCOffsets415BD0()
	if class != unsafe.Offsetof(server.Object{}.ObjClass) || typ != unsafe.Offsetof(server.Object{}.TypeInd) {
		t.Fatal("compiled C and native Go object offsets differ")
	}
	t.Logf("ARMOR_COEFFICIENT_NATIVE_LAYOUT_OK class=%d type=%d coefficient=%d native-pointer-width=%d", class, typ, unsafe.Offsetof(server.Modifier{}.DamageCoeffOrArmor64), unsafe.Sizeof(unsafe.Pointer(nil)))
}

func TestArmorCoefficientHighNative415BD0(t *testing.T) {
	var fixture struct {
		Cases []struct {
			Word uint32 `json:"word32"`
			Wide uint64 `json:"IEEE_default_conversion_bits64"`
		} `json:"cases"`
	}
	data, err := os.ReadFile("testdata/armor_coefficient_415bd0_ieee_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 2560 {
		t.Fatal("independent every-exponent/sign boundary domain")
	}
	item, freeItem := alloc.New(server.Object{})
	*item = server.Object{ObjClass: object.ClassArmor, TypeInd: 0x8001}
	defer freeItem()
	definition, freeDefinition := alloc.New(server.Modifier{})
	defer freeDefinition()
	if uintptr(unsafe.Pointer(item)) <= 0xffffffff || uintptr(unsafe.Pointer(definition)) <= 0xffffffff {
		t.Fatal("actual C object/definition must exceed 4GiB")
	}
	beforeItem := armorCoefficientBytes415BD0(item)
	var reads []armorCoefficientRead415BD0
	lookups := 0
	pop := pushArmorCoefficientObserver415BD0(&armorCoefficientObserver415BD0{
		Read: func(read armorCoefficientRead415BD0) { reads = append(reads, read) },
		Lookup: func(typ uint16) unsafe.Pointer {
			lookups++
			if typ != 0x8001 {
				t.Fatalf("unsigned WORD type = %#x", typ)
			}
			return unsafe.Pointer(definition)
		},
	})
	defer pop()
	calls := 0
	for _, vector := range fixture.Cases {
		definition.DamageCoeffOrArmor64 = math.Float32frombits(vector.Word)
		beforeDefinition := armorCoefficientBytes415BD0(definition)
		for _, call := range []func(unsafe.Pointer) float64{armorCoefficientNative415BD0, armorCoefficientActualC415BD0} {
			reads = nil
			got := math.Float64bits(call(unsafe.Pointer(item)))
			if got != vector.Wide {
				t.Fatalf("native DWORD %#08x -> %#016x want %#016x", vector.Word, got, vector.Wide)
			}
			if len(reads) != 3 || reads[0].Site != 0x415bd4 || reads[0].Width != 4 || reads[1].Site != 0x415bdf || reads[1].Width != 2 || reads[2].Site != 0x415bf0 || reads[2].Width != 4 || reads[2].Word != vector.Word {
				t.Fatal("class/type/definition read order or widths")
			}
			if reads[0].Address != unsafe.Add(unsafe.Pointer(item), unsafe.Offsetof(item.ObjClass)) || reads[1].Address != unsafe.Add(unsafe.Pointer(item), unsafe.Offsetof(item.TypeInd)) || reads[2].Address != unsafe.Add(unsafe.Pointer(definition), unsafe.Offsetof(definition.DamageCoeffOrArmor64)) {
				t.Fatal("full native field addresses")
			}
			calls++
		}
		if math.Float64bits(armorCoefficientTypedCConversion415BD0(unsafe.Pointer(definition))) != vector.Wide {
			t.Fatalf("independent typed C float-to-double %#08x", vector.Word)
		}
		if !bytes.Equal(beforeItem, armorCoefficientBytes415BD0(item)) || !bytes.Equal(beforeDefinition, armorCoefficientBytes415BD0(definition)) {
			t.Fatal("full actual object/definition changed")
		}
	}
	if calls != 5120 || lookups != calls {
		t.Fatal("complete actual Go/C domain")
	}
	t.Logf("ARMOR_COEFFICIENT_HIGH_NATIVE_OK actual-Go-C-calls=%d typed-C-conversions=2560 original-nonfinite-credit=false", calls)
}

func TestArmorCoefficientCachedReads415BD0(t *testing.T) {
	item, freeItem := alloc.New(server.Object{})
	defer freeItem()
	definition, freeDefinition := alloc.New(server.Modifier{})
	defer freeDefinition()
	for _, call := range []func(unsafe.Pointer) float64{armorCoefficientNative415BD0, armorCoefficientActualC415BD0} {
		item.ObjClass, item.TypeInd, definition.DamageCoeffOrArmor64 = object.ClassArmor, 0xffff, -0.25
		pop := pushArmorCoefficientObserver415BD0(&armorCoefficientObserver415BD0{
			Read: func(read armorCoefficientRead415BD0) {
				switch read.Site {
				case 0x415bd4:
					item.ObjClass = 0
				case 0x415bdf:
					item.TypeInd = 1
				case 0x415bf0:
					definition.DamageCoeffOrArmor64 = 0.75
				}
			},
			Lookup: func(typ uint16) unsafe.Pointer {
				if typ != 0xffff {
					t.Fatalf("cached unsigned type %#x", typ)
				}
				return unsafe.Pointer(definition)
			},
		})
		got := call(unsafe.Pointer(item))
		pop()
		if got != -0.25 {
			t.Fatal("completed read value must remain cached across observation")
		}
	}
}

type armorCoefficientLegacyServer415BD0 struct {
	Server
	srv *server.Server
}

func (s *armorCoefficientLegacyServer415BD0) S() *server.Server { return s.srv }

func TestArmorCoefficientDefaultLookup415BD0(t *testing.T) {
	srv := new(server.Server)
	previous := GetServer
	GetServer = func() Server { return &armorCoefficientLegacyServer415BD0{srv: srv} }
	defer func() { GetServer = previous }()
	item, freeItem := alloc.New(server.Object{})
	*item = server.Object{ObjClass: object.ClassArmor, TypeInd: 0x8001}
	defer freeItem()
	first, freeFirst := alloc.New(server.Modifier{})
	*first = server.Modifier{TypeInd: 0x8001, DamageCoeffOrArmor64: 0.125}
	defer freeFirst()
	last, freeLast := alloc.New(server.Modifier{})
	*last = server.Modifier{TypeInd: 0x8001, DamageCoeffOrArmor64: 0.75}
	defer freeLast()
	first.Next80 = last
	srv.Modif.Dword_5d4594_251608 = first
	for _, call := range []func(unsafe.Pointer) float64{armorCoefficientNative415BD0, armorCoefficientActualC415BD0} {
		for _, class := range []object.Class{0, object.ClassMonster, object.ClassArmor, object.Class(0xffffffff)} {
			item.ObjClass, item.TypeInd = class, 0x8001
			want := float64(0)
			if uint32(class)&0x02000000 != 0 {
				want = 0.125
			}
			if got := call(unsafe.Pointer(item)); got != want {
				t.Fatalf("actual default linked lookup class=%#x got=%v", class, got)
			}
		}
		item.ObjClass, item.TypeInd = object.ClassArmor, 0xffff
		if got := call(unsafe.Pointer(item)); math.Float64bits(got) != 0 {
			t.Fatal("absent definition returns positive zero")
		}
	}
	t.Log("ARMOR_COEFFICIENT_ACTUAL_DEFAULT_LOOKUP_OK observer=nil Go-and-C-calls=10 class-rejections=4 linked-lookups=6 first-match-and-missing=pass")
}

func TestArmorCoefficientPinnedGoHeap415BD0(t *testing.T) {
	item := &server.Object{ObjClass: object.ClassArmor, TypeInd: 0xffff}
	definition := &server.Modifier{DamageCoeffOrArmor64: math.Float32frombits(1)}
	var pin runtime.Pinner
	pin.Pin(item)
	pin.Pin(definition)
	defer pin.Unpin()
	pop := pushArmorCoefficientObserver415BD0(&armorCoefficientObserver415BD0{Lookup: func(typ uint16) unsafe.Pointer {
		runtime.GC()
		if typ != 0xffff {
			t.Fatal("pinned type")
		}
		return unsafe.Pointer(definition)
	}})
	defer pop()
	for _, call := range []func(unsafe.Pointer) float64{armorCoefficientNative415BD0, armorCoefficientActualC415BD0} {
		if got := call(unsafe.Pointer(item)); got != math.SmallestNonzeroFloat32 {
			t.Fatal("pinned Go-owned object/definition")
		}
	}
	runtime.KeepAlive(item)
	runtime.KeepAlive(definition)
}
