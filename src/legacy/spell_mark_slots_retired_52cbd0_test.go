package legacy

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestMarkSlotCastRetiredCEntry52CBD0(t *testing.T) {
	raw, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	_, section, found := strings.Cut(string(raw), "//----- (0052CBD0) --------------------------------------------------------")
	if !found {
		t.Fatal("original named Mark provenance is missing")
	}
	section, _, found = strings.Cut(section, "//----- (0052CCD0) --------------------------------------------------------")
	if !found || !strings.Contains(section, "#if 0\nint sub_52CBD0(") {
		t.Fatal("unsafe integer-pointer Mark 1..4 entry is still enabled")
	}
	_, body, found := strings.Cut(section, "#if 0\n")
	if !found {
		t.Fatal("disabled named Mark provenance start is missing")
	}
	body, _, found = strings.Cut(body, "\n#endif")
	if !found || fmt.Sprintf("%x", sha256.Sum256([]byte(strings.TrimSpace(body)))) != "82918b55e1db6a05b30993d8a76e693b67dc79aa45a5497ce23d4b827edfe487" {
		t.Fatal("retained original Mark 1..4 C body changed")
	}
	assertRetiredSpellHeaderPrototypes(t)
	entry, err := os.ReadFile("spells.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(entry), "C.sub_52CBD0") {
		t.Fatal("public Mark 1..4 selector still binds the retired C function")
	}
}
