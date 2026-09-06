package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	spellDurationDestroyRecord4FEDA0    = uint64(0x100000101)
	spellDurationDestroyCasterA4FEDA0   = uint64(0x200000202)
	spellDurationDestroyCasterB4FEDA0   = uint64(0x300000303)
	spellDurationDestroyCasterC4FEDA0   = uint64(0x400000404)
	spellDurationDestroyUpdate4FEDA0    = uint64(0x500000505)
	spellDurationDestroyPlayer4FEDA0    = uint64(0x600000606)
	spellDurationDestroyCallbackA4FEDA0 = uint64(0x700000707)
	spellDurationDestroyCallbackB4FEDA0 = uint64(0x800000808)
)

type spellDurationDestroyRecordState4FEDA0 struct {
	caster  uint64
	spell   uint32
	destroy uint64
}

type spellDurationDestroyObjectState4FEDA0 struct {
	class  byte
	update uint64
}

type spellDurationDestroyWorld4FEDA0 struct {
	records       map[uint64]*spellDurationDestroyRecordState4FEDA0
	objects       map[uint64]*spellDurationDestroyObjectState4FEDA0
	players       map[uint64]uint64
	playerClasses map[uint64]byte
	audioID       int32
	abilityResult int32
	events        []string
	after         map[string]func()
	faultAt       int
}

func newSpellDurationDestroyWorld4FEDA0() *spellDurationDestroyWorld4FEDA0 {
	return &spellDurationDestroyWorld4FEDA0{
		records: map[uint64]*spellDurationDestroyRecordState4FEDA0{
			spellDurationDestroyRecord4FEDA0: {
				caster:  spellDurationDestroyCasterA4FEDA0,
				spell:   0x8000002b,
				destroy: spellDurationDestroyCallbackA4FEDA0,
			},
		},
		objects: map[uint64]*spellDurationDestroyObjectState4FEDA0{
			spellDurationDestroyCasterA4FEDA0: {
				class:  spellDurationDestroyPlayerClass4FEDA0,
				update: spellDurationDestroyUpdate4FEDA0,
			},
			spellDurationDestroyCasterB4FEDA0: {
				class:  spellDurationDestroyPlayerClass4FEDA0 | spellDurationDestroyMonsterClass4FEDA0,
				update: spellDurationDestroyUpdate4FEDA0,
			},
			spellDurationDestroyCasterC4FEDA0: {},
		},
		players: map[uint64]uint64{
			spellDurationDestroyUpdate4FEDA0: spellDurationDestroyPlayer4FEDA0,
		},
		playerClasses: map[uint64]byte{
			spellDurationDestroyPlayer4FEDA0: 2,
		},
		audioID: -123,
		after:   make(map[string]func()),
	}
}

func (w *spellDurationDestroyWorld4FEDA0) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *spellDurationDestroyWorld4FEDA0) record(token uint64) *spellDurationDestroyRecordState4FEDA0 {
	state := w.records[token]
	if state == nil {
		panic(fmt.Sprintf("record:%x", token))
	}
	return state
}

func (w *spellDurationDestroyWorld4FEDA0) object(token uint64) *spellDurationDestroyObjectState4FEDA0 {
	state := w.objects[token]
	if state == nil {
		panic(fmt.Sprintf("object:%x", token))
	}
	return state
}

