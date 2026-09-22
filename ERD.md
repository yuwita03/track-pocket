# PocketTrack API — Backend PRD

## 1. Overview

PocketTrack API adalah REST API untuk aplikasi PocketTrack Mobile.

Backend bertanggung jawab terhadap:

* authentication;
* authorization;
* business logic;
* database;
* validation;
* caching;
* rate limiting;
* logging;
* API security.

Frontend tidak boleh mengakses PostgreSQL atau Redis secara langsung.

---

# 2. Repository

```text
pockettrack-api/
```

Backend berdiri sendiri dan dapat digunakan oleh lebih dari satu client di masa depan.

Contoh:

```text
React Native
     ↓
     API
     ↑
Web Client
```

---

# 3. Tech Stack

* Go
* Gin atau Chi
* PostgreSQL
* Redis
* JWT
* bcrypt / Argon2
* Docker
* Swagger/OpenAPI

---

# 4. Backend Structure

```text
pockettrack-api/
│
├── cmd/
│   └── api/
│
├── internal/
│   ├── config/
│   ├── handler/
│   ├── middleware/
│   ├── service/
│   ├── repository/
│   ├── model/
│   ├── database/
│   └── validation/
│
├── migrations/
├── docs/
├── Dockerfile
├── docker-compose.yml
├── .env.example
├── go.mod
└── README.md
```

---

# 5. API Base URL

Development:

```text
http://localhost:8080/api/v1
```

Production:

```text
https://api.<domain>/api/v1
```

---

# 6. Authentication API

```text
POST /auth/register
POST /auth/login
POST /auth/refresh
POST /auth/logout
GET  /auth/me
POST /auth/forgot-password
POST /auth/reset-password
POST /auth/verify-email
POST /auth/resend-verification
```

## Register

Request:

```text
name
email
password
```

Backend:

1. Validate request.
2. Check email.
3. Hash password.
4. Create user (status: `unverified`).
5. Generate email verification token, kirim email verifikasi.
6. Auto-seed default categories untuk user baru (lihat Section 8.1).
7. Generate authentication credentials.

## Token Policy

```text
Access token (JWT)  : expiry 15 menit
Refresh token       : expiry 7 hari, disimpan di tabel refresh_tokens
```

Refresh token **rotation**: setiap kali `/auth/refresh` dipanggil, refresh token lama di-revoke dan yang baru diterbitkan. Kalau refresh token yang sudah di-revoke dipakai lagi, backend menganggap itu potensi pencurian token dan me-revoke seluruh sesi user tersebut.

## Password Reset

```text
POST /auth/forgot-password   -> kirim email berisi reset token (TTL 15-30 menit)
POST /auth/reset-password    -> validasi token, set password baru
```

## Email Verification

```text
POST /auth/verify-email        -> validasi token dari email
POST /auth/resend-verification -> kirim ulang token verifikasi
```

User yang belum verifikasi email tetap bisa login, tapi endpoint tertentu (opsional, tergantung kebutuhan) bisa dibatasi sampai email terverifikasi.

---

# 7. User API

```text
GET   /users/me
PATCH /users/me
```

User hanya dapat mengakses profil miliknya sendiri.

---

# 8. Category API

```text
GET    /categories
POST   /categories
PATCH  /categories/:id
DELETE /categories/:id
```

Category memiliki:

```text
id
user_id
name
type
is_default
created_at
updated_at
```

Type:

```text
INCOME
EXPENSE
```

## 8.1 Default Categories

`is_default` bersifat **per user**, bukan global. Saat user baru register, backend otomatis meng-insert satu set default category (misal: Makanan, Transportasi, Gaji, dll) dengan `is_default = true` untuk user tersebut.

Category dengan `is_default = true`:

* tidak bisa di-delete;
* nama boleh diubah (opsional, tergantung kebutuhan produk).

## 8.2 Delete Category yang Punya Transaksi

Category **tidak boleh** dihapus secara hard-delete kalau masih punya transaksi terkait. Pilihan:

