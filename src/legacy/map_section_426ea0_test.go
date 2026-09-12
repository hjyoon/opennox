package legacy

import (
	"errors"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

func TestMapSection426EA0ViaC(t *testing.T) {
	oldHandler := Nox_xxx_mapReadSection_426EA0
	oldFile := cryptfile.Global()
	defer func() {
		Nox_xxx_mapReadSection_426EA0 = oldHandler
		cryptfile.SetGlobal(oldFile)
	}()
	var seenName string
	var seenContext uintptr
	var seenByte byte
	Nox_xxx_mapReadSection_426EA0 = func(a1 unsafe.Pointer, name string) (bool, error) {
		seenName = name
		seenContext = uintptr(a1)
		seenByte = *(*byte)(a1)
		switch name {
		case "Known":
			return true, nil
		case "Broken":
			return false, errors.New("section read failed")
		default:
			return false, nil
		}
	}
	for _, tc := range []struct {
		name       string
		wantResult int
		wantError  uint32
		wantClose  bool
	}{
		{"Known", 1, 0, false},
		{"known", 0, 0, false},
		{"Broken", 0, 1, true},
	} {
		cryptfile.SetGlobal(&cryptfile.CryptFile{})
		got := mapSectionViaC426EA0(tc.name)
		if got.result != tc.wantResult || got.error != tc.wantError ||
			(cryptfile.Global() == nil) != tc.wantClose {
			t.Fatalf("%q = %+v, closed=%v; want result=%d error=%d closed=%v",
				tc.name, got, cryptfile.Global() == nil, tc.wantResult, tc.wantError, tc.wantClose)
		}
		if seenName != tc.name || seenContext != got.context || seenByte != 0x5a {
			t.Fatalf("%q callback = name %q context %#x byte %#x, C context %#x", tc.name, seenName, seenContext, seenByte, got.context)
		}
		if unsafe.Sizeof(uintptr(0)) > 4 && got.context <= math.MaxUint32 {
			t.Fatalf("C stack context %#x was not above 4GiB", got.context)
		}
	}
}
