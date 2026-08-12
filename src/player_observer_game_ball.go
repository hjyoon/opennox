package opennox

import "github.com/opennox/opennox/v1/server"

// playerObserverFindGameBall_4E6230 preserves the original cache and object
// list access order. The cached type ID is reloaded for every object and the
// object's 16-bit type index is compared after zero extension.
func playerObserverFindGameBall_4E6230(
	ensureGameBallID func(),
	gameBallID func() uint32,
	firstObject func() *server.Object,
	nextObject func(*server.Object) *server.Object,
) *server.Object {
	ensureGameBallID()
	for obj := firstObject(); obj != nil; obj = nextObject(obj) {
		id := gameBallID()
		if uint32(obj.TypeInd) == id {
			return obj
		}
	}
	return nil
}
