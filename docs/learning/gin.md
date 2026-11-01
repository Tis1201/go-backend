# Chặng 0.2b: Chuyển backend sang Gin

> Quyết định: [ADR 0001](../adr/0001-use-gin.md). Từ chặng này dự án dùng Gin, phục vụ
> qua `http.Server` của chính ta. Bài này giải thích **mỗi dòng cấu hình production có
> mặt vì lý do gì**.

## Mục tiêu

Xong bài này, bạn giải thích được:

- Gin làm hộ bạn những gì, và đổi lại bạn chấp nhận điều gì.
- Vì sao `main.go` không phải sửa dòng nào khi đổi sang Gin.
- 5 cấu hình mặc định của Gin không an toàn cho production, và cách dự án xử lý từng cái.

---

## 1. Gin là gì?

Gin gồm 4 phần:

1. **Router** nhanh, dùng cây radix (lấy từ thư viện `httprouter`).
2. **`*gin.Context`**: gói `request`, `response` và nhiều hàm tiện ích vào một chỗ.
3. **Middleware chain**: chuỗi middleware, điều khiển bằng `c.Next()` và `c.Abort()`.
4. **Binding và validation**: `c.ShouldBindJSON(&dto)` kèm tag `binding:"required"`,
   phía dưới dùng `go-playground/validator`. Giống DTO + class-validator của NestJS.

**Điều quan trọng nhất cần nhớ:** `*gin.Engine` có method `ServeHTTP`, nghĩa là **nó
cũng là một `http.Handler`**. Toàn bộ Gin chỉ là một handler lớn. Vì thế:

- `httpapi.NewHandler` vẫn trả `http.Handler`, nên **`main.go` không đổi một dòng nào**.
  Server có 4 timeout từ chặng 0.1 vẫn giữ nguyên.
- Test viết bằng `httptest` vẫn chạy được.

## 2. Bảng đối chiếu

| | stdlib (Go 1.22+) | Gin | NestJS |
|---|---|---|---|
| Khai báo route | `mux.HandleFunc("GET /books", h)` | `r.GET("/books", h)` | `@Get('books')` |
| Chữ ký handler | `func(w http.ResponseWriter, r *http.Request)` | `func(c *gin.Context)` | method của controller |
| Path param | `r.PathValue("id")` | `c.Param("id")` | `@Param('id')` |
| Query | `r.URL.Query().Get("q")` | `c.Query("q")` | `@Query('q')` |
| Trả JSON | `writeJSON(w, 200, v)` (tự viết) | `c.JSON(200, v)` | `return v` |
| Đọc body | `json.NewDecoder(r.Body).Decode(&v)` + validate tự viết | `c.ShouldBindJSON(&v)` + tag `binding` | DTO + class-validator |
| Context cho I/O | `r.Context()` | `c.Request.Context()` | |
| Middleware | `func(http.Handler) http.Handler` | `func(c *gin.Context)` + `c.Next()` | Guard, Interceptor |
| Nhóm route | tự ghép prefix | `r.Group("/api/v1")` | prefix của controller |

---

## 3. Những gì đã thay đổi

| File | Thay đổi |
|---|---|
| `internal/httpapi/route.go` | `http.NewServeMux()` → `gin.New()` kèm cấu hình production |
| `internal/httpapi/health.go`, `ready.go` | Handler nhận `*gin.Context`, trả bằng `c.JSON` |
| `internal/httpapi/respond.go` | **Xóa**. `c.JSON` thay cho `writeJSON` |
| `internal/httpapi/main_test.go` | **Mới**: `TestMain` đặt `gin.TestMode` |
| `internal/httpapi/health_test.go` | So sánh media type thay vì nguyên văn `Content-Type` |
| `cmd/api/main.go` | **Không đổi** |
| `.env.example` | Thêm `GIN_MODE` |

### Test đã bắt được gì

Khi chạy bộ test cũ với code Gin, chưa sửa gì:

```
--- FAIL: TestHealthz/GET_returns_200
    health_test.go:64: expected application/json, got application/json; charset=utf-8
```

`c.JSON` thêm `; charset=utf-8`. Cả hai cách viết đều đúng chuẩn, nên ta sửa **test**:
dùng `mime.ParseMediaType` để chỉ so sánh phần `application/json`. Quy tắc: khi test đỏ,
hỏi "**hành vi** của API có sai không?". Nếu không sai thì test đang kiểm tra quá chi
tiết, và test mới là thứ cần sửa.

---

## 4. Năm cấu hình mặc định nguy hiểm (đã kiểm chứng trên Gin v1.12.0)

