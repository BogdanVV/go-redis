package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Handlers struct {
	Todos TodosHandlersInterface
}

func NewHandlers(redis *redis.Client, ctx context.Context) *Handlers {
	return &Handlers{
		Todos: NewTodosHandlerStruct(redis, ctx),
	}
}

type TodosHandlersInterface interface {
	GetAllTodos(c *gin.Context)
	GetTodoById(c *gin.Context)
}
