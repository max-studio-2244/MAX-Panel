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
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func main() {
	// Initialize database
	initDB()
	defer db.Close()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "MAX Panel v1.0",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Static files
	app.Use("/", static.New("./web"))

	// API routes
	api := app.Group("/api")
	
	// Authentication routes
	api.Post("/auth/login", login)
	api.Post("/auth/register", register)
	api.Post("/auth/logout", logout)
	api.Get("/auth/me", getMe)
	api.Post("/auth/forgot-password", forgotPassword)
	api.Post("/auth/reset-password", resetPassword)
	api.Post("/auth/verify-2fa", verifyTwoFA)
	api.Post("/auth/enable-2fa", enableTwoFA)
	api.Post("/auth/disable-2fa", disableTwoFA)
	
	// Server routes
	api.Get("/servers", getServers)
	api.Post("/servers", createServer)
	api.Get("/servers/:id", getServer)
	api.Post("/servers/:id/start", startServer)
	api.Post("/servers/:id/stop", stopServer)
	api.Post("/servers/:id/restart", restartServer)
	api.Post("/servers/:id/kill", killServer)
	api.Get("/servers/:id/stats", getServerStats)
	api.Delete("/servers/:id", deleteServer)
	
	// File management
	api.Get("/servers/:id/files", getFiles)
	api.Get("/servers/:id/files/download", downloadFile)
	api.Post("/servers/:id/files/upload", uploadFile)
	api.Put("/servers/:id/files/edit", editFile)
	api.Delete("/servers/:id/files/delete", deleteFile)
	
	// Backup management
	api.Get("/servers/:id/backups", getBackups)
	api.Post("/servers/:id/backups", createBackup)
	api.Post("/servers/:id/backups/:backup_id/restore", restoreBackup)
	api.Delete("/servers/:id/backups/:backup_id", deleteBackup)
	
	// Admin routes
	admin := api.Group("/admin")
	admin.Get("/users", getUsers)
	admin.Post("/users", createUser)
	admin.Put("/users/:id", updateUser)
	admin.Delete("/users/:id", deleteUser)
	
	admin.Get("/eggs", getEggs)
	admin.Post("/eggs", createEgg)
	admin.Put("/eggs/:id", updateEgg)
	admin.Delete("/eggs/:id", deleteEgg)
	
	admin.Get("/nodes", getNodes)
	admin.Post("/nodes", createNode)
	admin.Put("/nodes/:id", updateNode)
	admin.Delete("/nodes/:id", deleteNode)
	admin.Get("/nodes/:id/token", getNodeToken)
	
	admin.Get("/settings", getSettings)
	admin.Put("/settings", updateSettings)
	
	admin.Get("/logs", getActivityLogs)
	
	// Console WebSocket
	app.Get("/ws/:id", handleConsoleUpgrade)

	// Start server
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

	// Create tables
	createTables := `
	CREATE TABLE IF NOT EXISTS servers (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		game TEXT NOT NULL,
		egg_id TEXT NOT NULL,
		node_id TEXT NOT NULL,
		owner_id TEXT NOT NULL,
		port INTEGER NOT NULL,
		memory INTEGER NOT NULL,
		cpu REAL NOT NULL,
		disk INTEGER NOT NULL,
		status TEXT DEFAULT 'stopped',
		container_id TEXT,
		start_command TEXT,
		environment TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		role TEXT DEFAULT 'user',
		is_admin BOOLEAN DEFAULT FALSE,
		two_factor_enabled BOOLEAN DEFAULT FALSE,
		two_factor_secret TEXT,
		api_key TEXT,
		last_login DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS eggs (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		author TEXT,
		image TEXT NOT NULL,
		start_command TEXT NOT NULL,
		stop_command TEXT,
		config_files TEXT,
		environment_variables TEXT,
		ports TEXT,
		category TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS nodes (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		host TEXT NOT NULL,
		port INTEGER DEFAULT 8080,
		token TEXT NOT NULL,
		memory_total INTEGER NOT NULL,
		memory_allocated INTEGER DEFAULT 0,
		disk_total INTEGER NOT NULL,
		disk_allocated INTEGER DEFAULT 0,
		status TEXT DEFAULT 'offline',
		last_heartbeat DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS panel_settings (
		id INTEGER PRIMARY KEY,
		panel_name TEXT DEFAULT 'MAX Panel',
		panel_logo TEXT,
		primary_color TEXT DEFAULT '#3B82F6',
		secondary_color TEXT DEFAULT '#1F2937',
		accent_color TEXT DEFAULT '#10B981',
		theme TEXT DEFAULT 'dark',
		animations_enabled BOOLEAN DEFAULT TRUE,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS activity_logs (
		id TEXT PRIMARY KEY,
		user_id TEXT,
		action TEXT NOT NULL,
		description TEXT,
		ip_address TEXT,
		user_agent TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS backups (
		id TEXT PRIMARY KEY,
		server_id TEXT NOT NULL,
		name TEXT NOT NULL,
		size INTEGER,
		path TEXT,
		status TEXT DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(createTables); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	log.Println("✅ Database initialized")
}