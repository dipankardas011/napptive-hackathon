package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// make them accessable from redis database
type Album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var (
	redis_host     = os.Getenv("REDIS_HOST")
	redis_port     = os.Getenv("REDIS_PORT")
	redis_password = os.Getenv("REDIS_PASSWORD")

	ctx = context.Background()
	rdb *redis.Client
)

const (
	VERSION = "3.0"
	AUTHOR  = "Dipankar Das"
)

func deleteAlbum(id string) (err error) {
	_, err = rdb.Del(ctx, id).Result()
	if err != nil {
		return
	}
	return
}

func deleteAlbums() (err error) {
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		err := deleteAlbum(key)
		if err != nil {
			return err
		}
	}
	return nil
}

func getAlbum(id string) (album Album, err error) {
	value, err := rdb.Get(ctx, id).Result()

	if err != nil {
		return
	}

	if err != redis.Nil {
		err = json.Unmarshal([]byte(value), &album)
	}
	return
}

func getAlbums() ([]Album, error) {
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		return nil, err
	}
	var albums []Album

	for _, key := range keys {
		album, err := getAlbum(key)
		if err != nil {
			return nil, err
		}
		albums = append(albums, album)
	}
	return albums, nil
}

func savevideo(album Album) {

	albumbytes, err := json.Marshal(album)
	if err != nil {
		panic(err)
	}

	err = rdb.Set(ctx, album.ID, albumbytes, 0).Err()
	if err != nil {
		panic(err)
	}

}

func savevideos(albums []Album) {
	for _, album := range albums {
		savevideo(album)
	}
}

func GetAlbums(c *gin.Context) {
	albums, err := getAlbums()
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
	}
	c.IndentedJSON(http.StatusOK, albums)
}

func GetHome(c *gin.Context) {
	c.HTML(http.StatusOK, "posts/index.tmpl", gin.H{
		"title":       "Go Gin server backed by redis",
		"version":     VERSION,
		"author":      AUTHOR,
		"description": "Go Gin server to handle restfull requests with Redis as database",
	})
}

func PostAlbums(c *gin.Context) {
	var newAlbum Album
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}
	albums, err := getAlbums()
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
	}
	albums = append(albums, newAlbum)
	savevideos(albums)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

func GetAlbumByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}

	album, err := getAlbum(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found" + err.Error()})
	}
	c.IndentedJSON(http.StatusOK, album)
}

func DelAlbumByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	err := deleteAlbum(id)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Deleted " + id})
}

func DelAlbums(c *gin.Context) {
	err := deleteAlbums()
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Deleted all entries"})
}

func GetVersion(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"Version": VERSION, "Author": AUTHOR})
}

func GetHealth(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"Status": "Healthy"})
}

func main() {
	r := redis.NewClient(&redis.Options{
		Addr:     redis_host + ":" + redis_port,
		Password: redis_password, // no password set
		DB:       0,              // use default DB
	})

	rdb = r

	router := gin.Default()
	router.LoadHTMLGlob("templates/**/*")
	router.GET("/", GetHome)
	router.GET("/albums", GetAlbums)
	router.POST("/albums", PostAlbums)
	router.DELETE("/albums/:id", DelAlbumByID)
	router.DELETE("/albums", DelAlbums)
	router.GET("/albums/:id", GetAlbumByID)

	router.GET("/version", GetVersion)
	router.GET("/healthz", GetHealth)

	router.Run("0.0.0.0:8080")
}
