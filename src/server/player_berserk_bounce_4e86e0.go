package server

type playerBerserkBounceHooks4E86E0[O comparable] struct {
	mass     func(O) float32
	velocity func(O, int) float32
	store    func(O, int, float32)
}

// Keep each x87 arithmetic instruction at an explicit binary64 boundary. The
// Win32 executable uses the x87 53-bit precision mode here; separate helpers
// also prevent a target compiler from contracting multiply-plus-add to FMA.
//
//go:noinline
func playerBerserkAdd64_4E86E0(a, b float64) float64 { return a + b }

//go:noinline
func playerBerserkSub64_4E86E0(a, b float64) float64 { return a - b }

//go:noinline
func playerBerserkMul64_4E86E0(a, b float64) float64 { return a * b }

//go:noinline
func playerBerserkDiv64_4E86E0(a, b float64) float64 { return a / b }

// playerBerserkBounce4E86E0 preserves the two nil guards, live velocity load
// order, binary32 temporaries, and final X/Y store order of GAME.EXE 004E86E0.
// float64 models the Win32 x87 53-bit precision mode between explicit dword
// stores; values written to objects are rounded back to binary32.
func playerBerserkBounce4E86E0[O comparable](
	player, other O,
	hooks playerBerserkBounceHooks4E86E0[O],
) {
	var zero O
	if player == zero || other == zero {
		return
	}

	playerMass := hooks.mass(player)
	otherMass := hooks.mass(other)
	sum := float32(playerBerserkAdd64_4E86E0(float64(otherMass), float64(playerMass)))
	playerCoeff := playerBerserkDiv64_4E86E0(
		playerBerserkSub64_4E86E0(float64(playerMass), float64(otherMass)),
		float64(sum),
	)
	otherTwiceCoeff := float32(playerBerserkDiv64_4E86E0(
		playerBerserkAdd64_4E86E0(float64(otherMass), float64(otherMass)),
		float64(sum),
	))

	playerY := playerBerserkAdd64_4E86E0(
		playerBerserkMul64_4E86E0(float64(hooks.velocity(player, 1)), playerCoeff),
		playerBerserkMul64_4E86E0(float64(otherTwiceCoeff), float64(hooks.velocity(other, 1))),
	)
	otherCoeff := playerBerserkDiv64_4E86E0(
		playerBerserkSub64_4E86E0(float64(otherMass), float64(playerMass)),
		float64(sum),
	)
	playerTwiceCoeff := playerBerserkDiv64_4E86E0(
		playerBerserkAdd64_4E86E0(float64(playerMass), float64(playerMass)),
		float64(sum),
	)

	otherX := float32(playerBerserkAdd64_4E86E0(
		playerBerserkMul64_4E86E0(otherCoeff, float64(hooks.velocity(other, 0))),
		playerBerserkMul64_4E86E0(playerTwiceCoeff, float64(hooks.velocity(player, 0))),
	))
	otherY := float32(playerBerserkAdd64_4E86E0(
		playerBerserkMul64_4E86E0(otherCoeff, float64(hooks.velocity(other, 1))),
		playerBerserkMul64_4E86E0(playerTwiceCoeff, float64(hooks.velocity(player, 1))),
	))
	playerX := float32(playerBerserkAdd64_4E86E0(
		playerBerserkMul64_4E86E0(float64(hooks.velocity(player, 0)), playerCoeff),
		playerBerserkMul64_4E86E0(float64(otherTwiceCoeff), float64(hooks.velocity(other, 0))),
	))

	hooks.store(player, 0, playerX)
	hooks.store(player, 1, float32(playerY))
	hooks.store(other, 1, otherY)
	hooks.store(other, 0, otherX)
}
