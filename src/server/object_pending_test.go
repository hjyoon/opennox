package server

import (
	"testing"
	"time"
)

func pendingByScriptIDWithin(t *testing.T, objs *serverObjects, sid int) *Object {
	t.Helper()
	done := make(chan *Object, 1)
	go func() {
		done <- objs.PendingByScriptID(sid)
	}()
	select {
	case got := <-done:
		return got
	case <-time.After(time.Second):
		t.Fatalf("PendingByScriptID(%d) did not terminate", sid)
		return nil
	}
}

func TestPendingByScriptIDWalksWholeListAndTerminatesOnMiss(t *testing.T) {
	second := &Object{ScriptIDVal: 222}
	first := &Object{ScriptIDVal: 111, ObjNext: second}
	objs := &serverObjects{Pending: first}

	if got := pendingByScriptIDWithin(t, objs, 222); got != second {
		t.Fatalf("second-node lookup = %p, want %p", got, second)
	}
	if got := pendingByScriptIDWithin(t, objs, 333); got != nil {
		t.Fatalf("missing lookup = %p, want nil", got)
	}
}
