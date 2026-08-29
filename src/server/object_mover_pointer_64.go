//go:build amd64 || arm64

package server

func moverWaypointFromPE32(uint32) *Waypoint { return nil }

func moverObjectFromPE32(uint32) *Object { return nil }
