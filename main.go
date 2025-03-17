package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/pug/v2"
	"github.com/yuin/goldmark"
)

type MarkdownFile struct {
	Name  string
	Path  string
	Title string
}

// Anchor represents a heading in a markdown file
type Anchor struct {
	ID    string
	Text  string
	Level int
}

// SidebarItem represents an item in a sidebar section
type SidebarItem struct {
	Title    string `json:"Title"`
	Href     string `json:"Href"`
	External bool   `json:"External"`
}

// SidebarSection represents a section in the sidebar
type SidebarSection struct {
	Title string        `json:"Title"`
	Items []SidebarItem `json:"Items"`
}

func main() {
	// Initialize template engine
	engine := pug.New("./web/templates", ".pug")
	// Add template functions
	engine.AddFunc("title", func(s string) string {
		return strings.Title(s)
	})
	// Add function to render unescaped HTML
	engine.AddFunc("unescaped", func(s string) template.HTML {
		return template.HTML(s)
	})

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Static files
	app.Static("/", "./web/static")
	app.Static("/css", "./web/static/css")
	app.Static("/js", "./web/static/js")
	app.Static("/content", "./content")

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		files, err := getMarkdownFiles()
		if err != nil {
			return err
		}

		// Load sidebar data
		sidebar, err := loadSidebarData()
		if err != nil {
			log.Printf("Error loading sidebar data: %v", err)
			// Continue with empty sidebar data if there's an error
			sidebar = []SidebarSection{}
		}

		var defaultFile *MarkdownFile
		for _, file := range files {
			if file.Name == "welcome" {
				defaultFile = &file
				break
			}
		}

		if defaultFile == nil && len(files) > 0 {
			defaultFile = &files[0]
		}

		if defaultFile != nil {
			return c.Redirect("/" + defaultFile.Name)
		}

		return c.Render("layout", fiber.Map{
			"files":      files,
			"content":    "<p>No markdown files found in content directory.</p>",
			"activeFile": nil,
			"sidebar":    sidebar,
		})
	})

	app.Get("/:filename", func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		files, err := getMarkdownFiles()
		if err != nil {
			return err
		}

		// Load sidebar data
		sidebar, err := loadSidebarData()
		if err != nil {
			log.Printf("Error loading sidebar data: %v", err)
			// Continue with empty sidebar data if there's an error
			sidebar = []SidebarSection{}
		}

		filePath := filepath.Join("content", filename+".md")
		var activeFile *MarkdownFile

		for _, file := range files {
			if file.Name == filename {
				activeFile = &file
				break
			}
		}

		// Read the markdown file
		markdownData, err := os.ReadFile(filePath)
		if err != nil {
			return c.Status(fiber.StatusNotFound).Render("layout", fiber.Map{
				"files":      files,
				"content":    "<p>Markdown file not found.</p>",
				"activeFile": nil,
				"safe":       true,
				"sidebar":    sidebar,
			})
		}

		// Extract anchors from the markdown content
		anchors := extractAnchors(markdownData)

		// Render markdown to HTML
		content, err := renderMarkdown(filePath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).Render("layout", fiber.Map{
				"files":      files,
				"content":    "<p>Error rendering markdown content.</p>",
				"activeFile": nil,
				"safe":       true,
				"sidebar":    sidebar,
			})
		}

		// Add IDs to headings in the HTML content
		for _, anchor := range anchors {
			headingTag := fmt.Sprintf("<h%d>", anchor.Level)
			headingWithID := fmt.Sprintf("<h%d id=\"%s\">", anchor.Level, anchor.ID)
			content = strings.Replace(content, headingTag+anchor.Text, headingWithID+anchor.Text, 1)
		}

		return c.Render("layout", fiber.Map{
			"files":      files,
			"content":    content,
			"activeFile": activeFile,
			"anchors":    anchors,
			"safe":       true,
			"sidebar":    sidebar,
		})
	})

	// API endpoint to get raw markdown content
	app.Get("/api/content/:filename", func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		filePath := filepath.Join("content", filename+".md")

		content, err := renderMarkdown(filePath)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "File not found",
			})
		}

		return c.JSON(fiber.Map{
			"content": content,
		})
	})

	// Start server
	log.Println("Starting server on :3002")
	if err := app.Listen(":3002"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

// getMarkdownFiles returns a list of markdown files in the content directory
func getMarkdownFiles() ([]MarkdownFile, error) {
	files, err := os.ReadDir("content")
	if err != nil {
		return nil, err
	}

	var markdownFiles []MarkdownFile
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			name := strings.TrimSuffix(file.Name(), ".md")
			title := getMarkdownTitle(filepath.Join("content", file.Name()))
			markdownFiles = append(markdownFiles, MarkdownFile{
				Name:  name,
				Path:  file.Name(),
				Title: title,
			})
		}
	}

	return markdownFiles, nil
}

// getMarkdownTitle extracts the title from a markdown file
func getMarkdownTitle(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return filepath.Base(filePath)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}

	return strings.TrimSuffix(filepath.Base(filePath), ".md")
}

// extractAnchors extracts headings from markdown content
func extractAnchors(content []byte) []Anchor {
	lines := strings.Split(string(content), "\n")
	var anchors []Anchor

	for _, line := range lines {
		// Check if the line is a heading
		if strings.HasPrefix(line, "#") {
			// Count the number of # to determine heading level
			level := 0
			for i := 0; i < len(line) && line[i] == '#'; i++ {
				level++
			}

			// Extract the heading text
			text := strings.TrimSpace(line[level:])

			// Generate an ID for the heading
			id := strings.ToLower(text)
			id = strings.ReplaceAll(id, " ", "-")
			id = strings.ReplaceAll(id, ".", "")
			id = strings.ReplaceAll(id, "?", "")
			id = strings.ReplaceAll(id, "!", "")
			id = strings.ReplaceAll(id, ",", "")
			id = strings.ReplaceAll(id, ":", "")
			id = strings.ReplaceAll(id, ";", "")
			id = strings.ReplaceAll(id, "(", "")
			id = strings.ReplaceAll(id, ")", "")

			anchors = append(anchors, Anchor{
				ID:    id,
				Text:  text,
				Level: level,
			})
		}
	}

	return anchors
}

// renderMarkdown renders a markdown file to HTML
func renderMarkdown(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	md := goldmark.New()
	if err := md.Convert(content, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// loadSidebarData loads the sidebar structure from a JSON file
func loadSidebarData() ([]SidebarSection, error) {
	var sidebar []SidebarSection

	// Read the sidebar JSON file
	filePath := "./web/static/data/sidebar.json"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return sidebar, fmt.Errorf("error reading sidebar JSON: %w", err)
	}

	// Parse the JSON into the sidebar slice
	err = json.Unmarshal(content, &sidebar)
	if err != nil {
		return sidebar, fmt.Errorf("error parsing sidebar JSON: %w", err)
	}

	return sidebar, nil
}
