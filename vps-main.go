package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

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

type GameServer struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Game     string    `json:"game"`
	Port     int       `json:"port"`
	Memory   int       `json:"memory"`
	Status   string    `json:"status"`
	Process  *exec.Cmd `json:"-"`
	LogFile  *os.File  `json:"-"`
	WorkDir  string    `json:"work_dir"`
	Created  time.Time `json:"created"`
}

func main() {
	initDB()
	defer db.Close()

	// Create directories
	os.MkdirAll("./servers", 0755)
	os.MkdirAll("./backups", 0755)
	os.MkdirAll("./logs", 0755)

	app := fiber.New(fiber.Config{
		AppName: "MAX Panel VPS v1.0",
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
	
	// Cloudflare
	admin.Post("/cloudflare/setup", setupCloudflare)
	admin.Get("/cloudflare/config", getCloudflareConfig)
	admin.Post("/cloudflare/record", updateCloudflareRecord)
	admin.Delete("/cloudflare/remove", removeCloudflare)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 MAX Panel VPS starting on port %s", port)
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
		port INTEGER NOT NULL,
		memory INTEGER NOT NULL,
		status TEXT DEFAULT 'stopped',
		work_dir TEXT,
		egg_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		is_admin BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS backups (
		id TEXT PRIMARY KEY,
		server_id TEXT NOT NULL,
		name TEXT NOT NULL,
		path TEXT,
		size INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS panel_settings (
		id INTEGER PRIMARY KEY DEFAULT 1,
		panel_name TEXT DEFAULT 'MAX Panel',
		domain TEXT,
		node_token TEXT,
		cloudflare_config TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(createTables); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Create default admin user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	db.Exec(`INSERT OR IGNORE INTO users (id, username, email, password, is_admin) 
			  VALUES ('admin', 'admin', 'admin@maxpanel.com', ?, true)`, string(hashedPassword))

	// Initialize egg tables
	initEggTables()
	
	log.Println("✅ Database initialized")
}

// Auth handlers
func login(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	var user struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		IsAdmin  bool   `json:"is_admin"`
		Password string `json:"-"`
	}

	err := db.QueryRow(`SELECT id, username, email, password, is_admin FROM users WHERE username = ?`, 
		req.Username).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsAdmin)

	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id": user.ID,
			"username": user.Username,
			"email": user.Email,
			"is_admin": user.IsAdmin,
		},
		"token": generateToken(),
	})
}

func register(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	userID := uuid.New().String()
	_, err = db.Exec(`INSERT INTO users (id, username, email, password) VALUES (?, ?, ?, ?)`,
		userID, req.Username, req.Email, string(hashedPassword))

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Username or email already exists"})
	}

	return c.JSON(fiber.Map{"message": "User created successfully"})
}

func getMe(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"id": "admin",
		"username": "admin",
		"email": "admin@maxpanel.com",
		"is_admin": true,
	})
}

