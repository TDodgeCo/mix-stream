package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

const configFilePath = "config.json"

type Config struct {
	Directories  []string `json:"directories"`
	NgrokDomains []string `json:"ngrok_domains"`
}

var currentConfig Config
var templates = template.Must(template.ParseGlob("templates/*.html"))

type FileData struct {
	Name        string
	EscapedPath string
}

type DirectoryData struct {
	Path  string
	Files []FileData
}

type TemplateData struct {
	Directories  []DirectoryData
	NgrokDomains []string
}

func main() {
	// Load or create config
	currentConfig = loadConfig()

	// Run ngrok for each domain
	for _, domain := range currentConfig.NgrokDomains {
		go runNgrok(domain)
	}

	// Serve files and static assets
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(currentConfig.Directories[0]))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Web interface
	http.HandleFunc("/", handleFileList)
	http.HandleFunc("/update", handleUpdateConfig)

	// Start HTTP server
	fmt.Println("Serving music on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Load config or create default
func loadConfig() Config {
	var config Config

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		saveConfig(Config{}) // Create default config
		log.Println("Created default config file.")
	} else if err == nil {
		file, err := os.ReadFile(configFilePath)
		if err != nil {
			log.Fatal("Error reading config file:", err)
		}
		err = json.Unmarshal(file, &config)
		if err != nil {
			log.Fatal("Error parsing config file:", err)
		}
	} else {
		log.Fatal("Error checking config file:", err)
	}

	return config
}

// Save config
func saveConfig(config Config) {
	file, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatal("Error saving config:", err)
	}
	err = os.WriteFile(configFilePath, file, 0644)
	if err != nil {
		log.Fatal("Error writing config file:", err)
	}
}

// Run ngrok for a domain
func runNgrok(domain string) {
	cmd := exec.Command("ngrok", "http", "--domain="+domain, "8080")
	err := cmd.Start()
	if err != nil {
		log.Printf("Error starting ngrok for domain %s: %v\n", domain, err)
		return
	}
	fmt.Printf("ngrok is running at domain: %s\n", domain)
}

// Display files and directories
func handleFileList(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		NgrokDomains: currentConfig.NgrokDomains,
	}

	if len(currentConfig.Directories) > 0 {
		for _, dir := range currentConfig.Directories {
			dirData := DirectoryData{Path: dir}
			var files []FileData

			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.IsDir() && isAudioFile(path) {
					relativePath, err := filepath.Rel(dir, path)
					if err != nil {
						return err
					}
					files = append(files, FileData{
						Name:        filepath.Base(path),
						EscapedPath: url.PathEscape(relativePath),
					})
				}
				return nil
			})

			if err != nil {
				fmt.Fprintf(w, "<p>Error reading directory: %s</p>", err)
				continue
			}

			dirData.Files = files
			data.Directories = append(data.Directories, dirData)
		}
	}

	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Helper to check if the file is an audio file
func isAudioFile(fileName string) bool {
	ext := filepath.Ext(fileName)
	switch ext {
	case ".mp3", ".wav", ".ogg", ".flac", ".aac":
		return true
	default:
		return false
	}
}

// Update config with new directory and ngrok domain
func handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		newDirectory := r.FormValue("directory")
		newNgrok := r.FormValue("ngrok")

		if !contains(currentConfig.Directories, newDirectory) {
			currentConfig.Directories = append(currentConfig.Directories, newDirectory)
		}
		if !contains(currentConfig.NgrokDomains, newNgrok) {
			currentConfig.NgrokDomains = append(currentConfig.NgrokDomains, newNgrok)
		}

		saveConfig(currentConfig)
		go runNgrok(newNgrok)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Check if a slice contains an item
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
