# Chặng 0.1: Hướng dẫn từng bước

Bài này đi kèm `00-foundation.md`. Đây là **ví dụ mẫu có giải thích**: mình làm mẫu,
bạn gõ theo và làm thêm các bài "Tự làm". Mức hướng dẫn sẽ giảm dần: chặng 0.2 chỉ
còn khung code, từ chặng 0.3 trở đi chỉ còn gợi ý. Giống học bơi: có phao trước, rồi
bỏ phao dần.

**Quy tắc khi dùng bài này:**

- **Gõ lại, không copy-paste.** Gõ tay thì bạn mới để ý chi tiết, và tay quen dần với
  cú pháp Go.
- Đọc hết phần "Vì sao". Code chỉ là kết quả, phần đáng học là lý do đằng sau.
- Không bỏ qua bài "Tự làm" nào. Chúng là phần bắt buộc của chặng.

---

## 1. Từ NestJS sang Go: đổi cách nghĩ

Với NestJS, bạn gõ `nest new` là có sẵn khung MVC, DI container, decorator, rồi chỉ việc
điền code. **Go cố tình không có thứ đó.** Cộng đồng Go ưu tiên code tường minh hơn
"phép màu": thư viện chuẩn `net/http` chính là framework, và bạn tự nối các phần lại với
nhau trong khoảng 20 dòng ở `main`.

- **Được:** không có gì bị giấu. Đọc `main` là biết app được lắp ráp ra sao, debug dễ,
  khởi động nhanh, build ra đúng một file binary.
- **Mất:** phải gõ nhiều hơn một chút, và **bạn phải tự quyết cấu trúc**. Đó cũng chính
  là kỹ năng bạn đang học.

Bảng đối chiếu để bạn bắt nhịp nhanh:

| NestJS | Go (dự án này) |
|--------|----------------|
| `nest new` | `go mod init` + tự tạo thư mục |
| `package.json` / `package-lock.json` | `go.mod` / `go.sum` |
| `npm install <pkg>` | `go get <pkg>`, sau đó `go mod tidy` |
| `npm run start:dev` | `go run ./cmd/api` |
| `npm test` (Jest) | `go test ./...` (có sẵn, không cần cài gì) |
| Module (`@Module`) | Package (mỗi thư mục là một package) |
| Controller + `@Get('/healthz')` | Hàm handler + `mux.HandleFunc("GET /healthz", ...)` |
| Service + `@Injectable()` | Struct có method, tạo bằng hàm `NewXxx(...)` |
| DI container tự inject | Tự truyền dependency trong `main` (constructor injection) |
| DTO + class-validator | Struct + json tag + hàm validate tự viết |
| Middleware, Guard, Interceptor | Middleware dạng `func(http.Handler) http.Handler` |
| Exception filter | Hàm trả `error`, map lỗi sang HTTP status ở một chỗ |
| `ConfigModule` | Package `config` tự viết, đọc biến môi trường |
| TypeORM / Prisma | `sqlc` + `pgx` (bước 1) |
| `throw new Error()` | `return err`: **lỗi là một giá trị, Go không có exception** |
| `public` / `private` | Tên viết **Hoa** đầu = public (exported), viết **thường** = private |

Dòng "lỗi là một giá trị" là thay đổi lớn nhất khi chuyển sang Go. Mọi hàm có thể lỗi
đều trả về `error` làm giá trị cuối cùng, và bạn kiểm tra ngay bằng `if err != nil`.
Lúc đầu thấy dài dòng, nhưng nhờ vậy nhìn code là biết chỗ nào có thể lỗi, không có lỗi
nào "bay" ngầm qua nhiều tầng như exception.

---

## 2. Cài công cụ

1. **Go**: tải ở <https://go.dev/dl>, kiểm tra bằng `go version` (cần từ 1.22 trở lên).
2. **Docker**: Docker Desktop (macOS, Windows) hoặc Docker Engine (Linux). Kiểm tra bằng
   `docker version` và `docker compose version`.
3. **Editor**: VS Code với extension **Go** (của Go Team at Google). Extension sẽ đề nghị
   cài `gopls` (language server), hãy đồng ý. Bật *format on save*: trong Go, format là
   việc của công cụ, không ai tranh cãi chuyện tab hay space.

---

## 3. Docker từ con số 0

