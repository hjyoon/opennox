package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestItemPreDamageABI4E13B0(t *testing.T) {
	compiler := strings.Fields(os.Getenv("CC"))
	if len(compiler) == 0 {
		compiler = []string{"cc"}
	}
	binary := filepath.Join(t.TempDir(), "item-pre-damage-abi")
	args := append(compiler[1:], "-std=c11", "-Wall", "-Wextra", "-Werror",
		"testdata/item_pre_damage_4e13b0_abi_test.c", "-o", binary)
	if output, err := exec.Command(compiler[0], args...).CombinedOutput(); err != nil {
		t.Fatalf("public 004E13B0 C ABI: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("native pointer C ABI fixture: %v\n%s", err, output)
	}
}
