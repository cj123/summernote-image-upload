package summernote

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type ImageUploadHandler struct {
	uploadDirectory string
	urlPrefix       string
}

// NewImageUploadHandler creates a new image upload handler with a given upload directory and URL prefix.
func NewImageUploadHandler(uploadDirectory, urlPrefix string) *ImageUploadHandler {
	return &ImageUploadHandler{
		uploadDirectory: uploadDirectory,
		urlPrefix:       urlPrefix,
	}
}

// Upload is the endpoint to point the typescript summernote image uploader too.
func (iuh *ImageUploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("image")

	onError := func(err error, what string) {
		log.Printf("Experienced error while doing: %s. Err: %s", what, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	if err != nil {
		onError(err, "Read form file")
		return
	}

	defer file.Close()

	if !strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
		onError(nil, "Invalid MIME type")
		return
	}

	if _, err := os.Stat(iuh.uploadDirectory); os.IsNotExist(err) {
		if err := os.MkdirAll(iuh.uploadDirectory, 0755); err != nil {
			onError(err, "Create upload directory")
			return
		}
	} else if err != nil {
		onError(err, "Check upload directory")
		return
	}

	imageName := uuid.New().String() + filepath.Ext(header.Filename)

	uploadFilename := filepath.Join(iuh.uploadDirectory, imageName)

	f, err := os.OpenFile(uploadFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		onError(err, "Create uploaded file")
		return
	}

	defer f.Close()

	_, err = io.Copy(f, file)

	if err != nil {
		onError(err, "Upload file")
		return
	}

	_, _ = fmt.Fprintf(w, "%s/%s", iuh.urlPrefix, imageName)
}
