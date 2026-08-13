package opennox

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
)

func TestSpellProjectileExpire4E71F0NoTargetDeletesWithoutReadingUpdateData(t *testing.T) {
	type projectile struct{ id int }
	type update struct{}
	obj := &projectile{id: 1}
	var events []string

	spellProjectileExpire4E71F0(obj, spellProjectileExpire4E71F0Hooks[*projectile, *update]{
		updateData: func(got *projectile) *update {
			events = append(events, "update")
			if got != obj {
				t.Fatalf("update object: got %p, want %p", got, obj)
			}
			return nil
		},
		search: func(got *projectile, radius float32, arg *targetSearchArg4E6EA0[*projectile]) *projectile {
			events = append(events, "search")
			if got != obj {
				t.Fatalf("search object: got %p, want %p", got, obj)
			}
			if radius != 50 {
				t.Fatalf("radius: got %v, want 50", radius)
			}
			assertSpellProjectileSearchArg4E71F0(t, arg)
			return nil
		},
		level:  func(*update) int32 { t.Fatal("level read without target"); return 0 },
		owner:  func(*update) *projectile { t.Fatal("owner read without target"); return nil },
		spell:  func(*update) int32 { t.Fatal("spell read without target"); return 0 },
		source: func(*update) *projectile { t.Fatal("source read without target"); return nil },
		accept: func(int32, *projectile, *projectile, *projectile, *projectile, int32) int32 {
			t.Fatal("spell accepted without target")
			return 0
		},
		delayedDelete: func(got *projectile) {
			events = append(events, "delete")
			if got != obj {
				t.Fatalf("delete object: got %p, want %p", got, obj)
			}
		},
	})

	if want := []string{"update", "search", "delete"}; fmt.Sprint(events) != fmt.Sprint(want) {
		t.Fatalf("events: got %v, want %v", events, want)
	}
}

func TestSpellProjectileExpire4E71F0NilObjectFaultsBeforeSearch(t *testing.T) {
	type projectile struct{ update int }
	var events []string
	h := spellProjectileExpire4E71F0Hooks[*projectile, int]{
		updateData: func(got *projectile) int {
			events = append(events, "update")
			return got.update
		},
		search: func(*projectile, float32, *targetSearchArg4E6EA0[*projectile]) *projectile {
			t.Fatal("search called after nil object fault")
			return nil
		},
		level:  func(int) int32 { t.Fatal("level called after nil object fault"); return 0 },
		owner:  func(int) *projectile { t.Fatal("owner called after nil object fault"); return nil },
		spell:  func(int) int32 { t.Fatal("spell called after nil object fault"); return 0 },
		source: func(int) *projectile { t.Fatal("source called after nil object fault"); return nil },
		accept: func(int32, *projectile, *projectile, *projectile, *projectile, int32) int32 {
			t.Fatal("accept called after nil object fault")
			return 0
		},
		delayedDelete: func(*projectile) { t.Fatal("delete called after nil object fault") },
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		spellProjectileExpire4E71F0[*projectile](nil, h)
	}()
	if recovered == nil {
		t.Fatal("nil object did not fault on update-data read")
	}
	if want := []string{"update"}; fmt.Sprint(events) != fmt.Sprint(want) {
		t.Fatalf("events: got %v, want %v", events, want)
	}
}

