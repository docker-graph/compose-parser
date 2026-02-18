package compose_parser_test

import (
	"testing"
	"time"

	"github.com/docker-graph/compose-parser"
)

func TestNewComposeParser(t *testing.T) {
	parser := compose_parser.NewComposeParser()
	if parser == nil {
		t.Error("NewComposeParser() returned nil")
	}
}

func TestParseYAML_BasicService(t *testing.T) {
	yamlContent := `
version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
    environment:
      - NODE_ENV=production
  db:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: secret
`

	parser := compose_parser.NewComposeParser()
	project, err := parser.ParseYAML([]byte(yamlContent))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if project == nil {
		t.Fatal("ParseYAML returned nil project")
	}

	if project.Version != "3.8" {
		t.Errorf("Expected version '3.8', got '%s'", project.Version)
	}

	if len(project.Services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(project.Services))
	}

	// Check web service
	web, ok := project.Services["web"]
	if !ok {
		t.Fatal("Service 'web' not found")
	}

	if web.Name != "web" {
		t.Errorf("Expected service name 'web', got '%s'", web.Name)
	}

	if web.Image != "nginx:latest" {
		t.Errorf("Expected image 'nginx:latest', got '%s'", web.Image)
	}

	if len(web.Ports) != 1 {
		t.Errorf("Expected 1 port mapping for web, got %d", len(web.Ports))
	} else {
		port := web.Ports[0]
		if port.Target != 80 || port.Published != 80 {
			t.Errorf("Expected port 80:80, got %d:%d", port.Target, port.Published)
		}
	}

	if web.Environment == nil {
		t.Error("Environment map is nil")
	} else if web.Environment["NODE_ENV"] != "production" {
		t.Errorf("Expected NODE_ENV=production, got %s", web.Environment["NODE_ENV"])
	}

	// Check db service
	db, ok := project.Services["db"]
	if !ok {
		t.Fatal("Service 'db' not found")
	}

	if db.Image != "postgres:15" {
		t.Errorf("Expected image 'postgres:15', got '%s'", db.Image)
	}

	if db.Environment == nil {
		t.Error("Environment map is nil for db")
	} else if db.Environment["POSTGRES_PASSWORD"] != "secret" {
		t.Errorf("Expected POSTGRES_PASSWORD=secret, got %s", db.Environment["POSTGRES_PASSWORD"])
	}
}

func TestParseYAML_WithBuild(t *testing.T) {
	yamlContent := `
version: '3.8'
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
      args:
        NODE_ENV: development
    ports:
      - "3000:3000"
`

	parser := compose_parser.NewComposeParser()
	project, err := parser.ParseYAML([]byte(yamlContent))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	app, ok := project.Services["app"]
	if !ok {
		t.Fatal("Service 'app' not found")
	}

	if app.Build == nil {
		t.Fatal("Build config is nil")
	}

	if app.Build.Context != "." {
		t.Errorf("Expected build context '.', got '%s'", app.Build.Context)
	}

	if app.Build.Dockerfile != "Dockerfile.dev" {
		t.Errorf("Expected dockerfile 'Dockerfile.dev', got '%s'", app.Build.Dockerfile)
	}

	if app.Build.Args == nil {
		t.Error("Build args map is nil")
	} else if app.Build.Args["NODE_ENV"] != "development" {
		t.Errorf("Expected NODE_ENV=development, got %s", app.Build.Args["NODE_ENV"])
	}
}

func TestParseYAML_WithVolumes(t *testing.T) {
	yamlContent := `
version: '3.8'
services:
  app:
    image: nginx
    volumes:
      - ./data:/usr/share/nginx/html:ro
      - cache:/cache
volumes:
  cache:
    driver: local
`

	parser := compose_parser.NewComposeParser()
	project, err := parser.ParseYAML([]byte(yamlContent))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	app, ok := project.Services["app"]
	if !ok {
		t.Fatal("Service 'app' not found")
	}

	if len(app.Volumes) != 2 {
		t.Errorf("Expected 2 volumes, got %d", len(app.Volumes))
	}

	// Check bind mount
	bindMount := app.Volumes[0]
	if bindMount.Type != "bind" {
		t.Errorf("Expected type 'bind', got '%s'", bindMount.Type)
	}
	if bindMount.Source != "./data" {
		t.Errorf("Expected source './data', got '%s'", bindMount.Source)
	}
	if bindMount.Target != "/usr/share/nginx/html" {
		t.Errorf("Expected target '/usr/share/nginx/html', got '%s'", bindMount.Target)
	}
	if !bindMount.ReadOnly {
		t.Error("Expected read-only mount")
	}

	// Check volume mount
	volumeMount := app.Volumes[1]
	if volumeMount.Type != "volume" {
		t.Errorf("Expected type 'volume', got '%s'", volumeMount.Type)
	}
	if volumeMount.Source != "cache" {
		t.Errorf("Expected source 'cache', got '%s'", volumeMount.Source)
	}
	if volumeMount.Target != "/cache" {
		t.Errorf("Expected target '/cache', got '%s'", volumeMount.Target)
	}

	// Check volumes section
	if len(project.Volumes) != 1 {
		t.Errorf("Expected 1 volume definition, got %d", len(project.Volumes))
	}

	cacheVolume, ok := project.Volumes["cache"]
	if !ok {
		t.Fatal("Volume 'cache' not found")
	}

	if cacheVolume.Driver != "local" {
		t.Errorf("Expected driver 'local', got '%s'", cacheVolume.Driver)
	}
}

