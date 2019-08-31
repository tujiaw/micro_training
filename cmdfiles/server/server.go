package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const maxUploadSize = 10 * 1024 * 1024
const uploadPath = "./upload"

func main() {
	if !exists(uploadPath) {
		if err := os.Mkdir(uploadPath, os.ModePerm); err != nil {
			panic(err)
		}
	}

	http.HandleFunc("/upload", uploadFileHandler())

	fs := http.FileServer(http.Dir(uploadPath))
	http.Handle("/files/", http.StripPrefix("/files", fs))

	log.Println("listen on 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RequestURI)
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		}

		fileName := r.PostFormValue("filename")
		dir := r.PostFormValue("dir")
		dir = filepath.Join(uploadPath, dir)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			renderError(w, "INVALID_DIR", http.StatusBadRequest)
			return
		}

		newPath := filepath.Join(dir, fileName)
		multiindex, err := strconv.Atoi(r.PostFormValue("multiindex"))
		if err != nil {
			multiindex = 0
		}

		file, _, err := r.FormFile("uploadFile")
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}
		defer file.Close()

		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		if multiindex >= 1 {
			if multiindex == 1 {
				os.Remove(newPath)
			}
			err = writeFileAppend(newPath, fileBytes)
			if err != nil {
				fmt.Println(err)
				renderError(w, "WRITE_FILE_APPEND_ERROR", http.StatusInternalServerError)
			}
		} else {
			err = ioutil.WriteFile(newPath, fileBytes, os.ModePerm)
			if err != nil {
				fmt.Println(err)
			}
		}
		w.Write([]byte("SUCCESS"))
	})
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	log.Println("ERROR", message)
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}

func getMultiName(filename string) string {
	index := -1
	for i, c := range filename {
		if c == '_' {
			index = i
			break
		}
	}
	if index >= 0 {
		return filename[index+1:]
	}
	return ""
}

func writeFileAppend(filename string, fileBytes []byte) error {
	fi, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file error, %v", err)
	}
	defer fi.Close()
	_, err = fi.Write(fileBytes)
	return err
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
