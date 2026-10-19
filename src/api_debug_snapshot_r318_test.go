package opennox

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func debugAPIServerR318(t *testing.T) *Server {
	t.Helper()
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	old := noxServer
	noxServer = s
	t.Cleanup(func() { noxServer = old })
	return s
}

func preserveAPIGameFlagsR318(t *testing.T) {
	t.Helper()
	old := noxflags.GetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(old)
	})
}

func TestAPIDebugSnapshotR318CanceledRegisteredRoutes(t *testing.T) {
	debugAPIServerR318(t)
	preserveAPIGameFlagsR318(t)
	for _, path := range []string{
		"/debug/nox/types",
		"/debug/nox/memmap/snapshot",
		"/debug/nox/flags/game?set=0x40000000&unset=0x80000000",
	} {
		t.Run(path, func(t *testing.T) {
			before := noxflags.GetGame()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			w := httptest.NewRecorder()
			http.DefaultServeMux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil).WithContext(ctx))
			if w.Body.Len() != 0 || len(w.Header()) != 0 || noxflags.GetGame() != before {
				t.Fatalf("canceled route responded or changed the world: headers=%v bytes=%d flags=%v/%v", w.Header(), w.Body.Len(), noxflags.GetGame(), before)
			}
		})
	}
}

func TestAPIDebugSnapshotR318TypesOwnBytesBeforeRetirement(t *testing.T) {
	s := debugAPIServerR318(t)
	s.Types.Free()
	t.Cleanup(s.Types.Free)
	for _, name := range []string{"First Borrowed Type", "Second Borrowed Type"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	want := httptest.NewRecorder()
	writeJSONResp(want, s.Types.List())
	retired := false
	w := &retirementHeaderWriterR316{ResponseRecorder: httptest.NewRecorder(), retire: func() {
		retired = true
		// Free clears the same type records returned by List, including their
		// IDs. A shallow slice snapshot still fails after this retirement.
		s.Types.Free()
	}}
	serveAPIWorldRouteR316(t, s, w, "/debug/nox/types")
	if !retired || len(s.Types.List()) != 0 || w.Code != want.Code || w.Header().Get("Content-Type") != want.Header().Get("Content-Type") || !bytes.Equal(w.Body.Bytes(), want.Body.Bytes()) {
		t.Fatalf("type lifetime or JSON format changed: retired=%t count=%d status=%d headers=%v\ngot=%s\nwant=%s", retired, len(s.Types.List()), w.Code, w.Header(), w.Body, want.Body)
	}
}

func TestAPIDebugSnapshotR318EmptyTypesKeepNullFormat(t *testing.T) {
	s := debugAPIServerR318(t)
	s.Types.Free()
	w := runAPIObjectsRequestR316(t, s, "/debug/nox/types")
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/json" || w.Body.String() != "null\n" {
		t.Fatalf("empty type response changed: status=%d headers=%v body=%q", w.Code, w.Header(), w.Body.String())
	}
}

type firstWriteRetirementR318 struct {
	*httptest.ResponseRecorder
	retire func()
}

func (w *firstWriteRetirementR318) Write(p []byte) (int, error) {
	if w.retire != nil {
		retire := w.retire
		w.retire = nil
		retire()
	}
	return w.ResponseRecorder.Write(p)
}

