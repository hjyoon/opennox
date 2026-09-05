package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type spellDurationFreeRecursiveLegacyServer4FE980 struct {
	Server
	srv *server.Server
}

func (s *spellDurationFreeRecursiveLegacyServer4FE980) S() *server.Server {
	return s.srv
}

func useSpellDurationFreeRecursiveLegacyServer4FE980(t *testing.T) *server.Server {
	t.Helper()
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &spellDurationFreeRecursiveLegacyServer4FE980{srv: srv}
	}
	t.Cleanup(func() {
		GetServer = oldGetServer
	})
	return srv
}

func TestSpellDurationFreeRecursiveExport4FE980PreservesNativeTreeAndReuseOrder(t *testing.T) {
	srv := useSpellDurationFreeRecursiveLegacyServer4FE980(t)
	spells := &srv.Spells.Dur
	if got := spells.SpellCreateDurations4FE850(); got != 1 {
		t.Fatalf("allocator result = %d, want canonical 1", got)
	}
	t.Cleanup(spells.SpellFreeDurations4FE880)

	root := spells.NewRaw()
	sub108A := spells.NewRaw()
	sub108B := spells.NewRaw()
	grandchild := spells.NewRaw()
	sub104 := spells.NewRaw()
	for name, record := range map[string]*server.DurSpell{
		"root":       root,
		"Sub108 A":   sub108A,
		"Sub108 B":   sub108B,
		"grandchild": grandchild,
		"Sub104":     sub104,
	} {
		if record == nil {
			t.Fatalf("%s allocation returned nil", name)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(record)) <= math.MaxUint32 {
			t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, record)
		}
	}
	root.Sub108 = sub108A
	root.Sub104 = sub104
	sub108A.Next = sub108B
	sub108A.Sub108 = grandchild

	spellDurationFreeRecursiveExportCall4FE980(unsafe.Pointer(root))

	wantReuse := []*server.DurSpell{root, sub104, sub108B, sub108A, grandchild}
	for i, want := range wantReuse {
		got := spells.NewRaw()
		if got != want {
			t.Fatalf("CGo reuse %d = %p, want recursive FreeObjectFirst order %p", i, got, want)
		}
		if got.ID != uint16(i+6) {
			t.Fatalf("CGo reuse %d ID = %d, want %d", i, got.ID, i+6)
		}
	}
	runtime.KeepAlive(root)
	runtime.KeepAlive(sub108A)
	runtime.KeepAlive(sub108B)
	runtime.KeepAlive(grandchild)
	runtime.KeepAlive(sub104)
}
