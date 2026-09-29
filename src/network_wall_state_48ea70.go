package opennox

import (
	"encoding/binary"
	"image"

	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

type wallStateHooks48EA70 struct {
	connected    func() bool
	host         func() bool
	secretByID   func(uint16) *server.Wall
	wallAtGrid   func(image.Point) *server.Wall
	createAtGrid func(image.Point) *server.Wall
	deleteAtGrid func(image.Point)
}

func findSecretWallByIDNative48EA70(first *server.SecretWall, id uint16) *server.Wall {
	for it := first; it != nil; it = it.NextWall() {
		if it.Wall.Field10 == id {
			return it.Wall
		}
	}
	return nil
}

func handleWallStateNative48EA70(op netmsg.Op, data []byte, hooks wallStateHooks48EA70) int {
	switch op {
	case netmsg.MSG_OPEN_WALL, netmsg.MSG_CLOSE_WALL:
		const size = 3
		if len(data) < size {
			return -1
		}
		if !hooks.connected() || hooks.host() {
			return size
		}
		wl := hooks.secretByID(binary.LittleEndian.Uint16(data[1:size]))
		secret := wl.Secret()
		if secret == nil {
			return size
		}
		if op == netmsg.MSG_OPEN_WALL {
			secret.OpenDelay = 23
			secret.State = 3
		} else {
			secret.OpenDelay = 0
			secret.State = 1
		}
		return size

	case netmsg.MSG_CHANGE_OR_ADD_WALL_MAGIC:
		const size = 6
		if len(data) < size {
			return -1
		}
		if !hooks.connected() {
			return size
		}
		pos := image.Pt(int(data[4]), int(data[5]))
		wl := hooks.wallAtGrid(pos)
		if wl == nil {
			wl = hooks.createAtGrid(pos)
		}
		if wl != nil {
			wl.Tile1 = data[1]
			wl.Dir0 = data[2]
			wl.Field2 = data[3]
		}
		return size

	case netmsg.MSG_REMOVE_WALL_MAGIC:
		const size = 3
		if len(data) < size {
			return -1
		}
		if !hooks.connected() {
			return size
		}
		if wl := hooks.wallAtGrid(image.Pt(int(data[1]), int(data[2]))); wl != nil {
			hooks.deleteAtGrid(wl.GridPos())
		}
		return size
	default:
		return -1
	}
}

func (c *Client) handleWallStatePacketNative48EA70(op netmsg.Op, data []byte) int {
	return handleWallStateNative48EA70(op, data, wallStateHooks48EA70{
		connected: nox_client_isConnected,
		host:      func() bool { return noxflags.HasGame(noxflags.GameHost) },
		secretByID: func(id uint16) *server.Wall {
			return findSecretWallByIDNative48EA70(nox_xxx_wallSecretGetFirstWall_410780(), id)
		},
		wallAtGrid:   c.srv.Walls.GetWallAtGrid,
		createAtGrid: c.srv.Walls.CreateAtGrid,
		deleteAtGrid: c.srv.Walls.DeleteAtGrid,
	})
}
