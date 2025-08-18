# JioTV Go - Copilot Instructions

**ALWAYS follow these instructions first and fallback to search or additional context gathering ONLY when the information here is incomplete or found to be in error.**

JioTV Go is a web application that allows users to stream Live TV channels from JioTV on the web and IPTV clients. It acts as a web wrapper around the JioTV Android app, using the same APIs.

The key features include:
- Live TV streaming
- M3U playlist support for IPTV
- A web interface for watching TV
- EPG (Electronic Program Guide) support
- Authentication via Jio ID/Number (OTP or password)
- A command-line interface (CLI) for server management.

## Working Effectively

### Bootstrap and Build Repository
Run these commands in order. **NEVER CANCEL** any of these commands - wait for completion:

1. **Install Dependencies:**
   ```bash
   go mod tidy
   ```
   Takes ~15 seconds. NEVER CANCEL. Set timeout to 60+ seconds.

2. **Build Go Application:**
   ```bash
   go build -o build/jiotv_go .
   ```
   Takes ~25 seconds. NEVER CANCEL. Set timeout to 90+ seconds.

3. **Install Frontend Dependencies:**
   ```bash
   cd web && npm ci
   ```
   Takes ~50 seconds. NEVER CANCEL. Set timeout to 120+ seconds.

4. **Build Frontend (TailwindCSS):**
   ```bash
   cd web && npm run build
   ```
   Takes ~2 seconds. NEVER CANCEL. Set timeout to 30+ seconds.

### Run Tests
**CRITICAL**: Always run tests to verify your changes work correctly:

1. **Go Tests:**
   ```bash
   go test -v ./...
   ```
   Takes ~10 seconds. NEVER CANCEL. Set timeout to 60+ seconds.
   All tests should pass. Tests cover CLI utilities, handlers, and core functionality.

2. **Frontend Tests:**
   ```bash
   cd web && npm test -- --watchAll=false --ci
   ```
   Takes ~5 seconds. NEVER CANCEL. Set timeout to 30+ seconds.
   Tests use Jest with jsdom for JavaScript functionality.

### Run the Application

1. **Development Server (Manual Restart Required):**
   ```bash
   go run main.go serve --host 127.0.0.1 --port 5001
   ```
   Takes ~5 seconds to start. Server runs on http://127.0.0.1:5001

2. **Production Build:**
   ```bash
   ./build/jiotv_go serve --host 127.0.0.1 --port 5001
   ```

3. **Development with Auto-Reload (Docker):**
   ```bash
   docker compose up
   ```
   **NOTE**: May fail in some environments due to certificate issues. Use local development instead.

4. **Enable Debug Mode:**
   Set `JIOTV_DEBUG=true` environment variable for auto-reloading of templates and debug logs.

## Validation Scenarios

**ALWAYS test these scenarios after making changes to ensure the application works correctly:**

### CLI Validation
Test the main CLI commands work:
```bash
./build/jiotv_go --help
./build/jiotv_go serve --help  
./build/jiotv_go login --help
./build/jiotv_go epg --help
```

### Server Functionality  
1. Start the server: `./build/jiotv_go serve --host 127.0.0.1 --port 5001`
2. Verify server starts without critical errors (500 errors on main page are expected without authentication)
3. Test server responds: `curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:5001/` should return HTTP status code
4. Stop server with Ctrl+C

### Frontend Validation
After frontend changes:
1. Rebuild CSS: `cd web && npm run build`
2. Run frontend tests: `cd web && npm test -- --watchAll=false --ci`
3. Start server and verify web interface loads

## Tech Stack and Architecture

