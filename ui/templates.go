package ui

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"
)

// Template struct: Represents a template with a path and file name
type Template struct {
	Path         string
	FileName     string
	RenderMethod func(http.ResponseWriter, *http.Request)
}

// TemplateEngine struct: Manages the templates
type TemplateEngine struct {
	mux         *http.ServeMux
	templates   map[string]Template
	templateDir string
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(templateDir string) *TemplateEngine {
	return &TemplateEngine{
		templates:   make(map[string]Template),
		templateDir: templateDir,
		mux:         http.NewServeMux(),
	}
}

func (engine *TemplateEngine) Run(url string) error {
	return http.ListenAndServe(url, engine.mux)
}

// RegisterTemplate adds a template to the engine and registers the HTTP handler
func (engine *TemplateEngine) RegisterTemplate(path, fileName string, renderFunc func(http.ResponseWriter, *http.Request)) {
	engine.templates[path] = Template{
		Path:         path,
		FileName:     fileName,
		RenderMethod: renderFunc,
	}
	engine.mux.HandleFunc(path, renderFunc) // Automatically register HTTP route
}

// Render method: Loads and executes a template
func (engine *TemplateEngine) Render(templateFile string, data interface{}) (string, error) {
	tmplPath := filepath.Join(engine.templateDir, templateFile)
	fmt.Print(tmplPath)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
