package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/static"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	initDB()
	defer db.Close()

	app := fiber.New(fiber.Config{
		AppName: "MAX Panel v1.0",
	})

	app.Use(logger.New())
	app.Use(cors.New())
	app.Use("/", static.New("./web"))

	// API routes
	api := app.Group("/api")
	
	// Auth routes (mock for CodeSandbox)
	api.Post("/auth/login", mockLogin)
	api.Get("/auth/me", mockGetMe)
	
	// Server routes (mock for CodeSandbox)
	api.Get("/servers", mockGetServers)
	api.Post("/servers", mockCreateServer)
	api.Get("/servers/:id", mockGetServer)
	api.Post("/servers/:id/start", mockStartServer)
	api.Post("/servers/:id/stop", mockStopServer)
	
	// Admin routes (mock)
	admin := api.Group("/admin")
	admin.Get("/users", mockGetUsers)
	admin.Get("/eggs", mockGetEggs)
	admin.Get("/settings", mockGetSettings)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 MAX Panel starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./panel.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS servers (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		game TEXT NOT NULL,
		status TEXT DEFAULT 'stopped',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		is_admin BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(createTables); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	log.Println("✅ Database initialized")
}

// Mock handlers for CodeSandbox
func mockLogin(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id": "demo-user",
			"username": "demo",
			"email": "demo@maxpanel.com",
			"is_admin": true,
		},
		"token": "demo-token",
	})
}

func mockGetMe(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"id": "demo-user",
		"username": "demo",
		"email": "demo@maxpanel.com",
		"is_admin": true,
	})
}

func mockGetServers(c *fiber.Ctx) error {
	return c.JSON([]fiber.Map{
		{
			"id": "server-1",
			"name": "Minecraft Server",
			"game": "minecraft",
			"status": "running",
			"port": 25565,
			"memory": 2048,
		},
		{
			"id": "server-2", 
			"name": "CS:GO Server",
			"game": "csgo",
			"status": "stopped",
			"port": 27015,
			"memory": 1024,
		},
	})
}

func mockCreateServer(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"id": "new-server",
		"message": "Server created (demo mode)",
	})
}

func mockGetServer(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{
		"id": id,
		"name": "Demo Server",
		"game": "minecraft",
		"status": "running",
		"port": 25565,
		"memory": 2048,
	})
}

func mockStartServer(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Server started (demo mode)"})
}

func mockStopServer(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Server stopped (demo mode)"})
}

func mockGetUsers(c *fiber.Ctx) error {
	return c.JSON([]fiber.Map{
		{
			"id": "user-1",
			"username": "admin",
			"email": "admin@maxpanel.com",
			"is_admin": true,
		},
	})
}

func mockGetEggs(c *fiber.Ctx) error {
	return c.JSON([]fiber.Map{
		{
			"id": "minecraft-vanilla",
			"name": "Minecraft Vanilla",
			"category": "Minecraft",
			"image": "itzg/minecraft-server:latest",
		},
		{
			"id": "csgo",
			"name": "Counter-Strike: Global Offensive", 
			"category": "Source Engine",
			"image": "steamcmd/steamcmd:latest",
		},
	})
}

func mockGetSettings(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"panel_name": "MAX Panel",
		"primary_color": "#3B82F6",
		"theme": "dark",
	})
}