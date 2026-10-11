//go:build !server

// Package npcdrawprobe supplies a read-only C draw boundary for renderer tests.
// It copies opaque native render records; it knows no game/render field offsets.
// Production client code does not import this package.
package npcdrawprobe

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

typedef struct {
	const void* source;
	size_t size;
	unsigned count;
	unsigned char* snapshots[2];
	uintptr_t drawables[2];
} npc_palette_recorder;

static npc_palette_recorder* npc_palette_active;

static npc_palette_recorder* npc_palette_new(const void* source, size_t size) {
	if (!source || !size || npc_palette_active) return NULL;
	npc_palette_recorder* r = calloc(1, sizeof(*r));
	if (!r) return NULL;
	r->source = source;
	r->size = size;
	for (unsigned i = 0; i < 2; ++i) {
		r->snapshots[i] = malloc(size + 32);
		if (!r->snapshots[i]) {
			for (unsigned j = 0; j < i; ++j) free(r->snapshots[j]);
			free(r);
			return NULL;
		}
		memset(r->snapshots[i], 0x5a, size + 32);
	}
	npc_palette_active = r;
	return r;
}

static int npc_palette_draw(void* viewport, void* drawable) {
	npc_palette_recorder* r = npc_palette_active;
	if (!r || !viewport || !drawable) return 0;
	unsigned i = r->count++;
	if (i >= 2) return 0;
	memcpy(r->snapshots[i] + 16, r->source, r->size);
	r->drawables[i] = (uintptr_t)drawable;
	return 1;
}

static void* npc_palette_callback(void) { return (void*)npc_palette_draw; }
static void* npc_palette_snapshot(npc_palette_recorder* r, unsigned i) {
	return i < r->count && i < 2 ? r->snapshots[i] + 16 : NULL;
}
static uintptr_t npc_palette_drawable(npc_palette_recorder* r, unsigned i) {
	return i < r->count && i < 2 ? r->drawables[i] : 0;
}
static int npc_palette_guards(npc_palette_recorder* r) {
	if (r->count > 2) return 0;
	for (unsigned i = 0; i < 2; ++i) {
		for (unsigned j = 0; j < 16; ++j) {
			if (r->snapshots[i][j] != 0x5a || r->snapshots[i][16 + r->size + j] != 0x5a) return 0;
		}
	}
	return 1;
}
static void npc_palette_free(npc_palette_recorder* r) {
	if (!r) return;
	if (npc_palette_active == r) npc_palette_active = NULL;
	for (unsigned i = 0; i < 2; ++i) free(r->snapshots[i]);
	free(r);
}
*/
import "C"

import "unsafe"

type Recorder struct {
	p *C.npc_palette_recorder
}

func Callback() unsafe.Pointer {
	return C.npc_palette_callback()
}

func New(source unsafe.Pointer, size uintptr) *Recorder {
	p := C.npc_palette_new(source, C.size_t(size))
	if p == nil {
		panic("cannot allocate or nest NPC palette recorder")
	}
	return &Recorder{p: p}
}

func (r *Recorder) Close() {
	C.npc_palette_free(r.p)
	r.p = nil
}

func (r *Recorder) C() unsafe.Pointer {
	return unsafe.Pointer(r.p)
}

func (r *Recorder) Count() int {
	return int(r.p.count)
}

func (r *Recorder) Snapshot(i int) unsafe.Pointer {
	return C.npc_palette_snapshot(r.p, C.uint(i))
}

func (r *Recorder) Drawable(i int) uintptr {
	return uintptr(C.npc_palette_drawable(r.p, C.uint(i)))
}

func (r *Recorder) GuardsIntact() bool {
	return C.npc_palette_guards(r.p) != 0
}
