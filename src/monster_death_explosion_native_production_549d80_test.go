package opennox

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterDeathExplosionActualCProduction549D80(t *testing.T) {
	modes := []string{"ember", "ember-coop", "demon", "imp", "spider", "ember-live", "demon-live", "ember-coop-live"}
	mode := os.Getenv("OPENNOX_MONSTER_DEATH_PRODUCTION_CHILD")
	if mode == "" {
		for _, mode := range modes {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestMonsterDeathExplosionActualCProduction549D80$")
				cmd.Env = append(os.Environ(), "OPENNOX_MONSTER_DEATH_PRODUCTION_CHILD="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil || !bytes.Contains(out, []byte("MONSTER_DEATH_ACTUAL_C_PRODUCTION mode="+mode)) {
					t.Fatal("actual default C/root-server production child", mode, err)
				}
			})
		}
		return
	}
	if !slices.Contains(modes, mode) {
		t.Fatal("unknown production mode", mode)
	}
	entry := uint32(0x549d80)
	if mode == "demon" || mode == "demon-live" {
		entry = 0x549e00
	} else if mode == "imp" {
		entry = 0x549e70
	} else if mode == "spider" {
		entry = 0x54a250
	}
	coop := mode == "ember-coop" || mode == "ember-coop-live"
	live := mode == "ember-live" || mode == "demon-live" || mode == "ember-coop-live"
	explosion := entry == 0x549d80 || entry == 0x549e00
	// Expected arguments come from the frozen original output, not the new
	// Go death implementation. The raw target's damage endpoint is captured;
	// root map/rays, C callbacks, packets, audio allocation and deletion run.
	data, err := os.ReadFile("server/testdata/monster_death_explosion_native_original_549d80_pc53_tozero.json.gz")
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != "951ae5cda4a689cd7b717c1479e6a201a6add9380edcbe5104df62008a90f54f" {
		t.Fatal("frozen original production argument source", err)
	}
	z, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name   string
		Events []struct {
			Kind                        string
			Outer, Inner, Force, Damage uint32
			DamageType                  uint32 `json:"damage_type"`
			Code                        uint32
		}
	}
	if err := json.Unmarshal(decoded, &rows); err != nil {
		t.Fatal(err)
	}
	flag := 0
	if coop {
		flag = 1
	}
	name := fmt.Sprintf("%08x/unit0/position3/flag%08x", entry, flag)
	var outer, inner, force float32
	var damage, typ, fxCode uint32
	found := false
	for _, row := range rows {
		if row.Name != name {
			continue
		}
		found = true
		for _, event := range row.Events {
			switch event.Kind {
			case "push":
				outer, inner, force = math.Float32frombits(event.Outer), math.Float32frombits(event.Inner), math.Float32frombits(event.Force)
			case "damage":
				damage, typ = event.Damage, event.DamageType
			case "spark", "point":
				fxCode = event.Code
			}
		}
	}
	if !found {
		t.Fatal("exact frozen original production row", name)
	}
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer, oldGet, oldFlags, oldWallDamage := noxServer, legacy.GetServer, noxflags.GetGame(), doDamageWalls
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	noxflags.ResetGame()
	if coop {
		noxflags.SetGame(noxflags.GameModeCoop)
	}
	doDamageWalls = false
	t.Cleanup(func() {
		noxServer, legacy.GetServer, doDamageWalls = oldServer, oldGet, oldWallDamage
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	s.SetFrame(1400)
	s.SetTickRate(31)
	s.spells.duration.Init(s)
	if !s.Objs.InitHigh(8) {
		t.Fatal("native real root object allocator")
	}
	t.Cleanup(s.Objs.FreeObjects)
	s.Map.Init()
	t.Cleanup(s.Map.Free)
	if s.Walls.Init() == 0 {
		t.Fatal("native empty wall map")
	}
	t.Cleanup(s.Walls.Free)
	s.NetList.Init()
	t.Cleanup(s.NetList.Free)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	soundName := sound.ID(42).String()
	wire := binary.LittleEndian.AppendUint32(nil, 1)
	wire = append(wire, byte(len(soundName)))
	wire = append(wire, soundName...)
	wire = binary.LittleEndian.AppendUint16(wire, 8)
	wire = append(wire, 0)
	wire = binary.LittleEndian.AppendUint16(wire, 40)
	wire = append(wire, 0, 0, 0, 0, 1, 's', 0)
	file := luaCompatAudioFile{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(file) || file.Len() != 0 || s.Audio.Field12(42) != 1 {
		t.Fatal("real parsed AUD descriptor for original sound42")
	}
	s.Audio.Reset()
	unit, target := s.Objs.NewObject(&server.ObjectType{}), s.Objs.NewObject(&server.ObjectType{})
	if unit == nil || target == nil {
		t.Fatal("real native objects")
	}
	ud, freeUD := alloc.NewHigh(server.MonsterUpdateData{})
	td, freeTD := alloc.NewHigh(server.MonsterUpdateData{})
	hp, freeHP := alloc.NewHigh(server.HealthData{})
	th, freeTH := alloc.NewHigh(server.HealthData{})
	t.Cleanup(freeUD)
	t.Cleanup(freeTD)
	t.Cleanup(freeHP)
	t.Cleanup(freeTH)
	*hp, *th = server.HealthData{Cur: 1000, Max: 1000}, server.HealthData{Cur: 1000, Max: 1000}
	unit.ObjClass, unit.ObjFlags, unit.UpdateData, unit.HealthData = object.ClassMonster, object.FlagActive|object.FlagEnabled, unsafe.Pointer(ud), hp
	target.ObjClass, target.ObjFlags, target.UpdateData, target.HealthData = object.ClassMonster, object.FlagActive|object.FlagEnabled, unsafe.Pointer(td), th
	unit.PosVec, unit.NewPos, unit.PrevPos = types.Ptf(320, 330), types.Ptf(320, 330), types.Ptf(320, 330)
	target.PosVec, target.NewPos, target.PrevPos = types.Ptf(340, 330), types.Ptf(340, 330), types.Ptf(340, 330)
	unit.Mass, target.Mass = 1, 1
	unit.Shape.Kind, target.Shape.Kind = server.ShapeKindCenter, server.ShapeKindCenter
	radialDamageCaptureSetup4E25B0()
	target.Damage = legacy.RadialDamageCapturePointer4E25B0(false)
	s.Map.AddObjectToIndex(target)
	t.Cleanup(func() { s.Map.RemoveObjectFromIndex(target) })
	pl := s.Players.NewRaw(1001)
	if pl == nil {
		t.Fatal("real native FX receiver")
	}
	pu, freePU := alloc.NewHigh(server.PlayerUpdateData{})
	receiver, freeReceiver := alloc.NewHigh(server.Object{})
	t.Cleanup(freePU)
	t.Cleanup(freeReceiver)
	pu.Player = pl
	*receiver = server.Object{ObjClass: object.ClassPlayer, PosVec: unit.PosVec, UpdateData: unsafe.Pointer(pu)}
	if live {
		receiver.PosVec = types.Ptf(410, 415)
	}
	pl.PlayerUnit, pl.Field10, pl.Field12 = receiver, 100, 100
	var expectedForce types.Pointf
	var expectedDamage int32
	var visitDelta [2]uint32
	if explosion {
		beforeVisit := target.Field62
		legacy.MonsterDeathExplosionPushControl549D80(unit, outer, inner, force)
		expectedForce = target.ForceVec
		if expectedForce.X <= 0 || expectedForce.Y != 0 {
			t.Fatal("independent typed C push must reach the real indexed target", expectedForce)
		}
		legacy.MapRadialDamageCProbe4E25B0(&unit.PosVec, outer, inner, int32(damage), object.DamageType(typ), unit, nil)
		count, amount, kind, gotTarget, source, weapon := legacy.RadialDamageCaptureResult4E25B0()
		if count != 1 || amount <= 0 || kind != int32(typ) || gotTarget != target || source != unit || weapon != nil {
			t.Fatal("typed C real root radial damage positive control", count, amount, kind)
		}
		expectedDamage = amount
		for i := range visitDelta {
			visitDelta[i] = target.Field62[i] - beforeVisit[i]
		}
		target.ForceVec = types.Pointf{}
	}
	target.Damage = legacy.RadialDamageCapturePointer4E25B0(live)
	doDamageWalls = false
	newPos := unit.PosVec
	if live {
		newPos = types.Ptf(500, 500)
	}
	for _, pos := range []types.Pointf{unit.PosVec, newPos} {
		control := []byte{0x99, 0x5a}
		s.Nox_xxx_netSendFxAllCli_523030(pos, control)
		if got := s.NetList.CopyPacketsA(ntype.PlayerInd(pl.PlayerInd), netlist.Kind1); !bytes.Equal(got, control) {
			t.Fatal("independent unchanged-viewport admission of old/new live points", pos, got)
		}
	}
	packet := []byte{byte(fxCode)}
	if explosion {
		packet[0] = 0x93
	}
	packet = binary.LittleEndian.AppendUint16(packet, uint16(newPos.X))
	packet = binary.LittleEndian.AppendUint16(packet, uint16(newPos.Y))
	if explosion {
		packet = append(packet, byte(fxCode))
	}
	var audioIDs []sound.ID
	var audioPacket []byte
	audioOrderOK := true
	s.Audio.OnSound(func(id sound.ID, kind int, source *server.Object, pos types.Pointf) {
		audioIDs = append(audioIDs, id)
		count, amount, dtype, hit, origin, weapon := legacy.RadialDamageCaptureResult4E25B0()
		audioOrderOK = audioOrderOK && id == 42 && kind == 0 && source == unit && pos == newPos && !unit.ObjFlags.Has(object.FlagDestroyed) && count == 1 && amount == expectedDamage && dtype == int32(typ) && hit == target && origin == unit && weapon == nil && target.ForceVec == expectedForce
		audioPacket = s.NetList.CopyPacketsA(ntype.PlayerInd(pl.PlayerInd), netlist.Kind1)
	})
	expectedUnit, expectedTarget := *unit, *target
	if explosion {
		expectedUnit.ObjFlags |= object.FlagDestroyed
		expectedUnit.DeletedAt, expectedUnit.DeletedNext = s.Frame(), s.Objs.DeletedList
		expectedUnit.PosVec = newPos
		expectedTarget.ForceVec = expectedForce
		for i := range visitDelta {
			expectedTarget.Field62[i] += visitDelta[i]
		}
	}
	values := []struct {
		ptr  unsafe.Pointer
		size uintptr
	}{{unsafe.Pointer(ud), unsafe.Sizeof(*ud)}, {unsafe.Pointer(td), unsafe.Sizeof(*td)}, {unsafe.Pointer(hp), unsafe.Sizeof(*hp)}, {unsafe.Pointer(th), unsafe.Sizeof(*th)}, {unsafe.Pointer(pl), unsafe.Sizeof(*pl)}, {unsafe.Pointer(pu), unsafe.Sizeof(*pu)}, {unsafe.Pointer(receiver), unsafe.Sizeof(*receiver)}}
	var snapshots [][]byte
	for _, record := range append(values, struct {
		ptr  unsafe.Pointer
		size uintptr
	}{unsafe.Pointer(unit), unsafe.Sizeof(*unit)}, struct {
		ptr  unsafe.Pointer
		size uintptr
	}{unsafe.Pointer(target), unsafe.Sizeof(*target)}) {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(record.ptr) <= math.MaxUint32 {
			t.Fatal("all actual production records must be above4GiB")
		}
	}
	for _, record := range values {
		snapshots = append(snapshots, bytes.Clone(unsafe.Slice((*byte)(record.ptr), int(record.size))))
	}
	if result := legacy.MonsterDeathExplosionProductionC549D80(entry, unit); result != 1 {
		t.Fatal("original public C return", result)
	}
	for i, record := range values {
		if !bytes.Equal(snapshots[i], unsafe.Slice((*byte)(record.ptr), int(record.size))) {
			t.Fatal("whole native production record changed outside declared world effects", i)
		}
	}
	for i, pair := range [][2]*server.Object{{unit, &expectedUnit}, {target, &expectedTarget}} {
		a, b := unsafe.Slice((*byte)(unsafe.Pointer(pair[0])), int(unsafe.Sizeof(*pair[0]))), unsafe.Slice((*byte)(unsafe.Pointer(pair[1])), int(unsafe.Sizeof(*pair[1])))
		if !bytes.Equal(a, b) {
			for off, v := range a {
				if v != b[off] {
					t.Fatalf("complete native production Object differs record=%d offset=%x got=%02x want=%02x", i, off, v, b[off])
				}
			}
		}
	}
	if explosion {
		if !audioOrderOK || !slices.Equal(audioIDs, []sound.ID{42}) || !bytes.Equal(audioPacket, packet) || s.Objs.DeletedList != unit {
			t.Fatal("actual C push/damage/spark/audio/delete order or wire packet", audioOrderOK, audioIDs, audioPacket)
		}
		var audio []*server.AudioEvent
		s.Audio.EachEvent(func(ev *server.AudioEvent) { audio = append(audio, ev) })
		if len(audio) != 1 || audio[0].Sound != 42 || audio[0].Obj != unit || audio[0].Pos != newPos || audio[0].Kind != 0 || audio[0].Code != 0 {
			t.Fatal("real native AudioEvent allocation")
		}
	} else if len(audioIDs) != 0 || !bytes.Equal(s.NetList.CopyPacketsA(ntype.PlayerInd(pl.PlayerInd), netlist.Kind1), packet) || s.Objs.DeletedList != nil {
		t.Fatal("point-only original packet/lifecycle")
	}
	t.Logf("MONSTER_DEATH_ACTUAL_C_PRODUCTION mode=%s unit=%p target=%p real-root-map-ray-C-target-packet-audio-delete=true whole-native-records=9", mode, unit, target)
}
