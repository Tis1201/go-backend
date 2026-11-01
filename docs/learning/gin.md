# Bài phụ: Gin

> Học sau chặng 0.2, trước chặng 0.3. Mọi thứ nằm trong `playground/gin/`; app chính
> trong `backend/` **vẫn dùng stdlib**. Kết quả cuối cùng là ADR 0001: giữ stdlib hay
> chuyển sang Gin.

## Mục tiêu

Xong bài này, bạn:

- Viết lại được `/healthz` và `/readyz` bằng Gin, chạy được, có test.
- Giải thích được Gin làm hộ bạn những gì, và đổi lại bạn phải chấp nhận điều gì.
- Biết 5 cái bẫy của Gin mà người mới hay dính.
- Tự quyết định được bằng ADR: dự án hiệu sách có nên dùng Gin không.

---

## 1. Gin là gì?

Gin gồm 4 phần:

1. **Router** nhanh, dùng cây radix (lấy từ thư viện `httprouter`).
2. **`*gin.Context`**: gói `request`, `response` và nhiều hàm tiện ích vào một chỗ.
3. **Middleware chain**: chuỗi middleware, điều khiển bằng `c.Next()` và `c.Abort()`.
4. **Binding và validation**: `c.ShouldBindJSON(&dto)` kèm tag `binding:"required"`,
   phía dưới dùng thư viện `go-playground/validator`. Giống DTO + class-validator của
   NestJS.

**Điều quan trọng nhất cần nhớ:** `*gin.Engine` có method `ServeHTTP`, nghĩa là **nó
cũng là một `http.Handler`**. Toàn bộ Gin chỉ là một handler lớn. Vì thế:

- Gin cắm được vào `http.Server` có timeout mà bạn đã viết ở chặng 0.1.
- Gin test được bằng `httptest`, giống hệt cách bạn đã làm.

## 2. Bảng đối chiếu

| | stdlib (Go 1.22+) | Gin | NestJS |
|---|---|---|---|
| Khai báo route | `mux.HandleFunc("GET /books", h)` | `r.GET("/books", h)` | `@Get('books')` |
| Chữ ký handler | `func(w http.ResponseWriter, r *http.Request)` | `func(c *gin.Context)` | method của controller |
| Path param | `r.PathValue("id")` | `c.Param("id")` | `@Param('id')` |
| Query | `r.URL.Query().Get("q")` | `c.Query("q")` | `@Query('q')` |
| Trả JSON | `writeJSON(w, 200, v)` (tự viết) | `c.JSON(200, v)` | `return v` |
| Đọc body | `json.NewDecoder(r.Body).Decode(&v)` + validate tự viết | `c.ShouldBindJSON(&v)` + tag `binding` | DTO + class-validator |
| Context cho I/O | `r.Context()` | `c.Request.Context()` (xem bẫy 3) | |
| Middleware | `func(http.Handler) http.Handler` | `func(c *gin.Context)` + `c.Next()` | Guard, Interceptor |
| Nhóm route | tự ghép prefix | `r.Group("/api/v1")` | prefix của controller |

---

## 3. Năm cái bẫy (đã kiểm chứng trên Gin v1.12.0)

Hai bẫy đầu bạn sẽ **tự gặp** khi chạy test ở mục 5, nên ở đây chỉ nêu tên. Đừng mở
phần giải thích trước khi thấy test đỏ.

### Bẫy 1: Gọi sai method thì trả về gì?

<details>
<summary>Mở sau khi test <code>POST /healthz</code> đỏ</summary>

Mặc định Gin trả **404**, không phải 405 như stdlib, vì trường
`HandleMethodNotAllowed` của `Engine` mặc định là `false`. Đặt nó thành `true` thì Gin
trả 405 kèm header `Allow`. Test chặng 0.1 của bạn bắt được chính xác khác biệt này.
Đây là ví dụ đẹp cho câu "test bảo vệ hành vi của API": đổi framework mà không có test
thì có lẽ bạn không bao giờ phát hiện ra.

</details>

### Bẫy 2: Content-Type

<details>
<summary>Mở sau khi test Content-Type đỏ</summary>

