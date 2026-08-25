package opennox

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"image"
	"log/slog"
	"net"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/opennox/libs/client/keybind"
	noxlog "github.com/opennox/libs/log"
	"github.com/opennox/libs/noxnet"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/strman"
	"github.com/opennox/lobby"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/discover"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netstr"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

var (
	lobbyBroadcast      *net.UDPConn
	discoverDone        = make(chan []discover.Server, 1)
	dword_5d4594_815060 int
	winNoxWorld         *gui.Window
	noxWorldNative      *noxWorldNativeState
)

const (
	noxWorldMaxServers  = 2500
	noxWorldStringsFile = "C:\\NoxPost\\src\\client\\shell\\noxworld.c"
)

type noxWorldDiscoverFunc func(context.Context, *slog.Logger) ([]discover.Server, error)

type noxWorldDiscoveryResult struct {
	generation uint64
	servers    []discover.Server
	err        error
}

type noxWorldServerRow struct {
	server      discover.Server
	name        string
	players     string
	mode        string
	ping        string
	pingValue   int64
	status      string
	statusValue int
}

type noxWorldNativeState struct {
	root       *gui.Window
	status     *gui.Window
	list       *gui.Window
	columns    [5]*gui.Window
	slider     *gui.Window
	up         *gui.Window
	down       *gui.Window
	rows       []noxWorldServerRow
	sorting    int
	generation uint64
	results    chan noxWorldDiscoveryResult
	cancel     context.CancelFunc
	discover   noxWorldDiscoverFunc
}

func noxWorldLocalizedString(id string) string {
	return noxClient.Strings().GetStringInFile(strman.ID(id), noxWorldStringsFile)
}

func noxWorldGameModeString(mode lobby.GameMode, localize func(string) string) string {
	// GAME.EXE nox_gui_wol_gameModeString_43BCB0 tests these mode bits in
	// this order. In the modern lobby schema, elimination is the original
	// Highlander mode and coop/custom fall through to Arena, just as an
	// unrecognized legacy flag did.
	var id string
	switch mode {
	case lobby.ModeQuest:
		id = "Quest"
	case lobby.ModeCTF:
		id = "CTF"
	case lobby.ModeElimination:
		id = "Highlander"
	case lobby.ModeKOTR:
		id = "KotR"
	case lobby.ModeFlagBall:
		id = "Flagball"
	case lobby.ModeChat:
		id = "Chat"
	default:
		id = "Arena"
	}
	return localize(id)
}

func noxWorldTruncateServerName(s string) string {
	// The original result record stores 15 bytes plus a NUL terminator.
	// Keep that visible limit without leaving a split UTF-8 sequence behind.
	if len(s) <= 15 {
		return s
	}
	s = s[:15]
	for !utf8.ValidString(s) {
		s = s[:len(s)-1]
	}
	return s
}

func noxWorldServerAddress(s discover.Server) string {
	if s.Address != "" {
		return s.Address
	}
	if s.IP.IsValid() {
		return s.IP.String()
	}
	return ""
}

func noxWorldServerRows(servers []discover.Server, localize func(string) string) []noxWorldServerRow {
	rows := make([]noxWorldServerRow, 0, min(len(servers), noxWorldMaxServers))
	seen := make(map[string]struct{}, min(len(servers), noxWorldMaxServers))
	for _, srv := range servers {
		addr := noxWorldServerAddress(srv)
		// GAME.EXE clientOnLobbyServer rejects entries without an address.
		if addr == "" {
			continue
		}
		key := addr + "\x00" + strconv.Itoa(srv.Port)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		name := srv.Name
		if name == "" {
			name = fmt.Sprintf("%s:%d", addr, srv.Port)
		}
		name = noxWorldTruncateServerName(name)

		mode := noxWorldGameModeString(srv.Mode, localize)
		if srv.Mode == lobby.ModeQuest && srv.Quest != nil {
			mode = fmt.Sprintf("%s %d", mode, srv.Quest.Stage)
		}

		pingValue := int64(9999)
		ping := "--"
		if srv.Ping > 0 {
			pingValue = srv.Ping.Milliseconds()
			ping = strconv.FormatInt(pingValue, 10)
		}

		statusValue := 0
		var status string
		switch srv.Access {
		case lobby.AccessPassword:
			statusValue = 0x20
			status = localize("Noxworld.wnd:private")
		case lobby.AccessClosed:
			statusValue = 0x10
			status = localize("Noxworld.wnd:closed")
		default:
			if srv.Players.Cur < srv.Players.Max {
				status = localize("Open")
			} else {
				status = localize("Full")
			}
		}

		rows = append(rows, noxWorldServerRow{
			server:      srv,
			name:        name,
			players:     fmt.Sprintf("%d/%d", srv.Players.Cur, srv.Players.Max),
			mode:        mode,
			ping:        ping,
			pingValue:   pingValue,
			status:      status,
			statusValue: statusValue,
		})
		if len(rows) == noxWorldMaxServers {
			break
		}
	}
	return rows
}

