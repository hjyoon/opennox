package opennox

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/opennox/libs/console"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
)

var noxTelnet consoleTelnetService
var consoleTelnetOutput consoleTelnetCapture

func init() {
	consoleMux.Add(&consoleTelnetOutput)
	noxConsole.Register(&console.Command{Token: "telnet", HelpID: "telnethelp", Flags: console.Server, Sub: []*console.Command{
		{Token: "on", HelpID: "telnetonhelp", Flags: console.Server, Func: consoleTelnetOn},
		{Token: "off", HelpID: "telnetoffhelp", Flags: console.Server, Func: consoleTelnetOff},
	}})
}

func consoleTelnetOn(_ context.Context, c *console.Console, args []string) bool {
	if len(args) > 1 {
		return false
	}
	if noxflags.HasGame(noxflags.GameModeCoop) {
		return true
	}
	port := uint16(18500) // GAME.EXE 005797F0: default for ports <= 1024.
	if len(args) == 1 {
		v, err := strconv.ParseUint(args[0], 10, 16)
		if err != nil {
			return false
		}
		if v > 1024 {
			port = uint16(v)
		}
	}
	if noxServer == nil || !noxflags.HasGame(noxflags.GameHost) {
		c.Print(console.ColorRed, "No game server is available.")
		return true
	}
	if legacy.Nox_xxx_sysopGetPass_40A630() == "" {
		c.Print(console.ColorRed, "Set a non-empty sysop password before enabling telnet.")
		return true
	}
	prompt := c.Strings().GetStringInFile("Password", "telnetd.c")
	if prompt == "" {
		prompt = "Password"
	}
	bad := c.Strings().GetStringInFile("BadPassword", "telnetd.c")
	if bad == "" {
		bad = "Bad password"
	}
	if err := noxTelnet.start(port, consoleTelnetExecutor(noxServer), prompt, bad); err != nil {
		c.Printf(console.ColorRed, "Cannot start local telnet console: %v", err)
	} else {
		c.Printf(console.ColorRed, "Telnet console: %s (local only; unencrypted)", noxTelnet.address())
	}
	return true
}

func consoleTelnetOff(_ context.Context, _ *console.Console, args []string) bool {
	if len(args) != 0 {
		return false
	}
	if !noxflags.HasGame(noxflags.GameModeCoop) {
		noxTelnet.stop()
	}
	return true
}

// All native state and command execution stays on the game loop. Socket
// goroutines only queue work, receive copied text and perform bounded I/O.
func consoleTelnetExecutor(srv *Server) consoleTelnetExec {
	return func(ctx context.Context, authenticate bool, line string) (bool, []string) {
		type result struct {
			ok    bool
			lines []string
		}
		done := make(chan result, 1)
		srv.QueueInLoop(ctx, func() {
			if ctx.Err() != nil || noxServer != srv || !noxflags.HasGame(noxflags.GameHost) {
				done <- result{}
				return
			}
			if authenticate {
				password := legacy.Nox_xxx_sysopGetPass_40A630()
				done <- result{ok: password != "" && strings.EqualFold(password, line)}
				return
			}
			lines := consoleTelnetOutput.capture(func() { execConsoleCmdAuthed(ctx, line) })
			done <- result{ok: true, lines: lines}
		})
		select {
		case out := <-done:
			return out.ok, out.lines
		case <-ctx.Done():
			return false, nil
		}
	}
}

type consoleTelnetExec func(ctx context.Context, authenticate bool, line string) (bool, []string)
type consoleTelnetService struct {
	mu     sync.Mutex
	active *consoleTelnetRun
}
type consoleTelnetRun struct {
	listener net.Listener
	cancel   context.CancelFunc
	mu       sync.Mutex
	clients  map[net.Conn]struct{}
	wg       sync.WaitGroup
}

func (s *consoleTelnetService) start(port uint16, execute consoleTelnetExec, prompt, badPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active != nil {
		return nil
	}
	listener, err := net.Listen("tcp4", net.JoinHostPort("127.0.0.1", strconv.Itoa(int(port))))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	r := &consoleTelnetRun{listener: listener, cancel: cancel, clients: make(map[net.Conn]struct{})}
	s.active = r
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			r.mu.Lock()
			if ctx.Err() != nil || len(r.clients) >= 4 {
				r.mu.Unlock()
				_ = conn.Close()
				continue
			}
			r.clients[conn] = struct{}{}
			r.wg.Add(1)
			r.mu.Unlock()
			go func() {
				defer r.wg.Done()
				defer conn.Close()
				defer func() { r.mu.Lock(); delete(r.clients, conn); r.mu.Unlock() }()
				consoleTelnetServe(ctx, conn, execute, prompt, badPassword)
			}()
		}
	}()
	return nil
}

