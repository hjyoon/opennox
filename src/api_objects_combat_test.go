package opennox

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestAPIObjectCombatNativeSnapshot(t *testing.T) {
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	if !s.Objs.Init(1) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	o := s.Objs.NewObject(&server.ObjectType{})
	o.Extent, o.ScriptIDVal, o.ObjClass = 17, 1000347545, object.ClassMonster
	data, free := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(func() { o.UpdateData = nil; free() })
	o.UpdateData = unsafe.Pointer(data)
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(o)) <= 0xffffffff || uintptr(o.UpdateData) <= 0xffffffff) {
		t.Fatal("native object/update must be above4GiB")
	}
	data.StatusFlags = object.MonsterStatus(0x4020)
	o.SetAggression(0.625)
	o.SetRetreatLevel(0.25)
	o.SetRegroupLevel(0.875)
	o.Buffs = 1 << server.ENCHANT_ANTI_MAGIC
	o.BuffsDur[server.ENCHANT_ANTI_MAGIC], o.BuffsPower[server.ENCHANT_ANTI_MAGIC] = 120, 5
	s.Objs.SetObjects(o)
	old := noxServer
	noxServer = s
	t.Cleanup(func() { noxServer = old })
	read := func() debugObjectCombat {
		objectBytes := unsafe.Slice((*byte)(unsafe.Pointer(o)), int(unsafe.Sizeof(*o)))
		updateBytes := unsafe.Slice((*byte)(unsafe.Pointer(data)), int(unsafe.Sizeof(*data)))
		beforeObject, beforeUpdate := bytes.Clone(objectBytes), bytes.Clone(updateBytes)
		w := runAPIObjectsRequestR316(t, s, "/debug/nox/objects/combat?ind=17&flags=DEAD")
		if w.Code != http.StatusOK {
			t.Fatal("HTTP status", w.Code)
		}
		var result struct{ Objects []debugObjectCombat }
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || len(result.Objects) != 1 {
			t.Fatal("native combat response", err, len(result.Objects))
		}
		if !bytes.Equal(objectBytes, beforeObject) || !bytes.Equal(updateBytes, beforeUpdate) {
			t.Fatal("combat observer or query changed native object/update bytes")
		}
		return result.Objects[0]
	}
	first := read()
	if first.Ind != 17 || first.ScriptID != 1000347545 || first.Monster == nil || first.Monster.StatusFlags != 0x4020 || !first.Monster.CanCast || first.Monster.Aggression != 0.625 || first.Monster.Retreat != 0.25 || first.Monster.Regroup != 0.875 || !first.AntiMagic.Active || first.AntiMagic.Frames != 120 || first.AntiMagic.Power != 5 {
		t.Fatal("native combat fields", first)
	}
	data.StatusFlags, o.Buffs = 0, 0
	o.BuffsDur[server.ENCHANT_ANTI_MAGIC] = 0
	second := read()
	if second.Monster == nil || second.Monster.CanCast || second.AntiMagic.Active || second.AntiMagic.Frames != 0 || !first.Monster.CanCast || !first.AntiMagic.Active || first.AntiMagic.Frames != 120 {
		t.Fatal("snapshots must copy live native state", first, second)
	}
	o.UpdateData = nil
	if got := read(); got.Monster != nil {
		t.Fatal("missing monster record must not become zero status", got)
	}
	// A non-monster update record is a different union member.
	o.UpdateData, o.ObjClass = unsafe.Pointer(data), object.ClassPlayer
	if got := read(); got.Monster != nil {
		t.Fatal("non-monster record interpreted as monster data", got)
	}
}

func TestAPIObjectCombatCanceledRequestDoesNotAccessWorld(t *testing.T) {
	old := noxServer
	noxServer = &Server{Server: &server.Server{}}
	t.Cleanup(func() { noxServer = old })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/debug/nox/objects/combat", nil).WithContext(ctx))
	if w.Body.Len() != 0 || len(w.Header()) != 0 {
		t.Fatal("canceled combat request accessed/responded from world")
	}
}
