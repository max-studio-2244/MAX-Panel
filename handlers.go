package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Server struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Game        string `json:"game"`
	Image       string `json:"image"`
	Port        int    `json:"port"`
	Memory      int    `json:"memory"`
	CPU         float64 `json:"cpu"`
	Status      string `json:"status"`
	ContainerID string `json:"container_id"`
	CreatedAt   string `json:"created_at"`
}

var dockerClient *client.Client

func init() {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal("Failed to create Docker client:", err)
	}
}

func getServers(c *fiber.Ctx) error {
	// Get user from token (simplified - in production, implement proper JWT middleware)
	userID := getUserFromToken(c)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	
	// Check if user is admin
	var isAdmin bool
	err := db.QueryRow("SELECT is_admin FROM users WHERE id = ?", userID).Scan(&isAdmin)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	var query string
	var args []interface{}
	
	if isAdmin {
		// Admins see all servers
		query = "SELECT id, name, game, image, port, memory, cpu, status, container_id, created_at FROM servers"
	} else {
		// Regular users only see their assigned servers
		query = "SELECT id, name, game, image, port, memory, cpu, status, container_id, created_at FROM servers WHERE owner_id = ?"
		args = append(args, userID)
	}
	
	rows, err := db.Query(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var s Server
		var containerID sql.NullString
		err := rows.Scan(&s.ID, &s.Name, &s.Game, &s.Image, &s.Port, &s.Memory, &s.CPU, &s.Status, &containerID, &s.CreatedAt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		s.ContainerID = containerID.String
		servers = append(servers, s)
	}

	return c.JSON(servers)
}

func createServer(c *fiber.Ctx) error {
	// Check if user is admin
	userID := getUserFromToken(c)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	
	var isAdmin bool
	err := db.QueryRow("SELECT is_admin FROM users WHERE id = ?", userID).Scan(&isAdmin)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	if !isAdmin {
		return c.Status(403).JSON(fiber.Map{"error": "Only administrators can create servers"})
	}
	
	var req struct {
		Name    string  `json:"name"`
		Game    string  `json:"game"`
		Port    int     `json:"port"`
		Memory  int     `json:"memory"`
		CPU     float64 `json:"cpu"`
		OwnerID string  `json:"owner_id"` // Admin can assign server to specific user
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Game image mapping
	gameImages := map[string]string{
		"minecraft": "itzg/minecraft-server:latest",
		"csgo":      "steamcmd/steamcmd:latest",
		"rust":      "didstopia/rust-server:latest",
		"ark":       "steamcmd/steamcmd:latest",
	}

	image, exists := gameImages[req.Game]
	if !exists {
		return c.Status(400).JSON(fiber.Map{"error": "Unsupported game"})
	}

	serverID := uuid.New().String()
	
	// If no owner specified, assign to admin
	ownerID := req.OwnerID
	if ownerID == "" {
		ownerID = userID
	}

	// Insert into database with owner
	_, err = db.Exec(`
		INSERT INTO servers (id, name, game, image, port, memory, cpu, status, owner_id, egg_id, node_id, disk)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'stopped', ?, 'default', 'local', 5000)
	`, serverID, req.Name, req.Game, image, req.Port, req.Memory, req.CPU, ownerID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"id":      serverID,
		"message": "Server created successfully",
	})
}

func getServer(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var s Server
	var containerID sql.NullString
	err := db.QueryRow(`
		SELECT id, name, game, image, port, memory, cpu, status, container_id, created_at 
		FROM servers WHERE id = ?
	`, id).Scan(&s.ID, &s.Name, &s.Game, &s.Image, &s.Port, &s.Memory, &s.CPU, &s.Status, &containerID, &s.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	s.ContainerID = containerID.String
	return c.JSON(s)
}

func startServer(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var server Server
	var containerID sql.NullString
	err := db.QueryRow(`
		SELECT id, name, game, image, port, memory, cpu, status, container_id 
		FROM servers WHERE id = ?
	`, id).Scan(&server.ID, &server.Name, &server.Game, &server.Image, &server.Port, &server.Memory, &server.CPU, &server.Status, &containerID)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	ctx := context.Background()

	// If container exists, start it
	if containerID.String != "" {
		err = dockerClient.ContainerStart(ctx, containerID.String, types.ContainerStartOptions{})
		if err != nil {
			// Container might be removed, create new one
			return createAndStartContainer(id, server)
		}
	} else {
		// Create new container
		return createAndStartContainer(id, server)
	}

	// Update status
	_, err = db.Exec("UPDATE servers SET status = 'running' WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Server started"})
}

func createAndStartContainer(serverID string, server Server) error {
	ctx := context.Background()

	// Container configuration
	config := &container.Config{
		Image: server.Image,
		Env: []string{
			"EULA=TRUE",
			fmt.Sprintf("MEMORY=%dM", server.Memory),
		},
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", server.Port)): {},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", server.Port)): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: strconv.Itoa(server.Port),
				},
			},
		},
		Resources: container.Resources{
			Memory:   int64(server.Memory * 1024 * 1024),
			NanoCPUs: int64(server.CPU * 1000000000),
		},
	}

	networkConfig := &network.NetworkingConfig{}

	// Create container
	resp, err := dockerClient.ContainerCreate(ctx, config, hostConfig, networkConfig, nil, server.Name)
	if err != nil {
		return err
	}

	// Start container
	err = dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	// Update database with container ID
	_, err = db.Exec("UPDATE servers SET container_id = ?, status = 'running' WHERE id = ?", resp.ID, serverID)
	return err
}

