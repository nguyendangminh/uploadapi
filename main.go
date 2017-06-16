package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const PORT string = "1808"
const UPLOAD_DIR string = "upload"

func main() {
	http.HandleFunc("/api", UploadHandler)

	// serving static files at UPLOAD_DIR directory
	// For example: http://your_host:port/static/1203.jpg
	fs := http.FileServer(http.Dir(UPLOAD_DIR))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Printf("Listen and serve at http://host:%s/api", PORT)
	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Use POST method instead", http.StatusMethodNotAllowed)
		return
	}

	// parse request
	r.ParseMultipartForm(32 << 20) // left shift, 33554432, 32MB (maximum)
	image, _, err := r.FormFile("image")
	if err != nil {
		log.Println("Bad request:", err.Error())
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer image.Close()

	// save to a file with different name
	rand_name := strconv.FormatInt(time.Now().Unix(), 10) + "_" + strconv.Itoa(rand.Intn(1000)) + ".jpg"
	image_path := filepath.Join(UPLOAD_DIR, rand_name)
	f, err := os.Create(image_path)
	if err != nil {
		log.Println("Create file failed: ", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(f, image)

	// do something with the image by bash shell
	cmd := exec.Command("ls", "-l", image_path)
	var output, errout bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &errout
	if err = cmd.Run(); err != nil {
		log.Println("Execute command failed", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	log.Printf("Stdout:\n%s", output.String())
	log.Printf("Stderr:\n%s", errout.String())

	// returns result in JSON message
	result := fmt.Sprintf(`{"url": "/static/%s"}`, rand_name)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, result)
}
