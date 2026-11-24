package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestWOLServerFilterUnsignedPingAndQuestByte(t *testing.T) {
	mode, filter := wolServerFilterMode(0), wolServerFilterRecord(0)
	savedMode, saved := *mode, *filter
	t.Cleanup(func() { *mode, *filter = savedMode, saved })
	*mode = 2
	*filter = wolServerFilter{checkLatency: 1, maxLatency: 100, excludePrivate: 1, questLevel: 20}
	for _, tc := range []struct {
		name   string
		ping   int32
		status server.ServerFlags
		stage  byte
		want   int
	}{
		{"maximum ping", 100, 0, 0, 1},
		{"above maximum", 101, 0, 0, 0},
		{"unknown ping", 9999, 0, 0, 1},
		{"unsigned ping", -1, 0, 0, 0},
		{"private server", 100, server.ServerPrivate, 0, 0},
		{"closed remains eligible", 100, server.ServerClosed, 0, 1},
		{"flagged quest below minimum", 100, 0, 0x80 | 19, 0},
		{"flagged quest at minimum", 100, 0, 0x80 | 20, 1},
		{"unflagged quest byte", 100, 0, 19, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entry := Nox_gui_server_ent_t{PingVal: tc.ping, StatusVal: byte(tc.status), Field_25_2: tc.stage}
			if got := Nox_xxx_checkSomeFlagsOnJoin_4899C0(&entry); got != tc.want {
				t.Fatalf("filter returned %d, want %d", got, tc.want)
			}
		})
	}
}
