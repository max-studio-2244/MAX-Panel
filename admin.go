package main

import (
	"database/sql"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

type Egg struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	Author               string `json:"author"`
	Image                string `json:"image"`
	StartCommand         string `json:"start_command"`
	StopCommand          string `json:"stop_command"`
	ConfigFiles          string `json:"config_files"`
	EnvironmentVariables string `json:"environment_variables"`
	Ports                string `json:"ports"`
	Category             string `json:"category"`
	CreatedAt            string `json:"created_at"`
}

type Node struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Token           string `json:"token"`
	MemoryTotal     int    `json:"memory_total"`
	MemoryAllocated int    `json:"memory_allocated"`
	DiskTotal       int    `json:"disk_total"`
	DiskAllocated   int    `json:"disk_allocated"`
	Status          string `json:"status"`
	LastHeartbeat   string `json:"last_heartbeat"`
	CreatedAt       string `json:"created_at"`
}

type PanelSettings struct {
	ID                int    `json:"id"`
	PanelName         string `json:"panel_name"`
	PanelLogo         string `json:"panel_logo"`
	PrimaryColor      string `json:"primary_color"`
	SecondaryColor    string `json:"secondary_color"`
	AccentColor       string `json:"accent_color"`
	Theme             string `json:"theme"`
	AnimationsEnabled bool   `json:"animations_enabled"`
	UpdatedAt         string `json:"updated_at"`
}

// User Management
func getUsers(c *fiber.Ctx) error {
	rows, err := db.Query(`
		SELECT id, username, email, first_name, last_name, role, is_admin, created_at
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.FirstName, &u.LastName, &u.Role, &u.IsAdmin, &u.CreatedAt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		users = append(users, u)
	}

	return c.JSON(users)
}

func createUser(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	return register(c)
}

func updateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	_, err := db.Exec(`
		UPDATE users SET username = ?, email = ?, first_name = ?, last_name = ?, is_admin = ?
		WHERE id = ?
	`, req.Username, req.Email, req.FirstName, req.LastName, req.IsAdmin, id)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "User updated successfully"})
}

func deleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "User deleted successfully"})
}

// Egg Management
func getEggs(c *fiber.Ctx) error {
	// Return default eggs for now
	eggs := []Egg{
		{
			ID:          "minecraft-vanilla",
			Name:        "Minecraft Vanilla",
			Description: "Official Minecraft server",
			Author:      "Mojang",
			Image:       "itzg/minecraft-server:latest",
			Category:    "Minecraft",
			StartCommand: "java -Xms{{SERVER_MEMORY}}M -Xmx{{SERVER_MEMORY}}M -jar server.jar nogui",
		},
		{
			ID:          "minecraft-paper",
			Name:        "Minecraft Paper",
			Description: "High performance Minecraft server",
			Author:      "PaperMC",
			Image:       "itzg/minecraft-server:latest",
			Category:    "Minecraft",
			StartCommand: "java -Xms{{SERVER_MEMORY}}M -Xmx{{SERVER_MEMORY}}M -jar paper.jar nogui",
		},
		{
			ID:          "csgo",
			Name:        "Counter-Strike: Global Offensive",
			Description: "CS:GO dedicated server",
			Author:      "Valve",
			Image:       "steamcmd/steamcmd:latest",
			Category:    "Source Engine",
			StartCommand: "./srcds_run -game csgo -console -usercon +game_type 0 +game_mode 1 +mapgroup mg_active +map de_dust2",
		},
		{
			ID:          "rust",
			Name:        "Rust Server",
			Description: "Rust dedicated server",
			Author:      "Facepunch Studios",
			Image:       "didstopia/rust-server:latest",
			Category:    "Survival",
			StartCommand: "./RustDedicated -batchmode +server.port {{SERVER_PORT}} +server.identity \"rust\" +rcon.port {{RCON_PORT}} +rcon.web 1",
		},
	}

	return c.JSON(eggs)
}

func createEgg(c *fiber.Ctx) error {
	var egg Egg
	if err := c.BodyParser(&egg); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	egg.ID = generateID()

	_, err := db.Exec(`
		INSERT INTO eggs (id, name, description, author, image, start_command, stop_command, config_files, environment_variables, ports, category)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, egg.ID, egg.Name, egg.Description, egg.Author, egg.Image, egg.StartCommand, egg.StopCommand, egg.ConfigFiles, egg.EnvironmentVariables, egg.Ports, egg.Category)

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
		UPDATE eggs SET name = ?, description = ?, author = ?, image = ?, start_command = ?, stop_command = ?, config_files = ?, environment_variables = ?, ports = ?, category = ?
		WHERE id = ?
	`, egg.Name, egg.Description, egg.Author, egg.Image, egg.StartCommand, egg.StopCommand, egg.ConfigFiles, egg.EnvironmentVariables, egg.Ports, egg.Category, id)

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

// Node Management
func getNodes(c *fiber.Ctx) error {
	// Return default local node for now
	nodes := []Node{
		{
			ID:              "local",
			Name:            "Local Node",
			Description:     "Local Docker daemon",
			Host:            "localhost",
			Port:            8080,
			Status:          "online",
			MemoryTotal:     8192,
			MemoryAllocated: 2048,
			DiskTotal:       100000,
			DiskAllocated:   25000,
			Token:           generateToken(),
		},
	}

	return c.JSON(nodes)
}

func createNode(c *fiber.Ctx) error {
	var node Node
	if err := c.BodyParser(&node); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	node.ID = generateID()
	node.Token = generateToken()

	_, err := db.Exec(`
		INSERT INTO nodes (id, name, description, host, port, token, memory_total, disk_total)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, node.ID, node.Name, node.Description, node.Host, node.Port, node.Token, node.MemoryTotal, node.DiskTotal)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"id": node.ID, "token": node.Token, "message": "Node created successfully"})
}

