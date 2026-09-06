package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestConfusedAIActionRegistryHasNoRawPE32Callbacks(t *testing.T) {
	registered := server.GetAIAction(ai.ACTION_CONFUSED)
	action, ok := registered.(cgoAIAction)
	if !ok {
		t.Fatalf("registered confused action = %T, want legacy.cgoAIAction", registered)
	}
	if action.start != nil || action.update != nil || action.end != nil || action.cancel != nil {
		t.Fatalf("raw confused callbacks = start:%p update:%p end:%p cancel:%p", action.start, action.update, action.end, action.cancel)
	}
}
