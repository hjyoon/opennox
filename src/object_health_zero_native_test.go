package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	ns4 "github.com/opennox/noxscript/ns/v4"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// The real PotionDie and ImpEggDie registrations, HP setter, audio allocator,
// and delayed-deletion service are used. No death/HP/audio result is supplied.
// Player records use a declared inert protection token (0), not a protection
// callback replacement. Their death callback is the explicit fixture handler,
// not the stock PlayerDie lifecycle.
func healthZeroDefinitions(t *testing.T, s *Server) {
	t.Helper()
	wire := binary.LittleEndian.AppendUint32(nil, 2)
	for _, name := range []string{"PotionBreak", "EggBreak"} {
		wire = append(wire, byte(len(name)))
		wire = append(wire, name...)
		wire = binary.LittleEndian.AppendUint16(wire, 8)
		wire = append(wire, 0)
		wire = binary.LittleEndian.AppendUint16(wire, 40)
		wire = append(wire, 0, 0, 0, 0, 1, 's', 0)
	}
	input := luaCompatAudioFile{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(input) || input.Len() != 0 {
		t.Fatal("real AUD input was not completely consumed")
	}
	for _, id := range []sound.ID{sound.SoundPotionBreak, sound.SoundEggBreak} {
		if s.Audio.Field12(id) != 1 {
			t.Fatal("real death sound descriptor is disabled")
		}
	}
	s.Audio.Reset()
}

func TestObjectSetHealthZeroNativeDeathEffects(t *testing.T) {
	for _, class := range []object.Class{object.ClassSimple, object.ClassPlayer} {
		for _, handler := range []string{"none", "PotionDie", "ImpEggDie"} {
			for _, state := range []string{"alive", "already-zero", "zero-maximum", "no-health"} {
				for _, flags := range []object.Flags{object.FlagEnabled, object.FlagEnabled | object.FlagDead, object.FlagEnabled | object.FlagDestroyed} {
					for _, mode := range []string{"normal", "god", "sudden-death"} {
						for _, requested := range []int{-1, 0, 1, 65535, math.MaxInt} {
							t.Run(fmt.Sprintf("class-%x/%s/%s/flags-%x/%s/value-%d", class, handler, state, flags, mode, requested), func(t *testing.T) {
								s := scriptNSRoamFlagServer515C80(t)
								noxflags.ResetGame()
								noxflags.SetGame(noxflags.GameHost)
								oldEngine := noxflags.GetEngine()
								noxflags.UnsetEngine(noxflags.EngineGodMode)
								t.Cleanup(func() { noxflags.SetEngine(oldEngine); noxflags.UnsetEngine(^oldEngine) })
								if mode == "god" {
									noxflags.SetEngine(noxflags.EngineGodMode)
								}
								if mode == "sudden-death" {
									noxflags.SetGame(noxflags.GameSuddenDeath)
								}
								s.SetFrame(37)
								s.Audio.Init(s.Server)
								t.Cleanup(s.Audio.Free)
								healthZeroDefinitions(t, s)
								if !s.Objs.Init(2) {
									t.Fatal("cannot allocate real native objects")
								}
								t.Cleanup(s.Objs.FreeObjects)
								unit := s.Objs.NewObject(&server.ObjectType{})
								if unit == nil {
									t.Fatal("cannot allocate a server-associated object")
								}
								health, freeHealth := alloc.New(server.HealthData{})
								update, freeUpdate := alloc.New(server.PlayerUpdateData{})
								player, freePlayer := alloc.New(server.Player{})
								sentinel, freeSentinel := alloc.New(server.Object{})
								t.Cleanup(freeSentinel)
								t.Cleanup(freePlayer)
								t.Cleanup(freeUpdate)
								t.Cleanup(freeHealth)
								*health = server.HealthData{Cur: 17, Max: 31, Field2: 0x4321, Field16: 0x87654321}
								*player = server.Player{PlayerInd: 7, PlayerUnit: unit}
								*update = server.PlayerUpdateData{Player: player}
								*sentinel = server.Object{ObjFlags: object.FlagDestroyed, DeletedAt: 11, Extent: 0xe9876543}
								unit.ObjClass, unit.ObjFlags = class, flags
								unit.TypeInd, unit.Extent = 420, 0xe1234567
								unit.HealthData, unit.UpdateData = health, unsafe.Pointer(update)
								unit.PosVec = types.Ptf(31.25, -17.5)
								unit.Field38, unit.DeletedAt = 0x12345678, 0xabcdef01
								if handler != "none" {
									fn, size, ok := server.ObjectDeathHandler(handler)
									if !ok || fn == nil || size != 0 {
										t.Fatal("fixture's real registered death handler is missing")
									}
									unit.Death = fn
								}
								switch state {
								case "already-zero":
									health.Cur = 0
								case "zero-maximum":
									health.Max = 0
								case "no-health":
									unit.HealthData = nil
								}
								s.Objs.List, s.Objs.DeletedList = unit, sentinel
								t.Cleanup(func() { s.Objs.List, s.Objs.DeletedList = nil, nil })
								for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(health), unsafe.Pointer(update), unsafe.Pointer(player), unsafe.Pointer(sentinel)} {
									if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
										t.Fatalf("health/death fixture is not native-width: %p", p)
									}
								}
								initialUnit := *unit
								initialHealth, initialPlayer, initialUpdate, initialSentinel := *health, *player, *update, *sentinel
								reset := func() {
									*unit, *health, *player, *update, *sentinel = initialUnit, initialHealth, initialPlayer, initialUpdate, initialSentinel
									s.Objs.DeletedList = sentinel
									s.Audio.Reset()
								}
								wantHP := int(initialHealth.Cur)
								wanted := requested
								if wanted < 0 {
									wanted = 0
								}
								maximum := 0
								if initialUnit.HealthData != nil {
									maximum = int(initialHealth.Max)
								}
								if wanted > maximum {
									wanted = maximum
								}
								if mode != "sudden-death" && initialUnit.HealthData != nil {
									wantHP = wanted
								}
								death := mode != "sudden-death" && wanted == 0 && initialUnit.HealthData != nil &&
									initialHealth.Max != 0 && !initialUnit.ObjFlags.Has(object.FlagDead)
								check := func(route string) {
									t.Helper()
									wantUnit := initialUnit
									wantHealth := initialHealth
									if mode != "sudden-death" && initialUnit.HealthData != nil {
										wantUnit.Field38 = math.MaxUint32
										wantHealth.Cur = uint16(wantHP)
									}
									wantHead := sentinel
									if death {
										wantUnit.ObjFlags |= object.FlagDead
										switch handler {
										case "none", "PotionDie":
											if !initialUnit.ObjFlags.Has(object.FlagDestroyed) {
												wantUnit.ObjFlags |= object.FlagDestroyed
												wantUnit.DeletedAt, wantUnit.DeletedNext = 37, sentinel
												wantHead = unit
											}
										case "ImpEggDie":
											wantUnit.ObjFlags |= object.FlagNoCollide
										}
									}
									if !bytes.Equal(scriptNSRoamFlagBytes515C80(unsafe.Pointer(unit), unsafe.Sizeof(*unit)),
										scriptNSRoamFlagBytes515C80(unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit))) ||
										*health != wantHealth || *player != initialPlayer || *update != initialUpdate || *sentinel != initialSentinel ||
										s.Objs.DeletedList != wantHead || s.Objs.List != unit || s.Frame() != 37 {
										t.Errorf("%s HP/death/native record differs: HP=%d want=%d flags=%x want=%x head=%p want=%p", route, health.Cur, wantHP, unit.ObjFlags, wantUnit.ObjFlags, s.Objs.DeletedList, wantHead)
									}
									got, _ := luaCompatAudioEvents(t, s)
									var expected []luaCompatAudioSnapshot
									if death && handler != "none" {
										id := sound.SoundPotionBreak
										if handler == "ImpEggDie" {
											id = sound.SoundEggBreak
										}
										obj := unit
										if initialUnit.ObjFlags.Has(object.FlagDestroyed) {
											obj = nil
										}
										expected = append(expected, luaCompatAudioSnapshot{id, math.Float32bits(unit.PosVec.X), math.Float32bits(unit.PosVec.Y), obj, 0, 0})
									}
									if !fmtAudioSnapshotsEqual(got, expected) {
										t.Errorf("%s actual death sounds=%+v, want %+v", route, got, expected)
									}
								}
								var api ns4.Obj = s.noxScriptP().toObj(unit)
								for _, route := range []string{"root-object", "public-sdk", "legacy-object-interface"} {
									reset()
									switch route {
									case "root-object":
										asObjectS(unit).SetHealth(requested)
									case "public-sdk":
										api.SetHealth(requested)
									case "legacy-object-interface":
										scrObject{asObjectS(unit)}.SetHealth(requested)
									}
									check(route)
									// Repeated lethal setting must not replay a death callback.
									if wanted == 0 {
										switch route {
										case "root-object":
											asObjectS(unit).SetHealth(requested)
										case "public-sdk":
											api.SetHealth(requested)
										case "legacy-object-interface":
											scrObject{asObjectS(unit)}.SetHealth(requested)
										}
										check(route + " repeated")
									}
								}
								// Independent native lethal-damage control for the ordinary,
								// positive-health cases; God player damage is intentionally
								// not used as an oracle for the direct setter's assignment.
								if mode == "normal" && wanted == 0 && initialUnit.HealthData != nil && initialHealth.Cur != 0 && initialHealth.Max != 0 {
									reset()
									legacy.Nox_xxx_unitDamageClear_4EE5E0(unit, int(initialHealth.Cur))
									check("actual native lethal control")
								}
							})
						}
					}
				}
			}
		}
	}
}
