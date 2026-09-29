package opennox

import (
	"encoding/binary"
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/server"
)

func highAddressWall48EA70(t *testing.T) *server.Wall {
	t.Helper()
	wl := new(server.Wall)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(wl)) <= uintptr(math.MaxUint32) {
		t.Skipf("allocator returned a low address: %p", wl)
	}
	return wl
}

func TestFindSecretWallByIDNative48EA70PreservesHighAddressPointers(t *testing.T) {
	firstWall := highAddressWall48EA70(t)
	firstWall.Field10 = 0x1111
	want := highAddressWall48EA70(t)
	want.Field10 = 0x9234
	second := &server.SecretWall{Wall: want}
	first := &server.SecretWall{Next: second, Wall: firstWall}

	if got := findSecretWallByIDNative48EA70(first, 0x9234); got != want {
		t.Fatalf("secret-wall lookup = %p, want %p", got, want)
	}
	if got := findSecretWallByIDNative48EA70(first, 0xffff); got != nil {
		t.Fatalf("missing secret-wall lookup = %p, want nil", got)
	}
}

func TestHandleSecretWallStateNative48EA70PreservesHighAddressPointers(t *testing.T) {
	wl := highAddressWall48EA70(t)
	secret := new(server.SecretWall)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(secret)) <= uintptr(math.MaxUint32) {
		t.Skipf("allocator returned a low secret-wall address: %p", secret)
	}
	wl.AttachSecret(secret, 0x9234)
	lookups := 0
	hooks := wallStateHooks48EA70{
		connected: func() bool { return true },
		host:      func() bool { return false },
		secretByID: func(id uint16) *server.Wall {
			lookups++
			if id != 0x9234 {
				t.Fatalf("secret wall ID = %#x, want 0x9234", id)
			}
			return wl
		},
	}

	packet := []byte{byte(netmsg.MSG_OPEN_WALL), 0x34, 0x92, 0xaa}
	if got := handleWallStateNative48EA70(netmsg.MSG_OPEN_WALL, packet, hooks); got != 3 {
		t.Fatalf("open consumed bytes = %d, want 3", got)
	}
	if secret.OpenDelay != 23 || secret.State != 3 {
		t.Fatalf("open state = delay %d, state %d; want 23, 3", secret.OpenDelay, secret.State)
	}

	packet[0] = byte(netmsg.MSG_CLOSE_WALL)
	if got := handleWallStateNative48EA70(netmsg.MSG_CLOSE_WALL, packet, hooks); got != 3 {
		t.Fatalf("close consumed bytes = %d, want 3", got)
	}
	if secret.OpenDelay != 0 || secret.State != 1 {
		t.Fatalf("close state = delay %d, state %d; want 0, 1", secret.OpenDelay, secret.State)
	}
	if lookups != 2 {
		t.Fatalf("secret-wall lookups = %d, want 2", lookups)
	}
}

func TestHandleMagicWallChangeOrAddNative48EA70(t *testing.T) {
	packet := []byte{byte(netmsg.MSG_CHANGE_OR_ADD_WALL_MAGIC), 0xa1, 0xb2, 0xc3, 0x24, 0x36, 0xee}
	wantPacket := append([]byte(nil), packet...)
	pos := image.Pt(0x24, 0x36)
	wl := highAddressWall48EA70(t)
	wl.X5, wl.Y6, wl.Health7 = byte(pos.X), byte(pos.Y), 77
	var calls []string
	hooks := wallStateHooks48EA70{
		connected: func() bool { calls = append(calls, "connected"); return true },
		wallAtGrid: func(got image.Point) *server.Wall {
			calls = append(calls, "lookup")
			if got != pos {
				t.Fatalf("lookup position = %v, want %v", got, pos)
			}
			return wl
		},
		createAtGrid: func(image.Point) *server.Wall {
			calls = append(calls, "create")
			return nil
		},
	}
	if got := handleWallStateNative48EA70(netmsg.MSG_CHANGE_OR_ADD_WALL_MAGIC, packet, hooks); got != 6 {
		t.Fatalf("consumed bytes = %d, want 6", got)
	}
	if wl.Tile1 != 0xa1 || wl.Dir0 != 0xb2 || wl.Field2 != 0xc3 || wl.Health7 != 77 {
		t.Fatalf("wall fields = tile %#x dir %#x field %#x health %d", wl.Tile1, wl.Dir0, wl.Field2, wl.Health7)
	}
	if want := []string{"connected", "lookup"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("existing-wall calls = %v, want %v", calls, want)
	}
	if !reflect.DeepEqual(packet, wantPacket) {
		t.Fatalf("packet mutated: got %v, want %v", packet, wantPacket)
	}

	created := highAddressWall48EA70(t)
	created.X5, created.Y6 = byte(pos.X), byte(pos.Y)
	calls = nil
	hooks.wallAtGrid = func(got image.Point) *server.Wall {
		calls = append(calls, "lookup")
		return nil
	}
	hooks.createAtGrid = func(got image.Point) *server.Wall {
		calls = append(calls, "create")
		if got != pos {
			t.Fatalf("create position = %v, want %v", got, pos)
		}
		return created
	}
	if got := handleWallStateNative48EA70(netmsg.MSG_CHANGE_OR_ADD_WALL_MAGIC, packet, hooks); got != 6 {
		t.Fatalf("created consumed bytes = %d, want 6", got)
	}
	if created.Tile1 != 0xa1 || created.Dir0 != 0xb2 || created.Field2 != 0xc3 {
		t.Fatalf("created wall fields = tile %#x dir %#x field %#x", created.Tile1, created.Dir0, created.Field2)
	}
	if want := []string{"connected", "lookup", "create"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("created-wall calls = %v, want %v", calls, want)
	}
}

