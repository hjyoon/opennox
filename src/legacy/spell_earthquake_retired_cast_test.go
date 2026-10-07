package legacy

import (
	"os"
	"strings"
	"testing"
)

func TestEarthquakeCastRetiredCEntry52DE40(t *testing.T) {
	raw, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	_, body, found := strings.Cut(string(raw), "//----- (0052DE40) --------------------------------------------------------")
	if !found {
		t.Fatal("original Earthquake cast provenance is missing")
	}
	body, _, found = strings.Cut(body, "//----- (0052DEC0) --------------------------------------------------------")
	if !found || !strings.Contains(body, "#if 0\nint nox_xxx_castEquake_52DE40(") || !strings.Contains(body, "\n}\n#endif\n") {
		t.Fatal("unsafe six-int C cast is not retained behind a disabled fence")
	}
	goRaw, err := os.ReadFile("spells.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(goRaw), "C.nox_xxx_castEquake_52DE40") {
		t.Fatal("public Earthquake entry still binds the retired C cast")
	}
}
