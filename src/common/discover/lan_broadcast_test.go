package discover

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"reflect"
	"testing"
	"time"
)

type lanWriteCall struct {
	data []byte
	addr string
}

type lanTestWriter struct {
	calls []lanWriteCall
	write func([]byte, net.Addr) (int, error)
}

func (w *lanTestWriter) WriteTo(data []byte, addr net.Addr) (int, error) {
	w.calls = append(w.calls, lanWriteCall{append([]byte(nil), data...), addr.String()})
	return w.write(data, addr)
}

func TestLANIPv4Broadcast(t *testing.T) {
	for _, c := range []struct {
		name, cidr, want string
	}{
		{"private24", "10.130.27.14/24", "10.130.27.255"},
		{"private8", "10.1.2.3/8", "10.255.255.255"},
		{"private16", "172.16.3.4/16", "172.16.255.255"},
		{"private30", "192.168.5.1/30", "192.168.5.3"},
		{"private23", "192.168.4.8/23", "192.168.5.255"},
		{"point_to_point", "192.168.1.2/31", ""},
		{"single_host", "192.168.1.2/32", ""},
		{"limited", "192.168.1.2/0", ""},
		{"loopback", "127.0.0.1/8", ""},
		{"unspecified", "0.0.0.0/24", ""},
		{"multicast", "224.1.2.3/24", ""},
		{"ipv6", "2001:db8::1/64", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			ip, subnet, err := net.ParseCIDR(c.cidr)
			if err != nil {
				t.Fatal(err)
			}
			subnet.IP = ip
			beforeIP, beforeMask := append(net.IP(nil), subnet.IP...), append(net.IPMask(nil), subnet.Mask...)
			got := lanIPv4Broadcast(subnet)
			if c.want == "" && got != nil || c.want != "" && (got == nil || got.String() != c.want) {
				t.Fatalf("broadcast=%v want=%q", got, c.want)
			}
			if !bytes.Equal(beforeIP, subnet.IP) || !bytes.Equal(beforeMask, subnet.Mask) {
				t.Fatal("interface address/mask mutated")
			}
		})
	}
	t.Run("nil", func(t *testing.T) {
		if lanIPv4Broadcast(nil) != nil {
			t.Fatal("nil address acquired a broadcast")
		}
	})
	t.Run("noncontiguous_mask", func(t *testing.T) {
		if lanIPv4Broadcast(&net.IPNet{IP: net.IPv4(10, 1, 2, 3), Mask: net.IPMask{255, 0, 255, 0}}) != nil {
			t.Fatal("invalid mask acquired a broadcast")
		}
	})
}

func TestLANDiscoveryLimitedBroadcastUnchanged(t *testing.T) {
	data := encodeGameDiscovery(0x52323034)
	before := append([]byte(nil), data...)
	w := &lanTestWriter{write: func(p []byte, addr net.Addr) (int, error) { return len(p), nil }}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	err := sendLANDiscovery(context.Background(), log, w, data, 18590, func(*slog.Logger) []net.IP { t.Fatal("successful limited broadcast enumerated interfaces"); return nil })
	if err != nil || len(w.calls) != 1 || w.calls[0].addr != "255.255.255.255:18590" || !bytes.Equal(w.calls[0].data, before) || !bytes.Equal(data, before) {
		t.Fatalf("ordinary send changed: err=%v calls=%+v", err, w.calls)
	}
}

func TestLANDiscoveryFailedBroadcastFallback(t *testing.T) {
	original := errors.New("limited broadcast unavailable")
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, c := range []struct {
		name    string
		targets []net.IP
		success string
		want    []string
	}{
		{"one", []net.IP{net.IPv4(10, 130, 27, 255)}, "10.130.27.255:18590", []string{"255.255.255.255:18590", "10.130.27.255:18590"}},
		{"duplicates", []net.IP{net.IPv4(10, 130, 27, 255), net.IP{10, 130, 27, 255}, nil, net.ParseIP("2001:db8::1"), net.IPv4bcast}, "10.130.27.255:18590", []string{"255.255.255.255:18590", "10.130.27.255:18590"}},
		{"partial", []net.IP{net.IPv4(10, 130, 27, 255), net.IPv4(192, 168, 1, 255)}, "192.168.1.255:18590", []string{"255.255.255.255:18590", "10.130.27.255:18590", "192.168.1.255:18590"}},
		{"all_failed", []net.IP{net.IPv4(10, 130, 27, 255)}, "", []string{"255.255.255.255:18590", "10.130.27.255:18590"}},
		{"no_targets", nil, "", []string{"255.255.255.255:18590"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			data := encodeGameDiscovery(0x52323034)
			before := append([]byte(nil), data...)
			w := &lanTestWriter{write: func(p []byte, addr net.Addr) (int, error) {
				if addr.String() == c.success {
					return len(p), nil
				}
				return 0, original
			}}
			enumerations := 0
			err := sendLANDiscovery(context.Background(), log, w, data, 18590, func(*slog.Logger) []net.IP { enumerations++; return c.targets })
			if (c.success != "" && err != nil) || (c.success == "" && err != original) || enumerations != 1 {
				t.Fatalf("fallback err=%v enumerations=%d", err, enumerations)
			}
			var got []string
			for _, call := range w.calls {
				got = append(got, call.addr)
				if !bytes.Equal(call.data, before) {
					t.Fatal("fallback changed wire token/payload")
				}
			}
			if !reflect.DeepEqual(got, c.want) || !bytes.Equal(data, before) {
				t.Fatalf("destinations=%v want=%v", got, c.want)
			}
		})
	}
}

func TestLANDiscoveryCanceledFallbackDoesNotSend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	original := errors.New("limited broadcast unavailable")
	w := &lanTestWriter{write: func([]byte, net.Addr) (int, error) { return 0, original }}
	err := sendLANDiscovery(ctx, slog.Default(), w, encodeGameDiscovery(1), 18590, func(*slog.Logger) []net.IP { t.Fatal("canceled fallback read interfaces"); return nil })
	if err != context.Canceled || len(w.calls) != 1 {
		t.Fatalf("cancel err=%v calls=%d", err, len(w.calls))
	}
}

func TestLANDiscoveryFallbackActualDatagram(t *testing.T) {
	peer, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer peer.Close()
	sender, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer sender.Close()
	port := peer.LocalAddr().(*net.UDPAddr).Port
	data := encodeGameDiscovery(0x52323034)
	limited := errors.New("controlled OS limited-broadcast failure")
	w := &lanTestWriter{write: func(p []byte, addr net.Addr) (int, error) {
		if addr.(*net.UDPAddr).IP.Equal(net.IPv4bcast) {
			return 0, limited
		}
		return sender.WriteTo(p, addr)
	}}
	// The declared OS-address service supplies loopback only for this socket
	// integration fixture; production interface selection excludes loopback.
	err = sendLANDiscovery(context.Background(), slog.New(slog.NewTextHandler(io.Discard, nil)), w, data, port, func(*slog.Logger) []net.IP { return []net.IP{net.IPv4(127, 0, 0, 1)} })
	if err != nil {
		t.Fatal(err)
	}
	if err = peer.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	var packet [256]byte
	n, from, err := peer.ReadFromUDP(packet[:])
	if err != nil || !bytes.Equal(packet[:n], data) || from.Port != sender.LocalAddr().(*net.UDPAddr).Port || len(w.calls) != 2 {
		t.Fatalf("real datagram n=%d from=%v err=%v calls=%d", n, from, err, len(w.calls))
	}
}
