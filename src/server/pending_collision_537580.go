package server

import "unsafe"

// PendingCollisionQueue537580 owns the original collision links at native
// width. Field115 retains the legacy zero/removed marker; links never store
// truncated object addresses in that DWORD.
type PendingCollisionQueue537580 struct {
	head, tail *Object
	links      map[*Object]*Object
}

type PendingCollisionRuntime537610 struct {
	TypeByName         func(string) uint32
	Fist, UndeadKiller unsafe.Pointer
	PreparePath        func()
}

func (q *PendingCollisionQueue537580) Head() *Object             { return q.head }
func (q *PendingCollisionQueue537580) Tail() *Object             { return q.tail }
func (q *PendingCollisionQueue537580) Empty() bool               { return q.head == nil }
func (q *PendingCollisionQueue537580) Contains(o *Object) uint32 { return o.Field116 & 1 }

// Next reports whether a native link exists, including a known nil tail link.
// The C adapter retains raw DWORD results for objects outside this owner.
func (q *PendingCollisionQueue537580) Next(o *Object) (*Object, bool) {
	if o == nil {
		return nil, true
	}
	next, ok := q.links[o]
	return next, ok
}

func (q *PendingCollisionQueue537580) Remove(o *Object) {
	if q.Contains(o) == 0 {
		return
	}
	var previous *Object
	current := q.head
	for current != nil && current != o {
		previous, current = current, q.links[current]
	}
	if current == nil {
		return
	}
	next := q.links[o]
	if previous != nil {
		q.links[previous] = next
		previous.Field115 = 0
	} else {
		q.head = next
	}
	if q.tail == o {
		q.tail = previous
	}
	delete(q.links, o)
	o.Field115 = 0xffffffff
	o.Field116 &^= 1
}

// Pop requires a nonempty queue, as does original 00537700. Callers test the
// head before draining; an empty call retains the original fault boundary.
func (q *PendingCollisionQueue537580) Pop() *Object {
	o := q.head
	next := q.links[o]
	q.head = next
	if next == nil {
		q.tail = nil
	}
	delete(q.links, o)
	o.Field115 = 0xffffffff
	o.Field116 &^= 1
	return o
}

func (q *PendingCollisionQueue537580) append(o *Object) {
	if q.links == nil {
		q.links = make(map[*Object]*Object)
	}
	if q.tail != nil {
		q.links[q.tail] = o
		q.tail.Field115 = 0
	} else {
		q.head = o
	}
	q.tail = o
	q.links[o] = nil
	o.Field115 = 0
	o.Field116 |= 1
}

// Enqueue preserves the original gates, live reads after both type lookups,
// full lookup result words, low-byte return and path reset ordering.
func (q *PendingCollisionQueue537580) Enqueue(o *Object, runtime PendingCollisionRuntime537610) byte {
	result := byte(uintptr(o.Update))
	if o.Update == nil {
		result = byte(uintptr(o.Collide))
		if o.Collide == nil || uint32(o.ObjFlags)&0x40 != 0 {
			return result
		}
	}
	class := uint32(o.ObjClass)
	if class&0x400000 != 0 || uint32(o.ObjFlags)&8 != 0 {
		spike := runtime.TypeByName("Spike")
		periodic := runtime.TypeByName("PeriodicSpike")
		result = byte(periodic)
		class = uint32(o.ObjClass)
		if class&0xe080 == 0 && o.Collide != runtime.Fist && o.Collide != runtime.UndeadKiller && uint32(o.TypeInd) != spike && uint32(o.TypeInd) != periodic {
			return result
		}
	}
	if uint32(o.ObjFlags)&4 == 0 {
		return result
	}
	if class&0x2008 != 0 {
		runtime.PreparePath()
	}
	// Original 005485F0 is RET; it has no state or return-value effect here.
	result = byte(o.Field116)
	if result&1 == 0 {
		q.append(o)
		result = byte(o.Field116)
	}
	return result
}