func noxWorldNextSorting(current, buttonID int) int {
	base := (buttonID - 10047) * 2
	if current == base {
		return base + 1
	}
	return base
}

func noxWorldSortRows(rows []noxWorldServerRow, sorting int) {
	desc := sorting&1 != 0
	kind := sorting / 2
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		var order int
		switch kind {
		case 0:
			order = strings.Compare(strings.ToLower(a.name), strings.ToLower(b.name))
		case 1:
			order = cmp.Compare(a.server.Players.Cur, b.server.Players.Cur)
		case 2:
			order = strings.Compare(a.mode, b.mode)
		case 3:
			order = cmp.Compare(a.pingValue, b.pingValue)
		case 4:
			order = cmp.Compare(a.statusValue, b.statusValue)
		}
		if desc {
			return order > 0
		}
		return order < 0
	})
}

func newNoxWorldNativeState(win *gui.Window) *noxWorldNativeState {
	if win == nil {
		return nil
	}
	st := &noxWorldNativeState{
		root:     win,
		status:   win.ChildByID(10011),
		list:     win.ChildByID(10037),
		sorting:  0,
		results:  make(chan noxWorldDiscoveryResult, 4),
		discover: discover.ListServers,
	}
	for i, id := range [...]uint{10038, 10039, 10040, 10041, 10042} {
		st.columns[i] = win.ChildByID(id)
	}
	st.slider = win.ChildByID(10053)
	st.up = win.ChildByID(10043)
	st.down = win.ChildByID(10044)
	if st.status == nil || st.list == nil || st.slider == nil || st.up == nil || st.down == nil {
		return nil
	}
	for _, col := range st.columns {
		if col == nil {
			return nil
		}
		// Preserve the widget's structural parent, but route notifications to
		// the native-width controller instead of the legacy master list proc.
		col.DrawData().Window = win
		col.Func94(&WindowEvent0x4018{Win: st.up})
		col.Func94(&WindowEvent0x4019{Win: st.down})
		col.Func94(gui.AsWindowEvent(0x401A, uintptr(st.slider.C()), 0))
	}
	st.slider.DrawData().Window = win
	st.up.DrawData().Window = win
	st.down.DrawData().Window = win

	// The native controller opens directly in the original list-mode view.
	for _, id := range [...]uint{10020, 10021, 10033} {
		if child := win.ChildByID(id); child != nil {
			child.Hide()
		}
	}
	st.list.Show()
	if button := win.ChildByID(10004); button != nil {
		button.DrawData().Field0 |= 0x4
	}
	if button := win.ChildByID(10005); button != nil {
		button.DrawData().Field0 &^= 0x4
	}
	st.setStatus(noxWorldLocalizedString("ListJoinServer"))
	return st
}

func (st *noxWorldNativeState) setStatus(text string) {
	if st != nil && st.status != nil {
		st.status.DrawData().SetText(text)
	}
}

func (st *noxWorldNativeState) clearList() {
	if st == nil {
		return
	}
	for _, col := range st.columns {
		col.Func94(gui.AsWindowEvent(0x400F, 0, 0))
	}
}