### 3.1 Năm khái niệm đủ dùng cho dự án

| Khái niệm | Hiểu nhanh | So với lập trình |
|-----------|-----------|------------------|
| **Image** | Bản đóng gói chỉ đọc: một hệ điều hành tối giản + phần mềm (ví dụ `postgres:18`) | Giống class |
| **Container** | Một tiến trình chạy từ image. Xóa container thì mọi thay đổi bên trong mất | Giống instance |
| **Volume** | Ổ lưu trữ nằm ngoài container, nên dữ liệu còn khi container bị xóa | Giống database nằm ngoài app |
| **Port mapping** | `máy_bạn:container`. Container có mạng riêng, muốn kết nối vào thì phải mở cổng | Giống port forwarding |
| **Compose** | File YAML mô tả nhiều container, volume, network, chạy tất cả bằng một lệnh | Giống `package.json` cho hạ tầng |

### 3.2 Thử tay với `docker run` trước

Chạy thử bằng tay trước để hiểu Compose sẽ làm hộ bạn những gì:

```bash
docker run --rm -d --name pg-test -e POSTGRES_PASSWORD=secret -p 127.0.0.1:5432:5432 postgres:18
```

- `-d`: chạy nền. `--rm`: container dừng thì tự xóa.
- `-e`: biến môi trường. Image Postgres đọc biến này để đặt mật khẩu cho lần khởi tạo đầu.
- `-p 127.0.0.1:5432:5432`: mở cổng 5432 của container ra cổng 5432 trên máy bạn.

Thử tiếp các lệnh:

```bash
docker ps
```

```bash
docker logs pg-test
```

```bash
docker exec -it pg-test psql -U postgres
```

Trong `psql`, gõ `CREATE TABLE hello (id int);` rồi `\dt` để xem bảng, `\q` để thoát.
Sau đó:

```bash
docker stop pg-test
```

Chạy lại lệnh `docker run` ở trên và kiểm tra: bảng `hello` còn không? Không còn, vì
dữ liệu nằm trong container, mà container đã bị xóa (`--rm`). **Đó là lý do cần volume.**

### 3.3 Viết `docker-compose.yml`

Đặt ở thư mục gốc repo:

```yaml
services:
  db:
    image: postgres:18
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    ports:
      - "127.0.0.1:5432:5432"
    volumes:
      - db-data:<ĐƯỜNG_DẪN>   # Tự làm 1
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 5s
      timeout: 3s
      retries: 10

volumes:
  db-data:
```

Giải thích từng phần:

- `services.db`: một service tên `db`. Sau này thêm `api`, `redis`... cùng file.
- `image: postgres:18`: pin major version (lý do ở mục 1.3 của `00-foundation.md`).
- `${POSTGRES_USER}`: Compose tự đọc file `.env` cùng thư mục và thay giá trị vào.
- `"127.0.0.1:5432:5432"`: **chỉ** mở cổng cho máy của bạn. Nếu viết `"5432:5432"`, Docker
  mở cổng trên mọi network interface, và trên Linux nó còn **đi vòng qua firewall `ufw`**.
  Database dev với mật khẩu yếu sẽ lộ ra mạng Wi-Fi quán cà phê. Đây là lỗi bảo mật có
  thật, rất hay gặp.
- `volumes` ở dưới cùng khai báo một **named volume** do Docker quản lý. Dòng `db-data:...`
  trong service gắn volume đó vào một đường dẫn trong container.
- `healthcheck`: Docker chạy `pg_isready` mỗi 5 giây. Trạng thái container đi từ
  `starting` sang `healthy` khi Postgres thật sự nhận kết nối.

> Nếu máy bạn đã cài sẵn Postgres và chiếm cổng 5432, đổi thành `"127.0.0.1:5433:5432"`
> rồi kết nối qua cổng 5433.

**Tự làm 1:** thay `<ĐƯỜNG_DẪN>` bằng đường dẫn đúng. Đọc mục **PGDATA** trên trang
<https://hub.docker.com/_/postgres>. Lưu ý phần lớn tutorial trên mạng viết cho
Postgres 17 trở về trước.

<details>
<summary>Đáp án Tự làm 1 (chỉ mở sau khi đã tự tra)</summary>

