package server

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// These expectations were checked by executing unchanged GAME.EXE 00532440
// (248 bytes, SHA256 e2f19a2a945248439c050c0152ebc2f7a6b681e5b39e0e23f0c45c95af52ea54)
// and its unhooked 00424300 getter (31 bytes, SHA256
// 11676dd3929f946c37a619bdcca10071055d69e4a91851643782069ea685fa7e).
// Strike/audio/pop are bounded service callbacks, not a complete AI world.
type meleeSoundEvent532440 struct {
	Kind  string
	Value int
}

func meleeSoundStorage532440[T comparable](t *testing.T, cOwned bool) *T {
	t.Helper()
	if !cOwned {
		return new(T)
	}
	var zero T
	p, free := alloc.New(zero)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= uintptr(^uint32(0)) {
		t.Fatalf("C-owned test address = %p, want above 4 GiB", p)
	}
	return p
}

type meleeSoundFixture532440 struct {
	unit                 *Object
	old, fresh           *MonsterUpdateData
	definition           *MonsterDef
	oldSounds, newSounds *[10]uint32
}

func newMeleeSoundFixture532440(t *testing.T, cOwned bool, oldDone, newDone uint8) meleeSoundFixture532440 {
	t.Helper()
	f := meleeSoundFixture532440{
		unit:       meleeSoundStorage532440[Object](t, cOwned),
		old:        meleeSoundStorage532440[MonsterUpdateData](t, cOwned),
		fresh:      meleeSoundStorage532440[MonsterUpdateData](t, cOwned),
		definition: meleeSoundStorage532440[MonsterDef](t, cOwned),
		oldSounds:  meleeSoundStorage532440[[10]uint32](t, cOwned),
		newSounds:  meleeSoundStorage532440[[10]uint32](t, cOwned),
	}
	f.unit.ObjClass = object.ClassMonster
	f.unit.UpdateData = unsafe.Pointer(f.old)
	f.old.Field120_0, f.old.Field120_1, f.old.Field120_3 = 8, 4, oldDone
	f.fresh.Field120_0, f.fresh.Field120_1, f.fresh.Field120_3 = 8, 4, newDone
	f.old.MonsterDef = f.definition
	f.definition.MeleeAttackFrame108 = 4
	f.definition.MeleeStrikeFunc236 = unsafe.Pointer(f.definition)
	f.oldSounds[8], f.oldSounds[9] = 1008, 1009
	f.newSounds[8], f.newSounds[9] = 2008, 2009
	f.old.SoundSet122 = unsafe.Pointer(f.oldSounds)
	f.fresh.SoundSet122 = unsafe.Pointer(f.newSounds)
	return f
}

func (f meleeSoundFixture532440) mutate(t *testing.T, mode string) {
	t.Helper()
	switch mode {
	case "retain":
	case "swap_update":
		f.unit.UpdateData = unsafe.Pointer(f.fresh)
	case "swap_sound":
		f.old.SoundSet122 = unsafe.Pointer(f.newSounds)
	case "nil_sound":
		f.old.SoundSet122 = nil
	case "class_player":
		f.unit.ObjClass = object.ClassPlayer
	case "class_none":
		f.unit.ObjClass = 0
	case "monster_nil_update":
		f.unit.UpdateData = nil
	case "player_nil_update":
		f.unit.ObjClass = object.ClassPlayer
		f.unit.UpdateData = nil
	default:
		t.Fatalf("unknown mutation %q", mode)
	}
}

