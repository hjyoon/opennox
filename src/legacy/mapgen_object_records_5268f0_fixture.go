package legacy

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"
#include "server__mapgen__generate__populate.h"

extern uint32_t dword_5d4594_2487672;
extern uint32_t dword_5d4594_2487676;

typedef struct nox_test_mapgen_object_records_5268f0_result {
	uintptr_t records_address;
	uintptr_t resolved_address;
	uintptr_t entries[3];
	uint32_t records_token;
	uint32_t token_after_cleanup;
	int indices[4];
	int lookup_after_cleanup;
	int setup_clean;
} nox_test_mapgen_object_records_5268f0_result;

static nox_test_mapgen_object_records_5268f0_result nox_test_mapgen_object_records_5268f0(void) {
	nox_test_mapgen_object_records_5268f0_result out = {0};
	if (dword_5d4594_2487672 != 0) {
		return out;
	}
	out.setup_clean = 1;
	char* records = (char*)calloc(3, 64);
	if (!records) {
		return out;
	}
	strcpy(records, "Alpha");
	strcpy(records + 64, "Beta");
	strcpy(records + 128, "Gamma");
	out.records_address = (uintptr_t)records;
	dword_5d4594_2487672 = nox_mapgenLegacyPtrRegister(records);
	dword_5d4594_2487676 = 3;
	out.records_token = dword_5d4594_2487672;
	if (!out.records_token) {
		free(records);
		dword_5d4594_2487676 = 0;
		return out;
	}
	out.resolved_address = (uintptr_t)nox_mapgenLegacyPtrResolve(dword_5d4594_2487672);
	for (int i = 0; i < 3; ++i) {
		out.entries[i] = (uintptr_t)sub_526AA0(i);
	}
	out.indices[0] = sub_5268F0("Alpha");
	out.indices[1] = sub_5268F0("Beta");
	out.indices[2] = sub_5268F0("Gamma");
	out.indices[3] = sub_5268F0("Missing");
	sub_526A90();
	out.token_after_cleanup = dword_5d4594_2487672;
	out.lookup_after_cleanup = sub_5268F0("Alpha");
	dword_5d4594_2487676 = 0;
	return out;
}
*/
import "C"

type mapgenObjectRecordsResult5268F0 struct {
	recordsAddress     uintptr
	resolvedAddress    uintptr
	entries            [3]uintptr
	recordsToken       uint32
	tokenAfterCleanup  uint32
	indices            [4]int
	lookupAfterCleanup int
	setupClean         bool
}

func mapgenObjectRecordsFixture5268F0() mapgenObjectRecordsResult5268F0 {
	got := C.nox_test_mapgen_object_records_5268f0()
	out := mapgenObjectRecordsResult5268F0{
		recordsAddress:     uintptr(got.records_address),
		resolvedAddress:    uintptr(got.resolved_address),
		recordsToken:       uint32(got.records_token),
		tokenAfterCleanup:  uint32(got.token_after_cleanup),
		lookupAfterCleanup: int(got.lookup_after_cleanup),
		setupClean:         got.setup_clean != 0,
	}
	for i := range out.entries {
		out.entries[i] = uintptr(got.entries[i])
	}
	for i := range out.indices {
		out.indices[i] = int(got.indices[i])
	}
	return out
}
