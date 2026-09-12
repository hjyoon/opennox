package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterFarMoveToUsesNativeUpdate5445C0(t *testing.T) {
	action, ok := server.GetAIAction(ai.ACTION_FAR_MOVE_TO).(cgoAIAction)
	if !ok {
		t.Fatal("FAR_MOVE_TO registration is missing")
	}
	if action.update != nil {
		t.Fatalf("FAR_MOVE_TO still has a PE32 update callback: %p", action.update)
	}
}
