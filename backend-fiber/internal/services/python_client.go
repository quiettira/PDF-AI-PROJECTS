package services

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type PythonClient struct {
	BaseURL string
}

func NewPythonClient(baseURL string) *PythonClient {
	return &PythonClient{
		BaseURL: baseURL,
	}
}

func (c *PythonClient) Summarize(filePath string, style string) (string, int64, error) {
	start := time.Now() //untuk menghitung waktu prosesnya

	requestURL := c.BaseURL //untuk mempersiapkan request
	if style != "" { //untuk menambahkan query style ke request
		u, err := url.Parse(c.BaseURL)
		if err == nil {
			q := u.Query()
			q.Set("style", style)
			u.RawQuery = q.Encode()
			requestURL = u.String()
		}
	}

	file, err := os.Open(filePath) //untuk membuka file, ambil file dari disk utk dikirim ke python
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	body := &bytes.Buffer{} //untuk menyimpan body request
	writer := multipart.NewWriter(body) //untuk membuat writer

	part, err := writer.CreateFormFile("file", filepath.Base(filePath)) //untuk membuat file form, aplod
	if err != nil {
		return "", 0, err
	}

	_, err = io.Copy(part, file) //untuk menyalin file ke part
	if err != nil {
		return "", 0, err
	}

	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, requestURL, body) //untuk membuat request ke piton
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType()) //untuk menambahkan header

	resp, err := (&http.Client{}).Do(req) //kirim request ke python
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close() 

	b, err := io.ReadAll(resp.Body) //untuk membaca response dari python
	if err != nil {
		return "", 0, err
	}

	return string(b), time.Since(start).Milliseconds(), nil	//kembalikan response, waktu proses, dan error
}
//buat komunikasi ke python ya intinya 