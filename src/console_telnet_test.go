package opennox

import (
	"bufio"
	"context"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/strman"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
)

func TestConsoleTelnetLineBoundaries(t *testing.T) {
	r := consoleTelnetReader{r: bufio.NewReader(strings.NewReader("\xff\xfd\x01\xff\xfa\x18x\xff\xf0SET NAME 한글\r\nabc\x08D\r\x00next\n"))}
	for _, want := range []string{"SET NAME 한글", "abD", "next"} {
		got, err := r.line()
		if err != nil || got != want {
			t.Fatalf("got %q, %v; want %q", got, err, want)
		}
	}
	for _, input := range []string{strings.Repeat("x", 128) + "\r", strings.Repeat("🐺", 64) + "\r", "bad\xc0\r", strings.Repeat("x", 513), "\xff\xfa" + strings.Repeat("x", 1025)} {
		r := consoleTelnetReader{r: bufio.NewReader(strings.NewReader(input))}
		if _, err := r.line(); err == nil {
			t.Fatal("invalid/oversized line accepted")
		}
	}
	var echo strings.Builder
	r = consoleTelnetReader{r: bufio.NewReader(strings.NewReader("Secret\r")), password: true, echo: func(s string) error { _, err := echo.WriteString(s); return err }}
	if line, err := r.line(); err != nil || line != "Secret" || echo.String() != "******" {
		t.Fatal("password was not masked")
	}
}

func consoleTelnetTestStop(t *testing.T, s *consoleTelnetService) {
	t.Helper()
	if run := s.stop(); run != nil {
		done := make(chan struct{})
		go func() { run.wg.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Error("telnet stop left a blocked socket/queue worker")
		}
	}
}

func consoleTelnetTestReadUntil(t *testing.T, conn net.Conn, r *bufio.Reader, suffix string) string {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var out strings.Builder
	for !strings.HasSuffix(out.String(), suffix) {
		b, err := r.ReadByte()
		if err != nil {
			t.Fatalf("telnet response before %q: %v", suffix, err)
		}
		out.WriteByte(b)
		if out.Len() > 16384 {
			t.Fatal("unbounded telnet response")
		}
	}
	return out.String()
}

func TestConsoleTelnetLoopbackAuthenticationAndCommand(t *testing.T) {
	var s consoleTelnetService
	t.Cleanup(func() { consoleTelnetTestStop(t, &s) })
	commands := make(chan string, 4)
	if err := s.start(0, func(_ context.Context, auth bool, line string) (bool, []string) {
		if auth {
			return line == "secret", nil
		}
		commands <- line
		return true, []string{"result: " + line}
	}, "Password", "Bad password"); err != nil {
		t.Fatal(err)
	}
	host, _, err := net.SplitHostPort(s.address())
	if err != nil || host != "127.0.0.1" {
		t.Fatal("legacy telnet exposed outside loopback")
	}
	bad, err := net.Dial("tcp4", s.address())
	if err != nil {
		t.Fatal(err)
	}
	defer bad.Close()
	br := bufio.NewReader(bad)
	consoleTelnetTestReadUntil(t, bad, br, "Password: ")
	_, _ = io.WriteString(bad, "wrong\r\n")
	response, err := io.ReadAll(br)
	if err != nil || !strings.Contains(string(response), "Bad password") || strings.Contains(string(response), "wrong") {
		t.Fatal("bad login was not safely rejected")
	}
	if len(commands) != 0 {
		t.Fatal("unauthenticated command execution")
	}
	conn, err := net.Dial("tcp4", s.address())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	consoleTelnetTestReadUntil(t, conn, reader, "Password: ")
	_, _ = io.WriteString(conn, "\xff\xfd\x01secret\r\n")
	consoleTelnetTestReadUntil(t, conn, reader, "> ")
	_, _ = io.WriteString(conn, "SET NAME 한글\r\n")
	responseText := consoleTelnetTestReadUntil(t, conn, reader, "> ")
	if !strings.Contains(responseText, "result: SET NAME 한글") {
		t.Fatal("command result was not returned")
	}
	if got := <-commands; got != "SET NAME 한글" {
		t.Fatalf("command input changed: %q", got)
	}
	consoleTelnetTestStop(t, &s)
	if s.address() != "" {
		t.Fatal("off left the listener registered")
	}
}

