package ui

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/netsys-lab/scion-orchestrator/pkg/metrics"
)

// Comment struct
type Comment struct {
	ID      int
	Author  string
	Text    string
	Replies []Comment
}

//go:embed templates/*
var uiFiles embed.FS

func fileServerWithRedirect(fs http.FileSystem) http.Handler {

	fileServer := http.FileServer(fs)
	localFolder := filepath.Join("ui", "templates")
	// Check if the directory exists
	if _, err := os.Stat(localFolder); err == nil {
		fmt.Println("[UI] Local dev mode: Serving static files from:", localFolder)
		fileServer = http.FileServer(http.Dir(localFolder))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "ui/index.html")
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	metrics.Status.LastUpdated = time.Now().Format(time.RFC3339)
	json.NewEncoder(w).Encode(metrics.Status)
}

var engine *TemplateEngine // Global template engine

// Page handler using the engine
func renderPage(w http.ResponseWriter, r *http.Request) {
	content, err := engine.Render("index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(content))
}

// Comment handler using the engine
func renderComment(w http.ResponseWriter, r *http.Request) {
	comment := Comment{
		ID:     1,
		Author: "Alice",
		Text:   "This is a top-level comment!",
		Replies: []Comment{
			{ID: 2, Author: "Bob", Text: "This is a reply to Alice!"},
			{ID: 3, Author: "Charlie", Text: "Another reply to Alice!"},
		},
	}

	content, err := engine.Render("comment.html", comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(content))
}

func RunUIHTTPServer(url string) error {
	engine = NewTemplateEngine("ui/templates")
	// h := http.NewServeMux()

	// Register the templates
	engine.RegisterTemplate("/", "index.html", renderPage)
	engine.RegisterTemplate("/comment", "comment.html", renderComment)

	//h.Handle("/", fileServerWithRedirect(http.FS(uiFiles)))
	//h.HandleFunc("/status.json", statusHandler)
	log.Println("[UI] Server running on ", url)
	err := engine.Run(url)
	return err
}