func (w *spellDurationDestroyWorld4FEDA0) hooks() SpellDurationDestroyHooks4FEDA0[uint64, uint64, uint64, uint64, uint64] {
	return SpellDurationDestroyHooks4FEDA0[uint64, uint64, uint64, uint64, uint64]{
		LoadCaster: func(record uint64) uint64 {
			value := w.record(record).caster
			w.observe(fmt.Sprintf("caster:%x:%x", record, value))
			return value
		},
		LoadSpell: func(record uint64) uint32 {
			value := w.record(record).spell
			w.observe(fmt.Sprintf("spell:%x:%x", record, value))
			return value
		},
		SpellAudio: func(spellID, selector int32) int32 {
			value := w.audioID
			w.observe(fmt.Sprintf("spell-audio:%d:%d", spellID, selector))
			return value
		},
		AudioEvent: func(audio int32, object uint64, kind, code int32) {
			w.observe(fmt.Sprintf("audio:%d:%x:%d:%d", audio, object, kind, code))
		},
		LoadDestroy: func(record uint64) uint64 {
			value := w.record(record).destroy
			w.observe(fmt.Sprintf("destroy:%x:%x", record, value))
			return value
		},
		CallDestroy: func(callback, record uint64) {
			w.observe(fmt.Sprintf("call-destroy:%x:%x", callback, record))
		},
		LoadObjectClassLow: func(object uint64) byte {
			value := w.object(object).class
			w.observe(fmt.Sprintf("class:%x:%02x", object, value))
			return value
		},
		LoadPlayerUpdate: func(object uint64) uint64 {
			value := w.object(object).update
			w.observe(fmt.Sprintf("player-update:%x:%x", object, value))
			return value
		},
		LoadPlayer: func(update uint64) uint64 {
			value := w.players[update]
			w.observe(fmt.Sprintf("player:%x:%x", update, value))
			return value
		},
		LoadPlayerClass: func(player uint64) byte {
			value := w.playerClasses[player]
			w.observe(fmt.Sprintf("player-class:%x:%d", player, value))
			return value
		},
		AbilityActive: func(object uint64, ability int32) int32 {
			value := w.abilityResult
			w.observe(fmt.Sprintf("ability:%x:%d:%d", object, ability, value))
			return value
		},
		SetPlayerState: func(object uint64, state int32) {
			w.observe(fmt.Sprintf("state:%x:%d", object, state))
		},
		MonsterCancel: func(object uint64, spellID int32) {
			w.observe(fmt.Sprintf("monster-cancel:%x:%d", object, spellID))
		},
		Unlink: func(record uint64) {
			w.observe(fmt.Sprintf("unlink:%x", record))
		},
		FreeRecursive: func(record uint64) {
			w.observe(fmt.Sprintf("free:%x", record))
		},
	}
}

func TestSpellDurationDestroy4FEDA0ExactPlayerTraceAndLiveReloads(t *testing.T) {
	w := newSpellDurationDestroyWorld4FEDA0()
	w.after["spell-audio:-2147483605:2"] = func() {
		w.record(spellDurationDestroyRecord4FEDA0).caster = spellDurationDestroyCasterC4FEDA0
	}
	w.after["destroy:100000101:700000707"] = func() {
		w.record(spellDurationDestroyRecord4FEDA0).destroy = spellDurationDestroyCallbackB4FEDA0
	}
	w.after["call-destroy:700000707:100000101"] = func() {
		w.record(spellDurationDestroyRecord4FEDA0).caster = spellDurationDestroyCasterB4FEDA0
	}
	w.after["player-class:600000606:2"] = func() {
		w.record(spellDurationDestroyRecord4FEDA0).caster = spellDurationDestroyCasterC4FEDA0
	}

	SpellDurationDestroy4FEDA0(spellDurationDestroyRecord4FEDA0, w.hooks())

	want := []string{
		"caster:100000101:200000202",
		"spell:100000101:8000002b",
		"spell-audio:-2147483605:2",
		"audio:-123:200000202:0:0",
		"destroy:100000101:700000707",
		"call-destroy:700000707:100000101",
		"caster:100000101:300000303",
		"class:300000303:06",
		"player-update:300000303:500000505",
		"player:500000505:600000606",
		"player-class:600000606:2",
		"caster:100000101:400000404",
		"state:400000404:13",
		"unlink:100000101",
		"free:100000101",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("trace mismatch\n got: %q\nwant: %q", w.events, want)
	}
}

func TestSpellDurationDestroy4FEDA0WarriorAbilityResultIsExactZero(t *testing.T) {
	for _, tc := range []struct {
		name          string
		abilityResult int32
		wantState     bool
	}{
		{name: "active-positive", abilityResult: 1},
		{name: "active-negative", abilityResult: -7},
		{name: "inactive-exact-zero", abilityResult: 0, wantState: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newSpellDurationDestroyWorld4FEDA0()
			w.playerClasses[spellDurationDestroyPlayer4FEDA0] = spellDurationDestroyWarrior4FEDA0
			w.abilityResult = tc.abilityResult
			w.after[fmt.Sprintf("ability:200000202:1:%d", tc.abilityResult)] = func() {
				w.record(spellDurationDestroyRecord4FEDA0).caster = spellDurationDestroyCasterC4FEDA0
			}

			SpellDurationDestroy4FEDA0(spellDurationDestroyRecord4FEDA0, w.hooks())

			stateEvent := "state:400000404:13"
			gotState := false
			for _, event := range w.events {
				if event == stateEvent {
					gotState = true
				}
			}
			if gotState != tc.wantState {
				t.Fatalf("state event present = %v, want %v; trace %q", gotState, tc.wantState, w.events)
			}
			wantTail := []string{"unlink:100000101", "free:100000101"}
			if got := w.events[len(w.events)-2:]; !reflect.DeepEqual(got, wantTail) {
				t.Fatalf("cleanup tail = %q, want %q", got, wantTail)
			}
		})
	}
}

