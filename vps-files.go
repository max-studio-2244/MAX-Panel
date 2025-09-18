package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	CreatedAt string `json:"created_at"`
}

func getFiles(c *fiber.Ctx) error {
	serverID := c.Params("id")
	path := c.Query("path", "")

	var workDir string
	err := db.QueryRow(`SELECT work_dir FROM servers WHERE id = ?`, serverID).Scan(&workDir)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	targetPath := workDir
	if path != "" {
		targetPath = filepath.Join(workDir, path)
	}

	// Security check - ensure path is within server directory
	if !strings.HasPrefix(targetPath, workDir) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read directory"})
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		relativePath := path
		if relativePath != "" {
			relativePath = filepath.Join(relativePath, entry.Name())
		} else {
			relativePath = entry.Name()
		}

		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    relativePath,
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	return c.JSON(files)
}

func downloadFile(c *fiber.Ctx) error {
	serverID := c.Params("id")
	filePath := c.Query("path")

	if filePath == "" {
		return c.Status(400).JSON(fiber.Map{"error": "File path required"})
	}

	var workDir string
	err := db.QueryRow(`SELECT work_dir FROM servers WHERE id = ?`, serverID).Scan(&workDir)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	fullPath := filepath.Join(workDir, filePath)

	// Security check
	if !strings.HasPrefix(fullPath, workDir) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{"error": "File not found"})
	}

	return c.SendFile(fullPath)
}

func uploadFile(c *fiber.Ctx) error {
	serverID := c.Params("id")
	uploadPath := c.FormValue("path", "")

	var workDir string
	err := db.QueryRow(`SELECT work_dir FROM servers WHERE id = ?`, serverID).Scan(&workDir)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
	}

	targetDir := workDir
	if uploadPath != "" {
		targetDir = filepath.Join(workDir, uploadPath)
	}

	// Security check
	if !strings.HasPrefix(targetDir, workDir) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create directory"})
	}

	// Save file
	targetPath := filepath.Join(targetDir, file.Filename)
	if err := c.SaveFile(file, targetPath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save file"})
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

	var workDir string
	err := db.QueryRow(`SELECT work_dir FROM servers WHERE id = ?`, serverID).Scan(&workDir)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	fullPath := filepath.Join(workDir, req.Path)

	// Security check
	if !strings.HasPrefix(fullPath, workDir) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(req.Content), 0644); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save file"})
	}

	return c.JSON(fiber.Map{"message": "File saved successfully"})
}

func deleteFile(c *fiber.Ctx) error {
	serverID := c.Params("id")
	filePath := c.Query("path")

	if filePath == "" {
		return c.Status(400).JSON(fiber.Map{"error": "File path required"})
	}

	var workDir string
	err := db.QueryRow(`SELECT work_dir FROM servers WHERE id = ?`, serverID).Scan(&workDir)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	fullPath := filepath.Join(workDir, filePath)

	// Security check
	if !strings.HasPrefix(fullPath, workDir) {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	if err := os.RemoveAll(fullPath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete file"})
	}

	return c.JSON(fiber.Map{"message": "File deleted successfully"})
}

func getBackups(c *fiber.Ctx) error {
	serverID := c.Params("id")

	rows, err := db.Query(`SELECT id, server_id, name, size, path, created_at FROM backups WHERE server_id = ? ORDER BY created_at DESC`, serverID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var backups []Backup
	for rows.Next() {
		var b Backup
		var size *int64
		err := rows.Scan(&b.ID, &b.ServerID, &b.Name, &size, &b.Path, &b.CreatedAt)
		if err != nil {
			continue
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

	var workDir string
	err := db.QueryRow(`SELECT work_dir FROM servers WHERE id = ?`, serverID).Scan(&workDir)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Server not found"})
	}

	backupID := uuid.New().String()
	backupName := req.Name
	if backupName == "" {
		backupName = fmt.Sprintf("backup-%s", time.Now().Format("2006-01-02-15-04-05"))
	}

	backupPath := filepath.Join("./backups", fmt.Sprintf("%s-%s.zip", serverID, backupName))

	// Create backup asynchronously
	go func() {
		if err := createServerBackup(workDir, backupPath); err != nil {
			return
		}

		// Get file size
		stat, err := os.Stat(backupPath)
		if err != nil {
			return
		}

		// Save to database
		db.Exec(`INSERT INTO backups (id, server_id, name, path, size) VALUES (?, ?, ?, ?, ?)`,
			backupID, serverID, backupName, backupPath, stat.Size())
	}()

	return c.JSON(fiber.Map{
		"id":      backupID,
		"message": "Backup started",
	})
}

func createServerBackup(sourceDir, backupPath string) error {
	// Create backup directory
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return err
	}

	// Create zip file
	zipFile, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk through source directory
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Create file in zip
		zipFileWriter, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Copy file content
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(zipFileWriter, file)
		return err
	})
}

func deleteBackup(c *fiber.Ctx) error {
	serverID := c.Params("id")
	backupID := c.Params("backup_id")

	var backupPath string
	err := db.QueryRow(`SELECT path FROM backups WHERE id = ? AND server_id = ?`, backupID, serverID).Scan(&backupPath)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Backup not found"})
	}

	// Delete file
	if backupPath != "" {
		os.Remove(backupPath)
	}

	// Delete from database
	_, err = db.Exec(`DELETE FROM backups WHERE id = ?`, backupID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Backup deleted successfully"})
}