# Compose Parser

[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/badge/version-v0.3.4-blue.svg)](https://github.com/docker-graph/compose-parser)

A Go library for parsing Docker Compose files (docker-compose.yml) with support for Docker Compose specification v3 and v2.

**Version:** v0.3.4 
**Author:** Docker Graph Team  
**License:** MIT

## Features

- Parse Docker Compose YAML files into structured Go types
- Support for Docker Compose specification versions 2.x and 3.x
- Comprehensive type definitions for all Compose components:
  - Services with build configurations, ports, volumes, environment variables
  - Networks, volumes, secrets, and configs
  - Deployment configurations (replicas, resources, restart policies)
  - Health checks, logging, and labels
- Generate React Flow graph structures for visualization
- Error handling with detailed parsing errors

## Installation

```bash
go get github.com/docker-graph/compose-parser
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/docker-graph/compose-parser"
)

func main() {
    parser := compose_parser.NewComposeParser()
    
    // Parse a Docker Compose file
    project, err := parser.ParseFile("docker-compose.yml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Access parsed data
    fmt.Printf("Project: %s\n", project.Name)
    fmt.Printf("Version: %s\n", project.Version)
    fmt.Printf("Services: %d\n", len(project.Services))
    
    for name, service := range project.Services {
        fmt.Printf("Service: %s (Image: %s)\n", name, service.Image)
        fmt.Printf("  Ports: %v\n", service.Ports)
        fmt.Printf("  Volumes: %v\n", service.Volumes)
    }
}
```

## API Reference

### Types

#### `ComposeProjectConfig`
Main structure representing a Docker Compose project.

```go
type ComposeProjectConfig struct {
    Version      string                            `json:"version,omitempty"`
    Services     map[string]*ComposeServiceConfig  `json:"services"`
    Networks     map[string]*NetworkConfig         `json:"networks,omitempty"`
    Volumes      map[string]*VolumeConfig          `json:"volumes,omitempty"`
    Secrets      map[string]*SecretConfig          `json:"secrets,omitempty"`
    Configs      map[string]*ConfigConfig          `json:"configs,omitempty"`
    Name         string                            `json:"name"`
    // ... other fields
}
```

#### `ComposeServiceConfig`
Represents a single service configuration.

```go
type ComposeServiceConfig struct {
    Name       string       `json:"name"`
    Image      string       `json:"image,omitempty"`
    Build      *BuildConfig `json:"build,omitempty"`
    Ports      []PortMapping `json:"ports,omitempty"`
    Volumes    []VolumeMount `json:"volumes,omitempty"`
    // ... other fields
}
```

### Methods

#### `NewComposeParser() *ComposeParser`
Creates a new ComposeParser instance.

#### `ParseFile(filename string) (*ComposeProjectConfig, error)`
Parses a Docker Compose file from the filesystem.

#### `ParseYAML(yamlContent []byte) (*ComposeProjectConfig, error)`
Parses Docker Compose YAML content directly.

#### `ParseReader(reader io.Reader) (*ComposeProjectConfig, error)`
Parses Docker Compose content from an io.Reader.

#### `ParseFromDirectory(dirPath string) (*ComposeProjectConfig, error)`
Parses Docker Compose files from a directory (supports multiple compose files).

## Examples

### Example 1: Basic Parsing

```go
parser := compose_parser.NewComposeParser()
project, err := parser.ParseFile("docker-compose.yml")
if err != nil {
    log.Fatal(err)
}

// Work with parsed data
for serviceName, service := range project.Services {
    if service.Build != nil {
        fmt.Printf("%s uses build context: %s\n", serviceName, service.Build.Context)
    }
}
```

### Example 2: Generate React Flow Graph

```go
parser := compose_parser.NewComposeParser()
project, err := parser.ParseFile("docker-compose.yml")
if err != nil {
    log.Fatal(err)
}

// Generate graph for visualization
// (Note: Graph generation methods would need to be implemented based on your needs)
```

### Example 3: Parse Multiple Compose Files

```go
parser := compose_parser.NewComposeParser()
project, err := parser.ParseFromDirectory("./compose")
if err != nil {
    log.Fatal(err)
}
```

## Supported Docker Compose Features

- ✅ Services with build, image, command, entrypoint
- ✅ Port mappings (published, target, protocol, mode)
- ✅ Volume mounts (bind, volume, tmpfs, npipe)
- ✅ Environment variables and env files
- ✅ Networks and network modes
- ✅ Deploy configurations (replicas, resources, placement)
- ✅ Health checks and logging
- ✅ Secrets and configs
- ✅ Extends (service inheritance)
- ✅ Labels and custom metadata

## Error Handling

The parser returns detailed error messages including:
- YAML syntax errors
- Missing required fields
- Invalid field values
- Unsupported Compose versions

```go
project, err := parser.ParseFile("invalid-compose.yml")
if err != nil {
    if strings.Contains(err.Error(), "yaml:") {
        fmt.Println("YAML syntax error:", err)
    } else {
        fmt.Println("Compose validation error:", err)
    }
}
```

## Testing

Run tests with:

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Docker Compose Specification

This library supports the Docker Compose Specification:
- [Compose Specification v3](https://docs.docker.com/compose/compose-file/)
- [Compose Specification v2](https://docs.docker.com/compose/compose-file/compose-file-v2/)

## Related Projects

- [docker-graph](https://github.com/docker-graph) - Visualization tools for Docker Compose
- [docker/compose](https://github.com/docker/compose) - Official Docker Compose

## Support

For bugs, feature requests, or questions:
- Open an issue on GitHub
- Check existing issues before creating new ones

## Version History

- **v0.1.0** (Current): Initial release with basic Docker Compose parsing support

## Author

**Docker Graph Team**  
Maintainer of Docker Compose visualization and analysis tools

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.