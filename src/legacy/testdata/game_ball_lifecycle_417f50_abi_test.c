#include "../game_ball_lifecycle_417f50.h"

typedef int (*game_ball_reset_fn)(nox_object_t*);
typedef void (*game_ball_update_fn)(nox_object_t*);
typedef int (*game_ball_death_fn)(nox_object_t*);

_Static_assert(_Generic(&sub_417F50, game_ball_reset_fn: 1, default: 0),
	"00417F50 must receive a native-width object pointer");
_Static_assert(_Generic(&nox_xxx_updateGameBall_53DF40, game_ball_update_fn: 1, default: 0),
	"0053DF40 must receive a native-width object pointer");
_Static_assert(_Generic(&nox_xxx_dieGameBall_54E620, game_ball_death_fn: 1, default: 0),
	"0054E620 must receive a native-width object pointer");

int main(void) {
	return 0;
}
