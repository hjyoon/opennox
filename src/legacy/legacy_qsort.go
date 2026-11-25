package legacy

// sortLegacyQsort follows GAME.EXE's qsort at 0056624F, including its order of
// comparisons and swaps when records compare equal.
func sortLegacyQsort[T any](entries []T, compare func(*T, *T) int) {
	for len(entries) > 8 {
		middle := len(entries) / 2
		entries[0], entries[middle] = entries[middle], entries[0]
		low, high := 0, len(entries)
		for {
			low++
			for low < len(entries) && compare(&entries[low], &entries[0]) <= 0 {
				low++
			}
			high--
			for high > 0 && compare(&entries[high], &entries[0]) >= 0 {
				high--
			}
			if high < low {
				break
			}
			entries[low], entries[high] = entries[high], entries[low]
		}
		entries[0], entries[high] = entries[high], entries[0]
		lower, upper := entries[:high], entries[low:]
		// Visit the smaller partition first. The original visits the upper
		// partition first when their lengths match.
		if len(lower) >= len(upper) {
			lower, upper = upper, lower
		}
		sortLegacyQsort(lower, compare)
		entries = upper
	}
	for end := len(entries) - 1; end > 0; end-- {
		maximum := 0
		for i := 1; i <= end; i++ {
			if compare(&entries[i], &entries[maximum]) > 0 {
				maximum = i
			}
		}
		entries[maximum], entries[end] = entries[end], entries[maximum]
	}
}
