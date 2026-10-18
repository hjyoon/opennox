package opennox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type lifetimeSnapshotR314 struct {
	live *bool
	id   int
}

func (v lifetimeSnapshotR314) MarshalJSON() ([]byte, error) {
	if !*v.live {
		return nil, errors.New("world lifetime ended before serialization")
	}
	return []byte(fmt.Sprintf(`{"lifetime":%d}`, v.id)), nil
}

// The queue retires each world before returning to the HTTP handler. Retaining
// an object, or even a shallow object slice, instead of encoded bytes fails.
func TestAPISnapshotR314OwnsSerializedWorldBeforeRetirement(t *testing.T) {
	for id := 0; id < 32; id++ {
		live := false
		queue := func(_ context.Context, fn func()) {
			live = true
			fn()
			live = false
		}
		w := httptest.NewRecorder()
		writeJSONRespInLoop(w, httptest.NewRequest(http.MethodGet, "/debug/nox/objects", nil), queue, func() any {
			if !live {
				t.Fatal("world traversal outside server loop")
			}
			return []lifetimeSnapshotR314{{live: &live, id: id}}
		})
		var got []struct {
			Lifetime int `json:"lifetime"`
		}
		if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &got) != nil || len(got) != 1 || got[0].Lifetime != id {
			t.Fatalf("retired lifetime %d response: status=%d body=%s", id, w.Code, w.Body)
		}
		if w.Header().Get("Content-Type") != "application/json" || !strings.HasSuffix(w.Body.String(), "\n") {
			t.Fatalf("JSON response format changed: %v %q", w.Header(), w.Body.String())
		}
	}
}

func TestAPISnapshotR314CancellationBeforeEnqueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := httptest.NewRecorder()
	writeJSONRespInLoop(w, httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx), func(context.Context, func()) {
		t.Fatal("canceled request enqueued work")
	}, func() any { t.Fatal("canceled request traversed world"); return nil })
	if w.Body.Len() != 0 {
		t.Fatalf("canceled response body: %q", w.Body.String())
	}
}

func TestAPISnapshotR314CancellationAfterEnqueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queued := make(chan func(), 1)
	finished := make(chan struct{})
	w := httptest.NewRecorder()
	go func() {
		writeJSONRespInLoop(w, httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx), func(_ context.Context, fn func()) { queued <- fn }, func() any {
			t.Error("canceled queued request traversed world")
			return nil
		})
		close(finished)
	}()
	fn := <-queued
	cancel()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("handler ignored request cancellation")
	}
	fn()
	if w.Body.Len() != 0 {
		t.Fatalf("canceled response body: %q", w.Body.String())
	}
}

// Cancellation may win while encoding is already executing on the loop. The
// completion send must still let that loop proceed after the handler leaves.
func TestAPISnapshotR314CancellationDuringEncodingDoesNotBlockLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	queued := make(chan func(), 1)
	entered := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan struct{})
	loopFinished := make(chan struct{})
	w := httptest.NewRecorder()
	go func() {
		writeJSONRespInLoop(w, httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx), func(_ context.Context, fn func()) { queued <- fn }, func() any {
			close(entered)
			<-release
			return []string{"finished"}
		})
		close(finished)
	}()
	fn := <-queued
	go func() { fn(); close(loopFinished) }()
	<-entered
	cancel()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("handler ignored request cancellation")
	}
	close(release)
	select {
	case <-loopFinished:
	case <-time.After(time.Second):
		t.Fatal("disconnected client blocked server loop")
	}
	if w.Body.Len() != 0 {
		t.Fatalf("canceled response body: %q", w.Body.String())
	}
}

func TestAPISnapshotR314EncodeErrorIsAnHTTPError(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSONRespInLoop(w, httptest.NewRequest(http.MethodGet, "/", nil), func(_ context.Context, fn func()) { fn() }, func() any { return make(chan int) })
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("encode error status=%d body=%s", w.Code, w.Body)
	}
}
