package legacy

import (
	"fmt"
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type playerAttackTraceWallServer538330 struct {
	*playerReportPacketServer
	bounds image.Rectangle
	source *server.Object
}

func (s *playerAttackTraceWallServer538330) Nox_xxx_mapDamageToWalls_534FC0(
	bounds image.Rectangle, _ types.Pointf, _ float32, _ int, _ object.DamageType, source *server.Object,
) bool {
	s.bounds, s.source = bounds, source
	return false
}

func TestPlayerAttackTraceNativeCABI538330(t *testing.T) {
	var data server.PlayerAttackData538330
	goLayout := [5]uintptr{unsafe.Sizeof(data), unsafe.Offsetof(data.Owner), unsafe.Offsetof(data.Position),
		unsafe.Offsetof(data.Weapon), unsafe.Offsetof(data.DirectionMask)}
	if cLayout := attackTraceCLayout538330(); cLayout != goLayout {
		t.Fatalf("attack layout C=%v Go=%v", cLayout, goLayout)
	}

	s := playerVitalReportWorld(t)
	host := &playerAttackTraceWallServer538330{playerReportPacketServer: &playerReportPacketServer{world: s}}
	GetServer = func() Server { return host }
	noxflags.ResetGame()
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.InitHigh() == 0 {
		t.Fatal("allocate native walls")
	}
	t.Cleanup(s.Walls.Free)
	previous := publicAttackTrace538330
	t.Cleanup(func() { publicAttackTrace538330 = previous })

	owner := game4ConsumerValue(t, server.Object{PosVec: types.Ptf(1000, 1000)})
	candidate := game4ConsumerValue(t, server.Object{PosVec: types.Ptf(1030, 1000)})
	attack := game4ConsumerValue(t, server.PlayerAttackData538330{Owner: owner, Damage: 12.5, DirectionMask: 1})
	if unsafe.Sizeof(uintptr(0)) > 4 {
		for _, p := range []unsafe.Pointer{owner.CObj(), candidate.CObj(), unsafe.Pointer(attack)} {
			if uint64(uintptr(p)) <= 0xffffffff {
				t.Fatalf("expected native high pointer, got %p", p)
			}
		}
	}

	want := server.PlayerAttackTraceState538330{Distance: 120, Closest: candidate, Damaged: 29}
	publicAttackTrace538330 = want
	callAttackSelectC5386A0(owner, owner)
	if callAttackTraceC538330(nil, attack, false) != 0 || callAttackTraceC538330(owner, nil, true) != 0 ||
		publicAttackTrace538330 != want {
		t.Fatal("empty inputs changed shared attack scratch")
	}
	for _, indirect := range []bool{false, true} {
		publicAttackTrace538330.Damaged = 0
		callAttackHitC538510(candidate, attack, indirect)
		if publicAttackTrace538330.Damaged != 1 || attack.Owner != owner || attack.Damage != 12.5 {
			t.Fatalf("candidate dispatch indirect=%t changed attack state or pointers", indirect)
		}
	}

	// The selection callback uses the native direction table and the actual
	// spatial index. Check that both C roots hit the same object and only the
	// original public root publishes its selection scratch.
	x, y := memmap.PtrFloat32(0x587000, 194136), memmap.PtrFloat32(0x587000, 194140)
	oldX, oldY := *x, *y
	*x, *y = 1, 0
	t.Cleanup(func() { *x, *y = oldX, oldY })
	hits := 0
	damageFn := objectDamageNativeProbePtr()
	server.RegisterObjectDamageGo(fmt.Sprintf("PlayerAttackTrace%d", objectDamageNativeTestSequence.Add(1)), damageFn,
		func(target, source, weapon *server.Object, damage int32, typ object.DamageType) bool {
			if target != candidate || source != owner || weapon != nil || damage != 13 || typ != object.DamageType(0) {
				t.Fatalf("attack callback target=%p owner=%p weapon=%p damage=%d type=%v", target, source, weapon, damage, typ)
			}
			hits++
			return true
		})
	candidate.ObjFlags = object.FlagActive
	candidate.NewPos = candidate.PosVec
	candidate.Shape.Kind, candidate.Shape.Circle.R, candidate.Shape.Circle.R2 = server.ShapeKindCircle, 5, 25
	candidate.Damage = damageFn
	s.Map.AddObjectToIndex(candidate)
	t.Cleanup(func() { s.Map.RemoveObjectFromIndex(candidate) })
	attack.Position, attack.Radius = owner.PosVec, 35
	publicAttackTrace538330 = want
	if result := callAttackTraceC538330(owner, attack, true); result != 1 || hits != 1 || publicAttackTrace538330 != want {
		t.Fatalf("live C root returned %d, damage calls=%d, scratch=%+v", result, hits, publicAttackTrace538330)
	}
	if result := callAttackTraceC538330(owner, attack, false); result != 1 || hits != 2 ||
		publicAttackTrace538330.Closest != candidate || publicAttackTrace538330.Distance != 25 {
		t.Fatalf("public C root returned %d, damage calls=%d, scratch=%+v", result, hits, publicAttackTrace538330)
	}
	if host.source != owner || host.bounds != image.Rect(41, 41, 45, 45) {
		t.Fatalf("positive-radius wall dispatch source=%p bounds=%v", host.source, host.bounds)
	}
	attack.Radius = -35
	if result := callAttackTraceC538330(owner, attack, false); result != 0 || hits != 2 ||
		host.bounds != (image.Rectangle{Min: image.Pt(45, 45), Max: image.Pt(41, 41)}) {
		t.Fatalf("negative-radius C root returned %d, damage calls=%d, wall bounds=%v", result, hits, host.bounds)
	}
}