func TestMonsterMeleeLiveSound532440CallbackMatrix(t *testing.T) {
	modes := []string{"retain", "swap_update", "swap_sound", "nil_sound", "class_player", "class_none", "monster_nil_update", "player_nil_update"}
	for _, owner := range []struct {
		name string
		c    bool
	}{{"Go-owned", false}, {"C-owned", true}} {
		t.Run(owner.name, func(t *testing.T) {
			for _, mode := range modes {
				for _, hit := range []int{0, 1, -1} {
					for _, oldDone := range []uint8{0, 1} {
						for _, newDone := range []uint8{0, 1} {
							for _, audioMutation := range []string{"none", "old_done_toggle", "new_done_toggle"} {
								name := fmt.Sprintf("%s/hit=%d/old=%d/new=%d/audio=%s", mode, hit, oldDone, newDone, audioMutation)
								t.Run(name, func(t *testing.T) {
									f := newMeleeSoundFixture532440(t, owner.c, oldDone, newDone)
									beforeUnit, beforeOld, beforeNew, beforeDef := *f.unit, *f.old, *f.fresh, *f.definition
									var events []meleeSoundEvent532440
									handled, fault := false, any(nil)
									func() {
										defer func() { fault = recover() }()
										handled = monsterActionMeleeUpdate532440(f.unit, monsterActionMeleeUpdateHooks532440{
											canStrike: func(fn unsafe.Pointer) bool {
												if fn != unsafe.Pointer(f.definition) {
													t.Fatal("wrong strike identity")
												}
												events = append(events, meleeSoundEvent532440{"can", 0})
												return true
											},
											strike: func(unit *Object, fn unsafe.Pointer) int {
												if unit != f.unit || fn != unsafe.Pointer(f.definition) {
													t.Fatal("wrong strike arguments")
												}
												events = append(events, meleeSoundEvent532440{"strike", hit})
												f.mutate(t, mode)
												return hit
											},
											audio: func(id uint32, unit *Object) {
												if unit != f.unit {
													t.Fatal("wrong audio identity")
												}
												events = append(events, meleeSoundEvent532440{"audio", int(id)})
												if audioMutation == "old_done_toggle" {
													f.old.Field120_3 = 1 - f.old.Field120_3
												}
												if audioMutation == "new_done_toggle" {
													f.fresh.Field120_3 = 1 - f.fresh.Field120_3
												}
											},
											pop: func() int { events = append(events, meleeSoundEvent532440{"pop", 0}); return 0 },
										})
									}()
									wantFault := mode == "monster_nil_update"
									if (fault != nil) != wantFault || handled == wantFault {
										t.Fatalf("handled/fault = %v/%v, want handled=%v fault=%v", handled, fault, !wantFault, wantFault)
									}
									want := []meleeSoundEvent532440{{"can", 0}, {"strike", hit}}
									hasAudio := mode == "retain" || mode == "swap_update" || mode == "swap_sound"
									if hasAudio {
										id := 1000
										if mode == "swap_update" || mode == "swap_sound" {
											id = 2000
										}
										if hit != 0 {
											id += 8
										} else {
											id += 9
										}
										want = append(want, meleeSoundEvent532440{"audio", id})
									}
									postOldDone := oldDone
									if hasAudio && audioMutation == "old_done_toggle" {
										postOldDone = 1 - oldDone
									}
									if !wantFault && postOldDone != 0 {
										want = append(want, meleeSoundEvent532440{"pop", 0})
									}
									if !reflect.DeepEqual(events, want) {
										t.Fatalf("events = %v, want original prefix %v", events, want)
									}
									wantUnit, wantOld, wantNew := beforeUnit, beforeOld, beforeNew
									switch mode {
									case "swap_update":
										wantUnit.UpdateData = unsafe.Pointer(f.fresh)
									case "swap_sound":
										wantOld.SoundSet122 = unsafe.Pointer(f.newSounds)
									case "nil_sound":
										wantOld.SoundSet122 = nil
									case "class_player":
										wantUnit.ObjClass = object.ClassPlayer
									case "class_none":
										wantUnit.ObjClass = 0
									case "monster_nil_update":
										wantUnit.UpdateData = nil
									case "player_nil_update":
										wantUnit.ObjClass, wantUnit.UpdateData = object.ClassPlayer, nil
									}
									wantOld.Field120_3 = postOldDone
									if hasAudio && audioMutation == "new_done_toggle" {
										wantNew.Field120_3 = 1 - newDone
									}
									if *f.unit != wantUnit || *f.old != wantOld || *f.fresh != wantNew || *f.definition != beforeDef {
										t.Fatal("writes outside callback mutations: unit/update/definition changed")
									}
								})
							}
						}
					}
				}
			}
		})
	}
}

func TestMonsterMeleeLiveSound532440WithoutAudioService(t *testing.T) {
	for _, cOwned := range []bool{false, true} {
		for _, mode := range []string{"retain", "swap_update", "swap_sound", "nil_sound", "class_player", "class_none", "monster_nil_update", "player_nil_update"} {
			t.Run(fmt.Sprintf("C-owned=%v/%s", cOwned, mode), func(t *testing.T) {
				f := newMeleeSoundFixture532440(t, cOwned, 1, 0)
				var events []string
				handled, fault := false, any(nil)
				func() {
					defer func() { fault = recover() }()
					handled = monsterActionMeleeUpdate532440(f.unit, monsterActionMeleeUpdateHooks532440{
						canStrike: func(unsafe.Pointer) bool { return true },
						strike:    func(*Object, unsafe.Pointer) int { events = append(events, "strike"); f.mutate(t, mode); return 1 },
						pop:       func() int { events = append(events, "pop"); return 0 },
					})
				}()
				wantFault := mode == "monster_nil_update"
				want := []string{"strike"}
				if !wantFault {
					want = append(want, "pop")
				}
				if (fault != nil) != wantFault || handled == wantFault || !reflect.DeepEqual(events, want) {
					t.Fatalf("handled/fault/events = %v/%v/%v, want fault=%v events=%v", handled, fault, events, wantFault, want)
				}
			})
		}
	}
}

func TestMonsterMeleeLiveSound532440PreservesAdmissionAndQuietGates(t *testing.T) {
	for _, cOwned := range []bool{false, true} {
		for _, mode := range []string{"wrong_frame", "nonzero_subcounter", "unsupported_callback"} {
			t.Run(fmt.Sprintf("C-owned=%v/%s", cOwned, mode), func(t *testing.T) {
				f := newMeleeSoundFixture532440(t, cOwned, 1, 0)
				if mode == "wrong_frame" {
					f.old.Field120_1 = 3
				}
				if mode == "nonzero_subcounter" {
					f.old.Field120_2 = 1
				}
				popped := false
				handled := monsterActionMeleeUpdate532440(f.unit, monsterActionMeleeUpdateHooks532440{
					canStrike: func(unsafe.Pointer) bool {
						f.unit.UpdateData = nil
						return mode != "unsupported_callback"
					},
					strike: func(*Object, unsafe.Pointer) int { t.Fatal("quiet/unsupported gate called strike"); return 0 },
					audio:  func(uint32, *Object) { t.Fatal("quiet/unsupported gate called audio") },
					pop:    func() int { popped = true; return 0 },
				})
				want := mode != "unsupported_callback"
				if handled != want || popped != want {
					t.Fatalf("handled/popped = %v/%v, want %v", handled, popped, want)
				}
			})
		}
	}
}
