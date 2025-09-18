package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Egg struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Game         string            `json:"game"`
	Version      string            `json:"version"`
	BuildNumber  string            `json:"build_number"`
	Image        string            `json:"image"`
	StartCommand string            `json:"start_command"`
	StopCommand  string            `json:"stop_command"`
	InstallScript string           `json:"install_script"`
	ConfigFiles  string            `json:"config_files"`
	Environment  string            `json:"environment"`
	Ports        string            `json:"ports"`
	Category     string            `json:"category"`
	MinMemory    int               `json:"min_memory"`
	MaxMemory    int               `json:"max_memory"`
	CreatedAt    string            `json:"created_at"`
}

type ServerAssignment struct {
	ID       string `json:"id"`
	ServerID string `json:"server_id"`
	UserID   string `json:"user_id"`
	Role     string `json:"role"` // owner, admin, user
}

func initEggTables() {
	createEggTables := `
	CREATE TABLE IF NOT EXISTS eggs (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		game TEXT NOT NULL,
		version TEXT,
		build_number TEXT,
		image TEXT,
		start_command TEXT NOT NULL,
		stop_command TEXT,
		install_script TEXT,
		config_files TEXT,
		environment TEXT,
		ports TEXT,
		category TEXT,
		min_memory INTEGER DEFAULT 512,
		max_memory INTEGER DEFAULT 8192,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS server_assignments (
		id TEXT PRIMARY KEY,
		server_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		role TEXT DEFAULT 'user',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	INSERT OR IGNORE INTO eggs (id, name, description, game, version, start_command, category, min_memory, max_memory) VALUES
	('minecraft-vanilla', 'Minecraft Vanilla', 'Official Minecraft Server', 'minecraft', 'latest', 'java -Xms{MEMORY}M -Xmx{MEMORY}M -jar server.jar nogui', 'Minecraft', 1024, 8192),
	('minecraft-paper', 'Minecraft Paper', 'High performance Minecraft server', 'minecraft', 'latest', 'java -Xms{MEMORY}M -Xmx{MEMORY}M -jar paper.jar nogui', 'Minecraft', 1024, 8192),
	('nodejs-app', 'Node.js Application', 'Node.js web/game server', 'nodejs', 'latest', 'node server.js', 'Web', 256, 2048);`

	if _, err := db.Exec(createEggTables); err != nil {
		log.Fatal("Failed to create egg tables:", err)
	}
}

func getEggs(c *fiber.Ctx) error {
	rows, err := db.Query(`
		SELECT id, name, description, game, version, build_number, image, start_command, 
		       stop_command, install_script, config_files, environment, ports, category, 
		       min_memory, max_memory, created_at 
		FROM eggs ORDER BY category, name`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var eggs []Egg
	for rows.Next() {
		var e Egg
		var version, buildNumber, image, stopCommand, installScript, configFiles, environment, ports sql.NullString
		
		err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.Game, &version, &buildNumber, 
			&image, &e.StartCommand, &stopCommand, &installScript, &configFiles, 
			&environment, &ports, &e.Category, &e.MinMemory, &e.MaxMemory, &e.CreatedAt)
		if err != nil {
			continue
		}

		e.Version = version.String
		e.BuildNumber = buildNumber.String
		e.Image = image.String
		e.StopCommand = stopCommand.String
		e.InstallScript = installScript.String
		e.ConfigFiles = configFiles.String
		e.Environment = environment.String
		e.Ports = ports.String

		eggs = append(eggs, e)
	}

	return c.JSON(eggs)
}

