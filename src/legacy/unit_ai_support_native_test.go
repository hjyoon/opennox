package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestSupportAIActionsRegisteredWithoutPE32Callbacks(t *testing.T) {
	for _, typ := range []ai.ActionType{
		ai.ACTION_PICKUP_OBJECT,
		ai.ACTION_RETREAT_TO_MASTER,
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
