package main

import (
	"context"

	"github.com/bogdanvv/go-redis/handlers"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	r := gin.Default()

	ctx := context.Background()
	redisClient := redis.NewClient(&redis.Options{})
	h := handlers.NewHandlers(redisClient, ctx)

	api := r.Group("api")
	{
		api.GET("todos", h.Todos.GetAllTodos)
		api.GET("todos/:id", h.Todos.GetTodoById)
	}

	r.Run(":9999")
}
