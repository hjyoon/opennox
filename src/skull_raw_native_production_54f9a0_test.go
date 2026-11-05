package opennox

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func TestSkullRawActualCProduction54F9A0(t *testing.T) {
	mode := os.Getenv("OPENNOX_SKULL_RAW_PRODUCTION_CHILD")
	if mode == "" {
		for _, mode := range []string{"disabled", "waiting", "scan-empty", "wrapper-empty", "candidate-class", "candidate-flags", "candidate-hit", "candidate-behind", "projectile-missing", "projectile-cold", "projectile-warm", "update-fire"} {
			t.Run(mode, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSkullRawActualCProduction54F9A0$", "-test.v")
				cmd.Env = append(os.Environ(), "OPENNOX_SKULL_RAW_PRODUCTION_CHILD="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("actual default Skull C %s: %v\n%s", mode, err, out)
				if ctx.Err() != nil || err != nil || !bytes.Contains(out, []byte("SKULL_RAW_ACTUAL_DEFAULT_C_PRODUCTION")) {
					t.Fatal("actual default root-server services", err)
				}
			})
		}
		return
	}
	legacy.InitBlobData()
	handles.Init()
	t.Cleanup(handles.Release)
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	oldFlags, oldPlay := noxflags.GetGame(), noxflags.GetGamePlay()
	noxflags.ResetGame()
	noxflags.SetGamePlay(noxflags.GameplayFlag1)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
		noxflags.UnsetGamePlay(noxflags.GetGamePlay())
		noxflags.SetGamePlay(oldPlay)
	})
	s.SetFrame(1400)
	s.SetTickRate(30)
	s.spells.duration.Init(s)
	s.FreeObjectTypes()
	t.Cleanup(s.FreeObjectTypes)
	if !s.Objs.Init(12) {
		t.Fatal("actual object pool")
	}
	t.Cleanup(s.Objs.FreeObjects)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.Init() == 0 {
		t.Fatal("actual wall grid")
	}
	t.Cleanup(s.Walls.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	name := sound.ID(889).String()
	wire := binary.LittleEndian.AppendUint32(nil, 1)
	wire = append(wire, byte(len(name)))
	wire = append(wire, name...)
	wire = append(wire, 8, 0, 0, 40, 0, 0, 0, 0, 0, 1, 's', 0)
	file := luaCompatAudioFile{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(file) || file.Len() != 0 {
		t.Fatal("real ArrowTrapShoot descriptor")
	}
	s.Audio.Reset()
	dir, oldData := t.TempDir(), datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	if err := os.WriteFile(filepath.Join(dir, "gamedata.yml"), []byte("ArrowTrapDamage: 45.66777420043945\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Balance.Read(); err != nil || math.Float32bits(float32(s.Balance.Float("ArrowTrapDamage"))) != 0x4236abcd {
		t.Fatal("actual original binary32 balance", err)
	}
	for _, def := range []*things.Thing{
		{Name: "ArrowTrap1", OnUpdate: &things.ProcFunc{Name: "SkullUpdate"}},
		{Name: "ArrowTrap2", OnUpdate: &things.ProcFunc{Name: "SkullUpdate"}},
		{Name: "MercArcherArrow", OnCollide: &things.ProcFunc{Name: "ProjectileCollide"}},
		{Name: "RawSkullTarget", Class: []things.Class{"MONSTER"}, OnUpdate: &things.ProcFunc{Name: "MonsterUpdate"}},
	} {
		if err := s.Types.ReadObjectType(def); err != nil {
			t.Fatal(err)
		}
	}
	s.Types.ByID("MercArcherArrow").Speed = 1.25
	source := s.NewObjectByTypeID("ArrowTrap1")
	target := s.NewObjectByTypeID("ArrowTrap2")
	if source == nil || target == nil || source.UpdateData == nil || target.UpdateData == nil {
		t.Fatal("actual native constructors and52-byte update")
	}
	source.ObjFlags = object.FlagActive | object.FlagEnabled
	source.PosVec = types.Ptf(300, 300)
	source.Direction1, source.Direction2 = 0, 77
	source.Shape.Kind, source.Shape.Circle.R = server.ShapeKindCircle, 14
	target.PosVec = types.Ptf(320, 300)
	id := uint32(s.Types.IndByID("MercArcherArrow"))
	if id == 0 {
		t.Fatal("actual numeric type lookup")
	}
	d := (*server.SkullUpdateData)(source.UpdateData)
	*d = server.SkullUpdateData{ScanDelay: 17, FireDelay: 19, TargetReady: 1, ProjectileType: id, Enabled: 1}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(source)) <= math.MaxUint32 || uintptr(source.UpdateData) <= math.MaxUint32 || uintptr(unsafe.Pointer(target)) <= math.MaxUint32) {
		t.Fatal("actual native pointers must exceed4GiB")
	}
	state := unsafe.Slice(memmap.PtrUint32(0x5D4594, 2491764), 6)
	clear(state)
	state[0] = 99
	entry := 0
	switch mode {
	case "disabled":
		source.ObjFlags &^= object.FlagEnabled
	case "waiting":
	case "scan-empty":
		entry = 3
	case "wrapper-empty":
		entry = 2
	case "candidate-class":
		entry = 4
		target.ObjClass = 0
	case "candidate-flags":
		entry = 4
		target.ObjClass = object.ClassMonster
		target.ObjFlags |= object.FlagDead
	case "candidate-hit", "candidate-behind":
		entry = 4
		target = s.NewObjectByTypeID("RawSkullTarget")
		if target == nil || target.UpdateData == nil || !target.ObjClass.Has(object.ClassMonster) {
			t.Fatal("real native monster constructor/update")
		}
		target.PosVec = types.Ptf(320, 300)
		if mode == "candidate-behind" {
			target.PosVec = types.Ptf(280, 300)
		}
		if !s.IsEnemyTo(source, target) || !s.MapTraceVision(target, source) {
			t.Fatal("independent actual hostility and map controls")
		}
	case "projectile-missing":
		entry = 1
		id = 0
	case "projectile-cold":
		entry = 1
	case "projectile-warm":
		entry = 1
		state[1], state[2], state[3] = id, uint32(source.TypeInd), uint32(target.TypeInd)
	case "update-fire":
		d.FireDelay = 0
	default:
		t.Fatal("undeclared actual C condition", mode)
	}
	beforeData := *d
	beforeSource := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(source)), int(unsafe.Sizeof(*source))))
	sounds := 0
	s.Audio.OnSound(func(id sound.ID, kind int, unit *server.Object, position types.Pointf) {
		if id != 889 || kind != 0 || unit != source || position != source.PosVec {
			t.Fatal("real C audio identity", id, kind, unit)
		}
		sounds++
	})
	value, ok := legacy.SkullRawActualDefaultC54F9A0(entry, source, target, id)
	if !ok {
		t.Fatal("actual typed C/default services required")
	}
	created := mode == "projectile-cold" || mode == "projectile-warm" || mode == "update-fire"
	if created {
		p := s.Objs.Pending
		if p == nil || p.TypeInd != uint16(id) || p.ObjOwner != source || p.CollideData == nil || p.PosVec != types.Ptf(318, 300) || p.Direction1 != 0 || p.Direction2 != 0 || p.VelVec.X != p.SpeedCur || p.VelVec.Y != 0 {
			t.Fatal("actual C-created projectile, owner, position and velocity", p)
		}
		c := (*server.ProjectileCollideData)(p.CollideData)
		if c.Damage != 45 || c.Field4 != 45 || sounds != 1 || state[1] != id || state[2] != uint32(source.TypeInd) || state[3] != uint32(target.TypeInd) {
			t.Fatal("real C balance spill/conversion, collide stores, audio and caches", *c, sounds, state)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(p)) <= math.MaxUint32 || uintptr(p.CollideData) <= math.MaxUint32) {
			t.Fatal("actual projectile/collide pointers below4GiB")
		}
		if mode == "update-fire" {
			beforeData.ScanDelay, beforeData.FireDelay = 16, 29
			if value != 29 || state[4] != uint32(source.TypeInd) || state[5] != uint32(target.TypeInd) {
				t.Fatal("actual raw update timers/return/main caches", value, state)
			}
		}
	} else {
		if s.Objs.Pending != nil || sounds != 0 {
			t.Fatal("unexpected construction/audio")
		}
		if !bytes.Equal(beforeSource, unsafe.Slice((*byte)(unsafe.Pointer(source)), int(unsafe.Sizeof(*source)))) {
			t.Fatal("whole native source guard outside reached world services")
		}
		switch mode {
		case "disabled":
			beforeData.Enabled = 0
			if value != uint32(source.ObjFlags) {
				t.Fatal("disabled original flags return")
			}
		case "waiting":
			beforeData.ScanDelay, beforeData.FireDelay = 16, 18
			if value != 18 {
				t.Fatal("waiting cooldown return")
			}
		case "scan-empty":
			if value != 0 || state[0] != 0 {
				t.Fatal("real empty map scan and live marker")
			}
		case "wrapper-empty":
			beforeData.TargetReady = 0
			if state[0] != 0 {
				t.Fatal("real wrapper empty scan")
			}
		case "candidate-class", "candidate-flags":
			if state[0] != 99 {
				t.Fatal("real rejected callback preserves marker")
			}
		case "candidate-hit":
			if state[0] != 1 {
				t.Fatal("real enemy/map/facing callback accept")
			}
		case "candidate-behind":
			if state[0] != 99 {
				t.Fatal("real facing callback rejection preserves marker")
			}
		}
	}
	if *d != beforeData {
		t.Fatal("whole actual52-byte Skull record", *d, beforeData)
	}
	t.Logf("SKULL_RAW_ACTUAL_DEFAULT_C_PRODUCTION mode=%s entry=%d source=%p update=%p created=%v audio=%d default-world=true", mode, entry, source, source.UpdateData, created, sounds)
}