func noxWorldAddListLine(win *gui.Window, text string) {
	cstr, free := alloc.CString16(text)
	defer free()
	win.Func94(gui.AsWindowEvent(0x400D, uintptr(unsafe.Pointer(cstr)), 4))
}

func (st *noxWorldNativeState) renderRows() {
	if st == nil {
		return
	}
	st.clearList()
	for _, row := range st.rows {
		for i, text := range [...]string{row.name, row.players, row.mode, row.ping, row.status} {
			noxWorldAddListLine(st.columns[i], text)
		}
	}
}

func (st *noxWorldNativeState) startRefresh() {
	if st == nil {
		return
	}
	if st.cancel != nil {
		st.cancel()
	}
	st.generation++
	generation := st.generation
	ctx, cancel := context.WithCancel(context.Background())
	st.cancel = cancel
	st.rows = nil
	st.clearList()
	st.setStatus(noxWorldLocalizedString("Wolchat.c:PleaseWait"))

	results := st.results
	discoverServers := st.discover
	log := noxlog.WithSystem(noxClient.Log, "discover")
	go func() {
		servers, err := discoverServers(ctx, log)
		if ctx.Err() != nil {
			return
		}
		res := noxWorldDiscoveryResult{generation: generation, servers: servers, err: err}
		select {
		case results <- res:
		case <-ctx.Done():
		}
	}()
}

func noxWorldPollRefresh() bool {
	st := noxWorldNative
	if st == nil {
		return true
	}
	for {
		select {
		case res := <-st.results:
			if res.generation != st.generation {
				continue
			}
			if res.err != nil && !isCtxTimeout(res.err) {
				noxlog.WithSystem(noxClient.Log, "discover").Warn("server refresh failed", "err", res.err)
			}
			st.rows = noxWorldServerRows(res.servers, noxWorldLocalizedString)
			noxWorldSortRows(st.rows, st.sorting)
			st.renderRows()
			st.setStatus(noxWorldLocalizedString("ListJoinServer"))
			noxlog.WithSystem(noxClient.Log, "discover").Info("native server list updated", "servers", len(st.rows))
		default:
			return true
		}
	}
}

func closeNoxWorldNativeState() {
	if noxWorldNative != nil && noxWorldNative.cancel != nil {
		noxWorldNative.cancel()
	}
	noxWorldNative = nil
	noxServer.SetUpdateFunc2(nil)
}

// noxGameShowGameSelNative owns the Nox World window through native pointers.
// The original controller stores most child windows in 32-bit integer globals,
// which truncates their addresses on 64-bit hosts before the first frame.
func noxGameShowGameSelNative() int {
	if winNoxWorld != nil && !winNoxWorld.GetFlags().Has(gui.StatusDestroyed) {
		winNoxWorld.ShowModal()
		gui.SetAnimGlobalState(gui.AnimInDone)
		return 1
	}

	win := nox_new_window_from_file("noxworld.wnd", noxWorldWindowProc)
	if win == nil {
		return 0
	}
	winNoxWorld = win
	win.SetFunc93(noxWorldWindowProc)
	if win.ShowModal() != 0 {
		win.Destroy()
		winNoxWorld = nil
		return 0
	}
	noxWorldNative = newNoxWorldNativeState(win)
	if noxWorldNative == nil {
		noxClient.Log.Error("cannot initialize native Nox World controls")
		win.Destroy()
		winNoxWorld = nil
		return 0
	}
	noxServer.SetUpdateFunc2(noxWorldPollRefresh)

	noxClient.GameAddStateCode(client.StateServerList)
	sub4A24C0(true)
	sub_4A1BE0(0)
	gui.SetAnimGlobalState(gui.AnimInDone)
	noxWorldNative.startRefresh()
	return 1
}

