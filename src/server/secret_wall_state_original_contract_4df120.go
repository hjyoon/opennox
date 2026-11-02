package server

// SecretWallPacketRuntime4DF120 exposes the three original service boundaries.
// The reference contract precedes the native implementation. Traversal is live;
// callers must not substitute a player snapshot for FirstUnit/NextUnit.
type SecretWallPacketRuntime4DF120 struct {
	FirstUnit func() *Object
	NextUnit  func(*Object) *Object
	Send      func(int, []byte, *Object, int) int
}
