# Go 1.26.5 Linux/386 기준선

소스 수정 전 기준 커밋 `b184030e76be2b681a7f6d2bcdef52b091d94b9b`를 Go 1.26.5로 빌드한 결과다.

## 환경

- 호스트: macOS arm64, Docker Desktop
- 빌드 플랫폼: `linux/386`
- Go: `go1.26.5 linux/386`
- C 컴파일러: GCC 12.2.0, i686
- SDL: 2.26.5
- OpenAL: 1.19.1
- 빌드 이미지: `ghcr.io/opennox/docker-build@sha256:80a75e37704a2ca876cfa94a5c0b576355f5579edb4ce471eb732ef1127e526d`
- `GOTOOLCHAIN=local`, 빈 `GOEXPERIMENT`, `GO386=sse2`, `CGO_ENABLED=1`
- 소스 마운트: 읽기 전용
- 전체 빌드 시간: `24m38.9596903s`

## 산출물

| 파일 | 크기 | SHA-256 |
|---|---:|---|
| `opennox` | 약 45 MiB | `462caf42a1d6fd340f8e02ee5d1d6a091c09db306a48a4a04f46eb39c084f4e6` |
| `opennox-hd` | 약 46 MiB | `04d0d38f8170e19689151020a7053ec62fc5551cacec87749a1f4fa4ac81307f` |
| `opennox-server` | 약 43 MiB | `d34777ddcfe25e8ff702d9b64c60aa0aef38227094596d91ecbe75050d4a9fe5` |

세 파일은 모두 동적 링크된 `ELF 32-bit LSB executable, Intel 80386`이며, `go version -m`에서 다음을 확인했다.

- Go 버전: `go1.26.5`
- `GOOS=linux`
- `GOARCH=386`
- `GO386=sse2`
- `CGO_ENABLED=1`
- VCS revision: `b184030e76be2b681a7f6d2bcdef52b091d94b9b`
- VCS modified: `false`

세 바이너리를 같은 Linux/386 이미지에서 각각 `-h`로 실행했고 모두 종료 코드 0을 반환했다.

## 기준선에서 드러난 포팅 위험

빌드는 성공했지만 다음 C 파일군에서 포인터/정수 캐스트·비교 또는 서로 다른 함수 포인터 비교 경고가 발생했다.

- `legacy/cnxz/nxz_comp.c`
- `legacy/GAME1*.c`부터 `legacy/GAME5.c`
- `legacy/client__draw__*.c`
- `legacy/client__gui__*.c`
- `legacy/client__network__cdecode.c`
- `legacy/client__shell__*.c`
- `legacy/server__system__server.c`

서버 태그에서도 클라이언트 레거시 C 파일 상당수가 컴파일된다. 따라서 64비트·ARM 포팅은 경고 억제나 일괄 캐스트가 아니라 주소, 오프셋, 직렬화 필드, 핸들, 함수 포인터를 의미별 타입으로 복원해야 한다.

기준 바이너리는 저장소에 커밋하지 않았으며 로컬 `/private/tmp/opennox-go126-baseline`에만 보관했다.