`c.JSON` đặt header `application/json; charset=utf-8`, trong khi `writeJSON` của bạn đặt
`application/json`. Cả hai đều đúng chuẩn. Câu hỏi thiết kế: test nên so sánh **nguyên
văn chuỗi**, hay chỉ kiểm tra phần media type `application/json`? (Gợi ý: package
`mime`, hàm `ParseMediaType`.)

</details>

### Bẫy 3: `c` không mang tín hiệu hủy của request

`*gin.Context` có các method của `context.Context` nên truyền thẳng `c` vào
`db.Ping(c)` vẫn compile được. Nhưng với cấu hình mặc định, `c.Done()` trả về `nil`:
client đóng tab thì `c` **không bao giờ báo hủy**. Mình đã thử:

```
default engine: c.Done() == nil -> true
```

Hành vi này chỉ đổi khi bật `engine.ContextWithFallback = true`. Quy tắc an toàn: **luôn
dùng `c.Request.Context()`** cho mọi lời gọi I/O.

### Bẫy 4: `r.Run()` không có timeout

Mã nguồn của `Run` trong Gin v1.12.0:

```go
server := &http.Server{ // #nosec G112
	Addr:    address,
	Handler: engine.Handler(),
}
err = server.ListenAndServe()
```

Không có timeout nào cả. Comment `#nosec G112` để tắt cảnh báo của công cụ quét bảo mật
`gosec`, và G112 chính là cảnh báo **Slowloris** mà bạn đã tự đo ở chặng 0.1. Hầu hết
tutorial Gin trên mạng đều dùng `r.Run(":8080")`. Bạn thì đã biết nên tự tạo
`http.Server`.

### Bẫy 5: `gin.Default()` gắn sẵn những gì?

`gin.Default()` = `gin.New()` + `Logger()` + `Recovery()`.

- `Recovery()` bắt `panic` trong handler và trả 500. Nên dùng.
- `Logger()` in log dạng text của riêng Gin, **không phải** `slog` JSON. App của ta đã
  dùng `slog`, nên dùng `gin.New()` rồi tự chọn middleware.

**Thêm hai điều nên biết:**

- `*gin.Context` được **tái sử dụng** giữa các request (Gin dùng `sync.Pool` để giảm cấp
  phát bộ nhớ). Không dùng `c` trong goroutine chạy tiếp sau khi handler đã return. Nếu
  thật sự cần, dùng `c.Copy()`.
- Mặc định Gin tin mọi proxy, nên `c.ClientIP()` có thể bị giả mạo qua header
  `X-Forwarded-For`. Khi làm rate limit theo IP ở bước 4, ta sẽ cấu hình
  `SetTrustedProxies`.

---

## 4. Thiết lập module playground

Từ thư mục gốc repo:

```bash
mkdir -p playground/gin && cd playground/gin
```

```bash
go mod init github.com/Tis1201/go-backend/playground/gin
```

```bash
go get github.com/gin-gonic/gin github.com/jackc/pgx/v5
```

**Vì sao playground là một module riêng?** Để Gin và các dependency của nó không chui vào
`backend/go.mod`. App chính chỉ chứa những gì nó thật sự dùng.

**Tự làm 1:** trong playground, thử import
`github.com/Tis1201/go-backend/backend/internal/httpapi` để dùng lại interface `Pinger`.
Chuyện gì xảy ra? Đây là câu 5 của chặng 0.1, giờ bạn gặp trong thực tế. Hệ quả: trong
playground, bạn phải **khai báo lại** `Pinger` và `healthResponse`. Với code thử nghiệm
thì chấp nhận được.

---

## 5. Bài tập

### 5.1 `playground/gin/handlers.go`

```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const readyTimeout = 2 * time.Second

// Pinger reports whether a dependency can be reached.
type Pinger interface {
	Ping(ctx context.Context) error
}

type healthResponse struct {
	Status string `json:"status"`
}

func newRouter(logger *slog.Logger, db Pinger) *gin.Engine {
	r := gin.New()
	// TODO: add the Recovery middleware with r.Use(...)

	r.GET("/healthz", handleHealthz)
	r.GET("/readyz", handleReadyz(logger, db))
	return r
}

func handleHealthz(c *gin.Context) {
	// TODO: respond 200 with healthResponse{Status: "ok"} using c.JSON
}

func handleReadyz(logger *slog.Logger, db Pinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: same logic as backend/internal/httpapi/ready.go, but respond with c.JSON.
		//       Which context do you pass to Ping? (Bẫy 3)
	}
}
```

