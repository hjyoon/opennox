package legacy

/*
#include "client__gui__guiinv.h"
*/
import "C"

import (
	"image"
	"strconv"
	"unsafe"

	noxcolor "github.com/opennox/libs/color"
	"github.com/opennox/libs/player"
	"golang.org/x/image/font"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//export nox_client_inventory_total_weight_463880
func nox_client_inventory_total_weight_463880() C.int {
	var weight C.int
	for row := range inventoryRows {
		for column := range inventoryColumns {
			cell := inventoryGridCell(row, column)
			if cell.field_140 != 0 {
				weight += C.int(cell.field_140) * C.int(cell.field_0.field_74_3)
			}
		}
	}
	return weight
}

func clientInventoryStatText(id string, line int32, first, second int) string {
	var out [256]wchar2_t
	formatFixedWide(&out[0], clientInventoryString(id, line), formatValue{number: int32(first)}, formatValue{number: int32(second)})
	return GoWString(&out[0])
}

func clientInventoryStatInt(value float64) int {
	return int(server.X87TruncateSignedDWORD419A70(float32(value)))
}

//export nox_client_makePlayerStatsDlg_463880
func nox_client_makePlayerStatsDlg_463880(origin *C.int) {
	render := GetClient().R2()
	small := render.GetFonts().AsFont(inventoryFont)
	lineHeight := render.FontHeight(nil)
	offsetY := clientInventoryStatInt(float64(lineHeight-render.FontHeight(small))*0.5 + 0.5)
	pl := Get_dword_8531A0_2576()
	if pl == nil {
		return
	}
	pos := unsafe.Slice(origin, 2)
	x, y := int(pos[0])+11, int(pos[1])+15
	stats := GetServer().S().Players
	class := pl.PlayerClass()
	maximum := *stats.ClassStats(class)
	warrior := *stats.ClassStats(player.Warrior)
	info := pl.Info()
	draw := func(face font.Face, text string, x, y, width int) {
		render.DrawStringWrapped(face, text, image.Rect(x, y, x+width, y))
	}
	fill := func(color uint32, x, y, width, height int) {
		c := noxcolor.RGBA5551(color)
		render.Data().SetColor2(c)
		render.DrawRectFilledOpaque(x, y, width, height, c)
	}
	setTextColor := func(color uint32) {
		render.Data().SetTextColor(noxcolor.RGBA5551(color))
	}
	white := Get_nox_color_white_2523948()
	lightRed := memmap.Uint32(0x85B3FC, 940)
	darkBlue := memmap.Uint32(0x85B3FC, 944)
	blue := Get_nox_color_blue_2650684()
	yellow := Get_nox_color_yellow_2589772()
	setTextColor(white)
	fill(Get_nox_color_black_2650656(), x, y, 200, 200)
	x += 2
	y += 2*lineHeight + 3
	draw(nil, clientInventoryStatText("StatsLevel", 1878, int(pl.Level), 0), x, y, 200)
	y += lineHeight + 1
	if noxflags.HasGame(noxflags.GameModeCoop) {
		levelTarget := int(int32(int64(GetServer().S().Balance.FloatInd("XPTable", int(pl.Level)+1))))
		xp := int(memmap.Int32(0x5D4594, 1062544))
		draw(nil, clientInventoryStatText("StatsEXP", 1886, xp, levelTarget), x, y, 200)
	}
	y += 2*lineHeight + 2
	barX := x + 60
	bar := func(color uint32, width int) {
		fill(color, barX, y, width, lineHeight)
	}
	barWidth := func(value int32, max float32) int {
		return clientInventoryStatInt(float64(90*value) / float64(max))
	}
	limitText := func(current, max int, line int32) {
		text := clientInventoryStatText("MinMaxFormat", line, current, max)
		width := render.GetStringSizeWrapped(small, text, 0).X
		draw(small, text, x-width+193, y+offsetY, 200)
	}
	currentText := func(current int) {
		draw(small, strconv.Itoa(current), x+45, y+offsetY, 200)
	}

	draw(nil, GoWString(clientInventoryString("StatsHealth", 1896)), x, y, 200)
	bar(Get_nox_color_violet_2598268(), 90)
	bar(Get_nox_color_red_2589776(), barWidth(int32(info.Field2247()), maximum.Health))
	health := int32(sub_470CC0())
	bar(lightRed, barWidth(health, maximum.Health))
	limitText(int(int32(info.Field2247())), clientInventoryStatInt(float64(maximum.Health)), 1914)
	currentText(int(health))
	y += lineHeight + 1

	if class != player.Warrior {
		bar(darkBlue, 90)
		draw(nil, GoWString(clientInventoryString("StatsMana", 1941)), x, y, 200)
		bar(blue, barWidth(int32(info.Field2243()), maximum.Mana))
		mana := int32(nox_xxx_cliGetMana_470DD0())
		bar(Get_nox_color_cyan_2649820(), barWidth(mana, maximum.Mana))
		limitText(int(int32(info.Field2243())), clientInventoryStatInt(float64(maximum.Mana)), 1952)
		currentText(int(mana))
		y += lineHeight + 1
	}

	bar(memmap.Uint32(0x85B3FC, 956), 90)
	draw(nil, GoWString(clientInventoryString("StatsStrength", 1975)), x, y, 200)
	strength := int32(info.Field2239())
	bar(memmap.Uint32(0x5D4594, 2597996), barWidth(strength, maximum.Strength))
	limitText(int(strength), clientInventoryStatInt(float64(maximum.Strength)), 1982)
	currentText(int(strength))
	y += lineHeight + 1

	bar(Get_nox_color_orange_2614256(), 90)
	speed := int32(info.Field2235())
	width := clientInventoryStatInt(float64(90*speed)/float64(maximum.Speed) + 0.5)
	draw(nil, GoWString(clientInventoryString("StatsSpeed", 2006)), x, y, 200)
	bar(yellow, width)
	setTextColor(white)
	bonus := float32(float64(memmap.Float32(0x5D4594, 1063100)) / (float64(warrior.Speed) * 0.000001))
	if memmap.Uint8(0x5D4594, 1062541)&2 != 0 {
		bonus = float32((float64(width)+float64(bonus))*1.25 - float64(width))
	}
	if memmap.Uint8(0x5D4594, 1062540)&0x10 != 0 {
		bonus = float32((float64(width)+float64(bonus))*0.5 - float64(width))
	}
	switch {
	case bonus > 0:
		delta := min(90-width, clientInventoryStatInt(float64(bonus)))
		fill(yellow, barX+width, y, delta, lineHeight)
		setTextColor(blue)
	case bonus < 0:
		delta := clientInventoryStatInt(float64(bonus))
		fill(darkBlue, barX+width+delta, y, clientInventoryStatInt(float64(-bonus)), lineHeight)
	}
	bonusPercent := float32(float64(bonus) * 100 * 0.011111111)
	speedPercent := clientInventoryStatInt(float64(speed)*100/float64(warrior.Speed) + float64(bonusPercent) + 0.5)
	maxSpeed := clientInventoryStatInt(float64(maximum.Speed) * 100 / float64(warrior.Speed))
	limitText(speedPercent, maxSpeed, 2045)
	setTextColor(white)
	currentText(speedPercent)

	y += 2*lineHeight + 2
	if lang := nox_strman_get_lang_code(); lang == 6 || lang == 8 {
		x += 39
	}
	setTextColor(white)
	armorLabel := GoWString(clientInventoryString("StatsArmorLabel", 2072))
	armorWidth := render.GetStringSizeWrapped(nil, armorLabel, 0).X
	draw(nil, armorLabel, x, y, 0)
	valueX := x + armorWidth + 5
	armor := clientInventoryStatInt(float64(memmap.Float32(0x5D4594, 1062548))*1000 + 0.5)
	draw(nil, clientInventoryStatText("MinMaxFormat", 2076, armor, 1000), valueX, y, 0)

	y += lineHeight + 1
	weight := int(nox_client_inventory_total_weight_463880())
	capacity := int(pl.StatsCapacity())
	setTextColor(white)
	weightLabel := GoWString(clientInventoryString("DollWeight", 2098))
	labelWidth := render.GetStringSizeWrapped(nil, weightLabel, 0).X
	draw(nil, weightLabel, x+armorWidth-labelWidth, y, 0)
	if weight > capacity {
		setTextColor(lightRed)
	}
	draw(nil, clientInventoryStatText("MinMaxFormat", 2107, weight, capacity), valueX, y, 0)
}
