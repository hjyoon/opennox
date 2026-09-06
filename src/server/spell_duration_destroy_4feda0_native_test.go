package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	playerlib "github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func requireSpellDurationDestroyNativePointers4FEDA0(t *testing.T, values ...unsafe.Pointer) {
	t.Helper()
	if unsafe.Sizeof(uintptr(0)) != 8 {
		return
	}
	for i, value := range values {
		if value == nil || uintptr(value) <= math.MaxUint32 {
			t.Fatalf("pointer %d = %p, want native address above 4 GiB", i, value)
		}
	}
}

func TestSpellDurationDestroy4FEDA0NativeLayouts(t *testing.T) {
	type layout struct {
		durCaster    uintptr
		durDestroy   uintptr
		objectClass  uintptr
		objectUpdate uintptr
		updatePlayer uintptr
		playerClass  uintptr
	}
	wants := map[uintptr]layout{
		4: {
			durCaster: 16, durDestroy: 100, objectClass: 8,
			objectUpdate: 748, updatePlayer: 276, playerClass: 2251,
		},
		8: {
			durCaster: 24, durDestroy: 144, objectClass: 12,
			objectUpdate: 872, updatePlayer: 336, playerClass: 2255,
		},
	}
	ptrSize := unsafe.Sizeof(uintptr(0))
	want, ok := wants[ptrSize]
	if !ok {
		t.Fatalf("unsupported pointer size %d", ptrSize)
	}
	got := layout{
		durCaster:    unsafe.Offsetof(DurSpell{}.Caster16),
		durDestroy:   unsafe.Offsetof(DurSpell{}.Destroy),
		objectClass:  unsafe.Offsetof(Object{}.ObjClass),
		objectUpdate: unsafe.Offsetof(Object{}.UpdateData),
		updatePlayer: unsafe.Offsetof(PlayerUpdateData{}.Player),
		playerClass:  unsafe.Offsetof(Player{}.info) + unsafe.Offsetof(PlayerInfo{}.playerClass),
	}
	if got != want {
		t.Fatalf("native layout on %s/%s = %+v, want %+v", runtime.GOOS, runtime.GOARCH, got, want)
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"DurSpell.Spell", unsafe.Offsetof(DurSpell{}.Spell), 4},
		{"DurSpell caster pointer width", unsafe.Sizeof(DurSpell{}.Caster16), ptrSize},
		{"DurSpell callback pointer width", unsafe.Sizeof(DurSpell{}.Destroy), ptrSize},
		{"Object.ObjClass width", unsafe.Sizeof(Object{}.ObjClass), 4},
		{"PlayerState width", unsafe.Sizeof(PlayerState(0)), 1},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestSpellDurationDestroyNative4FEDA0PreservesPointersAndLiveReloads(t *testing.T) {
	player := new(Player)
	player.Info().SetPlayerClass(playerlib.Warrior)
	update := &PlayerUpdateData{Player: player}
	casterA := new(Object)
	casterB := &Object{
		ObjClass:   object.Class(0xffffff06),
		UpdateData: unsafe.Pointer(update),
	}
	casterC := new(Object)
	callbackA, callbackB := new(byte), new(byte)
	record := &DurSpell{
		Spell:    0x8000002b,
		Caster16: casterA,
		Destroy:  unsafe.Pointer(callbackA),
	}
	requireSpellDurationDestroyNativePointers4FEDA0(t,
		unsafe.Pointer(player), unsafe.Pointer(update), unsafe.Pointer(casterA),
		unsafe.Pointer(casterB), unsafe.Pointer(casterC), unsafe.Pointer(callbackA),
		unsafe.Pointer(callbackB), unsafe.Pointer(record),
	)

	var events []string
	spellDurationDestroyNative4FEDA0(record, spellDurationDestroyNativeDeps4FEDA0{
		spellAudio: func(spellID, selector int32) int32 {
			events = append(events, "spell-audio")
			if spellID != math.MinInt32+43 || selector != 2 {
				t.Fatalf("spell/selector = %d/%d, want %d/2", spellID, selector, int32(math.MinInt32+43))
			}
			record.Caster16 = casterC
			return -123
		},
		audioEvent: func(id int32, caster *Object, kind, code int32) {
			events = append(events, "audio")
			if id != -123 || caster != casterA || kind != 0 || code != 0 {
				t.Fatalf("audio args = %d/%p/%d/%d, want -123/%p/0/0", id, caster, kind, code, casterA)
			}
			record.Destroy = unsafe.Pointer(callbackB)
		},
		callDestroy: func(callback unsafe.Pointer, gotRecord *DurSpell) {
			events = append(events, "destroy")
			if callback != unsafe.Pointer(callbackB) || gotRecord != record {
				t.Fatalf("destroy args = %p/%p, want %p/%p", callback, gotRecord, callbackB, record)
			}
			record.Destroy = unsafe.Pointer(callbackA)
			record.Caster16 = casterB
		},
		abilityActive: func(caster *Object, ability int32) int32 {
			events = append(events, "ability")
			if caster != casterB || ability != 1 {
				t.Fatalf("ability args = %p/%d, want %p/1", caster, ability, casterB)
			}
			record.Caster16 = casterC
			return 0
		},
		setPlayerState: func(caster *Object, state int32) {
			events = append(events, "state")
			if caster != casterC || state != 13 {
				t.Fatalf("state args = %p/%d, want %p/13", caster, state, casterC)
			}
		},
		monsterCancel: func(*Object, int32) {
			t.Fatal("Player|Monster class reached Monster path")
		},
		unlink: func(got *DurSpell) {
			events = append(events, "unlink")
			if got != record {
				t.Fatalf("unlink record = %p, want %p", got, record)
			}
		},
		freeRecursive: func(got *DurSpell) {
			events = append(events, "free")
			if got != record {
				t.Fatalf("free record = %p, want %p", got, record)
			}
		},
	})
	wantEvents := []string{"spell-audio", "audio", "destroy", "ability", "state", "unlink", "free"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want %q", events, wantEvents)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(casterA)
	runtime.KeepAlive(casterB)
	runtime.KeepAlive(casterC)
	runtime.KeepAlive(callbackA)
	runtime.KeepAlive(callbackB)
	runtime.KeepAlive(record)
}

func TestSpellDurationDestroyNative4FEDA0MonsterUsesLiveSpell(t *testing.T) {
	caster := &Object{ObjClass: object.Class(0xffffff82)}
	record := &DurSpell{Spell: 31, Caster16: caster}
	var cancelledSpell int32
	spellDurationDestroyNative4FEDA0(record, spellDurationDestroyNativeDeps4FEDA0{
		spellAudio: func(spellID, selector int32) int32 {
			if spellID != 31 || selector != 2 {
				t.Fatalf("spell/selector = %d/%d, want 31/2", spellID, selector)
			}
			record.Spell = 59
			return 123
		},
		audioEvent: func(int32, *Object, int32, int32) {},
		callDestroy: func(unsafe.Pointer, *DurSpell) {
			t.Fatal("nil destroy callback was called")
		},
		abilityActive: func(*Object, int32) int32 {
			t.Fatal("Monster reached Player ability path")
			return 0
		},
		setPlayerState: func(*Object, int32) {
			t.Fatal("Monster reached Player state path")
		},
		monsterCancel: func(gotCaster *Object, spellID int32) {
			if gotCaster != caster {
				t.Fatalf("cancel caster = %p, want %p", gotCaster, caster)
			}
			cancelledSpell = spellID
		},
		unlink:        func(*DurSpell) {},
		freeRecursive: func(*DurSpell) {},
	})
	if cancelledSpell != 59 {
		t.Fatalf("cancelled spell = %d, want live value 59", cancelledSpell)
	}
}

func TestSpellDurationDestroyNative4FEDA0NilPlayerFaults(t *testing.T) {
	update := new(PlayerUpdateData)
	record := &DurSpell{
		Caster16: &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)},
	}
	defer func() {
		if got := recover(); got == nil {
			t.Fatal("nil Player did not fault at the original Player+2251 load")
		}
	}()
	spellDurationDestroyNative4FEDA0(record, spellDurationDestroyNativeDeps4FEDA0{
		spellAudio:     func(int32, int32) int32 { return 0 },
		audioEvent:     func(int32, *Object, int32, int32) {},
		callDestroy:    func(unsafe.Pointer, *DurSpell) {},
		abilityActive:  func(*Object, int32) int32 { return 0 },
		setPlayerState: func(*Object, int32) {},
		monsterCancel:  func(*Object, int32) {},
		unlink:         func(*DurSpell) {},
		freeRecursive:  func(*DurSpell) {},
	})
}