func TestSpellProjectileExpire4E71F0CachesUpdateAndPreservesReadOrder(t *testing.T) {
	type projectile struct{ id int }
	type update struct {
		level  int32
		owner  *projectile
		spell  int32
		source *projectile
	}
	obj := &projectile{id: 1}
	target := &projectile{id: 2}
	owner1 := &projectile{id: 3}
	owner2 := &projectile{id: 4}
	source := &projectile{id: 5}
	oldUD := &update{level: -1, owner: owner1, spell: -2, source: source}
	newUD := &update{level: 7, owner: owner2, spell: 8, source: owner2}
	currentUD := oldUD
	var events []string

	spellProjectileExpire4E71F0(obj, spellProjectileExpire4E71F0Hooks[*projectile, *update]{
		updateData: func(got *projectile) *update {
			events = append(events, "update")
			return currentUD
		},
		search: func(got *projectile, radius float32, arg *targetSearchArg4E6EA0[*projectile]) *projectile {
			events = append(events, "search")
			assertSpellProjectileSearchArg4E71F0(t, arg)
			currentUD = newUD
			return target
		},
		level: func(got *update) int32 {
			events = append(events, "level")
			if got != oldUD {
				t.Fatalf("level update pointer: got %p, want cached %p", got, oldUD)
			}
			return got.level
		},
		owner: func(got *update) *projectile {
			events = append(events, "owner")
			owner := got.owner
			got.owner = owner2
			return owner
		},
		spell: func(got *update) int32 {
			events = append(events, "spell")
			return got.spell
		},
		source: func(got *update) *projectile {
			events = append(events, "source")
			return got.source
		},
		accept: func(spellID int32, gotSource, gotOwner3, gotOwner4, gotTarget *projectile, level int32) int32 {
			events = append(events, "accept")
			if spellID != -2 || level != -1 {
				t.Fatalf("signed values: spell=%d level=%d, want -2/-1", spellID, level)
			}
			if gotSource != source || gotOwner3 != owner1 || gotOwner4 != owner1 || gotTarget != target {
				t.Fatalf("accept args: source=%p owners=%p/%p target=%p", gotSource, gotOwner3, gotOwner4, gotTarget)
			}
			return -7
		},
		delayedDelete: func(got *projectile) {
			events = append(events, "delete")
			if got != obj {
				t.Fatalf("delete object: got %p, want %p", got, obj)
			}
		},
	})

	want := []string{"update", "search", "level", "owner", "spell", "source", "accept", "delete"}
	if fmt.Sprint(events) != fmt.Sprint(want) {
		t.Fatalf("events: got %v, want %v", events, want)
	}
}

func TestSpellProjectileExpire4E71F0NilUpdateFaultsOnlyAfterTarget(t *testing.T) {
	type projectile struct{ id int }
	type update struct{ level int32 }
	obj := &projectile{id: 1}
	target := &projectile{id: 2}
	var events []string
	h := spellProjectileExpire4E71F0Hooks[*projectile, *update]{
		updateData: func(*projectile) *update {
			events = append(events, "update")
			return nil
		},
		search: func(*projectile, float32, *targetSearchArg4E6EA0[*projectile]) *projectile {
			events = append(events, "search")
			return target
		},
		level: func(got *update) int32 {
			events = append(events, "level")
			return got.level
		},
		owner:  func(*update) *projectile { t.Fatal("owner read after level fault"); return nil },
		spell:  func(*update) int32 { t.Fatal("spell read after level fault"); return 0 },
		source: func(*update) *projectile { t.Fatal("source read after level fault"); return nil },
		accept: func(int32, *projectile, *projectile, *projectile, *projectile, int32) int32 {
			t.Fatal("accept called after level fault")
			return 0
		},
		delayedDelete: func(*projectile) { t.Fatal("delete called after level fault") },
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		spellProjectileExpire4E71F0(obj, h)
	}()
	if recovered == nil {
		t.Fatal("nil update data did not fault on level read")
	}
	if want := []string{"update", "search", "level"}; fmt.Sprint(events) != fmt.Sprint(want) {
		t.Fatalf("events: got %v, want %v", events, want)
	}
}

func assertSpellProjectileSearchArg4E71F0[T comparable](t *testing.T, arg *targetSearchArg4E6EA0[T]) {
	t.Helper()
	if arg.Field0 != 15 || arg.Field4 != 1 || arg.Field8 != 0 ||
		arg.ClassAllow12 != object.MaskUnits || arg.ClassDisallow16 != 0 ||
		arg.SubClassAllow20 != object.SubClass(^uint32(0)) || arg.SubClassDisallow24 != 0 ||
		arg.FlagsAllow28 != object.Flags(^uint32(0)) ||
		arg.FlagsDisallow32 != object.FlagDead|object.FlagDestroyed {
		t.Fatalf("search arg mismatch: %+v", *arg)
	}
}
