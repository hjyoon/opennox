package legacy

import "unsafe"

// Raw 004E0B30 control flow. Sites and widths follow the sealed 1783-byte,
// 592-instruction parent, including its narrow/lazy reads and cached values.
// Registration, C exports and real service dispatch are deliberately separate.
// No original row, expected plan, fixture token or reference outcome is input.
func rawDefaultDamageNative4E0B30(target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
	rawDefaultDamageObserveCaller4E0B30(&damage, 0x4e0b30, true, true)
	read := func(base unsafe.Pointer, field, width int, site uint32) uint32 {
		return rawDefaultDamageReadNativeOperand4E0B30(base, field, width, site, false).Word
	}
	pointerWidth := int(unsafe.Sizeof(unsafe.Pointer(nil)))
	readPointer := func(base unsafe.Pointer, field int, site uint32) unsafe.Pointer {
		return rawDefaultDamageReadNativeOperand4E0B30(base, field, pointerWidth, site, true).Pointer
	}
	write := func(base unsafe.Pointer, field, width int, site, value uint32) {
		rawDefaultDamageWriteNativeOperand4E0B30(base, field, width, site, false, rawDefaultDamageNativeValue4E0B30{Word: value})
	}
	writePointer := func(base unsafe.Pointer, field int, site uint32, value unsafe.Pointer) {
		rawDefaultDamageWriteNativeOperand4E0B30(base, field, pointerWidth, site, true, rawDefaultDamageNativeValue4E0B30{Pointer: value})
	}
	assignDamage := func(site uint32, value int32) {
		rawDefaultDamageObserveCaller4E0B30(&damage, site, false, false)
		damage = value
		rawDefaultDamageObserveCaller4E0B30(&damage, site, true, false)
	}
	object := func(value unsafe.Pointer) rawDefaultDamageNativeArgument4E0B30 {
		return rawDefaultDamageNativeArgument4E0B30{Kind: "object", Pointer: value}
	}
	word := func(value uint32) rawDefaultDamageNativeArgument4E0B30 {
		return rawDefaultDamageNativeArgument4E0B30{Kind: "word", Word: value}
	}
	damageArgument := func() rawDefaultDamageNativeArgument4E0B30 {
		return rawDefaultDamageNativeArgument4E0B30{Kind: "damage", Damage: &damage}
	}
	text := func(value string) rawDefaultDamageNativeArgument4E0B30 {
		return rawDefaultDamageNativeArgument4E0B30{Kind: "string", String: value}
	}
	service := func(site uint32, arguments ...rawDefaultDamageNativeArgument4E0B30) rawDefaultDamageNativeReply4E0B30 {
		return rawDefaultDamageInvokeService4E0B30(rawDefaultDamageNativeService4E0B30{Site: site, Arguments: arguments})
	}
	cw := uint16(0x0e3f)
	if observed, ok := rawDefaultDamageObservedControlWord4E0B30(); ok {
		cw = observed
	}
	if target == nil {
		return 1
	}
	if service(0x4e0b46, object(target), word(23)).Word != 0 {
		if rawDefaultDamageNativeFrame4E0B30(0x4e0b52, 1)&3 == 0 {
			service(0x4e0b66, word(71), object(target), word(0), word(0))
		}
		return 1
	}
	if read(target, 1, 1, 0x4e0b7b)&2 != 0 {
		update := readPointer(target, 18, 0x4e0b86)
		write(update, 24, 4, 0x4e0b8f, 0)
		if typ == 1 || typ == 12 || typ == 7 || typ == 14 || typ == 6 {
			status := read(update, 21, 4, 0x4e0baf)
			write(update, 21, 4, 0x4e0baf, status|0x80000)
		}
	}
	if read(target, 3, 4, 0x4e0bb9)&0x8000 != 0 {
		if service(0x4e0bc2, object(target)).Word == 0 {
			return 1
		}
		attribution := weapon
		if attribution == nil {
			attribution = source
		}
		writePointer(target, 9, 0x4e0bde, attribution)
		write(target, 10, 4, 0x4e0be4, uint32(typ))
		frame := rawDefaultDamageNativeFrame4E0B30(0x4e0bea, 4)
		write(target, 13, 4, 0x4e0bf1, frame)
		return 1
	}
	// EDI keeps the original source pointer. The PE caller's source slot is
	// later reused as p32 scratch; it is not used to restore EDI afterwards.
	if service(0x4e0c05, word(1)).Word == 0 {
		owner := service(0x4e0c16, object(source)).Pointer
		if owner != nil && read(owner, 1, 1, 0x4e0c24)&6 != 0 {
			if service(0x4e0c2c, object(target), object(owner)).Word == 0 {
				if target != owner {
					return 1
				}
				if service(0x4e0c45, word(4096)).Word != 0 {
					return 1
				}
			}
		}
	}
	if source != nil && weapon != nil {
		if service(0x4e0c63, object(target), object(source)).Word == 0 {
			if read(target, 1, 1, 0x4e0c6f)&6 != 0 {
				if rawDefaultDamageInvokeAttack4E0B30(0x4e0c77, source, weapon) != 0 {
					if rawDefaultDamageInvokeFriendly4E0B30(0x4e0c84, weapon) == 0 {
						return 1
					}
				}
			}
		}
	}
	// This BYTE read is reached only after both child predicates above.
	if read(target, 3, 1, 0x4e0c94)&2 != 0 {
		return 1
	}
	if source != nil && service(0x4e0ca5, object(target), word(22)).Word != 0 {
		if read(source, 1, 1, 0x4e0cb1)&6 != 0 && weapon != nil {
			if rawDefaultDamageInvokeAttack4E0B30(0x4e0cbd, source, weapon) != 0 {
				service(0x4e0cd3, word(135), object(source), word(0), word(0))
				service(0x4e0cdb, object(target), word(22))
				coefficient := rawDefaultDamageFloatFromReply4E0B30(service(0x4e0cec, text("ShockDamage"), word(4)))
				spill := rawDefaultDamageFloatStore32Bits4E0B30(coefficient, cw)
				shock := rawDefaultDamageConvertNative4E0B30(0x4e0cf8, spill, cw)
				callback := readPointer(source, 15, 0x4e0d05)
				rawDefaultDamageBeforeNativeCallback4E0B30(0x4e0d05, callback)
				rawDefaultDamageInvokeService4E0B30(rawDefaultDamageNativeService4E0B30{
					Indirect: true, Site: 0x4e0d05, Callback: callback,
					Arguments: []rawDefaultDamageNativeArgument4E0B30{object(source), object(target), object(nil), word(uint32(shock)), word(9)},
				})
				if read(source, 1, 1, 0x4e0d0b)&4 != 0 {
					service(0x4e0d18, object(source), word(23))
				}
			}
		}
	}
	cachedType := typ
	if read(target, 1, 1, 0x4e0d20)&2 != 0 {
		subclass := read(target, 2, 4, 0x4e0d29)
		if subclass&0x200 != 0 && cachedType == 5 {
			return 1
		}
		if subclass&0x400 != 0 {
			if cachedType == 1 || cachedType == 12 {
				return 1
			}
			if cachedType == 7 {
				assignDamage(0x4e0d63, damage/2)
			}
		} else if subclass&0x800 != 0 && (cachedType == 9 || cachedType == 17) {
			return 1
		}
	}
	if cachedType == 1 || cachedType == 12 || cachedType == 7 {
		protection := rawDefaultDamageFloatFromReply4E0B30(service(0x4e0d9d, object(target)))
		p32 := rawDefaultDamageFloatStore32Bits4E0B30(protection, cw)
		rawDefaultDamageNativeConstant4E0B30(0x4e0da6, 0)
		if rawDefaultDamageFloatOrderedNonzero4E0B30(protection) && rawDefaultDamageNativeFrame4E0B30(0x4e0db6, 1)&3 == 0 {
			service(0x4e0dc5, word(104), object(target), word(0), word(0))
		}
		rawDefaultDamageNativeConstant4E0B30(0x4e0dcd, 0x3f800000)
		product := rawDefaultDamageFloatProtectionProduct4E0B30(p32, uint32(damage), cw)
		assignDamage(0x4e0de7, rawDefaultDamageConvertNative4E0B30(0x4e0ddf, product, cw))
		if damage == 0 {
			assignDamage(0x4e0def, 1)
		}
	}
	if cachedType == 9 || cachedType == 17 {
		protection := rawDefaultDamageFloatFromReply4E0B30(service(0x4e0e06, object(target)))
		p32 := rawDefaultDamageFloatStore32Bits4E0B30(protection, cw)
		rawDefaultDamageNativeConstant4E0B30(0x4e0e0f, 0)
		if rawDefaultDamageFloatOrderedNonzero4E0B30(protection) && rawDefaultDamageNativeFrame4E0B30(0x4e0e1f, 1)&3 == 0 {
			service(0x4e0e2e, word(108), object(target), word(0), word(0))
		}
		rawDefaultDamageNativeConstant4E0B30(0x4e0e36, 0x3f800000)
		product := rawDefaultDamageFloatProtectionProduct4E0B30(p32, uint32(damage), cw)
		assignDamage(0x4e0e50, rawDefaultDamageConvertNative4E0B30(0x4e0e48, product, cw))
		if damage == 0 {
			assignDamage(0x4e0e58, 1)
		}
		class := read(target, 1, 4, 0x4e0e60)
		if class&4 != 0 {
			update := readPointer(target, 18, 0x4e0e67)
			write(update, 20, 2, 0x4e0e6d, 2)
		} else if class&2 != 0 && read(target, 2, 1, 0x4e0e7c)&0x10 != 0 {
			update := readPointer(target, 18, 0x4e0e82)
			write(update, 22, 1, 0x4e0e88, 2)
		}
	}
	if source != nil {
		if weapon != nil {
			if read(weapon, 1, 4, 0x4e0ea1)&0x1001000 != 0 {
				x := read(source, 6, 4, 0x4e0eaa)
				write(target, 11, 4, 0x4e0ead, x)
				y := read(source, 7, 4, 0x4e0eb3)
				write(target, 12, 4, 0x4e0eb6, y)
			} else {
				x := read(weapon, 6, 4, 0x4e0ebe)
				write(target, 11, 4, 0x4e0ec1, x)
				y := read(weapon, 7, 4, 0x4e0ec7)
				write(target, 12, 4, 0x4e0eca, y)
			}
			if weapon != source && read(target, 1, 1, 0x4e0ed4)&2 != 0 {
				update := readPointer(target, 18, 0x4e0eda)
				write(update, 24, 4, 0x4e0ee2, 1)
				typeIndex := read(weapon, 0, 2, 0x4e0eec)
				write(update, 23, 4, 0x4e0f2a, typeIndex)
			}
		} else {
			x := read(source, 6, 4, 0x4e0ef2)
			write(target, 11, 4, 0x4e0ef8, x)
			y := read(source, 7, 4, 0x4e0efe)
			write(target, 12, 4, 0x4e0f01, y)
			if (cachedType == 10 || cachedType == 2) && read(target, 1, 1, 0x4e0f0e)&2 != 0 {
				update := readPointer(target, 18, 0x4e0f14)
				write(update, 24, 4, 0x4e0f1c, 1)
				typeIndex := read(source, 0, 2, 0x4e0f26)
				write(update, 23, 4, 0x4e0f2a, typeIndex)
			}
		}
		service(0x4e0f33, object(target), word(0))
	} else {
		write(target, 11, 4, 0x4e0f40, 0)
		write(target, 12, 4, 0x4e0f46, 0)
		if cachedType == 12 {
			service(0x4e0f51, object(target), word(0))
		}
	}
	class := read(target, 1, 4, 0x4e0f5d)
	if class&4 != 0 || class&2 != 0 && read(target, 2, 1, 0x4e0f68)&0x10 != 0 {
		service(0x4e0f77, object(target), object(source), object(weapon), damageArgument(), word(uint32(cachedType)))
	}
	attribution := weapon
	if attribution == nil {
		attribution = source
	}
	if weapon != nil {
		writePointer(target, 9, 0x4e0f83, attribution)
	} else {
		writePointer(target, 9, 0x4e0f8b, attribution)
	}
	cachedClass := read(target, 1, 1, 0x4e0f91)
	write(target, 10, 4, 0x4e0f94, uint32(cachedType))
	frame := rawDefaultDamageNativeFrame4E0B30(0x4e0f9a, 4)
	write(target, 13, 4, 0x4e0fa2, frame)
	if cachedClass&2 != 0 {
		update := readPointer(target, 18, 0x4e0faa)
		status := read(update, 21, 4, 0x4e0fb0)
		marker := read(update, 24, 4, 0x4e0fb6)
		write(update, 21, 4, 0x4e0fc1, status|0x200)
		if marker == 0 {
			write(update, 24, 4, 0x4e0fc9, 2)
			write(update, 23, 4, 0x4e0fd3, uint32(cachedType))
		}
	}
	if weapon != nil && read(weapon, 1, 4, 0x4e0fdd)&0x1001000 != 0 {
		service(0x4e0fef, object(target), object(source), object(weapon), damageArgument(), word(uint32(cachedType)))
	}
	suppressSound := false
	if target == weapon && read(target, 1, 4, 0x4e0ffb)&0x1001000 != 0 {
		suppressSound = true
	}
	if !suppressSound && source != nil && read(source, 1, 1, 0x4e1008)&2 != 0 {
		if readPointer(source, 18, 0x4e100e) != nil {
			soundSet := service(0x4e1019, object(source)).Pointer
			if soundSet != nil {
				soundID := read(soundSet, 26, 4, 0x4e1025)
				if soundID != 0 && int32(service(0x4e102d, word(soundID)).Word) > 0 {
					suppressSound = true
				}
			}
		}
	}
	if !suppressSound {
		callback := readPointer(target, 16, 0x4e1039)
		other := weapon
		if other == nil {
			other = source
		}
		if callback != nil {
			site := uint32(0x4e1049)
			if weapon == nil {
				site = 0x4e104f
			}
			rawDefaultDamageBeforeNativeCallback4E0B30(site, callback)
			rawDefaultDamageInvokeService4E0B30(rawDefaultDamageNativeService4E0B30{Indirect: true, Site: site, Callback: callback, Arguments: []rawDefaultDamageNativeArgument4E0B30{object(target), object(other)}})
		} else {
			service(0x4e105c, object(target), object(other))
		}
	}
	if source != nil && read(target, 1, 1, 0x4e106c)&6 != 0 && service(0x4e1079, object(source), word(13)).Word != 0 {
		service(0x4e1097, word(163), object(weapon), word(0), word(0))
		power := service(0x4e109f, object(source), word(13)).Word
		index := (power & 0xff) - 1
		coefficient := rawDefaultDamageFloatFromReply4E0B30(service(0x4e10b0, text("VampirismCoeff"), word(index)))
		product := rawDefaultDamageFloatStore32Bits4E0B30(rawDefaultDamageFloatMultiplyDamage4E0B30(coefficient, damage, cw), cw)
		amount := uint32(rawDefaultDamageConvertNative4E0B30(0x4e10bf, product, cw))
		if uint16(amount) < 1 {
			amount = 1
		}
		service(0x4e10de, object(source), word(amount&0xffff))
		var coordinates [4]uint32
		x := read(source, 4, 4, 0x4e10e3)
		coordinates[0] = uint32(rawDefaultDamageConvertNative4E0B30(0x4e10e7, x, cw))
		y := read(source, 5, 4, 0x4e10f0)
		sourceY := uint32(rawDefaultDamageConvertNative4E0B30(0x4e10f4, y, cw))
		x = read(target, 4, 4, 0x4e10f9)
		coordinates[1] = sourceY
		targetX := uint32(rawDefaultDamageConvertNative4E0B30(0x4e1101, x, cw))
		y = read(target, 5, 4, 0x4e1106)
		coordinates[2] = targetX
		coordinates[3] = uint32(rawDefaultDamageConvertNative4E0B30(0x4e110e, y, cw))
		service(0x4e1122, word(162), rawDefaultDamageNativeArgument4E0B30{Kind: "coordinates", Coordinates: coordinates}, word(amount))
	}
	service(0x4e1131, object(source), object(target), word(uint32(damage)))
	if read(target, 1, 1, 0x4e1136)&4 != 0 && damage >= 20 {
		update := readPointer(target, 18, 0x4e1147)
		state := read(update, 19, 1, 0x4e114d)
		if state != 1 && state != 15 {
			service(0x4e115b, object(target), word(30))
		}
	}
	if service(0x4e1168, word(6144)).Word == 1 {
		service(0x4e117c, object(source), object(target), damageArgument())
	}
	if source != nil {
		var immediateMonster unsafe.Pointer
		if read(source, 1, 1, 0x4e1188)&2 != 0 {
			immediateMonster = source
		} else {
			owner := readPointer(source, 8, 0x4e1192)
			if owner != nil && read(owner, 1, 1, 0x4e119c)&2 != 0 {
				immediateMonster = owner
			}
		}
		if immediateMonster != nil && service(0x4e11aa, object(target), object(immediateMonster)).Word != 0 {
			service(0x4e11b7, object(immediateMonster))
		}
	}
	if service(0x4e11c2, object(target), word(26)).Word != 0 && typ != 5 {
		other := weapon
		if other == nil {
			other = source
		}
		if typ != 15 || source != target {
			service(0x4e11f2, object(target), damageArgument(), word(uint32(typ)), object(other))
		}
		if damage == 0 {
			return 0
		}
	}
	service(0x4e1212, object(target), word(uint32(damage)))
	return 1
}
