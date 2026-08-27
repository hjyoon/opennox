package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

type rewardAnkhLegacyServer4F2110 struct {
	Server
	native  *server.Server
	created *server.Object
	owner   server.Obj
	point   types.Pointf
	deleted *server.Object
}

func (s *rewardAnkhLegacyServer4F2110) S() *server.Server {
	return s.native
}

func (s *rewardAnkhLegacyServer4F2110) CreateObjectAt(object, owner server.Obj, point types.Pointf) {
	s.created = server.ToObject(object)
	s.owner = owner
	s.point = point
}

func (s *rewardAnkhLegacyServer4F2110) DelayedDelete(object *server.Object) {
	s.deleted = object
}

func TestRewardAnkhLegacyRuntimePreservesNativePointersAndNullOwner4F2110(t *testing.T) {
	outer := &rewardAnkhLegacyServer4F2110{native: new(server.Server)}
	object := new(server.Object)
	owner := new(server.Object)
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(object)) <= math.MaxUint32 ||
		uintptr(unsafe.Pointer(owner)) <= math.MaxUint32) {
		t.Fatalf("object pointers do not exercise native width: %#x/%#x",
			uintptr(unsafe.Pointer(object)), uintptr(unsafe.Pointer(owner)))
	}
	point := types.Pointf{X: 61, Y: 62}

	ankhRuntime := rewardAnkhReplaceRuntime4F2110(outer)
	ankhRuntime.CreateAt(object, nil, point)
	if outer.created != object || outer.owner != nil || outer.point != point {
		t.Fatalf("Ankh create object/owner/point = %p/%v/%+v", outer.created, outer.owner, outer.point)
	}
	ankhRuntime.CreateAt(object, owner, point)
	if outer.owner == nil || server.ToObject(outer.owner) != owner {
		t.Fatalf("Ankh non-null owner = %v, want %p", outer.owner, owner)
	}
	ankhRuntime.DelayedDelete(object)
	if outer.deleted != object {
		t.Fatalf("Ankh delete object = %p, want %p", outer.deleted, object)
	}

	outer.owner = owner
	containerRuntime := rewardContainerRuntime4F1F20(outer)
	containerRuntime.CreateAt(object, nil, point)
	if outer.owner != nil {
		t.Fatalf("container null owner became typed-nil interface: %v", outer.owner)
	}
}