func noxWorldWindowProc(_ *gui.Window, ev gui.WindowEvent) gui.WindowEventResp {
	switch ev := ev.(type) {
	case *WindowEvent0x4005:
		clientPlaySoundSpecial(sound.SoundShellSelect, 100)
		return gui.RawEventResp(1)
	case *WindowEvent0x4007:
		if ev.Win == nil {
			return nil
		}
		switch ev.Win.ID() {
		case 10010: // Back
			closeNoxWorldToMainMenu()
		case 10006: // Refresh
			if noxWorldNative != nil {
				noxWorldNative.startRefresh()
			}
			clientPlaySoundSpecial(sound.SoundShellClick, 100)
		case 10043, 10044: // shared list up/down buttons
			if noxWorldNative != nil {
				for _, col := range noxWorldNative.columns {
					col.Func94(ev)
				}
			}
		case 10047, 10048, 10049, 10050, 10051:
			if noxWorldNative != nil {
				noxWorldNative.sorting = noxWorldNextSorting(noxWorldNative.sorting, int(ev.Win.ID()))
				noxWorldSortRows(noxWorldNative.rows, noxWorldNative.sorting)
				noxWorldNative.renderRows()
			}
			clientPlaySoundSpecial(sound.SoundShellClick, 100)
		default:
			clientPlaySoundSpecial(sound.SoundShellClick, 100)
		}
		return gui.RawEventResp(1)
	case *WindowEvent0x4009:
		if noxWorldNative != nil {
			for _, col := range noxWorldNative.columns {
				col.Func94(ev)
			}
		}
		return gui.RawEventResp(1)
	case *WindowEvent0x4010:
		if noxWorldNative != nil {
			selection := gui.AsWindowEvent(0x4013, ev.Val, 0)
			for _, col := range noxWorldNative.columns {
				col.Func94(selection)
			}
		}
		return gui.RawEventResp(1)
	case gui.WindowKeyPress:
		if ev.Key == keybind.KeyEsc && ev.Pressed {
			closeNoxWorldToMainMenu()
			return gui.RawEventResp(1)
		}
	}
	return nil
}

func closeNoxWorldToMainMenu() {
	closeNoxWorldNativeState()
	if winNoxWorld != nil {
		winNoxWorld.Destroy()
		winNoxWorld = nil
	}
	if noxClient.GameGetStateCode() == client.StateServerList {
		noxClient.GamePopState()
	}
	sub4A24C0(false)
	sub_4A1BE0(1)
	gui.SetAnimGlobalState(gui.AnimInDone)
	nox_game_showMainMenu_4A1C00()
	clientPlaySoundSpecial(sound.SoundShellClick, 100)
}

type LobbyServerInfo struct {
	discover.Server
	Status     byte
	Flags      uint16 // TODO: should be the same as GameFlags
	Level      uint16
	Field_11_0 int16
	Field_11_2 int16
	Version    uint32
	Field_25_1 byte
	Field_25_2 byte
	Field_26_1 uint16
	Field_26_3 uint16
	Field_33_3 [20]byte
	Field_38_3 uint32
	Field_39_3 uint32
}

func (s *LobbyServerInfo) String() string {
	addr := fmt.Sprintf("%s:%d", s.IP, s.Port)
	ping := ""
	if s.Ping > 0 {
		ping = fmt.Sprintf(" P:%s,", s.Ping)
	}
	return fmt.Sprintf("{%q, %q (%s), %d/%d,%s F:%v, M:%q, L:%d}", addr, s.Name, s.Source, s.Players.Cur, s.Players.Max, ping, s.Flags, s.Map, s.Level)
}