### Backend
- **Language:** Go 1.25 (REQUIRED - specified in go.mod)
- **Web Framework:** [Fiber](https://gofiber.io/) (`github.com/gofiber/fiber/v2`)
- **CLI Framework:** [urfave/cli](https://cli.urfave.org/) (`github.com/urfave/cli/v2`)
- **Configuration:** [cleanenv](https://github.com/ilyakaznacheev/cleanenv) loads from TOML, YAML, JSON or environment variables
- **Testing:** Standard Go testing (`go test`)

### Frontend
- **Styling:** [TailwindCSS v3](https://tailwindcss.com/) with [DaisyUI v4](https://daisyui.com/) component library
- **JavaScript:** Vanilla JavaScript (no major frameworks)
- **Video Players:** Flowplayer for HLS, Shaka Player for DRM-protected DASH
- **Testing:** [Jest](https://jestjs.io/) with `jsdom`

## Project Structure

### Key Directories
- `main.go`: Main entry point for the application
- `cmd/`: CLI command definitions and logic (serve, login, update, etc.)
- `internal/`: Core application logic split into sub-packages:
  - `config/`: Configuration management
  - `constants/`: Project-wide constants  
  - `handlers/`: HTTP handlers for Fiber web server
  - `middleware/`: Custom Fiber middleware
- `pkg/`: Reusable packages (EPG generation, scheduling, utilities)
- `web/`: All frontend assets
  - `static/`: CSS, JavaScript, icons, external libraries
  - `views/`: Go HTML templates
  - `package.json`: Frontend dependencies and build scripts
- `docs/`: Project documentation (built with mdbook)
- `.github/`: GitHub workflows, issue templates, and these instructions

### Build Artifacts
- `build/`: Contains compiled Go binary
- `web/static/internal/tailwind.css`: Generated TailwindCSS file

## Common Commands Reference

### Building
```bash
# Full build process
go mod tidy && go build -o build/jiotv_go . && cd web && npm ci && npm run build && cd ..

# Go only
go build -o build/jiotv_go .

# Frontend only  
cd web && npm run build
```

### Testing
```bash
# All tests
go test -v ./... && cd web && npm test -- --watchAll=false --ci

# Go tests only
go test -v ./...

# Frontend tests only
cd web && npm test -- --watchAll=false --ci
```

### Development
```bash
# Run server in development mode
JIOTV_DEBUG=true go run main.go serve --host 127.0.0.1 --port 5001

# Watch TailwindCSS changes
cd web && npm run watch

# Background server operations
./build/jiotv_go background start --args="serve --host 127.0.0.1 --port 5001"
./build/jiotv_go background stop
```

## GitHub Workflows

The project uses several GitHub Actions workflows:
- **`pr_tests.yml`**: Runs Go and frontend tests on PRs
- **`build-doc.yml`**: Builds mdbook documentation and deploys to GitHub Pages  
- **`docker.yml`**: Builds and pushes Docker images
- **`dependabot_action.yml`**: Rebuilds TailwindCSS when dependencies change
- **`release.yml`**: Creates releases with cross-platform binaries

## Development Notes

### Making Changes
- **Backend changes**: Modify Go files, run `go test ./...`, then test with `go run main.go serve`
- **Frontend changes**: Modify files in `web/`, run `cd web && npm run build`, then test server
- **Template changes**: Modify files in `web/views/`, enable `JIOTV_DEBUG=true` for auto-reload
- **CSS changes**: Modify `web/static/internal/input.css`, run `cd web && npm run build`

### Dependencies
- **Go**: Managed with `go.mod`, run `go mod tidy` after changes
- **Frontend**: Managed with `npm` in `web/` directory, use `npm ci` for clean installs

### Configuration
- Configuration loaded from files (TOML, YAML, JSON) or environment variables
- Default config files in `configs/` directory
- Runtime config path: `~/.jiotv_go/` or `JIOTV_PATH_PREFIX` environment variable

### Error Handling
- 500 errors on main page without authentication are expected behavior
- Use `JIOTV_DEBUG=true` for detailed logging
- Check logs for authentication and API connectivity issues

### Conventions and Standards

#### Commit Messages
Follow conventional commit format:
- `feat:` for new features
- `bug:` for bug fixes
- `security:` for security-related changes
- `chore:` for maintenance tasks
- `docs:` for documentation updates
- `test:` for test-related changes
- `refactor:` for code refactoring

#### Code Style
- **Go:** Follow standard Go conventions (`gofmt`). Code should be well-commented
- **Frontend:** Use TailwindCSS utility classes for styling

## License

The project is licensed under **Creative Commons Attribution 4.0 International**. See the `LICENSE` file for details.
