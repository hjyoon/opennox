package opennox

import (
	"bytes"
	"context"
	"image"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/libs/wall"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestAPIWorldSnapshotR316CanceledRegisteredRoutes(t *testing.T) {
	oldClient, oldServer := noxClient, noxServer
	noxServer = &Server{Server: &server.Server{}}
	noxClient = &Client{Client: &client.Client{}, srv: noxServer}
	t.Cleanup(func() { noxClient, noxServer = oldClient, oldServer })
	for _, path := range []string{
		"/debug/nox/players",
		"/debug/nox/players?all=1",
		"/debug/nox/walls",
		"/debug/nox/walls?x=-1&y=-1",
		"/debug/nox/walls?x=bad&y=1",
		"/debug/nox/waypoints",
		"/debug/nox/drawable",
	} {
		t.Run(path, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			w := httptest.NewRecorder()
			http.DefaultServeMux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil).WithContext(ctx))
			if w.Body.Len() != 0 || len(w.Header()) != 0 {
				t.Fatalf("disconnected route responded from world: headers=%v body=%q", w.Header(), w.Body.String())
			}
		})
	}
}

type retirementHeaderWriterR316 struct {
	*httptest.ResponseRecorder
	retire func()
}

func (w *retirementHeaderWriterR316) Header() http.Header {
	if w.retire != nil {
		retire := w.retire
		w.retire = nil
		retire()
	}
	return w.ResponseRecorder.Header()
}

func serveAPIWorldRouteR316(t *testing.T, s *Server, w http.ResponseWriter, path string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	finished := make(chan struct{})
	go func() {
		http.DefaultServeMux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil).WithContext(ctx))
		close(finished)
	}()
	for {
		select {
		case <-finished:
			return
		case <-ctx.Done():
			<-finished
			t.Fatal("world snapshot did not finish on the game loop")
		default:
			s.LoopSleep(time.Millisecond)
		}
	}
}

// The original serializer is the byte-format oracle. At the first HTTP header
// access, retire every borrowed native value in the response. Traversing on the
// loop but passing shallow native pointers to the HTTP encoder still fails.
func TestAPIWorldSnapshotR316OwnsBytesBeforeHTTPAndPreservesFormat(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(1) {
		t.Fatal("cannot initialize native object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	obj := s.Objs.NewObject(&server.ObjectType{})
	player := s.Players.ByIndRaw(1)
	inactive := s.Players.ByIndRaw(2)
	if s.Walls.Init() == 0 {
		t.Fatal("cannot initialize native wall registry")
	}
	t.Cleanup(s.Walls.Free)
	w1 := s.Walls.CreateAtGrid(image.Pt(12, 14))
	w2 := s.Walls.CreateAtGrid(image.Pt(34, 36))
	wp1, freeWP1 := alloc.New(server.Waypoint{})
	wp2, freeWP2 := alloc.New(server.Waypoint{})
	t.Cleanup(freeWP1)
	t.Cleanup(freeWP2)
	wp1.WpNext, wp2.WpPrev = wp2, wp1
	s.WPs.List = wp1
	cl := client.NewClient(nil, nil, s.Server)
	t.Cleanup(cl.Close)
	cl.Objs.Init(2)
	t.Cleanup(cl.Objs.Free)
	dr1, dr2 := cl.Objs.New(), cl.Objs.New()
	if w1 == nil || w2 == nil || dr1 == nil || dr2 == nil {
		t.Fatal("native fixture allocation failed")
	}
	dr1.NextPtr, cl.Objs.List1 = dr2, dr1
	oldClient, oldServer := noxClient, noxServer
	noxClient, noxServer = &Client{Client: cl, srv: s}, s
	t.Cleanup(func() { noxClient, noxServer = oldClient, oldServer })
	reset := func() {
		obj.Extent, obj.ObjFlags, obj.PosVec = 17, object.FlagEnabled, types.Ptf(12.5, 31.75)
		player.Active, player.PlayerInd, player.PlayerUnit = 1, 1, obj
		player.SetName("Active Player")
		player.Info().SetName("Original Player")
		inactive.Active = 0
		inactive.SetName("Inactive Slot")
		w1.Flags4, w1.Data, w1.Health7, w1.Tile1 = wall.FlagDoor, unsafe.Pointer(obj), 77, 5
		w2.Health7, w2.Tile1 = 41, 9
		wp1.Index, wp1.PosVec, wp1.Flags = 11, types.Ptf(100.25, 200.5), 1
		wp2.Index, wp2.PosVec, wp2.Flags2 = 12, types.Ptf(310.75, 420.125), 3
		wp1.SetName("First Waypoint")
		wp2.SetName("Second Waypoint")
		dr1.NetCode32, dr1.ObjClass, dr1.ObjFlags, dr1.PosVec = 77, object.ClassMissile, object.FlagActive, image.Pt(210, 320)
		dr2.NetCode32, dr2.ObjClass, dr2.ObjFlags, dr2.PosVec = 78, object.ClassImmobile, object.FlagEnabled, image.Pt(430, 540)
	}
	retire := func() {
		obj.Extent, obj.ObjFlags, obj.PosVec = 18, 0, types.Pointf{}
		player.SetName("Retired Player")
		inactive.SetName("Retired Slot")
		w1.Health7, w1.Tile1, w2.Health7, w2.Tile1 = 0, 0, 0, 0
		wp1.PosVec, wp2.PosVec = types.Pointf{}, types.Pointf{}
		wp1.SetName("Retired Waypoint")
		wp2.SetName("Retired Waypoint")
		dr1.NetCode32, dr1.ObjFlags, dr1.PosVec = 0, 0, image.Point{}
		dr2.NetCode32, dr2.ObjFlags, dr2.PosVec = 0, 0, image.Point{}
	}
	for _, tt := range []struct {
		name  string
		path  string
		value func() any
	}{
		{"players-active", "/debug/nox/players", func() any { return s.Players.List() }},
		{"players-empty-all", "/debug/nox/players?all=", func() any { return s.Players.List() }},
		{"players-nonempty-all", "/debug/nox/players?all=0", func() any { return s.Players.ListSlots() }},
		{"walls-all", "/debug/nox/walls", func() any { return s.Walls.All() }},
		{"walls-grid", "/debug/nox/walls?x=12&y=14", func() any { return w1 }},
		{"walls-missing", "/debug/nox/walls?x=254&y=254", func() any { return (*server.Wall)(nil) }},
		{"walls-invalid-fallback", "/debug/nox/walls?x=bad&y=14", func() any { return s.Walls.All() }},
		{"walls-partial-fallback", "/debug/nox/walls?x=12", func() any { return s.Walls.All() }},
		{"waypoints", "/debug/nox/waypoints", func() any { return s.WPs.All() }},
		{"drawables", "/debug/nox/drawable", func() any { return cl.Objs.AllList1() }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			reset()
			want := httptest.NewRecorder()
			writeJSONResp(want, tt.value())
			retired := false
			got := &retirementHeaderWriterR316{ResponseRecorder: httptest.NewRecorder(), retire: func() {
				retired = true
				retire()
			}}
			serveAPIWorldRouteR316(t, s, got, tt.path)
			if !retired || got.Code != want.Code || got.Header().Get("Content-Type") != want.Header().Get("Content-Type") || !bytes.Equal(got.Body.Bytes(), want.Body.Bytes()) {
				t.Fatalf("native lifetime or original JSON contract changed: retired=%t status=%d/%d headers=%v/%v\ngot=%s\nwant=%s", retired, got.Code, want.Code, got.Header(), want.Header(), got.Body, want.Body)
			}
		})
	}
}