func TestConsoleTelnetConnectionLimitAndCancellation(t *testing.T) {
	var s consoleTelnetService
	t.Cleanup(func() { consoleTelnetTestStop(t, &s) })
	queued := make(chan struct{}, 1)
	if err := s.start(0, func(ctx context.Context, _ bool, _ string) (bool, []string) {
		queued <- struct{}{}
		<-ctx.Done()
		return false, nil
	}, "Password", "Bad password"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		conn, err := net.Dial("tcp4", s.address())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.Close() })
		consoleTelnetTestReadUntil(t, conn, bufio.NewReader(conn), "Password: ")
		if i == 0 {
			_, _ = io.WriteString(conn, "pending\r\n")
		}
	}
	select {
	case <-queued:
	case <-time.After(3 * time.Second):
		t.Fatal("queue fixture never started")
	}
	extra, err := net.Dial("tcp4", s.address())
	if err != nil {
		t.Fatal(err)
	}
	defer extra.Close()
	_ = extra.SetReadDeadline(time.Now().Add(3 * time.Second))
	if _, err := extra.Read(make([]byte, 1)); err == nil {
		t.Fatal("fifth connection was accepted")
	} else if timeout, ok := err.(net.Error); ok && timeout.Timeout() {
		t.Fatal("fifth connection was left open")
	}
	consoleTelnetTestStop(t, &s)
}

func TestConsoleTelnetNativeLoopAndPermissionGuards(t *testing.T) {
	consoleCommandTestFlags(t)
	srv := NewServer(nil, nil, strman.New())
	oldServer, oldConsole := noxServer, noxConsole
	oldPassword := legacy.Nox_xxx_sysopGetPass_40A630()
	noxServer = srv
	noxflags.SetGame(noxflags.GameHost)
	noxConsole = console.New(&consoleTelnetOutput)
	noxConsole.Localize(strman.New())
	t.Cleanup(func() {
		legacy.Nox_xxx_sysopSetPass_40A610(oldPassword)
		noxServer, noxConsole = oldServer, oldConsole
		srv.Close()
	})
	calls := 0
	noxConsole.Register(&console.Command{Token: "probe", Flags: console.Server | console.Cheat, Func: func(_ context.Context, c *console.Console, _ []string) bool {
		calls++
		c.Print(console.ColorRed, "native result")
		return true
	}})
	execute := consoleTelnetExecutor(srv)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	run := func(auth bool, line string) (bool, []string) {
		type response struct {
			ok    bool
			lines []string
		}
		done := make(chan response, 1)
		go func() { ok, lines := execute(ctx, auth, line); done <- response{ok, lines} }()
		for {
			select {
			case result := <-done:
				return result.ok, result.lines
			case <-ctx.Done():
				t.Fatal("native queue did not complete")
				return false, nil
			default:
				srv.RunLoopHooks()
			}
		}
	}
	legacy.Nox_xxx_sysopSetPass_40A610("")
	if ok, _ := run(true, ""); ok {
		t.Fatal("empty sysop password authenticated")
	}
	legacy.Nox_xxx_sysopSetPass_40A610("SeCrEt")
	if ok, _ := run(true, "secret"); !ok {
		t.Fatal("native password/case-insensitive authentication failed")
	}
	if ok, _ := run(true, "wrong"); ok {
		t.Fatal("wrong native password authenticated")
	}
	run(false, "PROBE")
	if calls != 0 {
		t.Fatal("telnet bypassed the normal cheat gate")
	}
	noxConsole.SetCheats(true)
	if ok, lines := run(false, "PROBE"); !ok || calls != 1 || len(lines) != 1 || lines[0] != "native result" {
		t.Fatal("native queued execution/output failed")
	}
	noxflags.UnsetGame(noxflags.GameHost)
	if ok, _ := run(false, "PROBE"); ok || calls != 1 {
		t.Fatal("stopped server still executes remote commands")
	}
}
