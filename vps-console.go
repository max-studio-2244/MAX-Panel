package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/websocket/v2"
)

func handleConsole(c *websocket.Conn) {
	serverID := c.Params("id")

	serversMutex.RLock()
	server, exists := servers[serverID]
	serversMutex.RUnlock()

	if !exists {
		c.WriteMessage(websocket.TextMessage, []byte("Error: Server not found"))
		return
	}

	// Send initial message
	c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Connected to %s console\n", server.Name)))

	// Stream existing logs
	go streamLogs(c, server)

	// Handle incoming commands
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		command := strings.TrimSpace(string(msg))
		if command == "" {
			continue
		}

		// Handle special commands
		switch command {
		case "/clear":
			c.WriteMessage(websocket.TextMessage, []byte("\033[2J\033[H"))
		case "/status":
			status := "stopped"
			if server.Process != nil {
				status = "running"
			}
			c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Server status: %s\n", status)))
		case "/stop":
			if server.Process != nil {
				server.Process.Process.Kill()
				c.WriteMessage(websocket.TextMessage, []byte("Server stopped\n"))
			} else {
				c.WriteMessage(websocket.TextMessage, []byte("Server is not running\n"))
			}
		default:
			// Send command to server process
			if server.Process != nil {
				// For Minecraft servers, we can send commands via stdin
				if server.Game == "minecraft" {
					if stdin := server.Process.Stdin; stdin != nil {
						stdin.Write([]byte(command + "\n"))
						c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("> %s\n", command)))
					}
				} else {
					c.WriteMessage(websocket.TextMessage, []byte("Command execution not supported for this server type\n"))
				}
			} else {
				c.WriteMessage(websocket.TextMessage, []byte("Server is not running\n"))
			}
		}
	}
}

func streamLogs(c *websocket.Conn, server *GameServer) {
	logPath := filepath.Join(server.WorkDir, "server.log")

	// Send existing log content
	if file, err := os.Open(logPath); err == nil {
		scanner := bufio.NewScanner(file)
		lineCount := 0
		var lines []string

		// Read all lines
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		file.Close()

		// Send last 50 lines
		start := len(lines) - 50
		if start < 0 {
			start = 0
		}

		for i := start; i < len(lines); i++ {
			c.WriteMessage(websocket.TextMessage, []byte(lines[i]+"\n"))
		}
	}

	// Watch for new log entries
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastSize int64 = 0
	if stat, err := os.Stat(logPath); err == nil {
		lastSize = stat.Size()
	}

	for {
		select {
		case <-ticker.C:
			if stat, err := os.Stat(logPath); err == nil {
				if stat.Size() > lastSize {
					// File has grown, read new content
					if file, err := os.Open(logPath); err == nil {
						file.Seek(lastSize, 0)
						scanner := bufio.NewScanner(file)
						for scanner.Scan() {
							c.WriteMessage(websocket.TextMessage, []byte(scanner.Text()+"\n"))
						}
						file.Close()
						lastSize = stat.Size()
					}
				}
			}
		}
	}
}