### 5.2 `playground/gin/main.go`

Dùng lại pattern `run() error` từ `backend/cmd/api/main.go`, với 3 điểm khác:

1. Gọi `gin.SetMode(gin.ReleaseMode)` ở đầu `run`. Hãy thử chạy **không có** dòng này
   một lần trước, xem Gin in ra những gì ở chế độ debug.
2. `Addr: ":8081"`, để chạy song song với API stdlib ở cổng `:8080` và so sánh hai bên.
3. Trong playground, đọc `DATABASE_URL` bằng `os.Getenv` cho nhanh. App thật vẫn dùng
   `config.Load`.

**Không dùng `r.Run()`** (bẫy 4): tạo `http.Server` với đủ 4 timeout và truyền router
vào `Handler`.

Chạy từ `playground/gin/`:

```bash
set -a; source ../../.env; set +a; go run .
```

### 5.3 `playground/gin/handlers_test.go`: đem test cũ sang

Đây là phần đáng học nhất của bài.

1. Copy bảng test từ `backend/internal/httpapi/health_test.go` và `ready_test.go`, gộp
   thành một bảng `TestRoutes` có thêm trường `path`.
2. Thay `NewHandler(...)` bằng `newRouter(...)`. Phần còn lại **giữ nguyên**.
3. Thêm hàm sau để Gin không in cảnh báo debug khi chạy test:

   ```go
   func TestMain(m *testing.M) {
   	gin.SetMode(gin.TestMode)
   	m.Run()
   }
   ```

4. Chạy `go test ./...`. Bạn **sẽ** thấy test đỏ. Đọc thông báo lỗi, rồi mới mở phần giải
   thích của bẫy 1 và bẫy 2.
5. Sửa cho xanh. Với mỗi lỗi, tự quyết định: nên sửa **code** hay sửa **test**? Vì sao?

---

## 6. So sánh và câu hỏi

Chạy cả hai server cùng lúc: API stdlib ở `:8080`, playground ở `:8081`. Gọi
`curl -i` vào cả hai và so sánh header.

1. Đếm số dòng của hai bản (`wc -l`). Với 2 endpoint này, Gin bớt cho bạn được bao nhiêu
   dòng? Những dòng đó là gì?
2. Vì sao test cũ chạy được với Gin gần như nguyên vẹn?
3. Liệt kê những khác biệt về **hành vi** giữa hai bản mà test đã bắt được.
4. Với API hiện tại (chưa nhận body), Gin chưa thể hiện điểm mạnh lớn nhất của nó. Theo
   bạn, khi nào Gin sẽ giúp được nhiều hơn? Gợi ý: bước 2, endpoint tạo sách nhận body
   JSON cần validate.

---

## 7. ADR 0001: chọn router HTTP

Tạo `docs/adr/0001-http-router.md` (viết bằng tiếng Anh):

```markdown
# 1. HTTP router for the bookstore API

## Status
Accepted / Rejected / Proposed

## Context
<!-- Ta cần gì từ router: routing, nhóm route, middleware, binding, validation? -->

## Options
<!-- stdlib net/http, Chi, Gin: mỗi lựa chọn 2–3 dòng ưu và nhược,
     dựa trên những gì bạn đã tự thấy trong playground -->

## Decision

## Consequences
<!-- Được gì, mất gì, khi nào nên xem lại quyết định này? -->
```

Không có đáp án đúng duy nhất. Mình chỉ chấm **lập luận**: lý do có dựa trên những gì
bạn đã tự thấy hay không.

---

## 8. Commit

1. `chore(playground): add gin module`: `go.mod`, `go.sum`
2. `feat(playground): port health and readiness endpoints to gin`: `handlers.go`,
   `main.go`, `handlers_test.go`
3. `docs(adr): decide on the http router`: ADR 0001

Trước khi push, chạy trong `playground/gin/`:

```bash
gofmt -l . && go vet ./... && go test ./...
```