func onLobbyServerPacket(log *slog.Logger, addr string, port int, name string, packet []byte) bool {
	log = noxlog.WithSystem(log, "discover")
	log.Debug("ignoring server response", "addr", addr, "port", port, "name", name)
	return false
	/*
		ticks := uint64(binary.LittleEndian.Uint32(packet[44:]))
		if exp := *memmap.PtrUint32(0x5D4594, 814964); uint32(ticks) != exp {
			discover.Log.Printf("onLobbyServerPacket: ignoring server %q: invalid ts: 0x%x vs 0x%x", addr, ticks, exp)
			return false
		}
		mi := StrLenBytes(packet[10:])
		info := &LobbyServerInfo{
			Addr:       addr,
			Port:       port,
			Name:       name,
			Players:    int(packet[3]),
			MaxPlayers: int(packet[4]),
			Field_25_1: packet[5] | (packet[6] << 4),
			Field_38_3: binary.LittleEndian.Uint32(packet[7:]) & 0xffffff,
			Map:        string(packet[10 : 10+mi]),
			Field_25_2: packet[19],
			Status:     packet[20] | packet[21],
			Field_12:   binary.LittleEndian.Uint32(packet[24:]),
			GetFlags:      binary.LittleEndian.Uint16(packet[28:]),
			Field_39_3: binary.LittleEndian.Uint32(packet[32:]),
			Field_26_1: binary.LittleEndian.Uint16(packet[36:]),
			Field_26_3: binary.LittleEndian.Uint16(packet[38:]),
			Field_11_0: int16(binary.LittleEndian.Uint16(packet[40:])),
			Field_11_2: int16(binary.LittleEndian.Uint16(packet[42:])),
			Ping:       time.Duration(platformTicks()-ticks) * time.Millisecond,
			Level:      binary.LittleEndian.Uint16(packet[68:]),
		}
		copy(info.Field_33_3[:], packet[48:48+20])
		return onLobbyServer(info)
	*/
}

func nox_client_refreshServerList_4378B0() {
	if sub44A4A0() {
		legacy.Set_dword_5d4594_815104(1)
		return
	}

	*memmap.PtrUint64(0x5D4594, 815076) = platformTicks()
	dword_5d4594_815060 = 0
	legacy.Sub_4379C0()
	legacy.Get_dword_5d4594_815004().Func94(gui.AsWindowEvent(0x400F, 0, 0))
	legacy.Sub_49FFA0(1)
	legacy.Set_nox_wol_server_result_cnt_815088(0)

	ctx := context.Background()
	winNewDialogID(legacy.Get_nox_wol_wnd_world_814980(), "Wolchat.c:PleaseWait", "C:\\NoxPost\\src\\client\\shell\\noxworld.c")
	noxServer.NetStr.Responded = false
	go discoverAndPingServers(ctx, noxClient.Log)
	legacy.Set_dword_5d4594_815104(0)
	legacy.Set_qword_5d4594_815068(
		// next auto-refresh
		*memmap.PtrUint64(0x5D4594, 815076) + 120000)
}

func sub_438770_waitList() {
	if dword_5d4594_815060 != 0 {
		return
	}
	log := noxlog.WithSystem(noxClient.Log, "discover")
	timer := time.NewTimer(10 * time.Millisecond)
	defer timer.Stop()
	var list []discover.Server
	select {
	case <-timer.C:
		return
	case list = <-discoverDone:
	}
	for _, g := range list {
		level := 0
		if g.Quest != nil {
			level = g.Quest.Stage
		}
		var status byte
		switch g.Access {
		case lobby.AccessClosed:
			status |= 0x10
		case lobby.AccessPassword:
			status |= 0x20
		}
		// TODO: should propagate this via lobby
		v := noxProtoVersionLegacy
		if g.Res.HighRes || g.Res.Width > 1024 || g.Res.Height > 768 {
			v = noxProtoVersionHighRes
		}
		clientOnLobbyServer(log, &LobbyServerInfo{
			Server:  g,
			Flags:   gameModeToFlags(g.Mode),
			Status:  status,
			Level:   uint16(level),
			Version: uint32(v),
		})
	}
	sub_44A400()
	legacy.Sub_4379C0()
	legacy.Sub_4A0360()
	dword_5d4594_815060 = 1
}

func nox_xxx_createSocketLocal(port int) error {
	if lobbyBroadcast != nil {
		return nil
	}
	conn, err := netstr.Listen(netip.AddrPortFrom(netip.IPv4Unspecified(), uint16(port)))
	if err != nil {
		netstr.Log.Println("cannot bind broadcast socket:", err)
		return err
	}
	lobbyBroadcast = conn
	noxClient.SetDrawFunc(clientWaitForLobbyResults)
	return nil
}

func sub_554D10() int {
	if lobbyBroadcast != nil {
		_ = lobbyBroadcast.Close()
		lobbyBroadcast = nil
		noxClient.SetDrawFunc(nil)
	}
	return 0
}

