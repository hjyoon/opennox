package legacy

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
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

type fireballCompositionServer struct {
	Server
	native            *server.Server
	projectile, owner *server.Object
	position          types.Pointf
	placements        int
}

func (s *fireballCompositionServer) S() *server.Server { return s.native }
func (s *fireballCompositionServer) CreateObjectAt(obj, owner server.Obj, pos types.Pointf) {
	s.projectile, s.owner, s.position = server.ToObject(obj), server.ToObject(owner), pos
	s.placements++
	if s.projectile == nil || math.Float32bits(s.projectile.SpeedCur) != math.Float32bits(2.25) {
		panic("real type constructor did not supply definition speed72/32")
	}
}

// World placement is the declared observation boundary. Actual native type
// loading/allocation, packed type selection, map trace, balance, sound lookup
// and audio collection use the production runtime without a dependency override.
func TestFireballNativeDefaultComposition(t *testing.T) {
	const child = "OPENNOX_FIREBALL_DEFAULT_CHILD"
	if os.Getenv(child) != "1" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary, "-test.run=^TestFireballNativeDefaultComposition$", "-test.v", "-test.timeout=3m")
		cmd.Env = append(os.Environ(), child+"=1")
		out, err := cmd.CombinedOutput()
		t.Logf("actual Fireball native default composition:\n%s", out)
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
	speed := 72
	for _, name := range []string{"Fireball", "StrongFireball", "TitanFireball"} {
		if err := s.Types.ReadObjectType(&things.Thing{Name: name, Speed: &speed}); err != nil {
			t.Fatal(err)
		}
	}
	if !s.Objs.InitHigh(128) {
		t.Fatal("actual high native object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	s.Walls.InitHigh()
	t.Cleanup(s.Walls.Free)
	s.Audio.Init(s)
	t.Cleanup(s.Audio.Free)
	outer := &fireballCompositionServer{native: s}
	old, oldData, flags := GetServer, datapath.Data(""), noxflags.GetGame()
	GetServer = func() Server { return outer }
	noxflags.ResetGame()
	t.Cleanup(func() { GetServer = old; datapath.SetData(oldData); noxflags.ResetGame(); noxflags.SetGame(flags) })
	root := t.TempDir()
	if base := os.Getenv("OPENNOX_LOCAL_TEST_ARTIFACTS"); base != "" {
		if err := os.MkdirAll(base, 0700); err != nil {
			t.Fatal(err)
		}
		var err error
		root, err = os.MkdirTemp(base, "fireball-defaults-")
		if err != nil {
			t.Fatal(err)
		}
	}
	spellID := spell.ID(20)
	inputs := map[string]string{
		"gamedata.yml": "FireballSpeedCoeff: [1.1, 1.1, 1.1, 1.1, 1.1]\n",
		"spells.yml":   "- name: " + spellID.String() + "\n  cast_sound: FireballCast\n",
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
	wire := binary.LittleEndian.AppendUint32(nil, 1)
	name := "FireballCast"
	wire = append(wire, byte(len(name)))
	wire = append(wire, name...)
	wire = append(wire, make([]byte, 9)...)
	wire = append(wire, 1, 'f', 0)
	buf, _ := alloc.Make([]byte{}, len(wire))
	copy(buf, wire)
	aud := binfile.NewMemFile(unsafe.Pointer(&buf[0]), len(buf))
	if !s.Audio.ReadAUD(aud) {
		t.Fatal("actual AUD loader")
	}
	aud.Free()
	if s.Audio.Field12(sound.SoundFireballCast) != 1 {
		t.Fatal("actual audio descriptor")
	}
	firing := s.Objs.NewObject(&server.ObjectType{})
	if firing == nil {
		t.Fatal("actual firing object allocation")
	}
	arg := resourceSpellCompositionOwner(t, server.SpellAcceptArg{Pos: types.Pointf{X: -33, Y: 77}})
	var heard []sound.ID
	s.Audio.OnSound(func(id sound.ID, kind int, obj *server.Object, pos types.Pointf) {
		if obj != firing || pos != firing.Pos() || kind != 0 {
			t.Fatal("actual audio native owner/position", obj, pos)
		}
		heard = append(heard, id)
	})
	rows := fireballCases(t)
	compositions := 0
	for route := 0; route < 3; route++ {
		for _, c := range rows {
			if !strings.HasPrefix(c.Input.Name, "default-typed-providers-") || c.Input.Trace != 1 {
				continue
			}
			t.Run(fmt.Sprintf("route%d/%s", route, c.Input.Name), func(t *testing.T) {
				input := c.Before[0]
				firing.PosVec = types.Pointf{X: math.Float32frombits(input.Position[0]), Y: math.Float32frombits(input.Position[1])}
				firing.VelVec = types.Pointf{X: math.Float32frombits(input.Velocity[0]), Y: math.Float32frombits(input.Velocity[1])}
				firing.Shape.Circle.R = math.Float32frombits(input.Radius)
				firing.SpeedCur = math.Float32frombits(input.Speed)
				firing.Direction1 = server.Dir16(input.Directions[0])
				firing.Direction2 = server.Dir16(input.Directions[1])
				heard = nil
				s.Audio.Reset()
				outer.placements = 0
				outer.projectile = nil
				beforeArg := *arg
				if got := fireballActualC(route, c.Input.Spell, nil, firing, firing, arg, c.Input.Level); got != 1 || outer.placements != 1 || outer.owner != firing || outer.projectile == nil || *arg != beforeArg {
					t.Fatal("actual production dispatch/placement/native owner", got, outer.placements, outer.owner, outer.projectile)
				}
				projectile := outer.projectile
				var creation *fireballOriginalEvent
				for i := range c.Events {
					if c.Events[i].Kind == "create" {
						creation = &c.Events[i]
						break
					}
				}
				if creation == nil || math.Float32bits(outer.position.X) != uint32(creation.Args[2].(float64)) || math.Float32bits(outer.position.Y) != uint32(creation.Args[3].(float64)) {
					t.Fatal("actual placement position differs from unchanged original", outer.position, creation)
				}
				expectedName := c.Events[0].Args[0].(string)
				if int(projectile.TypeInd) != s.Types.IndByID(expectedName) {
					t.Fatal("real packed type lookup/constructor selected wrong type", projectile.TypeInd, expectedName)
				}
				actual := (&fireballArena{objects: [3]*server.Object{firing, projectile, firing}}).state()
				want := c.After[1]
				if actual[0] != c.Before[0] || actual[1].Speed != want.Speed || actual[1].Velocity != want.Velocity || actual[1].Directions != want.Directions {
					t.Fatal("actual production projectile fields differ from original", actual[0], actual[1], want)
				}
				if len(heard) != 1 || heard[0] != sound.SoundFireballCast {
					t.Fatal("actual sound definition/audio notification", heard)
				}
				queued := 0
				s.Audio.EachEvent(func(ev *server.AudioEvent) {
					queued++
					if ev.Sound != sound.SoundFireballCast || ev.Obj != firing || ev.Kind != 0 || ev.Code != 0 {
						t.Fatal("actual queued native audio", ev)
					}
				})
				if queued != 1 {
					t.Fatal("actual audio event count", queued)
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(projectile)) <= math.MaxUint32 || uintptr(unsafe.Pointer(firing)) <= math.MaxUint32) {
					t.Fatal("actual constructor must retain high native owners")
				}
				compositions++
			})
		}
	}
	if compositions != 45 {
		t.Fatal("actual default composition case count", compositions)
	}
}