// Use the original route's ZIP serialization as the byte-format oracle. The
// two substituted Blob.Data slices borrow C storage just like the real blobs;
// their original registry entries and storage are restored before cleanup.
func TestAPIDebugSnapshotR318ArchiveOwnsAllNativeBytesBeforeHTTP(t *testing.T) {
	s := debugAPIServerR318(t)
	blobs := memmap.Blobs()
	if len(blobs) < 2 {
		t.Fatal("native blob registry is missing")
	}
	first := memmap.BlobByAddr(blobs[0].Addr)
	last := memmap.BlobByAddr(blobs[len(blobs)-1].Addr)
	oldFirst, oldLast := first.Data, last.Data
	firstData, freeFirst := alloc.Make([]byte(nil), len(oldFirst))
	lastData, freeLast := alloc.Make([]byte(nil), len(oldLast))
	t.Cleanup(func() {
		first.Data, last.Data = oldFirst, oldLast
		freeFirst()
		freeLast()
	})
	state := uint32(0x1729468b)
	for _, payload := range [][]byte{firstData, lastData} {
		for i := range payload {
			state ^= state << 13
			state ^= state >> 17
			state ^= state << 5
			payload[i] = byte(state)
		}
	}
	first.Data, last.Data = firstData, lastData
	want := httptest.NewRecorder()
	zw := zip.NewWriter(want)
	for _, blob := range memmap.Blobs() {
		entry, err := zw.Create(fmt.Sprintf("blob_%X.dat", blob.Addr))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(blob.Data); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	retired := false
	w := &firstWriteRetirementR318{ResponseRecorder: httptest.NewRecorder(), retire: func() {
		retired = true
		clear(firstData)
		clear(lastData)
		first.Data, last.Data = nil, nil
	}}
	serveAPIWorldRouteR316(t, s, w, "/debug/nox/memmap/snapshot")
	if !retired || w.Code != want.Code || w.Result().Header.Get("Content-Type") != want.Result().Header.Get("Content-Type") || !bytes.Equal(w.Body.Bytes(), want.Body.Bytes()) {
		t.Fatalf("archive used retired native storage or changed ZIP bytes: retired=%t status=%d/%d type=%q/%q bytes=%d/%d", retired, w.Code, want.Code, w.Result().Header.Get("Content-Type"), want.Result().Header.Get("Content-Type"), w.Body.Len(), want.Body.Len())
	}
	archive, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	if err != nil || len(archive.File) != len(blobs) {
		t.Fatalf("archive members changed: err=%v", err)
	}
	for i, entry := range archive.File {
		if entry.Name != fmt.Sprintf("blob_%X.dat", blobs[i].Addr) {
			t.Fatalf("archive order changed at %d: %s", i, entry.Name)
		}
	}
}

func flagJSONResponseR318(t *testing.T, v any) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	if err := enc.Encode(v); err != nil {
		t.Fatal(err)
	}
	return w
}

func TestAPIDebugSnapshotR318GameMutationOrderingAndOriginalFormat(t *testing.T) {
	s := debugAPIServerR318(t)
	preserveAPIGameFlagsR318(t)
	for _, tc := range []struct {
		name     string
		initial  noxflags.GameFlag
		query    string
		expected noxflags.GameFlag
	}{
		{"set-then-unset", 0, "set=0xc0000000&unset=0x80000000", 0x40000000},
		{"invalid-set-valid-unset", 0xc0000000, "set=not-a-flag&unset=0x80000000", 0x40000000},
		{"valid-set-invalid-unset", 0, "set=0x80000000&unset=not-a-flag", 0x80000000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			noxflags.ResetGame()
			noxflags.SetGame(tc.initial)
			want := flagJSONResponseR318(t, tc.expected)
			w := runAPIObjectsRequestR316(t, s, "/debug/nox/flags/game?"+tc.query)
			if noxflags.GetGame() != tc.expected || w.Code != want.Code || w.Result().Header.Get("Content-Type") != want.Result().Header.Get("Content-Type") || !bytes.Equal(w.Body.Bytes(), want.Body.Bytes()) {
				t.Fatalf("game mutation order or original HTTP format changed: flags=%v/%v status=%d/%d headers=%v/%v body=%q/%q", noxflags.GetGame(), tc.expected, w.Code, want.Code, w.Result().Header, want.Result().Header, w.Body.String(), want.Body.String())
			}
		})
	}
}

func TestAPIDebugSnapshotR318CopiedFlagValueRoutesKeepImmediateResponses(t *testing.T) {
	old := noxServer
	noxServer = nil
	t.Cleanup(func() { noxServer = old })
	for _, tc := range []struct {
		path  string
		value any
	}{
		{"/debug/nox/flags", struct {
			Game   noxflags.GameFlag   `json:"game"`
			Engine noxflags.EngineFlag `json:"engine"`
		}{noxflags.GetGame(), noxflags.GetEngine()}},
		{"/debug/nox/flags/game", noxflags.GetGame()},
		{"/debug/nox/flags/game?set=not-a-flag&unset=0", noxflags.GetGame()},
		{"/debug/nox/flags/engine", noxflags.GetEngine()},
		{"/debug/nox/flags/engine?set=0&unset=0", noxflags.GetEngine()},
	} {
		t.Run(tc.path, func(t *testing.T) {
			want := flagJSONResponseR318(t, tc.value)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			w := httptest.NewRecorder()
			http.DefaultServeMux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.path, nil).WithContext(ctx))
			if w.Code != want.Code || w.Result().Header.Get("Content-Type") != want.Result().Header.Get("Content-Type") || !bytes.Equal(w.Body.Bytes(), want.Body.Bytes()) {
				t.Fatalf("copied value route unnecessarily depends on a loop or changed format: status=%d/%d headers=%v/%v body=%q/%q", w.Code, want.Code, w.Result().Header, want.Result().Header, w.Body.String(), want.Body.String())
			}
		})
	}
}
