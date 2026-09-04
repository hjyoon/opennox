package legacy

import (
	"bytes"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

type netSendFxAllLegacyServer523030 struct {
	Server
	srv *server.Server
}

func (s *netSendFxAllLegacyServer523030) S() *server.Server {
	return s.srv
}

func TestNetSendFxAllCliEntry523030KeepsNativePlayerPointers(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.NetList.Init()
	t.Cleanup(srv.NetList.Free)

	oldGetServer := GetServer
	GetServer = func() Server { return &netSendFxAllLegacyServer523030{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	type playerFixture struct {
		player *server.Player
		unit   *server.Object
		update *server.PlayerUpdateData
		camera *server.Object
	}
	newPlayer := func(id int, unitPos types.Pointf) playerFixture {
		pl := srv.Players.NewRaw(id)
		if pl == nil {
			t.Fatal("cannot allocate player fixture")
		}
		update := &server.PlayerUpdateData{Player: pl}
		unit := &server.Object{
			ObjClass:   object.ClassPlayer,
			PosVec:     unitPos,
			UpdateData: unsafe.Pointer(update),
		}
		pl.PlayerUnit = unit
		pl.Field10 = 25
		pl.Field12 = 20
		return playerFixture{player: pl, unit: unit, update: update}
	}

	direct := newPlayer(1001, types.Ptf(100, 100))
	follow := newPlayer(1002, types.Ptf(1000, 1000))
	follow.camera = &server.Object{PosVec: types.Ptf(100, 100)}
	follow.player.Field3680 = 1
	follow.player.CameraFollowObj = follow.camera
	distant := newPlayer(1003, types.Ptf(1000, 1000))

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]unsafe.Pointer{
			"direct player":  unsafe.Pointer(direct.player),
			"direct unit":    unsafe.Pointer(direct.unit),
			"direct update":  unsafe.Pointer(direct.update),
			"follow player":  unsafe.Pointer(follow.player),
			"follow unit":    unsafe.Pointer(follow.unit),
			"follow update":  unsafe.Pointer(follow.update),
			"follow camera":  unsafe.Pointer(follow.camera),
			"distant player": unsafe.Pointer(distant.player),
			"distant unit":   unsafe.Pointer(distant.unit),
			"distant update": unsafe.Pointer(distant.update),
		} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want address above the ABI32 range", name, ptr)
			}
		}
	}

	packet := []byte{0x99, 0x30, 0x23, 0x50}
	if got := netSendFxAllCliEntry523030(types.Ptf(120, 115), packet); got != 0 {
		t.Fatalf("entry return = %d, want 0", got)
	}
	for _, tc := range []struct {
		name string
		pl   *server.Player
		want []byte
	}{
		{name: "unit position", pl: direct.player, want: packet},
		{name: "camera position", pl: follow.player, want: packet},
		{name: "outside viewport", pl: distant.player},
	} {
		got := srv.NetList.CopyPacketsA(ntype.PlayerInd(tc.pl.PlayerInd), netlist.Kind1)
		if !bytes.Equal(got, tc.want) {
			t.Errorf("%s packet = % x, want % x", tc.name, got, tc.want)
		}
	}

	if got := netSendFxAllCliEntry523030(types.Ptf(25, 100), packet); got != 0 {
		t.Fatalf("boundary entry return = %d, want 0", got)
	}
	for _, pl := range []*server.Player{direct.player, follow.player, distant.player} {
		if got := srv.NetList.CopyPacketsA(ntype.PlayerInd(pl.PlayerInd), netlist.Kind1); len(got) != 0 {
			t.Errorf("strict viewport boundary packet = % x, want none", got)
		}
	}

	justInside := math.Nextafter32(25, float32(math.Inf(1)))
	if got := netSendFxAllCliEntry523030(types.Ptf(justInside, 100), packet); got != 0 {
		t.Fatalf("inside entry return = %d, want 0", got)
	}
	for _, tc := range []struct {
		name string
		pl   *server.Player
		want []byte
	}{
		{name: "unit just inside", pl: direct.player, want: packet},
		{name: "camera just inside", pl: follow.player, want: packet},
		{name: "distant just inside", pl: distant.player},
	} {
		got := srv.NetList.CopyPacketsA(ntype.PlayerInd(tc.pl.PlayerInd), netlist.Kind1)
		if !bytes.Equal(got, tc.want) {
			t.Errorf("%s packet = % x, want % x", tc.name, got, tc.want)
		}
	}

	runtime.KeepAlive(direct)
	runtime.KeepAlive(follow)
	runtime.KeepAlive(distant)
	runtime.KeepAlive(packet)
}
