package server

import "testing"

func TestTrackedRefreshLastBeforeFrame519710(t *testing.T) {
	update := &PlayerUpdateData{Field67: 80, Player: &Player{PlayerInd: 31}}
	var order []string
	got := TrackedRefreshNative519710(update, TrackedRefreshRuntime519710{
		Count: func(index int32) int32 {
			if index != 31 {
				t.Fatalf("index=%d", index)
			}
			order = append(order, "count")
			update.Player = nil
			return 4
		},
		Frame: func() uint32 {
			order = append(order, "frame")
			// Original instruction739 loads last before instruction740's
			// frame read. Replacing the frame accessor must retain that order.
			update.Field67 = 100
			return 100
		},
	})
	if got != 1 || len(order) != 2 || order[0] != "count" || order[1] != "frame" || update.Field67 != 100 || update.Player != nil {
		t.Fatalf("return=%d order=%v state=%+v", got, order, update)
	}
}
