package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	// 1. Koneksi ke RDS MySQL
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", dbUser, dbPass, dbHost, dbName)
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Println("Gagal konek DB:", err)
	} else {
		// Buat tabel jika belum ada
		db.Exec(`CREATE TABLE IF NOT EXISTS laporan (id INT AUTO_INCREMENT PRIMARY KEY, lokasi VARCHAR(255), foto_url VARCHAR(255))`)
		fmt.Println("Koneksi RDS Berhasil!")
	}

	// 2. Routing Web
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		
		// Ambil data dari RDS untuk ditampilkan
		rows, _ := db.Query("SELECT lokasi, foto_url FROM laporan")
		var laporanList []map[string]string
		for rows != nil && rows.Next() {
			var lokasi, foto string
			rows.Scan(&lokasi, &foto)
			laporanList = append(laporanList, map[string]string{"lokasi": lokasi, "foto_url": foto})
		}
		
		tmpl.Execute(w, laporanList)
	})

	http.HandleFunc("/upload", handleUpload)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server jalan di port:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	lokasi := r.FormValue("lokasi")
	file, header, err := r.FormFile("foto")
	if err != nil {
		http.Error(w, "Gagal baca file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3. Upload ke S3
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		http.Error(w, "Gagal buat session S3", http.StatusInternalServerError)
		return
	}

	uploader := s3manager.NewUploader(sess)
	filename := fmt.Sprintf("%d-%s", time.Now().Unix(), header.Filename)
	
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		http.Error(w, "Gagal upload ke S3", http.StatusInternalServerError)
		return
	}

	// 4. Simpan URL S3 ke RDS
	_, err = db.Exec("INSERT INTO laporan (lokasi, foto_url) VALUES (?, ?)", lokasi, result.Location)
	if err != nil {
		http.Error(w, "Gagal simpan ke database", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}