func (s *consoleTelnetService) address() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active == nil {
		return ""
	}
	return s.active.listener.Addr().String()
}

// Do not wait on socket goroutines from the game loop: one may be queued in
// QueueInLoop. Cancellation unblocks it, and closes even idle/slow sessions.
func (s *consoleTelnetService) stop() *consoleTelnetRun {
	s.mu.Lock()
	r := s.active
	s.active = nil
	s.mu.Unlock()
	if r != nil {
		r.cancel()
		_ = r.listener.Close()
		r.mu.Lock()
		for conn := range r.clients {
			_ = conn.Close()
		}
		r.mu.Unlock()
	}
	return r
}

func consoleTelnetWrite(conn net.Conn, text string) error {
	if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	_, err := io.WriteString(conn, text)
	return err
}

func consoleTelnetServe(ctx context.Context, conn net.Conn, execute consoleTelnetExec, prompt, badPassword string) {
	// WILL ECHO / WILL SUPPRESS-GO-AHEAD: the server masks password input.
	if consoleTelnetWrite(conn, "\xff\xfb\x01\xff\xfb\x03"+prompt+": ") != nil {
		return
	}
	r := consoleTelnetReader{r: bufio.NewReader(conn), echo: func(s string) error { return consoleTelnetWrite(conn, s) }, password: true}
	for {
		line, err := r.line()
		if err != nil {
			return
		}
		if line == "" {
			continue
		}
		ok, output := execute(ctx, r.password, line)
		if !ok {
			_ = consoleTelnetWrite(conn, "\r\n"+badPassword+"\r\n")
			return
		}
		r.password = false
		if consoleTelnetWrite(conn, "\r\n") != nil {
			return
		}
		for _, text := range output {
			if consoleTelnetWrite(conn, strings.ReplaceAll(text, "\n", "\r\n")+"\r\n") != nil {
				return
			}
		}
		if consoleTelnetWrite(conn, "> ") != nil {
			return
		}
	}
}

type consoleTelnetReader struct {
	r        *bufio.Reader
	echo     func(string) error
	password bool
	skipLF   bool
}

func (r *consoleTelnetReader) line() (string, error) {
	var buf []byte
	for {
		b, err := r.r.ReadByte()
		if err != nil {
			return "", err
		}
		if r.skipLF {
			r.skipLF = false
			if b == '\n' || b == 0 {
				continue
			}
		}
		if b == 255 {
			verb, err := r.r.ReadByte()
			if err != nil {
				return "", err
			}
			if verb >= 251 && verb <= 254 {
				if _, err := r.r.ReadByte(); err != nil {
					return "", err
				}
				continue
			}
			if verb == 250 {
				last := byte(0)
				for n := 0; ; n++ {
					v, err := r.r.ReadByte()
					if err != nil {
						return "", err
					}
					if last == 255 && v == 240 {
						break
					}
					if n >= 1024 {
						return "", errors.New("telnet negotiation exceeds limit")
					}
					last = v
				}
			}
			continue
		}
		if b == '\r' || b == '\n' {
			r.skipLF = b == '\r'
			if !utf8.Valid(buf) || len(utf16.Encode([]rune(string(buf)))) > 127 {
				return "", errors.New("invalid or oversized telnet console line")
			}
			return string(buf), nil
		}
		text := string([]byte{b})
		if b == 8 || b == 127 {
			if len(buf) == 0 {
				continue
			}
			_, size := utf8.DecodeLastRune(buf)
			buf = buf[:len(buf)-size]
			text = "\b \b"
		} else {
			if b < 32 && b != '\t' {
				continue
			}
			if len(buf) >= 512 {
				return "", errors.New("telnet console line exceeds limit")
			}
			buf = append(buf, b)
			if r.password {
				if b >= 0x80 && b < 0xC0 {
					continue
				}
				text = "*"
			}
		}
		if r.echo != nil {
			if err := r.echo(text); err != nil {
				return "", err
			}
		}
	}
}

type consoleTelnetCapture struct {
	mu     sync.Mutex
	active bool
	lines  []string
}

func (p *consoleTelnetCapture) Print(_ console.Color, text string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.active && len(p.lines) < 512 {
		if len(text) > 4096 {
			text = text[:4096]
		}
		p.lines = append(p.lines, text)
	}
}
func (p *consoleTelnetCapture) Printf(cl console.Color, format string, args ...interface{}) {
	p.Print(cl, fmt.Sprintf(format, args...))
}
func (p *consoleTelnetCapture) capture(fn func()) []string {
	p.mu.Lock()
	p.active, p.lines = true, nil
	p.mu.Unlock()
	defer func() { p.mu.Lock(); p.active = false; p.mu.Unlock() }()
	fn()
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.lines...)
}
