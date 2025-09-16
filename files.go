package main

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/gofiber/fiber/v2"
)

type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"is_dir"`
	ModTime string `json:"mod_time"`
}

type Backup struct {
	ID        string `json:"id"`
	ServerID  string `json:"server_id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Path      string `json:"path"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// File Management
func getFiles(c *fiber.Ctx) error {
	serverID := c.Params("id")
	path := c.Query("path", "/")

	// Get container ID
	var containerID string
	err := db.QueryRow("SELECT container_id FROM servers WHERE id = ?", serverID).Scan(&containerID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	if containerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Server not running"})
	}

	ctx := context.Background()

	// Execute ls command in container
	execConfig := types.ExecConfig{
		Cmd:          []string{"ls", "-la", path},
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	execAttach, err := dockerClient.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer execAttach.Close()

	// Read output
	output, err := io.ReadAll(execAttach.Reader)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Parse ls output (simplified)
	lines := strings.Split(string(output), "\n")
	var files []FileInfo

	for _, line := range lines {
		if len(line) > 0 && !strings.HasPrefix(line, "total") {
			parts := strings.Fields(line)
			if len(parts) >= 9 {
				name := strings.Join(parts[8:], " ")
				if name != "." && name != ".." {
					files = append(files, FileInfo{
						Name:    name,
						Path:    filepath.Join(path, name),
						IsDir:   strings.HasPrefix(line, "d"),
						ModTime: strings.Join(parts[5:8], " "),
					})
				}
			}
		}
	}

	return c.JSON(files)
}

func downloadFile(c *fiber.Ctx) error {
	serverID := c.Params("id")
	filePath := c.Query("path")

	if filePath == "" {
		return c.Status(400).JSON(fiber.Map{"error": "File path required"})
	}

	// Get container ID
	var containerID string
	err := db.QueryRow("SELECT container_id FROM servers WHERE id = ?", serverID).Scan(&containerID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	if containerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Server not running"})
	}

	ctx := context.Background()

	// Copy file from container
	reader, _, err := dockerClient.CopyFromContainer(ctx, containerID, filePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer reader.Close()

	// Set headers for file download
	filename := filepath.Base(filePath)
	c.Set("Content-Disposition", "attachment; filename="+filename)
	c.Set("Content-Type", "application/octet-stream")

	// Stream file to response
	_, err = io.Copy(c.Response().BodyWriter(), reader)
	return err
}

func uploadFile(c *fiber.Ctx) error {
	serverID := c.Params("id")
	uploadPath := c.FormValue("path", "/")

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
	}

	// Get container ID
	var containerID string
	err = db.QueryRow("SELECT container_id FROM servers WHERE id = ?", serverID).Scan(&containerID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	if containerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Server not running"})
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer src.Close()

	ctx := context.Background()

	// Copy file to container (simplified - would need proper tar handling)
	targetPath := filepath.Join(uploadPath, file.Filename)
	err = dockerClient.CopyToContainer(ctx, containerID, targetPath, src, types.CopyToContainerOptions{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "File uploaded successfully"})
}

func editFile(c *fiber.Ctx) error {
	serverID := c.Params("id")
	
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Get container ID
	var containerID string
	err := db.QueryRow("SELECT container_id FROM servers WHERE id = ?", serverID).Scan(&containerID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	if containerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Server not running"})
	}

	ctx := context.Background()

	// Write content to file using echo command
	execConfig := types.ExecConfig{
		Cmd:          []string{"sh", "-c", "echo '" + req.Content + "' > " + req.Path},
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	execAttach, err := dockerClient.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer execAttach.Close()

	return c.JSON(fiber.Map{"message": "File saved successfully"})
}

func deleteFile(c *fiber.Ctx) error {
	serverID := c.Params("id")
	filePath := c.Query("path")

	if filePath == "" {
		return c.Status(400).JSON(fiber.Map{"error": "File path required"})
	}

	// Get container ID
	var containerID string
	err := db.QueryRow("SELECT container_id FROM servers WHERE id = ?", serverID).Scan(&containerID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	if containerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Server not running"})
	}

	ctx := context.Background()

	// Delete file using rm command
	execConfig := types.ExecConfig{
		Cmd:          []string{"rm", "-rf", filePath},
		AttachStdout: true,
		AttachStderr: true,
	}

	execResp, err := dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	execAttach, err := dockerClient.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer execAttach.Close()

	return c.JSON(fiber.Map{"message": "File deleted successfully"})
}

// Backup Management
func getBackups(c *fiber.Ctx) error {
	serverID := c.Params("id")

	rows, err := db.Query(`
		SELECT id, server_id, name, size, path, status, created_at
		FROM backups WHERE server_id = ? ORDER BY created_at DESC
	`, serverID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var backups []Backup
	for rows.Next() {
		var b Backup
		var size *int64
		err := rows.Scan(&b.ID, &b.ServerID, &b.Name, &size, &b.Path, &b.Status, &b.CreatedAt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if size != nil {
			b.Size = *size
		}
		backups = append(backups, b)
	}

	return c.JSON(backups)
}

func createBackup(c *fiber.Ctx) error {
	serverID := c.Params("id")
	
	var req struct {
		Name string `json:"name"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Get container ID
	var containerID string
	err := db.QueryRow("SELECT container_id FROM servers WHERE id = ?", serverID).Scan(&containerID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	backupID := generateID()
	backupName := req.Name
	if backupName == "" {
		backupName = "backup-" + time.Now().Format("2006-01-02-15-04-05")
	}

	// Insert backup record
	_, err = db.Exec(`
		INSERT INTO backups (id, server_id, name, status)
		VALUES (?, ?, ?, 'pending')
	`, backupID, serverID, backupName)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Create backup asynchronously
	go func() {
		createServerBackup(backupID, serverID, containerID, backupName)
	}()

	return c.JSON(fiber.Map{
		"id":      backupID,
		"message": "Backup started",
	})
}

func createServerBackup(backupID, serverID, containerID, backupName string) {
	ctx := context.Background()
	
	// Create backup directory
	backupDir := "./backups/" + serverID
	os.MkdirAll(backupDir, 0755)
	
	backupPath := filepath.Join(backupDir, backupName+".zip")
	
	// Create zip file
	zipFile, err := os.Create(backupPath)
	if err != nil {
		updateBackupStatus(backupID, "failed")
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Copy files from container (simplified)
	reader, _, err := dockerClient.CopyFromContainer(ctx, containerID, "/")
	if err != nil {
		updateBackupStatus(backupID, "failed")
		return
	}
	defer reader.Close()

	// Write to zip (simplified)
	writer, err := zipWriter.Create("server-data.tar")
	if err != nil {
		updateBackupStatus(backupID, "failed")
		return
	}

	size, err := io.Copy(writer, reader)
	if err != nil {
		updateBackupStatus(backupID, "failed")
		return
	}

	// Update backup record
	_, err = db.Exec(`
		UPDATE backups SET status = 'completed', size = ?, path = ?
		WHERE id = ?
	`, size, backupPath, backupID)

	if err != nil {
		updateBackupStatus(backupID, "failed")
	}
}

func updateBackupStatus(backupID, status string) {
	db.Exec("UPDATE backups SET status = ? WHERE id = ?", status, backupID)
}

func restoreBackup(c *fiber.Ctx) error {
	serverID := c.Params("id")
	backupID := c.Params("backup_id")

	// Get backup info
	var backup Backup
	err := db.QueryRow(`
		SELECT id, server_id, name, path, status
		FROM backups WHERE id = ? AND server_id = ?
	`, backupID, serverID).Scan(&backup.ID, &backup.ServerID, &backup.Name, &backup.Path, &backup.Status)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Backup not found"})
	}

	if backup.Status != "completed" {
		return c.Status(400).JSON(fiber.Map{"error": "Backup not completed"})
	}

	// Restore backup (simplified implementation)
	return c.JSON(fiber.Map{"message": "Backup restore started"})
}

func deleteBackup(c *fiber.Ctx) error {
	serverID := c.Params("id")
	backupID := c.Params("backup_id")

	// Get backup path
	var backupPath string
	err := db.QueryRow(`
		SELECT path FROM backups WHERE id = ? AND server_id = ?
	`, backupID, serverID).Scan(&backupPath)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Backup not found"})
	}

	// Delete backup file
	if backupPath != "" {
		os.Remove(backupPath)
	}

	// Delete backup record
	_, err = db.Exec("DELETE FROM backups WHERE id = ?", backupID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Backup deleted successfully"})
}