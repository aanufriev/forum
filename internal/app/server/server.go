package server

import (
	"log"

	"github.com/aanufriev/forum/configs"
	forumDelivery "github.com/aanufriev/forum/internal/pkg/forum/delivery"
	forumRepository "github.com/aanufriev/forum/internal/pkg/forum/repository"
	forumUsecase "github.com/aanufriev/forum/internal/pkg/forum/usecase"
	userDelivery "github.com/aanufriev/forum/internal/pkg/user/delivery"
	userRepository "github.com/aanufriev/forum/internal/pkg/user/repository"
	userUsecase "github.com/aanufriev/forum/internal/pkg/user/usecase"
	"github.com/buaazp/fasthttprouter"
	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
)

func StartApiServer() {
	db, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			User:     "docker",
			Password: "docker",
			Port:     5432,
			Database: "forum",
		},
		MaxConnections: 100,
	})

	if err != nil {
		log.Fatal(err)
	}

	userRepository := userRepository.New(db)
	userUsecase := userUsecase.New(userRepository)
	userDelivery := userDelivery.New(userUsecase)

	forumRepository := forumRepository.New(db)
	forumUsecase := forumUsecase.New(forumRepository)
	forumDelivery := forumDelivery.New(forumUsecase, userRepository)

	router := fasthttprouter.New()

	router.POST("/api/user/:nickname/create", userDelivery.Create)
	router.GET("/api/user/:nickname/profile", userDelivery.Get)
	router.POST("/api/user/:nickname/profile", userDelivery.Update)

	router.POST("/api/forum/:slug", forumDelivery.Create)
	router.GET("/api/forum/:slug/details", forumDelivery.Get)
	router.POST("/api/forum/:slug/create", forumDelivery.CreateThread)
	router.GET("/api/forum/:slug/threads", forumDelivery.GetThreads)
	router.GET("/api/forum/:slug/users", forumDelivery.GetUsersFromForum)

	router.POST("/api/thread/:slug_or_id/create", forumDelivery.CreatePosts)
	router.GET("/api/thread/:slug_or_id/details", forumDelivery.GetThread)
	router.POST("/api/thread/:slug_or_id/vote", forumDelivery.Vote)
	router.GET("/api/thread/:slug_or_id/posts", forumDelivery.GetPosts)
	router.POST("/api/thread/:slug_or_id/details", forumDelivery.UpdateThread)

	router.GET("/api/post/:id/details", forumDelivery.GetPostDetails)
	router.POST("/api/post/:id/details", forumDelivery.UpdatePost)

	router.POST("/api/service/clear", forumDelivery.ClearService)
	router.GET("/api/service/status", forumDelivery.GetServiceInfo)

	log.Printf("server started at port %v", configs.ApiPort)
	log.Fatal(fasthttp.ListenAndServe(":5000", router.Handler))
}
