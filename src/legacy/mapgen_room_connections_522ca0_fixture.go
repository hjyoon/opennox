package legacy

/*
#include <stdint.h>
#include <stdlib.h>

#include "GAME4_1.h"
#include "GAME4_2.h"
#include "mapgen_legacy_ptr.h"
#include "memmap.h"

typedef struct nox_test_mapgen_room_connections_522ca0_result {
	uintptr_t theme_address;
	uintptr_t room_address;
	uintptr_t room_at_cell_address;
	uintptr_t hall_addresses[4];
	uintptr_t hall_resolved_addresses[4];
	uintptr_t neighbor_addresses[2];
	uintptr_t linked_neighbor_addresses[2];
	uintptr_t legacy_center_return_address;
	uintptr_t waypoint_addresses[2];
	uintptr_t waypoint_link_addresses[2];
	uint32_t theme_token;
	uint32_t room_token;
	uint32_t hall_tokens[4];
	uint32_t waypoint_tokens[2];
	float room_point[2];
	float near_points[4][2];
	float far_points[4][2];
	float center_points[4][2];
	float legacy_near_point[2];
	float legacy_far_point[2];
	float legacy_center_point[2];
	float hall_bounds[4][4];
	uint8_t room_point_count;
	uint8_t waypoint_counts[2];
	uint8_t waypoint_kinds[2];
	int grid_initialized;
	int room_added;
	int add_first;
	int add_duplicate;
	int add_legacy_duplicate;
	int native_lines_returned_null;
	int native_points_returned_null;
	int legacy_lines_returned_null;
	int legacy_points_returned_null;
	int hall_links_added;
	int adjust_rejected;
	int waypoint_native_link;
	int waypoint_duplicate_rejected;
	int waypoint_legacy_link;
	int waypoint_self_rejected;
	int waypoint_capacity_rejected;
	int cleanup_head_is_null;
} nox_test_mapgen_room_connections_522ca0_result;

static nox_test_mapgen_room_connections_522ca0_result nox_test_mapgen_room_connections_522ca0(void) {
	nox_test_mapgen_room_connections_522ca0_result out = {0};
	uint8_t* theme = NULL;
	uint8_t* room = NULL;
	uint8_t* halls[4] = {0};
	uint8_t* neighbors[2] = {0};
	float2* legacy_point = NULL;
	nox_waypoint_t* waypoints[2] = {0};
	int room_added = 0;
	double* coordinate_epsilon = getMemDoublePtr(0x581450, 10432);
	double saved_coordinate_epsilon = *coordinate_epsilon;
	*coordinate_epsilon = 0.001;

	theme = (uint8_t*)calloc(1, 0x45C);
	if (!theme) {
		goto cleanup;
	}
	out.theme_address = (uintptr_t)theme;
	out.theme_token = nox_mapgenLegacyPtrRegister(theme);
	if (!out.theme_token) {
		goto cleanup;
	}
	*(uint32_t*)(theme + 68) = 32;
	out.grid_initialized = sub_520EA0(theme);
	if (!out.grid_initialized) {
		goto cleanup;
	}

	room = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(2, 2);
	if (!room) {
		goto cleanup;
	}
	out.room_address = (uintptr_t)room;
	out.room_token = nox_mapgenLegacyPtrRegister(room);
	float2 room_position = {0.0f, 0.0f};
	nox_xxx_mapGenSetRoomPos_521880((uint32_t*)room, &room_position);
	out.room_added = nox_xxx_mapGenAddNewRoom_521730((uint32_t*)room) != 0;
	if (!out.room_added) {
		goto cleanup;
	}
	room_added = 1;

	int2 room_cell = {0};
	nox_xxx_mapGenRoundFloatToPtr_520DF0((float2*)(room + 20), &room_cell);
	out.room_at_cell_address = (uintptr_t)nox_mapgenRoomAtCellNative_521290(&room_cell);
	out.native_lines_returned_null = nox_mapgenConnectRoomLinesNative_522D30(theme) == NULL;
	out.native_points_returned_null = nox_mapgenConnectRoomPointsNative_522F40(theme) == NULL;
	out.legacy_lines_returned_null = sub_522D30((int)out.theme_token) == NULL;
	out.legacy_points_returned_null =
		nox_xxx_mapGenTryNextRoom_522F40((uint32_t*)(uintptr_t)out.theme_token) == NULL;

	float2 point = {17.25f, -31.5f};
	out.add_first = nox_mapgenAddRoomPointNative_522CA0(room, &point);
	out.add_duplicate = nox_mapgenAddRoomPointNative_522CA0(room, &point);
	out.add_legacy_duplicate = sub_522CA0((int)out.room_token, &point.field_0);
	out.room_point_count = room[352];
	out.room_point[0] = *(float*)(room + 224);
	out.room_point[1] = *(float*)(room + 228);

	for (int i = 0; i < 4; ++i) {
		int type = i + 2;
		halls[i] = (uint8_t*)nox_mapgenMakeHallStructNative_523E30(type, 2, 4);
		if (!halls[i]) {
			goto cleanup;
		}
		out.hall_addresses[i] = (uintptr_t)halls[i];
		out.hall_tokens[i] = nox_mapgenLegacyPtrRegister(halls[i]);
		out.hall_resolved_addresses[i] =
			(uintptr_t)nox_mapgenLegacyPtrResolve(out.hall_tokens[i]);
		float2 position = {(float)(100 + i * 300), (float)(200 + i * 300)};
		nox_xxx_mapGenSetRoomPos_521880((uint32_t*)halls[i], &position);
		out.hall_bounds[i][0] = *(float*)(halls[i] + 36);
		out.hall_bounds[i][1] = *(float*)(halls[i] + 40);
		out.hall_bounds[i][2] = *(float*)(halls[i] + 44);
		out.hall_bounds[i][3] = *(float*)(halls[i] + 48);
		nox_mapgenConnectionPointNearNative_523C30(
			halls[i], (float2*)out.near_points[i]);
		nox_mapgenConnectionPointFarNative_523CB0(
			halls[i], (float2*)out.far_points[i]);
		nox_mapgenConnectionPointCenterNative_523D30(
			halls[i], (float2*)out.center_points[i]);
	}

	for (int i = 0; i < 2; ++i) {
		neighbors[i] = (uint8_t*)nox_xxx_mapGenMakeRoomStruct_521940(1, 1);
		if (!neighbors[i]) {
			goto cleanup;
		}
		out.neighbor_addresses[i] = (uintptr_t)neighbors[i];
		float2 position = {(float)(10000 + i * 100), (float)(10000 + i * 100)};
		nox_xxx_mapGenSetRoomPos_521880((uint32_t*)neighbors[i], &position);
	}
	out.hall_links_added = 1;
	for (int i = 0; i < 4; ++i) {
		int first_direction = i < 2 ? 0 : 2;
		out.hall_links_added = out.hall_links_added &&
			nox_mapgenRoomLinkNative_521900(halls[i], neighbors[0], first_direction) &&
			nox_mapgenRoomLinkNative_521900(halls[i], neighbors[1], first_direction + 1);
	}
	out.linked_neighbor_addresses[0] =
		(uintptr_t)nox_mapgenLegacyPtrResolve(*(uint32_t*)(halls[0] + 88));
	out.linked_neighbor_addresses[1] =
		(uintptr_t)nox_mapgenLegacyPtrResolve(*(uint32_t*)(halls[0] + 120));
	for (int i = 0; i < 4; ++i) {
		nox_mapgenConnectionPointCenterNative_523D30(
			halls[i], (float2*)out.center_points[i]);
	}
	out.adjust_rejected = !nox_mapgenAdjustHallNative_523A10(halls[0], (float*)neighbors[0]);

	legacy_point = (float2*)calloc(1, sizeof(*legacy_point));
	if (!legacy_point) {
		goto cleanup;
	}
	uint32_t legacy_point_token = nox_mapgenLegacyPtrRegister(legacy_point);
	if (!legacy_point_token) {
		goto cleanup;
	}
	sub_523C30((int)out.hall_tokens[0], (int)legacy_point_token);
	out.legacy_near_point[0] = legacy_point->field_0;
	out.legacy_near_point[1] = legacy_point->field_4;
	sub_523CB0((int)out.hall_tokens[0], (int)legacy_point_token);
	out.legacy_far_point[0] = legacy_point->field_0;
	out.legacy_far_point[1] = legacy_point->field_4;
	out.legacy_center_return_address = (uintptr_t)sub_523D30(
		(float*)(uintptr_t)out.hall_tokens[0], (float*)(uintptr_t)legacy_point_token);
	out.legacy_center_point[0] = legacy_point->field_0;
	out.legacy_center_point[1] = legacy_point->field_4;

	for (int i = 0; i < 2; ++i) {
		waypoints[i] = (nox_waypoint_t*)calloc(1, sizeof(*waypoints[i]));
		if (!waypoints[i]) {
			goto cleanup;
		}
		out.waypoint_addresses[i] = (uintptr_t)waypoints[i];
		out.waypoint_tokens[i] = nox_mapgenLegacyPtrRegister(waypoints[i]);
		if (!out.waypoint_tokens[i]) {
			goto cleanup;
		}
	}
	out.waypoint_native_link =
		nox_mapgenWaypointLinkNative_51D300(waypoints[0], waypoints[1], 7);
	out.waypoint_duplicate_rejected =
		!nox_mapgenWaypointLinkNative_51D300(waypoints[0], waypoints[1], 7);
	out.waypoint_legacy_link =
		sub_51D300((int)out.waypoint_tokens[1], (int)out.waypoint_tokens[0], 9);
	out.waypoint_self_rejected =
		!nox_mapgenWaypointLinkNative_51D300(waypoints[0], waypoints[0], 7);
	out.waypoint_counts[0] = waypoints[0]->points_cnt;
	out.waypoint_counts[1] = waypoints[1]->points_cnt;
	out.waypoint_link_addresses[0] = (uintptr_t)waypoints[0]->points[0].waypoint;
	out.waypoint_link_addresses[1] = (uintptr_t)waypoints[1]->points[0].waypoint;
	out.waypoint_kinds[0] = waypoints[0]->points[0].ind;
	out.waypoint_kinds[1] = waypoints[1]->points[0].ind;
	waypoints[0]->points_cnt = 31;
	out.waypoint_capacity_rejected =
		!nox_mapgenWaypointLinkNative_51D300(waypoints[0], waypoints[1], 8);

cleanup:
	*coordinate_epsilon = saved_coordinate_epsilon;
	if (room_added) {
		nox_xxx_mapGenFreeTopRoom_521A40();
		room = NULL;
	} else if (room) {
		sub_521A10(room);
		room = NULL;
	}
	for (int i = 0; i < 4; ++i) {
		if (halls[i]) {
			sub_521A10(halls[i]);
		}
	}
	for (int i = 0; i < 2; ++i) {
		if (neighbors[i]) {
			sub_521A10(neighbors[i]);
		}
		if (waypoints[i]) {
			nox_mapgenLegacyPtrForget(waypoints[i]);
			free(waypoints[i]);
		}
	}
	if (legacy_point) {
		nox_mapgenLegacyPtrForget(legacy_point);
		free(legacy_point);
	}
	sub_520F80();
	if (theme) {
		nox_mapgenLegacyPtrForget(theme);
		free(theme);
	}
	out.cleanup_head_is_null = nox_xxx_mapGenGetTopRoom_521710() == NULL;
	return out;
}
*/
import "C"

