package server

import (
	"reflect"
	"testing"
)

type awardSpellCollideTestData4EAD20 struct {
	spell uint32
}

type awardSpellCollideTestObject4EAD20 struct {
	data *awardSpellCollideTestData4EAD20
}

func TestAwardSpellCollide4EAD20NilTargetPrecedesSource(t *testing.T) {
	var events []string
	hooks := awardSpellCollideHooks4EAD20[
		*awardSpellCollideTestObject4EAD20,
		*awardSpellCollideTestData4EAD20,
	]{
		loadCollideData: func(*awardSpellCollideTestObject4EAD20) *awardSpellCollideTestData4EAD20 {
			t.Fatal("source read for nil target")
			return nil
		},
		loadSpell: func(*awardSpellCollideTestData4EAD20) uint32 {
			t.Fatal("spell read for nil target")
			return 0
		},
		grantSpell: func(*awardSpellCollideTestObject4EAD20, uint32, int32, int32, int32) int32 {
			t.Fatal("grant called for nil target")
			return 0
		},
	}

	got := awardSpellCollide4EAD20(
		(*awardSpellCollideTestObject4EAD20)(nil),
		(*awardSpellCollideTestObject4EAD20)(nil),
		&struct{ poison bool }{poison: true},
		hooks,
	)
	if got != 0 {
		t.Fatalf("return = %d, want 0", got)
	}
	if len(events) != 0 {
		t.Fatalf("events = %#v", events)
	}
}

func TestAwardSpellCollide4EAD20LoadsCachedDataAndReturnsFullGrantResult(t *testing.T) {
	data := &awardSpellCollideTestData4EAD20{spell: 0xf1234567}
	replacement := &awardSpellCollideTestData4EAD20{spell: 0x11223344}
	source := &awardSpellCollideTestObject4EAD20{data: data}
	target := &awardSpellCollideTestObject4EAD20{}
	collision := &struct{ poison bool }{poison: true}
	var events []string

	hooks := awardSpellCollideHooks4EAD20[
		*awardSpellCollideTestObject4EAD20,
		*awardSpellCollideTestData4EAD20,
	]{
		loadCollideData: func(obj *awardSpellCollideTestObject4EAD20) *awardSpellCollideTestData4EAD20 {
			events = append(events, "data")
			return obj.data
		},
		loadSpell: func(got *awardSpellCollideTestData4EAD20) uint32 {
			events = append(events, "spell")
			if got != data {
				t.Fatalf("data = %p, want cached %p", got, data)
			}
			source.data = replacement
			return got.spell
		},
		grantSpell: func(gotTarget *awardSpellCollideTestObject4EAD20, spell uint32, mode, fourth, fifth int32) int32 {
			events = append(events, "grant")
			if gotTarget != target {
				t.Fatalf("target = %p, want %p", gotTarget, target)
			}
			if spell != 0xf1234567 || mode != 1 || fourth != 0 || fifth != 0 {
				t.Fatalf("grant args = %#x/%d/%d/%d", spell, mode, fourth, fifth)
			}
			return int32(-0x1234567)
		},
	}

	got := awardSpellCollide4EAD20(source, target, collision, hooks)
	if got != int32(-0x1234567) {
		t.Fatalf("return = %#x", uint32(got))
	}
	if want := []string{"data", "spell", "grant"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestAwardSpellCollide4EAD20FaultOrder(t *testing.T) {
	t.Run("nil source faults after target guard", func(t *testing.T) {
		target := &awardSpellCollideTestObject4EAD20{}
		var events []string
		hooks := awardSpellCollideHooks4EAD20[
			*awardSpellCollideTestObject4EAD20,
			*awardSpellCollideTestData4EAD20,
		]{
			loadCollideData: func(obj *awardSpellCollideTestObject4EAD20) *awardSpellCollideTestData4EAD20 {
				events = append(events, "data")
				return obj.data
			},
			loadSpell: func(*awardSpellCollideTestData4EAD20) uint32 {
				t.Fatal("spell read after source fault")
				return 0
			},
		}

		var recovered any
		func() {
			defer func() { recovered = recover() }()
			awardSpellCollide4EAD20(
				(*awardSpellCollideTestObject4EAD20)(nil),
				target,
				struct{}{}, hooks,
			)
		}()
		if recovered == nil {
			t.Fatal("nil source did not fault")
		}
		if want := []string{"data"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %#v, want %#v", events, want)
		}
	})

	t.Run("nil data faults before grant", func(t *testing.T) {
		source := &awardSpellCollideTestObject4EAD20{}
		target := &awardSpellCollideTestObject4EAD20{}
		var events []string
		hooks := awardSpellCollideHooks4EAD20[
			*awardSpellCollideTestObject4EAD20,
			*awardSpellCollideTestData4EAD20,
		]{
			loadCollideData: func(obj *awardSpellCollideTestObject4EAD20) *awardSpellCollideTestData4EAD20 {
				events = append(events, "data")
				return obj.data
			},
			loadSpell: func(data *awardSpellCollideTestData4EAD20) uint32 {
				events = append(events, "spell")
				return data.spell
			},
			grantSpell: func(*awardSpellCollideTestObject4EAD20, uint32, int32, int32, int32) int32 {
				t.Fatal("grant called after data fault")
				return 0
			},
		}

		var recovered any
		func() {
			defer func() { recovered = recover() }()
			awardSpellCollide4EAD20(source, target, struct{}{}, hooks)
		}()
		if recovered == nil {
			t.Fatal("nil data did not fault")
		}
		if want := []string{"data", "spell"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %#v, want %#v", events, want)
		}
	})
}
