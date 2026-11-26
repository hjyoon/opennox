package legacy

/*
#include "GAME2_1.h"

*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/wall"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type clientMinimapView struct {
	originX, originY int32
	top, worldSize   int32
	zoom, layer      int32
}

func (m clientMinimapView) project(x, y int32) [2]int32 {
	return [2]int32{100 * (x - m.originX) / m.zoom, m.top + 100*(y-m.originY)/m.zoom}
}

func (m clientMinimapView) projectDebug(x, y float32) [2]int32 {
	return [2]int32{
		int32(100*(int64(x)-int64(m.originX))) / m.zoom,
		m.top + int32(100*(int64(y)-int64(m.originY)))/m.zoom,
	}
}

func drawClientMinimapBackground(top, size int32) {
	currentDrawClipRect(0, 0, clientDisplayStorage.width, clientDisplayStorage.height)
	viewportX := int32(nox_draw_getViewport_437250().x1)
	if viewportX == 0 {
		nox_client_drawRectFilledAlpha_49CF10(0, top, size, size)
	} else {
		nox_client_drawSetColor_434460(int32(renderPalette.black))
		width := min(viewportX, size)
		nox_client_drawRectFilledOpaque_49CE30(0, top, width, size)
		if viewportX < size {
			nox_client_drawRectFilledAlpha_49CF10(width, top, size-width, size)
		}
	}
	nox_client_drawEnableAlpha_434560(1)
	nox_client_drawSetColor_434460(int32(renderPalette.black))
	for i, alpha := range [...]uint8{0x5a, 0x3c, 0x28} {
		extra := C.int(i + 1)
		nox_client_drawSetAlpha_434580(C.uchar(alpha))
		nox_client_drawRectLines_473510(-extra, C.int(top)-extra, C.int(size)+2*extra, C.int(size)+2*extra)
	}
	nox_client_drawEnableAlpha_434560(0)
	currentDrawClipRect(0, C.int(top), C.int(size), C.int(size))
}

func minimapWallVisible(tile *server.Wall) bool {
	if !noxflags.HasGame(noxflags.GameNotQuest) && tile.Field12 == 0 {
		return false
	}
	return (!tile.Flags4.Has(wall.FlagSecret) || !secretWallHiddenMinimap472600(tile.Data)) &&
		!tile.Flags4.Has(wall.FlagBroken)
}

func minimapWallAt(x, y int32) *server.Wall {
	return asWallP(nox_server_getWallAtGrid_410580(x, y))
}

func minimapDoorVisible(x, y int32) bool {
	offsets := memmap.PtrT[[4][2]int32](0x587000, 149240)
	for _, pair := range offsets {
		dx, dy := pair[0], pair[1]
		if diagonal := minimapWallAt(x+dx, y+dy); diagonal != nil {
			if diagonal.Field12 != 0 {
				return true
			}
			continue
		}
		if neighbor := minimapWallAt(x+dx, y); neighbor != nil && neighbor.Field12 != 0 {
			return true
		}
		if neighbor := minimapWallAt(x, y+dy); neighbor != nil && neighbor.Field12 != 0 {
			return true
		}
	}
	return false
}

func minimapMutedColor() int32 { return memmap.Int32(0x85B3FC, 940) }

func (m clientMinimapView) drawDoor(tile *server.Wall) {
	drawable := (*nox_drawable)(tile.ClientData)
	if drawable == nil || !minimapDoorVisible(int32(tile.X5), int32(tile.Y6)) {
		return
	}
	at := m.project(int32(drawable.pos.x), int32(drawable.pos.y))
	offset := 196184 + 8*uintptr(drawable.field_74_4)
	dx, dy := memmap.Int32(0x587000, offset), memmap.Int32(0x587000, offset+4)
	nox_client_drawSetColor_434460(minimapMutedColor())
	nox_client_drawAddPoint_49F500(at[0], at[1])
	nox_xxx_rasterPointRel_49F570(100*dx/m.zoom, 100*dy/m.zoom)
	nox_client_drawLineFromPoints_49E4B0()
}

func (m clientMinimapView) drawWalls() {
	invisible := memmap.PtrT[[2]byte](0x5D4594, 1096300)
	if invisible[0] == 0 {
		invisible[0] = byte(nox_xxx_wallTileByName_410D60(internCStr("InvisibleWallSet")))
		invisible[1] = byte(nox_xxx_wallTileByName_410D60(internCStr("InvisibleBlockingWallSet")))
	}
	minX, maxX := m.originX/wall.GridStep, (m.originX+m.worldSize)/wall.GridStep
	minY, maxY := m.originY/wall.GridStep, (m.originY+m.worldSize)/wall.GridStep
	for y := minY; y <= maxY; y++ {
		for tile := asWallP(sub_4106A0(y)); tile != nil; tile = tile.NextByY24 {
			layer := int32(byte(tile.Field8))
			x := int32(tile.X5)
			if tile.Tile1 == invisible[0] || tile.Tile1 == invisible[1] ||
				layer != 0 && layer != m.layer || x < minX {
				continue
			}
			if x > maxX {
				break
			}
			if tile.Flags4.Has(wall.FlagDoor) {
				m.drawDoor(tile)
			} else if minimapWallVisible(tile) {
				p := m.project(x*wall.GridStep, y*wall.GridStep)
				at := C.int2{field_0: C.int(p[0]), field_4: C.int(p[1])}
				sub_4730D0(&at, C.uchar(tile.Dir0), C.int(100*wall.GridStep/m.zoom))
			}
		}
	}
}

func (m clientMinimapView) drawAI() {
	count, raw := sub_50CB00(), sub_50CB10()
	if count >= 2 && raw != nil {
		points := unsafe.Slice((*[2]float32)(raw), count)
		nox_client_drawSetColor_434460(int32(worldWordStorage.local_player_color))
		for i := 1; i < len(points); i++ {
			a, b := points[i-1], points[i]
			p, q := m.projectDebug(a[0], a[1]), m.projectDebug(b[0], b[1])
			drawMinimapLine(p[0], p[1], q[0], q[1])
		}
	}
	for unit := minimapMonsterFirstImpl50AAE0(); unit != nil; unit = minimapMonsterNextImpl50AB10() {
		nox_client_drawSetColor_434460(minimapMutedColor())
		at := m.projectDebug(unit.PosVec.X, unit.PosVec.Y)
		nox_xxx_minimapDrawPoint_473570(C.int(at[0]), C.int(at[1]))
	}
}

func minimapSetTeamColor(id int32) {
	if team := nox_xxx_getTeamByID_418AB0(id); team != nil {
		nox_client_drawSetColor_434460(int32(nox_xxx_materialGetTeamColor_418D50(team)))
	}
}

func minimapCrownShown() bool {
	if nox_server_teamFirst_418B10() != nil {
		return true
	}
	for player := nox_xxx_cliGetSpritePlayer_45A000(); player != nil; player = sub_45A010(player) {
		if bool(nox_client_drawable_testBuff_4356C0(player, 30)) {
			return false
		}
	}
	return true
}

func (m clientMinimapView) drawTrackedObjects(local *nox_drawable) {
	crown := memmap.PtrUint32(0x5D4594, 1096304)
	ball := memmap.PtrUint32(0x5D4594, 1096308)
	if *crown == 0 {
		*crown = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("Crown")))
		*ball = uint32(nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr("GameBall")))
	}
	for item := nox_xxx_cliFirstMinimapObj_459EB0(); item != nil; item = nox_xxx_cliNextMinimapObj_459EC0(item) {
		position := [2]int32{int32(item.pos.x), int32(item.pos.y)}
		layer := int32(1)
		if polygon := findMapPolygonAt(&position, 0, false); polygon != nil {
			layer = int32(polygon.AudioZone)
		}
		if layer != m.layer {
			continue
		}
		at := m.project(position[0], position[1])
		drawAt := (*C.uint32_t)(unsafe.Pointer(&at[0]))
		color := minimapMutedColor()
		if item.flags28&0x400000 != 0 && item.flags29&8 != 0 {
			color = int32(renderPalette.blue)
		}
		nox_client_drawSetColor_434460(color)
		switch {
		case uint32(item.field_27) == *crown:
			if minimapCrownShown() {
				nox_client_drawSetColor_434460(int32(worldWordStorage.local_player_color))
				if member := teamObjectByNetCode418C80(uint32(item.field_32)); member != nil {
					minimapSetTeamColor(int32(member.ID))
				}
				sub_473420(drawAt)
			}
		case uint32(item.field_27) == *ball:
			member := teamObjectByNetCode418C80(uint32(item.field_32))
			if member.Has() {
				minimapSetTeamColor(int32(member.ID))
			} else {
				nox_client_drawSetColor_434460(int32(renderPalette.white))
			}
			nox_video_drawCircleRad3_4734F0((*C.int)(unsafe.Pointer(&at[0])))
		case item.flags28&0x10000000 != 0:
			if item.flags30&0x1000000 != 0 {
				nox_client_drawSetColor_434460(int32(renderPalette.white))
				minimapSetTeamColor(int32(sub_4B94E0(item)))
				sub_4733B0(drawAt)
			}
		case item.flags28&4 == 0:
			nox_xxx_minimapDrawPoint_473570(C.int(at[0]), C.int(at[1]))
		case !noxflags.HasGame(noxflags.GameModeCTF):
			if item == local {
				sub_4735C0(C.int(at[0]), C.int(at[1]))
			} else {
				nox_xxx_minimapDrawPoint_473570(C.int(at[0]), C.int(at[1]))
			}
		default:
			player := nox_common_playerInfoGetByID_417040(int32(item.field_32))
			if player != nil && player.field_4&1 != 0 {
				nox_client_drawSetColor_434460(int32(renderPalette.white))
				if member := teamObjectByNetCode418C80(uint32(item.field_32)); member != nil {
					opposite := int32(1)
					if member.ID == 1 {
						opposite = 2
					}
					minimapSetTeamColor(opposite)
				}
				sub_4733B0(drawAt)
			}
		}
	}
}

func (m clientMinimapView) drawPlayers(local *nox_drawable, localTeam *server.ObjectTeam) {
	playerColor := int32(worldWordStorage.local_player_color)
	for drawable := nox_xxx_cliGetSpritePlayer_45A000(); drawable != nil; drawable = sub_45A010(drawable) {
		crowned := bool(nox_client_drawable_testBuff_4356C0(drawable, 30))
		member := teamObjectByNetCode418C80(uint32(drawable.field_32))
		player := nox_common_playerInfoGetByID_417040(int32(drawable.field_32))
		sameTeam := localTeam.SameAs(member)
		if !sameTeam && !crowned && drawable != local {
			observer := Get_dword_8531A0_2576()
			if observer == nil || observer.Field3680&1 == 0 {
				continue
			}
		}
		position := [2]int32{int32(drawable.pos.x), int32(drawable.pos.y)}
		polygon := findMapPolygonAt(&position, 0, false)
		if player == nil || player.field_3680&1 != 0 || polygon != nil && int32(polygon.AudioZone) != m.layer {
			continue
		}
		at := m.project(position[0], position[1])
		color := minimapMutedColor()
		if drawable == local || sameTeam {
			color = playerColor
		}
		nox_client_drawSetColor_434460(color)
		if crowned {
			if member != nil {
				minimapSetTeamColor(int32(member.ID))
			}
			sub_473420((*C.uint32_t)(unsafe.Pointer(&at[0])))
		} else {
			nox_xxx_minimapDrawPoint_473570(C.int(at[0]), C.int(at[1]))
		}
	}
}

//export nox_xxx_cliDrawMinimap_472600
func nox_xxx_cliDrawMinimap_472600(_ *C.nox_drawable, layer C.int) C.int {
	zoom := int32(minimapZoom())
	if zoom == 0 {
		return 0
	}
	size := int32(clientDisplayStorage.width / 6)
	worldSize := size * zoom / 100
	top := (int32(clientDisplayStorage.height) - size) / 2
	camera := GetClient().Viewport().World.Max
	view := clientMinimapView{
		originX: int32(camera.X) - worldSize/2,
		originY: int32(camera.Y) - worldSize/2,
		top:     top, worldSize: worldSize, zoom: zoom, layer: int32(layer),
	}
	nox_client_drawEnableAlpha_434560(0)
	nox_xxx_wndDraw_49F7F0()
	drawClientMinimapBackground(top, size)
	view.drawWalls()
	if noxflags.HasEngine(noxflags.EngineShowAI) {
		view.drawAI()
	}
	local := (*nox_drawable)(*memmap.PtrPtr(0x852978, 8))
	team := teamObjectByNetCode418C80(clientPlayerNetCode)
	view.drawTrackedObjects(local)
	view.drawPlayers(local, team)
	return sub_49F860()
}
