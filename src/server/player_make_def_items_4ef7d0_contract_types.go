//go:build ignore

package server

// This file is passed explicitly to `go test` together with
// player_make_def_items_4ef7d0.go and its semantic test. It keeps the pure
// contract matrix independent from the native server graph and CGo while
// retaining the exact state type and value consumed by the recovered body.
type PlayerState byte

const PlayerState13 = PlayerState(13)
