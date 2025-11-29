# Integrasi Data Provinsi Indonesia untuk Filtering Lokasi Franchise

## Deskripsi
Implementasi ini mengintegrasikan file JSON provinsi Indonesia (`indonesia-province-simple.json`) ke dalam sistem untuk filtering lokasi franchise berdasarkan batas provinsi yang sebenarnya. Sistem menggunakan **hanya batas provinsi** untuk filtering, menghilangkan kompleksitas algoritma lama.

## Perubahan yang Dibuat

### 1. Struct untuk Data GeoJSON
- `IndonesiaProvince`: Representasi satu provinsi dari file GeoJSON
- `IndonesiaProvincesCollection`: Koleksi lengkap data provinsi
- `provincesData`: Variabel global untuk menyimpan data provinsi yang sudah dimuat

### 2. Fungsi Utama

#### `LoadProvincesData()`
- Memuat data provinsi dari file `indonesia-province-simple.json`
- **Lazy loading**: Dipanggil otomatis saat pertama kali dibutuhkan
- Menyimpan data ke variabel global `provincesData`

#### `FindProvinceByName(cityName string)`
- Mencari provinsi berdasarkan nama kota
- **Lazy loading**: Memuat data provinsi jika belum dimuat
- Menggunakan mapping kota-ke-provinsi yang telah didefinisikan
- Fallback ke pencarian berdasarkan kemiripan nama

#### `IsCoordinateInProvince(lat, lng float64, province *IndonesiaProvince)`
- Mengecek apakah koordinat berada dalam batas provinsi
- Menggunakan algoritma ray casting untuk polygon
- Mendukung geometry type Polygon dan MultiPolygon

#### `GetProvinceBoundaries(province *IndonesiaProvince)`
- Mengembalikan bounding box (min/max lat/lng) dari provinsi
- Berguna untuk optimasi pencarian

#### `PointInPolygon(lat, lng float64, coordinates [][][]float64)`
- Implementasi algoritma ray casting untuk mengecek titik dalam polygon

### 3. Modifikasi Fungsi `GetFranchiseLocations`
- **Hanya menggunakan batas provinsi** untuk filtering
- Kriteria inclusion yang disederhanakan:
  - **Hanya**: Jika koordinat dalam batas provinsi → include
- Menghapus algoritma lama (distance-based, viewport-based, address-based filtering)

### 4. Mapping Kota ke Provinsi
Mapping komprehensif kota-kota utama Indonesia ke provinsi masing-masing, termasuk:
- Kota-kota besar (Jakarta, Bandung, Surabaya, dll.)
- Ibukota provinsi
- Kota-kota penting lainnya

## Keuntungan

1. **Akurasi Tinggi**: Filtering berdasarkan batas provinsi yang sebenarnya
2. **Sederhana**: Hanya satu kriteria filtering - dalam batas provinsi atau tidak
3. **Cakupan Luas**: Dapat menemukan franchise di seluruh provinsi
4. **Performance**: Lazy loading - data dimuat hanya saat dibutuhkan
5. **Maintainable**: Kode lebih sederhana dan mudah dipahami

## Penggunaan

Sistem akan otomatis menggunakan batas provinsi ketika parameter `city` diberikan dalam request:

```
GET /franchise/locations?brand=KFC&city=Jakarta
```

Sistem akan:
1. **Lazy load** data provinsi jika belum dimuat
2. Mencari provinsi untuk kota "Jakarta" (DKI JAKARTA)
3. Menggunakan batas provinsi DKI Jakarta untuk filtering
4. Mengembalikan **hanya** franchise yang berada dalam batas provinsi tersebut

## File yang Dimodifikasi

- `service/google_maps.go`: Implementasi utama dengan lazy loading
- `indonesia-province-simple.json`: Data GeoJSON provinsi (sudah ada)

## Catatan Teknis

- **Lazy Loading**: Data provinsi dimuat otomatis saat pertama kali dibutuhkan
- Data GeoJSON mendukung geometry type Polygon dan MultiPolygon
- Algoritma ray casting digunakan untuk point-in-polygon testing
- Mapping kota-ke-provinsi dapat diperluas sesuai kebutuhan
- Error handling untuk kasus file tidak ditemukan atau parsing gagal
- **Tidak ada loading di startup** - aplikasi start lebih cepat
