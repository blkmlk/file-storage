package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/blkmlk/file-storage/internal/services/api/controllers"
)

func main() {
	ctx := context.Background()

	err := uploadFile(ctx, "127.0.0.1:19090", "/tmp/test.mp4")
	if err != nil {
		log.Fatal(err)
	}
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
