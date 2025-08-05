# JioTV Go - Copilot Instructions

This document provides instructions for GitHub Copilot to better understand the JioTV Go project and provide more relevant assistance.

## 1. Project Overview

JioTV Go is a web application that allows users to stream Live TV channels from JioTV on the web and IPTV clients. It acts as a web wrapper around the JioTV Android app, using the same APIs.

The key features include:
- Live TV streaming
- M3U playlist support for IPTV
- A web interface for watching TV
- EPG (Electronic Program Guide) support
- Authentication via Jio ID/Number (OTP or password)
- A command-line interface (CLI) for server management.

## 2. Tech Stack

The project is a monorepo containing a Go backend and a vanilla HTML/CSS/JS frontend.

### Backend
- **Language:** Go
- **Web Framework:** [Fiber](https://gofiber.io/) (`github.com/gofiber/fiber/v2`)
- **CLI Framework:** [urfave/cli](https://cli.urfave.org/) (`github.com/urfave/cli/v2`)
- **Configuration:** [cleanenv](https://github.com/ilyakaznacheev/cleanenv) is used for loading configurations from files (TOML, YAML, JSON) or environment variables.

### Frontend
- **Styling:** [TailwindCSS](https://tailwindcss.com/) with the [DaisyUI](https://daisyui.com/) component library.
- **JavaScript:** Vanilla JavaScript. No major frameworks like React or Vue are used.
- **Video Players:**
    - [Flowplayer](https://flowplayer.com/) for HLS streaming.
    - [Shaka Player](https://shaka-player-demo.appspot.com/) for DRM-protected DASH streaming.

### Testing
- **Backend:** Standard Go testing (`go test`).
- **Frontend:** [Jest](https://jestjs.io/) with `jsdom`.

## 3. Project Structure

- `main.go`: Main entry point for the application.
- `cmd/`: Contains the CLI command definitions and their logic (e.g., `serve`, `login`, `update`).
- `internal/`: Houses all the core application logic, split into sub-packages:
    - `config/`: Configuration management.
    - `constants/`: Project-wide constants.
    - `handlers/`: HTTP handlers for the Fiber web server.
    - `middleware/`: Custom Fiber middleware.
- `pkg/`: Contains reusable packages like EPG generation, scheduling, and utilities.
- `web/`: All frontend assets.
    - `static/`: CSS, JavaScript, icons, and external libraries.
    - `views/`: Go HTML templates.
- `docs/`: Project documentation, built using `mdbook`.
- `.github/`: Contains GitHub-specific files like workflows, issue templates, and these instructions.
- `Dockerfile`: Defines the production Docker image.
- `docker-compose.yml` & `dev.Dockerfile`: For development using Docker Compose.

## 4. Development Workflow

### Building the Project
- **Go Backend:** From the root directory, run `go build .`
- **Frontend CSS:** Navigate to the `web/` directory and run `npm run build` to compile TailwindCSS.

### Running Tests
- **Go Tests:** From the root directory, run `go test -v ./...`
- **Frontend Tests:** Navigate to the `web/` directory and run `npm test`.

### Running the Application
- **Locally:** `go run main.go serve`
- **With Docker (for development):** `docker-compose up`

### Documentation
The documentation in the `docs/` directory is built using `mdbook`.
- To build the book: `mdbook build docs`

## 5. Conventions and Standards

### Commit Messages
Please follow the conventional commit format. The issue templates (`.github/ISSUE_TEMPLATE/`) provide prefixes like:
- `feat:` for new features.
- `bug:` for bug fixes.
- `security:` for security-related changes.
- `chore:` for maintenance tasks.
- `docs:` for documentation updates.
- `test:` for test-related changes.
- `refactor:` for code refactoring.

### Branching Strategy
- Use `main` for production-ready code.
- Use `develop` for ongoing development.
- Feature branches should be named `feature/<description>`.
- Bugfix branches should be named `bugfix/<description>`.
- Use `hotfix/<description>` for urgent fixes that need to go directly to `main`.

### Code Style
- **Go:** Follow standard Go conventions (`gofmt`). Code should be well-commented, especially public functions and complex logic.
- **Frontend:** Adhere to the existing style. Use TailwindCSS utility classes for styling.

### Dependencies
- **Go:** Manage with `go.mod`. Run `go mod tidy` after changing dependencies.
- **Frontend:** Manage with `npm` in the `web/` directory. Use `package.json` and `package-lock.json`.

## 6. License

The project is licensed under the **Creative Commons Attribution 4.0 International** license. See the `LICENSE` file for details.