func clientOnLobbyServer(log *slog.Logger, info *LobbyServerInfo) int {
	log = noxlog.WithSystem(log, "discover")
	log.Info("server response", "info", info)
	if legacy.Get_nox_wol_server_result_cnt_815088() >= 2500 || legacy.Get_dword_5d4594_815044() != 0 || dword_5d4594_815060 != 0 {
		log.Warn("ignoring server: don't need more results", "addr", info.Address)
		return 0
	}
	if info.Address == "" {
		log.Warn("ignoring server: invalid address", "addr", info.Address)
		return 0
	}
	if !legacy.Sub_4A0410(info.Address, info.Port) {
		log.Warn("ignoring server: duplicate?", "addr", info.Address)
		return 0
	}
	srv, freeSrv := alloc.New(legacy.Nox_gui_server_ent_t{})
	defer freeSrv()
	srv.Field_11_0 = info.Field_11_0
	srv.Field_11_2 = info.Field_11_2
	srv.SetVersion(info.Version)
	if info.Ping <= 0 {
		srv.PingVal = 9999 // UI interprets it as N/A
	} else {
		srv.PingVal = int32(info.Ping / time.Millisecond)
	}
	srv.StatusVal = info.Status
	srv.Field_25_1 = info.Field_25_1
	srv.Field_25_2 = info.Field_25_2
	if info.Players.Cur < 0 || info.Players.Cur > 0xff {
		srv.PlayersVal = 255
	} else {
		srv.PlayersVal = byte(info.Players.Cur)
	}
	if info.Players.Max < 0 || info.Players.Cur > 0xff {
		srv.MaxPlayersVal = 255
	} else {
		srv.MaxPlayersVal = byte(info.Players.Max)
	}
	*(*uint16)(unsafe.Pointer(&srv.Field_26_1)) = info.Field_26_1
	*(*uint16)(unsafe.Pointer(&srv.Field_26_3)) = info.Field_26_3
	srv.SetMapName(info.Map)
	copy(srv.Field_33_3[:], info.Field_33_3[:])
	*(*uint32)(unsafe.Pointer(&srv.Field_38_3)) = info.Field_38_3
	*(*uint32)(unsafe.Pointer(&srv.Field_39_3)) = info.Field_39_3
	srv.SetFlags(noxflags.GameFlag(info.Flags))
	srv.SetQuestLevel(info.Level)
	srv.Field_42 = 0
	if legacy.Get_dword_587000_87412() == -1 || legacy.Sub_437860(int(srv.Field_11_0), int(srv.Field_11_2)) == legacy.Get_dword_587000_87412() {
		if legacy.Nox_xxx_checkSomeFlagsOnJoin_4899C0(srv) != 0 {
			srv.SetAddr(info.Address)
			srv.Field_9 = uint32(legacy.Get_nox_wol_server_result_cnt_815088())
			srv.Field_7 = 0
			srv.SetPort(uint16(info.Port))
			srv.SetServerName(info.Name)
			srv.SetFlags(noxflags.GameFlag(info.Flags))
			legacy.Nox_wol_servers_addResult_4A0030(srv)
			legacy.Inc_nox_wol_server_result_cnt_815088()
		}
	}
	return 0
}

func clientWaitForLobbyResults() bool {
	waitForLobbyResults(lobbyBroadcast, netstr.RecvCanRead)
	return true
}

