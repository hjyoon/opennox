package legacy

import (
	"os"
	"strings"
	"testing"
)

func TestEarthquakeDamageRetiredCCallback52DEC0(t *testing.T) {
	raw, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	_, body, found := strings.Cut(string(raw), "//----- (0052DEC0) --------------------------------------------------------")
	if !found {
		t.Fatal("original Earthquake callback provenance is missing")
	}
	body, _, found = strings.Cut(body, "//----- (0052E020) --------------------------------------------------------")
	if !found || !strings.Contains(body, "#if 0\nshort nox_xxx_equakeDamage_52DEC0(") || !strings.Contains(body, "\n}\n#endif\n") {
		t.Fatal("unsafe int-pointer C callback is not retained behind a disabled fence")
	}
	goRaw, err := os.ReadFile("spell_earthquake_52de40.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(goRaw), "C.") || strings.Contains(string(goRaw), "Nox_spells_call_intint6_go") {
		t.Fatal("native Earthquake bridge still binds the retired callback or dispatcher")
	}
}