`/var/lib/postgresql`. Từ `postgres:18`, `PGDATA` mặc định là
`/var/lib/postgresql/18/docker` và volume của image chuyển lên `/var/lib/postgresql`. Nhờ
đó sau này nâng cấp major version bằng `pg_upgrade --link` sẽ thuận tiện hơn. Nếu mount
theo tutorial cũ vào `/var/lib/postgresql/data`, container sẽ in lỗi giải thích chuyện
này rồi thoát.

</details>

### 3.4 File `.env` và `.env.example`

`.env` (**không commit**):

```
POSTGRES_USER=bookstore
POSTGRES_PASSWORD=bookstore
POSTGRES_DB=bookstore
```

`.env.example` (**có commit**): cùng các key, giá trị để dạng mẫu. Người mới clone repo
về chỉ cần `cp .env.example .env`. Mật khẩu đơn giản thì chấp nhận được ở môi trường dev,
nhưng tuyệt đối không dùng lại cho môi trường thật.

### 3.5 Các lệnh Compose hằng ngày

```bash
docker compose up -d
```

```bash
docker compose ps
```

```bash
docker compose logs -f db
```

```bash
docker compose exec db psql -U bookstore -d bookstore
```

```bash
docker compose down
```

```bash
docker compose down -v
```

Lần lượt các lệnh: khởi động nền; xem trạng thái (đợi đến khi thấy `healthy`); xem log
liên tục (Ctrl+C để thoát); mở `psql` trong container; dừng và xóa container nhưng
**giữ** volume; dừng và **xóa luôn** volume.

**Tự làm 2:** tạo một bảng trong `psql`, chạy `down` rồi `up -d`. Bảng còn không? Sau
đó chạy `down -v` rồi `up -d`. Giải thích kết quả bằng lời của bạn.

**Tự làm 3:** chạy `down -v`, đổi đường dẫn volume thành `/var/lib/postgresql/data`, chạy
`up -d` rồi đọc `docker compose logs db`. Đọc hiểu thông báo lỗi, sau đó sửa lại đường dẫn
cho đúng. Kỹ năng đọc log lỗi đến cuối là kỹ năng bạn sẽ dùng hằng ngày.

---

## 4. Khởi tạo Go module

Từ thư mục gốc repo:

```bash
mkdir -p backend/cmd/api backend/internal/httpapi
```

```bash
cd backend && go mod init github.com/Tis1201/go-backend/backend
```

`go.mod` vừa tạo trông như sau (số phiên bản tùy máy bạn):

```
module github.com/Tis1201/go-backend/backend

go 1.24.7
```

- `module ...`: tên module. Đường dẫn import của mọi package trong module bắt đầu bằng
  tên này. Ví dụ thư mục `internal/httpapi` được import bằng
  `github.com/Tis1201/go-backend/backend/internal/httpapi`.
- `go 1.24.7`: phiên bản Go tối thiểu để build module.
- **Vì sao có `cmd/api`?** Đây là quy ước: mỗi thư mục con trong `cmd/` là một chương
  trình chạy được. Sau này sẽ có thêm `cmd/worker` (chạy background job) hay `cmd/seed`
  (nạp dữ liệu mẫu), cùng dùng chung code trong `internal/`.

---

## 5. Server đầu tiên

Ta tạo 4 file. Đọc giải thích bên dưới mỗi file.

### 5.1 `backend/internal/httpapi/routes.go`

```go
// Package httpapi contains the HTTP routes and handlers of the API.
package httpapi

import "net/http"

// NewHandler returns the root HTTP handler with every route registered.
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	return mux
}
```

- **`http.Handler` là khái niệm trung tâm của `net/http`.** Đó là một interface chỉ có
  một method: `ServeHTTP(http.ResponseWriter, *http.Request)`. Router, handler, middleware
  đều là `http.Handler`. Hiểu interface này là hiểu 80% cách web server Go vận hành.
- `http.NewServeMux()` tạo router (Go gọi là "mux"). Bản thân mux cũng là một `http.Handler`.
- `"GET /healthz"`: method và path nằm chung trong pattern (Go 1.22+). Gọi sai method thì
  mux tự trả `405`.
- `NewHandler` viết **Hoa** vì package `main` cần gọi nó. `handleHealthz` viết **thường**
  vì chỉ dùng trong package `httpapi`.
