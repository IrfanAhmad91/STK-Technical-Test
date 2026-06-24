# Panduan Menjalankan Backend (Bahasa Indonesia)

## ⚠️ Masalah: Go Sudah Diinstall Tapi Tidak Terdeteksi

**Penyebab:** Go sudah terinstall di `C:\Program Files\Go`, tetapi belum ditambahkan ke PATH environment variable.

**Solusi Ada 2 Cara:**

---

## ✅ Cara 1: Menggunakan Script Otomatis (TERMUDAH)

Cukup double-click file ini:

```
run_backend.bat
```

Script ini akan:
1. ✅ Otomatis menambahkan Go ke PATH
2. ✅ Mengecek koneksi database
3. ✅ Install dependencies
4. ✅ Menjalankan backend server

**Lokasi file:** `c:\laragon\www\STK - Technical Test Fullstack Web\run_backend.bat`

---

## ✅ Cara 2: Menambahkan Go ke PATH Secara Permanen

### Langkah-langkah:

1. **Buka System Properties**
   - Tekan `Win + X`
   - Pilih "System"
   - Klik "Advanced system settings" (di sebelah kanan)

2. **Buka Environment Variables**
   - Klik tombol "Environment Variables..." (di bagian bawah)

3. **Edit PATH Variable**
   - Di bagian "User variables" atau "System variables"
   - Cari dan pilih variable bernama "Path"
   - Klik "Edit..."

4. **Tambahkan Go ke PATH**
   - Klik "New"
   - Ketik: `C:\Program Files\Go\bin`
   - Klik "OK" pada semua dialog

5. **Restart Terminal/Command Prompt**
   - Tutup semua window PowerShell/CMD yang terbuka
   - Buka terminal baru

6. **Verifikasi Instalasi**
   ```bash
   go version
   ```
   
   Seharusnya muncul: `go version go1.26.4 windows/amd64`

---

## 🚀 Menjalankan Backend

### Opsi A: Menggunakan Batch File (Termudah)

```bash
# Dari File Explorer, double-click:
run_backend.bat
```

### Opsi B: Dari Command Line

```bash
cd "c:\laragon\www\STK - Technical Test Fullstack Web"
go mod tidy
go run cmd/api/main.go
```

### Hasil yang Diharapkan:

```
Successfully connected to database: postgres@localhost:5432/menu_tree_db
Starting server on port 8080...
[GIN-debug] Listening and serving HTTP on :8080
```

---

## ✅ Verifikasi Backend Berjalan

Buka browser dan akses:

```
http://localhost:8080/api/v1/health
```

Atau gunakan curl:

```bash
curl http://localhost:8080/api/v1/health
```

**Response yang diharapkan:**
```json
{
  "status": "healthy",
  "message": "Menu Tree API is running"
}
```

---

## 🔧 Troubleshooting

### Problem 1: "go: command not found" atau "go is not recognized"

**Solusi:**
- Gunakan `run_backend.bat` (sudah include Go ke PATH)
- ATAU ikuti Cara 2 di atas untuk menambahkan Go ke PATH secara permanen
- ATAU restart komputer setelah install Go

### Problem 2: "failed to connect to database"

**Kemungkinan Penyebab:**

1. **PostgreSQL tidak running**
   ```bash
   # Cek service PostgreSQL di Windows Services
   # Atau start dari Laragon control panel
   ```

2. **Database belum dibuat**
   ```bash
   # Jalankan script inisialisasi database
   cd "c:\laragon\www\STK - Technical Test Fullstack Web"
   scripts\init_database.bat
   ```

3. **Password database salah**
   - Edit file `.env` (buat dari `.env.example` jika belum ada)
   - Sesuaikan username/password PostgreSQL Anda

### Problem 3: Port 8080 sudah digunakan

**Solusi 1:** Stop aplikasi lain yang menggunakan port 8080

**Solusi 2:** Ganti port backend
```bash
# Buat file .env
copy .env.example .env

# Edit .env dan tambahkan:
PORT=8081
```

Kemudian jalankan lagi backend.

### Problem 4: "go mod tidy" gagal

**Kemungkinan:**
- Tidak ada koneksi internet
- Firewall/proxy memblokir download dari GitHub

**Solusi:**
```bash
# Coba dengan proxy (jika perlu)
set HTTP_PROXY=http://proxy-server:port
set HTTPS_PROXY=http://proxy-server:port
go mod tidy
```

---

## 📁 Struktur Project

```
STK - Technical Test Fullstack Web\
├── cmd\api\
│   └── main.go               # Entry point aplikasi
├── config\
│   └── database.go           # Konfigurasi database
├── internal\
│   ├── domain\              # Model data
│   ├── repository\          # Akses database
│   ├── service\             # Business logic (Task 3)
│   └── handler\             # HTTP handlers (Task 3)
├── migrations\              # SQL schema database
├── run_backend.bat          # Script untuk jalankan backend
└── go.mod                   # Dependencies Go
```

---

## 🎯 Status Implementasi

- ✅ Task 1: Database schema (SELESAI)
- ✅ Task 2: Backend data layer (SELESAI)
  - ✅ Project structure
  - ✅ Domain models
  - ✅ Repository pattern
  - ✅ PostgreSQL queries
  - ✅ Transaction support
- ⏳ Task 3: Business logic & API handlers (NEXT)
- ⏳ Task 4: Frontend React
- ⏳ Task 5: Integration testing

---

## 📝 Catatan Penting

1. **Selalu gunakan `run_backend.bat`** untuk menjalankan backend jika PATH belum di-set
2. **Database harus running** sebelum start backend
3. **Port 8080** harus tersedia
4. **Restart terminal** setelah mengubah environment variables

---

## 🆘 Butuh Bantuan?

Jika masih ada masalah:

1. Cek apakah Go terinstall: `"C:\Program Files\Go\bin\go.exe" version`
2. Cek PostgreSQL running: `psql -U postgres -c "SELECT version();"`
3. Cek database exists: `psql -U postgres -l | findstr menu_tree_db`
4. Lihat log error lengkap di terminal

---

## 🎉 Backend Siap!

Setelah backend berjalan, Anda bisa:
- ✅ Test health check: `http://localhost:8080/api/v1/health`
- ⏳ Lanjut ke Task 3: Implementasi business logic
- ⏳ Lanjut ke Task 4: Build frontend React

**Selamat coding! 🚀**
