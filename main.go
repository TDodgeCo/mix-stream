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
	// Load or prompt for config
	currentConfig = loadConfig()

	// Run ngrok for each domain
	for _, domain := range currentConfig.NgrokDomains {
		go runNgrok(domain)
	}

	// Serve files over HTTP if there are directories
	if len(currentConfig.Directories) > 0 {
		http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(currentConfig.Directories[0]))))
	} else {
		fmt.Println("No directories configured. Please add one through the web interface.")
	}

	// Simple web interface to display files and settings
	http.HandleFunc("/", handleFileList)
	http.HandleFunc("/update", handleUpdateConfig)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start HTTP server
	fmt.Println("Serving music on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":3333", nil))
}

// Load the config or return an empty config
func loadConfig() Config {
	var config Config

	// Check if config file exists
	if _, err := os.Stat(configFilePath); err == nil {
		// Load config from file
		file, err := os.ReadFile(configFilePath)
		if err != nil {
			log.Fatal("Error reading config file:", err)
		}
		err = json.Unmarshal(file, &config)
		if err != nil {
			log.Fatal("Error parsing config file:", err)
		}
	}

	return config
}

// Save the config to a file
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

// Run the ngrok command in the background
func runNgrok(domain string) {
	cmd := exec.Command("ngrok", "http", "--domain="+domain, "8080")
	err := cmd.Start() // Run ngrok command in the background
	if err != nil {
		log.Fatal("Error starting ngrok:", err)
	}
	fmt.Printf("ngrok is running at domain: %s\n", domain)
}

// Serve a list of music files and directories as a simple HTML page
func handleFileList(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		NgrokDomains: currentConfig.NgrokDomains,
	}

	if len(currentConfig.Directories) > 0 {
		for _, dir := range currentConfig.Directories {
			dirData := DirectoryData{
				Path: dir,
			}

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
					escapedFilePath := url.PathEscape(relativePath)
					files = append(files, FileData{
						Name:        filepath.Base(path),
						EscapedPath: escapedFilePath,
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

// Helper function to check if the file is an audio file (e.g., .mp3, .wav)
func isAudioFile(fileName string) bool {
	ext := filepath.Ext(fileName)
	switch ext {
	case ".mp3", ".wav", ".ogg", ".flac", ".aac":
		return true
	default:
		return false
	}
}

// Handle updating the directories and ngrok domains
func handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Get new directory and ngrok domain from form input
		newDirectory := r.FormValue("directory")
		newNgrok := r.FormValue("ngrok")

		// Check if directory is valid
		if _, err := os.Stat(newDirectory); os.IsNotExist(err) {
			fmt.Fprintf(w, "<html><body><h1>Directory Not Found</h1>")
			fmt.Fprintf(w, "<p>Error: %s</p>", err)
			fmt.Fprintf(w, "<a href=\"/\">Go back</a></body></html>")
			return
		}

		// Add the new directory and ngrok domain to the config
		currentConfig.Directories = append(currentConfig.Directories, newDirectory)
		currentConfig.NgrokDomains = append(currentConfig.NgrokDomains, newNgrok)

		// Save the updated config
		saveConfig(currentConfig)

		// Run ngrok for the new domain
		go runNgrok(newNgrok)

		// Redirect to the main page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