// Server handlers
func getServers(c *fiber.Ctx) error {
	rows, err := db.Query(`SELECT id, name, game, port, memory, status, work_dir, created_at FROM servers`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var serverList []GameServer
	for rows.Next() {
		var s GameServer
		var workDir sql.NullString
		err := rows.Scan(&s.ID, &s.Name, &s.Game, &s.Port, &s.Memory, &s.Status, &workDir, &s.Created)
		if err != nil {
			continue
		}
		s.WorkDir = workDir.String
		serverList = append(serverList, s)
	}

	return c.JSON(serverList)
}

func createServer(c *fiber.Ctx) error {
	var req struct {
		Name   string `json:"name"`
		Game   string `json:"game"`
		Port   int    `json:"port"`
		Memory int    `json:"memory"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	serverID := uuid.New().String()
	workDir := filepath.Join("./servers", serverID)
	
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create server directory"})
	}

	// Download server files based on game type
	if err := setupGameServer(req.Game, workDir, req.Port); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to setup game server: " + err.Error()})
	}

	_, err := db.Exec(`INSERT INTO servers (id, name, game, port, memory, work_dir) VALUES (?, ?, ?, ?, ?, ?)`,
		serverID, req.Name, req.Game, req.Port, req.Memory, workDir)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"id": serverID, "message": "Server created successfully"})
}

func setupGameServer(game, workDir string, port int) error {
	switch game {
	case "minecraft":
		return setupMinecraftServer(workDir, port)
	case "nodejs":
		return setupNodeJSServer(workDir, port)
	default:
		return fmt.Errorf("unsupported game type: %s", game)
	}
}

func setupMinecraftServer(workDir string, port int) error {
	// Download Minecraft server jar
	jarURL := "https://piston-data.mojang.com/v1/objects/84194a2f286ef7c14ed7ce0090dba59902951553/server.jar"
	jarPath := filepath.Join(workDir, "server.jar")
	
	if err := downloadFile(jarURL, jarPath); err != nil {
		return err
	}

	// Create server.properties
	properties := fmt.Sprintf(`server-port=%d
motd=MAX Panel Minecraft Server
online-mode=false
difficulty=easy
gamemode=survival
max-players=20
spawn-protection=0
`, port)

	if err := os.WriteFile(filepath.Join(workDir, "server.properties"), []byte(properties), 0644); err != nil {
		return err
	}

	// Create eula.txt
	eula := "eula=true\n"
	return os.WriteFile(filepath.Join(workDir, "eula.txt"), []byte(eula), 0644)
}

func setupNodeJSServer(workDir string, port int) error {
	// Create simple Node.js server
	serverJS := fmt.Sprintf(`const http = require('http');
const server = http.createServer((req, res) => {
  res.writeHead(200, {'Content-Type': 'text/html'});
  res.end('<h1>Game Server Running on Port %d</h1><p>Server managed by MAX Panel</p>');
});
server.listen(%d, () => {
  console.log('Server running on port %d');
});`, port, port, port)

	return os.WriteFile(filepath.Join(workDir, "server.js"), []byte(serverJS), 0644)
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func startServer(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var server GameServer
	var workDir sql.NullString
	err := db.QueryRow(`SELECT id, name, game, port, memory, work_dir FROM servers WHERE id = ?`, id).
		Scan(&server.ID, &server.Name, &server.Game, &server.Port, &server.Memory, &workDir)
	
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}
	
	server.WorkDir = workDir.String

	serversMutex.Lock()
	defer serversMutex.Unlock()

	if existingServer, exists := servers[id]; exists && existingServer.Process != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Server already running"})
	}

	var cmd *exec.Cmd
	switch server.Game {
	case "minecraft":
		cmd = exec.Command("java", fmt.Sprintf("-Xmx%dM", server.Memory), "-jar", "server.jar", "nogui")
	case "nodejs":
		cmd = exec.Command("node", "server.js")
	default:
		return c.Status(400).JSON(fiber.Map{"error": "Unsupported game type"})
	}

	cmd.Dir = server.WorkDir
	
	// Create log file
	logFile, err := os.Create(filepath.Join(server.WorkDir, "server.log"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create log file"})
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to start server: " + err.Error()})
	}

	server.Process = cmd
	server.LogFile = logFile
	server.Status = "running"
	servers[id] = &server

	// Update database
	db.Exec(`UPDATE servers SET status = 'running' WHERE id = ?`, id)

	// Monitor process
	go func() {
		cmd.Wait()
		serversMutex.Lock()
		defer serversMutex.Unlock()
		
		if s, exists := servers[id]; exists {
			s.Status = "stopped"
			if s.LogFile != nil {
				s.LogFile.Close()
			}
			s.Process = nil
			db.Exec(`UPDATE servers SET status = 'stopped' WHERE id = ?`, id)
		}
	}()

	return c.JSON(fiber.Map{"message": "Server started successfully"})
}

func stopServer(c *fiber.Ctx) error {
	id := c.Params("id")
	
	serversMutex.Lock()
	defer serversMutex.Unlock()

	server, exists := servers[id]
	if !exists || server.Process == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Server not running"})
	}

	if err := server.Process.Process.Kill(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to stop server"})
	}

	server.Status = "stopped"
	if server.LogFile != nil {
		server.LogFile.Close()
	}
	server.Process = nil

	db.Exec(`UPDATE servers SET status = 'stopped' WHERE id = ?`, id)

	return c.JSON(fiber.Map{"message": "Server stopped successfully"})
}

func restartServer(c *fiber.Ctx) error {
	stopServer(c)
	time.Sleep(2 * time.Second)
	return startServer(c)
}

func deleteServer(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Stop server if running
	serversMutex.Lock()
	if server, exists := servers[id]; exists && server.Process != nil {
		server.Process.Process.Kill()
		delete(servers, id)
	}
	serversMutex.Unlock()

	// Get work directory
	var workDir sql.NullString
	db.QueryRow(`SELECT work_dir FROM servers WHERE id = ?`, id).Scan(&workDir)

	// Delete from database
	_, err := db.Exec(`DELETE FROM servers WHERE id = ?`, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Delete server directory
	if workDir.String != "" {
		os.RemoveAll(workDir.String)
	}

	return c.JSON(fiber.Map{"message": "Server deleted successfully"})
}

func getServer(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var server GameServer
	var workDir sql.NullString
	err := db.QueryRow(`SELECT id, name, game, port, memory, status, work_dir, created_at FROM servers WHERE id = ?`, id).
		Scan(&server.ID, &server.Name, &server.Game, &server.Port, &server.Memory, &server.Status, &workDir, &server.Created)
	
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}
	
	server.WorkDir = workDir.String
	return c.JSON(server)
}

func getServerStats(c *fiber.Ctx) error {
	id := c.Params("id")
	
	serversMutex.RLock()
	server, exists := servers[id]
	serversMutex.RUnlock()

	if !exists || server.Process == nil {
		return c.JSON(fiber.Map{
			"cpu": 0,
			"memory": 0,
			"uptime": 0,
			"status": "stopped",
		})
	}

	return c.JSON(fiber.Map{
		"cpu": 15.5,
		"memory": server.Memory * 0.7,
		"uptime": time.Since(server.Created).Seconds(),
		"status": "running",
	})
}

func generateToken() string {
	return uuid.New().String()
}