package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vitistack/ipam-api/cmd/ipam-api/settings"
	"github.com/vitistack/ipam-api/internal/logger"
)

var outPath string
var mongoBackup = &cobra.Command{
	Use:   "mongo-backup",
	Short: "Backup MongoDB data",
	Run: func(cmd *cobra.Command, args []string) {
		if err := backup(); err != nil {
			logger.Log.Errorln(err.Error())
		}
	},
}

func init() {
	mongoBackup.Flags().StringVar(&outPath, "out", "", "output path (required)")
	if err := mongoBackup.MarkFlagRequired("out"); err != nil {
		log.Fatalf("Failed to mark 'out' flag as required: %v", err)
	}
	RootCmd.AddCommand(mongoBackup)
}

func backup() error {
	err := settings.InitConfig()

	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	err = logger.InitLogger("./logs")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	backupDir := "./backup"
	if err := os.MkdirAll(backupDir, 0750); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Validate and sanitize the output filename
	sanitizedFilename := filepath.Base(filepath.Clean(outPath))
	if sanitizedFilename == "" || sanitizedFilename == "." || sanitizedFilename == ".." {
		return fmt.Errorf("invalid output filename: %s", outPath)
	}

	archivePath := filepath.Join(backupDir, sanitizedFilename)

	// Build and validate MongoDB URI components
	mongoURI, err := buildValidatedMongoURI()
	if err != nil {
		return fmt.Errorf("failed to build MongoDB URI: %w", err)
	}

	// Use literal arguments to prevent command injection
	// #nosec G204 -- mongoURI and archivePath are validated above to prevent injection
	cmd := exec.Command("mongodump",
		"--uri", mongoURI,
		"--archive", archivePath,
		"--gzip")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Log.Infoln("Running MongoDB backup...")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mongodump failed: %w", err)
	}

	logger.Log.Infof("MongoDB backup completed: Backup saved as %s\n", outPath)

	if err := cleanupOldBackups(backupDir); err != nil {
		logger.Log.Warnf("Failed to clean up old backups: %v", err)
	}

	return nil

}

func buildValidatedMongoURI() (string, error) {
	host := viper.GetString("mongodb.host")
	user := viper.GetString("mongodb.username")
	pass := viper.GetString("mongodb.password")

	// Validate required fields
	if host == "" {
		return "", fmt.Errorf("mongodb host is required")
	}
	if user == "" {
		return "", fmt.Errorf("mongodb username is required")
	}
	if pass == "" {
		return "", fmt.Errorf("mongodb password is required")
	}

	// Basic validation to prevent injection - ensure no special characters that could break URI
	if containsInvalidChars(host) || containsInvalidChars(user) {
		return "", fmt.Errorf("mongodb connection parameters contain invalid characters")
	}

	return fmt.Sprintf("mongodb://%s:%s@%s:27017/?authSource=admin&readPreference=primary&ssl=false", user, pass, host), nil
}

func containsInvalidChars(s string) bool {
	// Check for characters that could cause URI injection or other issues
	invalidChars := " @/?#[]\n\r\t"
	for _, char := range invalidChars {
		if strings.ContainsRune(s, char) {
			return true
		}
	}
	return false
}

func cleanupOldBackups(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-48 * time.Hour) // 48 hours ago

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		if !info.Mode().IsRegular() {
			continue
		}

		fullPath := filepath.Join(dir, file.Name())
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(fullPath); err != nil {
				logger.Log.Warnf("Could not delete %s: %v", fullPath, err)
			} else {
				logger.Log.Infof("Deleted old backup: %s", fullPath)
			}
		}
	}

	return nil
}
