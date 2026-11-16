package legacy

import "github.com/opennox/opennox/v1/server"

// game4TradeFindItem50FFE0 returns the first item with the original full32bit
// network code. This leaf retains the original nonnil-item contract; the
// server's separate uint16 wire lookup also validates its session/items.
func game4TradeFindItem50FFE0(head *server.TradeItem, code uint32) *server.TradeItem {
	for node := head; node != nil; node = node.Field8 {
		if node.Item0.NetCode == code {
			return node
		}
	}
	return nil
}