func TestSpellDurationDestroy4FEDA0MonsterUsesLiveSpellAndPlayerPrecedence(t *testing.T) {
	t.Run("monster", func(t *testing.T) {
		w := newSpellDurationDestroyWorld4FEDA0()
		w.record(spellDurationDestroyRecord4FEDA0).destroy = 0
		w.object(spellDurationDestroyCasterA4FEDA0).class = 0x82
		w.after["class:200000202:82"] = func() {
			w.record(spellDurationDestroyRecord4FEDA0).spell = 59
		}

		SpellDurationDestroy4FEDA0(spellDurationDestroyRecord4FEDA0, w.hooks())

		wantTail := []string{
			"class:200000202:82",
			"spell:100000101:3b",
			"monster-cancel:200000202:59",
			"unlink:100000101",
			"free:100000101",
		}
		if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
			t.Fatalf("monster tail = %q, want %q", got, wantTail)
		}
	})

	t.Run("player-precedes-monster", func(t *testing.T) {
		w := newSpellDurationDestroyWorld4FEDA0()
		w.object(spellDurationDestroyCasterA4FEDA0).class = 6

		SpellDurationDestroy4FEDA0(spellDurationDestroyRecord4FEDA0, w.hooks())

		for _, event := range w.events {
			if len(event) >= len("monster-cancel:") && event[:len("monster-cancel:")] == "monster-cancel:" {
				t.Fatalf("Player|Monster class reached Monster path: %q", w.events)
			}
		}
	})
}

func TestSpellDurationDestroy4FEDA0NilCasterPhases(t *testing.T) {
	t.Run("initial-nil-still-runs-destroy", func(t *testing.T) {
		w := newSpellDurationDestroyWorld4FEDA0()
		w.record(spellDurationDestroyRecord4FEDA0).caster = 0
		w.after["call-destroy:700000707:100000101"] = func() {
			w.record(spellDurationDestroyRecord4FEDA0).caster = spellDurationDestroyCasterA4FEDA0
		}

		SpellDurationDestroy4FEDA0(spellDurationDestroyRecord4FEDA0, w.hooks())

		wantPrefix := []string{
			"caster:100000101:0",
			"destroy:100000101:700000707",
			"call-destroy:700000707:100000101",
			"caster:100000101:200000202",
		}
		if got := w.events[:len(wantPrefix)]; !reflect.DeepEqual(got, wantPrefix) {
			t.Fatalf("prefix = %q, want %q", got, wantPrefix)
		}
	})

	t.Run("destroy-clears-caster", func(t *testing.T) {
		w := newSpellDurationDestroyWorld4FEDA0()
		w.after["call-destroy:700000707:100000101"] = func() {
			w.record(spellDurationDestroyRecord4FEDA0).caster = 0
		}

		SpellDurationDestroy4FEDA0(spellDurationDestroyRecord4FEDA0, w.hooks())

		wantTail := []string{
			"call-destroy:700000707:100000101",
			"caster:100000101:0",
			"unlink:100000101",
			"free:100000101",
		}
		if got := w.events[len(w.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
			t.Fatalf("tail = %q, want %q", got, wantTail)
		}
	})
}

func TestSpellDurationDestroy4FEDA0FaultPrefixes(t *testing.T) {
	baseline := newSpellDurationDestroyWorld4FEDA0()
	SpellDurationDestroy4FEDA0(spellDurationDestroyRecord4FEDA0, baseline.hooks())
	want := append([]string(nil), baseline.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newSpellDurationDestroyWorld4FEDA0()
			w.faultAt = faultAt
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				SpellDurationDestroy4FEDA0(spellDurationDestroyRecord4FEDA0, w.hooks())
			}()
			if recovered == nil {
				t.Fatal("fault sentinel was not recovered")
			}
			if prefix := want[:faultAt]; !reflect.DeepEqual(w.events, prefix) {
				t.Fatalf("events = %q, want fault prefix %q", w.events, prefix)
			}
		})
	}
}

func TestSpellDurationDestroy4FEDA0DoesNotGuardNilRecord(t *testing.T) {
	w := newSpellDurationDestroyWorld4FEDA0()
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		SpellDurationDestroy4FEDA0(uint64(0), w.hooks())
	}()
	if recovered == nil {
		t.Fatal("nil record token did not fault on the first caster load")
	}
	if len(w.events) != 0 {
		t.Fatalf("events before nil-record fault = %q, want none", w.events)
	}
}