type mapgenRoomConnectionsResult522CA0 struct {
	themeAddress              uintptr
	roomAddress               uintptr
	roomAtCellAddress         uintptr
	hallAddresses             [4]uintptr
	hallResolvedAddresses     [4]uintptr
	neighborAddresses         [2]uintptr
	linkedNeighborAddresses   [2]uintptr
	legacyCenterReturnAddress uintptr
	waypointAddresses         [2]uintptr
	waypointLinkAddresses     [2]uintptr
	themeToken                uint32
	roomToken                 uint32
	hallTokens                [4]uint32
	waypointTokens            [2]uint32
	roomPoint                 [2]float32
	nearPoints                [4][2]float32
	farPoints                 [4][2]float32
	centerPoints              [4][2]float32
	legacyNearPoint           [2]float32
	legacyFarPoint            [2]float32
	legacyCenterPoint         [2]float32
	hallBounds                [4][4]float32
	roomPointCount            uint8
	waypointCounts            [2]uint8
	waypointKinds             [2]uint8
	gridInitialized           bool
	roomAdded                 bool
	addFirst                  int
	addDuplicate              int
	addLegacyDuplicate        int
	nativeLinesReturnedNull   bool
	nativePointsReturnedNull  bool
	legacyLinesReturnedNull   bool
	legacyPointsReturnedNull  bool
	hallLinksAdded            bool
	adjustRejected            bool
	waypointNativeLink        bool
	waypointDuplicateRejected bool
	waypointLegacyLink        bool
	waypointSelfRejected      bool
	waypointCapacityRejected  bool
	cleanupHeadIsNull         bool
}

