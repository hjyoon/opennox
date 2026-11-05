package legacy

import (
	"bytes"
	"testing"
)

func TestGeneratorCreditReloadWindowInitial4D61B0(t *testing.T) {
	inputs, initial := generatorCreditInputs4D61B0(t)
	for i, input := range inputs {
		if input.Alias != "counter-is-update-player" {
			continue
		}
		m := newGeneratorCreditMemory4D61B0(t, input, initial[i])
		peAt := int(m.chosenPE-0x21000000) + 4692
		nativeAt := m.chosenNative + int(m.off[18])
		original, native := m.pe[peAt:peAt+5], m.buf[nativeAt:nativeAt+5]
		if !bytes.Equal(native, original) {
			t.Fatalf("complete initial flags/reload window native=%x original=%x", native, original)
		}
	}
}
