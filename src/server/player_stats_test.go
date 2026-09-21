package server

import (
	"encoding/binary"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	"github.com/opennox/libs/player"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func playerStatsTestBalance() *balance.File {
	return &balance.File{
		Global: balance.Config{},
		Tags: map[balance.Tag]balance.Config{
			balance.TagArena: {
				"basehealth":          balance.Float(10),
				"basemana":            balance.Float(20),
				"basespeed":           balance.Float(30),
				"basestrength":        balance.Float(40),
				"warriormaxhealth":    balance.Float(100),
				"warriormaxmana":      balance.Float(200),
				"warriormaxspeed":     balance.Float(300),
				"warriormaxstrength":  balance.Float(400),
				"wizardmaxhealth":     balance.Float(110),
				"wizardmaxmana":       balance.Float(210),
				"wizardmaxspeed":      balance.Float(310),
				"wizardmaxstrength":   balance.Float(410),
				"conjurermaxhealth":   balance.Float(120),
				"conjurermaxmana":     balance.Float(220),
				"conjurermaxspeed":    balance.Float(320),
				"conjurermaxstrength": balance.Float(420),
			},
			balance.TagSolo: {
				"basehealth":          balance.Float(1010),
				"basemana":            balance.Float(1020),
				"basespeed":           balance.Float(1030),
				"basestrength":        balance.Float(1040),
				"warriormaxhealth":    balance.Float(1100),
				"warriormaxmana":      balance.Float(1200),
				"warriormaxspeed":     balance.Float(1300),
				"warriormaxstrength":  balance.Float(1400),
				"wizardmaxhealth":     balance.Float(1110),
				"wizardmaxmana":       balance.Float(1210),
				"wizardmaxspeed":      balance.Float(1310),
				"wizardmaxstrength":   balance.Float(1410),
				"conjurermaxhealth":   balance.Float(1120),
				"conjurermaxmana":     balance.Float(1220),
				"conjurermaxspeed":    balance.Float(1320),
				"conjurermaxstrength": balance.Float(1420),
			},
		},
	}
}

func playerStatsTestRestoreCoop(t *testing.T) {
	t.Helper()
	wasCoop := noxflags.HasGame(noxflags.GameModeCoop)
	t.Cleanup(func() {
		if wasCoop {
			noxflags.SetGame(noxflags.GameModeCoop)
		} else {
			noxflags.UnsetGame(noxflags.GameModeCoop)
		}
	})
}

func TestClassStatsNativeLayout(t *testing.T) {
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "size", got: unsafe.Sizeof(ClassStats{}), want: 16},
		{name: "health", got: unsafe.Offsetof(ClassStats{}.Health), want: 0},
		{name: "mana", got: unsafe.Offsetof(ClassStats{}.Mana), want: 4},
		{name: "speed", got: unsafe.Offsetof(ClassStats{}.Speed), want: 8},
		{name: "strength", got: unsafe.Offsetof(ClassStats{}.Strength), want: 12},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("ClassStats %s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPlayerClassStatsSelectorsAndSnapshots(t *testing.T) {
	var players serverPlayers
	players.Stats.Base = ClassStats{Health: 1, Mana: 2, Speed: 3, Strength: 4}
	players.Stats.Warrior = ClassStats{Health: 5, Mana: 6, Speed: 7, Strength: 8}
	players.Stats.Wizard = ClassStats{Health: 9, Mana: 10, Speed: 11, Strength: 12}
	players.Stats.Conjurer = ClassStats{Health: 13, Mana: 14, Speed: 15, Strength: 16}
	for _, tc := range []struct {
		class player.Class
		want  *ClassStats
	}{
		{player.Warrior, &players.Stats.Warrior},
		{player.Wizard, &players.Stats.Wizard},
		{player.Conjurer, &players.Stats.Conjurer},
		{player.Class(0xff), &players.Stats.Base},
	} {
		if got := players.ClassStats(tc.class); got != tc.want {
			t.Errorf("ClassStats(%d) = %p, want %p", tc.class, got, tc.want)
		}
	}

	players.DefaultStatMult()
	wantDefault := struct {
		Warrior  ClassStats
		Wizard   ClassStats
		Conjurer ClassStats
	}{
		Warrior:  ClassStats{Health: 3, Mana: 1, Speed: 1, Strength: 1},
		Wizard:   ClassStats{Health: 3, Mana: 3, Speed: 1, Strength: 1},
		Conjurer: ClassStats{Health: 3, Mana: 3, Speed: 1, Strength: 1},
	}
	if players.Mult != wantDefault {
		t.Fatalf("default multipliers = %#v, want %#v", players.Mult, wantDefault)
	}
	for _, class := range []player.Class{player.Warrior, player.Wizard, player.Conjurer} {
		if players.ClassStatsMult(class) == nil {
			t.Errorf("ClassStatsMult(%d) = nil", class)
		}
	}
	if got := players.ClassStatsMult(player.Class(0xff)); got != nil {
		t.Errorf("ClassStatsMult(invalid) = %p, want nil", got)
	}

	players.SaveStats2()
	saved := players.Mult
	players.Mult = struct {
		Warrior  ClassStats
		Wizard   ClassStats
		Conjurer ClassStats
	}{}
	players.LoadStats2()
	if players.Mult != saved {
		t.Fatalf("restored multipliers = %#v, want %#v", players.Mult, saved)
	}
}

func TestCalcClassStatsUsesModeAndMatchingMultipliers(t *testing.T) {
	playerStatsTestRestoreCoop(t)
	noxflags.UnsetGame(noxflags.GameModeCoop)

	srv := new(Server)
	srv.Balance.file = playerStatsTestBalance()
	srv.Players.Mult.Warrior = ClassStats{Health: 2, Mana: 3, Speed: 5, Strength: 7}
	srv.Players.Mult.Wizard = ClassStats{Health: 11, Mana: 13, Speed: 17, Strength: 19}
	srv.Players.Mult.Conjurer = ClassStats{Health: 23, Mana: 29, Speed: 31, Strength: 37}

	srv.CalcClassStats()
	if got, want := srv.Players.Stats.Base, (ClassStats{Health: 10, Mana: 20, Speed: 30, Strength: 40}); got != want {
		t.Errorf("arena base = %#v, want %#v", got, want)
	}
	if got, want := srv.Players.Stats.Warrior, (ClassStats{Health: 200, Mana: 600, Speed: 1500, Strength: 2800}); got != want {
		t.Errorf("arena warrior = %#v, want %#v", got, want)
	}
	if got, want := srv.Players.Stats.Wizard, (ClassStats{Health: 1210, Mana: 2730, Speed: 5270, Strength: 7790}); got != want {
		t.Errorf("arena wizard = %#v, want %#v", got, want)
	}
	if got, want := srv.Players.Stats.Conjurer, (ClassStats{Health: 2760, Mana: 6380, Speed: 9920, Strength: 15540}); got != want {
		t.Errorf("arena conjurer = %#v, want %#v", got, want)
	}

	noxflags.SetGame(noxflags.GameModeCoop)
	srv.CalcClassStats()
	if got, want := srv.Players.Stats.Base, (ClassStats{Health: 1010, Mana: 1020, Speed: 1030, Strength: 1040}); got != want {
		t.Errorf("solo base = %#v, want %#v", got, want)
	}
	if got, want := srv.Players.Stats.Warrior, (ClassStats{Health: 2200, Mana: 3600, Speed: 6500, Strength: 9800}); got != want {
		t.Errorf("solo warrior = %#v, want %#v", got, want)
	}
	if got, want := srv.Players.Stats.Wizard, (ClassStats{Health: 12210, Mana: 15730, Speed: 22270, Strength: 26790}); got != want {
		t.Errorf("solo wizard = %#v, want %#v", got, want)
	}
	if got, want := srv.Players.Stats.Conjurer, (ClassStats{Health: 25760, Mana: 35380, Speed: 40920, Strength: 52540}); got != want {
		t.Errorf("solo conjurer = %#v, want %#v", got, want)
	}
}

func TestOnClassStatsUpdatesOnlySelectedClass(t *testing.T) {
	playerStatsTestRestoreCoop(t)
	noxflags.UnsetGame(noxflags.GameModeCoop)
	srv := new(Server)
	srv.Balance.file = playerStatsTestBalance()
	srv.Players.Mult.Warrior = ClassStats{Health: 1, Mana: 1, Speed: 1, Strength: 1}
	srv.Players.Mult.Wizard = ClassStats{Health: 1, Mana: 1, Speed: 1, Strength: 1}
	srv.Players.Mult.Conjurer = ClassStats{Health: 1, Mana: 1, Speed: 1, Strength: 1}

	wantMult := ClassStats{Health: 2, Mana: 3, Speed: 5, Strength: 7}
	srv.OnClassStats(player.Wizard, wantMult)
	if got := srv.Players.Mult.Wizard; got != wantMult {
		t.Fatalf("wizard multiplier = %#v, want %#v", got, wantMult)
	}
	if got, want := srv.Players.Mult.Warrior, (ClassStats{Health: 1, Mana: 1, Speed: 1, Strength: 1}); got != want {
		t.Errorf("warrior multiplier changed to %#v", got)
	}
	if got, want := srv.Players.Mult.Conjurer, (ClassStats{Health: 1, Mana: 1, Speed: 1, Strength: 1}); got != want {
		t.Errorf("conjurer multiplier changed to %#v", got)
	}
	if got, want := srv.Players.Stats.Wizard, (ClassStats{Health: 220, Mana: 630, Speed: 1550, Strength: 2870}); got != want {
		t.Errorf("wizard stats = %#v, want %#v", got, want)
	}
}

func TestClassStatsNetWireOrder(t *testing.T) {
	stats := ClassStats{Health: 1.25, Mana: -2.5, Speed: 3.75, Strength: -4.5}
	msg := classStatsToNet(stats)
	if msg.Health != stats.Health || msg.Mana != stats.Mana || msg.Speed != stats.Speed || msg.Strength != stats.Strength {
		t.Fatalf("network stats = %#v, want %#v", msg, stats)
	}
	buf := make([]byte, msg.EncodeSize())
	if n, err := msg.Encode(buf); err != nil || n != len(buf) {
		t.Fatalf("encode = %d, %v", n, err)
	}
	wantBits := []uint32{
		math.Float32bits(stats.Health),
		math.Float32bits(stats.Mana),
		math.Float32bits(stats.Strength),
		math.Float32bits(stats.Speed),
	}
	for i, want := range wantBits {
		if got := binary.LittleEndian.Uint32(buf[4*i:]); got != want {
			t.Errorf("wire field %d = %#x, want %#x", i, got, want)
		}
	}
}
