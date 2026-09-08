package legacy

import (
	"testing"

	"github.com/opennox/noxscript/ns/asm"
)

func TestNativeJournalBuiltinsHaveNoLegacyCFallback(t *testing.T) {
	for _, builtin := range []asm.Builtin{
		asm.BuiltinGetQuestStatus,
		asm.BuiltinJournalDelete,
		asm.BuiltinJournalEdit,
	} {
		if _, ok := CallScriptBuiltin(builtin); ok {
			t.Fatalf("builtin %d still has a legacy C fallback", builtin)
		}
	}
}
