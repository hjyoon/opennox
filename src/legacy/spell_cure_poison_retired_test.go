package legacy

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestCurePoisonCastRetiredCEntry52CDB0(t *testing.T) {
	raw, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	_, body, found := strings.Cut(string(raw), "//----- (0052CDB0) --------------------------------------------------------")
	if !found {
		t.Fatal("original CurePoison provenance is missing")
	}
	body, _, found = strings.Cut(body, "//----- (0052CE60) --------------------------------------------------------")
	if !found || !strings.Contains(body, "#if 0\nint nox_xxx_castCurePoison_52CDB0(") || !strings.Contains(body, "\n}\n#endif\n") {
		t.Fatal("unsafe int* target load is not retained behind a disabled fence")
	}
	_, function, found := strings.Cut(body, "#if 0\n")
	if !found {
		t.Fatal("disabled provenance start is missing")
	}
	function, _, found = strings.Cut(function, "\n#endif")
	if !found || fmt.Sprintf("%x", sha256.Sum256([]byte(strings.TrimSpace(function)))) != "718575d317aecc6086ab0c14303029056cd6e071585a4cc372c2418218141740" {
		t.Fatal("retained C body changed while retiring the selector")
	}
	goRaw, err := os.ReadFile("spells.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(goRaw), "C.nox_xxx_castCurePoison_52CDB0") {
		t.Fatal("public CurePoison selector still binds the retired C function")
	}
}
