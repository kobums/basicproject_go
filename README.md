<div align="center">

# 🚀 basicproject_go

**Fiber 기반 Go REST API 시작 템플릿 + gomachine 코드 생성**

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Fiber](https://img.shields.io/badge/Fiber-v2-00ACD7?style=for-the-badge&logo=go&logoColor=white)](https://gofiber.io/)
[![MariaDB](https://img.shields.io/badge/MariaDB-11.4-003545?style=for-the-badge&logo=mariadb&logoColor=white)](https://mariadb.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)

</div>

---

## 📖 소개

**basicproject_go** 는 새 Go 백엔드를 바로 시작하기 위한 템플릿이다.
"이 템플릿 + 표준 DB(`db/schema.sql`) + [gomachine](https://github.com/kobums/gomachine)" 으로 초기 구성이 끝난다.

- 템플릿은 **인증 / 인프라만** 가진다: 로그인 · 회원가입 · 내 정보, 이미지 업로드, health, DB 풀, 설정, 로그.
- **테이블 CRUD 는 모두 gomachine 이 만든다**: `models/{table}.go`, `controllers/rest/{table}.go`, `router/routers/{table}.go`, `router/routers/generated_routes.go`.
  표준 스키마(user / board)로 만든 파일이 커밋되어 있어 DB 없이도 바로 빌드된다.
- 토큰은 Go / Spring / Rust 템플릿이 같이 쓰는 **통일 JWT** 다 (한 백엔드에서 받은 토큰을 다른 백엔드에서도 쓴다).
  인증 / 업로드 API 도 gomachine README 의 "인증 API 계약" 그대로라 프런트(React / Vue)는 백엔드를 바꿔도 같은 코드로 로그인한다.
- gomachine 의 `buildtool-model init` (workspace 모드)이 이 저장소를 새 프로젝트의 Go 백엔드로 복사해 쓴다.
  그래서 저장소 밖의 경로에 의존하지 않는다 (스키마 · 설정 · 생성 파일이 모두 이 안에 있다).

---

## ✨ 주요 기능

| 기능 | 설명 |
|:---|:---|
| ⚙️ **코드 자동 생성** | `model.json` + DB 스키마로 모델 / REST 컨트롤러 / 라우터 생성 (`make gen`) |
| 🔐 **JWT 인증** | HS256 통일 JWT, bcrypt 비밀번호, `controllers.Controller.Session` 으로 생성 코드에 전달 |
| 🗂️ **업로드** | 이미지(jpg/png/gif/webp, 5MB)만, 내용 시그니처 검사, 서버가 파일명 생성, 요청 본문 20MB 제한 |
| 🔒 **설정** | 환경변수 우선 (`.env` → 환경변수 → `.env.yml`), 비밀값은 이미지에 넣지 않는다 |
| 🐬 **DB** | MariaDB 11.4 (docker compose), 공유 커넥션 풀 하나 |
| 🐳 **Docker** | multi-stage 빌드, 비 root 실행, compose 로 db + api |

---

## ⚡ 빠른 시작

준비물: Go 1.26+, Docker (DB), [gomachine](https://github.com/kobums/gomachine) (`make gen` 할 때만).

```bash
# 1. 설정 파일
cp .env.example .env              # JWT_SECRET 은 openssl rand -base64 48 로 바꾼다

# 2. DB (MariaDB, db/schema.sql 로 user_tb / board_tb 생성)
make db-up

# 3. 코드 생성 (선택: 표준 스키마로 만든 파일은 이미 커밋되어 있다)
make gen

# 4. 실행 → http://localhost:8007
make run
```

포트: Go API `8007` (`PORT`). 같은 계약의 다른 백엔드 템플릿은 Spring `8081`, Rust `8008` 이다.

확인:

```bash
# 회원가입 (level 1) → 토큰
curl -s -X POST localhost:8007/api/auth/signup -H 'Content-Type: application/json' \
  -d '{"email":"a@b.c","name":"홍길동","password":"password123"}'

# 로그인
TOKEN=$(curl -s -X POST localhost:8007/api/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"a@b.c","password":"password123"}' | sed 's/.*"token":"\([^"]*\)".*/\1/')

curl -s localhost:8007/api/auth/me -H "Authorization: Bearer $TOKEN"

# 생성된 board API (작성자 = 로그인 사용자, 로그인 사용자는 모든 글을 읽는다)
curl -s -X POST localhost:8007/api/board -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"title":"첫 글","content":"내용"}'
curl -s 'localhost:8007/api/board?page=1&pagesize=10&title=첫' -H "Authorization: Bearer $TOKEN"

# 이미지 업로드 → {filename, url}
curl -s -X POST localhost:8007/api/upload/index -H "Authorization: Bearer $TOKEN" -F upload=@photo.png
```

---

## 🏗️ 프로젝트 구조

```
basicproject_go/
├── main.go                      # 엔트리포인트 (설정 검증 → 서버 시작)
├── model.json                   # gomachine 설정 (테이블, 인증, 권한)
├── db/schema.sql                # 표준 스키마 (gomachine bootstrap/db/schema.sql 과 같다)
├── docker-compose.yml           # db (MariaDB) + api (profile)
├── dockerfile                   # multi-stage 빌드
├── .env.example                 # 환경변수 예 (cp .env.example .env)
├── services/http.go             # Fiber 앱 (CORS, 본문 제한, 압축, 정적 파일, 라우터)
├── router/
│   ├── router.go                # /api 라우트 등록 (auth → upload → 생성 라우트)
│   └── routers/
│       ├── auth.go              # 로그인 / 회원가입 / me          (손으로 관리)
│       ├── auth_middleware.go   # JwtRequired, 에러 응답           (손으로 관리)
│       ├── upload.go            # 이미지 업로드 {filename, url}    (손으로 관리)
│       ├── {table}.go           # 테이블 라우트                    (gomachine 생성)
│       └── generated_routes.go  # SetupGeneratedRoutes             (gomachine 생성)
├── controllers/
│   ├── controllers.go           # 공통 Controller + Session        (손으로 관리)
│   ├── dto/dto.go               # 인증 응답 DTO                    (손으로 관리)
│   └── rest/{table}.go          # 테이블 컨트롤러                  (gomachine 생성)
├── models/
│   ├── db.go                    # 공유 DB 풀, Connection, WhereClause (손으로 관리)
│   ├── cache.go                 # 시작 시 캐시 훅                  (손으로 관리)
│   ├── {table}.go               # 테이블 모델 / 매니저             (gomachine 생성)
│   └── {table}/{table}.go       # 컬럼 상수                        (gomachine 생성)
└── global/
    ├── config/                  # 설정 (환경변수 / .env / .env.yml)
    ├── jwt/                     # 통일 JWT 발급 / 검증, bcrypt
    ├── log/                     # zerolog + lumberjack
    ├── upload/                  # 이미지 검증 / 저장
    └── ...                      # 시간, 문자열, 메일 / SMS 등 유틸
```

- **gomachine 생성 파일** 은 첫 줄이 `// Code generated by gomachine. DO NOT EDIT.` 이다. 직접 고치지 말고 `model.json` / 스키마를 바꾼 뒤 `make gen` 한다.
- 첫 줄이 다른 파일(위의 "손으로 관리" 파일)은 gomachine 이 덮어쓰지 않는다.
- 생성 파일을 손으로 고쳐야 하면 머리줄을 지운다 (그 뒤로 gomachine 이 건너뛴다).

---

## 🔐 인증 흐름

1. `POST /api/auth/signup` 또는 `POST /api/auth/login` 이 토큰을 돌려준다.
2. 클라이언트는 이후 요청에 `Authorization: Bearer <token>` 을 붙인다.
3. 모든 컨트롤러의 `Init` 이 토큰을 검증해 `controllers.Controller.Session` 을 채운다 (없거나 무효면 `nil`).
4. 생성 코드는 `c.Session == nil` 이면 `{"code": "auth error"}` 를 돌려준다. `Session.Level >= adminLevel`(3) 이면 관리자다.

### 생성 API 의 권한 (model.json)

| 테이블 | 읽기 (목록 / 개수 / 검색 / 단건) | 쓰기 |
|:---|:---|:---|
| `board` (`session` 작성자 + `read: "login"`) | 로그인 사용자는 **모든 글** (비로그인 `auth error`). 목록 / 단건에 작성자 정보 `extra.user` (join) | 생성: 작성자 = `c.Session.Id` (본문 값 무시). 수정 / 삭제 / batch 삭제: **작성자만**, 남의 글은 `auth error` |
| `user` (`protect: ["level"]`) | **본인 행만** (`id = 세션 id` 조건, 남의 `GET /api/user/:id` 는 `auth error`). 관리자(level ≥ 3)는 모든 회원 | 수정 / 삭제는 본인만 (관리자 제외). `level` 은 관리자만 바꾼다. 생성 때 level 은 DB 기본값 `1` |

- `read` 값: `"owner"` (세션 컬럼이 본인인 행만, `session` 이 있을 때 기본), `"login"` (로그인 사용자 모두), `"public"` (비로그인도).
  게시판을 공개하려면 `"read": "public"` 으로 바꾸고 `make gen`.
- 문자열 목록 필터는 **부분 일치**다: `?title=ab` → `b_title like '%ab%' escape '!'` (`%`, `_`, `!` 는 글자 그대로).
  정확히 같은 값만 찾으려면 model.json `table[].compare` 에 `{"name": "title", "type": "="}`.

### 인증 API

gomachine README "인증 API 계약" 과 같다 (Spring / Rust 템플릿도 같은 경로 · 본문 · 상태 코드).

| 요청 | 설명 | 성공 | 실패 |
|:---|:---|:---|:---|
| `POST /api/auth/signup` `{email, password, name}` | 회원가입 (level 1, 비밀번호 8~72자, name 이 없으면 email 앞부분) | 201 `TokenResponse` | 400 검증, 409 이미 있는 email |
| `POST /api/auth/login` `{email, password}` | 로그인 | 200 `TokenResponse` | 401 |
| `GET /api/auth/me` | 내 정보 (토큰 필요) | 200 `user` | 401 (토큰 없음 / 무효 / 회원 없음) |
| `POST /api/upload/index` (multipart field `upload`) | 이미지 업로드 (토큰 필요) | 200 `{filename, url}` | 400 (파일 없음 / 이미지 아님), 401 |
| `GET /api/health` | 상태 확인 | 200 `{"status": "ok"}` | |

`TokenResponse` (`user` = `{id, email, name, level, createddate}`, 생성된 User 응답과 같은 key, 비밀번호 없음):

```json
{ "token": "eyJ...", "tokenType": "Bearer", "expiresInMs": 86400000,
  "user": { "id": 42, "email": "a@b.c", "name": "홍길동", "level": 1, "createddate": "2026-01-01 09:00:00" } }
```

- 에러 본문은 `{status, message, fieldErrors?}` 이고 `status` 는 HTTP 상태 코드와 같다. `fieldErrors` (`{필드: 메시지}`)는 검증 실패(400)에만 있다.
  생성 API 의 HTTP 200 + `code` 에러와 다르다.
- 업로드: 이미지(jpg/jpeg/png/gif/webp, 5MB 이하)만, 저장 이름은 서버가 만든다. `{UPLOAD_PATH}/uploads/{filename}` 에 저장하고
  `url` = `/uploads/{filename}` 으로 정적 서빙한다 (`board.img` 에 `filename` 이나 `url` 을 넣는다).

### JWT 형식 (통일 JWT)

- 알고리즘 **HS256** 고정, 비밀키 `JWT_SECRET` (**32 바이트 이상**, 짧으면 서버가 시작하지 않는다. production 은 `change-me...` 예시 값도 거부)
- 유효 시간 기본 24시간 (`JWT_EXPIRATION_MS`)
- payload:

```json
{ "sub": "42", "id": 42, "email": "a@b.c", "level": 1, "iat": 1767225600, "exp": 1767312000 }
```

`sub` 는 id 문자열, `iat` / `exp` 는 초 단위 정수다. 읽을 때 `id` 가 없으면 `sub` 를 숫자로 쓴다 (다른 백엔드 토큰 호환).

### 생성 API (예: board)

| 요청 | 설명 |
|:---|:---|
| `GET /api/board?page=1&pagesize=10&orderby=createddate desc&title=...` | 목록 `{items, total, code}` (page 가 없으면 전체) |
| `GET /api/board/count` | 개수 `{total, count, code}` |
| `GET /api/board/:id` | 단건 `{item, code}` |
| `GET /api/board/search/{field}?{field}=값` | 같은 값 검색 |
| `POST /api/board`, `POST /api/board/batch` | 생성 (`{id, code}`), batch 는 1000 행까지 |
| `PUT /api/board/:id`, `PATCH /api/board/:id` | 보낸 필드만 수정 |
| `DELETE /api/board/:id`, `DELETE /api/board/batch` `[{"id":1}]` | 삭제 |

- 모두 토큰이 필요하다 (읽기 / 쓰기 권한은 위 표). 예외: `POST /api/user` (gomachine 의 공개 회원가입, level 은 보호 컬럼이라
  DB 기본값 1 로 저장된다). 이 템플릿의 회원가입은 토큰까지 주는 `POST /api/auth/signup` 을 쓴다.
- `user` 의 비밀번호는 응답 / 필터에서 빠지고, 저장할 때 bcrypt 로 해시한다. `level` 은 관리자만 바꾼다 (`protect`).
- 에러는 HTTP 200 + `code: "error" | "auth error"`, 잘못된 요청은 400 `{"code": "bad request"}` 이다.
- 자세한 규칙은 gomachine README 의 "Go REST 동작", "HTTP 계약" 을 본다.

---

## ➕ 테이블 추가하기

1. **스키마** — `db/schema.sql` 에 테이블을 추가한다. 이름 규칙: 테이블 `{name}_tb`, 컬럼 `{접두사}_{이름}`,
   `{접두사}_id` (bigint auto_increment), `{접두사}_createddate` (Insert 때 현재 시각).

   ```sql
   CREATE TABLE IF NOT EXISTS `comment_tb` (
     `c_id` bigint NOT NULL AUTO_INCREMENT,
     `c_board` bigint NOT NULL,
     `c_user` bigint NOT NULL,
     `c_content` text NOT NULL,
     `c_createddate` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
     PRIMARY KEY (`c_id`)
   ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
   ```

2. **DB 반영** — 개발 DB 를 다시 만들거나 (`make db-reset`, 데이터 삭제) 직접 실행한다
   (`docker compose exec -T db mariadb -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < db/schema.sql`).

3. **model.json** — 권한이 필요하면 `table` 에 추가한다 (없으면 로그인한 누구나 읽고 쓴다).

   ```json
   { "name": "comment", "session": [{ "key": "id", "column": "user" }], "read": "login",
     "join": [{ "name": "user", "column": "user" }] }
   ```

   `session` 만 있으면 읽기도 본인 행만(`"owner"`)이다.

4. **생성** — `make gen` → `models/comment.go`, `controllers/rest/comment.go`, `router/routers/comment.go` 가 생기고
   `generated_routes.go` 에 `SetupCommentRoutes` 가 추가된다. `router.go` 는 고치지 않는다.

5. `make test && make run`

---

## 🧩 model.json

```jsonc
{
  "database": "project",            // gomachine 이 읽을 DB 이름 (.env 의 DB_NAME / MARIADB_DATABASE 와 같게)
  "server": "127.0.0.1",            // DB_HOST / DB_USER / DB_PASSWORD 환경변수가 있으면 그것을 쓴다
  "auth": "jwt",                    // 모든 생성 API 가 Session 을 요구한다
  "adminLevel": "3",                // Session.Level 이 이 값 이상이면 관리자
  "languages": { "go": true },
  "goModelFilePath": "./",          // 생성 위치 (model.json 기준)
  "table": [
    { "name": "user", "protect": ["level"] },         // level 은 관리자만 변경, 읽기는 본인(관리자는 전체)
    {
      "name": "board",
      "session": [{ "key": "id", "column": "user" }], // 작성자 = 로그인 사용자, 수정 / 삭제는 작성자만
      "join": [{ "name": "user", "column": "user" }],  // 응답 extra.user 에 작성자 정보
      "read": "login"                                  // 로그인 사용자는 모든 글을 읽는다
    }
  ]
}
```

DB 접속 정보(IP, 비밀번호)는 model.json 에 커밋하지 않는다. `make gen` 이 `.env` 의 `DB_HOST / DB_USER / DB_PASSWORD` 를 넘긴다.
gomachine 사용법(`buildtool-model <dir> [languages] [package] [--force]`, 커스텀 메서드, join 등)은 gomachine README 를 본다.

`make gen` 은 `buildtool-model` 을 PATH 에서 찾고, 없으면 `~/bin/buildtool-model` 을 쓴다 (gomachine 에서 `make install`).
템플릿은 `$HOME/bin/buildtool/` 에서 읽으므로 `HOME` 을 바꾸지 않는다. 다른 위치면 `make gen BUILDTOOL=/path/to/buildtool-model`.
gomachine 은 `.env` 의 `DB_HOST / DB_PORT / DB_NAME / DB_USER / DB_PASSWORD` 로 붙고, 생성 Go 파일을 gofmt 해서 쓴다
(다시 만들어도 diff 가 없다).

---

## 📝 환경 변수

앱은 시작할 때 `.env` 를 읽는다 (이미 설정된 환경변수가 우선). `.env.yml` (선택, `.env.yml.example`) 은 환경변수가 없는 값만 채운다.

| 변수 | 기본값 | 설명 |
|:---|:---|:---|
| `APP_MODE` | `develop` | `develop` / `production` (`--mode production` 인자도 된다) |
| `PORT` | `8007` | HTTP 포트 |
| `DB_HOST` / `DB_PORT` | `127.0.0.1` / `3306` | DB 주소 |
| `DB_USER` / `DB_PASSWORD` | | DB 계정 |
| `DB_NAME` | `MARIADB_DATABASE` | DB 이름 (필수) |
| `DB_DSN` | | 직접 지정하는 MySQL DSN (있으면 위 값 대신) |
| `DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` / `DB_CONN_MAX_LIFETIME_SEC` | `25` / `10` / `300` | 커넥션 풀 |
| `JWT_SECRET` | | HS256 비밀키, **32 바이트 이상 필수** |
| `JWT_EXPIRATION_MS` | `86400000` | 토큰 유효 시간 (24시간) |
| `CORS_ORIGINS` | develop: `http://localhost:5173,http://localhost:3000,http://localhost:9007` | 허용 출처 (쉼표 구분). production 기본값 없음 |
| `UPLOAD_PATH` | `webdata` | 업로드 / 로그 루트. 업로드 이미지는 `{UPLOAD_PATH}/uploads` → `/uploads` 로 서빙 (`/webdata` 서빙은 `log`·`certs` 제외) |
| `DOCUMENT_ROOT` | `dist` | SPA 빌드 결과 (`/*` → `index.html`) |
| `LOG_LEVEL` / `LOG_FILE` / `LOG_CONSOLE` / `LOG_WEB` | `debug`(운영 `info`) / `webdata/log/system.log` / `Y` / `Y` | 로그 |
| `LOG_DATABASE` | `N` | SQL + 파라미터 로그 (bcrypt 해시 등 민감값이 남는다) |
| `TLS_USE` / `TLS_CERT` / `TLS_KEY` | | production 에서 직접 TLS |
| `MAIL_*` / `SMS_*` | | 메일 / SMS 유틸 (선택) |
| `MARIADB_DATABASE` / `MARIADB_USER` / `MARIADB_PASSWORD` / `MARIADB_ROOT_PASSWORD` | | docker compose `db` 서비스 |

CORS 는 `Authorization` 헤더와 `GET, POST, HEAD, PUT, PATCH, DELETE, OPTIONS` 를 허용한다.

---

## 🛠️ Make 명령

| 명령 | 설명 |
|:---|:---|
| `make gen` | gomachine 으로 테이블 코드 생성 (DB 필요) |
| `make db-up` / `make db-down` / `make db-reset` | MariaDB 시작 / 중지 / 볼륨 삭제 후 다시 만들기 |
| `make run` | `go run .` |
| `make build` | `bin/basicproject_go` |
| `make test` / `make vet` / `make fmt` | 테스트 / vet / gofmt |
| `make linux` | 정적 Linux 바이너리 |
| `make docker` | Docker 이미지 (`IMAGE=... TAG=...`) |
| `make docker-up` / `make docker-down` | compose 로 db + api |

---

## 🐳 Docker

```bash
make docker                       # 이미지 빌드 (설정 / 비밀값은 이미지에 없다)
make docker-up                    # db + api (api 는 db healthcheck 뒤에 시작, .env 를 읽는다)
docker run --env-file .env -e DB_HOST=host.docker.internal -p 8007:8007 basicproject_go:latest
```

- 컨테이너는 `APP_MODE=production` 이 기본이다. `CORS_ORIGINS` 와 실제 `JWT_SECRET` 을 넣는다.
- 업로드 / 로그는 `/app/webdata` 볼륨에 남는다.

---

## ✅ 테스트

```bash
make test
```

- `global/jwt`: 토큰 발급 → 해석 (claim 이름 / 형식, sub 대체, 만료 / 다른 키 / HS384 / none 거부), bcrypt
- `global/config`: JWT_SECRET 길이 검사, `.env` 읽기
- `services`: MariaDB 없이 앱을 띄운다. 테스트용 `database/sql` 드라이버 `standin` (`services/standin_test.go`)이
  생성 코드가 보낸 SQL / 값을 기록하고 정해 둔 행을 돌려준다.
  - board: 비로그인 읽기 `auth error`, 로그인 사용자는 남의 글 목록 / 단건을 읽음 (작성자 조건 없음, 작성자 join),
    남의 글 수정 / 삭제 / batch 삭제는 `auth error` 이고 쓰기 SQL 이 없음, 작성자는 수정 / 삭제 가능, 생성 때 작성자 = 세션
  - user: 비관리자 목록 / 개수 / 검색 SQL 에 `u_id = 세션 id`, 남의 단건 `auth error`, 관리자는 조건 없음, `POST /api/user` 의 level 은 1
  - 인증 API: signup 201 (level 1, 해시, 토큰 claim), 중복 409, 검증 400 `fieldErrors`, login 200 / 401, me 200 / 401,
    응답 key (`token, tokenType, expiresInMs, user`, `user` = `id, email, name, level, createddate`)
  - 업로드: `{filename, url}`, `url` 정적 서빙, 이미지 아님 / 파일 없음 400
  - 토큰 처리(없음 / 깨짐 / 만료), CORS preflight(Authorization, PATCH)
- `models`: `WhereClause` SQL 인젝션 방지

---

## 🔒 보안 메모

- `.env`, `.env.yml`, `webdata/` 는 커밋하지 않는다 (`.gitignore`, `.dockerignore`).
- 요청 본문 20MB, 이미지 5MB (`global/upload.MaxImageSize`), 확장자 + 내용 시그니처 검사, 저장 파일명은 서버가 만든다.
- `/webdata/log`, `/webdata/certs` 는 정적 서빙에서 제외한다.
- 생성 모델의 `Where` 는 컬럼 / 연산자를 검증하고 값은 바인딩한다 (`models.WhereClause`). 원시 SQL 은 `Custom.Params` 로 바인딩한다.
- 생성 `GET /api/user` 는 본인 행만 돌려준다 (관리자는 전체). 회원 목록(이메일 포함)은 관리자만 본다.
