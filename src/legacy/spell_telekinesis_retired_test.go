package legacy

import (
	"os"
	"strings"
	"testing"
)

func TestTelekinesisCastRetiredCEntry52D330(t *testing.T) {
	raw, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	_, body, found := strings.Cut(string(raw), "//----- (0052D330) --------------------------------------------------------")
	if !found {
		t.Fatal("original Telekinesis provenance is missing")
	}
	body, _, found = strings.Cut(body, "//----- (0052D3C0) --------------------------------------------------------")
	if !found || !strings.Contains(body, "#if 0\nint nox_xxx_castTelekinesis_52D330(") || !strings.Contains(body, "\n}\n#endif\n") {
		t.Fatal("unsafe six-int C entry is not retained behind a disabled fence")
	}
	goRaw, err := os.ReadFile("spells.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(goRaw), "C.nox_xxx_castTelekinesis_52D330") {
		t.Fatal("public Telekinesis entry still binds the retired C function")
	}
}
