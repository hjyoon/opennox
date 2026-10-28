package opennox

import (
	"encoding/json"
	"net/http"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestAPIObjectsNativeHealthSnapshot(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(1) {
		t.Fatal("native allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	obj := s.Objs.NewObject(&server.ObjectType{})
	obj.Extent = 17
	obj.ObjClass = object.ClassMonster
	s.Objs.SetObjects(obj)
	health, free := alloc.Make([]server.HealthData(nil), 1)
	obj.HealthData = &health[0]
	t.Cleanup(func() { obj.HealthData = nil; free() })
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(obj)) <= 0xffffffff || uintptr(unsafe.Pointer(obj.HealthData)) <= 0xffffffff) {
		t.Fatal("expected native object and health above4 GiB")
	}
	old := noxServer
	noxServer = s
	t.Cleanup(func() { noxServer = old })
	read := func() map[string]any {
		w := runAPIObjectsRequestR316(t, s, "/debug/nox/objects")
		if w.Code != http.StatusOK {
			t.Fatalf("HTTP status%d", w.Code)
		}
		var rows []map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &rows); err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 {
			t.Fatal("native object selection", len(rows))
		}
		return rows[0]
	}
	health[0] = server.HealthData{Cur: 17, Max: 23}
	first := read()
	health[0].Cur = 0
	obj.ObjFlags |= object.FlagDead
	second := read()
	if first["health"].(map[string]any)["current"] != float64(17) || first["health"].(map[string]any)["maximum"] != float64(23) {
		t.Fatal("published native HP changed with later world state")
	}
	if second["health"].(map[string]any)["current"] != float64(0) || second["health"].(map[string]any)["maximum"] != float64(23) {
		t.Fatal("native zero HP must remain explicit")
	}
	if health[0].Cur != 0 || health[0].Max != 23 {
		t.Fatal("read-only route modified native health")
	}
	obj.HealthData = nil
	if _, ok := read()["health"]; ok {
		t.Fatal("absent health must not become a fabricated zero-HP record")
	}
}
