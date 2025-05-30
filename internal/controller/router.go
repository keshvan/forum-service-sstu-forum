package controller

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/keshvan/forum-service-sstu-forum/docs"
	"github.com/keshvan/forum-service-sstu-forum/internal/chat"
	"github.com/keshvan/forum-service-sstu-forum/internal/client"
	"github.com/keshvan/forum-service-sstu-forum/internal/controller/middleware"
	"github.com/keshvan/forum-service-sstu-forum/internal/usecase"
	"github.com/keshvan/go-common-forum/jwt"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetRoutes(engine *gin.Engine, categoryUsecase usecase.CategoryUsecase, topicUsecase usecase.TopicUsecase, postUsecase usecase.PostUsecase, jwt *jwt.JWT, log *zerolog.Logger, hub *chat.Hub, chatUsecase usecase.ChatUsecase, userClient client.UserClient) {
	categoryHandler := &CategoryHandler{categoryUsecase, log}
	topicHandler := &TopicHandler{topicUsecase, log}
	postHandler := &PostHandler{postUsecase, log}
	auth := middleware.NewAuthMiddleware(jwt)
	chatHandler := NewChatHandler(hub, chatUsecase, userClient, log)

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	engine.GET("/ws", auth.ChatAuth(), chatHandler.ServeWs)

	categories := engine.Group("/categories")
	{
		categories.GET("", categoryHandler.GetAll)
		categories.GET("/:id", categoryHandler.GetByID)

		adminCategories := categories.Group("")
		adminCategories.Use(auth.Auth(), middleware.RequireAdmin())
		{
			adminCategories.POST("", categoryHandler.Create)
			adminCategories.DELETE("/:id", categoryHandler.Delete)
			adminCategories.PATCH("/:id", categoryHandler.Update)
		}
	}

	engine.GET("/categories/:id/topics", topicHandler.GetByCategory)
	engine.POST("/categories/:id/topics", auth.Auth(), topicHandler.Create)

	engine.GET("/topics/:id", topicHandler.GetByID)
	topics := engine.Group("/topics").Use(auth.Auth())
	{
		topics.DELETE("/:id", topicHandler.Delete)
		topics.PATCH("/:id", topicHandler.Update)
	}

	engine.GET("/topics/:id/posts", postHandler.GetByTopic)
	engine.POST("/topics/:id/posts", auth.Auth(), postHandler.Create)

	posts := engine.Group("/posts").Use(auth.Auth())
	{
		posts.DELETE("/:id", postHandler.Delete)
		posts.PATCH("/:id", postHandler.Update)
	}

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
