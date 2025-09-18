package main

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PanelSettings struct {
	ID        int    `json:"id"`
	PanelName string `json:"panel_name"`
	Domain    string `json:"domain"`
	NodeToken string `json:"node_token"`
}

type Node struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Token       string `json:"token"`
	Status      string `json:"status"`
	LastPing    string `json:"last_ping"`
	ServerCount int    `json:"server_count"`
}

func getSettings(c *fiber.Ctx) error {
	var settings PanelSettings
	err := db.QueryRow(`SELECT id, panel_name, domain, node_token FROM panel_settings WHERE id = 1`).
		Scan(&settings.ID, &settings.PanelName, &settings.Domain, &settings.NodeToken)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return default settings
			settings = PanelSettings{
				ID:        1,
				PanelName: "MAX Panel",
				Domain:    "",
				NodeToken: uuid.New().String(),
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

	// Generate new token if not provided
	if settings.NodeToken == "" {
		settings.NodeToken = uuid.New().String()
	}

	_, err := db.Exec(`
		INSERT OR REPLACE INTO panel_settings (id, panel_name, domain, node_token) 
		VALUES (1, ?, ?, ?)
	`, settings.PanelName, settings.Domain, settings.NodeToken)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Settings updated successfully"})
}

func getNodes(c *fiber.Ctx) error {
	// For VPS version, return local node info
	nodes := []Node{
		{
			ID:          "local",
			Name:        "Local VPS Node",
			Host:        "localhost",
			Port:        8080,
			Token:       "local-node-token",
			Status:      "online",
			LastPing:    "2024-01-01 12:00:00",
			ServerCount: len(servers),
		},
	}

	// Add external nodes if configured
	var domain string
	db.QueryRow(`SELECT domain FROM panel_settings WHERE id = 1`).Scan(&domain)
	
	if domain != "" {
		nodes = append(nodes, Node{
			ID:          "external",
			Name:        "External Node",
			Host:        domain,
			Port:        8080,
			Token:       uuid.New().String(),
			Status:      "pending",
			LastPing:    "",
			ServerCount: 0,
		})
	}

	return c.JSON(nodes)
}

func createNode(c *fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	nodeID := uuid.New().String()
	nodeToken := uuid.New().String()

	return c.JSON(fiber.Map{
		"id":    nodeID,
		"token": nodeToken,
		"message": "Node created successfully",
		"instructions": fiber.Map{
			"install_command": fmt.Sprintf("curl -sSL https://raw.githubusercontent.com/maxpanel/installer/main/install.sh | bash -s -- --token=%s --host=%s --port=%d", nodeToken, req.Host, req.Port),
			"docker_command":  fmt.Sprintf("docker run -d --name maxpanel-node -p %d:8080 -e PANEL_TOKEN=%s -e PANEL_HOST=%s maxpanel/node:latest", req.Port, nodeToken, req.Host),
		},
	})
}

func getAdminUsers(c *fiber.Ctx) error {
	rows, err := db.Query(`SELECT id, username, email, is_admin, created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var user map[string]interface{} = make(map[string]interface{})
		var id, username, email, createdAt string
		var isAdmin bool
		
		err := rows.Scan(&id, &username, &email, &isAdmin, &createdAt)
		if err != nil {
			continue
		}
		
		user["id"] = id
		user["username"] = username
		user["email"] = email
		user["is_admin"] = isAdmin
		user["created_at"] = createdAt
		
		users = append(users, user)
	}

	return c.JSON(users)
}