```text
- Soft delete: tandai category.is_active = false, transaksi lama tetap merujuk ke category tsb.
- Hard delete ditolak (409 Conflict) selama masih ada transaksi terkait.
```

Category yang `is_active = false` tidak muncul lagi di `GET /categories` untuk pemilihan baru, tapi transaksi lama tetap menampilkan nama category-nya.

---

# 9. Transaction API

```text
GET    /transactions
GET    /transactions/:id
POST   /transactions
PATCH  /transactions/:id
DELETE /transactions/:id
```

Transaction:

```text
id
user_id
category_id
type
amount            -- NUMERIC(14,2), jangan pakai FLOAT/DOUBLE
note
transaction_date  -- disimpan sebagai TIMESTAMPTZ (UTC), dikonversi ke local time di frontend
created_at
updated_at
```

## 9.1 Idempotency

`POST /transactions` menerima header opsional:

```text
Idempotency-Key: <client-generated UUID>
```

Kalau key yang sama dikirim ulang (misal karena retry akibat network issue di mobile), backend mengembalikan response transaksi yang sudah dibuat sebelumnya alih-alih membuat duplikat. Key disimpan sementara (misal di Redis, TTL 24 jam).

---

# 10. Transaction Query

Endpoint:

```text
GET /transactions
```

Support:

```text
page
limit
search
type
category_id
start_date
end_date
```

Contoh:

```text
GET /transactions?page=1&limit=20&type=EXPENSE
```

---

# 11. Dashboard API

```text
GET /dashboard/summary
GET /dashboard/category-expenses
```

Summary:

```text
total_income
total_expense
balance
transaction_count
```

Category expenses:

```text
category_id
category_name
amount
percentage
```

---

# 12. Budget API

Fitur tahap kedua. **Scope tahap ini murni CRUD** — belum termasuk notifikasi/alert saat over-budget (dicatat sebagai fitur lanjutan, lihat Section 12.1).

```text
GET    /budgets
POST   /budgets
PATCH  /budgets/:id
DELETE /budgets/:id
```

Budget:

```text
id
user_id
category_id
amount
month
year
created_at
updated_at
```

## 12.1 Future: Budget Alert (out of scope untuk MVP)

Bisa ditambahkan belakangan: notifikasi/flag saat total pengeluaran per category mendekati atau melebihi budget bulan berjalan.

---

# 13. Database

PostgreSQL digunakan sebagai source of truth.

Tables:

```text
users
refresh_tokens
categories
transactions
budgets
```

Relationship:

```text
users
 ├── refresh_tokens
 ├── categories
 ├── transactions
 └── budgets

categories
 └── transactions
```

FK behavior:

```text
transactions.category_id -> categories.id   : ON DELETE RESTRICT (lihat Section 8.2, delete pakai soft-delete)
transactions.user_id     -> users.id        : ON DELETE CASCADE
categories.user_id       -> users.id        : ON DELETE CASCADE
budgets.category_id      -> categories.id   : ON DELETE RESTRICT
refresh_tokens.user_id   -> users.id        : ON DELETE CASCADE
```

---

# 14. Data Isolation

Setiap query terhadap user-owned resource harus mempertimbangkan authenticated user.

Contoh:

```text
WHERE transaction.user_id = authenticated_user_id
```

User tidak boleh:

* membaca transaksi user lain;
* mengubah transaksi user lain;
* menghapus transaksi user lain;
* menggunakan category user lain.

---

# 15. Redis

Redis digunakan setelah core API stabil.

## Dashboard Cache

```text
GET /dashboard/summary
        ↓
      Redis
     /     \
   HIT     MISS
   ↓         ↓
return   PostgreSQL
             ↓
          Redis SET
             ↓
           return
```

Contoh key:

```text
dashboard:summary:{user_id}:{year}:{month}
```

TTL:

```text
300 seconds
```

---

# 16. Cache Invalidation

