//go:build server

package legacy

import "image"

func AddSentryRay4C5020(from, to image.Point) {}

func SentryRayCount4C5020() int {
	return 0
}

func SentryRayAt4C5020(index int) (image.Point, image.Point, bool) {
	return image.Point{}, image.Point{}, false
}
