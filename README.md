# Website

This project is a Go web server that provides several services, including a dice rolling utility. It uses a SQLite database for data persistence and `htmx` for dynamic frontend interactions. The application is containerized using Docker.

## Getting Started

### Prerequisites

- Docker
- Go (for local development)

### Building and Running with Docker

The project can be built and run using the provided `rebuild.sh` script. This script will:

1.  Pull the latest changes from the git repository.
2.  Build a new Docker image with the tag `website:latest`.
3.  Stop the existing `website` container, relying on the system to automatically restart it with the new image.

To run the project, execute the following command:

```bash
./rebuild.sh
```

**Note:** The `rebuild.sh` script assumes a NixOS environment where the `website` service is configured to restart automatically.

### Local Development

For development, you can run the application directly using the Go toolchain:

```bash
go run .
```

The server will start on port 8080 by default. You can change the port by setting the `PORT` environment variable. The database connection string can be configured using the `DB_DSN` environment variable.

## Project Structure

-   `main.go`: The main entry point of the application. It sets up the database connection, defines the HTTP routes, and starts the web server.
-   `dice.go`: Contains the handlers and logic for the dice rolling feature.
-   `go.mod`, `go.sum`: Go module files that manage the project's dependencies.
-   `migrations/`: Contains SQL scripts for database schema migrations.
-   `ui/`: Contains the HTML templates and static assets for the frontend.
    -   `base.tmpl`: The base HTML template.
    -   `dice_home.tmpl`, `dice_dist.tmpl`: Templates for the dice rolling feature.
    -   `static/`: Contains static assets like CSS and JavaScript files.
-   `Dockerfile`: Defines the Docker image for the application.
-   `rebuild.sh`: A script for building and restarting the application.

## Development

-   The web server is built using the standard Go `net/http` package.
-   Frontend interactions are enhanced with `htmx`.
-   The database is SQLite, and migrations are handled with raw SQL scripts.
-   The application is designed to be deployed as a Docker container.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
