package server

import "github.com/opennox/libs/strman"

// DirectExperienceGrantRuntime4EF3A0 supplies fixed-width protection and the
// two object-bearing callbacks that remain outside the native Server adapter.
type DirectExperienceGrantRuntime4EF3A0 struct {
	ProtectExperience func(uint32, float32)
	SendLineMessage   func(*Object, string, uint32)
	SyncLevel         func(*Object)
}

type directExperienceGrantNativeDeps4EF3A0 struct {
	protectExperience func(uint32, float32)
	reportExperience  func(*Object)
	loadString        func(string, string, int) string
	sendLineMessage   func(*Object, string, uint32)
	syncLevel         func(*Object)
}

func directExperienceGrantNative4EF3A0(
	unit *Object,
	award float32,
	deps directExperienceGrantNativeDeps4EF3A0,
) {
	directExperienceGrant4EF3A0(directExperienceGrantHooks4EF3A0[
		*Object,
		*PlayerUpdateData,
		*Player,
		string,
	]{
		loadAwardArg: func() float32 {
			return award
		},
		loadUnitArg: func() *Object {
			return unit
		},
		loadExperience: func(unit *Object) float32 {
			return unit.Experience
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		storeExperience: func(unit *Object, experience float32) {
			unit.Experience = experience
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadExperienceToken: func(player *Player) uint32 {
			return player.ProtUnitExperience
		},
		protectExperience: deps.protectExperience,
		reportExperience:  deps.reportExperience,
		loadString:        deps.loadString,
		sendLineMessage:   deps.sendLineMessage,
		syncLevel:         deps.syncLevel,
	})
}

func directExperienceGrantServerDeps4EF3A0(
	s *Server,
	runtime DirectExperienceGrantRuntime4EF3A0,
) directExperienceGrantNativeDeps4EF3A0 {
	return directExperienceGrantNativeDeps4EF3A0{
		protectExperience: runtime.ProtectExperience,
		reportExperience:  s.NetReportExperience,
		loadString: func(key, path string, line int) string {
			_ = line // retained by the generic provenance contract
			return s.Strings().GetStringInFile(strman.ID(key), path)
		},
		sendLineMessage: runtime.SendLineMessage,
		syncLevel:       runtime.SyncLevel,
	}
}

// DirectExperienceGrant4EF3A0 binds GAME.EXE 004EF3A0 to native-width
// Object, PlayerUpdateData, and Player layouts. Protection values and message
// points stay fixed-width while every object-bearing service keeps native
// pointer width.
func (s *Server) DirectExperienceGrant4EF3A0(
	unit *Object,
	award float32,
	runtime DirectExperienceGrantRuntime4EF3A0,
) {
	directExperienceGrantNative4EF3A0(
		unit,
		award,
		directExperienceGrantServerDeps4EF3A0(s, runtime),
	)
}
