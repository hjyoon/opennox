# Original monster animation selector

The producer executes unchanged GAME.EXE, SHA-256
`0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a`.
Its static audit dependency is `original-functions-and-tables-static-v1.json`.

Startup precision-53 / ToZero (`0xe3f`) and diagnostic nearest (`0x23f`)
each execute 14,840 unique contexts and 15,064 selector calls. They execute
all 281 selector, 47 Mimic, 47 Plant, 79 Zombie and 28 WORD-name-ID helper
bytes: 482 body bytes in total. All 52 jump-table and 31 action-index bytes
are actually read. No padding or linear disassembly is counted as execution.
Both profiles produce identical rows, including per-context predicate-entry
observations, lookup/cache publication, declared mutations and full guards.

Only dictionary4E3830 is substituted: it returns a valid type record or nil
and applies declared input effects. The original WORD-ID helper and all
predicates execute. No predicate result, animation index or cache result is
supplied. Earlier duplicate inputs and preparation failures remain preserved;
the installed v4 producer adds read-only observations without changing any
v2 input, name or existing expected output.

The frozen runtime comparisons use real C-owned native records, the actual
public C Mimic/Plant/Zombie entries and a real registered type dictionary.
GetServer and the existing Plant/Zombie runtime seam observe dictionary
entry and apply the declared input mutations. No predicate result is injected.
Both direct Go and public C selectors must match original DWORD returns,
predicate entry/order/state, dictionary observations, caches, Go cache
publications and every remaining guarded native object/update/target/cache byte.
The unchanged C Mimic store count is not instrumented; its separately
qualified primitive, real cache snapshots and dictionary observations are
checked. This is not a new claim of C instruction-level write-count tracing.

Before porting, ordinary native Go action/animation controls passed and the
actual public C selector reproduced a pointer-truncation SIGSEGV. Allocation
and real-registry controls pass all 14,840 rows without selector calls.
The unchanged older private Go body, with its existing eager argument
evaluation and actual predicates, differs on 3,510 inputs. That test adapter
does not represent every part of the old public Go wrapper. Its source and
failed results are retained. Primary expected values and inputs are unchanged.

Independent default-binding modes use actual public C and ordinary Go calls,
an unmodified predicate factory, real registered types and original words.
They include cached Mimic idle/morph paths that must avoid server lookup.

The original cached update, action-before-predicate computation, signed stack
addressing and early returns remain the native core's contract. The older
private helper retains its existing invalid-head containment. Native layouts
and guarded typed update getters are unchanged. The two retained extracted
C callers still contain PE32 layout assumptions; adapting their root call
does not port those entire parent chains. Product/campaign and native-platform
checks remain separate. Overall 100% is not established by these references.
