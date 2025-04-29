package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/keshvan/forum-service-sstu-forum/internal/controller/middleware"
	"github.com/keshvan/forum-service-sstu-forum/internal/usecase"
	"github.com/keshvan/go-common-forum/jwt"
)

func SetRoutes(engine *gin.Engine, categoryUsecase usecase.CategoryUsecase, topicUsecase usecase.TopicUsecase, postUsecase usecase.PostUsecase, jwt *jwt.JWT) {
	categoryHandler := &CategoryHandler{categoryUsecase}
	topicHandler := &TopicHandler{topicUsecase}
	postHandler := &PostHandler{postUsecase}
	auth := middleware.NewAuthMiddleware(jwt)

	categories := engine.Group("/categories")
	{
		categories.GET("", categoryHandler.GetAll)

		adminCategories := categories.Group("")
		adminCategories.Use(auth.Auth(), middleware.RequireAdmin())
		{
			adminCategories.POST("", categoryHandler.Create)
			adminCategories.DELETE("/:id", categoryHandler.Delete)
			adminCategories.PATCH("/:id", categoryHandler.Update)
		}

		categoriesID := categories.Group("/:id")
		{
			categoriesID.GET("/topics", topicHandler.GetByCategory)
			categoriesID.POST("/topics", auth.Auth(), topicHandler.Create)
			categoriesID.DELETE("/topics/:topic_id", auth.Auth(), topicHandler.Delete)
			categoriesID.PATCH("/topics/:topic_id", auth.Auth(), topicHandler.Update)
		}

		topicsID := categoriesID.Group("/topics/:topic_id")
		{
			topicsID.GET("/posts", postHandler.GetByTopic)
			topicsID.POST("/posts", auth.Auth(), postHandler.Create)
			topicsID.DELETE("/posts/:post_id", auth.Auth(), postHandler.Delete)
			topicsID.PATCH("/posts/:post_id", auth.Auth(), postHandler.Update)
		}
	}

	//Topics
	//engine.GET("/categories/:category_id/topics", topicHandler.GetByCategory)
	//engine.POST("/categories/:category_id/topics", auth.Auth(), topicHandler.Create)
	/*
		topics := engine.Group("/topics")
		topics.Use(auth.Auth())
		{
			topics.DELETE("/:id", topicHandler.Delete)
			topics.PATCH("/:id", topicHandler.Update)

			topics.GET("/:id/posts", postHandler.GetByTopic)
			topics.POST("/:id/posts", auth.Auth(), postHandler.Create)
		}

		//Posts
		//engine.GET("/topics/:topic_id/posts", postHandler.GetByTopic)
		//engine.POST("/topics/:topic_id/posts", auth.Auth(), postHandler.Create)

		posts := engine.Group("/posts")
		posts.Use(auth.Auth())
		{
			posts.DELETE("/:id", postHandler.Delete)
			posts.PATCH("/:id", postHandler.Update)
		}*/
}
