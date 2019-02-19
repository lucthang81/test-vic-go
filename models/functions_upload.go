package models

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/vic/vic_go/utils"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func (models *Models) uploadImage(c martini.Context, r *http.Request) {
	response := &HttpResponse{}
	c.Map(response)
	file, _, err := r.FormFile("upload_file")
	if err != nil {
		response.err = err
		return
	}
	data, err := models.saveImageFile(file)
	file.Close()
	response.response = data
	response.err = err
}

func (models *Models) SaveImageFile(file multipart.File) (data map[string]interface{}, err error) {
	fileName := fmt.Sprintf("%s.png", utils.RandSeq(20))
	filePath := fmt.Sprintf("%s/images/%s", models.mediaFolderAddress, fileName)
	out, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		return nil, err
	}
	// the header contains useful info, like the original file name
	relativeUrl := fmt.Sprintf("images/%s", fileName)
	absoluteUrl := fmt.Sprintf("%s/%s", models.mediaRoot, relativeUrl)
	data = map[string]interface{}{
		"relative_url": relativeUrl,
		"absolute_url": absoluteUrl,
		"media_root":   models.mediaRoot,
	}
	return data, nil
}

func (models *Models) saveImageFile(file multipart.File) (data map[string]interface{}, err error) {
	fileName := fmt.Sprintf("%s.png", utils.RandSeq(20))
	filePath := fmt.Sprintf("%s/images/%s", models.mediaFolderAddress, fileName)
	out, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		return nil, err
	}
	// the header contains useful info, like the original file name
	relativeUrl := fmt.Sprintf("images/%s", fileName)
	absoluteUrl := fmt.Sprintf("%s/%s", models.mediaRoot, relativeUrl)
	data = map[string]interface{}{
		"relative_url": relativeUrl,
		"absolute_url": absoluteUrl,
		"media_root":   models.mediaRoot,
	}
	return data, nil
}
