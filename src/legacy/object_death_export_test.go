package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

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
	createCallback, _, ok := server.ObjectDeathHandler("CreateObjectDie")
	if !ok {
		t.Fatal("CreateObjectDie is not registered")
	}
	spawnCallback, _, ok := server.ObjectDeathHandler("SpawnObjectDie")
	if !ok {
		t.Fatal("SpawnObjectDie is not registered")
	}
	server.CallObjectDeath(createCallback, source)
	server.CallObjectDeath(spawnCallback, source)
	if createCalls != 2 || spawnCalls != 2 {
		t.Fatalf("export/native calls = create:%d spawn:%d, want 2/2", createCalls, spawnCalls)
	}
	runtime.KeepAlive(source)
}

func TestPlayerGlyphDieExportsAndDispatchKeepNativePointerWidth(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	obj := &server.Object{}
	if uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
		t.Fatalf("object pointer = %p, want address above the ABI32 range", obj)
	}

	oldPlayer := playerDieCall54D2B0
	oldGlyph := glyphDieCall54DF30
	t.Cleanup(func() {
		playerDieCall54D2B0 = oldPlayer
		glyphDieCall54DF30 = oldGlyph
	})

	var playerCalls, glyphCalls int
	check := func(name string, got *server.Object) {
		if got != obj {
			t.Fatalf("%s object = %p, want %p", name, got, obj)
		}
	}
	playerDieCall54D2B0 = func(got *server.Object) {
		playerCalls++
		check("PlayerDie", got)
	}
	glyphDieCall54DF30 = func(got *server.Object) {
		glyphCalls++
		check("GlyphDie", got)
	}

	playerDieExportCall54D2B0(obj)
	glyphDieExportCall54DF30(obj)
	for _, tc := range []struct {
		name string
	}{
		{name: "PlayerDie"},
		{name: "GlyphDie"},
	} {
		callback, size, ok := server.ObjectDeathHandler(tc.name)
		if !ok || callback == nil || size != 0 {
			t.Fatalf("ObjectDeathHandler(%q) = %p/%d/%t, want non-nil/0/true", tc.name, callback, size, ok)
		}
		server.CallObjectDeath(callback, obj)
	}
	if playerCalls != 2 || glyphCalls != 2 {
		t.Fatalf("export/native calls = player:%d glyph:%d, want 2/2", playerCalls, glyphCalls)
	}
	runtime.KeepAlive(obj)
}

func TestChestCollideDispatchesRegisteredDeathNatively4E9C40(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("native-width routing regression applies to 64-bit builds")
	}

	createCallback, _, ok := server.ObjectDeathHandler("CreateObjectDie")
	if !ok {
		t.Fatal("CreateObjectDie is not registered")
	}
	spawnCallback, _, ok := server.ObjectDeathHandler("SpawnObjectDie")
	if !ok {
		t.Fatal("SpawnObjectDie is not registered")
	}

	oldCreate := createObjectDieCall54E010
	oldSpawn := spawnObjectDieCall54E070
	t.Cleanup(func() {
		createObjectDieCall54E010 = oldCreate
		spawnObjectDieCall54E070 = oldSpawn
	})

	var active string
	var activeSource *server.Object
	var events []string
	createObjectDieCall54E010 = func(got *server.Object) {
		events = append(events, "death:create")
		if active != "create" || got != activeSource {
			t.Fatalf("CreateObjectDie dispatch = %q/%p, want create/%p", active, got, activeSource)
		}
	}
	spawnObjectDieCall54E070 = func(got *server.Object) {
		events = append(events, "death:spawn")
		if active != "spawn" || got != activeSource {
			t.Fatalf("SpawnObjectDie dispatch = %q/%p, want spawn/%p", active, got, activeSource)
		}
	}

	for _, tc := range []struct {
		name     string
		callback unsafe.Pointer
	}{
		{name: "create", callback: createCallback},
		{name: "spawn", callback: spawnCallback},
	} {
		t.Run(tc.name, func(t *testing.T) {
			active = tc.name
			events = events[:0]
			source := &server.Object{Death: tc.callback}
			activeSource = source
			target := &server.Object{ObjClass: object.ClassPlayer}
			if uintptr(source.CObj()) <= math.MaxUint32 {
				t.Fatalf("source pointer = %p, want address above the ABI32 range", source)
			}

			(&server.Server{}).ChestCollide4E9C40(source, target, nil, server.ChestCollideRuntime4E9C40{
				ChestOpen: func(gotSource, gotTarget *server.Object) {
					events = append(events, "open")
					if gotSource != source || gotTarget != target {
						t.Fatalf("ChestOpen objects = %p/%p, want %p/%p", gotSource, gotTarget, source, target)
					}
				},
				DropAllItems: func(got *server.Object) {
					events = append(events, "drop")
					if got != source {
						t.Fatalf("DropAllItems object = %p, want %p", got, source)
					}
				},
			})

			want := []string{"death:" + tc.name, "open", "drop"}
			if len(events) != len(want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
			for i := range want {
				if events[i] != want[i] {
					t.Fatalf("events = %v, want %v", events, want)
				}
			}
			runtime.KeepAlive(source)
			runtime.KeepAlive(target)
		})
	}
}
