package opennox

import "github.com/opennox/libs/types"

// Independent observer model for the stock unenchanted Troll: radial falloff
// truncates first, then DefaultDamage's EXPLOSION fire branch promotes zero to
// one. A near-80 edge hit is still admitted; outside/occluded objects are not.
func e2eMeteorShowerImpactDamage(raw int32, from, to types.Pointf, unoccluded bool) int32 {
	amount, inside := e2eMeteorRadialDamage(raw, from, to)
	if !inside || !unoccluded {
		return 0
	}
	if amount == 0 {
		return 1
	}
	return amount
}
