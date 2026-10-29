package opennox

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Actual C entries compose with production cancellation, buff exports,
// position-delta and synchronization. Spell definitions are synthetic; player
// protection handle zero is inactive, and no peer connection is opened.
func TestSpellOvalShieldPublicCNativeComposition531490(t *testing.T) {
	const childEnv = "NOX_OVAL_SHIELD_NATIVE_COMPOSITION_CHILD"
	mode := os.Getenv(childEnv)
	if mode == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"player-lifecycle", "player-refresh-cancel", "player-word-byte-wrap", "player-blocked-update", "player-destroyed-update", "monster-stationary-update", "monster-moved-update"} {
			t.Run(mode, func(t *testing.T) {
				cmd := exec.Command(binary, "-test.run=^TestSpellOvalShieldPublicCNativeComposition531490$", "-test.v")
				cmd.Env = append(os.Environ(), childEnv+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual native C OvalShield composition: %v", err)
				}
			})
		}
		return
	}
	handles.Init()
	t.Cleanup(handles.Release)
	s := NewServer(nil, nil, strman.New())
	t.Cleanup(s.Close)
	type cancellationReport struct {
		recipient, remove, sequence, result int
		packet                              []byte
		related                             *server.Object
	}
	var reports []cancellationReport
	productionSend := s.NetSendPacketXxx
	s.NetSendPacketXxx = func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
		result := productionSend(recipient, packet, related, remove, sequence)
		reports = append(reports, cancellationReport{recipient, remove, sequence, result, append([]byte(nil), packet...), related})
		return result
	}
	t.Cleanup(func() { s.NetSendPacketXxx = productionSend })
	oldServer, oldGet := noxServer, legacy.GetServer
	noxServer, legacy.GetServer = s, func() legacy.Server { return s }
	t.Cleanup(func() { noxServer, legacy.GetServer = oldServer, oldGet })
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(oldFlags)
	t.Cleanup(func() { noxflags.UnsetGame(noxflags.GetGame()); noxflags.SetGame(oldFlags) })
	s.spells.duration.Init(s)
	s.SetFrame(142)
	s.SetTickRate(30)
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	dir, oldData := t.TempDir(), datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	defs := []things.Spell{
		{ID: spell.SPELL_FIREBALL.String(), Flags: things.SpellOffensive},
		{ID: spell.SPELL_OVAL_SHIELD.String(), Flags: things.SpellDefensive},
	}
	if err := things.WriteSpellsYAML(filepath.Join(dir, "spells.yml"), defs); err != nil {
		t.Fatal(err)
	}
	data, _ := alloc.Make([]byte{}, 4) // Empty original SPEL section; production loader reads YAML.
	file := binfile.NewMemFile(unsafe.Pointer(&data[0]), len(data))
	t.Cleanup(file.Free)
	if err := s.Spells.Read(file, nil, false); err != nil {
		t.Fatal(err)
	}
	if !s.Spells.Flags(spell.SPELL_FIREBALL).Has(things.SpellOffensive) || s.Spells.Flags(spell.SPELL_OVAL_SHIELD).Has(things.SpellOffensive) {
		t.Fatal("production spell flags")
	}
	type recordFrame struct {
		Before [32]byte
		Record server.DurSpell
		After  [32]byte
	}
	type objectFrame struct {
		Before [32]byte
		Object server.Object
		After  [32]byte
	}
	type dataFrame struct {
		Before [32]byte
		Data   server.PlayerUpdateData
		After  [32]byte
	}
	type playerFrame struct {
		Before [32]byte
		Player server.Player
		After  [32]byte
	}
	f, freeRecord := alloc.New(recordFrame{})
	t.Cleanup(freeRecord)
	*f = recordFrame{}
	source, freeSource := alloc.New(objectFrame{})
	t.Cleanup(freeSource)
	*source = objectFrame{}
	other, freeOther := alloc.New(objectFrame{})
	t.Cleanup(freeOther)
	*other = objectFrame{}
	update, freeUpdate := alloc.New(dataFrame{})
	t.Cleanup(freeUpdate)
	*update = dataFrame{}
	info, freePlayer := alloc.New(playerFrame{})
	t.Cleanup(freePlayer)
	*info = playerFrame{}
	offensive, freeOffensive := alloc.New(recordFrame{})
	t.Cleanup(freeOffensive)
	*offensive = recordFrame{}
	defensive, freeDefensive := alloc.New(recordFrame{})
	t.Cleanup(freeDefensive)
	*defensive = recordFrame{}
	otherSpell, freeOtherSpell := alloc.New(recordFrame{})
	t.Cleanup(freeOtherSpell)
	*otherSpell = recordFrame{}
	for i := range f.Before {
		f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
		source.Before[i], source.After[i] = byte(0x31+i), byte(0xc7-i)
		other.Before[i], other.After[i] = byte(0x21+i), byte(0xb7-i)
		update.Before[i], update.After[i] = byte(0x71+i), byte(0xa7-i)
		info.Before[i], info.After[i] = byte(0x11+i), byte(0x97-i)
		offensive.Before[i], offensive.After[i] = byte(0x41+i), byte(0x87-i)
		defensive.Before[i], defensive.After[i] = byte(0x61+i), byte(0x77-i)
		otherSpell.Before[i], otherSpell.After[i] = byte(0x81+i), byte(0x67-i)
	}
	update.Data.Player = &info.Player
	source.Object = server.Object{ObjClass: object.ClassPlayer, TypeInd: 713, UpdateData: unsafe.Pointer(&update.Data), PosVec: types.Ptf(100, 200)}
	other.Object = server.Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(7, 11)}
	const shield = server.ENCHANT_REFLECTIVE_SHIELD
	const unrelated = server.ENCHANT_LIGHT
	source.Object.Buffs = 1 << unrelated
	source.Object.BuffsDur[unrelated], source.Object.BuffsPower[unrelated] = 19, 2
	f.Record = server.DurSpell{ID: 0xabcd, Spell: uint32(spell.SPELL_OVAL_SHIELD), Level: 3, Obj12: &source.Object, Caster16: &source.Object, Target48: &source.Object, Field72: 0x55aa, Pos: types.Ptf(math.Float32frombits(0x3fdccccc), -22.25), Frame68: 0x87654321, Flags88: 0x77556601}
	wantDuration, wantPower, wantExpiry := uint16(1800), uint8(3), uint32(1942)
	monster := mode == "monster-stationary-update" || mode == "monster-moved-update"
	if monster {
		source.Object.ObjClass = object.ClassMonster
		source.Object.UpdateData = nil
		f.Record.Pos = source.Object.PosVec
	}
	if mode == "player-word-byte-wrap" {
		s.SetTickRate(0xffffffff)
		f.Record.Level = 0xffffffff
		wantDuration, wantPower, wantExpiry = 20, 255, 162
	}
	if mode == "player-refresh-cancel" {
		source.Object.Buffs |= 1<<shield | 1<<server.ENCHANT_INVISIBLE
		source.Object.BuffsDur[shield], source.Object.BuffsPower[shield] = 9, 2
		source.Object.BuffsDur[server.ENCHANT_INVISIBLE], source.Object.BuffsPower[server.ENCHANT_INVISIBLE] = 12, 4
		offensive.Record = server.DurSpell{Spell: uint32(spell.SPELL_FIREBALL), Caster16: &source.Object, Flags88: 0xaabbcc00, Next: &defensive.Record}
		defensive.Record = server.DurSpell{Spell: uint32(spell.SPELL_OVAL_SHIELD), Caster16: &source.Object, Flags88: 0x11223300, Next: &otherSpell.Record}
		otherSpell.Record = server.DurSpell{Spell: uint32(spell.SPELL_FIREBALL), Caster16: &other.Object, Flags88: 0x44556600}
		s.Spells.Dur.List = &offensive.Record
		t.Cleanup(func() { s.Spells.Dur.List = nil }) // C-owned fixture nodes outlive list cleanup.
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{f.Record.C(), source.Object.CObj(), other.Object.CObj(), unsafe.Pointer(&update.Data), unsafe.Pointer(&info.Player), offensive.Record.C(), defensive.Record.C(), otherSpell.Record.C()} {
			if uintptr(p) <= 0xffffffff {
				t.Fatal("native C-owned OvalShield links must exceed 4 GiB")
			}
		}
	}
	raw := func(p unsafe.Pointer, n uintptr) []byte { return unsafe.Slice((*byte)(p), n) }
	copyBytes := func(p unsafe.Pointer, n uintptr) []byte { return append([]byte(nil), raw(p, n)...) }
	expectedRecord, expectedOffensive := *f, *offensive
	dataBefore := copyBytes(unsafe.Pointer(update), unsafe.Sizeof(*update))
	playerBefore := copyBytes(unsafe.Pointer(info), unsafe.Sizeof(*info))
	otherBefore := copyBytes(unsafe.Pointer(other), unsafe.Sizeof(*other))
	defensiveBefore := copyBytes(unsafe.Pointer(defensive), unsafe.Sizeof(*defensive))
	otherSpellBefore := copyBytes(unsafe.Pointer(otherSpell), unsafe.Sizeof(*otherSpell))
	sourceBefore := copyBytes(unsafe.Pointer(source), unsafe.Sizeof(*source))
	allowed := make([]bool, len(sourceBefore))
	allow := func(p unsafe.Pointer, n uintptr) {
		off := uintptr(p) - uintptr(unsafe.Pointer(source))
		for i := off; i < off+n; i++ {
			allowed[i] = true
		}
	}
	allow(unsafe.Pointer(&source.Object.Field38), unsafe.Sizeof(source.Object.Field38))
	allow(unsafe.Pointer(&source.Object.Field140), unsafe.Sizeof(source.Object.Field140))
	allow(unsafe.Pointer(&source.Object.Buffs), unsafe.Sizeof(source.Object.Buffs))
	for _, enchant := range []server.EnchantID{shield, server.ENCHANT_INVISIBLE} {
		allow(unsafe.Pointer(&source.Object.BuffsDur[enchant]), unsafe.Sizeof(source.Object.BuffsDur[enchant]))
		allow(unsafe.Pointer(&source.Object.BuffsPower[enchant]), unsafe.Sizeof(source.Object.BuffsPower[enchant]))
	}
	if mode == "player-destroyed-update" {
		allow(unsafe.Pointer(&source.Object.ObjFlags), unsafe.Sizeof(source.Object.ObjFlags))
	}
	if mode == "monster-moved-update" {
		allow(unsafe.Pointer(&source.Object.PosVec), unsafe.Sizeof(source.Object.PosVec))
	}
	got := ccall.CallIntPtr(legacy.Get_sub_531490(), f.Record.C())
	if monster {
		if got != 1 || f.Record.Frame68 != expectedRecord.Record.Frame68 || source.Object.HasEnchant(shield) {
			t.Fatal("actual nonplayer create guard")
		}
	} else {
		if got != 0 || f.Record.Frame68 != wantExpiry || !source.Object.HasEnchant(shield) || source.Object.BuffsDur[shield] != wantDuration || source.Object.BuffsPower[shield] != wantPower {
			t.Fatalf("actual C create: return=%d expiry=%d duration=%d power=%d", got, f.Record.Frame68, source.Object.BuffsDur[shield], source.Object.BuffsPower[shield])
		}
		expectedRecord.Record.Frame68 = wantExpiry
	}
	if mode == "player-refresh-cancel" {
		if offensive.Record.Flags88 != 0xaabbcc01 || source.Object.HasEnchant(server.ENCHANT_INVISIBLE) || source.Object.BuffsDur[server.ENCHANT_INVISIBLE] != 0 || source.Object.BuffsPower[server.ENCHANT_INVISIBLE] != 0 || s.Spells.Dur.List != &offensive.Record {
			t.Fatal("production offensive cancellation and Invisibility removal")
		}
		if len(reports) != 1 {
			t.Fatalf("actual production cancellation reports=%d", len(reports))
		}
		report := reports[0]
		if report.recipient != 0 || report.remove != 1 || report.sequence != 0 || report.related != nil || report.result != 1 || len(report.packet) != 6 || report.packet[0] != byte(netmsg.MSG_REPORT_SPELL_STAT) || binary.LittleEndian.Uint32(report.packet[1:5]) != uint32(spell.SPELL_FIREBALL) || report.packet[5] != 15 {
			t.Fatal("actual C important-packet allocation and cancellation status payload")
		}
		expectedOffensive.Record.Flags88 = 0xaabbcc01
	} else if len(reports) != 0 {
		t.Fatal("empty cancellation list emitted a report")
	}
	wantUpdate := 0
	switch mode {
	case "player-blocked-update":
		source.Object.Buffs |= 1 << 8
		wantUpdate = 1
	case "player-destroyed-update":
		source.Object.ObjFlags |= object.FlagDestroyed
		wantUpdate = 1
	case "monster-moved-update":
		source.Object.PosVec.X += 5
		wantUpdate = 1
	}
	if got := ccall.CallIntPtr(legacy.Get_sub_5314F0(), f.Record.C()); got != wantUpdate {
		t.Fatalf("actual C update=%d want=%d", got, wantUpdate)
	}
	ccall.CallVoidPtr(legacy.Get_sub_531560(), f.Record.C())
	wantBuffs := uint32(1) << unrelated
	if mode == "player-blocked-update" {
		wantBuffs |= 1 << 8
	}
	if source.Object.Buffs != wantBuffs || source.Object.BuffsDur[shield] != 0 || source.Object.BuffsPower[shield] != 0 || source.Object.BuffsDur[unrelated] != 19 || source.Object.BuffsPower[unrelated] != 2 {
		t.Fatal("actual C destroy must clear only Shield; unrelated buff survives")
	}
	for _, check := range []struct {
		before []byte
		p      unsafe.Pointer
		size   uintptr
	}{
		{raw(unsafe.Pointer(&expectedRecord), unsafe.Sizeof(expectedRecord)), unsafe.Pointer(f), unsafe.Sizeof(*f)},
		{raw(unsafe.Pointer(&expectedOffensive), unsafe.Sizeof(expectedOffensive)), unsafe.Pointer(offensive), unsafe.Sizeof(*offensive)},
		{dataBefore, unsafe.Pointer(update), unsafe.Sizeof(*update)}, {playerBefore, unsafe.Pointer(info), unsafe.Sizeof(*info)},
		{otherBefore, unsafe.Pointer(other), unsafe.Sizeof(*other)}, {defensiveBefore, unsafe.Pointer(defensive), unsafe.Sizeof(*defensive)}, {otherSpellBefore, unsafe.Pointer(otherSpell), unsafe.Sizeof(*otherSpell)},
	} {
		if !bytes.Equal(check.before, raw(check.p, check.size)) {
			t.Fatal("record/update/player/other-object frame or canary changed")
		}
	}
	for i, value := range raw(unsafe.Pointer(source), unsafe.Sizeof(*source)) {
		if value != sourceBefore[i] && !allowed[i] {
			t.Fatalf("nonoutput source canary at%d: before=%02x after=%02x", i, sourceBefore[i], value)
		}
	}
	t.Logf("OVAL SHIELD PUBLIC C COMPOSITION PASS:mode=%s record=%p target=%p frame=%d", mode, &f.Record, &source.Object, s.Frame())
}
