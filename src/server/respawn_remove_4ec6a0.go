package server

// RespawnRemoveHooks4EC6A0 separates the pointer-width-independent GAME.EXE
// contract from the legacy allocator, record, and list storage.
type RespawnRemoveHooks4EC6A0[O comparable, R comparable, A any] struct {
	LoadHead      func() R
	LoadObject    func(R) O
	LoadNext      func(R) R
	LoadPrev      func(R) R
	StoreHead     func(R)
	StoreNext     func(R, R)
	StorePrev     func(R, R)
	LoadAllocator func() A
	FreeFirst     func(A, R)
}

// RespawnRemove4EC6A0 preserves GAME.EXE 004EC6A0, including its head
// fast-path and live link reloads. In particular, the first head object's
// identity is read before the head is checked for zero.
func RespawnRemove4EC6A0[O comparable, R comparable, A any](obj O, hooks RespawnRemoveHooks4EC6A0[O, R, A]) {
	var zero R
	head := hooks.LoadHead()
	if hooks.LoadObject(head) == obj {
		firstNext := hooks.LoadNext(head)
		hooks.StoreHead(firstNext)
		secondNext := hooks.LoadNext(head)
		if secondNext != zero {
			hooks.StorePrev(secondNext, zero)
		}
		allocator := hooks.LoadAllocator()
		hooks.FreeFirst(allocator, head)
		return
	}
	if head == zero {
		return
	}

	rec := head
	for {
		if hooks.LoadObject(rec) == obj {
			break
		}
		rec = hooks.LoadNext(rec)
		if rec == zero {
			return
		}
	}

	firstPrev := hooks.LoadPrev(rec)
	if firstPrev != zero {
		firstNext := hooks.LoadNext(rec)
		hooks.StoreNext(firstPrev, firstNext)
	}
	secondNext := hooks.LoadNext(rec)
	if secondNext != zero {
		secondPrev := hooks.LoadPrev(rec)
		hooks.StorePrev(secondNext, secondPrev)
	}
	allocator := hooks.LoadAllocator()
	hooks.FreeFirst(allocator, rec)
}
