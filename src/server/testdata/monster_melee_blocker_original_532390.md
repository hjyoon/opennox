# Original melee native-entry references

`original_monster_melee_actions_532130.py` executes the unchanged original
start `00532130` and update `00532440`, calling the original blocker,
Push/Pop/reset, stamina, RNG, morph, sound getter and enemy predicates.
Each FPU profile contains 3,664 inputs: 2,255 start and 1,409 update inputs.
All 600 start, 161 blocker and 248 update body bytes execute. Their complete
intervals contain 1,040 bytes including 31 padding bytes. The complete
251-byte Push, 254-byte Pop and 67-byte RNG also execute.

Startup `0x0e3f` supplies the expected values. Nearest `0x023f` is a retained
control, with 302 differing action outputs. The original RNG's 4,096 DWORD
values equal the native prand table. No FP, predicate, RNG, Push admission
or Pop result is substituted. There are 646 original blocker calls and
358 original enemy checks per profile, with 176 flee contexts.

Spatial enumeration supplies an ordered world and invokes the original
blocker. Buff-flag publication, audio with a declared zero return, diagnostics,
player attack and registered strike/cancel/end callbacks are external engine
services. Type and resource caches are initialized input state. These references
do not certify first type lookup, every enemy branch, actual audio return
behavior or all transitive attack/action bodies. Full allocation guards,
unchanged code/RNG/caches/registry, nonvolatile registers, stack bounds,
FPU control and empty TOP are checked.

`original_monster_melee_blocker_532390.py` independently executes the service-free
176-byte callback interval for 14,112 inputs per profile, covering all 161 body
bytes. It checks guarded objects, packed globals, original direction table
and constants, stack, code, nonvolatile registers and FPU control. There are
1,326 differing profile outputs. FPSW/FPTAG values are not row outputs and
are not claimed equivalent to native Go/C.

The native actual-C comparators use guarded C-owned records above 4 GiB on
64-bit hosts, native stack/object/update/sound pointers, all 24 stack slots,
unused arguments, packed-global neighbors, raw return words or char bytes,
and actual native RNG/enemy code. Their allocation controls are separate
from equivalence. The Go action comparator checks the same state/services;
the old handled-bool interface has no original return word, while the new
native method and actual C comparator both check it.

The private first action profiles are retained under the validation ROOT
and explicitly excluded: a stale generated enumeration adapter omitted
nonempty queries. Corrected v2 invalidates that adapter's translation,
retains its 3,646-input prefix and adds 18 condition-pop inputs. No original
reference was changed after engine implementation.

The shared pointer uses `memmap.PtrPtr(0x5d4594, 2487948)` and distance uses
the separate packed float at offset 2487952. `Field516` remains a DWORD;
native weapon associations preserve exact cached update identity and known
nil values. Ordinary C equip/dequip and owner-address reuse have a separate
regression. Passing these bounded references does not establish overall
porting completion or the final nine chapter 1-3 audits.
