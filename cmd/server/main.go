package main

import (
	"flag"
	"log"

	"xboard/internal/config"
	"xboard/internal/handler"
	"xboard/internal/model"
	"xboard/internal/repository"
	"xboard/internal/service"
	"xboard/pkg/cache"
	"xboard/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(
		&model.User{},
		&model.Plan{},
		&model.Server{},
		&model.Order{},
		&model.Setting{},
		&model.Stat{},
		&model.Ticket{},
		&model.TicketMessage{},
		&model.Payment{},
		&model.Coupon{},
		&model.InviteCode{},
		&model.CommissionLog{},
		&model.Notice{},
		&model.Knowledge{},
		&model.Host{},
		&model.ServerNode{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize cache
	cacheClient, err := cache.New(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect redis: %v", err)
	}

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize services
	services := service.NewServices(repos, cacheClient, cfg)

	// Start scheduler
	go services.Scheduler.Start()
	log.Println("Scheduler started")

	// Start node sync service
	go services.NodeSync.StartSyncLoop()
	log.Println("Node sync service started")

	// Initialize HTTP server
	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	handler.RegisterRoutes(r, services, cfg)

	log.Printf("Server starting on %s", cfg.App.Listen)
	if err := r.Run(cfg.App.Listen); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
