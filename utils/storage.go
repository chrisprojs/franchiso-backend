package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/disintegration/imaging"
)

// Helper function to upload file to storage proxy
func UploadToStorageProxy(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", err
	}
	writer.Close()

	resp, err := http.Post("http://localhost:8081/upload", writer.FormDataContentType(), body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var result struct {
		FileUrl string `json:"fileUrl"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.FileUrl, nil
}

// CropImageToSquare receives an image file from multipart.FileHeader, crops it to a square, and returns the cropped buffer along with the file format ("jpeg"/"png").
func CropImageToSquare(fileHeader *multipart.FileHeader) (*bytes.Buffer, string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	// Detect file format
	ext := strings.ToLower(fileHeader.Filename)
	var img image.Image
	var format string
	if strings.HasSuffix(ext, ".jpg") || strings.HasSuffix(ext, ".jpeg") {
		img, err = jpeg.Decode(file)
		format = "jpeg"
	} else if strings.HasSuffix(ext, ".png") {
		img, err = png.Decode(file)
		format = "png"
	} else {
		return nil, "", fmt.Errorf("format file tidak didukung")
	}
	if err != nil {
		return nil, "", err
	}

	// Crop menjadi persegi
	min := img.Bounds().Dx()
	if img.Bounds().Dy() < min {
		min = img.Bounds().Dy()
	}
	cropped := imaging.CropCenter(img, min, min)

	// Encode hasil crop ke buffer
	buf := new(bytes.Buffer)
	if format == "jpeg" {
		err = jpeg.Encode(buf, cropped, nil)
	} else {
		err = png.Encode(buf, cropped)
	}
	if err != nil {
		return nil, "", err
	}
	return buf, format, nil
}

// BufferToFileHeader mengubah buffer hasil crop menjadi *multipart.FileHeader agar bisa diupload
func BufferToFileHeader(buf *bytes.Buffer, filename, format string) *multipart.FileHeader {
	// Buat multipart writer ke buffer baru
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filename)
	part.Write(buf.Bytes())
	writer.Close()

	// Buat multipart.Reader dari body
	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(int64(body.Len()))
	files := form.File["file"]
	if len(files) > 0 {
		return files[0]
	}
	return nil
}
