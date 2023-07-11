package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/blkmlk/file-storage/internal/services/api/controllers"
)

func main() {
	ctx := context.Background()

	tmpdir, err := os.MkdirTemp("/tmp", "file-storage")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(tmpdir)

	filePath, err := createTempFile(tmpdir)
	if err != nil {
		log.Fatal(err)
	}
	fileName := filepath.Base(filePath)

	if err := uploadFile(ctx, "127.0.0.1:19090", filePath); err != nil {
		log.Fatal(err)
	}

	data, err := downloadFile(ctx, "127.0.0.1:19090", fileName)
	if err != nil {
		log.Fatal(err)
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	if !bytes.Equal(fileData, data) {
		log.Fatal("bytes are not equal")
	}

	fmt.Println("OK")
}

func createTempFile(dir string) (string, error) {
	content := make([]byte, 1024*1024*10)
	_, err := rand.Read(content)
	if err != nil {
		return "", err
	}

	file, err := os.CreateTemp(dir, "upload")
	if err != nil {
		return "", err
	}
	if _, err = file.Write(content); err != nil {
		return "", err
	}
	_ = file.Close()

	return file.Name(), nil
}

func uploadFile(ctx context.Context, host, filePath string) error {
	uploadLink, err := getUploadURL(ctx, host)
	if err != nil {
		return err
	}

	return sendFile(ctx, filePath, uploadLink)
}

func getUploadURL(ctx context.Context, host string) (string, error) {
	reqUrl := fmt.Sprintf("http://%s/api/v1/upload", host)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var getResp controllers.GetUploadLinkResponse
	if err = json.NewDecoder(resp.Body).Decode(&getResp); err != nil {
		return "", err
	}

	return getResp.UploadLink, nil
}

func sendFile(ctx context.Context, fileName, uploadLink string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	body := multipart.NewWriter(&buffer)

	writer, err := body.CreateFormFile("file", file.Name())
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		return err
	}
	_ = body.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadLink, &buffer)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", body.FormDataContentType())
	req.Header.Set("Content-Length", strconv.Itoa(buffer.Len()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func downloadFile(ctx context.Context, host, name string) ([]byte, error) {
	reqUrl := fmt.Sprintf("http://%s/api/v1/download/%s", host, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
