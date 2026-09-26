# Plan: Sembunyikan Judul Native macOS Title Bar yang Bertabrakan dengan Custom Brand AetherGrok

## Ringkasan Masalah
Pada tangkapan layar yang diberikan pengguna:
- Di pojok kiri atas jendela aplikasi macOS, tepat di sebelah kanan traffic lights (tombol merah, kuning, hijau), muncul teks bawaan jendela sistem operasi `"AetherGrok"`.
- Di saat yang bersamaan, elemen custom header web aplikasi (`PanelLeftClose`, icon bintang `Sparkles`, dan logo teks custom `AetherGrok v1.0.0`) dirender di area yang sama karena title bar macOS dibuat transparan (`TitlebarAppearsTransparent: true`).
- Akibatnya, teks bawaan sistem operasi bertabrakan dan menimpa elemen icon/sidebar toggle kustom kita.

## Akar Masalah Teknis
Di file `main.go`, konfigurasi opsi macOS Wails:
```go
Mac: &mac.Options{
    TitleBar: &mac.TitleBar{
        TitlebarAppearsTransparent: true,
        HideTitle:                  false, // <-- Masalah: bernilai false, sehingga macOS tetap merender string Title ("AetherGrok") di atas webview!
        HideTitleBar:              false,
        FullSizeContent:            false,
        UseToolbar:                 false,
    },
    Appearance:           mac.NSAppearanceNameDarkAqua,
    WebviewIsTransparent: false,
    WindowIsTranslucent:  false,
},
```

Jika `TitlebarAppearsTransparent: true` namun `HideTitle: false`, macOS merender teks `Title: "AetherGrok"` langsung pada title bar sistem. Mengubahnya menjadi:
```go
HideTitle: true,
FullSizeContent: true,
```
akan menyembunyikan teks judul bawaan macOS secara native, mempertahankan tombol traffic lights macOS (close/minimize/zoom) yang rapi di pojok kiri atas, dan membiarkan header kustom aplikasi web (`AetherGrok v1.0.0`) ditampilkan secara eksklusif dan jernih tanpa tabrakan.

## Rencana Langkah Eksekusi:
1. **Perbarui `main.go`**:
   - Atur `HideTitle: true` dan `FullSizeContent: true` pada `Mac.TitleBar`.
2. **Periksa padding header di `App.svelte`**:
   - Pastikan padding kiri pada header (`pl-20` atau setara) tetap memberikan ruang aman yang nyaman untuk traffic lights native macOS.
3. **Verifikasi & Build**:
   - Jalankan `go test ./test/... -v`.
   - Jalankan `wails build` atau `go build` untuk memastikan kompilasi backend Go berhasil tanpa error.
4. **Commit**:
   - Catat perubahan ke Git.
