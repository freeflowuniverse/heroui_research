# MD Viewer

A modern Markdown viewer application that dynamically renders Markdown files with a clean UI, built with Go Fiber, PicoCSS, and Unpoly.

## Features

- Clean, responsive design with PicoCSS
- Server-side rendering with Go Fiber and Pug templates
- Dynamic content loading without full page refreshes using Unpoly
- Markdown rendering with proper syntax highlighting
- Automatic navigation generation from Markdown files
- On-page anchor navigation for headings
- Configurable sidebar with custom sections
- Image support in Markdown files
- Back-to-top button for easy navigation

## Project Structure

```
md-viewer/
├── main.go                # Main application file
├── content/               # Markdown content files
│   ├── about.md
│   ├── features.md
│   ├── image-examples.md
│   ├── usage.md
│   └── welcome.md
├── web/
│   ├── static/            # Static assets
│   │   ├── css/
│   │   │   └── style.css
│   │   ├── data/
│   │   │   └── sidebar.json  # Sidebar configuration
│   │   ├── images/        # Image assets
│   │   └── js/
│   │       └── main.js
│   └── templates/         # Pug templates
│       ├── layout.pug     # Main layout template
│       ├── search.pug     # Search results template
│       └── sidebar.pug    # Sidebar template
└── go.mod                 # Go module file
```

## Getting Started

1. Install dependencies:
   ```
   go mod tidy
   ```

2. Start the server:
   ```
   go run main.go
   ```

3. Open your browser and navigate to `http://localhost:3000`

## Adding Content

To add new Markdown files, simply place them in the `content` directory. The application will automatically detect and list them in the navigation.

## Customizing the Sidebar

The sidebar is configurable through the `web/static/data/sidebar.json` file. You can add custom sections and links by modifying this file.

## Technologies Used

- **Go Fiber v2.52.0**: Fast, flexible web framework for Go
- **Pug Templates**: Template engine for HTML
- **PicoCSS v2**: Minimal CSS framework for clean designs
- **Unpoly v2.7.1**: JavaScript library for dynamic content loading
- **Goldmark v1.6.0**: Markdown parser and compiler for Go