func TestHandleMagicWallRemoveNative48EA70(t *testing.T) {
	packet := []byte{byte(netmsg.MSG_REMOVE_WALL_MAGIC), 0x24, 0x36, 0xee}
	wl := highAddressWall48EA70(t)
	wl.X5, wl.Y6 = 0x24, 0x36
	var deleted image.Point
	hooks := wallStateHooks48EA70{
		connected: func() bool { return true },
		wallAtGrid: func(pos image.Point) *server.Wall {
			if pos != image.Pt(0x24, 0x36) {
				t.Fatalf("lookup position = %v", pos)
			}
			return wl
		},
		deleteAtGrid: func(pos image.Point) { deleted = pos },
	}
	if got := handleWallStateNative48EA70(netmsg.MSG_REMOVE_WALL_MAGIC, packet, hooks); got != 3 {
		t.Fatalf("consumed bytes = %d, want 3", got)
	}
	if deleted != wl.GridPos() {
		t.Fatalf("deleted position = %v, want %v", deleted, wl.GridPos())
	}
}

func TestHandleWallStateNative48EA70Guards(t *testing.T) {
	tests := []struct {
		op   netmsg.Op
		size int
	}{
		{netmsg.MSG_OPEN_WALL, 3},
		{netmsg.MSG_CLOSE_WALL, 3},
		{netmsg.MSG_CHANGE_OR_ADD_WALL_MAGIC, 6},
		{netmsg.MSG_REMOVE_WALL_MAGIC, 3},
	}
	for _, tc := range tests {
		for n := 0; n < tc.size; n++ {
			if got := handleWallStateNative48EA70(tc.op, make([]byte, n), wallStateHooks48EA70{}); got != -1 {
				t.Fatalf("op %v short packet of %d bytes consumed %d, want -1", tc.op, n, got)
			}
		}
	}

	lookups := 0
	disconnected := wallStateHooks48EA70{
		connected:  func() bool { return false },
		host:       func() bool { t.Fatal("host queried while disconnected"); return false },
		secretByID: func(uint16) *server.Wall { lookups++; return nil },
		wallAtGrid: func(image.Point) *server.Wall { lookups++; return nil },
	}
	if got := handleWallStateNative48EA70(netmsg.MSG_OPEN_WALL, []byte{59, 1, 0}, disconnected); got != 3 {
		t.Fatalf("disconnected open consumed %d, want 3", got)
	}
	if got := handleWallStateNative48EA70(netmsg.MSG_CHANGE_OR_ADD_WALL_MAGIC, []byte{61, 1, 2, 3, 4, 6}, disconnected); got != 6 {
		t.Fatalf("disconnected magic wall consumed %d, want 6", got)
	}
	if lookups != 0 {
		t.Fatalf("disconnected lookups = %d, want 0", lookups)
	}

	host := wallStateHooks48EA70{
		connected:  func() bool { return true },
		host:       func() bool { return true },
		secretByID: func(uint16) *server.Wall { lookups++; return nil },
	}
	if got := handleWallStateNative48EA70(netmsg.MSG_CLOSE_WALL, []byte{60, 1, 0}, host); got != 3 {
		t.Fatalf("host close consumed %d, want 3", got)
	}
	if lookups != 0 {
		t.Fatalf("host secret lookups = %d, want 0", lookups)
	}
	if got := handleWallStateNative48EA70(netmsg.MSG_CODE2, []byte{2}, wallStateHooks48EA70{}); got != -1 {
		t.Fatalf("unsupported opcode consumed %d, want -1", got)
	}
}

func TestHandleSecretWallStateNative48EA70MissingOrNonSecret(t *testing.T) {
	packet := make([]byte, 3)
	packet[0] = byte(netmsg.MSG_OPEN_WALL)
	binary.LittleEndian.PutUint16(packet[1:], 7)
	wl := highAddressWall48EA70(t)
	hooks := wallStateHooks48EA70{
		connected:  func() bool { return true },
		host:       func() bool { return false },
		secretByID: func(uint16) *server.Wall { return wl },
	}
	if got := handleWallStateNative48EA70(netmsg.MSG_OPEN_WALL, packet, hooks); got != 3 {
		t.Fatalf("non-secret consumed bytes = %d, want 3", got)
	}
	hooks.secretByID = func(uint16) *server.Wall { return nil }
	if got := handleWallStateNative48EA70(netmsg.MSG_OPEN_WALL, packet, hooks); got != 3 {
		t.Fatalf("missing wall consumed bytes = %d, want 3", got)
	}
}
