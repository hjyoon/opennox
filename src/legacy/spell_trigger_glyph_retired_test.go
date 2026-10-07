package legacy

import (
	"os"
	"strings"
	"testing"
)

func TestTriggerGlyphCastRetiredCEntry52CCD0(t *testing.T) {
	raw, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	_, body, found := strings.Cut(string(raw), "//----- (0052CCD0) --------------------------------------------------------")
	if !found {
		t.Fatal("original TriggerGlyph provenance is missing")
	}
	body, _, found = strings.Cut(body, "//----- (0052CDB0) --------------------------------------------------------")
	if !found || !strings.Contains(body, "#if 0\nint sub_52CCD0(") || !strings.Contains(body, "\n}\n#endif\n") {
		t.Fatal("unsafe three-int C entry is not retained behind a disabled fence")
	}
	goRaw, err := os.ReadFile("spells.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(goRaw), "C.sub_52CCD0") {
		t.Fatal("public TriggerGlyph entry still binds the retired C function")
	}
}