func mapgenRoomConnectionsFixture522CA0() mapgenRoomConnectionsResult522CA0 {
	got := C.nox_test_mapgen_room_connections_522ca0()
	var out mapgenRoomConnectionsResult522CA0
	out.themeAddress = uintptr(got.theme_address)
	out.roomAddress = uintptr(got.room_address)
	out.roomAtCellAddress = uintptr(got.room_at_cell_address)
	for i := range out.hallAddresses {
		out.hallAddresses[i] = uintptr(got.hall_addresses[i])
		out.hallResolvedAddresses[i] = uintptr(got.hall_resolved_addresses[i])
		out.hallTokens[i] = uint32(got.hall_tokens[i])
		for j := range out.nearPoints[i] {
			out.nearPoints[i][j] = float32(got.near_points[i][j])
			out.farPoints[i][j] = float32(got.far_points[i][j])
			out.centerPoints[i][j] = float32(got.center_points[i][j])
		}
		for j := range out.hallBounds[i] {
			out.hallBounds[i][j] = float32(got.hall_bounds[i][j])
		}
	}
	for i := range out.neighborAddresses {
		out.neighborAddresses[i] = uintptr(got.neighbor_addresses[i])
		out.linkedNeighborAddresses[i] = uintptr(got.linked_neighbor_addresses[i])
		out.waypointAddresses[i] = uintptr(got.waypoint_addresses[i])
		out.waypointLinkAddresses[i] = uintptr(got.waypoint_link_addresses[i])
		out.waypointTokens[i] = uint32(got.waypoint_tokens[i])
		out.waypointCounts[i] = uint8(got.waypoint_counts[i])
		out.waypointKinds[i] = uint8(got.waypoint_kinds[i])
		out.roomPoint[i] = float32(got.room_point[i])
		out.legacyNearPoint[i] = float32(got.legacy_near_point[i])
		out.legacyFarPoint[i] = float32(got.legacy_far_point[i])
		out.legacyCenterPoint[i] = float32(got.legacy_center_point[i])
	}
	out.legacyCenterReturnAddress = uintptr(got.legacy_center_return_address)
	out.themeToken = uint32(got.theme_token)
	out.roomToken = uint32(got.room_token)
	out.roomPointCount = uint8(got.room_point_count)
	out.gridInitialized = got.grid_initialized != 0
	out.roomAdded = got.room_added != 0
	out.addFirst = int(got.add_first)
	out.addDuplicate = int(got.add_duplicate)
	out.addLegacyDuplicate = int(got.add_legacy_duplicate)
	out.nativeLinesReturnedNull = got.native_lines_returned_null != 0
	out.nativePointsReturnedNull = got.native_points_returned_null != 0
	out.legacyLinesReturnedNull = got.legacy_lines_returned_null != 0
	out.legacyPointsReturnedNull = got.legacy_points_returned_null != 0
	out.hallLinksAdded = got.hall_links_added != 0
	out.adjustRejected = got.adjust_rejected != 0
	out.waypointNativeLink = got.waypoint_native_link != 0
	out.waypointDuplicateRejected = got.waypoint_duplicate_rejected != 0
	out.waypointLegacyLink = got.waypoint_legacy_link != 0
	out.waypointSelfRejected = got.waypoint_self_rejected != 0
	out.waypointCapacityRejected = got.waypoint_capacity_rejected != 0
	out.cleanupHeadIsNull = got.cleanup_head_is_null != 0
	return out
}
