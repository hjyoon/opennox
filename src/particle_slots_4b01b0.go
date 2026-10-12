package opennox

import "reflect"

// Keep the original logical arena in Go memory: particleFx contains Go
// callbacks/references and must never be allocated in C-owned storage.
// The owner's existing pool constructor holds the arena without a global
// owner registry. Native handles remain opaque, native-width lookup keys.
func (pfx *partFXes) particleSlots4B01B0() func() *particleFx {
	const block = 512 // 004AF8E5 initial capacity; 004B01D1 grows by 0x200.
	slots := make([]*particleFx, block)
	active := pfx.byHandle
	return func() *particleFx {
		// Init/Free replace the active map. Both map references stay live,
		// so identity comparison cannot mistake recycled map storage.
		if reflect.ValueOf(active).Pointer() != reflect.ValueOf(pfx.byHandle).Pointer() {
			active = pfx.byHandle
			next := make([]*particleFx, block)
			copy(next, slots)
			for _, p := range next {
				if p != nil {
					*p = particleFx{}
				}
			}
			slots = next
		}
		for i, p := range slots {
			if p == nil {
				p = new(particleFx)
				slots[i] = p
			}
			if p.chnd == nil || active[p.chnd] != p {
				return p
			}
		}
		index := len(slots)
		slots = append(slots, make([]*particleFx, block)...)
		p := new(particleFx)
		slots[index] = p
		return p
	}
}
