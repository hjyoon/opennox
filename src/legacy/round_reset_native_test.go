package legacy

import (
	"bytes"
	"encoding/binary"
	"testing"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type roundResetSettingsTestServer struct{ Server }

func (roundResetSettingsTestServer) GetFlag3592() bool { return false }

func TestRoundResetGameSettingsPacketsExcludeHost(t *testing.T) {
	objectStatusTestServer(t, 4)
	setImportantRecipientMaskC((1 << server.HostPlayerIndex) | (1 << 6))
	originalServer := GetServer
	GetServer = func() Server { return roundResetSettingsTestServer{originalServer()} }
	t.Cleanup(func() { GetServer = originalServer })

	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameFlag(0x80030013))
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	settings := (*server.Settings2)(cliGamedataNative416590(0))
	oldSettings := *settings
	name := memmap.Slice(0x5d4594, 1324)[:16]
	oldName := bytes.Clone(name)
	subflags := (*uint32)(settingFlagsNativeCell409E40(0))
	playerLimit := memmap.PtrUint32(0x5d4594, 3464)
	active := memmap.PtrUint32(0x587000, 4660)
	deadline := memmap.PtrUint32(0x5d4594, 3468)
	modeIndex := uintptr(sub409A70Native(0x10))
	modeLimit := memmap.PtrUint16(0x5d4594, 3488+2*modeIndex)
	timed := memmap.PtrUint8(0x5d4594, 3500+modeIndex)
	oldSubflags, oldLimit, oldActive, oldDeadline := *subflags, *playerLimit, *active, *deadline
	oldModeLimit, oldTimed := *modeLimit, *timed
	t.Cleanup(func() {
		*settings = oldSettings
		copy(name, oldName)
		*subflags, *playerLimit, *active, *deadline = oldSubflags, oldLimit, oldActive, oldDeadline
		*modeLimit, *timed = oldModeLimit, oldTimed
	})
	*settings = server.Settings2{
		Field24: server.Settings3{Vals: [5]uint32{0x04030201, 0x08070605, 0x0c0b0a09, 0x100f0e0d, 0x14131211}},
		Field44: [4]byte{0x15, 0x16, 0x17, 0x18},
		Field48: 0x1c1b1a19,
		Field52: 0x10,
	}
	copy(name, "0123456789abcde\x00")
	*subflags, *playerLimit, *active, *deadline = 0x88776655, 0x12345640, 0, 0x12340000
	*modeLimit, *timed = 0xfedc, 0x81

	restore := pushTimerNativeObserver40A1F0(&timerNativeObserver40A1F0{
		Service: func(site uint32, args []uint32) (timerNativeReply40A1F0, bool) {
			if site != 0x40a233 || len(args) != 0 {
				t.Fatalf("timer service %#x with %x", site, args)
			}
			return timerNativeReply40A1F0{EAX: 0x1020}, true
		},
	})
	defer restore()

	game := [20]byte{0xaf, 19, 0, 0, 0, 0x9a, 3, 1, 0, 0x10, 0, 3, 0, 0x55, 0x66, 0x77, 0x88, 0x40, 0xdc, 0x81}
	details := [49]byte{0xb0}
	copy(details[1:16], name[:15])
	for i := range details[17:45] {
		details[17+i] = byte(i + 1)
	}
	for _, enabled := range []uint32{0, 1} {
		*active = enabled
		if got := nox_xxx_netGameSettings_4DEF00(); got != 1 {
			t.Fatalf("settings returned %d with timer flag %d", got, enabled)
		}
		first, _ := importantNativeListC()
		// Both protocol builds use the same packet layout and a distinct version.
		if first == nil {
			t.Fatal("missing settings details packet")
		}
		previous := importantPacketNextC(first)
		if previous == nil {
			t.Fatal("missing game settings packet")
		}
		version := binary.LittleEndian.Uint32((*importantPacketLegacy)(previous).Payload[5:9])
		if version != 0x0001039a && version != 0x000f039a {
			t.Fatalf("unexpected protocol version %#08x", version)
		}
		binary.LittleEndian.PutUint32(game[5:9], version)
		if enabled != 0 {
			binary.LittleEndian.PutUint32(details[45:49], 0x1233efe0)
		}
		checkObjectStatusPacket(t, first, 0x9f, details[:], 1, 0)
		checkObjectStatusPacket(t, previous, 0x9f, game[:], 1, 0)
	}
}
