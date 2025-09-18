package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
)

type CloudflareConfig struct {
	APIToken string `json:"api_token"`
	ZoneID   string `json:"zone_id"`
	Domain   string `json:"domain"`
	Email    string `json:"email"`
}

type CloudflareDNSRecord struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

type CloudflareResponse struct {
	Success bool                  `json:"success"`
	Errors  []CloudflareError     `json:"errors"`
	Result  []CloudflareDNSRecord `json:"result"`
}

type CloudflareError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func setupCloudflare(c *fiber.Ctx) error {
	var config CloudflareConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Validate Cloudflare credentials
	if err := validateCloudflareCredentials(config); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid Cloudflare credentials: " + err.Error()})
	}

	// Get server IP
	serverIP := getServerIP()
	if serverIP == "" {
		return c.Status(500).JSON(fiber.Map{"error": "Could not determine server IP"})
	}

	// Create DNS records
	records := []CloudflareDNSRecord{
		{
			Type:    "A",
			Name:    config.Domain,
			Content: serverIP,
			TTL:     300,
			Proxied: true,
		},
		{
			Type:    "A",
			Name:    "*." + config.Domain,
			Content: serverIP,
			TTL:     300,
			Proxied: false, // Wildcard can't be proxied
		},
	}

	var createdRecords []string
	for _, record := range records {
		recordID, err := createCloudflareRecord(config, record)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create DNS record: " + err.Error()})
		}
		createdRecords = append(createdRecords, recordID)
	}

	// Save Cloudflare config to database
	configJSON, _ := json.Marshal(config)
	_, err := db.Exec(`
		INSERT OR REPLACE INTO panel_settings (id, panel_name, domain, cloudflare_config) 
		VALUES (1, (SELECT panel_name FROM panel_settings WHERE id = 1), ?, ?)`,
		config.Domain, string(configJSON))

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save configuration"})
	}

	// Generate SSL certificate via Cloudflare
	if err := generateCloudflareSSL(config); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate SSL certificate: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Cloudflare setup completed successfully",
		"domain":  config.Domain,
		"records": createdRecords,
		"ssl":     "Generated",
	})
}

func validateCloudflareCredentials(config CloudflareConfig) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s", config.ZoneID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+config.APIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid credentials or zone ID")
	}

	return nil
}

func createCloudflareRecord(config CloudflareConfig, record CloudflareDNSRecord) (string, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", config.ZoneID)
	
	jsonData, err := json.Marshal(record)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+config.APIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var cfResp struct {
		Success bool `json:"success"`
		Result  struct {
			ID string `json:"id"`
		} `json:"result"`
		Errors []CloudflareError `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return "", err
	}

	if !cfResp.Success {
		if len(cfResp.Errors) > 0 {
			return "", fmt.Errorf(cfResp.Errors[0].Message)
		}
		return "", fmt.Errorf("unknown error")
	}

	return cfResp.Result.ID, nil
}

func generateCloudflareSSL(config CloudflareConfig) error {
	// Create SSL directory
	sslDir := "./ssl"
	os.MkdirAll(sslDir, 0755)

	// Generate Cloudflare Origin Certificate
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/certificates")
	
	certRequest := map[string]interface{}{
		"type": "origin-ca",
		"hostnames": []string{
			config.Domain,
			"*." + config.Domain,
		},
		"requested_validity": 5475, // 15 years
		"request_type":       "origin-ca",
	}

	jsonData, err := json.Marshal(certRequest)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+config.APIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var certResp struct {
		Success bool `json:"success"`
		Result  struct {
			Certificate string `json:"certificate"`
			PrivateKey  string `json:"private_key"`
		} `json:"result"`
		Errors []CloudflareError `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&certResp); err != nil {
		return err
	}

	if !certResp.Success {
		if len(certResp.Errors) > 0 {
			return fmt.Errorf(certResp.Errors[0].Message)
		}
		return fmt.Errorf("failed to generate certificate")
	}

	// Save certificate and key
	certPath := filepath.Join(sslDir, "cert.pem")
	keyPath := filepath.Join(sslDir, "key.pem")

	if err := os.WriteFile(certPath, []byte(certResp.Result.Certificate), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(keyPath, []byte(certResp.Result.PrivateKey), 0600); err != nil {
		return err
	}

	return nil
}

func getServerIP() string {
	// Try to get public IP from multiple sources
	sources := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ipecho.net/plain",
	}

	for _, source := range sources {
		resp, err := http.Get(source)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			body := make([]byte, 64)
			n, err := resp.Body.Read(body)
			if err == nil {
				ip := strings.TrimSpace(string(body[:n]))
				if ip != "" {
					return ip
				}
			}
		}
	}

	return ""
}

func getCloudflareConfig(c *fiber.Ctx) error {
	var configJSON sql.NullString
	err := db.QueryRow(`SELECT cloudflare_config FROM panel_settings WHERE id = 1`).Scan(&configJSON)
	
	if err != nil || !configJSON.Valid {
		return c.JSON(fiber.Map{
			"configured": false,
			"domain":     "",
		})
	}

	var config CloudflareConfig
	if err := json.Unmarshal([]byte(configJSON.String), &config); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid configuration"})
	}

	// Don't return sensitive data
	return c.JSON(fiber.Map{
		"configured": true,
		"domain":     config.Domain,
		"email":      config.Email,
	})
}

func updateCloudflareRecord(c *fiber.Ctx) error {
	var req struct {
		RecordType string `json:"record_type"`
		Name       string `json:"name"`
		Content    string `json:"content"`
		Proxied    bool   `json:"proxied"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Get Cloudflare config
	var configJSON sql.NullString
	err := db.QueryRow(`SELECT cloudflare_config FROM panel_settings WHERE id = 1`).Scan(&configJSON)
	if err != nil || !configJSON.Valid {
		return c.Status(400).JSON(fiber.Map{"error": "Cloudflare not configured"})
	}

	var config CloudflareConfig
	if err := json.Unmarshal([]byte(configJSON.String), &config); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid configuration"})
	}

	record := CloudflareDNSRecord{
		Type:    req.RecordType,
		Name:    req.Name,
		Content: req.Content,
		TTL:     300,
		Proxied: req.Proxied,
	}

	recordID, err := createCloudflareRecord(config, record)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create record: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":   "DNS record created successfully",
		"record_id": recordID,
	})
}

func removeCloudflare(c *fiber.Ctx) error {
	// Remove Cloudflare configuration
	_, err := db.Exec(`
		UPDATE panel_settings 
		SET domain = '', cloudflare_config = NULL 
		WHERE id = 1`)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Remove SSL certificates
	os.RemoveAll("./ssl")

	return c.JSON(fiber.Map{"message": "Cloudflare configuration removed"})
}