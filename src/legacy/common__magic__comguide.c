#include "GAME1_2.h"
#include "GAME2.h"
#include "common__strman.h"

const char* nox_guide_names_native[NOX_GUIDE_COUNT] = {
	"GUIDE_INVALID", "Bat",          "BlackBear",       "Bear",          "Beholder",
	"Bomber",        "CarnivorousPlant", "AlbinoSpider", "SmallAlbinoSpider", "EvilCherub",
	"EmberDemon",    "Ghost",        "GiantLeech",      "Imp",           "FlyingGolem",
	"MechanicalGolem", "Mimic",      "GruntAxe",        "OgreBrute",     "OgreWarlord",
	"Scorpion",      "Shade",        "Skeleton",        "SkeletonLord",  "Spider",
	"SmallSpider",   "SpittingSpider", "StoneGolem",    "Troll",         "Urchin",
	"Wasp",          "WillOWisp",    "Wolf",            "BlackWolf",     "WhiteWolf",
	"Zombie",        "VileZombie",   "Demon",           "Lich",          "WizardGreen",
	"UrchinShaman",
};
nox_guide_entry_t nox_guide_entries_native[NOX_GUIDE_COUNT] = {0};

//----- (00427070) --------------------------------------------------------
int nox_xxx_loadGuides_427070() {
	memset(nox_guide_entries_native, 0, sizeof(nox_guide_entries_native));
	for (int index = 1; index < NOX_GUIDE_COUNT; index++) {
		const char* creature = nox_guide_names_native[index];
		nox_thing* type = sub_44D330((char*)creature);
		if (!type) {
			return 0;
		}
		nox_guide_entry_t* entry = &nox_guide_entries_native[index];
		char text[256];
		nox_sprintf(text, "creature:%s", creature);
		entry->name = nox_strman_loadString_40F1D0(
			text, 0, "C:\\NoxPost\\src\\common\\Magic\\ComGuide.c", 57);
		entry->thing_type = strcmp(type->name, "Bomber") == 0
								? 0
								: nox_xxx_getTTByNameSpriteMB_44CFC0((char*)creature);
		nox_sprintf(text, "creature_desc:%s", creature);
		entry->description = nox_strman_loadString_40F1D0(
			text, 0, "C:\\NoxPost\\src\\common\\Magic\\ComGuide.c", 65);
		nox_sprintf(text, "CreatureCage%s", creature);
		entry->cage_image = nox_xxx_gLoadImg_42F970(text);
		nox_sprintf(text, "Spellbook%s", creature);
		entry->book_image = nox_xxx_gLoadImg_42F970(text);
		entry->field_20 = 0;
		entry->unit_size = (type->sub_class & 1) ? 1 : ((type->sub_class & 2) ? 2 : 4);
		*getMemU8Ptr(0x5D4594, 740100 + 28 * index) = entry->unit_size;
	}
	return 1;
}
