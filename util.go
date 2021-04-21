package summernote

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/net/html"
)

// MigrateBase64ImageToDisk takes a base64 encoded image and writes it out to the directory specified, using a uuid.UUID
// as its filename. The filename is returned.
func MigrateBase64ImageToDisk(directory, data string) (string, error) {
	data = strings.TrimPrefix(data, "data:")
	split := strings.SplitN(data, ";", 2)

	if len(split) != 2 {
		return "", fmt.Errorf("base64 image did not have correct format")
	}

	mimeType := split[0]
	base64Encoded := strings.TrimPrefix(split[1], "base64,")

	decoded, err := base64.StdEncoding.DecodeString(base64Encoded)

	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", err
	}

	exts, err := mime.ExtensionsByType(mimeType)

	if err != nil {
		return "", err
	}

	if len(exts) == 0 {
		return "", fmt.Errorf("no valid mime type found")
	}

	filename := uuid.New().String() + exts[0]

	return filename, os.WriteFile(filepath.Join(directory, filename), decoded, 0644)
}

func recurseImagesInHTML(n *html.Node, fn func(n *html.Node)) {
	switch n.Data {
	case "img":
		fn(n)
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		recurseImagesInHTML(child, fn)
	}
}

// MigrateDiskImageToBase64 reads an image from path and returns it as a base64 encoded image.
func MigrateDiskImageToBase64(path string) (string, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("data:%s;base64,%s", mime.TypeByExtension(filepath.Ext(path)), base64.StdEncoding.EncodeToString(data)), nil
}

// InlineImages takes inHTML and looks at all <img> tags, replacing their src attributes with a base64 encoded variant.
func InlineImages(inHTML, imageURLPrefix, imageUploadPath string) (outHTML string, numSuccessful int, err error) {
	outHTML, err = processImagesInHTML(inHTML, func(n *html.Node) {
		for attrIndex := range n.Attr {
			if n.Attr[attrIndex].Key == "src" {
				imageSource := n.Attr[attrIndex].Val

				if strings.HasPrefix(imageSource, imageURLPrefix) {
					var str string

					str, err = MigrateDiskImageToBase64(filepath.Join(imageUploadPath, filepath.Base(imageSource)))

					if err != nil {
						log.Printf("Unable to inline image: %s, error: %s", imageSource, err)
					} else {
						n.Attr[attrIndex].Val = str
						numSuccessful++
					}
				}
			}
		}
	})

	return outHTML, numSuccessful, err
}

// DeInlineImages takes inHTML and looks at all <img> tags, reading all base64 inlined images and writing them to disk,
// replacing the source attributes in the process.
func DeInlineImages(inHTML, imageURLPrefix, imageUploadPath string) (outHTML string, numSuccessful int, err error) {
	outHTML, err = processImagesInHTML(inHTML, func(n *html.Node) {
		for attrIndex := range n.Attr {
			if n.Attr[attrIndex].Key == "src" {
				imageSource := n.Attr[attrIndex].Val

				if strings.HasPrefix(imageSource, "data:image/") {
					filename, err := MigrateBase64ImageToDisk(imageUploadPath, imageSource)

					if err == nil {
						n.Attr[attrIndex].Val = fmt.Sprintf("%s/%s", imageURLPrefix, filename)
						numSuccessful++
					} else {
						log.Printf("Unable to de-inline image, error: %s", err)
					}
				}
			}
		}
	})

	return outHTML, numSuccessful, err
}

func processImagesInHTML(inHTML string, processFunc func(n *html.Node)) (outHTML string, err error) {
	n, err := html.Parse(strings.NewReader(inHTML))

	if err != nil {
		return outHTML, err
	}

	recurseImagesInHTML(n, processFunc)

	buf := new(bytes.Buffer)

	if err := html.Render(buf, n); err != nil {
		return outHTML, err
	}

	outHTML = buf.String()

	return outHTML, nil
}