- Comment ngay trên tên được export, bắt đầu bằng chính tên đó (`// NewHandler returns...`),
  là quy ước của Go. Công cụ `go doc` đọc các comment này.
- `mux := ...`: khai báo và gán kiểu ngắn gọn, Go tự suy ra kiểu. Chỉ dùng được bên trong hàm.

### 5.2 `backend/internal/httpapi/health.go`

```go
package httpapi

import "net/http"

type healthResponse struct {
	Status string `json:"status"`
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
```

- `healthResponse` đóng vai trò DTO. Tag `` `json:"status"` `` quy định tên field khi
  chuyển sang JSON.
- ⚠️ **Bẫy kinh điển:** field phải viết **Hoa** (`Status`) thì `encoding/json` mới nhìn
  thấy. Nếu viết `status string`, JSON trả về sẽ là `{}` mà không có lỗi nào báo cả.
- Chữ ký `func(w http.ResponseWriter, r *http.Request)` là chữ ký chuẩn của mọi handler:
  `w` để ghi response, `r` chứa request.
- Dùng hằng `http.StatusOK` thay vì số `200` để code dễ đọc hơn.

### 5.3 `backend/internal/httpapi/respond.go`

```go
package httpapi

import (
	"encoding/json"
	"net/http"
)

// writeJSON sends v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// The status line is already sent, so an encoding error can no longer be
	// reported to the client. Milestone 0.2 adds a logger to record it.
	_ = json.NewEncoder(w).Encode(v)
}
```

- ⚠️ **Thứ tự rất quan trọng:** set header → `WriteHeader(status)` → ghi body. Header
  set **sau** `WriteHeader` sẽ bị bỏ qua mà không báo lỗi. Gọi `Write` khi chưa gọi
  `WriteHeader` thì Go tự gửi status `200`.
- `any` là bí danh của `interface{}`, nghĩa là nhận giá trị kiểu gì cũng được.
- `_ = ...` là cố tình bỏ qua lỗi. Quy tắc của dự án là mỗi lần bỏ qua lỗi phải có
  comment giải thích vì sao, như ở trên.
- Vì sao tách thành hàm riêng? Endpoint nào cũng cần trả JSON. Gom vào một chỗ thì cả API
  nhất quán, và sau này muốn đổi (thêm log, đổi header) chỉ cần sửa một nơi.

### 5.4 `backend/cmd/api/main.go`

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Tis1201/go-backend/backend/internal/httpapi"
)

