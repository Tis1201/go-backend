# Lộ trình học: Bookstore fullstack

Mục tiêu không phải là "làm xong một trang web bán sách", mà là sau mỗi bước, bạn
**giải thích được** vì sao hệ thống được thiết kế như vậy, đúng như cách một kỹ sư
mid/senior trả lời trong buổi review thiết kế hay buổi phỏng vấn.

## Cách chúng ta làm việc

### Vòng lặp của mỗi bước

1. **Đọc lesson** `docs/learning/NN-*.md`. Đọc phần kiến thức nền trước, chưa code.
2. **Trả lời phần "trước khi code"** rồi gửi mình. Mình sẽ phản biện. Thiết kế sai
   trên giấy thì sửa mất 5 phút, còn sai trong code thì mất cả buổi.
3. **Code theo các commit nhỏ** trên branch `step-N/<tên>`.
4. **Tự review** theo checklist trong lesson, sau đó mở PR. Phần mô tả PR chứa câu
   trả lời các câu hỏi.
5. **Mình review như review thật.** Mỗi comment có nhãn:
   - `blocking:` phải sửa mới merge được
   - `suggestion:` nên sửa, nhưng bạn có thể phản biện
   - `question:` mình hỏi để kiểm tra bạn có hiểu không, chưa chắc là code sai
   - `nit:` chi tiết nhỏ (tên biến, format)
   - `praise:` chỗ làm tốt; nó cho bạn biết nên giữ thói quen nào
6. **Sửa, merge, rồi viết retro 3 dòng** vào `docs/learning/journal.md`: học được gì,
   vướng ở đâu, lần sau sẽ làm khác thế nào.

### Thang gợi ý

Khi bí, nhắn "gợi ý 1", "gợi ý 2" hoặc "gợi ý 3":

| Bậc | Bạn nhận được |
|-----|---------------|
| 1 | Hướng đi, khái niệm cần tìm hiểu |
| 2 | Tên package, hàm, tài liệu cụ thể |
| 3 | Khung code có `TODO`, không phải lời giải |

**Timebox:** nếu bí một chỗ quá 45 phút thì hỏi. Vật lộn với vấn đề giúp bạn học,
còn xoay vòng tại chỗ thì không.

### Ai viết code?

| Loại việc | Ai viết |
|-----------|---------|
| Kỹ năng cốt lõi: concurrency, transaction, auth, cache, schema, thanh toán, quyết định render ở FE | **Bạn**. Agent chỉ gợi ý và review |
| Lần đầu gặp một pattern (CRUD đầu tiên, form đầu tiên, test đầu tiên) | **Bạn** |
| Lặp lại pattern bạn đã tự làm | Agent viết, bạn review **từng dòng** |
| Boilerplate (file config, Makefile) | Agent viết được, nhưng bạn phải đọc hiểu |

Muốn agent viết thì nói rõ **"write it"**.

**Luật bất di bất dịch: không merge dòng code nào mà bạn không giải thích được.**

### Quyết định lớn thì ghi ADR

Mỗi lựa chọn có trade-off (công cụ migration, session hay JWT, có dùng Redis không...)
được ghi thành một ADR ngắn trong `docs/adr/NNNN-ten-quyet-dinh.md`, gồm bối cảnh,
các phương án, lựa chọn và hệ quả. Sáu tháng sau, ADR là thứ trả lời câu "sao hồi đó
lại làm thế này?".

## Lộ trình

Mỗi bước xoay quanh **một câu hỏi lớn**. Xong bước thì bạn phải trả lời được câu đó
bằng lời của mình.

| # | Bước | Câu hỏi lớn |
|---|------|-------------|
| 0 | Nền móng | Một server Go "production-ready" tối thiểu cần những gì? |
| 1 | Thiết kế dữ liệu | Dữ liệu của hiệu sách trông như thế nào, và database bảo vệ tính đúng đắn của nó ra sao? |
| 2 | Catalog API | Nên chia lớp code backend thế nào cho dễ đổi, dễ test, và khi nào chia lớp là thừa? |
| 3 | Frontend nền tảng | Trang nào nên render ở server, trang nào ở client, và vì sao? |
| 4 | Auth & quản trị | Làm sao biết request này của ai, và người đó được làm gì? |
| 5 | Giỏ hàng, đơn hàng, tồn kho | Hai người cùng bấm mua cuốn sách cuối cùng thì ai mua được? |
| 6 | Tìm kiếm | Làm sao gõ "nha gia kim" mà ra "Nhà Giả Kim" trong 10ms? |
| 7 | Hiệu năng & cache | Chậm ở đâu, đo bằng cách nào? Cache cái gì, và khi nào xóa cache? |
| 8 | Bất đồng bộ & thanh toán | Việc nào không cần làm ngay trong request, và làm sao để nó không bao giờ bị mất? |
| 9 | Vận hành production | Hệ thống lỗi lúc 2 giờ sáng thì bạn biết bằng cách nào, và tìm nguyên nhân ra sao? |
| + | AI trong sản phẩm | AI giúp gì cho người mua sách, và tốn bao nhiêu tiền? |

**Bài phụ (sau chặng 0.2):** [Gin](gin.md). Viết lại `/healthz` và `/readyz` bằng Gin
trong `playground/gin/`, so sánh với stdlib, rồi quyết định bằng ADR 0001.

