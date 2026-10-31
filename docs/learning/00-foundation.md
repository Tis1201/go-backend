# Bước 0: Nền móng

> Thời lượng gợi ý: 1 đến 2 tuần, chia làm 3 chặng. Đừng vội. Bước này ít dòng code
> nhưng chứa rất nhiều chi tiết mà junior hay bỏ qua, còn senior thì nhìn ra ngay khi review.

## Mục tiêu học

Xong bước này, bạn tự giải thích được:

- Cách tổ chức một monorepo Go + frontend, và vì sao Go có thư mục `internal/`.
- Connection pool là gì, vì sao chỉ tạo một lần.
- Server "chạy được" khác server "chạy được trên production" ở đâu (timeout, shutdown).
- Liveness khác readiness thế nào, và vì sao nhầm hai thứ này có thể làm sập cả hệ thống.
- Dependency injection trong Go (không cần framework) giúp viết test như thế nào.

## Kết quả cần đạt (Definition of Done)

1. `docker compose up -d` chạy PostgreSQL, có healthcheck, dữ liệu nằm trong named volume.
2. `make run` chạy API ở địa chỉ lấy từ biến môi trường (mặc định `:8080`).
3. `GET /healthz` trả `200` khi process còn sống, **không** phụ thuộc vào DB.
4. `GET /readyz` trả `200` khi ping DB thành công trong tối đa 2 giây, ngược lại trả `503`.
   Body là JSON.
5. Thiếu `DATABASE_URL` thì app thoát ngay lúc khởi động, kèm log lỗi rõ ràng (fail fast).
6. Khi nhận Ctrl+C (SIGINT) hoặc SIGTERM: server ngừng nhận request mới, chờ request
   đang chạy xong (tối đa N giây), đóng pool DB, rồi thoát với exit code 0.
7. Log dạng JSON qua `log/slog`. Không có `fmt.Println` hay `log.Println` trong code.
8. Có unit test cho handler `/readyz`, cả trường hợp DB ổn và DB lỗi, chạy được
   **khi không có Postgres**.
9. `go vet ./...` sạch, code đã qua `gofmt`.
10. Người khác clone repo về, làm theo README là chạy được.

## Chia thành 3 chặng

Bạn mới dùng Go khoảng một tuần, nên bước 0 được chia thành 3 chặng. Mỗi chặng chỉ thêm
một hai khái niệm Go mới, và mỗi chặng là **một PR riêng**. PR nhỏ thì review nhanh, sửa
dễ; ở công ty thật cũng nên làm như vậy. DoD 9 (`gofmt`, `go vet`) áp dụng cho mọi chặng.

### Chặng 0.1: Server đầu tiên và test đầu tiên

- **Hướng dẫn từng bước:** [`00-1-walkthrough.md`](00-1-walkthrough.md).
- **Làm:** chạy Postgres bằng Docker Compose (code Go chưa cần kết nối DB). Chạy
  `go mod init`, viết `main.go` với `http.Server` có timeout, endpoint `GET /healthz` trả
  JSON `{"status":"ok"}`, và một test cho `/healthz`. Tạm hard-code địa chỉ `:8080` cũng
  được, chặng 0.2 sẽ đưa nó vào config. Hard-code tạm là chuyện bình thường, miễn là bạn
  biết sau này sẽ thay ở đâu.
- **Khái niệm Go mới:** package và module, handler của `net/http`, `encoding/json`,
  `testing` với `httptest`.
