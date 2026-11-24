package legacy

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestJoinMinimapRevealNativeCEntry4DE380(t *testing.T) {
	s := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer := GetServer
	GetServer = func() Server { return &chatLegacyServer528AC0{srv: s} }
	ballID, crownID := memmap.PtrUint32(0x5D4594, 1563284), memmap.PtrUint32(0x5D4594, 1563272)
	oldBall, oldCrown := *ballID, *crownID
	*ballID, *crownID = 221, 223
	t.Cleanup(func() {
		GetServer = oldServer
		*ballID, *crownID = oldBall, oldCrown
	})

	objs, free := alloc.Make([]server.Object{}, 5)
	t.Cleanup(free)
	ball, crown := &objs[0], &objs[1]
	ball.TypeInd, ball.ObjFlags = 221, object.FlagActive
	crown.TypeInd, crown.ObjFlags = 223, object.FlagActive
	objs[2].TypeInd = 221
	objs[3].TypeInd, objs[3].ObjFlags = 224, object.FlagActive
	objs[4].TypeInd, objs[4].ObjFlags = 223, object.FlagActive
	for i := 0; i < 3; i++ {
		objs[i].ObjNext = &objs[i+1]
	}
	s.Objs.SetObjects(&objs[0])
	s.Objs.Pending = &objs[4]
	const index = 7
	t.Cleanup(func() {
		for i := range objs {
			s.Players.Nox_xxx_netUnmarkMinimapObj_417300(index, &objs[i], ^uint32(0))
		}
	})
	s.Players.Nox_xxx_netMarkMinimapObject_417190(index, ball, 2)
	if got := callJoinMinimapRevealC(index); got != 0 {
		t.Fatalf("minimap reveal returned %d, want 0", got)
	}
	player := s.Players.ByIndRaw(index)
	head := player.Field4580
	if player.MinimapTrackCount() != 2 || head == nil || head.Field4 != crown || head.Field0 != 1 ||
		head.Field8.Field4 != ball || head.Field8.Field0 != 3 {
		t.Fatal("minimap did not reveal the two active objectives in order")
	}
}

func TestJoinObjectSyncNativeCEntry4DE410(t *testing.T) {
	s := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer, oldSync := GetServer, objectHasSyncData
	t.Cleanup(func() { GetServer, objectHasSyncData = oldServer, oldSync })
	GetServer = func() Server { return &chatLegacyServer528AC0{srv: s} }

	objs, free := alloc.Make([]server.Object{}, 3)
	t.Cleanup(free)
	for i, class := range []object.Class{object.ClassImmobile, object.ClassClientPersist, object.ClassMonster} {
		objs[i].ObjClass = class
		if i+1 < len(objs) {
			objs[i].ObjNext = &objs[i+1]
		}
	}
	s.Objs.SetObjects(&objs[0])

	type call struct {
		obj *server.Object
		key uint32
	}
	for _, index := range []int32{0, 31} {
		mask := uint32(1) << index
		var calls []call
		objectHasSyncData = func(obj *server.Object, key uint32) bool {
			calls = append(calls, call{obj, key})
			return key == 1 || key == 0x8000
		}
		for i := range objs {
			objs[i].Field37 = 0xffffffff
			objs[i].Field38 = 0x01000100
			for p := range objs[i].Field140 {
				objs[i].Field140[p] = 0x76543abc
			}
		}
		callJoinObjectSyncC(index)
		if len(calls) != 32 {
			t.Fatalf("client %d received %d sync checks, want 32", index, len(calls))
		}
		for i, got := range calls {
			if want := (call{&objs[i/16], uint32(1) << (i % 16)}); got != want {
				t.Fatalf("sync check %d = %+v, want %+v", i, got, want)
			}
		}
		for i := range objs {
			want37, wantSync := uint32(0xffffffff), uint32(0x80010abc)
			if i == 2 {
				want37 &^= mask
				wantSync = 0xabc
			}
			obj := &objs[i]
			if obj.Field37 != want37 || obj.Field38 != 0x01000100|mask || obj.Field140[index] != wantSync {
				t.Fatalf("client %d, object %d masks = %x %x %x", index, i, obj.Field37, obj.Field38, obj.Field140[index])
			}
			for other, value := range obj.Field140 {
				if int32(other) != index && value != 0x76543abc {
					t.Fatalf("sync for client %d modified client %d", index, other)
				}
			}
		}
	}
}
