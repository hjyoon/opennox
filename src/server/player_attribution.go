package server

// The original Player fields at 3600..3608 retain the most recent opposing
// player responsible for an offensive action. Player death handling accepts
// this attribution for ten seconds, then clears the pending flag.
func (p *Player) SetLastAggressorPending(v uint32) {
	p.Field3600 = v
}

func (p *Player) SetLastAggressorPlayerIndex(v uint32) {
	p.Field3604 = v
}

func (p *Player) SetLastAggressorFrame(v uint32) {
	p.field3608 = v
}

func (p *Player) LastAggressorState() (pending, playerIndex, frame uint32) {
	return p.Field3600, p.Field3604, p.field3608
}
