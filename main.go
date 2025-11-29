// main.go
package main

import (
	"github.com/chrisprojs/Franchiso/config"
	"github.com/chrisprojs/Franchiso/middleware"
	"github.com/chrisprojs/Franchiso/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Server struct {
	app *config.App
	r   *gin.Engine
}

func NewServer() *Server {
	db := config.NewPostgres()
	es := config.NewElastic()
	redis := config.NewRedis()
	midtrans := config.NewMidtrans()
	google_maps := config.NewGoogleMaps()
	email := config.NewEmailConfig()
	app := &config.App{DB: db, ES: es, Redis: redis, Midtrans: midtrans, GoogleMaps: google_maps, Email: email}
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))
	return &Server{app: app, r: r}
}

func (s *Server) setupRoutes() {
	// Healthcheck endpoint
	s.r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes group
	s.r.POST("/register", func(c *gin.Context) {
		service.Register(c, s.app)
	})
	s.r.POST("/verify-email", func(c *gin.Context) {
		service.VerifyEmail(c, s.app)
	})
	s.r.POST("/login", func(c *gin.Context) {
		service.Login(c, s.app)
	})
	s.r.GET("/profile", middleware.AuthMiddleware(s.app, func(c *gin.Context) {
		service.GetProfile(c, s.app)
	}))

	// Franchise routes group
	franchise := s.r.Group("/franchise")
	{
		franchise.GET("/my_franchises", middleware.AuthMiddleware(s.app, func(c *gin.Context) {
			service.DisplayMyFranchises(c, s.app)
		}))
		franchise.POST("/upload", middleware.AuthMiddleware(s.app, func(c *gin.Context) {
			service.UploadFranchise(c, s.app)
		}))
		franchise.PUT("/edit/:id", middleware.AuthMiddleware(s.app, func(c *gin.Context) {
			service.EditFranchise(c, s.app)
		}))
		franchise.GET("/:id", func(c *gin.Context) {
			showPrivate := c.DefaultQuery("showPrivate", "false")
			if showPrivate == "true" {
				middleware.AuthMiddleware(s.app, func(c *gin.Context) {
					service.DisplayFranchiseDetailByID(c, s.app)
				})(c)
				return
			}
			service.DisplayFranchiseDetailByID(c, s.app)
		})
		franchise.DELETE("delete/:id", middleware.AuthMiddleware(s.app, func(c *gin.Context) {
			service.DeleteFranchise(c, s.app)
		}))
		franchise.GET("", func(c *gin.Context) {
			service.SearchingFranchise(c, s.app)
		})
		franchise.GET("/categories", func(c *gin.Context) {
			service.CategoryList(c, s.app)
		})
		franchise.GET("/locations", func(c *gin.Context) {
			service.GetFranchiseLocations(c, s.app)
		})
	}

	// Boost routes group
	boost := s.r.Group("/boost")
	{
		boost.POST("/:id", middleware.AuthMiddleware(s.app, func(c *gin.Context) {
			service.BoostFranchise(c, s.app)
		}))
	}

	// Mid trans callback routes group
	s.r.POST("/mid_trans/call_back", func(c *gin.Context) {
		service.BoostPurchaseCallback(c, s.app)
	})

	// Admin routes group
	admin := s.r.Group("/admin")
	{
		admin.GET("/verify-franchise", middleware.AuthMiddleware(s.app, func(c *gin.Context) {
			service.DisplayAllRequestForVerificationFranchise(c, s.app)
		}))
		admin.PUT("/verify-franchise/:id", middleware.AuthMiddleware(s.app, func(c *gin.Context) {
			service.VerifyFranchise(c, s.app)
		}))
	}
}

func (s *Server) Run() {
	s.setupRoutes()
	s.r.Run()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	go RunStorageProxy()

	srv := NewServer()
	srv.Run()
}
