package legacy

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMapLoadFlagNativeCEntries(t *testing.T) {
	flag := memmap.PtrUint32(0x5D4594, 1563072)
	old := *flag
	t.Cleanup(func() { *flag = old })
	for _, value := range []int32{0, 1, -1, -2147483648, 2147483647} {
		if got := setMapLoadFlagC(value); got != value {
			t.Fatalf("setter returned %d, want %d", got, value)
		}
		if got := Nox_xxx_mapLoadRequired_4DCC80(); got != int(value) || *flag != uint32(value) {
			t.Fatalf("Go read %d and stored word %#x, want %d", got, *flag, value)
		}
		Nox_xxx_mapLoadOrSaveMB_4DCC70(int(value))
		if got := mapLoadFlagC(); got != value {
			t.Fatalf("C read %d, want %d", got, value)
		}
	}
}

func TestSaveGateNativeCEntry4DCCB0(t *testing.T) {
	s := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(s.Close)
	unit, freeUnit := alloc.NewHigh(server.Object{})
	t.Cleanup(freeUnit)
	host := s.Players.ByIndRaw(31)
	host.PlayerUnit = unit

	oldServer, oldFrame := GetServer, gameFrameHook
	oldBusy, oldCheck := Sub_4DCC90, Sub_4DCC10
	oldFlags := noxflags.GetGame()
	cursor := memmap.PtrUint32(0x5D4594, 1096672)
	stamp := memmap.PtrUint32(0x5D4594, 1563068)
	pause := (*uint32)(startupPauseDefaultCell4134D0(5))
	oldCursor, oldStamp, oldPause := *cursor, *stamp, *pause
	t.Cleanup(func() {
		GetServer, gameFrameHook = oldServer, oldFrame
		Sub_4DCC90, Sub_4DCC10 = oldBusy, oldCheck
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
		*cursor, *stamp, *pause = oldCursor, oldStamp, oldPause
	})
	GetServer = func() Server { return &chatLegacyServer528AC0{srv: s} }

	for _, tc := range []struct {
		name             string
		inactive, noUnit bool
		nonCoop          bool
		busy             int
		cursor, paused   uint32
		frame, last      uint32
		flags            object.Flags
		saveCheck        int
		want, wantChecks int32
	}{
		{name: "non-coop", nonCoop: true, want: 1},
		{name: "no active host", inactive: true},
		{name: "host without unit", noUnit: true},
		{name: "save in progress", busy: 1, frame: 30},
		{name: "startup paused", paused: 1, frame: 30},
		{name: "cooldown", frame: 29},
		{name: "cursor busy", cursor: 0x80000000, frame: 30},
		{name: "airborne", frame: 30, flags: object.FlagAirborne},
		{name: "failed final check", frame: 30, wantChecks: 1},
		{name: "save permitted", frame: 30, saveCheck: 2, want: 1, wantChecks: 1},
		{name: "frame wraps", frame: 4, last: 0xffffffe6, saveCheck: 1, want: 1, wantChecks: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			noxflags.ResetGame()
			if !tc.nonCoop {
				noxflags.SetGame(noxflags.GameModeCoop)
			}
			host.Active, host.PlayerUnit = 1, unit
			if tc.inactive {
				host.Active = 0
			}
			if tc.noUnit {
				host.PlayerUnit = nil
			}
			unit.ObjFlags = tc.flags
			*cursor, *stamp, *pause = tc.cursor, tc.last, tc.paused
			if got := saveGateCursorC(); got != int32(tc.cursor) {
				t.Fatalf("cursor word read as %d, want %d", got, int32(tc.cursor))
			}
			gameFrameHook = func() uint32 { return tc.frame }
			Sub_4DCC90 = func() int { return tc.busy }
			var checks int32
			Sub_4DCC10 = func(got *server.Object) int {
				if got != unit {
					t.Fatalf("checked unit %p, want %p", got, unit)
				}
				checks++
				return tc.saveCheck
			}
			if got := callSaveGateC(); got != tc.want || checks != tc.wantChecks {
				t.Fatalf("gate = %d with %d final checks, want %d with %d checks", got, checks, tc.want, tc.wantChecks)
			}
		})
	}
}
