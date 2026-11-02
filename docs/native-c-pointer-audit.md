# Native C entry pointer audit

This inventory prioritizes remaining public C ABIs after the native point,
occupancy and retreat ports. It does not supply an overall porting denominator or
count confirmed runtime crashes.

At `616b61d`, host macOS ARM64 normal preprocessing of all **191** top-level
`src/legacy/*.c` files finds **3,502** active function definitions. Real clang
preprocessing excludes the preserved `#if 0` original bodies; all 191
preprocessor logs have zero warnings. Full preprocessed outputs, input
hashes, function/caller inventories and scripts are retained under
`ROOT/monster-retreat-predicates-native-entry-port`.

The heuristic finds **303** definitions casting an integer parameter to a
pointer. **289** have a fixed 32-bit `int`/`int32_t`/`uint32_t` operand, while
**14** use native-width or other integer types. Of the 289, **26** have direct
non-test Go `C.symbol` references and **171** have direct active C callers.
References may store or compare callback addresses rather than invoke them.
Forwarding entries without their own cast, cgo preambles, aliases, pointer
returns, globals and integer fields need separate analysis. Other target
preprocessor configurations also remain outside this host inventory.

Semantic inspection is required before changing a candidate:

| Example | Classification |
| --- | --- |
| `548630` collision target and `548830` door data | Already use `uintptr_t`; these are not fixed 32-bit argument truncations. |
| `4597E0` server-options hide argument | Integer control flag, also returned through a legacy pointer-shaped result; a cast alone does not establish a pointer input. |
| `56F9E0` protected-mana argument | Integer protection handle resolved by `56F590`, with sentinel results; changing it to an object pointer would change its meaning. |
| `549D80` and related death symbols | Some Go references classify callback addresses and select existing native implementations. Raw C ABI and actual dispatcher behavior must be checked separately. |
| `545520`, `545580`, `5455B0` | Actual legacy-int public entries each SIGSEGV above 4 GiB after independent native-layout, typed sound/action/buff and whole-allocation positive controls pass. Frozen original comparison precedes the native port. |

The three retreat predicates now use typed `nox_object_t*` public arguments
and shared native Go cores. Their result remains the original canonical
integer. Raw entries retain original faults and short circuits; ordinary
callers keep their earlier admission checks. Original C bodies remain exact
under `#if 0`, and the frozen original getters/consumers are unchanged.

At ported source `43ae3d9`, the repeated host scan finds **3,499 active C
definitions / 301 integer-cast candidates / 287 fixed-32-bit candidates**.
Only the resume and cast predicates disappear from the cast inventory; the
third, forwarding retreat predicate had no cast of its own. All three typed
public ABIs are independently typechecked and compared with the original.
The fourteen native-width/other candidates remain separated. This scan
introduces no new fixed-32-bit candidate or preprocessor warning.

At that preceding `43ae3d9` source, raw `5456D0` still has an integer object
argument. No direct production call is found outside its definition/header;
the ordinary dispatcher already uses a native Go update. That classification
does not close indirect reachability. The subsequent actual C reproduction
SIGSEGV above 4 GiB after independent layouts, real native owner/Push
composition and typed sound/action/buff/predicate controls pass.

Runtime `8212e8c` ports raw `5456D0` to a typed `nox_object_t*` entry and shares
its update core with the ordinary dispatcher. The raw entry retains original
faults, pointer-load order, cached MUD/live owner behavior, unspilled distance
arithmetic and ordered action writes. Ordinary admission, including guarded
missing health, remains intact. Five original references are frozen before
engine changes, including a corrected C-prefix-versus-Go-EXT layout check;
the initial preparation failure and actual old C crash remain preserved.

The repeated `8212e8c` host scan finds **3,498 active C definitions / 300
integer-cast candidates / 286 fixed-32-bit candidates**. Only `5456D0` leaves
the fixed-width inventory. Fourteen native-width/other candidates remain
separate; no new candidate or preprocessor warning is introduced. Integer
object entries `5455E0`, `545790` and `5457C0` remain follow-up targets.
Their ordinary native dispatcher implementations do not port those raw C
ABIs. In particular, ROAM cancel has signed unchecked stack-index arithmetic
and a pointer-shaped return; widening its first argument alone is insufficient.

Build, strict-pointer, product and CI results belong in
[porting-completion.md](porting-completion.md), with their exact source
revision. Static inventory changes alone earn no campaign, audio, runtime
stability or overall 100% completion credit.
