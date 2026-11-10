package legacy

import "unsafe"

// These five dependencies used PE32 packed event offsets in a separate C file.
// Selection arrays are scalar DWORDs; sample and buffer results are pointers.
func game2AudioStart451CA0(e *game2AudioEvent) unsafe.Pointer {
	e.Remaining = e.SampleCount
	if e.SampleCount == 0 {
		return nil
	}
	if e.SampleCount > 32 {
		e.Remaining = 32
	} // malformed native state policy.
	for i := uint32(0); i < e.Remaining; i++ {
		e.Order[i] = i
	}
	e.Cursor = -1
	return game2AudioNext451CF0(e)
}
func game2AudioNext451CF0(e *game2AudioEvent) unsafe.Pointer {
	d := game2AudioDef(e.Definition)
	if d == nil {
		return nil
	}
	if e.Remaining != 0 {
		if e.Remaining > 32 {
			return nil
		}
		if d.word(1)&2 != 0 {
			i := int(game2AudioRandom(0, int32(e.Remaining)-1, 376))
			if i < 0 || i >= int(e.Remaining) {
				return nil
			}
			e.Cursor = int32(e.Order[i])
			copy(e.Order[i:], e.Order[i+1:e.Remaining])
		} else {
			e.Cursor++
		}
		e.Remaining--
		if e.Cursor < 0 || e.Cursor >= 32 {
			return nil
		}
		// The old handle+24 buffer is represented by the real C-owned sample
		// token; only the native backend may interpret it, never packed C.
		return e.Samples[e.Cursor]
	}
	if d.word(1)&1 != 0 {
		if d.word(15) != 0 {
			e.Loops++
			if int32(e.Loops) >= int32(d.word(15)) {
				return nil
			}
		}
		return game2AudioStart451CA0(e)
	}
	return nil
}
func game2AudioLoadSample451F30(e *game2AudioEvent, index int32) int32 {
	d := game2AudioDef(e.Definition)
	if d == nil || index < 0 || index >= 32 || e.SampleCount >= 32 {
		return 0
	}
	p := game2AudioSampleAt(d.sample(int(index)))
	e.Samples[e.SampleCount] = p
	if p != nil {
		game2AudioRetain(p)
		e.SampleCount++
		return int32(e.SampleCount)
	}
	return 0 // failed sample has no scalar pointer-as-result alias.
}
func game2AudioReleaseSamples451F90(e *game2AudioEvent) int32 {
	count := e.SampleCount
	if count > 32 {
		count = 32
	}
	for i := uint32(0); i < count; i++ {
		game2AudioRelease(e.Samples[i])
		e.Samples[i] = nil
	}
	e.SampleCount = 0
	return int32(count)
}
func game2AudioLoad451DC0(e *game2AudioEvent) int32 {
	d := game2AudioDef(e.Definition)
	if d == nil {
		return 0
	}
	if e.SampleCount != 0 {
		if d.word(17) < 33 {
			return int32(e.SampleCount)
		}
		game2AudioReleaseSamples451F90(e)
	}
	if d.word(1)&4 != 0 {
		if d.word(17) >= 33 {
			return game2AudioLoadSample451F30(e, game2AudioShuffle451E80(e))
		}
		count := int32(d.word(48))
		if count > 32 {
			count = 32
		}
		for i := int32(0); i < count; i++ {
			game2AudioLoadSample451F30(e, i)
		}
		return int32(d.word(48))
	}
	i := int32(0)
	if d.word(1)&2 != 0 {
		i = game2AudioRandom(0, int32(d.word(48))-1, 536)
	}
	return game2AudioLoadSample451F30(e, i)
}
func game2AudioShuffle451E80(e *game2AudioEvent) int32 {
	d := game2AudioDef(e.Definition)
	if d == nil {
		return 0
	}
	if int32(e.ShuffleRemaining) <= 0 {
		e.ShuffleRemaining = d.word(48)
		if e.ShuffleRemaining > 32 {
			return 0
		}
		for i := uint32(0); i < e.ShuffleRemaining; i++ {
			e.Shuffle[i] = e.ShuffleRemaining - i - 1
		}
	}
	if e.ShuffleRemaining == 0 {
		return 0
	}
	e.ShuffleRemaining--
	if d.word(1)&2 == 0 {
		return int32(e.Shuffle[e.ShuffleRemaining])
	}
	i := int(game2AudioRandom(0, int32(e.ShuffleRemaining), 431))
	if i < 0 || i > int(e.ShuffleRemaining) {
		return 0
	}
	v := e.Shuffle[i]
	copy(e.Shuffle[i:], e.Shuffle[i+1:e.ShuffleRemaining+1])
	return int32(v)
}
func game2AudioSetup452580(e *game2AudioEvent) int32 {
	d := game2AudioDef(e.Definition)
	if d == nil || d.word(48) == 0 {
		return 0
	}
	e.Loops = 0
	e.Voice = game2AudioAcquire452810(int32(d.word(12))+e.Category, 0)
	v := e.Voice
	if v == nil {
		return 0
	}
	pitch := game2AudioRandom(int32(d.word(19)), int32(d.word(20)), 1482)
	v.Group.Timers[1].SetRaw(uint32(pitch + 100))
	v.Flags |= 0x10
	v.Owner = e
	e.State = 1
	if d.word(1)&8 != 0 {
		delay := game2AudioRandom(int32(d.word(17)), int32(d.word(18)), 1497)
		if delay > 33 {
			game2AudioDelay452690(e, int64(delay), 1)
		}
	}
	return 1
}
func game2AudioNeedBuffer452770(v *game2AudioVoice) int32 {
	if v == nil || v.Owner == nil {
		return 0
	}
	e := v.Owner
	d := game2AudioDef(e.Definition)
	if d == nil {
		return 0
	}
	p := game2AudioNext451CF0(e)
	if d.word(18) < 33 {
		v.Buffer = p
		if p != nil {
			game2AudioVoicePlay(v)
		}
		return 0
	}
	v.Buffer = nil
	if d.word(1)&8 == 0 || p != nil || e.ShuffleRemaining != 0 {
		delay := game2AudioRandom(int32(d.word(17)), int32(d.word(18)), 706)
		if delay < 33 {
			v.Buffer = p
			if p != nil {
				game2AudioVoicePlay(v)
			}
			return 0
		}
		e.Delay = uint32(delay)
		e.Buffer = p
	}
	return 0
}
