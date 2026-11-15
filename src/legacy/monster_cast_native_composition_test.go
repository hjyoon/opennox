package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/monster_cast_default_services_541300_5413b0_541490_v1.json.gz
var monsterCastDefaultOriginal []byte

// Spell acceptance is the explicit observation boundary. The public entries,
// exact default runtime, bot record morphs, direction, RNG, buff/voice/audio,
// actual spell-power and 4FDD20 dispatch all run without a runtime override.
func TestMonsterCastNativeDefaultComposition(t *testing.T) {
	const child = "OPENNOX_MONSTER_CAST_DEFAULT_CHILD"
	if os.Getenv(child) != "1" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary, "-test.run=^TestMonsterCastNativeDefaultComposition$", "-test.v", "-test.timeout=3m")
		cmd.Env = append(os.Environ(), child+"=1")
		out, err := cmd.CombinedOutput()
		t.Logf("actual monster cast default composition:\n%s", out)
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	InitBlobData()
	s := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(s.Close)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if err := s.Types.ReadObjectType(&things.Thing{Name: "ImaginaryCaster"}); err != nil {
		t.Fatal(err)
	}
	s.Audio.Init(s)
	t.Cleanup(s.Audio.Free)
	old, oldData, flags, oldCast := GetServer, datapath.Data(""), noxflags.GetGame(), Nox_xxx_castSpellByUser_4FDD20
	GetServer = func() Server { return &resourceSpellLegacyServer{srv: s} }
	noxflags.ResetGame()
	t.Cleanup(func() {
		GetServer = old
		datapath.SetData(oldData)
		noxflags.ResetGame()
		noxflags.SetGame(flags)
		Nox_xxx_castSpellByUser_4FDD20 = oldCast
	})
	root := t.TempDir()
	if base := os.Getenv("OPENNOX_LOCAL_TEST_ARTIFACTS"); base != "" {
		if err := os.MkdirAll(base, 0700); err != nil {
			t.Fatal(err)
		}
		var err error
		root, err = os.MkdirTemp(base, "monster-cast-defaults-")
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "spells.yml"), []byte("- name: "+spell.ID(41).String()+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	datapath.SetData(root)
	mem, _ := alloc.Make([]byte{}, 4)
	f := binfile.NewMemFile(unsafe.Pointer(&mem[0]), len(mem))
	if err := s.Spells.Read(f, nil, false); err != nil {
		t.Fatal(err)
	}
	f.Free()
	wire := binary.LittleEndian.AppendUint32(nil, 1)
	name := "FireballCast"
	wire = append(wire, byte(len(name)))
	wire = append(wire, name...)
	wire = append(wire, make([]byte, 9)...)
	wire = append(wire, 1, 'x', 0)
	buf, _ := alloc.Make([]byte{}, len(wire))
	copy(buf, wire)
	aud := binfile.NewMemFile(unsafe.Pointer(&buf[0]), len(buf))
	if !s.Audio.ReadAUD(aud) {
		t.Fatal("actual AUD loader")
	}
	aud.Free()
	if sound.SoundFireballCast != 41 || s.Audio.Field12(sound.SoundFireballCast) != 1 {
		t.Fatal("original sound41 and actual audio descriptor")
	}
	var rows []monsterCastCase
	npcVoiceOriginalDecode(t, monsterCastDefaultOriginal, "f06e25d439363f6664a3603d053490a4cef23cc0f15386c6d11104b2ca670346", "16086451eafe8e95ac32066a94e197f22355a89dc5c7c3498c2d76e2cafa7b03", 44, &rows)
	var heard []sound.ID
	var owner *server.Object
	s.Audio.OnSound(func(id sound.ID, kind int, obj *server.Object, pos types.Pointf) {
		if obj != owner || kind != 0 || pos != owner.Pos() {
			t.Fatal("actual audio owner/position")
		}
		heard = append(heard, id)
	})
	compositions := 0
	for route := 0; route < 3; route++ {
		for _, c := range rows {
			t.Run(fmt.Sprintf("route%d/%s", route, c.Input.Name), func(t *testing.T) {
				v := monsterCastArenaNew(t, 8)
				v.prepare(c)
				// Default services see valid typed records. The original matrix
				// separately checks arbitrary non-field bytes and ordered mutations.
				for _, o := range v.objects {
					position, velocity, dir1, dir2, class, sub, data := o.PosVec, o.VelVec, o.Direction1, o.Direction2, o.ObjClass, o.ObjSubClass, o.UpdateData
					*o = server.Object{ObjClass: class, ObjSubClass: sub, PosVec: position, VelVec: velocity, Direction1: dir1, Direction2: dir2, UpdateData: data, TypeInd: uint16(s.Types.IndByID("ImaginaryCaster") + 1)}
				}
				for i, u := range v.updates {
					flags, accuracy, frame, ready, index, stack := u.StatusFlags, u.Field330, u.Field120_1, u.Field120_2, u.AIStackInd, u.AIStack
					*u = server.MonsterUpdateData{StatusFlags: flags, Field330: accuracy, Field120_1: frame, Field120_2: ready, AIStackInd: index, AIStack: stack, MonsterDef: v.definitions[i], SoundSet122: v.voice, Field545: v.players[i], Field510: 3}
					if !c.Input.VoiceAvailable {
						u.SoundSet122 = nil
					}
					pl := resourceSpellCompositionOwner(t, server.Player{})
					pl.SpellLvl[41] = 4
					*v.players[i] = server.PlayerUpdateData{Field73: u, Player: pl}
				}
				if c.Input.BuffReturn != 0 {
					v.objects[0].Buffs = 1 << server.ENCHANT_ANTI_MAGIC
				}
				s.Rand.Logic = prand.New(int(c.Input.RNGIndex))
				v.randomIndex = func() uint32 { return uint32(s.Rand.Logic.Index()) }
				if !reflect.DeepEqual(v.state(), c.Before) {
					t.Fatalf("default input state got%+v want%+v", v.state(), c.Before)
				}
				before := bytes.Clone(v.raw)
				heard = nil
				owner = v.objects[0]
				s.Audio.Reset()
				accepted := 0
				var castEvent *monsterCastEvent
				wantAudio := 0
				for i := range c.Events {
					if c.Events[i].Kind == "cast" {
						castEvent = &c.Events[i]
					}
					if c.Events[i].Kind == "audio" {
						wantAudio++
					}
				}
				Nox_xxx_castSpellByUser_4FDD20 = func(id int32, caster *server.Object, arg *server.SpellAcceptArg) int32 {
					return s.CastSpellByUser4FDD20(spell.ID(id), caster, arg, server.CastSpellByUserRuntime4FDD20{
						SpellGetPower:    func(id spell.ID, obj *server.Object) int32 { return s.SpellPower4FE7B0(id, obj) },
						DisableEnchant:   func(*server.Object, server.EnchantID) { t.Fatal("unexpected offensive removal") },
						CancelDuration:   func(spell.ID, *server.Object) { t.Fatal("unexpected duration removal") },
						CreateProjectile: func(*server.Object, *server.Object, spell.ID) { t.Fatal("unexpected targeted projectile") },
						SpellAccept: func(id spell.ID, second, third, fourth *server.Object, arg *server.SpellAcceptArg, power int32) int32 {
							v.activeArg = arg
							accepted++
							if castEvent == nil || int32(id) != int32(castEvent.Args[0].(float64)) || second != owner || third != owner || fourth != owner || !reflect.DeepEqual(v.argumentState(arg), castEvent.Before.Argument) {
								t.Fatal("actual4FDD20 native dispatch differs from original cast event", id, second, third, fourth, v.argumentState(arg), castEvent)
							}
							wantPower := int32(3)
							if owner.ObjClass.Has(object.ClassPlayer) {
								wantPower = 4
							}
							if power != wantPower {
								t.Fatal("actual live monster/player spell power", power, wantPower)
							}
							return int32(c.Input.CastReturn)
						},
					})
				}
				result := monsterCastActualC(route, c.Input.Root, c.Input.Spell, owner, v.object(c.Input.Target), v.arg, c.Input.Mode, v.out)
				if c.Return != nil && result != *c.Return {
					t.Fatal("default direct return", result, *c.Return)
				}
				wantCasts := 0
				if castEvent != nil {
					wantCasts = 1
				}
				if accepted != wantCasts {
					t.Fatal("actual acceptance count", accepted, wantCasts)
				}
				if !reflect.DeepEqual(v.state(), c.After) {
					t.Fatalf("default final state got%+v want%+v", v.state(), c.After)
				}
				if !bytes.Equal(v.raw, v.expected(before, c)) {
					t.Fatal("default whole native owner canaries")
				}
				if len(heard) != wantAudio {
					t.Fatal("actual audio notifications", heard, wantAudio)
				}
				queued := 0
				s.Audio.EachEvent(func(ev *server.AudioEvent) {
					queued++
					if ev.Sound != sound.SoundFireballCast || ev.Obj != owner || ev.Kind != 0 || ev.Code != 0 {
						t.Fatal("actual queued native voice audio", ev)
					}
				})
				if queued != wantAudio {
					t.Fatal("actual audio queue count", queued, wantAudio)
				}
				compositions++
			})
		}
	}
	if compositions != 132 {
		t.Fatal("real default composition count", compositions)
	}
}
