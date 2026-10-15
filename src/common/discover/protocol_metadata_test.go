package discover

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"testing"
	"time"

	wire "github.com/opennox/libs/noxnet/discover"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/lobby"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

var discoveryProtocolCases = []uint32{0x0001039a, 0x000f039a, 0, 0x8001039a, 0xffffffff}

// JSON deliberately exercises the public Server representation without assuming
// how the optional protocol metadata is stored internally.
func requireDiscoveryProtocol(t *testing.T, value any, want uint32) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Protocol *uint32 `json:"protocol_version"`
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Protocol == nil || *got.Protocol != want {
		t.Fatalf("protocol metadata = %s, want known %#08x", data, want)
	}
}

func discoveryProtocolPacket(t *testing.T, version, token uint32) []byte {
	t.Helper()
	m := &wire.MsgServerInfo{
		PlayersCur: 2, PlayersMax: 8, MapName: "WIZ01A",
		ServerName: "WireName", Flags: uint16(noxflags.GameModeArena), Token: token,
	}
	binary.LittleEndian.PutUint32(m.Unk19[3:], version)
	data, err := netmsg.Append([]byte{0, 0}, m)
	if err != nil {
		t.Fatal(err)
	}
	if binary.LittleEndian.Uint32(data[24:28]) != version {
		t.Fatal("fixture protocol is not at original datagram offset 24")
	}
	return data
}

func TestDiscoveryProtocolConversionCopiesExactWireVersion(t *testing.T) {
	for _, version := range discoveryProtocolCases {
		t.Run(fmt.Sprintf("%08x", version), func(t *testing.T) {
			data := discoveryProtocolPacket(t, version, 0x12345678)
			before := bytes.Clone(data)
			m := decodeGameInfo(data)
			if m == nil {
				t.Fatal("valid response rejected")
			}
			g := convGameInfo(netip.MustParseAddrPort("127.0.0.1:18590"), m, data)
			requireDiscoveryProtocol(t, g, version)
			if !bytes.Equal(data, before) {
				t.Fatal("conversion mutated the response")
			}
			clear(data)
			clear(m.Unk19[:])
			requireDiscoveryProtocol(t, g, version)
		})
	}
}

func TestDiscoveryProtocolRejectsTruncatedResponse(t *testing.T) {
	data := discoveryProtocolPacket(t, 0x000f039a, 1)
	for n := 0; n < 73; n++ {
		if decodeGameInfo(data[:n]) != nil {
			t.Fatalf("accepted %d-byte response without full fixed fields", n)
		}
	}
}

// This is a UDP wire fixture, not a running game server or a world simulation.
func discoveryProtocolResponder(t *testing.T, version uint32) (netip.AddrPort, *net.UDPConn) {
	t.Helper()
	remote, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	local, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		remote.Close()
		t.Fatal(err)
	}
	template := discoveryProtocolPacket(t, version, 0)
	done := make(chan error, 1)
	go func() {
		buf := make([]byte, 256)
		_ = remote.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, addr, err := remote.ReadFromUDPAddrPort(buf)
		if err == nil {
			if n != 12 || buf[2] != byte(netmsg.MSG_SERVER_DISCOVER) {
				err = fmt.Errorf("unexpected discovery request: %x", buf[:n])
			} else {
				data := bytes.Clone(template)
				copy(data[44:48], buf[8:12])
				_, err = remote.WriteToUDPAddrPort(data, addr)
			}
		}
		done <- err
	}()
	t.Cleanup(func() {
		local.Close()
		remote.Close()
		if err := <-done; err != nil {
			t.Errorf("UDP fixture: %v", err)
		}
	})
	return remote.LocalAddr().(*net.UDPAddr).AddrPort(), local
}

