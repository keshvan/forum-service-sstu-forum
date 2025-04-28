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

	//Categories
	category := engine.Group("/categories")
	{
		category.GET("/", categoryHandler.GetAll)

		adminCategory := category.Group("")
		adminCategory.Use(auth.Auth(), middleware.RequireAdmin())
		{
			adminCategory.POST("/", categoryHandler.Create)
			adminCategory.DELETE("/:id", categoryHandler.Delete)
			adminCategory.PATCH("/:id", categoryHandler.Update)
		}

		//Topics
		topic := category.Group("/:category_id/topics")
		{
			topic.GET("/", topicHandler.GetByCategory)

			needAuthTopic := topic.Group("")
			needAuthTopic.Use(auth.Auth())
			{
				needAuthTopic.POST("/", topicHandler.Create)
				needAuthTopic.DELETE("/:id", topicHandler.Delete)
				needAuthTopic.PATCH("/:id", topicHandler.Update)
			}

			//Posts
			post := topic.Group("/:topic_id/posts")
			{
				post.GET("/", postHandler.GetByTopic)

				needAuthPost := post.Group("")
				needAuthPost.Use(auth.Auth())
				{
					needAuthPost.POST("/", postHandler.Create)
					needAuthPost.DELETE("/:id", postHandler.Delete)
					needAuthPost.PATCH("/:id", postHandler.Update)
				}
			}
		}
	}
}
