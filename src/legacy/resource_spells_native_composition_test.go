package legacy

import (
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
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type resourceSpellLegacyServer struct {
	Server
	srv *server.Server
}

func (s *resourceSpellLegacyServer) S() *server.Server { return s.srv }

func resourceSpellCompositionOwner[T comparable](t *testing.T, value T) *T {
	t.Helper()
	type frame struct {
		Before [32]byte
		Value  T
		After  [32]byte
	}
	p, free := alloc.NewHigh(frame{})
	*p = frame{Value: value}
	t.Cleanup(free)
	for i := range p.Before {
		p.Before[i], p.After[i] = byte(0x51+i), byte(0xe7-i)
	}
	t.Cleanup(func() {
		for i := range p.Before {
			if p.Before[i] != byte(0x51+i) || p.After[i] != byte(0xe7-i) {
				t.Error("resource spell composition owner canary changed")
			}
		}
	})
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= 0xffffffff {
		t.Fatal("resource spell composition owner below4GiB")
	}
	return &p.Value
}

// Only the audio notification is observed. The public entry, shared callback,
// real balance/spell/AUD loaders, HP/mana setters, protection and audio queue run.
func TestResourceSpellsNativeDefaultComposition(t *testing.T) {
	const child = "OPENNOX_RESOURCE_SPELL_DEFAULT_CHILD"
	if os.Getenv(child) != "1" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary, "-test.run=^TestResourceSpellsNativeDefaultComposition$", "-test.v")
		cmd.Env = append(os.Environ(), child+"=1")
		out, err := cmd.CombinedOutput()
		t.Logf("native resource spell default composition:\n%s", out)
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	s := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(s.Close)
	old, oldData, flags := GetServer, datapath.Data(""), noxflags.GetGame()
	GetServer = func() Server { return &resourceSpellLegacyServer{srv: s} }
	noxflags.ResetGame()
	t.Cleanup(func() { GetServer = old; datapath.SetData(oldData); noxflags.ResetGame(); noxflags.SetGame(flags) })
	root := t.TempDir()
	if base := os.Getenv("OPENNOX_LOCAL_TEST_ARTIFACTS"); base != "" {
		var err error
		if err = os.MkdirAll(base, 0700); err != nil {
			t.Fatal(err)
		}
		root, err = os.MkdirTemp(base, "resource-spell-defaults-")
		if err != nil {
			t.Fatal(err)
		}
	}
	inputs := map[string]string{
		"gamedata.yml": "LesserHealAmount: 13.75\n",
		"spells.yml":   "- name: SPELL_LESSER_HEAL\n  mana_cost: 17\n  on_sound: LesserHealEffect\n",
	}
	for name, text := range inputs {
		if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	datapath.SetData(root)
	if err := s.Balance.Read(); err != nil {
		t.Fatal(err)
	}
	mem, _ := alloc.Make([]byte{}, 4)
	f := binfile.NewMemFile(unsafe.Pointer(&mem[0]), len(mem))
	if err := s.Spells.Read(f, nil, false); err != nil {
		t.Fatal(err)
	}
	f.Free()
	s.Audio.Init(s)
	t.Cleanup(s.Audio.Free)
	wire := binary.LittleEndian.AppendUint32(nil, 3)
	for _, name := range []string{"RestoreHealth", "RestoreMana", "LesserHealEffect"} {
		wire = append(wire, byte(len(name)))
		wire = append(wire, name...)
		wire = append(wire, make([]byte, 9)...)
		wire = append(wire, 1, 'x', 0)
	}
	audioBytes, _ := alloc.Make([]byte{}, len(wire))
	copy(audioBytes, wire)
	audioFile := binfile.NewMemFile(unsafe.Pointer(&audioBytes[0]), len(wire))
	if !s.Audio.ReadAUD(audioFile) {
		t.Fatal("real AUD descriptor loader")
	}
	audioFile.Free()
	for _, id := range []sound.ID{sound.SoundRestoreHealth, sound.SoundRestoreMana, sound.SoundLesserHealEffect} {
		if s.Audio.Field12(id) != 1 {
			t.Fatal("real AUD sample count", id)
		}
	}
	s.Players.Mult.Warrior.Health, s.Players.Mult.Wizard.Health, s.Players.Mult.Conjurer.Health = 2, 3, 0.5
	var heard []sound.ID
	var wantObject *server.Object
	s.Audio.OnSound(func(id sound.ID, kind int, obj *server.Object, pos types.Pointf) {
		if obj != wantObject || kind != 0 || pos != obj.Pos() {
			t.Fatal("actual audio owner/position", obj, wantObject)
		}
		heard = append(heard, id)
	})
	for route := 0; route < 3; route++ {
		for _, c := range []struct {
			name                                   string
			mode                                   int
			class                                  player.Class
			hp, mana, maximumMana                  uint16
			self, nilTarget                        bool
			wantHP, wantField2, wantMana, wantPrev uint16
			sound                                  sound.ID
		}{
			{"health", 0, player.Warrior, 10, 10, 100, false, false, 100, 100, 10, 99, sound.SoundRestoreHealth},
			{"mana-clamp", 1, player.Warrior, 10, 10, 100, false, false, 10, 9, 100, 10, sound.SoundRestoreMana},
			{"mana-word-wrap", 1, player.Warrior, 10, 1, 65535, false, false, 10, 9, 0, 1, sound.SoundRestoreMana},
			{"lesser-warrior", 2, player.Warrior, 10, 10, 100, false, false, 37, 9, 10, 99, sound.SoundLesserHealEffect},
			{"lesser-wizard", 2, player.Wizard, 10, 10, 100, false, false, 51, 9, 10, 99, sound.SoundLesserHealEffect},
			{"lesser-conjurer", 2, player.Conjurer, 10, 10, 100, false, false, 16, 9, 10, 99, sound.SoundLesserHealEffect},
			{"lesser-refund", 2, player.Warrior, 100, 10, 100, true, false, 100, 9, 27, 10, 0},
			{"nil-health", 0, player.Warrior, 10, 10, 100, false, true, 10, 9, 10, 99, 0},
			{"nil-mana", 1, player.Warrior, 10, 10, 100, false, true, 10, 9, 10, 99, 0},
			{"nil-lesser", 2, player.Warrior, 10, 10, 100, false, true, 10, 9, 10, 99, 0},
		} {
			t.Run(fmt.Sprintf("route%d/%s", route, c.name), func(t *testing.T) {
				pl := resourceSpellCompositionOwner(t, server.Player{})
				pl.Info().SetPlayerClass(c.class)
				update := resourceSpellCompositionOwner(t, server.PlayerUpdateData{Player: pl, ManaCur: c.mana, ManaPrev: 99, ManaMax: c.maximumMana})
				health := resourceSpellCompositionOwner(t, server.HealthData{Cur: c.hp, Field2: 9, Max: 100})
				obj := resourceSpellCompositionOwner(t, server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update), HealthData: health})
				arg := resourceSpellCompositionOwner(t, server.SpellAcceptArg{Obj: obj, Pos: types.Pointf{X: 17.5, Y: -23.25}})
				if c.nilTarget {
					arg.Obj = nil
				}
				var second *server.Object
				if c.self {
					second = obj
				}
				heard = nil
				wantObject = obj
				s.Audio.Reset()
				beforeArg := *arg
				result := resourceSpellActualC(route, c.mode, int32(spell.SPELL_LESSER_HEAL), second, obj, nil, arg, -1)
				wantResult := int32(1)
				if c.nilTarget {
					wantResult = 0
				}
				if result != wantResult || health.Cur != c.wantHP || health.Field2 != c.wantField2 || update.ManaCur != c.wantMana || update.ManaPrev != c.wantPrev || update.ManaMax != c.maximumMana || *arg != beforeArg {
					t.Fatal("actual resource helper composition", result, *health, *update, *arg)
				}
				var want []sound.ID
				if c.sound != 0 {
					want = []sound.ID{c.sound}
				}
				if !reflect.DeepEqual(heard, want) {
					t.Fatal("actual audio event notifications", heard, want)
				}
				var queued []sound.ID
				s.Audio.EachEvent(func(ev *server.AudioEvent) {
					if ev.Obj != obj || ev.Kind != 0 || ev.Code != 0 {
						t.Fatal("actual queued audio identity", ev)
					}
					queued = append(queued, ev.Sound)
				})
				if !reflect.DeepEqual(queued, want) {
					t.Fatal("actual audio queue", queued, want)
				}
			})
		}
	}
}
