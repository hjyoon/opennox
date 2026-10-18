package opennox

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

// These are the registered routes, with no world initialized. A disconnected
// caller must not traverse a world, mutate it, or receive a response at all.
func TestAPIObjectsSnapshotR316CanceledRoutesDoNotAccessWorld(t *testing.T) {
	oldServer := noxServer
	noxServer = &Server{Server: &server.Server{}}
	t.Cleanup(func() { noxServer = oldServer })
	for _, path := range []string{
		"/debug/nox/objects",
		"/debug/nox/objects?ind=17&flags=ACTIVE",
		"/debug/nox/objects/updatable",
		"/debug/nox/objects/non-missiles",
		"/debug/nox/objects/missiles",
		"/debug/nox/objects/index",
		"/debug/nox/teams",
	} {
		t.Run(path, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, path, nil).WithContext(ctx)
			http.DefaultServeMux.ServeHTTP(w, r)
			if w.Body.Len() != 0 || len(w.Header()) != 0 {
				t.Fatalf("canceled route accessed/responded from world: status=%d headers=%v body=%q", w.Code, w.Header(), w.Body.String())
			}
		})
	}
}

func runAPIObjectsRequestR316(t *testing.T, s *Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	w := httptest.NewRecorder()
	finished := make(chan struct{})
	go func() {
		http.DefaultServeMux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil).WithContext(ctx))
		close(finished)
	}()
	for {
		select {
		case <-finished:
			return w
		case <-ctx.Done():
			<-finished
			t.Fatal("diagnostic route did not complete on the server loop")
		default:
			s.LoopSleep(time.Millisecond)
		}
	}
}

func TestAPIObjectsSnapshotR316IndexedFlagsAndErrors(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(1) {
		t.Fatal("cannot initialize native object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	obj := s.Objs.NewObject(&server.ObjectType{})
	obj.Extent = 17
	s.Objs.SetObjects(obj)
	oldServer := noxServer
	noxServer = s
	t.Cleanup(func() { noxServer = oldServer })
	initial := object.FlagBelow | object.FlagEnabled
	for _, tt := range []struct {
		name  string
		query string
		want  object.Flags
		code  int
		body  string
	}{
		{"ordered", "ind=17&flags=ACTIVE,-ACTIVE,ACTIVE,NO_COLLIDE,-ENABLED", object.FlagBelow | object.FlagActive | object.FlagNoCollide, http.StatusOK, ""},
		{"repeated", "ind=17&flags=-BELOW,BELOW,-BELOW", object.FlagEnabled, http.StatusOK, ""},
		{"empty", "ind=17", initial, http.StatusOK, ""},
		{"invalid-atomic", "ind=17&flags=ACTIVE,BOGUS", initial, http.StatusBadRequest, "invalid flag name: \"BOGUS\"\n"},
		{"missing-before-flags", "ind=99&flags=BOGUS", initial, http.StatusFound, "Not found\n"},
		{"bad-index", "ind=nope&flags=BOGUS", initial, http.StatusBadRequest, "Bad index\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			obj.ObjFlags = initial
			w := runAPIObjectsRequestR316(t, s, "/debug/nox/objects?"+tt.query)
			if w.Code != tt.code || obj.Flags() != tt.want {
				t.Fatalf("status=%d flags=%v, want %d %v; body=%q", w.Code, obj.Flags(), tt.code, tt.want, w.Body.String())
			}
			if tt.code != http.StatusOK {
				if w.Body.String() != tt.body || w.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
					t.Fatalf("HTTP error changed: headers=%v body=%q", w.Header(), w.Body.String())
				}
				return
			}
			var got struct {
				Ind   int      `json:"ind"`
				Flags []string `json:"flags"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got.Ind != 17 {
				t.Fatalf("snapshot does not include the ordered update: %#v err=%v body=%s", got, err, w.Body)
			}
			var gotFlags object.Flags
			for _, name := range got.Flags {
				flag, err := object.ParseFlag(name)
				if err != nil {
					t.Fatal(err)
				}
				gotFlags |= flag
			}
			if gotFlags != tt.want {
				t.Fatalf("snapshot flags=%v, want %v", gotFlags, tt.want)
			}
			if w.Header().Get("Content-Type") != "application/json" || !strings.HasSuffix(w.Body.String(), "\n") || !strings.Contains(w.Body.String(), "\n\t\"ind\": 17") {
				t.Fatalf("JSON response format changed: headers=%v body=%q", w.Header(), w.Body.String())
			}
		})
	}
	t.Run("canceled-does-not-mutate", func(t *testing.T) {
		obj.ObjFlags = initial
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		w := httptest.NewRecorder()
		http.DefaultServeMux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/debug/nox/objects?ind=17&flags=ACTIVE", nil).WithContext(ctx))
		if obj.Flags() != initial || w.Body.Len() != 0 || len(w.Header()) != 0 {
			t.Fatalf("canceled flag update: flags=%v headers=%v body=%q", obj.Flags(), w.Header(), w.Body.String())
		}
	})
}

func TestAPIObjectsSnapshotR316QueuedHTTPError(t *testing.T) {
	live := false
	w := httptest.NewRecorder()
	writeJSONRespInLoopWithError(w, httptest.NewRequest(http.MethodGet, "/", nil), func(_ context.Context, fn func()) {
		live = true
		fn()
		live = false
	}, func() (any, error) {
		if !live {
			t.Fatal("error lookup escaped the server loop")
		}
		return nil, apiSnapshotHTTPError{status: http.StatusFound, message: "Not found"}
	})
	if live || w.Code != http.StatusFound || w.Body.String() != "Not found\n" || w.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Fatalf("queued HTTP error response changed: status=%d headers=%v body=%q", w.Code, w.Header(), w.Body.String())
	}
}