### 4.1 `r.Run()` không có timeout

Mã nguồn của `Run`:

```go
server := &http.Server{ // #nosec G112
	Addr:    address,
	Handler: engine.Handler(),
}
err = server.ListenAndServe()
```

Không có timeout nào. Comment `#nosec G112` để tắt cảnh báo **Slowloris** của công cụ
quét bảo mật `gosec`, đúng kiểu tấn công bạn đã tự đo bằng `nc` ở chặng 0.1. Hầu hết
tutorial Gin đều dùng `r.Run(":8080")`. **Dự án không bao giờ gọi `Run`.**

### 4.2 Sai method trả 404 thay vì 405

`HandleMethodNotAllowed` mặc định là `false`. Ta đặt `true`, và Gin trả:

```
HTTP/1.1 405 Method Not Allowed
Allow: GET
```

Khác với stdlib: route `GET` của Gin **không tự nhận** method `HEAD`. Một số load balancer
dùng `HEAD` để health check; nếu sau này gặp, khai báo thêm `r.HEAD(...)`.

### 4.3 Tin mọi proxy

Mặc định Gin tin mọi proxy (`0.0.0.0/0`), nên bất kỳ client nào cũng có thể giả IP bằng
header `X-Forwarded-For`, và `c.ClientIP()` sẽ trả IP giả. Ở bước 4, rate limit đăng nhập
theo IP sẽ vô dụng nếu bỏ qua điểm này. Ta gọi `SetTrustedProxies(nil)` để không tin proxy
nào. Khi deploy sau một proxy thật (bước 9), ta sẽ khai báo đúng IP của nó.

### 4.4 `c` không mang tín hiệu hủy của request

`*gin.Context` có các method của `context.Context`, nên truyền `c` vào `db.Ping(c)` vẫn
compile được. Nhưng với cấu hình mặc định, `c.Done()` trả `nil`: client đóng tab thì `c`
**không bao giờ báo hủy**. Mình đã thử:

```
default engine: c.Done() == nil -> true
```

Vì vậy `ready.go` dùng `c.Request.Context()`.

### 4.5 `gin.Default()` và chế độ debug

`gin.Default()` = `gin.New()` + `Logger()` + `Recovery()`. `Logger()` in log text của riêng
Gin, trong khi dự án dùng `slog` JSON. Vì vậy ta dùng `gin.New()` và chỉ gắn `Recovery()`.
Ở chặng 0.3 ta sẽ viết middleware log request bằng `slog`.

Ở chế độ debug (mặc định khi không đặt `GIN_MODE`), Gin in ra stdout:

```
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
[GIN-debug] GET    /healthz   --> .../httpapi.handleHealthz (2 handlers)
{"time":"...","level":"INFO","msg":"server started","addr":":8080"}
```

Text lẫn với JSON, nên hệ thống thu log sẽ không parse được. **Mọi môi trường deploy phải
có `GIN_MODE=release`.** Khi dev ở máy thì để `debug` cũng được, vì bảng route hữu ích.

**Thêm một điều nên biết:** `*gin.Context` được tái sử dụng giữa các request (Gin dùng
`sync.Pool`). Không dùng `c` trong goroutine chạy tiếp sau khi handler đã return; nếu cần
thì dùng `c.Copy()`.

Mọi quy tắc trên đã được ghi vào mục **Gin rules** trong `AGENTS.md`.

---

## 5. Bài tập nhỏ (15 phút)

1. Trong `route.go`, comment dòng `r.HandleMethodNotAllowed = true` rồi chạy
   `go test ./...`. Test nào đỏ, và thông báo lỗi là gì? Sau đó bỏ comment.
2. Chạy app với `GIN_MODE=debug`, rồi chạy lại với `GIN_MODE=release`. So sánh stdout.
3. Gửi request `HEAD` bằng `curl -I localhost:8080/healthz` (`-I` là cách đúng để gửi
   `HEAD`; `-X HEAD` sẽ làm curl chờ một body không bao giờ tới). Kết quả có khớp với mục
   4.2 không?

## 6. Câu hỏi kiểm tra

1. Vì sao đổi sang Gin mà `main.go` không phải sửa dòng nào?
2. Nếu một đồng nghiệp viết `db.Ping(c)` thay vì `db.Ping(c.Request.Context())`, code vẫn
   chạy và test vẫn xanh. Vậy lỗi nằm ở đâu, và khi nào nó gây hại?
3. Vì sao `NewHandler` trả `http.Handler` chứ không trả `*gin.Engine`?
