# 비공개 원본 오라클 매니페스트

이 디렉터리에는 사용자가 보유한 `nox/` 기준본의 **경로, 바이트 수, SHA-256**만 보관한다. `GAME.EXE`, 맵, 음성, 영상 등 원본 자산 자체를 소스 저장소나 공개 CI에 복사하지 않는다.

`nox-2023-1003-01.json`은 다음 규칙으로 봉인한다.

- 모든 일반 파일을 상대 경로 순서로 정렬한다.
- 각 파일의 크기와 SHA-256을 기록한다.
- `(경로 NUL 크기 NUL 파일 SHA-256 LF)` 레코드 전체의 SHA-256을 `tree_sha256`으로 기록한다.
- 심볼릭 링크, 특수 파일, 대소문자만 다른 충돌 경로와 이식 불가능한 경로는 거부한다.
- `seal`은 기존 매니페스트를 덮어쓰지 않는다. 기준본을 바꾸려면 기존 파일을 보존하고 새 ID와 새 파일명으로 별도 검토한다.

검증은 저장소 루트에서 실행한다.

```sh
make oracle-verify
make oracle-code-verify
make oracle-test
```

`game-exe-functions.json`은 포팅한 원본 함수의 VA, 정확한 함수 길이와 범위 SHA-256을 기록한다. 검증기는 전체 트리 매니페스트와 오라클 ID 및 `GAME.EXE` 해시를 교차 확인하고, PE32/I386 형식, 이미지 베이스 `0x00400000`, 함수와 jump table이 실행 섹션에 있는지, 상수 테이블이 비실행 섹션에 있는지와 각 범위 해시를 검사한다. 현재는 `0x004D8270`, `0x004D9EB0`, `0x004E44F0..0x004E80BE`, `0x0054E170..0x0054E45D`의 115개 함수와 `0x004E4C90`·`0x004E6CE0`이 사용하는 두 dispatch table 범위를 합계 117개 코드 범위로 봉인했다. `004E6E50`의 입력인 256개 정수 방향 벡터·임계값·9×16 반환표, `004E7410`의 float32 `85.0` 상수, `004E7470`의 두 드롭 표·타입 문자열과 WeaponDie/ArmorDie 등록 레코드는 여섯 데이터 범위로 별도 봉인했다. 함수 하나를 포팅할 때마다 어셈블리에서 경계를 확인한 뒤 이 목록과 의미 설명을 확장한다.

`oracle-test`는 의미 비교 전과 후에 O0과 코드 범위 검증을 실행한다. 테스트 입력은 읽기 전용으로 취급해야 하며, 컨테이너/VM에서는 가능하면 `nox/`를 read-only로 마운트한다. 사후 검사는 잘못된 테스트가 원본을 수정한 경우도 실패로 남긴다.

다른 위치의 정당한 보유본을 쓰려면 절대 경로나 저장소 루트 기준 경로를 넘긴다.

```sh
make oracle-verify NOX_ORACLE_ROOT=/absolute/path/to/nox
```

Windows에서 GNU Make를 사용하지 않는 경우에도 같은 검증기를 실행할 수 있다.

```powershell
$NoxRoot = (Resolve-Path 'C:\Games\Nox').Path
$Manifest = (Resolve-Path '.\toolchain\oracle\nox-2023-1003-01.json').Path
.\scripts\go.ps1 -C src run .\internal\noxoracle verify `
  -root $NoxRoot -manifest $Manifest

$CodeManifest = (Resolve-Path '.\toolchain\oracle\game-exe-functions.json').Path
.\scripts\go.ps1 -C src run .\internal\noxoracle code-verify `
  -root $NoxRoot -manifest $CodeManifest -oracle-manifest $Manifest
```

전체 트리 매니페스트는 원본의 **신원과 무결성**을 보장하는 Gate O0이다. 함수 범위 매니페스트는 포팅한 소스가 어느 원본 주소·바이트 범위에서 유래했는지 기계적으로 고정하지만, 해시만으로 의미 동등성을 증명하지는 않는다. 각 함수에는 어셈블리 검토와 원본 또는 검증된 기준 구현을 사용한 차등 시험이 추가로 필요하고, 게임 상태·패킷·프레임 동등성은 상위 게이트에서 별도로 검증한다.
