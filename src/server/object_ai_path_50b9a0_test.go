package server

import (
	"testing"

	"github.com/opennox/libs/types"
)

func TestAIPathMaybeAppendWorkPath50B9A0PreservesStatusGate(t *testing.T) {
	paths := new(serverAIPaths)
	paths.points = []types.Pointf{{X: 1}, {X: 2}, {X: 3}, {X: 4}}
	paths.pointsCnt = len(paths.points)

	out := make([]types.Pointf, 3)
	if got := paths.MaybeAppendWorkPath(out); got != 2 {
		t.Fatalf("truncated point count = %d, want 2", got)
	}
	if paths.PathStatus != 1 {
		t.Fatalf("successful oversized path status = %d, want 1", paths.PathStatus)
	}
	if out[0] != paths.points[0] || out[1] != paths.points[1] || out[2] != (types.Pointf{}) {
		t.Fatalf("truncated path = %v, want first two points and reserved final slot", out)
	}

	paths.PathStatus = 2
	clear(out)
	if got := paths.MaybeAppendWorkPath(out); got != 2 {
		t.Fatalf("failed oversized point count = %d, want 2", got)
	}
	if paths.PathStatus != 2 {
		t.Fatalf("failed path status was overwritten: got %d, want 2", paths.PathStatus)
	}

	paths.PathStatus = 0
	paths.pointsCnt = len(out)
	clear(out)
	if got := paths.MaybeAppendWorkPath(out); got != 2 {
		t.Fatalf("fitting point count = %d, want 2", got)
	}
	if paths.PathStatus != 0 {
		t.Fatalf("fitting path status = %d, want 0", paths.PathStatus)
	}
}
