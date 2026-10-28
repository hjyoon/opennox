package server

type SparkUpdateRuntime53ADC0 struct {
	DelayedDelete func(*Object)
}

// SparkUpdate53ADC0 restores GAME.EXE 0053ADC0 without dereferencing the
// PE32 Object+748 update-data slot, which is a different field on 64-bit hosts.
func (s *Server) SparkUpdate53ADC0(obj *Object, runtime SparkUpdateRuntime53ADC0) {
	if obj == nil || obj.UpdateData == nil {
		return
	}
	SparkUpdateRaw53ADC0(obj, func(source *Object) {
		if runtime.DelayedDelete != nil {
			runtime.DelayedDelete(source)
		}
	})
}
