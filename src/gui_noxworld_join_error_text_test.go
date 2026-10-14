package opennox

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Exercise the production C dispatcher and C-to-Go string lookup, not just the
// dialog order. Existing join tests run without the relocated PE32 string blob.
func TestNativeNoxWorldJoinErrorText438BD0(t *testing.T) {
	const child = "OPENNOX_TEST_JOIN_ERROR_TEXT_438BD0"
	if os.Getenv(child) == "" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestNativeNoxWorldJoinErrorText438BD0$", "-test.v")
		cmd.Env = append(os.Environ(), child+"=true")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("native join error text failed: %v\n%s", err, out)
		}
		if !bytes.Contains(out, []byte("NATIVE JOIN ERROR TEXT PASS")) {
			t.Fatalf("missing complete native dispatch receipt:\n%s", out)
		}
		return
	}
	strings := strman.New()
	s := server.New(nil, nil, strings)
	t.Cleanup(s.Close)
	noxServer = &Server{Server: s}
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	noxClient = &Client{Client: &client.Client{Server: s, GUI: g}}
	gui.MainBg = g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
	g.NewWindowRaw(gui.MainBg, gui.StatusEnabled, 0, 0, 10, 10, nil).SetID(99)
	var events []string
	legacy.Sub_44A4A0 = func() int { return 1 }
	legacy.Sub_449E00 = func(text string) int { events = append(events, "title:"+text); return 1 }
	legacy.Sub_449E30 = func(text string) int { events = append(events, "text:"+text); return 1 }
	legacy.Sub_449EA0 = func(flags gui.DialogFlags) { events = append(events, fmt.Sprintf("buttons:%d", flags)) }
	legacy.Sub_44A360 = func(size int) { events = append(events, fmt.Sprintf("font:%d", size)) }
	legacy.PlatformTicks = func() uint64 { return 5000 }
	// Exact sealed GAME.EXE 59C578 entries, ordered by the protocol error code.
	ids := []string{"ErrLowPing", "ErrHighPing", "ErrLowLevel", "ErrHighLevel", "ErrClosed", "ErrBanned", "ErrWrongPassword", "ErrIllegalClass", "ErrTimeOut", "ErrFindFailed", "ErrNeedRefresh", "ErrFull", "ErrDupSerial", "ErrWrongVer"}
	check := func(t *testing.T, code int, id string) {
		events = nil
		startNoxWorldJoin()
		legacy.Set_nox_client_connError_814552(code)
		legacy.Sub_43AF90(2)
		if !s.Update2() || legacy.Sub_43AF80() != 1 || legacy.Sub_43B6D0() != 0 {
			t.Fatal("error dispatch did not return to idle and clear the join gate")
		}
		want := []string{}
		if code != 8 && code != 9 && code != 10 {
			want = append(want, "title:MISSING:'noxworld.c:ConnError'")
		}
		want = append(want, "text:MISSING:'noxworld.c:"+id+"'", "buttons:1", "font:1")
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("original error key/order for code %d: got=%q want=%q", code, events, want)
		}
	}
	for code, id := range ids {
		t.Run(id, func(t *testing.T) { check(t, code, id) })
	}
	for _, code := range []int{14, 255, -1} {
		t.Run(fmt.Sprintf("invalid-%d", code), func(t *testing.T) { check(t, code, "ConnError") })
	}
	t.Log("NATIVE JOIN ERROR TEXT PASS")
}
