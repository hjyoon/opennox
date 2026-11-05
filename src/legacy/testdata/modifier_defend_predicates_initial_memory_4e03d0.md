# Independent native initial memory before the three raw modifier ports

The reference freeze `26086de87` precedes these controls. No predicate or
traversal engine is changed. `GAME3_2.c` remains byte-identical. The added
four-`int` traversal declaration in `GAME3_2.h` only declares its existing
definition so the pre-engine ABI control can inspect all three C symbols.

`generate_modifier_defend_predicates_initial_4e03d0.py` uses only the frozen
independent inputs. It reconstructs each complete 64-KiB PE32 initial heap,
declares its scalar/pointer initialization, and independently describes the
selected field reads. It executes neither an original nor a native body.
All 298 complete heaps match the actual original initial states exactly.
The encoded/decoded initial fixture SHA-256 values are:

```
8a1ca3db5a6e2c4ded9dd7d33165046f957ccc8355ca8ce2b941dc54b6c2ccc1
5eea6cc5cdefce04991b2ed96214b4fca33f854e80e08180cd7899c49093a1bc
```

The native fixture allocates a C-owned 64-KiB heap with a protected predecessor.
It independently checks thirteen C/Go layout facts, including complete
ModifierEff size, callback and signed-strength fields, the object C prefix,
class/init-data fields, complete modifier-array extent and all four pointer
slots. The Go object's two trailing handle words are outside the C prefix.
The last-readable-strength case certifies that word only; neither ABI has a
complete ModifierEff allocation after its boundary.

Only explicitly observed scalar and pointer windows are relocated. The rest
of the full allocation remains guarded. Actual C symbol identities replace
their original callback tokens; invalid addresses occupy the protected native
domain. Whole native bytes, original projection and independent original
initial hashes are checked before and after all field reads. Both typed Go
reads and actual C getters observe every valid selected domain. There are
136 actual nil/protected typed-read faults per profile. These independent
initial-field probes include domains which the original admission gates may
skip; they are distinct from the 18 original function faults per profile.
This does not claim arbitrary aliases between unrelated widened structures.

Go 1.27.3 normal and strict HD (cgocheck2/checkptr=2) each pass 306 events,
including all 298 initial-memory leaves and the 136 read-fault domains.
Four actual parent mains are preserved before execution. Each profile also
re-executes its exact retained legacy main for three isolated actual old-C
entry faults above 4 GiB. Whole native memory, typed C/Go getter controls and
old integer prototypes pass before each raw Inversion/Grip/traversal call.
The child then produces the expected C pointer fault. Parent mains and Go
drivers exit zero without forced termination; the intentional child failures
and their complete stacks/logs remain in the verified archives.

Evidence: `ART/modifier-defend-predicates-native-entry-port/independent-initial-memory-agrees-with-original-before-native-v1.json`
and `native-initial-memory-normal-and-strict-qualified-v1.json`. The complete
WIP source is byte-verified in `modifier-predicates-initial-source-v1.zip`.
The actual main files, JSON events and all logs are in `local-suite-archives`;
successful WORK/report/executable duplicates are removed after verification.

All three raw C bodies still accept integers and use the old offsets. The
signed-comparison mismatch also remains active. Complete original-output
consumers must next compile, run and freeze before any native engine change.
