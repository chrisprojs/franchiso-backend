# Unboost Franchise Service

Aplikasi ini berfungsi untuk mengecek dan menghapus boost franchise yang sudah expired.

## Fitur

- Mengecek boost yang sudah expired berdasarkan `end_date`
- Update `is_boosted` menjadi `false` di PostgreSQL dan Elasticsearch
- Menghapus boost yang expired dari tabel `boosts`

## Cara Menjalankan

### 1. Setup Environment Variables

Pastikan environment variables berikut sudah diset:

#### Option 1: Set Environment Variables secara manual
```bash
# PostgreSQL
PG_ADDR=localhost:5432
PG_USER=your_username
PG_PASSWORD=your_password
PG_DATABASE=your_database

# Elasticsearch
ELASTIC_URL=http://localhost:9200
```

#### Option 2: Gunakan file .env
Buat file `.env` di folder `unboost_franchise` dengan isi:
```bash
# PostgreSQL
PG_ADDR=localhost:5432
PG_USER=your_username
PG_PASSWORD=your_password
PG_DATABASE=your_database

# Elasticsearch
ELASTIC_URL=http://localhost:9200
```

### 2. Menjalankan Aplikasi

#### Linux/Mac
```bash
cd unboost_franchise
chmod +x run.sh
./run.sh
```

#### Windows
```cmd
cd unboost_franchise
run.bat
```

#### Manual
```bash
cd unboost_franchise
go run main.go
```

### 2.1. Testing Aplikasi

Untuk memverifikasi bahwa aplikasi berfungsi dengan benar:

```bash
cd unboost_franchise
go run test.go
```

Test ini akan:
- Mengecek boost yang sudah expired
- Mengecek boost yang masih aktif
- Mengecek franchise dengan `is_boosted = true`
- Mengecek koneksi Elasticsearch

### 3. Menjalankan sebagai Cron Job

Untuk menjalankan secara otomatis, Anda bisa menggunakan cron job:

#### Linux/Mac
```bash
# Edit crontab
crontab -e

# Tambahkan salah satu dari berikut:
# Menjalankan setiap jam
0 * * * * cd /path/to/Franchiso/unboost_franchise && ./run.sh

# Menjalankan setiap hari jam 00:00
0 0 * * * cd /path/to/Franchiso/unboost_franchise && ./run.sh

# Menjalankan setiap 30 menit
*/30 * * * * cd /path/to/Franchiso/unboost_franchise && ./run.sh
```

#### Windows (Task Scheduler)
1. Buka Task Scheduler
2. Create Basic Task
3. Set trigger sesuai kebutuhan (hourly, daily, etc.)
4. Set action: Start a program
5. Program: `cmd.exe`
6. Arguments: `/c "cd /d C:\path\to\Franchiso\unboost_franchise && run.bat"`

## Log Output

Aplikasi akan menampilkan log seperti:

```
Found 3 expired boosts
Successfully processed expired boost 123e4567-e89b-12d3-a456-426614174000 for franchise 987fcdeb-51a2-43d1-9f12-345678901234
Successfully processed expired boost 456e7890-e89b-12d3-a456-426614174001 for franchise 654fcdeb-51a2-43d1-9f12-345678901235
Successfully processed expired boost 789e0123-e89b-12d3-a456-426614174002 for franchise 321fcdeb-51a2-43d1-9f12-345678901236
Successfully checked and removed expired boosts
```

## Struktur Database

Aplikasi ini bekerja dengan tabel berikut:

### Tabel `boosts`
- `id` (UUID) - Primary Key
- `franchise_id` (UUID) - Foreign Key ke tabel franchises
- `start_date` (TIMESTAMP) - Tanggal mulai boost
- `end_date` (TIMESTAMP) - Tanggal berakhir boost
- `is_active` (BOOLEAN) - Status aktif boost
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

### Tabel `franchises`
- `id` (UUID) - Primary Key
- `is_boosted` (BOOLEAN) - Status boost franchise
- ... field lainnya

## Elasticsearch Mapping

Aplikasi ini menggunakan index `franchises` dengan mapping yang sudah didefinisikan untuk field `is_boosted`.

## Error Handling

Aplikasi memiliki error handling yang robust:
- Jika gagal mendapatkan data franchise, akan skip dan lanjut ke boost berikutnya
- Jika gagal update PostgreSQL, akan skip dan lanjut
- Jika gagal update Elasticsearch, akan skip dan lanjut
- Jika gagal delete boost, akan skip dan lanjut

Semua error akan di-log untuk monitoring. 