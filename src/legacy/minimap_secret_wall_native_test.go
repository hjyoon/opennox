package legacy

import (
	"bytes"
	"math"
	"os"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestSecretWallMinimap472600NativeStateAndLegacyOffset(t *testing.T) {
	const guardSize = 16
	mem, free := alloc.Malloc(unsafe.Sizeof(server.SecretWall{}) + 2*guardSize)
	next, freeNext := alloc.New(server.SecretWall{})
	wall, freeWall := alloc.New(server.Wall{})
	t.Cleanup(free)
	t.Cleanup(freeNext)
	t.Cleanup(freeWall)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]unsafe.Pointer{
			"record": mem, "next": unsafe.Pointer(next), "wall": unsafe.Pointer(wall),
		} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("%s = %p, want native pointer above 4 GiB", name, ptr)
			}
		}
	}
	raw := unsafe.Slice((*byte)(mem), int(unsafe.Sizeof(server.SecretWall{}))+2*guardSize)
	for i := range raw {
		raw[i] = 0xa5
	}
	ptr := unsafe.Add(mem, guardSize)
	secret := (*server.SecretWall)(ptr)
	*secret = server.SecretWall{
		Next: next, X: 111, Y: 57, Wall: wall, OpenWait: 0x12345678,
		Flags: 0x9a, OpenDelay: 23, LastOpen: 0xabcdef01, PlayerBits: 0x87654321,
	}
	const legacyStateOffset = 21 // GAME.EXE 00472A40: mov dl,[esi+15h].
	poisonCount := 1
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if got := unsafe.Offsetof(secret.State); got != 29 {
			t.Fatalf("native state offset = %d, want 29", got)
		}
		poisonCount = 256
	} else if got := unsafe.Offsetof(secret.State); got != legacyStateOffset {
		t.Fatalf("PE32 state offset = %d, want 21", got)
	}
	legacyByte := raw[guardSize+legacyStateOffset]
	defer func() { raw[guardSize+legacyStateOffset] = legacyByte }()
	checks := 0
	for state := 0; state < 256; state++ {
		secret.State = byte(state)
		// This is the independent original state table, not the implementation.
		want := state == 2 || state == 3
		for poison := 0; poison < poisonCount; poison++ {
			if poisonCount > 1 {
				// On 64-bit hosts the old offset lies inside the wall pointer.
				// The predicate must neither consult nor dereference that byte.
				raw[guardSize+legacyStateOffset] = byte(poison)
			}
			before := append([]byte(nil), raw...)
			if got := secretWallHiddenMinimap472600(ptr); got != want {
				t.Fatalf("state=%d legacy byte=%d: hidden=%t, want %t", state, poison, got, want)
			}
			if !bytes.Equal(raw, before) {
				t.Fatalf("state=%d legacy byte=%d: minimap predicate changed record or guards", state, poison)
			}
			checks++
		}
	}
	t.Logf("checked %d CGo state/legacy-offset combinations with unchanged guarded C memory", checks)
	runtime.KeepAlive(secret)
	runtime.KeepAlive(next)
	runtime.KeepAlive(wall)
}

func TestMinimap472600UsesNativeSecretWallState(t *testing.T) {
	source, err := os.ReadFile("GAME2_1.c")
	if err != nil {
		t.Fatal(err)
	}
	const signature = "int nox_xxx_cliDrawMinimap_472600("
	if strings.Count(string(source), signature) != 1 {
		t.Fatal("production minimap signature is not unique")
	}
	_, body, _ := strings.Cut(string(source), signature)
	body, _, ok := strings.Cut(body, "\n//----- (004730D0)")
	if !ok {
		t.Fatal("production minimap boundary is missing")
	}
	const call = "nox_secret_wall_hidden_minimap_472600((const nox_secret_wall_t*)nox_server_wallData(v11))"
	if strings.Count(body, call) != 1 || strings.Contains(body, "wallData[21]") {
		t.Fatal("production minimap does not use the tested native secret-wall predicate")
	}
}
