package discover

import (
	"context"
	"log/slog"
	"net"
)

type lanDiscoveryWriter interface {
	WriteTo([]byte, net.Addr) (int, error)
}

// Keep the limited broadcast first. Some macOS interface-scoped routes reject
// it even though a directed broadcast on the same active subnet succeeds.
// Only retry that failed send, using the same socket, token and wire bytes.
func sendLANDiscovery(ctx context.Context, log *slog.Logger, conn lanDiscoveryWriter, data []byte, port int, targets func(*slog.Logger) []net.IP) error {
	_, originalErr := conn.WriteTo(data, &net.UDPAddr{IP: net.IPv4(255, 255, 255, 255), Port: port})
	if originalErr == nil {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	sent := false
	seen := make(map[string]bool)
	for _, ip := range targets(log) {
		if err := ctx.Err(); err != nil {
			return err
		}
		v4 := ip.To4()
		if v4 == nil || v4.Equal(net.IPv4bcast) || seen[v4.String()] {
			continue
		}
		seen[v4.String()] = true
		addr := &net.UDPAddr{IP: v4, Port: port}
		if _, err := conn.WriteTo(data, addr); err != nil {
			log.Warn("subnet discovery broadcast failed", "addr", addr, "err", err)
			continue
		}
		log.Info("using subnet discovery broadcast", "addr", addr)
		sent = true
	}
	if sent {
		return nil
	}
	return originalErr
}

func lanBroadcastTargets(log *slog.Logger) []net.IP {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Warn("cannot enumerate LAN broadcast interfaces", "err", err)
		return nil
	}
	var out []net.IP
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagBroadcast == 0 || iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagPointToPoint != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Warn("cannot read LAN interface addresses", "interface", iface.Name, "err", err)
			continue
		}
		for _, addr := range addrs {
			if subnet, ok := addr.(*net.IPNet); ok {
				if ip := lanIPv4Broadcast(subnet); ip != nil {
					out = append(out, ip)
				}
			}
		}
	}
	return out
}

func lanIPv4Broadcast(subnet *net.IPNet) net.IP {
	if subnet == nil {
		return nil
	}
	ip := subnet.IP.To4()
	ones, bits := subnet.Mask.Size()
	// /31 and /32 have no broadcast host; /0 is the limited broadcast.
	if ip == nil || bits != 32 || ones <= 0 || ones >= 31 || ip.IsLoopback() || ip.IsUnspecified() || ip.IsMulticast() {
		return nil
	}
	out := make(net.IP, net.IPv4len)
	for i := range out {
		out[i] = ip[i] | ^subnet.Mask[i]
	}
	return out
}
