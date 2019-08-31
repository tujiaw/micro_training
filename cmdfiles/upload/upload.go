package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const maxUploadSize = 5 * 1024 * 1024
const tmpDir = "./tmp"

func main() {
	fromPtr := flag.String("f", "./filename", "from address")
	toPtr := flag.String("t", "localhost:8080/", "to localhost:8080/")
	flag.Parse()

	os.RemoveAll(tmpDir)
	if !exists(tmpDir) {
		if err := os.Mkdir(tmpDir, os.ModePerm); err != nil {
			panic(err)
		}
	}

	from := *fromPtr
	to := *toPtr
	f, err := os.Stat(from)
	if err != nil {
		panic(err)
		return
	}

	if f.IsDir() {
		panic(fmt.Errorf("%s is not file", from))
		return
	}

	pos := strings.Index(to, ":")
	if pos == -1 {
		to = "localhost:8080/" + to
	}
	pos = strings.Index(to, "/")
	if pos == -1 {
		panic(errors.New("to error"))
		return
	}

	url := fmt.Sprintf("http://%s/upload", to[:pos])
	dir := to[pos+1:]

	filename := from
	fileSize := getFileSize(filename)
	if fileSize <= 0 {
		panic("file size error")
	}

	fields := map[string]string{
		"filename": filepath.Base(filename),
		"dir":      dir,
	}

	if fileSize < maxUploadSize {
		err := postFile(filename, url, fields)
		panicIfErr(err)
	} else {
		pathChan := make(chan string, 5)
		go splitFile(filename, pathChan)
		index := 0
		for path := range pathChan {
			index += 1
			fields["multiindex"] = strconv.Itoa(index)
			err := postFile(path, url, fields)
			panicIfErr(err)
		}
	}
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func splitFile(filename string, pathChan chan string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	index := 0
	buf := make([]byte, maxUploadSize)
	r := bufio.NewReader(f)
	randStr := randToken(12)
	for {
		n, err := r.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if 0 == n {
			break
		}

		index += 1
		smallFileName := fmt.Sprintf("%s/multi%s-%d_%s", tmpDir, randStr, index, filepath.Base(filename))
		err = ioutil.WriteFile(smallFileName, buf[:n], os.ModePerm)
		if err != nil {
			panic(err)
		}
		pathChan <- smallFileName
	}
	close(pathChan)
}

func postFile(filename string, targetUrl string, fileds map[string]string) error {
	fmt.Println("post", targetUrl, filename)
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("uploadFile", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	for k, v := range fileds {
		bodyWriter.WriteField(k, v)
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	resp, err := http.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(resp_body))
	return nil
}

func getFileSize(filename string) int64 {
	fi, err := os.Stat(filename)
	if err != nil {
		return -1
	}
	return fi.Size()
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
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
