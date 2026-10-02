package opennox

import (
	"fmt"
	"math/bits"
	"slices"

	"github.com/opennox/libs/noximage"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type e2eEquipmentColorSubject struct {
	unit      *server.Object
	drawable  *client.Drawable
	holder    server.ArmorAndWeaponHolder
	animation int
}

// CheckEquipmentColors observes already-equipped inventory, network replay and
// the current animation. Mode 0 checks the player; mode 1 checks a visible
// equipped NPC. It neither grants items nor alters live equipment or colors.
// The stock layer is rendered in scratch buffers through DrawWeapon/DrawArmor
// and compared with an independent palette and the same native draw callback.
func (sc *e2eScenario) CheckEquipmentColors(mode int, name string) {
	if mode != 0 && mode != 1 {
		panic("equipment color check requires player mode 0 or NPC mode 1")
	}
	var subject *e2eEquipmentColorSubject
	sc.addWhen(0, name, 1200, func() bool {
		subject = e2eEquipmentColorTarget(mode)
		return subject != nil
	}, func() {
		if err := e2eCheckEquipmentColorSubject(subject); err != nil {
			e2eError(err)
		}
	})
}

func e2eEquipmentColorTarget(mode int) *e2eEquipmentColorSubject {
	if noxServer == nil || noxClient == nil || noxClient.ClientPlayerUnit() == nil || noxClient.Viewport() == nil {
		return nil
	}
	var best *e2eEquipmentColorSubject
	bestItems := 0
	player := noxServer.Players.HostUnit()
	drawData := (*client.PlayerDrawData)(noxClient.ClientPlayerUnit().DrawData)
	if drawData == nil {
		return nil
	}
	for unit := noxServer.Objs.First(); unit != nil; unit = unit.Next() {
		if mode == 0 && unit != player || mode == 1 && !unit.Class().Has(object.ClassMonster) ||
			unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			continue
		}
		code := noxServer.GetUnitNetCode(unit)
		if code <= 0 || code > int(^uint16(0)) {
			continue
		}
		dr := noxClient.Objs.ByNetCode(uint16(code))
		if dr == nil || hasChatBubble(dr) || int(dr.AnimDir) >= 9 || dr.AnimDir == 4 ||
			dr.ZVal != 0 || dr.HasEnchant(server.ENCHANT_FREEZE) || dr.HasEnchant(server.ENCHANT_INVULNERABLE) ||
			!noxClient.Viewport().ToScreenPos(dr.Pos()).In(noxClient.Viewport().Screen) {
			continue
		}
		var holder server.ArmorAndWeaponHolder
		if mode == 0 && dr.DrawFuncPtr == legacy.Get_nox_thing_player_draw() {
			holder = noxClient.Server.Players.ByID(int(dr.NetCode32))
		} else if mode == 1 && dr.DrawFuncPtr == legacy.Get_nox_thing_npc_draw() {
			holder = noxClient.Server.NPCs.ByID(int(dr.NetCode32))
		} else {
			continue
		}
		// Avoid a typed nil interface before calling its data accessors.
		if mode == 0 && holder.(*server.Player) == nil || mode == 1 && holder.(*server.NPC) == nil {
			continue
		}
		weapon, armor, _ := e2eEquippedNPCMasks(unit)
		clientWeapon, _ := holder.WeaponData()
		clientArmor, _ := holder.ArmorData()
		if weapon != clientWeapon || armor != clientArmor {
			continue // allow the native equipment packets to arrive
		}
		animation := int(dr.AnimInd)
		if mode == 0 && dr.AnimInd == 4 && dr.HasEnchant(server.ENCHANT_SNEAK) {
			animation = 53
		}
		if mode == 1 {
			// NPC animation numbers are not player animation numbers. In
			// particular, NPC state 8 selects player idle animation 0. Avoid
			// the random attack branches before observing the selector on a copy.
			if dr.AnimInd >= 1 && dr.AnimInd <= 4 || (dr.AnimInd == 5 || dr.AnimInd == 6) && weapon&0x400 != 0 {
				continue
			}
			probe := *dr
			animation = noxClient.nox_xxx_spriteNPCInfo_49A4B0(&probe, weapon, armor)
		}
		if animation < 0 || animation >= len(drawData.Anim) {
			continue
		}
		base := &drawData.Anim[animation].Base
		if base.Cnt40 == 0 || base.Kind == client.AnimRandom || base.Kind == client.AnimOneShotRemove {
			continue // the observer must not consume RNG or delete a drawable
		}
		count := bits.OnesCount32(weapon&^1) + bits.OnesCount32(armor)
		if count > bestItems {
			best = &e2eEquipmentColorSubject{unit: unit, drawable: dr, holder: holder, animation: animation}
			bestItems = count
		}
	}
	return best
}

