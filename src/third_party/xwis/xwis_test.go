package xwis

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func init() {
	if false {
		DebugLog = log.New(os.Stderr, "", 0)
	}
}

func TestHost(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	const servUser = "testserv"
	info := GameInfo{
		Access:     AccessOpen,
		Resolution: Res640x480,
		Players:    3,
		MaxPlayers: 32,
		Map:        "headache",
		Name:       "Test Server",
		MapType:    MapTypeChat,
		FragLimit:  5,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cli, err := NewClient(ctx, servUser, servUser)
	require.NoError(t, err)
	defer cli.Close()

	err = cli.HostGame(ctx, info)
	require.NoError(t, err)
}

func TestRegister(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	const servUser = "testserv"
	info := GameInfo{
		Access:     AccessOpen,
		Resolution: Res640x480,
		Players:    4,
		MaxPlayers: 32,
		Map:        "headache",
		Name:       "Test Server",
		MapType:    MapTypeChat,
		FragLimit:  5,
	}
	info.setDefaults()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cli, err := NewClient(ctx, servUser, servUser)
	require.NoError(t, err)
	defer cli.Close()

	g, err := cli.RegisterGame(ctx, info)
	require.NoError(t, err)
	defer g.Close()

	assertGame := func() {
		list, err := cli.ListRooms(ctx)
		require.NoError(t, err)

		found := false
		for _, r := range list {
			if r.Game == nil {
				continue
			}
			if r.Game.Name == info.Name {
				require.NotEmpty(t, r.Game.Addr)
				info.Addr = r.Game.Addr
				require.Equal(t, info, *r.Game)
				found = true
				break
			}
		}
		require.True(t, found)
	}

	assertGame()

	info.Name += " 1"
	info.Map += "1"
	info.Players += 2

	err = g.Update(ctx, info)
	require.NoError(t, err)

	assertGame()
}

func TestList(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cli, err := NewClient(ctx, "", "")
	require.NoError(t, err)
	defer cli.Close()

	list, err := cli.ListRooms(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, list)
	for _, r := range list {
		if r.Game != nil {
			t.Logf("%q - %q - %s", r.ID, r.Name, r.Game.Addr)
		} else {
			t.Logf("%q - %q", r.ID, r.Name)
		}
	}
}

func TestListRoomsContinuesAfterTruncatedGameInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, peer := net.Pipe()
	defer client.Close()
	defer peer.Close()
	deadline, _ := ctx.Deadline()
	require.NoError(t, peer.SetDeadline(deadline))

	cli := &Client{c: client, r: newReader(client), w: newWriter(client), stop: make(chan struct{})}
	go cli.readLoop()
	defer cli.Close()

	result := make(chan error, 1)
	go func() {
		defer peer.Close()
		r := newReader(peer)
		request, err := r.ReadLine()
		if err == nil && request != "LIST -1 37" {
			err = fmt.Errorf("unexpected request %q", request)
		}
		if err == nil {
			_, err = io.WriteString(peer,
				":fixture 326 probe #truncated 0 0 0 0 0 2130706433 :128::\n"+
					":fixture 326 probe #valid 0 0 0 0 0 2130706433 :"+encodedInfoHdr+"\n"+
					":fixture 323 probe :End of list\n")
		}
		if err == nil {
			// Keep the stream open until the client finishes processing the list.
			_, err = r.WaitFor(ctx, "QUIT")
		}
		result <- err
	}()

	rooms, err := cli.ListRooms(ctx)
	require.NoError(t, err)
	require.Len(t, rooms, 2)
	require.Equal(t, Room{ID: "#truncated", Name: "truncated"}, rooms[0])
	require.Equal(t, "#valid", rooms[1].ID)
	require.Equal(t, "NoxCommunity EU", rooms[1].Name)
	require.NotNil(t, rooms[1].Game)
	require.Equal(t, "127.0.0.1", rooms[1].Game.Addr)
	require.NoError(t, cli.Close())
	require.NoError(t, <-result)
}

func TestListLobbyServers(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	list, err := ListLobbyServers(ctx)
	require.NoError(t, err)
	require.Equal(t, []LobbyServer{
		{Addr: "xwis.net:4000", Name: "XWIS"},
	}, list)
}
