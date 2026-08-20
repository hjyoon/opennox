package server

const (
	sageSetMessage4EF4F0   = "sageset"
	sageUnsetMessage4EF4F0 = "sageunset"
)

// SageCommandRuntime4EF4F0 supplies the observations made by the two decoded
// parsecmd callers at 00442480 and 004424C0. SetSage is kept explicit even
// though the original callee is a no-op, so the caller order remains testable.
type SageCommandRuntime4EF4F0 struct {
	QuestMode  func() bool
	SetSage    func(uint32)
	LoadString func(string) string
	Print      func(string)
}

// SageNoop4EF4F0 preserves GAME.EXE 004EF4F0: one RET instruction. The pushed
// enable dword is never read and no game, memory, or callback state is touched.
// All three decoded callers ignore the unconstrained return register.
//
//go:noinline
func SageNoop4EF4F0(_ uint32) {}

// SageCommand4EF4F0 preserves the observable portions of the two original
// command callers. "set sage" checks Quest mode first and is completely quiet
// in Quest; otherwise it passes one to the no-op before loading and printing
// sageset. "unset sage" does not query Quest mode and passes zero before loading
// and printing sageunset. Neither caller inspects command arguments.
func SageCommand4EF4F0(enable bool, runtime SageCommandRuntime4EF4F0) {
	if enable {
		if runtime.QuestMode() {
			return
		}
		runtime.SetSage(1)
		message := runtime.LoadString(sageSetMessage4EF4F0)
		runtime.Print(message)
		return
	}

	runtime.SetSage(0)
	message := runtime.LoadString(sageUnsetMessage4EF4F0)
	runtime.Print(message)
}
