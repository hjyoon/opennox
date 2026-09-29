#include "client__gui__guijourn.h"

#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2_3.h"
#include "common__strman.h"
#include "operators.h"

extern uint32_t nox_color_white_2523948;
extern uint32_t nox_color_yellow_2589772;
extern uint32_t nox_color_black_2650656;
extern uintptr_t dword_8531A0_2576;
extern uint32_t dword_8531A0_2572;

//----- (00469BC0) --------------------------------------------------------
void nox_xxx_cliBuildJournalString_469BC0() {
	nox_playerInfo* player = (nox_playerInfo*)dword_8531A0_2576;
	if (!player) {
		return;
	}
	int font_height = nox_xxx_guiFontHeightMB_43F320(0);
	int total_height = -font_height;
	for (nox_playerInfo_journal* entry = player->field_3644; entry; entry = entry->next) {
		char key[76] = "Journal:";
		wchar2_t text[2048];
		int text_height;
		strncat(key, entry->entry, sizeof(key) - strlen(key) - 1);
		switch (entry->field_3) {
		case 2:
			nox_wcsncpy(text,
				nox_strman_loadString_40F1D0("Journal:QuestLabel", 0,
					"C:\\NoxPost\\src\\client\\Gui\\GUIJourn.c", 56),
				sizeof(text) / sizeof(text[0]));
			break;
		case 4:
			nox_wcsncpy(text,
				nox_strman_loadString_40F1D0("Journal:CompletedLabel", 0,
					"C:\\NoxPost\\src\\client\\Gui\\GUIJourn.c", 60),
				sizeof(text) / sizeof(text[0]));
			break;
		case 8:
			nox_wcsncpy(text,
				nox_strman_loadString_40F1D0("Journal:HintLabel", 0,
					"C:\\NoxPost\\src\\client\\Gui\\GUIJourn.c", 64),
				sizeof(text) / sizeof(text[0]));
			break;
		default:
			text[0] = 0;
			break;
		}
		nox_wcscat(text, L" ");
		nox_wcscat(text,
			nox_strman_loadString_40F1D0(key, 0, "C:\\NoxPost\\src\\client\\Gui\\GUIJourn.c", 74));
		nox_xxx_drawGetStringSize_43F840(0, text, 0, &text_height, 240);
		total_height += font_height + text_height;
	}
	*getMemU32Ptr(0x5D4594, 1064848) = total_height <= 0 ? 0 : total_height;
}

//----- (00469D40) --------------------------------------------------------
void nox_xxx_guiDrawJournal_469D40(int xLeft, int yTop, int a3) {
	nox_playerInfo* player = (nox_playerInfo*)dword_8531A0_2576;
	int draw_y = yTop - a3;
	if (player) {
		nox_client_drawSetColor_434460(nox_color_black_2650656);
		nox_client_drawRectFilledOpaque_49CE30(xLeft, yTop, 260, 150);
		nox_playerInfo_journal* entry = player->field_3644;
		if (entry) {
			while (entry->next) {
				entry = entry->next;
			}
			int font_height = nox_xxx_guiFontHeightMB_43F320(0);
			do {
				char key[76] = "Journal:";
				wchar2_t text[2048];
				uint32_t color;
				int text_height;
				strncat(key, entry->entry, sizeof(key) - strlen(key) - 1);
				switch (entry->field_3) {
				case 1:
					color = nox_color_white_2523948;
					text[0] = 0;
					break;
				case 2:
					color = *getMemU32Ptr(0x85B3FC, 940);
					nox_wcscpy(text, nox_strman_loadString_40F1D0("Journal:QuestLabel", 0,
						"C:\\NoxPost\\src\\client\\Gui\\GUIJourn.c", 135));
					break;
				case 4:
					color = *getMemU32Ptr(0x85B3FC, 956);
					nox_wcscpy(text, nox_strman_loadString_40F1D0("Journal:CompletedLabel", 0,
						"C:\\NoxPost\\src\\client\\Gui\\GUIJourn.c", 140));
					break;
				case 8:
					color = nox_color_yellow_2589772;
					nox_wcscpy(text, nox_strman_loadString_40F1D0("Journal:HintLabel", 0,
						"C:\\NoxPost\\src\\client\\Gui\\GUIJourn.c", 145));
					break;
				default:
					color = dword_8531A0_2572;
					text[0] = 0;
					break;
				}
				nox_wcscat(text, L" ");
				nox_wcscat(text,
					nox_strman_loadString_40F1D0(key, 0, "C:\\NoxPost\\src\\client\\Gui\\GUIJourn.c", 155));
				nox_xxx_drawGetStringSize_43F840(0, text, 0, &text_height, 240);
				int text_bottom = text_height + draw_y;
				if (text_bottom > yTop) {
					nox_xxx_drawSetTextColor_434390(color);
					nox_xxx_drawStringWrap_43FAF0(0, text, xLeft + 10, draw_y, 240, 0);
				}
				draw_y = text_bottom + font_height;
				if (draw_y > yTop + 150) {
					break;
				}
				entry = entry->prev;
			} while (entry);
		}
	}
}