### Bước 0: Nền móng

- **Backend:** layout monorepo, Go modules, `internal/`, config theo 12-factor, `pgxpool`,
  `http.Server` có timeout, routing Go 1.22+, `log/slog`, graceful shutdown.
- **Hệ thống:** Docker Compose, healthcheck, liveness và readiness.
- **Kỹ năng nghề:** commit nhỏ, Conventional Commits, test handler với `httptest`,
  dependency injection không cần framework.

### Bước 1: Thiết kế dữ liệu

- ERD cho sách, tác giả, thể loại, tồn kho, người dùng, đơn hàng.
- Chuẩn hóa dữ liệu, ràng buộc (`PRIMARY KEY`, `FOREIGN KEY`, `UNIQUE`, `CHECK`, `NOT NULL`).
  Dùng database làm lớp bảo vệ cuối cùng.
- Kiểu dữ liệu cho tiền: vì sao không dùng `float`.
- Migrations với `goose`: viết **ADR đầu tiên** ghi lại vì sao chọn goose thay vì golang-migrate hay Atlas.
- `sqlc`: viết SQL thật, sinh code Go type-safe. Không dùng ORM, và hiểu vì sao.
- Seed data để dev.

### Bước 2: Catalog API

- Phân lớp handler → service → store, và khi nào **không** cần lớp service.
- Xử lý lỗi: sentinel error, wrap, `errors.Is`/`errors.As`, map lỗi sang HTTP status.
- Quy ước JSON, validate input, format lỗi thống nhất.
- Phân trang: offset hay keyset (cursor), đo trên dữ liệu lớn.
- Integration test với Postgres thật (testcontainers).
- Viết OpenAPI làm hợp đồng giữa BE và FE.

### Bước 3: Frontend nền tảng

- Next.js App Router: Server Component và Client Component.
- Chiến lược render: static, dynamic, streaming. Caching và revalidation của Next.js.
- API client có type, sinh từ OpenAPI.
- Trang danh sách và chi tiết sách.
- Core Web Vitals (LCP, CLS, INP): đo bằng Lighthouse, tối ưu ảnh bìa với `next/image`.
- Chọn giải pháp styling (ADR).

### Bước 4: Auth & quản trị

- Hash mật khẩu (argon2id, bcrypt) và vì sao không dùng SHA-256.
- Session cookie hay JWT: phân tích trade-off rồi viết ADR.
- Cookie flags (`HttpOnly`, `Secure`, `SameSite`), CSRF, rate limit cho đăng nhập.
- Middleware chain, phân quyền theo vai trò (customer, admin).
- Admin CRUD sách: bạn viết endpoint đầu tiên, agent viết phần lặp lại, bạn review.
- **FE:** form, validate ở cả hai phía, bảo vệ route.

### Bước 5: Giỏ hàng, đơn hàng, tồn kho

- Transaction, isolation level, các hiện tượng lost update, write skew.
- Bài toán "cuốn sách cuối cùng": `SELECT ... FOR UPDATE`, `UPDATE ... WHERE stock >= n`,
  hay optimistic locking. Viết test chạy đồng thời để chứng minh code đúng.
- Idempotency key: user bấm "Đặt hàng" hai lần thì sao?
- State machine của đơn hàng, deadlock và cách tránh.
- `go test -race`.
- **FE:** state của giỏ hàng, optimistic UI và cách rollback khi lỗi.

### Bước 6: Tìm kiếm

- Full-text search của Postgres: `tsvector`, GIN index.
- Tiếng Việt có dấu (`unaccent`), gõ sai chính tả (`pg_trgm`).
- `EXPLAIN ANALYZE`: đọc query plan, biết khi nào index được dùng.
- **FE:** debounce, lưu state tìm kiếm trên URL, hủy request cũ (`AbortController`),
  autocomplete.

### Bước 7: Hiệu năng & cache

- Đo rồi mới tối ưu: load test bằng k6, profile bằng `pprof`.
- N+1 query, tinh chỉnh connection pool.
- HTTP caching (`Cache-Control`, `ETag`).
- Cache-aside: có cần Redis không (ADR)? Invalidation, cache stampede.
- **FE:** phân tích bundle, code splitting, lazy load, memo đúng chỗ, React Profiler.

### Bước 8: Bất đồng bộ & thanh toán

- Background job: chọn job queue (ADR).
- Transactional outbox: không bao giờ mất sự kiện "đơn hàng đã tạo".
- Retry với exponential backoff, xử lý at-least-once nên consumer phải idempotent.
- Webhook thanh toán giả lập: xác thực chữ ký HMAC, xử lý webhook đến trùng hoặc
  sai thứ tự.
- Gửi email xác nhận đơn.

### Bước 9: Vận hành production

- Observability: log có request ID, metrics (Prometheus), tracing (OpenTelemetry).
- Dockerfile multi-stage, image nhỏ và an toàn.
- CI với GitHub Actions: lint, test, build.
- Chạy migration khi deploy, quản lý secret, backup database.
- Deploy lên một môi trường thật.

### Bonus: AI trong sản phẩm

- Tìm kiếm theo ngữ nghĩa với `pgvector` ("sách giống Nhà giả kim").
- Gợi ý sách, tóm tắt đánh giá.
- Kiểm soát chi phí token: cache, batch, chọn model phù hợp.
