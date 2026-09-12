// Package noxscriptqueue preserves the fixed PE32 deferred-script callback
// queue's insertion and removal order without constraining native pointers.
package noxscriptqueue

const Capacity = 32

// Append implements the 005025A0 bounded insertion. A full queue is unchanged.
func Append[T any](queue []T, entry T) []T {
	if len(queue) >= Capacity {
		return queue
	}
	return append(queue, entry)
}

// Remove implements the 005025E0 forward scan. After a match is shifted out,
// the scan advances, deliberately skipping an adjacent match moved into its slot.
// The now-unused tail slot is left intact, as in the original fixed array.
func Remove[T comparable](queue []T, entry T) []T {
	for i := 0; i < len(queue); i++ {
		if queue[i] != entry {
			continue
		}
		copy(queue[i:], queue[i+1:])
		queue = queue[:len(queue)-1]
	}
	return queue
}
