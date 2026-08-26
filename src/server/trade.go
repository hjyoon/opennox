package server

import (
	"encoding/binary"
	"strings"
	"unicode/utf16"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const ShopStartPacketSize50F0F0 = 86

type TradeSession struct {
	Field0  uint32        // 0, 0
	Field4  uint32        // 1, 4
	Field8  *Object       // 2, 8
	Field12 *Object       // 3, 12
	Field16 uint32        // 4, 16
	Field20 *TradeItem    // 5, 20
	Field24 uint32        // 6, 24
	Field28 uint32        // 7, 28
	Field32 *TradeItem    // 8, 32
	Field36 *TradeItem    // 9, 36
	Field40 uint32        // 10, 40
	Field44 uint32        // 11, 44
	Field48 *Object       // 12, 48
	Field52 *Object       // 13, 52
	Field56 *TradeSession // 14, 56
	Field60 *TradeSession // 15, 60
}

type TradeItem struct {
	Item0   *Object    // 0, 0
	Cost4   uint32     // 1, 4
	Field8  *TradeItem // 2, 8
	Field12 *TradeItem // 3, 12
}

type serverTradeNativeState struct {
	sessions map[*TradeSession]func()
}

func (t *serverTradeNativeState) init() {
	if t.sessions == nil {
		t.sessions = make(map[*TradeSession]func())
	}
}

func (t *serverTradeNativeState) close() {
	for session, free := range t.sessions {
		delete(t.sessions, session)
		free()
	}
}

// NewShopSessionNative50E8F0 allocates a pointer-width-safe shop session on
// the C heap. The original PE32 pool is fixed at 64 bytes, while the same
// typed structure grows with native pointers on 64-bit targets.
func (s *Server) NewShopSessionNative50E8F0(player, merchant *Object) *TradeSession {
	s.tradeNative.init()
	session, free := alloc.New(TradeSession{})
	session.Field8 = player
	session.Field12 = merchant
	session.Field16 = 1
	s.tradeNative.sessions[session] = free
	return session
}

// IsTradeSessionNative reports whether session is owned by the native-width
// allocator rather than the original PE32 pool.
func (s *Server) IsTradeSessionNative(session *TradeSession) bool {
	if session == nil || s.tradeNative.sessions == nil {
		return false
	}
	_, ok := s.tradeNative.sessions[session]
	return ok
}

// ReleaseTradeSessionNative510000 releases only sessions allocated by the
// native-width trade subsystem. It is deliberately safe for legacy sessions.
func (s *Server) ReleaseTradeSessionNative510000(session *TradeSession) bool {
	if session == nil || s.tradeNative.sessions == nil {
		return false
	}
	free, ok := s.tradeNative.sessions[session]
	if !ok {
		return false
	}
	delete(s.tradeNative.sessions, session)
	free()
	return true
}

func shopObjectNameKey4E39F0(objectID, typeID string) string {
	id := objectID
	if id == "" {
		id = typeID
	}
	if i := strings.LastIndexByte(id, ':'); i >= 0 {
		id = id[i+1:]
	}
	return "NPC:" + strings.ReplaceAll(id, "_", "")
}

// ShopObjectName4E39F0 reproduces the original object database display-name
// lookup without passing a native-width Object pointer through the PE32 helper.
func (s *Server) ShopObjectName4E39F0(obj *Object) string {
	if obj == nil {
		return ""
	}
	typeID := ""
	if typ := s.Types.ByInd(int(obj.TypeInd)); typ != nil {
		typeID = typ.ID()
	}
	key := shopObjectNameKey4E39F0(obj.ID(), typeID)
	return s.Strings().GetStringInFile(strman.ID(key), "C:\\NoxPost\\src\\Server\\DBase\\objdb.c")
}

// BuildShopStartPacket50F0F0 constructs the original 86-byte C9/0D packet:
// merchant type, 24 UTF-16LE name code units plus terminator, and a bounded
// 32-byte shop-text C string.
func BuildShopStartPacket50F0F0(typeInd uint16, name string, shopText []byte) [ShopStartPacketSize50F0F0]byte {
	var packet [ShopStartPacketSize50F0F0]byte
	packet[0] = byte(netmsg.MSG_TRADE)
	packet[1] = 0x0d
	binary.LittleEndian.PutUint16(packet[2:4], typeInd)
	name16 := utf16.Encode([]rune(name))
	if len(name16) > 24 {
		name16 = name16[:24]
	}
	for i, ch := range name16 {
		binary.LittleEndian.PutUint16(packet[4+2*i:6+2*i], ch)
	}
	// packet[52:54] remains the explicit UTF-16 terminator.
	for i, ch := range shopText {
		if i >= 31 || ch == 0 {
			break
		}
		packet[54+i] = ch
	}
	return packet
}
