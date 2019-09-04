package main

import (
	"bufio"
	"bytes"
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

	"github.com/tujiaw/goutil"
)

const maxUploadSize = 5 * 1024 * 1024

var tmpDir = filepath.Join(os.TempDir(), "cmdfiles")

func main() {
	fmt.Println(os.TempDir())
	if len(os.Args) == 1 {
		fmt.Println("usage: upload <command> [<args>]")
		return
	}

	uploadCommand := flag.NewFlagSet("upload", flag.ExitOnError)
	uploadFromPtr := uploadCommand.String("from", "", "local file path")
	uploadToPtr := uploadCommand.String("to", "localhost:8080/", "to localhost:8080/")

	downloadCommand := flag.NewFlagSet("download", flag.ExitOnError)
	downloadFromPtr := downloadCommand.String("from", "", "remote file path")
	downloadToPtr := downloadCommand.String("to", "", "local file dir")

	deleteCommand := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteFromPtr := deleteCommand.String("from", "", "remote file path")

	switch os.Args[1] {
	case "upload":
		uploadCommand.Parse(os.Args[2:])
	case "download":
		downloadCommand.Parse(os.Args[2:])
	case "delete":
		deleteCommand.Parse(os.Args[2:])
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}

	if uploadCommand.Parsed() {
		uploadFile(*uploadFromPtr, *uploadToPtr)
	} else if downloadCommand.Parsed() {
		downloadFile(*downloadFromPtr, *downloadToPtr)
	} else if deleteCommand.Parsed() {
		deleteFile(*deleteFromPtr)
	} else {
		panic("command error")
	}
}

func uploadFile(from string, to string) {
	if len(from) == 0 {
		panic(errors.New("local file path is empty!"))
	}

	f, err := os.Stat(from)
	if err != nil {
		panic(err)
	}

	if f.IsDir() {
		panic(fmt.Errorf("%s is not file", from))
	}

	if !goutil.Exists(tmpDir) {
		if err := os.Mkdir(tmpDir, os.ModePerm); err != nil {
			panic(err)
		}
	}

	tmp, err := os.Open(tmpDir)
	names, _ := tmp.Readdirnames(-1)
	for _, name := range names {
		os.RemoveAll(filepath.Join(tmpDir, name))
	}

	pos := strings.Index(to, ":")
	if pos == -1 {
		to = "localhost:8080/" + to
	}
	pos = strings.Index(to, "/")
	if pos == -1 {
		panic(errors.New("to error"))
	}

	url := fmt.Sprintf("http://%s/upload", to[:pos])
	dir := to[pos+1:]

	filename := from
	fileSize := goutil.GetFileSize(filename)
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

func downloadFile(from string, to string) {
	if len(from) == 0 {
		panic(errors.New("remote file path is empty!"))
	}

	from = fmt.Sprintf("http://localhost:8080/files/%s", from)
	pos := strings.LastIndex(from, "/")
	if pos == len(from)-1 {
		pos = strings.LastIndex(from[:len(from)-1], "/")
	}

	if len(to) == 0 {
		to = "./"
	}
	to = to + from[pos+1:]

	fmt.Printf("download from:%s, to:%s\n", from, to)
	resp, err := http.Get(from)
	if err != nil {
		panic(err)
	}

	os.Remove(to)
	buf := make([]byte, maxUploadSize/2)
	total := 0
	for {
		n, err := resp.Body.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		if n == 0 {
			break
		}
		if err = goutil.WriteFileAppend(to, buf[:n]); err != nil {
			panic(err)
		}
		total += n
		fmt.Printf("\r%s\t", FormatBytes(float64(total)))
	}
	fmt.Println()
	fmt.Println("SUCCESS")
}

func deleteFile(from string) {
	if len(from) == 0 {
		panic("from is empty")
	}
	url := ""
	pos := strings.Index(from, ":")
	if pos == -1 {
		url = fmt.Sprintf("http://localhost:8080/delete/%s", from)
	} else {
		pos2 := strings.Index(from[pos:], "/")
		if pos2 == -1 {
			panic("from path error")
		}
		url = fmt.Sprintf("%s:%s/delete/%s", from[:pos], from[pos+1:pos2], from[pos2+1:])
	}
	fmt.Println("delete url:", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(resp_body))
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
	uuidv4, _ := goutil.Uuidv4()
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
		smallFileName := fmt.Sprintf("%s/%s-%d_%s", tmpDir, uuidv4, index, filepath.Base(filename))
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

func FormatBytes(size float64) string {
	const KB = float64(1024)
	const MB = float64(KB * KB)
	const GB = float64(MB * MB)

	trimZero := func(str string) string {
		for i := len(str) - 1; i >= 0; i-- {
			if str[i] != 48 && str[i] != 46 {
				return str[:i+1]
			}
		}
		return str
	}

	switch {
	case size < KB:
		return "1 KB"
	case size >= KB && size < MB:
		return trimZero(fmt.Sprintf("%.2f", size/KB)) + " KB"
	case size >= MB && size < GB:
		return trimZero(fmt.Sprintf("%.2f", size/MB)) + " MB"
	default:
		return trimZero(fmt.Sprintf("%.2f", size/GB)) + " GB"
	}
}
