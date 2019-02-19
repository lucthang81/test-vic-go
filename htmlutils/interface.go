package htmlutils

import (
	"mime/multipart"
)

type SaveImageInterface interface {
	SaveImageFile(file multipart.File) (data map[string]interface{}, err error)
}
