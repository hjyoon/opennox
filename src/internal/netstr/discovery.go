package netstr

import (
	"errors"
	"net"
	"net/netip"

	"github.com/opennox/libs/noxnet"
	"github.com/opennox/libs/noxnet/netmsg"
)

type LobbyWaitOptions struct {
	OnResult       func(addr netip.AddrPort, data []byte)
	OnPassRequired func()
	OnPing         func(addr netip.AddrPort, data []byte)
	OnConnectErr   func(errcode noxnet.ConnectError) bool
	OnJoinOK       func()
	OnJoinFail     func()
}

func WaitForLobbyResults(conn net.PacketConn, srvAddr netip.Addr, flag RecvFlags, opts LobbyWaitOptions) (int, error) {
	argp := 0
	if flag.Has(RecvCanRead) {
		var err error
		argp, err = canReadConn(DebugSockets, Log, conn)
		if err != nil {
			return 0, err
		} else if argp == 0 {
			// TODO: is it an error at all?
			return 0, errors.New("nothing to read")
		}
	} else {
		argp = 1
	}
	buf := make([]byte, 256)
	for {
		buf = buf[:cap(buf)]
		n, from, err := readFrom(DebugSockets, Log, conn, buf)
		if err != nil {
			return 0, err
		}
		buf = buf[:n]
		// A callback must never consume fields beyond this datagram, even
		// though the reused receive buffer has capacity for a larger reply.
		if len(buf) >= 3 {
			op := netmsg.Op(buf[2])
			if op < netmsg.MSG_CLIENT_ACCEPT {
				if op == netmsg.MSG_SERVER_INFO || srvAddr == from.Addr() {
					switch op {
					case netmsg.MSG_SERVER_INFO:
						// The fixed header ends where the server name begins.
						if len(buf) >= 72 && from.Addr().IsValid() {
							opts.OnResult(from, buf)
						}
					case netmsg.MSG_PASSWORD_REQUIRED:
						opts.OnPassRequired()
					case netmsg.MSG_SERVER_PING:
						if len(buf) >= 8 {
							opts.OnPing(from, buf)
						}
					case netmsg.MSG_SERVER_ERROR:
						if len(buf) >= 4 && !opts.OnConnectErr(noxnet.ConnectError(buf[3])) {
							break
						}
					case netmsg.MSG_SERVER_JOIN_OK:
						opts.OnJoinOK()
					case netmsg.MSG_SERVER_JOIN_FAIL:
						opts.OnJoinFail()
					}
				}
			}
		}
		if flag&RecvCanRead == 0 || (flag&RecvJustOne) != 0 {
			return n, nil
		}
		argp, err = canReadConn(DebugSockets, Log, conn)
		if err != nil {
			return n, err
		} else if argp == 0 {
			return n, nil
		}
	}
}
