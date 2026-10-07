package legacy

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

func lockAssertRetired52CE90(t *testing.T, address, next, signature, hash string) {
	t.Helper()
	raw, err := os.ReadFile("GAME4_2.c")
	if err != nil {
		t.Fatal(err)
	}
	_, section, found := strings.Cut(string(raw), "//----- ("+address+") --------------------------------------------------------")
	if !found {
		t.Fatal("original Lock provenance is missing: " + address)
	}
	section, _, found = strings.Cut(section, "//----- ("+next+") --------------------------------------------------------")
	if !found || !strings.Contains(section, "#if 0\n"+signature) {
		t.Fatal("integer-pointer C body is not disabled: " + address)
	}
	_, body, found := strings.Cut(section, "#if 0\n")
	if !found {
		t.Fatal("disabled provenance start is missing")
	}
	body, _, found = strings.Cut(body, "\n#endif")
	if !found || fmt.Sprintf("%x", sha256.Sum256([]byte(strings.TrimSpace(body)))) != hash {
		t.Fatal("retained original C body changed: " + address)
	}
	header, err := os.ReadFile("GAME4_2.h")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(header)) != "53bb95d58f580db49b1939dd9a220a4ed2c5874b0fce06233ea674bfc21f0be4" {
		t.Fatal("original declarations changed while retiring the Lock cluster")
	}
}

func TestLockCastRetiredCEntry52CE90(t *testing.T) {
	lockAssertRetired52CE90(t, "0052CE90", "0052CF90", "int nox_xxx_castLock_52CE90(", "f90efdc6b712d4d79b3acab634af9aeda855623c48158d615dc175c40dc70045")
	raw, err := os.ReadFile("spells.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "C.nox_xxx_castLock_52CE90") {
		t.Fatal("public Lock selector still binds the retired C entry")
	}
}