func main() {
	srv := &http.Server{
		Addr:              ":8080", // TODO(0.2): read from config
		Handler:           httpapi.NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
```

- `&http.Server{...}`: tạo struct bằng cách gán theo tên field, `&` lấy con trỏ tới nó.
  Luôn gán theo tên field, đừng dựa vào thứ tự.
- `5 * time.Second`: thời gian trong Go có kiểu riêng (`time.Duration`), không phải số
  mili giây trần trụi. Nhờ vậy không thể nhầm giây với mili giây.
- Import chia hai nhóm, cách nhau một dòng trống: thư viện chuẩn trước, package của dự
  án sau. `goimports` (extension Go tự chạy khi lưu file) sẽ sắp xếp giúp bạn.
- `if err := ...; err != nil`: khai báo biến ngay trong câu `if`, biến `err` chỉ sống
  trong khối `if` đó. Đây là cách viết rất phổ biến trong Go.
- `log` sẽ được thay bằng `slog` ở chặng 0.2. `TODO(0.2)` đánh dấu chỗ sẽ sửa sau.

### 5.5 Chạy thử

Trong thư mục `backend/`:

```bash
go run ./cmd/api
```

Mở một terminal khác:

```bash
curl -i localhost:8080/healthz
```

```bash
curl -i -X POST localhost:8080/healthz
```

Lệnh thứ nhất trả `200` với `{"status":"ok"}`. Lệnh thứ hai trả `405`, kèm header
`Allow: GET, HEAD`. Bạn không viết dòng nào cho `405` cả, mux tự xử lý.

---

## 6. Test đầu tiên của bạn

Đọc mục 1.11 của `00-foundation.md` trước. Sau đó tạo
`backend/internal/httpapi/health_test.go` và điền vào các `TODO`:

```go
package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	// Arrange
	// TODO: tạo handler bằng NewHandler()
	// TODO: tạo request GET /healthz bằng httptest.NewRequest
	// TODO: tạo recorder bằng httptest.NewRecorder

	// Act
	// TODO: gọi ServeHTTP(recorder, request) trên handler

	// Assert
	// TODO: status code phải là http.StatusOK (dùng t.Fatalf nếu sai: vì sao Fatalf?)
	// TODO: header Content-Type phải là "application/json"
	// TODO: decode body vào một biến healthResponse; field Status phải là "ok"
}
```

Test nằm cùng package `httpapi`, nên dùng được cả những thứ viết thường như
`healthResponse`.

Chạy:

```bash
go test ./...
```

```bash
go test -v -run TestHealthz ./internal/httpapi
```

**Tự làm 4:** viết xong test, cố tình sửa handler trả `"okk"`, chạy test và xem nó đỏ ra
sao. Sau đó sửa lại cho xanh.

**Tự làm 5:** chuyển `TestHealthz` sang dạng table-driven với 2 trường hợp: `GET` trả
`200`, `POST` trả `405`. Gợi ý: trường hợp `405` không có body JSON. Test của bạn xử lý
chỗ này thế nào?

---

## 7. Chuẩn viết code Go: những điều cần nhớ ngay

1. **Format do công cụ lo:** `gofmt`. Không cần tranh luận về style.
2. **Đặt tên:**
   - Dùng `camelCase` hoặc `PascalCase`, không dùng `snake_case`.
   - Từ viết tắt viết hoa toàn bộ: `userID`, `httpClient`, `parseURL`, không viết `userId`.
   - Phạm vi càng hẹp thì tên càng ngắn (`r`, `w`, `ctx`, `err`, `i`), phạm vi càng rộng
     thì tên càng rõ nghĩa.
   - Không lặp lại tên package: viết `httpapi.NewHandler`, đừng viết
     `httpapi.NewHTTPAPIHandler`.
3. **Lỗi:**
   - Thêm ngữ cảnh khi trả lỗi lên trên: `return fmt.Errorf("load config: %w", err)`.
   - Message lỗi viết thường, không có dấu chấm cuối câu, vì nó sẽ được nối thành chuỗi
     kiểu `start server: load config: DATABASE_URL is required`.
   - **Mỗi lỗi chỉ xử lý một lần:** hoặc log, hoặc return, không làm cả hai. Làm cả hai
     thì cùng một lỗi xuất hiện nhiều lần trong log, rất khó lần ra gốc.
   - `panic` không dùng cho lỗi thông thường.
4. **Return sớm** thay vì lồng `if/else`:

   ```go
   if err != nil {
   	return err
   }
   // happy path tiếp tục ở mức thụt lề ngoài cùng
   ```

5. **Comment** cho mọi thứ được export, bắt đầu bằng tên của nó.
6. Muốn đọc thêm: *Effective Go*, *Go Code Review Comments*, *Google Go Style Guide*.

---

## 8. Git và commit

```bash
git switch -c step-0/1-server
```

File `.gitignore` ở gốc repo, tối thiểu gồm:

```
.env
bin/
*.out
.DS_Store
```

Ba commit của chặng này:

1. `chore: add gitignore and env example`: `.gitignore`, `.env.example`
2. `chore(db): add postgres via docker compose`: `docker-compose.yml`
3. `feat(api): bootstrap http server with liveness endpoint`: toàn bộ `backend/`, kể cả test

Trước mỗi commit, chạy `git status` và `git diff --staged` để thấy chính xác mình sắp
commit những gì. Đây là thói quen giúp bạn không bao giờ lỡ commit file `.env`.

---

## 9. Trước khi gửi review

- [ ] Xong Tự làm 1 đến 5.
- [ ] `gofmt -l .` không in ra gì, `go vet ./...` và `go test ./...` đều sạch (chạy
      trong `backend/`).
- [ ] `git ls-files` không có `.env`.
- [ ] Mô tả PR (hoặc tin nhắn gửi mình) có:
  - câu trả lời câu hỏi 5 và 8 trong `00-foundation.md`;
  - giải thích bằng lời của bạn, mỗi timeout một câu: `ReadHeaderTimeout`, `ReadTimeout`,
    `WriteTimeout`, `IdleTimeout` đo từ lúc nào đến lúc nào;
  - kết quả Tự làm 2 và 3.
