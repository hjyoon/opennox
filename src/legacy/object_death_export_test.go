package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestCreateSpawnObjectDieExportsKeepNativePointerWidth54E010(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	source := new(server.Object)
	if uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
		t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
	}

	oldCreate := createObjectDieCall54E010
	oldSpawn := spawnObjectDieCall54E070
	t.Cleanup(func() {
		createObjectDieCall54E010 = oldCreate
		spawnObjectDieCall54E070 = oldSpawn
	})
	var createCalls, spawnCalls int
	createObjectDieCall54E010 = func(got *server.Object) {
		createCalls++
		if got != source {
			t.Fatalf("CreateObjectDie source = %p, want %p", got, source)
		}
	}
	spawnObjectDieCall54E070 = func(got *server.Object) {
		spawnCalls++
		if got != source {
			t.Fatalf("SpawnObjectDie source = %p, want %p", got, source)
		}
	}

	createObjectDieExportCall54E010(source)
	spawnObjectDieExportCall54E070(source)
	if createCalls != 1 || spawnCalls != 1 {
		t.Fatalf("export calls = create:%d spawn:%d, want 1/1", createCalls, spawnCalls)
	}
	runtime.KeepAlive(source)
}