- **Đọc:** mục 1.1, 1.2, 1.3, 1.6, 1.11. [Tour of Go](https://go.dev/tour) phần Basics.
  [*Learn Go with Tests*](https://quii.gitbook.io/learn-go-with-tests): các chương
  Hello, World; Integers; Iteration.
- **Xong khi:** đạt DoD 1, 3; trả lời câu hỏi 5, 8.
- **Branch:** `step-0/1-server`

### Chặng 0.2: Config, pool DB và readiness

- **Làm:** đọc config từ biến môi trường, tạo `pgxpool`, endpoint `GET /readyz`, chuyển
  log sang `slog`, viết table-driven test cho `/readyz` với pinger giả.
- **Khái niệm Go mới:** xử lý lỗi (`if err != nil`, wrap bằng `%w`), interface,
  `context.WithTimeout`, `defer`.
- **Đọc:** mục 1.4, 1.5, 1.7, 1.9, 1.10. *Learn Go with Tests*: Structs, methods &
  interfaces; Pointers & errors; Dependency Injection.
- **Xong khi:** đạt DoD 4, 5, 7, 8; trả lời câu hỏi 2, 3 (trước khi code) và 4, 7
  (sau khi code).
- **Branch:** `step-0/2-readiness`

### Chặng 0.3: Graceful shutdown

- **Làm:** tắt server đúng cách khi nhận SIGINT hoặc SIGTERM, viết Makefile và README.
- **Khái niệm Go mới:** goroutine, channel, `select`, bắt signal.
- **Đọc:** mục 1.8. *Learn Go with Tests*: Concurrency, Select, Context.
  [*Go by Example*](https://gobyexample.com): Goroutines, Channels, Select, Signals.
- **Xong khi:** đạt DoD 2, 6, 10; trả lời câu hỏi 1 (trước khi code) và 6 (sau khi code).
- **Branch:** `step-0/3-shutdown`

---

## Phần 1: Kiến thức nền (đọc trước khi code)

### 1.1 Cấu trúc thư mục

```
.
├── AGENTS.md
├── Makefile
├── docker-compose.yml
├── .env.example
├── backend/
│   ├── go.mod
│   ├── cmd/api/main.go    ← chỉ lắp ráp (wiring): đọc config, tạo pool, tạo server, chạy
│   └── internal/
│       ├── config/         ← đọc và validate biến môi trường
│       └── httpapi/        ← router, handler
└── docs/
```

- **`main` phải mỏng.** `main` là nơi duy nhất biết hết mọi thứ: nó tạo các dependency
  rồi nối chúng lại với nhau. Logic không nằm ở đây vì `main` rất khó test.
- **`internal/`**: compiler của Go cấm code nằm ngoài thư mục cha của `internal/` import
  các package bên trong nó. Nhờ vậy bạn tự do đổi API nội bộ mà không sợ làm vỡ code của ai.
- **Đừng dựng kiến trúc trước khi có code.** Đừng tạo sẵn `domain/`, `usecase/`,
  `repository/`, `pkg/utils/` khi chưa có code cần đến chúng. Cấu trúc nên lớn lên theo
  nhu cầu. Đến bước 2, khi đã có nghiệp vụ thật, ta sẽ bàn chuyện phân lớp.
- **Tên package**: ngắn, viết thường, là danh từ. Tránh `utils`, `common`, `helpers` vì
  những tên đó không nói lên package làm gì. Không đặt tên package là `http` vì sẽ trùng
  với `net/http`.

### 1.2 Go module

- `go mod init <module path>`. Module path nên trùng nơi code thực sự nằm, ví dụ
  `github.com/Tis1201/go-backend/backend`.
- Commit cả `go.mod` lẫn `go.sum`. Tự tìm hiểu: `go.sum` dùng để làm gì, và nó liên quan
  gì đến an toàn chuỗi cung ứng (supply chain)?
- Dòng `go 1.xx` trong `go.mod` là phiên bản ngôn ngữ tối thiểu. Bạn cần ít nhất 1.22 để
  dùng routing pattern dạng `"GET /healthz"`.

### 1.3 PostgreSQL bằng Docker Compose

- Không cài Postgres thẳng vào máy. Cả team chạy cùng một phiên bản thì hết cảnh
  "máy em chạy được".
- **Pin phiên bản**: dùng `postgres:18`, không dùng `postgres:latest`. `latest` có thể
  nhảy lên major version mới vào một ngày đẹp trời, và dữ liệu cũ sẽ không đọc được nữa.
- Dùng **named volume** để dữ liệu còn nguyên sau `docker compose down`. Lưu ý
  `docker compose down -v` sẽ xóa volume. Biết khi nào nên dùng lệnh đó.
- ⚠️ **Bẫy:** nhiều tutorial trên mạng mount volume vào `/var/lib/postgresql/data`. Từ image
  `postgres:18` trở đi, cách bố trí thư mục dữ liệu đã thay đổi. Đọc mục **PGDATA** trên
  trang Docker Hub của image `postgres` để mount cho đúng. Bài học lớn hơn ở đây: **đọc
  docs chính thức của đúng phiên bản bạn dùng**, đừng copy tutorial cũ.
- **Healthcheck** bằng `pg_isready`. Container ở trạng thái "running" chưa có nghĩa là
  Postgres đã sẵn sàng nhận kết nối: Postgres cần vài giây để khởi động, lần chạy đầu còn
  lâu hơn vì phải khởi tạo dữ liệu.
- Lấy user và mật khẩu từ file `.env` (không commit), và commit `.env.example`. Compose
  tự đọc file `.env` nằm cùng thư mục.
- Không viết dòng `version: "3.x"` ở đầu file. Docker Compose hiện tại coi nó là lỗi thời.

### 1.4 Config theo 12-factor

- Những gì khác nhau giữa các môi trường (dev, staging, prod) nằm trong **biến môi
  trường**, không hard-code trong code.
- Đọc config **một lần** lúc khởi động vào một struct, validate, rồi truyền struct đó đi.
  Đừng gọi `os.Getenv` rải rác khắp nơi: vừa khó test, vừa không ai biết app cần những
  biến nào.
- **Fail fast:** thiếu config bắt buộc thì thoát ngay lúc khởi động, đừng đợi đến request
  đầu tiên mới lỗi.
- Bước này chưa cần thư viện config như viper. Thư viện chuẩn là đủ.

Biến môi trường tối thiểu: `DATABASE_URL`, `HTTP_ADDR` (có giá trị mặc định),
`SHUTDOWN_TIMEOUT` (có giá trị mặc định).

### 1.5 Connection pool

- Mở một kết nối tới Postgres (TCP, có thể thêm TLS, rồi xác thực) tốn hàng chục mili giây,
  và mỗi kết nối là một process riêng phía Postgres, tốn RAM. Vì vậy ta không mở kết nối
  mới cho mỗi request mà dùng **pool**: giữ sẵn một số kết nối và cho mượn.
- Dùng `pgxpool` (`github.com/jackc/pgx/v5/pgxpool`). Ta chọn pgx thay vì `database/sql`
  vì hiệu năng tốt hơn, hỗ trợ đầy đủ kiểu dữ liệu của Postgres, và `sqlc` ở bước 1 sinh
  code cho pgx/v5 rất tốt.
- Pool **an toàn khi nhiều goroutine dùng cùng lúc**, nên chỉ tạo **một lần** trong `main`
  rồi truyền vào nơi cần. Không tạo pool trong handler, không dùng biến global.
- ⚠️ **Bẫy:** tạo pool thành công **chưa chắc** đã kết nối được DB. Bạn sẽ tự kiểm chứng
  điều này ở câu hỏi 4.
- Đóng pool khi app tắt.

### 1.6 HTTP server: "chạy được" và "production-ready"

- `http.ListenAndServe(":8080", mux)` chạy được, nhưng server mặc định **không có timeout
  nào cả**. Một client có thể mở kết nối rồi gửi header thật chậm, vài byte mỗi 10 giây,
  và giữ kết nối đó mãi. Vài nghìn client như vậy là server cạn tài nguyên. Kiểu tấn công
  này tên là **Slowloris**.
- Hãy tự tạo `&http.Server{...}` và tìm hiểu 4 trường `ReadHeaderTimeout`, `ReadTimeout`,
  `WriteTimeout`, `IdleTimeout`. Với mỗi trường, viết một câu giải thích nó đo từ thời
  điểm nào đến thời điểm nào.
- Routing của Go 1.22+: `mux.HandleFunc("GET /healthz", ...)` khai báo method và path
  ngay trong pattern. Gọi sai method thì server tự trả `405`. Dự án này không cần
  chi, gin hay echo. Tự kiểm tra: `curl -i -X POST localhost:8080/healthz` trả về gì,
  và header `Allow` chứa gì?

### 1.7 Health check: liveness và readiness

- **Liveness** ("tôi còn sống không?") kiểm tra process có bị treo hay deadlock không.
  Nếu fail, orchestrator (ví dụ Kubernetes) **giết rồi khởi động lại** container.
- **Readiness** ("tôi có sẵn sàng nhận traffic không?") kiểm tra các dependency như DB.
  Nếu fail, load balancer **ngừng gửi traffic** tới instance đó nhưng **không giết** nó.
- Nhầm hai thứ này là một lỗi kinh điển. Câu hỏi 2 bên dưới sẽ cho bạn thấy vì sao.
- Health check phải nhanh và có timeout. Lệnh ping DB phải chạy trong
  `context.WithTimeout`, nếu không thì health check có thể treo đúng lúc hệ thống đang
  gặp sự cố.
- Đừng trả nguyên văn lỗi nội bộ (connection string, hostname...) ra body response.
  Ghi lỗi vào log, còn response chỉ cần cho biết trạng thái.

### 1.8 Graceful shutdown

Khi deploy phiên bản mới, orchestrator gửi SIGTERM cho process cũ. Nếu app chết ngay
lập tức, request đang xử lý dở sẽ bị cắt ngang: người dùng thấy lỗi, còn dữ liệu có thể
bị bỏ lại ở trạng thái nửa vời.

Các việc cần làm (bạn tự sắp thứ tự và giải thích ở câu hỏi 1):

- bắt tín hiệu,
- ngừng nhận kết nối mới và chờ request đang chạy xong, có giới hạn thời gian,
- đóng tài nguyên (pool DB...),
- thoát.

Những API cần tra: `signal.NotifyContext`, `(*http.Server).Shutdown`, `http.ErrServerClosed`.

⚠️ **Bẫy:** đọc kỹ docs của `Shutdown`. Có một câu nói về việc `ListenAndServe` trả về
**ngay lập tức** khi `Shutdown` được gọi. Nếu không để ý câu đó, chương trình sẽ thoát
trước khi request kịp chạy xong, tức là đúng điều mà graceful shutdown muốn tránh.

### 1.9 Dependency injection và khả năng test

- Handler `/readyz` cần "một thứ gì đó ping được DB". Nếu handler nhận thẳng
  `*pgxpool.Pool` thì muốn test nó, bạn phải có Postgres thật.
- Go có câu quen thuộc: **"accept interfaces, return structs"**. Interface được **khai báo
  ở phía dùng nó**, và càng nhỏ càng tốt. Handler chỉ cần `Ping(ctx) error`, nên ta khai
  báo một interface một method ngay trong package `httpapi`. `*pgxpool.Pool` tự động
  thỏa mãn interface đó, vì interface trong Go là ngầm định: không cần từ khóa
  `implements`.
- Để test, dùng `net/http/httptest` (`httptest.NewRecorder`, `httptest.NewRequest`) với
  một pinger giả trả lỗi hoặc không. Test chạy trong vài mili giây và không cần Docker.
- Viết theo kiểu **table-driven test**, cách viết test chuẩn của cộng đồng Go.

### 1.10 Logging có cấu trúc

- Dùng `log/slog` (có trong thư viện chuẩn từ Go 1.21). Handler JSON cho máy đọc, sau
  này đẩy vào hệ thống log tập trung. Lúc dev dùng handler dạng text cũng được.
- Log theo cặp key-value: `logger.Info("server started", "addr", cfg.HTTPAddr)`, đừng nối
  chuỗi. Khi đó bạn lọc và tìm log theo từng field được.
- Logger cũng là một dependency: tạo trong `main`, rồi truyền vào nơi cần.

### 1.11 Test trong Go, bắt đầu từ con số 0

Go có sẵn công cụ test, không cần thư viện ngoài.

- Test nằm trong file có đuôi `_test.go`, đặt cạnh file code và cùng package.
- Mỗi test là một hàm `func TestXxx(t *testing.T)`. Chạy tất cả bằng `go test ./...`,
  thêm `-v` để xem từng test, thêm `-run TestTen` để chỉ chạy một test.
- `t.Errorf` báo lỗi nhưng **chạy tiếp**. `t.Fatalf` báo lỗi rồi **dừng** test đó. Dùng
  `Fatalf` khi các bước sau không còn ý nghĩa, ví dụ không decode được JSON thì kiểm tra
  từng field cũng vô ích.
- Một test thường có 3 phần: chuẩn bị (arrange), thực hiện (act), kiểm tra (assert).

**Test một handler** không cần chạy server thật:

1. Tạo request giả: `httptest.NewRequest(http.MethodGet, "/healthz", nil)`.
2. Tạo đối tượng ghi lại response: `httptest.NewRecorder()`.
3. Gọi thẳng `ServeHTTP(rec, req)`. Nếu gọi qua router (`mux`) thay vì gọi handler,
   bạn test luôn được cả việc đăng ký route và method.
4. Kiểm tra `rec.Code` và `rec.Body`.

**Table-driven test:** khi có nhiều trường hợp cùng dạng, gom chúng vào một bảng. Ví dụ
với một hàm không liên quan đến dự án:

```go
func TestAbs(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{name: "positive", in: 3, want: 3},
		{name: "negative", in: -3, want: 3},
		{name: "zero", in: 0, want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Abs(tc.in); got != tc.want {
				t.Errorf("Abs(%d) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}
```

Muốn thêm trường hợp mới thì chỉ cần thêm một dòng vào bảng. Thông báo lỗi theo mẫu
"got X, want Y" là quy ước của cộng đồng Go.

**Thói quen quan trọng nhất: phải thấy test fail ít nhất một lần.** Viết test xong, cố
tình làm sai code (ví dụ trả `500` thay vì `200`) rồi chạy lại. Nếu test vẫn xanh thì nó
không kiểm tra gì cả. Một test chưa từng đỏ là một test chưa được kiểm chứng.

---

## Phần 2: Câu hỏi

Viết câu trả lời vào mô tả PR, mỗi câu 3 đến 5 câu. Câu nào chưa chắc thì cứ ghi
"mình đoán là... vì...". Một phỏng đoán có lý do vẫn tốt hơn bỏ trống. Senior review cả
cách bạn suy nghĩ, không chỉ review code.

Số thứ tự câu hỏi cố định. Mục "Chia thành 3 chặng" ghi rõ câu nào thuộc chặng nào.

### Trả lời trước khi code

1. **(Chặng 0.3)** Liệt kê theo thứ tự những việc `main()` làm, từ lúc khởi động đến lúc
   thoát. Chỉ cần gạch đầu dòng, chưa cần code.
2. **(Chặng 0.2)** Postgres chết 30 giây. Bạn có 10 pod API, và `livenessProbe` của Kubernetes trỏ vào
   một endpoint **có ping DB**. Mô tả chuyện gì xảy ra theo từng giai đoạn. Nếu thay bằng
   `readinessProbe` thì sao?
3. **(Chặng 0.2)** App có nên ping DB ngay lúc khởi động không? Nêu một lý do để làm,
   một lý do để không làm, rồi chọn.

### Trả lời bằng thực nghiệm (sau khi code)

4. **(Chặng 0.2)** Gọi `pgxpool.New` với `DATABASE_URL` trỏ tới một DB đang tắt. Nó có trả lỗi không?
   Bạn kiểm chứng bằng cách nào?
5. **(Chặng 0.1)** Chạy `nc localhost 8080` rồi không gõ gì. Kết nối được giữ bao lâu
   trước và sau khi bạn đặt `ReadHeaderTimeout`?
6. **(Chặng 0.3)** Tạo tạm một endpoint ngủ 5 giây, gọi nó, rồi nhấn Ctrl+C ngay. Request
   có trả về thành công không? Nếu bỏ đoạn code chờ `Shutdown` chạy xong thì chuyện gì
   xảy ra?
7. **(Chặng 0.2)** Vì sao interface `Pinger` nằm ở package `httpapi` chứ không nằm ở
   package tạo pool?
8. **(Chặng 0.1)** Chuyện gì xảy ra nếu một module khác import một package nằm trong
   `backend/internal/`?

---

## Phần 3: Bài tập, chia theo commit

Mỗi chặng là một branch và một PR. Mỗi commit làm một việc, message theo Conventional
Commits. Test đi chung commit với code mà nó kiểm tra: mỗi commit phải tự đứng được,
checkout commit nào thì `go test ./...` cũng phải xanh.

**Chặng 0.1** (`step-0/1-server`)

1. `chore: add gitignore and env example`
2. `chore(db): add postgres via docker compose`
3. `feat(api): bootstrap http server with liveness endpoint`
   (`go mod init`, `main`, `http.Server` có timeout, `/healthz` và test của nó)

**Chặng 0.2** (`step-0/2-readiness`)

4. `feat(api): load config from environment`
5. `feat(api): add database pool and readiness endpoint` (kèm test với pinger giả)

**Chặng 0.3** (`step-0/3-shutdown`)

6. `feat(api): shut down gracefully on SIGINT and SIGTERM`
7. `chore: add Makefile` (nhớ cập nhật mục Commands trong `AGENTS.md`)
8. `docs: add README with local setup`

Makefile tối thiểu có các target `db-up`, `db-down`, `run`, `test`, `lint`
(`lint` = `gofmt -l` + `go vet`).

---

## Phần 4: Gợi ý theo bậc (chỉ mở khi bí quá 30 đến 45 phút)

<details>
<summary>Config: gợi ý 1</summary>

Viết một hàm `Load() (Config, error)` trong `internal/config`. Hàm đó đọc biến môi
trường, gán giá trị mặc định, validate, rồi trả về struct. `main` gọi hàm này đúng một lần.

</details>

<details>
<summary>Config: gợi ý 2</summary>

- So sánh `os.LookupEnv` với `os.Getenv`: hàm nào phân biệt được "biến chưa được set" với
  "biến được set là chuỗi rỗng"?
- `time.ParseDuration` để đọc giá trị kiểu `"10s"`.
- Gom tất cả lỗi lại bằng `errors.Join` thay vì dừng ở lỗi đầu tiên, để người chạy app
  thấy hết các biến bị thiếu trong một lần.
- Để hàm dễ test, cân nhắc cho nó nhận một hàm `getenv func(string) (string, bool)` thay
  vì gọi thẳng `os.LookupEnv`. Thử nghĩ xem vì sao cách này dễ test hơn.

</details>

<details>
<summary>Graceful shutdown: gợi ý 1</summary>

Bạn cần hai việc chạy song song: một việc chạy server (lời gọi này block), một việc chờ
tín hiệu. Việc nào nên chạy trong goroutine?

</details>

<details>
<summary>Graceful shutdown: gợi ý 2</summary>

- `ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`
- Chạy `ListenAndServe` trong một goroutine. Nếu có lỗi khác `http.ErrServerClosed` thì
  gửi lỗi đó qua một channel.
- `main` dùng `select` để chờ hoặc `<-ctx.Done()`, hoặc lỗi từ channel.
- Khi shutdown, tạo **một context mới có timeout** để truyền cho `Shutdown`. Đừng dùng lại
  `ctx` vừa bị cancel. Vì sao?

</details>

<details>
<summary>Graceful shutdown: gợi ý 3 (khung code)</summary>

```go
func run(ctx context.Context, logger *slog.Logger) error {
	// TODO: load config
	// TODO: create pool
	// TODO: defer pool.Close(): đặt ở đâu thì thứ tự đóng mới đúng?

	srv := &http.Server{
		// TODO: Addr, Handler, các timeout
	}

	errCh := make(chan error, 1) // câu hỏi: vì sao cần buffer 1?
	go func() {
		// TODO: ListenAndServe; bỏ qua http.ErrServerClosed
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	// TODO: tạo shutdownCtx với timeout lấy từ config
	// TODO: srv.Shutdown(shutdownCtx)
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(ctx, logger); err != nil {
		// TODO: log lỗi rồi os.Exit(1)
	}
}
```

Pattern "`main` gọi `run() error`": `main` chỉ lo exit code, mọi thứ khác trả về `error`.
Nhờ vậy các lệnh `defer` trong `run` chạy đầy đủ. Lưu ý `os.Exit` **bỏ qua** mọi `defer`.

</details>

<details>
<summary>Test handler: gợi ý 1</summary>

```go
type fakePinger struct{ err error }

func (f fakePinger) Ping(ctx context.Context) error { return f.err }
```

Mỗi dòng của bảng test gồm tên case, lỗi mà pinger trả về, và status code mong đợi.
Lặp qua bảng, chạy mỗi case bằng `t.Run`.

</details>

<details>
<summary>Test handler: gợi ý 2</summary>

Kiểm tra cả status code lẫn body. Hãy decode body JSON rồi so sánh từng field, đừng so
sánh nguyên chuỗi. Vì sao so sánh chuỗi JSON dễ làm test vỡ dù code không sai?

</details>

<details>
<summary>Makefile và .env</summary>

Make không tự đọc file `.env`. Tìm hiểu `include` và `export` trong GNU Make. Một cách
khác là để app tự đọc `.env`, nhưng hãy cân nhắc: app có nên phụ thuộc vào một file
`.env` không, và 12-factor nói gì về việc này?

</details>

---

## Phần 5: Checklist tự review trước khi mở PR

Mục nào thuộc chặng sau thì để dành đến chặng đó.

- [ ] `gofmt -l .` không in ra gì, `go vet ./...` sạch.
- [ ] `go test ./...` pass, kể cả khi Postgres đang **tắt**.
- [ ] Không có biến global cho pool, logger hay config.
- [ ] Không có `.env` trong git: `git ls-files | grep env` chỉ ra `.env.example`.
- [ ] Không có lỗi nào bị bỏ qua bằng `_` mà không có lý do.
- [ ] Lỗi được wrap kèm ngữ cảnh, ví dụ `fmt.Errorf("create db pool: %w", err)`.
- [ ] Tắt Postgres: `/readyz` trả `503` trong tối đa 2 giây, `/healthz` vẫn trả `200`.
- [ ] Nhấn Ctrl+C khi đang có request chậm: request vẫn trả về, log ghi "shutdown complete",
      exit code là 0.
- [ ] Chạy `docker compose down -v` rồi làm lại từ đầu theo README: mọi thứ vẫn chạy.
- [ ] Đọc lại toàn bộ diff một lượt như thể bạn là người review.

---

## Phần 6: Thử thách thêm (không bắt buộc)

- **Middleware ghi log request** (method, path, status, thời gian xử lý). Bạn sẽ thấy
  `http.ResponseWriter` không cho đọc lại status code đã ghi. Hãy tự tìm cách giải quyết
  (gợi ý: bọc nó lại). Bài này luyện rất tốt về embedding interface.
- **Request ID:** sinh ID cho mỗi request, gắn vào context và log, trả về client qua
  header `X-Request-ID`.
- Thêm `golangci-lint` với một cấu hình tối thiểu.

## Đọc thêm

- [Tour of Go](https://go.dev/tour), [*Learn Go with Tests*](https://quii.gitbook.io/learn-go-with-tests),
  [*Go by Example*](https://gobyexample.com). Ba nguồn này đi kèm suốt bước 0.
- Go docs: package `net/http` (`Server`, routing pattern của `ServeMux`), `os/signal`,
  `log/slog`, `net/http/httptest`.
- Cloudflare blog: *The complete guide to Go net/http timeouts*.
- Mat Ryer: *How I write HTTP services in Go after 13 years* (pattern `run()`).
- *The Twelve-Factor App*, mục III: Config.
- Kubernetes docs: *Configure Liveness, Readiness and Startup Probes*.
- Docker Hub, image `postgres`: mục Environment Variables và PGDATA.
- pgx v5 docs: package `pgxpool`.
