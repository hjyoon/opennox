package legacy

import (
	"bytes"
	"testing"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestAIDebugPrintf5341A0VariadicCEntry(t *testing.T) {
	const flag = noxflags.EngineShowAI
	if noxflags.HasEngine(flag) {
		t.Cleanup(func() { noxflags.SetEngine(flag) })
	} else {
		t.Cleanup(func() { noxflags.UnsetEngine(flag) })
	}
	text := memmap.Slice(0x5D4594, 2487996)[:32]
	old := bytes.Clone(text)
	t.Cleanup(func() { copy(text, old) })
	copy(text, "saved\x00")

	noxflags.UnsetEngine(flag)
	cfixture.AIDebugPrintf5341A0(nil, 37)
	if got := alloc.GoString(&text[0]); got != "saved" {
		t.Fatalf("disabled debug changed buffer to %q", got)
	}

	noxflags.SetEngine(flag)
	format := []byte("%04d %s\x00")
	cfixture.AIDebugPrintf5341A0(&format[0], 37)
	if got := alloc.GoString(&text[0]); got != "0037 monster" {
		t.Fatalf("debug buffer = %q, want %q", got, "0037 monster")
	}
}
