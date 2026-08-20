package server

type healthLinksNextHooks4EE440[O, H any] struct {
	loadHealth func(O) H
	loadNext   func(H) O
}

// healthLinksNext4EE440 preserves the null-object branch and the two exact
// memory reads performed by GAME.EXE 004EE440. A non-null object has no
// null-HealthData guard, so loadNext must still be called with a null record.
func healthLinksNext4EE440[O, H comparable](obj O, hooks healthLinksNextHooks4EE440[O, H]) O {
	var nilObject O
	if obj == nilObject {
		return nilObject
	}
	health := hooks.loadHealth(obj)
	return hooks.loadNext(health)
}
