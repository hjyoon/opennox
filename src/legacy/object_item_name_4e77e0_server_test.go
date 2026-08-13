package legacy

import (
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

type modifierEffPrefix4E77E0 struct {
	name  *byte
	ind   uint32
	desc  *uint16
	sec   *uint16
	ident *uint16
}

func modifierEffForItemNameTest4E77E0(desc, ident *uint16) *server.ModifierEff {
	mod := &server.ModifierEff{}
	prefix := (*modifierEffPrefix4E77E0)(unsafe.Pointer(mod))
	prefix.desc = desc
	prefix.ident = ident
	return mod
}

func TestObjectItemNameNative4E77E0NamedFieldMapping(t *testing.T) {
	stringsByPointer := make(map[*uint16]string)
	newString := func(s string) *uint16 {
		p := new(uint16)
		stringsByPointer[p] = s
		return p
	}
	primary := newString("Primary")
	secondary := newString("Secondary")
	definitionDesc := newString("Sword")
	modifierTwo := newString("Two")
	modifierThree := newString("Three")
	def := &server.Modifier{Desc8: definitionDesc}
	attrs := &server.ModifierInitData{Modifiers: [4]*server.ModifierEff{
		modifierEffForItemNameTest4E77E0(primary, nil),
		modifierEffForItemNameTest4E77E0(secondary, nil),
		modifierEffForItemNameTest4E77E0(modifierTwo, nil),
		modifierEffForItemNameTest4E77E0(nil, modifierThree),
	}}
	obj := &server.Object{
		TypeInd:  0x8001,
		ObjClass: object.ClassWeapon,
		InitData: unsafe.Pointer(attrs),
	}

	oldRuntime := objectItemNameRuntime
	defer func() { objectItemNameRuntime = oldRuntime }()
	var output strings.Builder
	buffer := new(uint16)
	weaponLookups := 0
	objectItemNameRuntime = objectItemNameRuntime4E77E0{
		weaponDef: func(ind uint16) *server.Modifier {
			weaponLookups++
			if ind != 0x8001 {
				t.Fatalf("type index = %#x, want %#x", ind, uint16(0x8001))
			}
			return def
		},
		armorDef: func(uint16) *server.Modifier {
			t.Fatal("weapon object used armor lookup")
			return nil
		},
		buffer: func() *uint16 { return buffer },
		clear:  func() { output.Reset() },
		copy: func(s *uint16) {
			output.Reset()
			output.WriteString(stringsByPointer[s])
		},
		formatNoInfo: func(*uint16, *byte) {
			t.Fatal("definition unexpectedly used NoInfo")
		},
		append: func(s *uint16) { output.WriteString(stringsByPointer[s]) },
		appendSpace: func() {
			output.WriteByte(' ')
		},
	}

	if got := objectItemNameNative4E77E0(obj); got != buffer {
		t.Fatalf("buffer = %p, want %p", got, buffer)
	}
	if got := output.String(); got != "Primary Secondary Sword Two Three" {
		t.Fatalf("output = %q, want composed item name", got)
	}
	if weaponLookups != 1 {
		t.Fatalf("weapon lookups = %d, want 1", weaponLookups)
	}
}

func TestObjectItemNameNative4E77E0NilObjectFault(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil object did not fault at the class load")
		}
	}()
	objectItemNameNative4E77E0(nil)
}