func TestSpellDurationDestroy4FEDA0ServerBinding(t *testing.T) {
	var srv Server
	srv.Spells.byID = map[spell.ID]*SpellDef{
		31: {OffSound: sound.ID(987)},
	}
	srv.Spells.init(&srv)
	srv.Abils.init(&srv)
	if got := srv.Spells.Dur.SpellCreateDurations4FE850(); got != 1 {
		t.Fatalf("allocator result = %d, want 1", got)
	}
	t.Cleanup(srv.Spells.Dur.Free)

	unit, freeUnit := alloc.New(Object{})
	update, freeUpdate := alloc.New(PlayerUpdateData{})
	player, freePlayer := alloc.New(Player{})
	callback, freeCallback := alloc.New(byte(0))
	t.Cleanup(func() {
		freeUnit()
		freeUpdate()
		freePlayer()
		freeCallback()
	})
	player.Info().SetPlayerClass(playerlib.Wizard)
	update.Player = player
	unit.ObjClass = object.ClassPlayer
	unit.UpdateData = unsafe.Pointer(update)
	record := srv.Spells.Dur.SpellDurationNew4FE950()
	if record == nil {
		t.Fatal("duration allocator returned nil")
	}
	record.Spell = 31
	record.Caster16 = unit
	record.Destroy = unsafe.Pointer(callback)
	srv.Spells.Dur.SpellDurationInsert4FED40(record)
	requireSpellDurationDestroyNativePointers4FEDA0(t,
		unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(player),
		unsafe.Pointer(callback), unsafe.Pointer(record),
	)

	var destroyCalls, stateCalls int
	srv.Spells.Dur.SpellDurationDestroy4FEDA0(record, SpellDurationDestroyRuntime4FEDA0{
		CallDestroy: func(gotCallback unsafe.Pointer, gotRecord *DurSpell) {
			destroyCalls++
			if gotCallback != unsafe.Pointer(callback) || gotRecord != record {
				t.Fatalf("destroy args = %p/%p, want %p/%p", gotCallback, gotRecord, callback, record)
			}
		},
		SetPlayerState: func(gotUnit *Object, state PlayerState) {
			stateCalls++
			if gotUnit != unit || state != PlayerState13 {
				t.Fatalf("state args = %p/%d, want %p/13", gotUnit, state, unit)
			}
		},
	})
	if destroyCalls != 1 || stateCalls != 1 {
		t.Fatalf("destroy/state calls = %d/%d, want 1/1", destroyCalls, stateCalls)
	}
	if srv.Spells.Dur.List != nil {
		t.Fatalf("duration list head = %p, want nil", srv.Spells.Dur.List)
	}
	if len(srv.Audio.delayedObj) != 1 {
		t.Fatalf("queued audio events = %d, want 1", len(srv.Audio.delayedObj))
	}
	audio := srv.Audio.delayedObj[0]
	if audio.ID != 987 || audio.Obj != unit || audio.Kind != 0 || audio.Code != 0 {
		t.Fatalf("queued audio = %+v, want 987/%p/0/0", audio, unit)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
	runtime.KeepAlive(callback)
}
