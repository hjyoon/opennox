package legacy

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestToxicCloudCastRetiredCEntry52DB60(t *testing.T) {
	raw, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	_, section, found := strings.Cut(string(raw), "//----- (0052DB60) --------------------------------------------------------")
	if !found {
		t.Fatal("original ToxicCloud provenance is missing")
	}
	section, _, found = strings.Cut(section, "//----- (0052DC80) --------------------------------------------------------")
	if !found || !strings.Contains(section, "#if 0\nint nox_xxx_castToxicCloud_52DB60(") {
		t.Fatal("unsafe integer-pointer ToxicCloud entry is still enabled")
	}
	_, body, found := strings.Cut(section, "#if 0\n")
	if !found {
		t.Fatal("disabled ToxicCloud provenance start is missing")
	}
	body, _, found = strings.Cut(body, "\n#endif")
	if !found || fmt.Sprintf("%x", sha256.Sum256([]byte(strings.TrimSpace(body)))) != "ce1f2a0717c4e5543cf05b8dff6130ac543279b0b5de4997081f08138aed171e" {
		t.Fatal("retained original ToxicCloud C body changed")
	}
	header, err := os.ReadFile("GAME4_2.h")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(header)) != "53bb95d58f580db49b1939dd9a220a4ed2c5874b0fce06233ea674bfc21f0be4" {
		t.Fatal("original C declarations changed while retiring ToxicCloud")
	}
	entry, err := os.ReadFile("spells.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(entry), "C.nox_xxx_castToxicCloud_52DB60") {
		t.Fatal("public ToxicCloud selector still binds the retired C function")
	}
}
