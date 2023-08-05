package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Todo struct {
	UserId    int    `json:"userId"`
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

const REDIS_CACHING_TIME = time.Minute * 5

func main() {
	r := gin.Default()

	ctx := context.Background()
	redisClient := redis.NewClient(&redis.Options{})

	api := r.Group("api")
	{
		api.GET("todos", func(c *gin.Context) {
			handleGetAllTodos(c, *redisClient, ctx)
		})
		api.GET("todos/:id", func(c *gin.Context) {
			handleGetTodoById(c, *redisClient, ctx)
		})
	}

	r.Run(":9999")
}

// NOTE: seems like gin has built-in caching functionality
// because even without redis all requests to jsonplaceholder take 3-4 times less time
// after calling it for the first time

// redis reduces response time from 110-120ms to 2-8 ms (local environment + postman)

func handleGetAllTodos(c *gin.Context, r redis.Client, ctx context.Context) {
	var todos []Todo

	// return data from redis if exists
	redisTodos, err := r.Get(ctx, "todos").Result()
	if err == nil && redisTodos != "" {
		if err := json.Unmarshal([]byte(redisTodos), &todos); err == nil {
			c.JSON(http.StatusOK, gin.H{"data": todos})
			return
		} else {
			fmt.Printf("failed to extract todos from redis into struct, err>>> %s\n", err.Error())
		}
	} else {
		fmt.Println("no data in redis")
	}

	// get the latest data from external API, save it into redis and return to the user
	response, err := http.Get("https://jsonplaceholder.typicode.com/todos")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load the data from external API"})
		return
	}

	if err := json.NewDecoder(response.Body).Decode(&todos); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse external API's response"})
		return
	}

	// cache
	jsonTodos, err := json.Marshal(todos)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse json data for redis"})
	}

	err = r.Set(ctx, "todos", string(jsonTodos), REDIS_CACHING_TIME).Err()
	if err != nil {
		fmt.Printf("failed to save todos in redis")
	}

	c.JSON(http.StatusOK, gin.H{"data": todos})
}

func handleGetTodoById(c *gin.Context, r redis.Client, ctx context.Context) {
	todoId := c.Param("id")
	_, err := strconv.Atoi(todoId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "invalid id"})
		return
	}

	var todo Todo
	redisTodo, err := r.Get(ctx, fmt.Sprintf("todo:%s", todoId)).Result()
	if err == nil && redisTodo != "" {
		err = json.Unmarshal([]byte(redisTodo), &todo)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"data": todo})
			return
		}
	}

	response, err := http.Get(fmt.Sprintf("https://jsonplaceholder.typicode.com/todos/%s", todoId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch data from external API"})
		return
	}

	if err := json.NewDecoder(response.Body).Decode(&todo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse data from external API"})
		return
	}

	todoJson, err := json.Marshal(todo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse failed to parse data from external API for redis"})
		return
	}

	err = r.Set(ctx, fmt.Sprintf("todo:%s", todoId), []byte(todoJson), REDIS_CACHING_TIME).Err()
	if err != nil {
		fmt.Printf("failed to save todo with id %s in redis", todoId)
	}

	c.JSON(http.StatusOK, gin.H{"data": todo})
}
