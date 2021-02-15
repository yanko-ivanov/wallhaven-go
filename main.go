package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	models "main/models"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {

	dbUsername := os.Getenv("DB_USERNAME")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbDatabase := os.Getenv("DB_DATABASE")
	dbPort := os.Getenv("DB_PORT")

	// "wallche:wallchepass@tcp(db:3306)/wallche?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUsername, dbPass, dbHost, dbPort, dbDatabase)

	// println(dsn)
	conn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	sqlDB, err := conn.DB()
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	db = conn

	// Migrate the schema
	db.AutoMigrate(&models.Wallpaper{})

	app := gin.Default()
	app.Static("/img", "./download")
	app.GET("/get", getWallpaper)

	app.Run(":" + os.Getenv("PORT")) // listen and serve on 0.0.0.0:80
}

func getWallpaper(ctx *gin.Context) {

	url := ctx.Request.URL.Query().Get("url")

	var wallpaper models.Wallpaper

	fullpath, thumbPath := "", ""

	db.Where("url = ?", url).First(&wallpaper)
	if wallpaper.ID == 0 {

		fullpath, err := DownloadFile("./download", url)

		if err != nil {
			panic(err)
		}

		thumbPath := ResizeImage(fullpath)

		wallpaper := models.Wallpaper{Url: url, Path: fullpath, ThumbPath: thumbPath}

		db.Create(&wallpaper)

	} else {
		fullpath := wallpaper.Path
		thumbPath := wallpaper.ThumbPath
	}

	ctx.JSON(200, gin.H{
		"full":  ("/img" + fullpath[strings.LastIndex(fullpath, "/"):]),
		"thumb": ("/img" + thumbPath[strings.LastIndex(thumbPath, "/"):]),
	})

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

	uuid := uuid.NewV4()

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
