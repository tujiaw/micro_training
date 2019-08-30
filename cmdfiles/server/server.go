package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const maxUploadSize = 10 * 1024 * 1024
const uploadPath = "./upload"

func main() {
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

		if fileName[:5] == "multi" {
			fileName = getMultiName(fileName)
			if len(fileName) == 0 {
				panic("file name error")
			}
			newPath := filepath.Join(uploadPath, fileName)
			err = writeFileAppend(newPath, fileBytes)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			newPath := filepath.Join(uploadPath, fileName)
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
	fi, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fi.Close()
	_, err = fi.Write(fileBytes)
	return err
}
