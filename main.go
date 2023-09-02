package main

import (
	"context"
	"encoding/json"
	"fmt"
	"homework2/conn"
	"homework2/models"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-kivik/couchdb"
	"github.com/go-kivik/kivik"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var Client *kivik.Client
var DB *kivik.DB

func AddDataIntoDb(value models.Student) {
	id := uuid.New()
	rev, err := DB.Put(context.TODO(), id.String(), value)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Data inserted with revision %s\n", rev)
}

func PrintOutJSONasString(parm1 models.Student) string {
	res2B, _ := json.Marshal(parm1)
	return string(res2B)
}

func GetDataFromDB(id string) models.Student {
	row := DB.Get(context.TODO(), id)
	var st models.Student
	row.ScanDoc(&st)
	// data := PrintOutJSONasString(st)
	return st
}

func RgisterRouter(r *gin.Engine) {
	router := r.Group("api/v1")
	router.POST("/students", func(ctx *gin.Context) {
		var req_body models.Student
		ctx.BindJSON(&req_body)
		AddDataIntoDb(req_body)
		body := "Success"
		ctx.JSON(http.StatusOK, body)
	})

	router.GET("/students/:id", func(c *gin.Context) {
		id := c.Param("id")
		fmt.Println("id : ", id)
		row := GetDataFromDB(id)
		c.JSON(http.StatusOK, row)
	})

	router.PUT("/students/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		body := DB.Get(context.TODO(), id)
		var req_body models.Student
		body.ScanDoc(&req_body)
		ctx.BindJSON(&req_body)
		DB.Put(context.TODO(), id, req_body)
		ctx.JSON(http.StatusOK, "Success")
	})

	router.DELETE("/students/:id/:rev", func(ctx *gin.Context) {
		id := ctx.Param("id")
		rev := ctx.Param("rev")
		DB.Delete(context.TODO(), id, rev)
		ctx.JSON(http.StatusOK, "Delete Success")
	})

	router.GET("/students/class/:value", func(ctx *gin.Context) {
		value := ctx.Param("value")
		rows, _ := DB.Query(context.TODO(), "_design/flitterbyclass", "_view/new-view", kivik.Options{
			"key": value,
		})

		var results []interface{}
		for rows.Next() {
			var value interface{}
			if err := rows.ScanValue(&value); err != nil {
				ctx.JSON(500, gin.H{"error": "Error scanning key"})
				return
			}
			results = append(results, value)
		}
		ctx.JSON(200, results)

	})

	router.GET("/students/class", func(ctx *gin.Context) {
		rows, _ := DB.Query(context.TODO(), "_design/flitterbyclass", "_view/new-view")

		var results []interface{}
		for rows.Next() {
			var value interface{}
			if err := rows.ScanValue(&value); err != nil {
				ctx.JSON(500, gin.H{"error": "Error scanning key"})
				return
			}
			results = append(results, value)
		}
		ctx.JSON(200, results)

	})

	router.GET("students/name/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		rows, _ := DB.Query(context.TODO(), "_design/student_name", "_view/student_name", kivik.Options{
			"key": name,
		})
		var results []interface{}
		for rows.Next() {
			var value interface{}
			if err := rows.ScanValue(&value); err != nil {
				ctx.JSON(500, gin.H{"error": "Error scanning key"})
				return
			}
			results = append(results, value)
		}
		ctx.JSON(200, results)
	})

	router.POST("/upload/:id/:rev", func(ctx *gin.Context) {
		id := ctx.Param("id")
		rev := ctx.Param("rev")

		file, header, err := ctx.Request.FormFile("content")
		if err != nil {
			ctx.String(http.StatusBadRequest, fmt.Sprintf("file err : %s", err.Error()))
			return
		}

		fileExt := filepath.Ext(header.Filename)
		originalFileName := strings.TrimSuffix(filepath.Base(header.Filename), filepath.Ext(header.Filename))
		now := time.Now()
		filename := strings.ReplaceAll(strings.ToLower(originalFileName), " ", "-") + "-" + fmt.Sprintf("%v", now.Unix()) + fileExt

		out, err := os.Create("images/" + filename)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			log.Fatal(err)
		}

		fileopen, _ := os.Open("./images/" + filename)
		attachment := &kivik.Attachment{
			Filename:    filename,
			ContentType: "image/jpeg",
			Content:     fileopen, // Replace with the actual binary data from your file
		}

		DB.PutAttachment(context.TODO(), id, rev, attachment)
		os.Remove("./images/" + filename)
		ctx.JSON(http.StatusOK, gin.H{"Status": "Success"})
	})

}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	var db_name string = "my-data"

	Client = conn.ConnectionDB()
	DB = Client.DB(context.TODO(), db_name)

	router := gin.Default()
	RgisterRouter(router)
	router.Run("localhost:8080")

	DB.Client().Close(context.TODO())
}
