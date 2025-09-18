package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/static"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	db           *sql.DB
	servers      = make(map[string]*GameServer)
	serversMutex sync.RWMutex
)

func main() {
	// Create directories
	os.MkdirAll("./servers", 0755)
	os.MkdirAll("./backups", 0755)
	os.MkdirAll("./logs", 0755)

	initDB()
	defer db.Close()

	app := fiber.New(fiber.Config{
		AppName: "MAX Panel CodeSandbox v1.0",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.Use("/", static.New("./web"))

	// API routes
	api := app.Group("/api")
	
	// Auth
	api.Post("/auth/login", login)
	api.Post("/auth/register", register)
	api.Get("/auth/me", getMe)
	
	// Servers
	api.Get("/servers", getServers)
	api.Post("/servers", createServer)
	api.Get("/servers/:id", getServer)
	api.Post("/servers/:id/start", startServer)
	api.Post("/servers/:id/stop", stopServer)
	api.Post("/servers/:id/restart", restartServer)
	api.Delete("/servers/:id", deleteServer)
	api.Get("/servers/:id/stats", getServerStats)
	
	// Files
	api.Get("/servers/:id/files", getFiles)
	api.Get("/servers/:id/files/download", downloadFile)
	api.Post("/servers/:id/files/upload", uploadFile)
	api.Put("/servers/:id/files/edit", editFile)
	api.Delete("/servers/:id/files/delete", deleteFile)
	
	// Backups
	api.Get("/servers/:id/backups", getBackups)
	api.Post("/servers/:id/backups", createBackup)
	api.Delete("/servers/:id/backups/:backup_id", deleteBackup)
	
	// Console WebSocket
	app.Get("/ws/:id", websocket.New(handleConsole))
	
	// Admin
	admin := api.Group("/admin")
	admin.Get("/settings", getSettings)
	admin.Put("/settings", updateSettings)
	admin.Get("/nodes", getNodes)
	admin.Post("/nodes", createNode)
	admin.Get("/users", getAdminUsers)
	
	// Egg management
	admin.Get("/eggs", getEggs)
	admin.Post("/eggs", createEgg)
	admin.Put("/eggs/:id", updateEgg)
	admin.Delete("/eggs/:id", deleteEgg)
	
	// Server management
	admin.Post("/servers/create-from-egg", createServerFromEgg)
	admin.Post("/servers/assign", assignServerToUser)
	admin.Get("/servers/:id/assignments", getServerAssignments)
	admin.Delete("/assignments/:assignment_id", removeServerAssignment)
	
	// Cloudflare (mock for CodeSandbox)
	admin.Post("/cloudflare/setup", mockCloudflareSetup)
	admin.Get("/cloudflare/config", mockCloudflareConfig)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 MAX Panel CodeSandbox starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

// Mock Cloudflare functions for CodeSandbox
func mockCloudflareSetup(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Cloudflare setup completed (CodeSandbox Demo)",
		"domain":  "demo.codesandbox.io",
		"ssl":     "Generated (Mock)",
	})
}

func mockCloudflareConfig(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"configured": true,
		"domain":     "demo.codesandbox.io",
	})
}