func discoveryLobbyRecord(addr netip.AddrPort) Server {
	return Server{
		Game: lobby.Game{
			Name: "LobbyName", Address: addr.Addr().String(), Port: int(addr.Port()),
			Map: "estate", Mode: lobby.ModeCTF, Vers: "v1.10.0-dev",
			Res:     lobby.Resolution{HighRes: true, Width: 1920, Height: 1080},
			Players: lobby.PlayersInfo{Cur: 0, Max: 32},
		},
		IP: addr.Addr(), Source: "fixture", Priority: -2,
	}
}

func requireDiscoveryLobbyPreserved(t *testing.T, got Server, want Server, version uint32) {
	t.Helper()
	requireDiscoveryProtocol(t, got, version)
	if got.Name != want.Name || got.Address != want.Address || got.Port != want.Port ||
		got.Map != want.Map || got.Mode != want.Mode || got.Vers != want.Vers ||
		got.Res != want.Res || got.Source != want.Source || got.Priority != want.Priority || got.IP != want.IP {
		t.Fatalf("lobby metadata changed: got %+v, want %+v", got, want)
	}
	if got.Players.Cur != 2 || got.Players.Max != 8 || !got.NoPing || got.Ping <= 0 {
		t.Fatalf("missing live ping fields: %+v", got)
	}
}

func TestDiscoveryProtocolSurvivesAllPingMergePaths(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, api := range []string{"PingAll", "PingEachServer", "ListServersWith"} {
		t.Run(api, func(t *testing.T) {
			for _, version := range discoveryProtocolCases {
				t.Run(fmt.Sprintf("%08x", version), func(t *testing.T) {
					addr, local := discoveryProtocolResponder(t, version)
					original := discoveryLobbyRecord(addr)
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					defer cancel()
					var got []Server
					if api == "PingAll" {
						got = []Server{original}
						if err := PingAll(ctx, log, local, got); err != nil {
							t.Fatal(err)
						}
					} else {
						oldBackends, oldFallback := backends, fallback
						t.Cleanup(func() { backends, fallback = oldBackends, oldFallback })
						backends = map[string]SearchFunc{"fixture": func(ctx context.Context, _ *slog.Logger, out chan<- Server) error {
							out <- original
							if api == "ListServersWith" {
								lessPreferred := original
								lessPreferred.Priority = 1
								lessPreferred.Name = "LowerPriorityName"
								out <- lessPreferred
							}
							return nil
						}}
						fallback = map[string]SearchFunc{}
						var err error
						if api == "PingEachServer" {
							err = PingEachServer(ctx, log, local, func(s Server) error { got = append(got, s); return nil })
						} else {
							got, err = ListServersWith(ctx, log, local)
						}
						if err != nil {
							t.Fatal(err)
						}
					}
					if len(got) != 1 {
						t.Fatalf("got %d servers, want one", len(got))
					}
					requireDiscoveryLobbyPreserved(t, got[0], original, version)
				})
			}
		})
	}
}

func TestDiscoveryProtocolKeepsPublicPingerGameAPI(t *testing.T) {
	for _, api := range []string{"Ping", "SendPing"} {
		t.Run(api, func(t *testing.T) {
			addr, local := discoveryProtocolResponder(t, 0x000f039a)
			p := NewPinger(slog.New(slog.NewTextHandler(io.Discard, nil)), local)
			defer p.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			var game *lobby.Game
			if api == "Ping" {
				var err error
				game, err = p.Ping(ctx, addr)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				out := make(chan *lobby.Game, 1)
				if err := p.SendPing(out, addr); err != nil {
					t.Fatal(err)
				}
				select {
				case game = <-out:
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
			}
			if game.Name != "WireName" || game.Map != "wiz01a" || game.Mode != lobby.ModeArena ||
				game.Players.Cur != 2 || game.Players.Max != 8 || game.Address != addr.Addr().String() || game.Port != int(addr.Port()) ||
				game.Vers != "" || game.Res != (lobby.Resolution{}) {
				t.Fatalf("public game payload changed: %+v", game)
			}
		})
	}
}
