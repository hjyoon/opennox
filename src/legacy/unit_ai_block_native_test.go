package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestBlockAIActionRegistryHasNoRawPE32Callbacks(t *testing.T) {
	for _, typ := range []ai.ActionType{
		ai.ACTION_BLOCK_ATTACK,
		ai.ACTION_BLOCK_FINISH,
		ai.ACTION_WEAPON_BLOCK,
	} {
		registered := server.GetAIAction(typ)
		action, ok := registered.(cgoAIAction)
		if !ok {
			t.Fatalf("registered %s action = %T, want legacy.cgoAIAction", typ, registered)
		}
		if action.start != nil || action.update != nil || action.end != nil || action.cancel != nil {
			t.Fatalf("raw %s callbacks = start:%p update:%p end:%p cancel:%p", typ, action.start, action.update, action.end, action.cancel)
		}
	}
}
