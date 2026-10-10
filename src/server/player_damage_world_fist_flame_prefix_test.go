package server

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

// Independent unchanged PE32 root/carry/owner/FISTP outputs. DefaultDamage,
// armor wear, facing, flags, buffs and audio are declared external services;
// this seal is a prefix comparison, not a whole Windows gameplay claim.
func TestPlayerDamageWorldFist4E17B0OriginalPrefixTranscript(t *testing.T) {
	got := nonunitDamagePrefixTranscript4E17B0(t, 1)
	const want = "edb5fb672123a1c171647968296512acf69f6d496544924dc332bebc6a1a83f7"
	if sum := fmt.Sprintf("%x", sha256.Sum256(got)); sum != want || len(got) != 59708 {
		t.Fatalf("1164 original fist prefix rows bytes=%d SHA=%s want=%s", len(got), sum, want)
	}
	t.Logf("1164 unchanged PE32 fist prefix rows bytes=%d SHA=%s", len(got), want)
}

func TestPlayerDamageImaginaryFlame4E17B0OriginalPrefixTranscript(t *testing.T) {
	got := nonunitDamagePrefixTranscript4E17B0(t, 2)
	const want = "dcca890db4bcb542d9699c856bd027fd864646841af8b972ba12286401288a79"
	if sum := fmt.Sprintf("%x", sha256.Sum256(got)); sum != want || len(got) != 58296 {
		t.Fatalf("1164 original flame prefix rows bytes=%d SHA=%s want=%s", len(got), sum, want)
	}
	t.Logf("1164 unchanged PE32 flame prefix rows bytes=%d SHA=%s", len(got), want)
}
