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

var redis_host = os.Getenv("REDIS_HOST")
var redis_port = os.Getenv("REDIS_PORT")
var redis_password = os.Getenv("REDIS_PASSWORD")

var ctx = context.Background()
var rdb *redis.Client

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
	c.IndentedJSON(http.StatusOK, gin.H{"home": "Welcome to restful http server backed by redis", "purpose": "napptive hackathon"})
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

func main() {
	r := redis.NewClient(&redis.Options{
		Addr:     redis_host + ":" + redis_port,
		Password: redis_password, // no password set
		DB:       0,              // use default DB
	})
	rdb = r

	router := gin.Default()
	router.GET("/", GetHome)
	router.GET("/albums", GetAlbums)
	router.POST("/albums", PostAlbums)
	router.GET("/albums/:id", GetAlbumByID)

	router.Run("0.0.0.0:8080")
}