func stopServer(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var containerID sql.NullString
	err := db.QueryRow("SELECT container_id FROM servers WHERE id = ?", id).Scan(&containerID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	if containerID.String != "" {
		ctx := context.Background()
		err = dockerClient.ContainerStop(ctx, containerID.String, container.StopOptions{})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	// Update status
	_, err = db.Exec("UPDATE servers SET status = 'stopped' WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Server stopped"})
}

func restartServer(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var containerID sql.NullString
	err := db.QueryRow("SELECT container_id FROM servers WHERE id = ?", id).Scan(&containerID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	if containerID.String != "" {
		ctx := context.Background()
		err = dockerClient.ContainerRestart(ctx, containerID.String, container.StopOptions{})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(fiber.Map{"message": "Server restarted"})
}

func killServer(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var containerID sql.NullString
	err := db.QueryRow("SELECT container_id FROM servers WHERE id = ?", id).Scan(&containerID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	if containerID.String != "" {
		ctx := context.Background()
		err = dockerClient.ContainerKill(ctx, containerID.String, "SIGKILL")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	// Update status
	_, err = db.Exec("UPDATE servers SET status = 'stopped' WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Server killed"})
}

func getServerStats(c *fiber.Ctx) error {
	id := c.Params("id")
	
	var containerID sql.NullString
	err := db.QueryRow("SELECT container_id FROM servers WHERE id = ?", id).Scan(&containerID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	if containerID.String == "" {
		return c.JSON(fiber.Map{
			"cpu": 0,
			"memory_used": 0,
			"disk_used": 0,
			"uptime": 0,
		})
	}

	ctx := context.Background()
	
	// Get container stats
	stats, err := dockerClient.ContainerStats(ctx, containerID.String, false)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer stats.Body.Close()

	// Parse stats (simplified)
	return c.JSON(fiber.Map{
		"cpu": 25,        // Mock data
		"memory_used": 512,
		"disk_used": 1024,
		"uptime": 3600,
	})
}

func getUserFromToken(c *fiber.Ctx) string {
	// Simplified token validation - in production, implement proper JWT
	auth := c.Get("Authorization")
	if auth == "" {
		return ""
	}
	
	// For now, return mock admin user ID
	// In production, decode JWT and return actual user ID
	return "admin"
}

func deleteServer(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Check if user is admin
	userID := getUserFromToken(c)
	if userID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	
	var isAdmin bool
	err := db.QueryRow("SELECT is_admin FROM users WHERE id = ?", userID).Scan(&isAdmin)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	if !isAdmin {
		return c.Status(403).JSON(fiber.Map{"error": "Only administrators can delete servers"})
	}
	
	var containerID sql.NullString
	err = db.QueryRow("SELECT container_id FROM servers WHERE id = ?", id).Scan(&containerID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	// Stop and remove container
	if containerID.String != "" {
		ctx := context.Background()
		dockerClient.ContainerStop(ctx, containerID.String, container.StopOptions{})
		dockerClient.ContainerRemove(ctx, containerID.String, types.ContainerRemoveOptions{})
	}

	// Delete from database
	_, err = db.Exec("DELETE FROM servers WHERE id = ?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Server deleted"})
}

func getFiles(c *fiber.Ctx) error {
	// Simplified file listing - would need proper implementation
	return c.JSON([]string{"server.properties", "world/", "logs/"})
}

func downloadFile(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "File download not implemented yet"})
}

func uploadFile(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "File upload not implemented yet"})
}

func handleConsoleUpgrade(c *fiber.Ctx) error {
	return websocket.New(handleConsole)(c)
}

func handleConsole(c *websocket.Conn) {
	serverID := c.Params("id")
	
	var containerID sql.NullString
	err := db.QueryRow("SELECT container_id FROM servers WHERE id = ?", serverID).Scan(&containerID)
	if err != nil {
		c.WriteMessage(websocket.TextMessage, []byte("Error: Server not found"))
		return
	}

	if containerID.String == "" {
		c.WriteMessage(websocket.TextMessage, []byte("Error: Container not running"))
		return
	}

	ctx := context.Background()
	
	// Get container logs
	logs, err := dockerClient.ContainerLogs(ctx, containerID.String, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "100",
	})
	if err != nil {
		c.WriteMessage(websocket.TextMessage, []byte("Error getting logs: "+err.Error()))
		return
	}
	defer logs.Close()

	// Stream logs to WebSocket
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := logs.Read(buf)
			if err != nil {
				if err != io.EOF {
					c.WriteMessage(websocket.TextMessage, []byte("Error reading logs: "+err.Error()))
				}
				break
			}
			
			// Remove Docker log headers (first 8 bytes)
			logData := buf[8:n]
			if len(logData) > 0 {
				c.WriteMessage(websocket.TextMessage, logData)
			}
		}
	}()

	// Handle incoming messages (commands)
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		// Execute command in container
		command := strings.TrimSpace(string(msg))
		if command != "" {
			execConfig := types.ExecConfig{
				Cmd:          []string{"sh", "-c", command},
				AttachStdout: true,
				AttachStderr: true,
			}

			execResp, err := dockerClient.ContainerExecCreate(ctx, containerID.String, execConfig)
			if err != nil {
				c.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
				continue
			}

			execAttach, err := dockerClient.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
			if err != nil {
				c.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
				continue
			}

			// Read command output
			output, _ := io.ReadAll(execAttach.Reader)
			c.WriteMessage(websocket.TextMessage, output)
			execAttach.Close()
		}
	}
}