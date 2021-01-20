package main

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/disintegration/imaging"

	uuid "github.com/satori/go.uuid"
)

func main() {

	app := gin.Default()
	app.Static("/img", "./download")
	app.GET("/get", func(ctx *gin.Context) {

		fullpath, err := DownloadFile("./download", ctx.Request.URL.Query().Get("url"))

		if err != nil {
			panic(err)
		}

		thumbPath := ResizeImage(fullpath)

		ctx.JSON(200, gin.H{
			"full":  "/img" + fullpath[strings.LastIndex(fullpath, "/"):],
			"thumb": "/img" + thumbPath[strings.LastIndex(thumbPath, "/"):],
		})
	})
	app.Run() // listen and serve on 0.0.0.0:8080
}

func ResizeImage(path string) string {

	src, err := imaging.Open(path)

	if err != nil {
		panic(err)
	}

	src = imaging.Resize(src, 200, 0, imaging.Lanczos)

	extension := path[strings.LastIndex(path, "."):]
	filename := path[:strings.LastIndex(path, ".")]

	thumbPath := filename + "_thumb" + extension

	imaging.Save(src, thumbPath)

	return thumbPath
}

func DownloadFile(filepath string, url string) (string, error) {
	resp, err := http.Get(url)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	uuid, err := uuid.NewV4()

	if err != nil {
		return "", err
	}

	extension := url[strings.LastIndex(url, "."):]
	fullpath := filepath + "/" + uuid.String() + extension
	out, err := os.Create(fullpath)

	if err != nil {
		return "", err
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	if err != nil {
		return "", err
	}

	return fullpath, nil

}
