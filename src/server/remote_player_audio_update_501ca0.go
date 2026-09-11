package server

const (
	remotePlayerAudioFollowMask501CA0    = uint8(0x03)
	remotePlayerAudioPlayerClass501CA0   = uint8(0x04)
	remotePlayerAudioUninitialized501CA0 = uint32(0xDEADFACE)
	remotePlayerAudioFirstPhoneme501CA0  = int32(186)
	remotePlayerAudioLastPhoneme501CA0   = int32(193)
)

// remotePlayerAudioHooks501CA0 describes every observable load and call made
// by GAME.EXE 00501CA0. Pointer-shaped values are kept as comparable generic
// handles so neither the control flow nor the tests can accidentally narrow
// them to the original executable's 32-bit representation.
type remotePlayerAudioHooks501CA0[O, U, P, G, T, E, Q comparable] struct {
	loadUpdate           func(O) U
	loadPlayer           func(U) P
	loadPlayerFlagsLow   func(P) uint8
	loadCameraTarget     func(P) O
	loadClassLow         func(O) uint8
	loadObjectPositionX  func(O) float32
	floatToInt           func(float32) int32
	loadObjectPositionY  func(O) float32
	polygonAtPoint       func([2]int32, uint32) G
	loadPolygonZone      func(G) uint8
	loadCurrentPolygonID func(P) uint32
	questCheckSecretArea func(O)
	loadPlayerAudioZone  func(P) uint8

	resetBitmap       func()
	hasTeam           func(O) bool
	loadTeamID        func(O) uint32
	teamByID          func(uint32) T
	firstEvent        func() E
	loadEventKind     func(E) int32
	loadEventCode     func(E) uint32
	loadObjectNetCode func(O) uint32
	loadEventObject   func(E) O
	eventPosition     func(E) Q
	eventZone         func(Q, O) uint8
	loadPhonemeState  func(U) uint8
	loadEventSound    func(E) int32
	playerPosition    func(P) Q
	fade              func(int32, Q, Q) int32
	loadSoundField20  func(int32) int32
	addAudio          func(E, int32)
	sendDirect        func(O, E, int32)
	loadEventNext     func(E) E
	flush             func(O)
}

func remotePlayerAudioHalfFade501CA0(value int32) int32 {
	return value >> 1
}

// remotePlayerAudioUpdate501CA0 preserves GAME.EXE 00501CA0's branch, reload,
// and callback order. In particular, the Player pointer is reloaded after the
// secret-area callback and before fade calculation, a non-player camera target
// is reloaded between its X and Y accesses, and an event's live next link is
// read only after all callbacks for that event have completed.
func remotePlayerAudioUpdate501CA0[O, U, P, G, T, E, Q comparable](
	unit O,
	hooks remotePlayerAudioHooks501CA0[O, U, P, G, T, E, Q],
) {
	var nilObject O
	var nilPolygon G
	var nilTeam T
	var nilEvent E

	update := hooks.loadUpdate(unit)
	player := hooks.loadPlayer(update)
	var listenerZone uint8
	if hooks.loadPlayerFlagsLow(player)&remotePlayerAudioFollowMask501CA0 == 0 {
		if hooks.loadCurrentPolygonID(player) == remotePlayerAudioUninitialized501CA0 {
			hooks.questCheckSecretArea(unit)
		}
		player = hooks.loadPlayer(update)
		listenerZone = hooks.loadPlayerAudioZone(player)
	} else {
		follow := hooks.loadCameraTarget(player)
		if follow == nilObject {
			if hooks.loadCurrentPolygonID(player) == remotePlayerAudioUninitialized501CA0 {
				hooks.questCheckSecretArea(unit)
			}
			player = hooks.loadPlayer(update)
			listenerZone = hooks.loadPlayerAudioZone(player)
		} else if hooks.loadClassLow(follow)&remotePlayerAudioPlayerClass501CA0 != 0 {
			followUpdate := hooks.loadUpdate(follow)
			followPlayer := hooks.loadPlayer(followUpdate)
			listenerZone = hooks.loadPlayerAudioZone(followPlayer)
		} else {
			follow = hooks.loadCameraTarget(player)
			x := hooks.loadObjectPositionX(follow)
			pointX := hooks.floatToInt(x)
			player = hooks.loadPlayer(update)
			follow = hooks.loadCameraTarget(player)
			y := hooks.loadObjectPositionY(follow)
			pointY := hooks.floatToInt(y)
			polygon := hooks.polygonAtPoint([2]int32{pointX, pointY}, 0)
			if polygon != nilPolygon {
				listenerZone = hooks.loadPolygonZone(polygon)
			}
		}
	}

	hooks.resetBitmap()
	var listenerTeam T
	if hooks.hasTeam(unit) {
		listenerTeam = hooks.teamByID(hooks.loadTeamID(unit))
	}

	for event := hooks.firstEvent(); event != nilEvent; event = hooks.loadEventNext(event) {
		eligible := true
		kind := hooks.loadEventKind(event)
		if kind == 1 {
			eventTeam := hooks.teamByID(hooks.loadEventCode(event))
			if listenerTeam == nilTeam || eventTeam == nilTeam || listenerTeam != eventTeam {
				eligible = false
			}
		} else if kind == 2 {
			if hooks.loadObjectNetCode(unit) != hooks.loadEventCode(event) {
				eligible = false
			}
		}

		if eligible {
			eventObject := hooks.loadEventObject(event)
			position := hooks.eventPosition(event)
			zone := hooks.eventZone(position, eventObject)
			if zone == listenerZone || zone == 0 {
				play := hooks.loadPhonemeState(update) != 0
				if !play {
					soundID := hooks.loadEventSound(event)
					play = soundID < remotePlayerAudioFirstPhoneme501CA0 || soundID > remotePlayerAudioLastPhoneme501CA0
					if !play {
						play = unit != hooks.loadEventObject(event)
					}
				}
				if play {
					player = hooks.loadPlayer(update)
					soundID := hooks.loadEventSound(event)
					listenerPosition := hooks.playerPosition(player)
					fade := remotePlayerAudioHalfFade501CA0(hooks.fade(soundID, position, listenerPosition))
					if fade > 0 {
						soundID = hooks.loadEventSound(event)
						if hooks.loadSoundField20(soundID) != 0 {
							hooks.addAudio(event, fade)
						} else {
							hooks.sendDirect(unit, event, fade)
						}
					}
				}
			}
		}
	}
	hooks.flush(unit)
}
