package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy"
)

func TestMapSection426EA0RootBinding(t *testing.T) {
	if legacy.Nox_xxx_mapReadSection_426EA0 == nil {
		t.Fatal("legacy map section callback is not bound")
	}
	if ok, err := legacy.Nox_xxx_mapReadSection_426EA0(nil, "__missing_section__"); ok || err != nil {
		t.Fatalf("unknown section = %v, %v; want false, nil", ok, err)
	}
	var context byte = 0x5a
	seen := false
	section := mapSection{Name: "__map_section_426ea0_probe__"}
	section.Fnc = func(_ *cryptfile.CryptFile, a1 unsafe.Pointer) error {
		seen = true
		if a1 != unsafe.Pointer(&context) || *(*byte)(a1) != 0x5a {
			t.Fatalf("section context = %p; want %p", a1, &context)
		}
		return nil
	}
	noxMapSectByName[section.Name] = &section
	defer delete(noxMapSectByName, section.Name)
	if ok, err := legacy.Nox_xxx_mapReadSection_426EA0(unsafe.Pointer(&context), section.Name); !ok || err != nil || !seen {
		t.Fatalf("known section = %v, %v, seen=%v; want true, nil, true", ok, err, seen)
	}
	if ok, err := legacy.Nox_xxx_mapReadSection_426EA0(nil, "__MAP_SECTION_426EA0_PROBE__"); ok || err != nil {
		t.Fatalf("case-changed section = %v, %v; want false, nil", ok, err)
	}
}
