package file

import (
	"io"
	"mime/multipart"
	"net/http"
)

func DetectContentType(file multipart.File) (string, error) {
	buf := make([]byte, 512)

	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(buf[:n]), nil
}

func MimeExtension(ct string) string {
	switch ct {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "video/mp4":
		return ".mp4"
	case "video/mkv":
		return ".mkv"
	default:
		return ""
	}
}