func updateNode(c *fiber.Ctx) error {
	id := c.Params("id")
	var node Node
	if err := c.BodyParser(&node); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	_, err := db.Exec(`
		UPDATE nodes SET name = ?, description = ?, host = ?, port = ?, memory_total = ?, disk_total = ?
		WHERE id = ?
	`, node.Name, node.Description, node.Host, node.Port, node.MemoryTotal, node.DiskTotal, id)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Node updated successfully"})
}

func deleteNode(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := db.Exec("DELETE FROM nodes WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Node deleted successfully"})
}

func getNodeToken(c *fiber.Ctx) error {
	id := c.Params("id")

	var token string
	err := db.QueryRow("SELECT token FROM nodes WHERE id = ?", id).Scan(&token)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Node not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"token": token})
}

// Settings Management
func getSettings(c *fiber.Ctx) error {
	var settings PanelSettings
	err := db.QueryRow(`
		SELECT id, panel_name, panel_logo, primary_color, secondary_color, accent_color, theme, animations_enabled, updated_at
		FROM panel_settings LIMIT 1
	`).Scan(&settings.ID, &settings.PanelName, &settings.PanelLogo, &settings.PrimaryColor, &settings.SecondaryColor, &settings.AccentColor, &settings.Theme, &settings.AnimationsEnabled, &settings.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return default settings
			settings = PanelSettings{
				PanelName:         "MAX Panel",
				PrimaryColor:      "#3B82F6",
				SecondaryColor:    "#1F2937",
				AccentColor:       "#10B981",
				Theme:             "dark",
				AnimationsEnabled: true,
			}
		} else {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(settings)
}

func updateSettings(c *fiber.Ctx) error {
	var settings PanelSettings
	if err := c.BodyParser(&settings); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Insert or update settings
	_, err := db.Exec(`
		INSERT OR REPLACE INTO panel_settings (id, panel_name, panel_logo, primary_color, secondary_color, accent_color, theme, animations_enabled, updated_at)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, settings.PanelName, settings.PanelLogo, settings.PrimaryColor, settings.SecondaryColor, settings.AccentColor, settings.Theme, settings.AnimationsEnabled)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Settings updated successfully"})
}

// Activity Logs
func getActivityLogs(c *fiber.Ctx) error {
	rows, err := db.Query(`
		SELECT id, user_id, action, description, ip_address, user_agent, created_at
		FROM activity_logs ORDER BY created_at DESC LIMIT 100
	`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var log map[string]interface{} = make(map[string]interface{})
		var id, userID, action, description, ipAddress, userAgent, createdAt string
		
		err := rows.Scan(&id, &userID, &action, &description, &ipAddress, &userAgent, &createdAt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		
		log["id"] = id
		log["user_id"] = userID
		log["action"] = action
		log["description"] = description
		log["ip_address"] = ipAddress
		log["user_agent"] = userAgent
		log["created_at"] = createdAt
		
		logs = append(logs, log)
	}

	return c.JSON(logs)
}