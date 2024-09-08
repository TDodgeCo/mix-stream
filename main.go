package main

import (
	"encoding/json"
	"fmt"
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

	// Start HTTP server
	fmt.Println("Serving music on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
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
	fmt.Fprintf(w, "<html><body><h1>Music Files and Directories</h1>")

	// Check if there are any directories to display
	if len(currentConfig.Directories) == 0 {
		fmt.Fprintf(w, "<p>No directories found. Please add one below.</p>")
	} else {
		// Display each directory
		for _, dir := range currentConfig.Directories {
			fmt.Fprintf(w, "<h2>Directory: %s</h2>", dir)
			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					// Handle any error encountered while walking the path
					fmt.Fprintf(w, "<p>Error reading file: %s</p>", err)
					return err
				}

				if !info.IsDir() {
					escapedFilePath := url.PathEscape(path[len(dir)+1:])
					fmt.Fprintf(w, "<li><a href=\"/files/%s\">%s</a></li>", escapedFilePath, filepath.Base(path))
				}
				return nil
			})
			if err != nil {
				fmt.Fprintf(w, "<p>Error reading directory: %s</p>", err)
			}
		}
	}

	fmt.Fprintf(w, "<h2>ngrok Domains</h2><ul>")
	if len(currentConfig.NgrokDomains) == 0 {
		fmt.Fprintf(w, "<p>No ngrok domains configured. Please add one below.</p>")
	} else {
		for _, domain := range currentConfig.NgrokDomains {
			fmt.Fprintf(w, "<li>%s</li>", domain)
		}
	}
	fmt.Fprintf(w, "<h2>Update Config</h2>")
	fmt.Fprintf(w, "<form action=\"/update\" method=\"POST\">")
	fmt.Fprintf(w, "<h3>Add a new directory</h3>")
	fmt.Fprintf(w, "Directory Path: <input type=\"text\" name=\"directory\" required><br>")
	fmt.Fprintf(w, "<h3>Add a new ngrok domain</h3>")
	fmt.Fprintf(w, "ngrok Domain: <input type=\"text\" name=\"ngrok\" required><br>")
	fmt.Fprintf(w, "<input type=\"submit\" value=\"Update\">")
	fmt.Fprintf(w, "</form>")
	fmt.Fprintf(w, "</body></html>")
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
