# Go 도구체인 정책

이 포팅 브랜치의 유일한 지원 도구체인은 `go1.26.5`이다. 버전의 단일 텍스트 기준은 `go-version.txt`이며, `src/go.mod`에는 언어 기준 `go 1.26.0`과 권장 도구체인 `toolchain go1.26.5`를 함께 선언한다.

빌드와 테스트는 다음 래퍼로 실행한다.

```sh
./scripts/go.sh version
./scripts/go.sh -C src test ./internal/noxbuild
```

Linux 64비트의 native-width CGo 회귀는 `make test-linux-pie`로 root·server·legacy 전체를 시험한다. 기본 비-PIE 실행 파일에서는 C heap이 4GiB 아래에 놓일 수 있어 고주소 포인터 시험의 전제가 성립하지 않을 수 있다. PIE 시험은 그 전제를 유지하는 별도 검증 게이트이며, 일반 제품의 빌드·실행 검증을 대체하지 않는다.

Windows PowerShell에서는 다음과 같이 실행한다.

```powershell
.\scripts\go.ps1 version
.\scripts\go.ps1 -C src test ./internal/noxbuild
```

실행할 제품이 현재 clean source revision과 정확히 일치하는지는 `noxbuild -verify`로 확인한다. 이 검사는 ELF, PE, Mach-O를 호스트 종류와 무관하게 읽어 Go 1.26.5, 지원 대상 tuple, 전체 `vcs.revision`, `vcs.modified=false`를 강제한다. 특히 무시된 `build/` 아래에 남은 구 실행 파일을 다시 실행하는 일을 차단한다.

```sh
./scripts/go.sh -C src run ./internal/noxbuild \
  -go=../scripts/go.sh -verify \
  ../build/linux-amd64/opennox ../build/linux-amd64/opennox-server
```

```powershell
.\scripts\go.ps1 -C src run ./internal/noxbuild `
  -go=go -verify `
  ..\build\windows-amd64\opennox.exe ..\build\windows-amd64\opennox-server.exe
```

래퍼는 외부 `GOROOT`를 제거하고 `GOTOOLCHAIN=go1.26.5`와 빈 `GOEXPERIMENT`를 강제한 뒤, 실제 `GOVERSION`이 정확히 일치하는지 검사한다. 이로써 goenv 같은 버전 관리자가 다른 표준 라이브러리 경로를 주입하는 경우도 차단한다. `GO` 환경 변수에는 공백 없는 Go 실행 파일 경로 하나만 지정할 수 있다. `internal/noxbuild`도 자신을 컴파일한 Go와 자식 빌드에 쓰는 Go를 각각 검사한다.

따라서 `GOEXPERIMENT=cgocheck2 ./scripts/go.sh ...`는 strict CGo 검증이 아니다. POSIX 환경의 테스트 전용 실행 경로는 고정 도구체인을 먼저 찾은 뒤 `cgocheck2`와 `CGO_ENABLED=1`의 실제 적용을 확인한다. 릴리스 빌드 래퍼의 정책은 바꾸지 않으며, 테스트 인수를 생략하면 전체 패키지를 실행한다.

```sh
sh ./scripts/test-cgocheck2.sh
sh ./scripts/test-cgocheck2.sh -run TestNative -count=3 . ./legacy ./server
```

macOS/ARM64의 GUI 회귀는 `NOX_E2E_SEAT=headless`로 실행한다. 다음 통합 시나리오는 Con01a 곰 전투, Con02a 네크로맨서 소환·공격·퇴장, 늑대 Charm 60회와 Henrick 추종, 이후 ESC 메뉴 입력을 같은 세션에서 검증한다. 맵 trigger로 플레이어 위치만 준비하며 대화 버튼은 실제 mouse 입력으로 누른다. 일반·HD 제품 모두 통과했으며, 네크로맨서 소환 시 frozen 상태와 이벤트 종료 후 조작/cinematic 복귀를 구분한다. 이 검증은 SDL 화면 출력이나 실제 OpenAL 재생 검증을 대신하지 않는다.

```sh
NOX_E2E_SEAT=headless bash ./scripts/run-headless-gui-e2e.sh \
  /absolute/path/to/nox ./scripts/e2e/solo-conjurer-headless.yaml
NOX_E2E_SEAT=headless NOX_E2E_CLIENT_TARGET=client-hd \
  bash ./scripts/run-headless-gui-e2e.sh /absolute/path/to/nox \
  ./scripts/e2e/solo-conjurer-headless.yaml
```

`host-game-flame-monster-food.yaml`도 일반·HD headless 실행에서 player HP `150→148`, Spider의 RedApple `40→45`, Meat `40→50`, Mushroom 독 `4→0`과 소비된 필드 오브젝트 제거를 확인했다. 휴면 tutorial NPC/Wolf의 원본 `Cur=Max=0`은 사망 상태가 아니며, 일반 몬스터의 food/retreat AI와 구분한다. 통합 이벤트의 종료가 통과하더라도 미이식 피해 분기는 별도로 검사해야 한다.

패치 버전을 올릴 때에는 다음 항목을 한 변경으로 갱신한다.

1. `toolchain/go-version.txt`
2. `src/go.mod`의 `toolchain` 지시자
3. `src/internal/noxbuild/main.go`의 `requiredGoVersion`
4. Docker 이미지 태그와 CI 매트릭스
5. `BASELINE-linux-386.md` 및 전체 대상 빌드 결과

현재 포팅 단계, 정적 검색 재현법과 확인된 64비트 ABI 단절은 `PORTING-INVENTORY.md`에 기록한다.

첫 구조체 분리의 근거, C/Go 오프셋과 실제 Linux/386 산출물 검증값은 [`ABI-OBJECT.md`](ABI-OBJECT.md)에 기록한다.

사용자가 보유한 원본 데이터에 의존하는 시험은 [`oracle/README.md`](oracle/README.md)의 O0 무결성 게이트를 먼저 통과해야 한다. `make oracle-test`는 봉인된 `nox/` 이외의 데이터로 기대 결과가 바뀌는 것을 차단한다.

임의의 `GOEXPERIMENT`, 호스트 기본 아키텍처 최적화, `latest` 도구체인은 릴리스 산출물에 사용하지 않는다.

Linux Docker 빌더의 Go 베이스는 다음 멀티아키텍처 인덱스로 고정한다.

```text
golang:1.26.5-trixie@sha256:98988b42f3293b627bf07c884ff17181a59501769cd8c06c7ba901e0ce2c9853
```

이 인덱스에서 이 프로젝트가 사용하는 플랫폼은 `linux/386`, `linux/amd64`, `linux/arm/v7`, `linux/arm64/v8`이다.