func waitForLobbyResults(conn net.PacketConn, flag netstr.RecvFlags) (int, error) {
	if conn == nil {
		return 0, client.ErrLobbyNoSocket
	}
	return netstr.WaitForLobbyResults(conn, legacy.Nox_client_getServerAddr_43B300(), flag, netstr.LobbyWaitOptions{
		OnResult: func(addr netip.AddrPort, data []byte) {
			saddr := addr.Addr().String()
			port := int(addr.Port())
			name := data[72:]
			name = name[:alloc.StrLenS(name)]
			onLobbyServerPacket(noxClient.Log, saddr, port, string(name), data)
		},
		OnPassRequired: func() {
			if legacy.Sub_43B6D0() != 0 {
				legacy.Sub_43AF90(5)
			}
		},
		OnPing: func(addr netip.AddrPort, buf []byte) {
			if legacy.Sub_43B6D0() != 0 {
				legacy.Sub_43AF90(4)
				buf[2] = byte(netmsg.MSG_SERVER_PONG)
				sendToServer(addr, buf[:8])
			}
		},
		OnConnectErr: func(errcode noxnet.ConnectError) bool {
			if errcode != noxnet.ErrDupSerial {
				if legacy.Sub_43B6D0() != 0 {
					nox_client_setConnError_43AFA0(errcode)
				}
				return false
			}
			// TODO: Code above is disabled because it causes issues with players reconnecting to the server.
			//       For some reason the player record gets stuck in the server's player list, so this check fails.
			gameLog.Printf("connect error: %d (%s, ignored)", errcode, errcode.Name())
			// from code20
			if legacy.Sub_43B6D0() != 0 && legacy.Sub_43AF80() == 3 {
				legacy.Sub_43AF90(7)
			}
			return true
		},
		OnJoinOK: func() {
			if legacy.Sub_43B6D0() != 0 && legacy.Sub_43AF80() == 3 {
				legacy.Sub_43AF90(7)
			}
		},
		OnJoinFail: func() {
			if legacy.Sub_43B6D0() != 0 {
				legacy.Sub_43AF90(8)
			}
		},
	})
}

func sendToServer(addr netip.AddrPort, data []byte) (int, error) {
	if lobbyBroadcast == nil {
		return 0, client.ErrLobbyNoSocket
	}
	if len(data) < 2 {
		return 0, nil
	}
	if lobbyBroadcast == nil {
		return 0, errors.New("no broadcast socket")
	}
	return lobbyBroadcast.WriteTo(data, net.UDPAddrFromAddrPort(addr))
}

func sub_420100() int { return int(memmap.Uint32(0x587000, 60072) >> 8) }

func nox_client_setConnError_43AFA0(err noxnet.ConnectError) {
	gameLog.Printf("connect error: %d (%s)", err, err.Name())
	legacy.Set_nox_client_connError_814552(int(err))
	legacy.Sub_43AF90(2)
}

func sub_4373A0() {
	c := noxClient
	if win := legacy.Get_dword_5d4594_815000(); !win.GetFlags().IsHidden() {
		win.Hide()
		legacy.Set_dword_5d4594_815056(0)
		win.StackPop()
		legacy.Get_nox_wol_wnd_world_814980().Focus()
	}
	if legacy.Get_dword_587000_87408() == 1 || legacy.Get_dword_587000_87412() == -1 {
		if legacy.Get_nox_game_createOrJoin_815048() == 1 {
			legacy.Set_nox_game_createOrJoin_815048(0)
			c.SetMouseBounds(image.Rect(0, 0, nox_win_width-1, nox_win_height-1))
			v2 := c.Strings().GetStringInFile("ChooseArea", "C:\\NoxPost\\src\\client\\shell\\noxworld.c")
			legacy.Get_dword_5d4594_814996().Func94(&gui.StaticTextSetText{Str: v2})
			clientPlaySoundSpecial(sound.SoundPermanentFizzle, 100)
		} else {
			c.nox_game_checkStateSwitch_43C1E0()
			legacy.Sub_49FF20()
			sub_4A1BE0(1)
			clientPlaySoundSpecial(sound.SoundPermanentFizzle, 100)
		}
	} else if legacy.Get_nox_game_createOrJoin_815048() == 1 {
		legacy.Set_nox_game_createOrJoin_815048(0)
		c.SetMouseBounds(image.Rect(0, 0, nox_win_width-1, nox_win_height-1))
		legacy.Get_dword_5d4594_814984().Capture(false)
		legacy.Sub_4375C0(1)
		v0 := c.Strings().GetStringInFile("JoinServer", "C:\\NoxPost\\src\\client\\shell\\noxworld.c")
		legacy.Get_dword_5d4594_814996().Func94(&gui.StaticTextSetText{Str: v0})
		clientPlaySoundSpecial(sound.SoundPermanentFizzle, 100)
	} else {
		legacy.Sub_49FF20()
		c.nox_game_checkStateSwitch_43C1E0()
	}
}