func TestParseYAML_WithNetworks(t *testing.T) {
	yamlContent := `
version: '3.8'
services:
  web:
    image: nginx
    networks:
      - frontend
      - backend
networks:
  frontend:
    driver: bridge
  backend:
    driver: overlay
    attachable: true
`

	parser := compose_parser.NewComposeParser()
	project, err := parser.ParseYAML([]byte(yamlContent))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	web, ok := project.Services["web"]
	if !ok {
		t.Fatal("Service 'web' not found")
	}

	if len(web.Networks) != 2 {
		t.Errorf("Expected 2 networks, got %d", len(web.Networks))
	}

	// Check networks section
	if len(project.Networks) != 2 {
		t.Errorf("Expected 2 network definitions, got %d", len(project.Networks))
	}

	frontend, ok := project.Networks["frontend"]
	if !ok {
		t.Fatal("Network 'frontend' not found")
	}

	if frontend.Driver != "bridge" {
		t.Errorf("Expected driver 'bridge', got '%s'", frontend.Driver)
	}

	backend, ok := project.Networks["backend"]
	if !ok {
		t.Fatal("Network 'backend' not found")
	}

	if backend.Driver != "overlay" {
		t.Errorf("Expected driver 'overlay', got '%s'", backend.Driver)
	}

	if !backend.Attachable {
		t.Error("Expected attachable to be true")
	}
}

func TestParseYAML_InvalidYAML(t *testing.T) {
	invalidYAML := `
version: '3.8'
services:
  web:
    image: nginx
    ports
      - "80:80"  # Missing colon after ports
`

	parser := compose_parser.NewComposeParser()
	project, err := parser.ParseYAML([]byte(invalidYAML))
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
	if project != nil {
		t.Error("Expected nil project for invalid YAML")
	}
}

func TestParseYAML_EmptyServices(t *testing.T) {
	yamlContent := `
version: '3.8'
services:
`

	parser := compose_parser.NewComposeParser()
	project, err := parser.ParseYAML([]byte(yamlContent))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if project == nil {
		t.Fatal("ParseYAML returned nil project")
	}

	if len(project.Services) != 0 {
		t.Errorf("Expected 0 services, got %d", len(project.Services))
	}
}

func TestParseYAML_WithDeploy(t *testing.T) {
	yamlContent := `
version: '3.8'
services:
  app:
    image: nginx
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
`

	parser := compose_parser.NewComposeParser()
	project, err := parser.ParseYAML([]byte(yamlContent))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	app, ok := project.Services["app"]
	if !ok {
		t.Fatal("Service 'app' not found")
	}

	if app.Deploy == nil {
		t.Fatal("Deploy config is nil")
	}

	if app.Deploy.Replicas != 3 {
		t.Errorf("Expected 3 replicas, got %d", app.Deploy.Replicas)
	}

	if app.Deploy.Resources == nil {
		t.Fatal("Resources config is nil")
	}

	if app.Deploy.Resources.Limits == nil {
		t.Fatal("Resource limits is nil")
	}

	if app.Deploy.Resources.Limits.CPUs != "0.5" {
		t.Errorf("Expected CPU limit '0.5', got '%s'", app.Deploy.Resources.Limits.CPUs)
	}

	if app.Deploy.Resources.Limits.Memory != "512M" {
		t.Errorf("Expected memory limit '512M', got '%s'", app.Deploy.Resources.Limits.Memory)
	}

	if app.Deploy.RestartPolicy == nil {
		t.Fatal("Restart policy is nil")
	}

	if app.Deploy.RestartPolicy.Condition != "on-failure" {
		t.Errorf("Expected condition 'on-failure', got '%s'", app.Deploy.RestartPolicy.Condition)
	}

	if app.Deploy.RestartPolicy.Delay != "5s" {
		t.Errorf("Expected delay '5s', got '%s'", app.Deploy.RestartPolicy.Delay)
	}

	if app.Deploy.RestartPolicy.MaxAttempts != 3 {
		t.Errorf("Expected max attempts 3, got %d", app.Deploy.RestartPolicy.MaxAttempts)
	}
}

func TestProjectMetadata(t *testing.T) {
	yamlContent := `
version: '3.8'
services:
  web:
    image: nginx
`

	parser := compose_parser.NewComposeParser()
	project, err := parser.ParseYAML([]byte(yamlContent))
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if project.Name == "" {
		t.Error("Project name should not be empty")
	}

	if project.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	if project.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}

	if project.Status == "" {
		t.Error("Status should be set")
	}

	// Check that CreatedAt is before or equal to UpdatedAt
	if project.CreatedAt.After(project.UpdatedAt) {
		t.Error("CreatedAt should be before or equal to UpdatedAt")
	}

	// Check that timestamps are recent (within last minute)
	now := time.Now()
	if project.CreatedAt.Before(now.Add(-time.Minute)) {
		t.Error("CreatedAt should be recent")
	}
}
