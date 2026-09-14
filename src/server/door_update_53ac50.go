package server

// DoorUpdateRuntime53AC50 keeps the audio and collision-queue operations at
// their existing boundaries while the per-frame door state runs in Go.
type DoorUpdateRuntime53AC50 struct {
	AudioEvent func(uint32, *Object)
	QueueDoor  func(*DoorUpdateData)
	WakeDoor   func(*Object)
}

// DoorUpdate53AC50 restores GAME.EXE 0053AC50 using native-width Object and
// DoorUpdateData pointers. The original callback remains registered as the
// thing.bin identity, but the object-update dispatcher invokes this directly.
func (s *Server) DoorUpdate53AC50(door *Object, runtime DoorUpdateRuntime53AC50) {
	if door == nil || door.UpdateData == nil {
		return
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
		update.SyncedDirection = current
	}
	if uint32(door.ObjFlags)&0x1000000 == 0 ||
		s.Frame()-update.LastMoveFrame <= s.TickRate()>>1 || current == target {
		return
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
		runtime.WakeDoor(door)
	}
}

func doorUpdateSound53AC50(door *Object, reachedTarget bool) uint32 {
	class := uint32(door.ObjClass)
	switch {
	case class&4 != 0:
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
	case class&1 != 0:
		if reachedTarget {
			return 248
		}
		return 247
	case class&0x1000 != 0:
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