Cache dashboard harus di-invalidate ketika transaksi yang mempengaruhi dashboard berubah.

Event:

```text
CREATE transaction
UPDATE transaction
DELETE transaction
```

Flow:

```text
Transaction mutation
        ↓
Database update
        ↓
Invalidate dashboard cache
```

---

# 17. Rate Limiting

Rate limiting minimal diterapkan pada:

```text
POST /auth/login
POST /auth/register
POST /auth/refresh
POST /auth/forgot-password
```

Redis dapat digunakan sebagai shared rate-limit storage.

---

# 18. Security

Backend harus:

* menggunakan HTTPS pada production;
* melakukan password hashing;
* memvalidasi seluruh request;
* menggunakan JWT dengan secret yang hanya berada di backend;
* melakukan authorization;
* mengisolasi data berdasarkan user;
* menerapkan rate limiting;
* tidak mengekspos sensitive information;
* tidak menyimpan secret di Git;
* menggunakan environment variables/secrets;
* melakukan refresh token rotation dan deteksi reuse (lihat Section 6).

---

# 19. Error Response

Format error konsisten.

```json
{
  "success": false,
  "message": "Transaction not found",
  "code": "TRANSACTION_NOT_FOUND"
}
```

Status code:

```text
400 Bad Request
401 Unauthorized
403 Forbidden
404 Not Found
409 Conflict
422 Validation Error
429 Too Many Requests
500 Internal Server Error
```

---

# 20. Logging

Log:

```text
HTTP method
path
status
duration
request ID
server errors
authentication failures
```

Jangan log:

```text
password
JWT
refresh token
database password
Redis credentials
```

---

# 21. Documentation

Swagger/OpenAPI:

```text
/api/docs
```

Dokumentasi harus menjelaskan:

* endpoint;
* request;
* response;
* authentication;
* query parameters;
* error response.

---

# 22. Testing

Minimal:

### Authentication

* register;
* login;
* invalid credentials;
* refresh (termasuk skenario refresh token reuse yang harus di-reject);
* logout;
* forgot/reset password;
* email verification.

### Authorization

* user A cannot access user B's transaction;
* user A cannot modify user B's transaction.

### Transaction

* create;
* read;
* update;
* delete;
* filtering;
* pagination;
* idempotency key mencegah duplikasi.

### Category

* delete category yang masih punya transaksi harus ditolak/soft-delete.

### Dashboard

* income calculation;
* expense calculation;
* balance calculation;
* category summary.

---

# 23. Development Phase

## Phase 1

Go server + PostgreSQL.

## Phase 2

Authentication (termasuk password reset & email verification).

## Phase 3

Categories + Transactions.

## Phase 4

Dashboard.

## Phase 5

React Native integration.

## Phase 6

Redis caching.

## Phase 7

Rate limiting.

## Phase 8

Testing + deployment + documentation.

---

# 24. Backend Definition of Done

* REST API berjalan.
* PostgreSQL terintegrasi.
* Authentication bekerja (termasuk reset password & email verification).
* Authorization bekerja.
* Transaction CRUD bekerja (dengan idempotency).
* Category CRUD bekerja (dengan soft-delete untuk category yang punya transaksi).
* Filtering dan pagination bekerja.
* Dashboard bekerja.
* User data isolation bekerja.
* Error handling konsisten.
* Logging tersedia.
* Swagger tersedia.
* Redis caching bekerja.
* Rate limiting tersedia.
* Test utama tersedia.
* Docker setup tersedia.
* Production deployment menggunakan HTTPS.

---

# 25. Repository Independence

Backend harus dapat dijalankan tanpa React Native.

```text
pockettrack-api/
    ↓
docker compose
    ↓
Go API
    ↓
PostgreSQL
    ↓
Redis
```

Frontend harus dapat menggunakan backend tanpa mengetahui implementasi internal Go.

```text
pockettrack-mobile
        ↓
      REST
        ↓
pockettrack-api
```

Contract antara keduanya adalah **HTTP API**, bukan shared source code.