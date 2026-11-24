package server

// DoorUpdateRuntime53AC50 keeps the audio and collision-queue operations at
// their existing boundaries while the per-frame door state runs in Go.
type DoorUpdateRuntime53AC50 struct {
	AudioEvent func(uint32, *Object)
	QueueDoor  func(*DoorUpdateData)
	WakeDoor   func(*Object) byte
}

// DoorUpdate53AC50 restores GAME.EXE 0053AC50 using native-width Object and
// DoorUpdateData pointers. It returns the legacy result byte for ABI callers.
func (s *Server) DoorUpdate53AC50(door *Object, runtime DoorUpdateRuntime53AC50) byte {
	if door == nil || door.UpdateData == nil {
		return 0
	}
	update := door.UpdateDataDoor()
	target := update.TargetDirection
	current := update.CurrentDirection
	if update.SyncedDirection == target {
		if current != target && runtime.AudioEvent != nil {
			runtime.AudioEvent(doorUpdateSound53AC50(door, false), door)
		}
	} else if current == target {
		if runtime.AudioEvent != nil {
			runtime.AudioEvent(doorUpdateSound53AC50(door, true), door)
		}
		s.Objs.RemoveFromUpdatable(door)
	}

	// Audio and removal are runtime calls; keep the original live read here.
	current = update.CurrentDirection
	if update.SyncedDirection != current {
		door.NeedSync()
		current = update.CurrentDirection
		update.SyncedDirection = current
	}
	if uint32(door.ObjFlags)&0x1000000 == 0 ||
		s.Frame()-update.LastMoveFrame <= uint32(int32(s.TickRate())>>1) {
		return byte(current)
	}
	target = update.TargetDirection
	if current == target {
		return byte(current)
	}
	delta := int32(-2)
	diff := current - target
	if diff < 0 {
		diff += 32
	}
	if diff >= 16 {
		delta = 2
	}
	direction := (int32(update.FractionalDir) + delta) % 256
	if direction < 0 {
		direction += 256
	}
	update.FractionalDir = int16(direction)
	if runtime.QueueDoor != nil {
		runtime.QueueDoor(update)
	}
	if runtime.WakeDoor != nil {
		return runtime.WakeDoor(door)
	}
	return byte(current)
}

func doorUpdateSound53AC50(door *Object, reachedTarget bool) uint32 {
	// GAME.EXE selects door sounds from ObjSubClass at offset 0x0c.
	subclass := uint32(door.ObjSubClass)
	switch {
	case subclass&4 != 0:
		if door.Material&8 != 0 {
			if reachedTarget {
				return 246
			}
			return 245
		}
		if reachedTarget {
			return 243
		}
		return 241
	case subclass&1 != 0:
		if reachedTarget {
			return 248
		}
		return 247
	case subclass&0x1000 != 0:
		if reachedTarget {
			return 1015
		}
		return 1014
	default:
		if reachedTarget {
			return 239
		}
		return 237
	}
}
