package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestMovementActionsRegisteredWithoutPE32Callbacks(t *testing.T) {
	for _, typ := range []ai.ActionType{
		ai.ACTION_MOVE_TO,
		ai.ACTION_ROAM,
		ai.ACTION_FLEE,
	} {
		action, ok := server.GetAIAction(typ).(cgoAIAction)
		if !ok {
			t.Fatalf("%s registration type = %T, want cgoAIAction", typ, server.GetAIAction(typ))
		}
		if action.start != nil || action.update != nil || action.end != nil || action.cancel != nil {
			t.Fatalf("%s retained PE32 callbacks: %#v", typ, action)
		}
	}
}
