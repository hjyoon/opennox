package opennox

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unicode/utf16"
	"unsafe"

	"github.com/opennox/libs/noxnet"
	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

// These tests call the production callback. The retained pre-port C comparison
// used (wchar_t*)settings + 39, whereas configuration stores ServerPass78.
// This is a source-backed offset contract, not original-binary execution proof.
type networkPasswordFixtureR283 struct {
	settings         *server.Settings
	window           []byte
	prepared         []byte
	preparedSettings server.Settings
}

func newNetworkPasswordFixtureR283(t *testing.T, password string, minimum, maximum int16) *networkPasswordFixtureR283 {
	t.Helper()
	sst := getServerSettings()
	const guard = 16
	window := unsafe.Slice((*byte)(memmap.PtrOff(0x5D4594, 371516-guard)), guard+int(unsafe.Sizeof(*sst))+guard)
	original := append([]byte(nil), window...)
	t.Cleanup(func() {
		copy(window, original)
		if !bytes.Equal(window, original) {
			t.Error("Settings and adjacent-byte cleanup did not restore the original record")
		}
	})
	// The record contains no Go pointers. Do not poison adjacent memmap globals.
	settingsBytes := window[guard : len(window)-guard]
	for i := range settingsBytes {
		settingsBytes[i] = 0xa7
	}
	// A valid, differently located NUL-terminated UTF-16 string at the old byte
	// offset makes a mistaken read deterministic and in-bounds on both word sizes.
	copy(settingsBytes[39:], []byte{'d', 0, 'e', 0, 'c', 0, 'o', 0, 'y', 0, 0, 0})
	clear(sst.ServerPass78[:])
	words := utf16.Encode([]rune(password))
	if len(words) > 8 {
		t.Fatal("fixture password exceeds the configured eight UTF-16-word limit")
	}
	copy(sst.ServerPass78[:], words)
	binary.LittleEndian.PutUint16(sst.RestrictedPingMind105[:], uint16(minimum))
	binary.LittleEndian.PutUint16(sst.RestrictedPingMax107[:], uint16(maximum))
	return &networkPasswordFixtureR283{settings: sst, window: window, prepared: append([]byte(nil), window...), preparedSettings: *sst}
}

func (f *networkPasswordFixtureR283) checkUnchanged(t *testing.T) {
	t.Helper()
	if *f.settings != f.preparedSettings || !bytes.Equal(f.window, f.prepared) {
		t.Error("authentication changed Settings, padding, or adjacent canary bytes")
	}
}

func checkNetworkPasswordResultR283(t *testing.T, got netmsg.Message, want string) {
	t.Helper()
	switch want {
	case "join":
		if _, ok := got.(*noxnet.MsgJoinOK); !ok {
			t.Errorf("authentication result = %T (%v), want actual MsgJoinOK", got, got)
		}
	case "wrong":
		message, ok := got.(*noxnet.MsgServerError)
		if !ok || message.Err != noxnet.ErrWrongPassword {
			t.Errorf("authentication result = %T (%v), want ErrWrongPassword", got, got)
		}
	case "ping":
		if got != nil {
			t.Errorf("authentication result = %T (%v), want nil to continue ping admission", got, got)
		}
	default:
		t.Fatal("unknown literal expected authentication result")
	}
}

func TestNetworkPasswordSettingsLayoutR283(t *testing.T) {
	if unsafe.Offsetof(server.Settings{}.ServerPass78) != 78 || unsafe.Sizeof(server.Settings{}.ServerPass78[0]) != 2 || len(server.Settings{}.ServerPass78) != 9 {
		t.Fatal("server password no longer begins at byte 78 as nine UTF-16 words")
	}
	if unsafe.Sizeof(server.Settings{}) != 168 || unsafe.Offsetof(server.Settings{}.RestrictedPingMind105) != 105 || unsafe.Offsetof(server.Settings{}.RestrictedPingMax107) != 107 {
		t.Fatal("scalar Settings or original signed ping fields changed layout")
	}
	f := newNetworkPasswordFixtureR283(t, "real", -1, -1)
	if unsafe.Pointer(&f.settings.ServerPass78[0]) != unsafe.Add(unsafe.Pointer(f.settings), 78) || binary.LittleEndian.Uint16(f.window[16+39:]) != 'd' || f.settings.ServerPass78[0] != 'r' {
		t.Fatal("fixture does not distinguish old byte offset 39 from UTF-16 word offset 39")
	}
	f.checkUnchanged(t)
}

func TestNetworkPasswordCallbackR283(t *testing.T) {
	cases := []struct {
		name, configured, supplied, want string
		minimum, maximum                 int16
	}{
		{"real-not-decoy", "Secret", "Secret", "join", -1, -1},
		{"decoy-is-not-password", "Secret", "decoy", "wrong", -1, -1},
		{"wrong", "Secret", "other", "wrong", -1, -1},
		{"case-sensitive", "Secret", "secret", "wrong", -1, -1},
		{"empty-real-not-decoy", "", "", "join", -1, -1},
		{"empty-does-not-accept-decoy", "", "decoy", "wrong", -1, -1},
		{"non-ASCII", "비밀鍵", "비밀鍵", "join", -1, -1},
		{"maximum-eight-words", "Ab12Cd34", "Ab12Cd34", "join", -1, -1},
		{"surrogate-pair", "Nox😀鍵", "Nox😀鍵", "join", -1, -1},
		{"eight-words-with-pair", "Ab😀1234", "Ab😀1234", "join", -1, -1},
		{"minimum-zero", "Secret", "Secret", "ping", 0, -1},
		{"maximum-zero", "Secret", "Secret", "ping", -1, 0},
		{"minimum-plus-one", "Secret", "Secret", "ping", 1, -1},
		{"maximum-plus-one", "Secret", "Secret", "ping", -1, 1},
		{"both-plus-one", "Secret", "Secret", "ping", 1, 1},
		{"minimum-minus-two-not-sentinel", "Secret", "Secret", "ping", -2, -1},
		{"maximum-minus-two-not-sentinel", "Secret", "Secret", "ping", -1, -2},
		{"wrong-before-ping", "Secret", "other", "wrong", 1, 1},
		{"decoy-before-ping", "Secret", "decoy", "wrong", 1, -1},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			f := newNetworkPasswordFixtureR283(t, test.configured, test.minimum, test.maximum)
			packet := &noxnet.MsgServerPass{Unk0: 0x72, Pass: test.supplied}
			before := *packet
			got := noxnetCheckPassword(packet)
			checkNetworkPasswordResultR283(t, got, test.want)
			if *packet != before {
				t.Error("authentication mutated the decoded password message")
			}
			f.checkUnchanged(t)
		})
	}
}