func createEgg(c *fiber.Ctx) error {
	var egg Egg
	if err := c.BodyParser(&egg); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	egg.ID = uuid.New().String()

	_, err := db.Exec(`
		INSERT INTO eggs (id, name, description, game, version, build_number, image, 
		                  start_command, stop_command, install_script, config_files, 
		                  environment, ports, category, min_memory, max_memory)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		egg.ID, egg.Name, egg.Description, egg.Game, egg.Version, egg.BuildNumber,
		egg.Image, egg.StartCommand, egg.StopCommand, egg.InstallScript,
		egg.ConfigFiles, egg.Environment, egg.Ports, egg.Category,
		egg.MinMemory, egg.MaxMemory)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"id": egg.ID, "message": "Egg created successfully"})
}

func updateEgg(c *fiber.Ctx) error {
	id := c.Params("id")
	var egg Egg
	if err := c.BodyParser(&egg); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	_, err := db.Exec(`
		UPDATE eggs SET name = ?, description = ?, game = ?, version = ?, build_number = ?,
		               image = ?, start_command = ?, stop_command = ?, install_script = ?,
		               config_files = ?, environment = ?, ports = ?, category = ?,
		               min_memory = ?, max_memory = ?
		WHERE id = ?`,
		egg.Name, egg.Description, egg.Game, egg.Version, egg.BuildNumber,
		egg.Image, egg.StartCommand, egg.StopCommand, egg.InstallScript,
		egg.ConfigFiles, egg.Environment, egg.Ports, egg.Category,
		egg.MinMemory, egg.MaxMemory, id)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Egg updated successfully"})
}

func deleteEgg(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := db.Exec("DELETE FROM eggs WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Egg deleted successfully"})
}

func createServerFromEgg(c *fiber.Ctx) error {
	var req struct {
		Name     string `json:"name"`
		EggID    string `json:"egg_id"`
		Port     int    `json:"port"`
		Memory   int    `json:"memory"`
		OwnerID  string `json:"owner_id"`
		Environment map[string]string `json:"environment"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Get egg details
	var egg Egg
	var version, buildNumber, image, installScript, configFiles, environment, ports sql.NullString
	err := db.QueryRow(`
		SELECT id, name, game, version, build_number, image, start_command, 
		       install_script, config_files, environment, ports, min_memory, max_memory
		FROM eggs WHERE id = ?`, req.EggID).Scan(
		&egg.ID, &egg.Name, &egg.Game, &version, &buildNumber, &image,
		&egg.StartCommand, &installScript, &configFiles, &environment, &ports,
		&egg.MinMemory, &egg.MaxMemory)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Egg not found"})
	}

	egg.Version = version.String
	egg.BuildNumber = buildNumber.String
	egg.Image = image.String
	egg.InstallScript = installScript.String
	egg.ConfigFiles = configFiles.String
	egg.Environment = environment.String
	egg.Ports = ports.String

	// Validate memory limits
	if req.Memory < egg.MinMemory || req.Memory > egg.MaxMemory {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Memory must be between %d and %d MB", egg.MinMemory, egg.MaxMemory),
		})
	}

	serverID := uuid.New().String()
	workDir := filepath.Join("./servers", serverID)

	if err := os.MkdirAll(workDir, 0755); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create server directory"})
	}

	// Setup server using egg
	if err := setupServerFromEgg(egg, workDir, req.Port, req.Memory, req.Environment); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to setup server: " + err.Error()})
	}

	// Create server record
	_, err = db.Exec(`
		INSERT INTO servers (id, name, game, port, memory, work_dir, egg_id) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		serverID, req.Name, egg.Game, req.Port, req.Memory, workDir, req.EggID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Assign server to user
	if req.OwnerID != "" {
		assignmentID := uuid.New().String()
		db.Exec(`
			INSERT INTO server_assignments (id, server_id, user_id, role) 
			VALUES (?, ?, ?, 'owner')`,
			assignmentID, serverID, req.OwnerID)
	}

	return c.JSON(fiber.Map{"id": serverID, "message": "Server created successfully"})
}

func setupServerFromEgg(egg Egg, workDir string, port, memory int, envVars map[string]string) error {
	// Replace variables in start command
	startCommand := strings.ReplaceAll(egg.StartCommand, "{MEMORY}", fmt.Sprintf("%d", memory))
	startCommand = strings.ReplaceAll(startCommand, "{PORT}", fmt.Sprintf("%d", port))

	// Parse environment variables
	var eggEnv map[string]string
	if egg.Environment != "" {
		json.Unmarshal([]byte(egg.Environment), &eggEnv)
	}

	// Merge with user-provided environment
	finalEnv := make(map[string]string)
	for k, v := range eggEnv {
		finalEnv[k] = v
	}
	for k, v := range envVars {
		finalEnv[k] = v
	}

	// Run install script if provided
	if egg.InstallScript != "" {
		if err := runInstallScript(egg.InstallScript, workDir, finalEnv); err != nil {
			return err
		}
	}

	// Create config files
	if egg.ConfigFiles != "" {
		var configFiles map[string]string
		if err := json.Unmarshal([]byte(egg.ConfigFiles), &configFiles); err == nil {
			for filename, content := range configFiles {
				// Replace variables in config content
				content = strings.ReplaceAll(content, "{PORT}", fmt.Sprintf("%d", port))
				content = strings.ReplaceAll(content, "{MEMORY}", fmt.Sprintf("%d", memory))
				
				for k, v := range finalEnv {
					content = strings.ReplaceAll(content, "{"+k+"}", v)
				}

				filePath := filepath.Join(workDir, filename)
				os.MkdirAll(filepath.Dir(filePath), 0755)
				os.WriteFile(filePath, []byte(content), 0644)
			}
		}
	}

	// Save start command
	os.WriteFile(filepath.Join(workDir, "start.sh"), []byte("#!/bin/bash\n"+startCommand), 0755)

	return nil
}

func runInstallScript(script, workDir string, env map[string]string) error {
	// Create install script file
	scriptPath := filepath.Join(workDir, "install.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\n"+script), 0755); err != nil {
		return err
	}

	// Execute install script (simplified - in production, use proper sandboxing)
	return nil
}

func assignServerToUser(c *fiber.Ctx) error {
	var req struct {
		ServerID string `json:"server_id"`
		UserID   string `json:"user_id"`
		Role     string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	assignmentID := uuid.New().String()
	_, err := db.Exec(`
		INSERT INTO server_assignments (id, server_id, user_id, role) 
		VALUES (?, ?, ?, ?)`,
		assignmentID, req.ServerID, req.UserID, req.Role)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Server assigned successfully"})
}

func getServerAssignments(c *fiber.Ctx) error {
	serverID := c.Params("id")

	rows, err := db.Query(`
		SELECT sa.id, sa.server_id, sa.user_id, sa.role, u.username, u.email
		FROM server_assignments sa
		JOIN users u ON sa.user_id = u.id
		WHERE sa.server_id = ?`, serverID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var assignments []map[string]interface{}
	for rows.Next() {
		var assignment map[string]interface{} = make(map[string]interface{})
		var id, serverID, userID, role, username, email string

		err := rows.Scan(&id, &serverID, &userID, &role, &username, &email)
		if err != nil {
			continue
		}

		assignment["id"] = id
		assignment["server_id"] = serverID
		assignment["user_id"] = userID
		assignment["role"] = role
		assignment["username"] = username
		assignment["email"] = email

		assignments = append(assignments, assignment)
	}

	return c.JSON(assignments)
}

func removeServerAssignment(c *fiber.Ctx) error {
	assignmentID := c.Params("assignment_id")

	_, err := db.Exec("DELETE FROM server_assignments WHERE id = ?", assignmentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Assignment removed successfully"})
}