// GAME.EXE 004B8E10/004B8CA0 use definition slots 1..6 and apply the
// effectiveness, material, primary and secondary colors in that order. The
// reference deliberately does not call the production palette or slot helpers.
func e2eEquipmentPalette(def *server.Modifier, mods [4]*server.ModifierEff, initial [16]noxrender.Color16) [16]noxrender.Color16 {
	want := initial
	for slot := 1; slot <= 6; slot++ {
		color := def.Colors12[slot]
		want[slot] = noxrender.Color16{R: uint16(color.R), G: uint16(color.G), B: uint16(color.B)}
	}
	slots := [4]int32{def.Effectiveness36, def.Material40, def.PriEnchant44, def.SecEnchant48}
	for index, mod := range mods {
		slot := slots[index]
		if mod != nil && slot >= 0 && slot < int32(len(want)) {
			color := mod.Color24
			want[slot] = noxrender.Color16{R: uint16(color.R), G: uint16(color.G), B: uint16(color.B)}
		}
	}
	return want
}

func e2eCheckEquipmentColorSubject(subject *e2eEquipmentColorSubject) error {
	c := noxClient
	data := c.r.Data()
	liveData := *data
	livePix := c.r.PixBuffer()
	defer func() {
		*data = liveData
		c.r.SetPixBuffer(livePix)
	}()
	// The native sprite callback writes its cached image/extent into the
	// drawable. Use C-owned copies, leaving the live drawable/viewport intact.
	dr, freeDrawable := alloc.New(*subject.drawable)
	defer freeDrawable()
	vp, freeViewport := alloc.New(*c.Viewport())
	defer freeViewport()
	drawData := (*client.PlayerDrawData)(c.ClientPlayerUnit().DrawData)
	if drawData == nil || subject.animation < 0 || subject.animation >= len(drawData.Anim) || drawData.Anim[subject.animation].Base.Cnt40 == 0 {
		return fmt.Errorf("equipment colors: no current player animation")
	}
	panim := &drawData.Anim[subject.animation]
	if panim.Base.Kind == client.AnimRandom || panim.Base.Kind == client.AnimOneShotRemove {
		return fmt.Errorf("equipment colors: current animation cannot be observed without RNG/deletion")
	}
	frame := c.getAnimFrameInd(dr, &panim.Base)
	if frame < 0 || frame >= int(panim.Base.Cnt40) {
		return fmt.Errorf("equipment colors: current frame %d out of range", frame)
	}
	_, weapons := subject.holder.WeaponData()
	_, armor := subject.holder.ArmorData()
	beforeWeapons, beforeArmor := *weapons, *armor
	var layers, pixels, modifiers [2]int
	for item := subject.unit.InvFirstItem; item != nil; item = item.InvNextItem {
		if !item.Flags().Has(object.FlagEquipped) {
			continue
		}
		kind := 0
		var bit uint32
		var entries []server.EquipmentData
		var layer *client.PlayerEquipAnimation
		var def *server.Modifier
		if item.Class().HasAny(object.ClassWeapon | object.ClassWand) {
			bit = noxServer.Weapons.Nox_xxx_weaponInventoryEquipFlags_415820(item)
			entries, def = weapons[:], c.Server.Modif.Dword_5d4594_251600
			index := bits.TrailingZeros32(bit)
			if index < len(panim.Weapon) {
				layer = panim.Weapon[index]
			}
		} else if item.Class().Has(object.ClassArmor) {
			kind = 1
			bit = noxServer.Armor.Nox_xxx_unitArmorInventoryEquipFlags_415C70(item)
			entries, def = armor[:], c.Server.Modif.Dword_5d4594_251608
			index := bits.TrailingZeros32(bit)
			if index < len(panim.Armor) {
				layer = panim.Armor[index]
			}
		} else {
			continue // flags have a separate draw path, not a material layer
		}
		if bit == 0 || bit&(bit-1) != 0 || layer == nil || layer.Frames[dr.AnimDir] == nil {
			return fmt.Errorf("equipment colors: type=%d bit=%#x has no directional layer", item.TypeInd, bit)
		}
		for def != nil && def.TypeInd != uint32(item.TypeInd) {
			def = def.Next80
		}
		if def == nil {
			return fmt.Errorf("equipment colors: stock type %d has no material definition", item.TypeInd)
		}
		var record *server.EquipmentData
		for i := range entries {
			if entries[i].Field0 == bit {
				record = &entries[i]
				break
			}
		}
		if record == nil {
			return fmt.Errorf("equipment colors: %s bit=%#x has no replayed record", def.Name(), bit)
		}
		var mods [4]*server.ModifierEff
		if item.InitData != nil {
			mods = item.InitDataModifier().Modifiers
		}
		colorSlots := [4]int32{def.Effectiveness36, def.Material40, def.PriEnchant44, def.SecEnchant48}
		for i, mod := range mods {
			replayed := (*server.ModifierEff)(record.Field4[i])
			if (replayed == nil) != (mod == nil) {
				return fmt.Errorf("equipment colors: %s modifier %d color_slot=%d replay=%p inventory=%p", def.Name(), i, colorSlots[i], replayed, mod)
			}
			if mod != nil && (replayed.Name() != mod.Name() || replayed.Index() != mod.Index() || replayed.Color24 != mod.Color24) {
				return fmt.Errorf("equipment colors: %s modifier %d color_slot=%d replay=%s/%d/%v inventory=%s/%d/%v", def.Name(), i, colorSlots[i], replayed.Name(), replayed.Index(), replayed.Color24, mod.Name(), mod.Index(), mod.Color24)
			}
			if mod != nil {
				modifiers[kind]++
			}
		}
		var initial [16]noxrender.Color16
		for slot := range initial {
			initial[slot] = noxrender.Color16{R: uint16(11 + 7*slot), G: uint16(213 - 5*slot), B: uint16(29 + 9*slot)}
		}
		want := e2eEquipmentPalette(def, mods, initial)
		beforeDef := *def
		var beforeMods [4]server.ModifierEff
		for i, mod := range mods {
			if mod != nil {
				beforeMods[i] = *mod
			}
		}
		actual := noximage.NewImage16(livePix.Rect)
		*data = liveData
		for slot, color := range initial {
			data.SetMaterialRGB(slot, int(color.R), int(color.G), int(color.B))
		}
		c.r.SetPixBuffer(actual)
		if kind == 0 {
			c.DrawWeapon(vp, dr, bit, weapons, panim, frame)
		} else {
			c.DrawArmor(vp, dr, bit, armor, panim, frame)
		}
		for slot, color := range want {
			if got := data.ColorMultOp(slot); got != color {
				return fmt.Errorf("equipment colors: %s bit=%#x slot=%d got=%v want=%v", def.Name(), bit, slot, got, color)
			}
		}
		reference := noximage.NewImage16(livePix.Rect)
		*data = liveData
		for slot, color := range want {
			data.SetMaterialRGB(slot, int(color.R), int(color.G), int(color.B))
		}
		c.r.SetPixBuffer(reference)
		image := panim.FramesSlice(layer.Frames[dr.AnimDir])[frame]
		legacy.Nox_xxx_drawObject_4C4770_draw(vp, dr, image)
		count, err := e2eEquipmentLayerPixels(actual, reference)
		if err != nil {
			return fmt.Errorf("equipment colors: %s bit=%#x: %w", def.Name(), bit, err)
		}
		if *weapons != beforeWeapons || *armor != beforeArmor || *def != beforeDef {
			return fmt.Errorf("equipment colors: render changed equipment/definition data")
		}
		for i, mod := range mods {
			if mod != nil && *mod != beforeMods[i] {
				return fmt.Errorf("equipment colors: render changed modifier %d", i)
			}
		}
		layers[kind]++
		pixels[kind] += count
		e2eLog.Printf("EQUIPMENT COLOR LAYER VERIFIED: unit=%q item=%s bit=%#x animation=%d frame=%d direction=%d image=%v pixels=%d materials=%v", subject.unit.ID(), def.Name(), bit, subject.animation, frame, dr.AnimDir, c.r.Bag.AsImage(image), count, want[1:7])
	}
	if layers[0]+layers[1] == 0 {
		return fmt.Errorf("equipment colors: no material layers checked")
	}
	e2eLog.Printf("EQUIPMENT COLORS VERIFIED: unit=%q player=%t weapon_layers=%d armor_layers=%d weapon_pixels=%d armor_pixels=%d weapon_modifiers=%d armor_modifiers=%d live_equipment_unchanged=true", subject.unit.ID(), subject.unit.Class().Has(object.ClassPlayer), layers[0], layers[1], pixels[0], pixels[1], modifiers[0], modifiers[1])
	return nil
}

func e2eEquipmentLayerPixels(actual, reference *noximage.Image16) (int, error) {
	if actual == nil || reference == nil || actual.Rect != reference.Rect || actual.Stride != reference.Stride ||
		actual.Rect.Empty() || actual.Stride != actual.Rect.Dx() ||
		len(actual.Pix) != actual.Rect.Dx()*actual.Rect.Dy() || !slices.Equal(actual.Pix, reference.Pix) {
		return 0, fmt.Errorf("stock layer pixels differ or have incompatible bounds")
	}
	count := 0
	for _, pixel := range reference.Pix {
		if pixel != 0 {
			count++
		}
	}
	if count == 0 {
		return 0, fmt.Errorf("stock layer drew no pixels")
	}
	return count, nil
}
