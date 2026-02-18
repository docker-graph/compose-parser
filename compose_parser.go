// Package compose_parser provides a Go library for parsing Docker Compose files.
// Version: v0.1.0
// Author: Docker Graph Team
// License: MIT
package compose_parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ComposeParser представляет парсер Docker Compose файлов
type ComposeParser struct{}

// NewComposeParser создает новый парсер Docker Compose файлов
func NewComposeParser() *ComposeParser {
	return &ComposeParser{}
}

// ParseFile парсит Docker Compose файл и возвращает конфигурацию проекта
func (p *ComposeParser) ParseFile(filePath string) (*ComposeProjectConfig, error) {
	return p.ParseFileWithName(filePath, "")
}

func (p *ComposeParser) ParseFileWithName(filePath string, projectName string) (*ComposeProjectConfig, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".yaml" && ext != ".yml" {
		return nil, fmt.Errorf("unsupported file extension: %s, expected .yaml or .yml", ext)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	if projectName == "" {
		baseName := filepath.Base(filePath)
		projectName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	}

	return p.parseYAML(data, projectName)
}

// ParseYAML парсит YAML данные и возвращает конфигурацию проекта
// Если имя проекта не указано, используется "docker-compose"
func (p *ComposeParser) ParseYAML(data []byte) (*ComposeProjectConfig, error) {
	return p.parseYAML(data, "docker-compose")
}

// ParseYAMLWithName парсит YAML данные с указанным именем проекта
func (p *ComposeParser) ParseYAMLWithName(data []byte, projectName string) (*ComposeProjectConfig, error) {
	return p.parseYAML(data, projectName)
}

// parseYAML парсит YAML данные и возвращает конфигурацию проекта
func (p *ComposeParser) parseYAML(data []byte, projectName string) (*ComposeProjectConfig, error) {
	// Парсим YAML с сохранением порядка ключей
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	// Создаем конфигурацию проекта
	now := time.Now()
	project := &ComposeProjectConfig{
		Name:         projectName,
		Services:     make(map[string]*ComposeServiceConfig),
		ServiceOrder: make([]string, 0),
		VolumeOrder:  make([]string, 0),
		Networks:     make(map[string]*NetworkConfig),
		Volumes:      make(map[string]*VolumeConfig),
		Secrets:      make(map[string]*SecretConfig),
		Configs:      make(map[string]*ConfigConfig),
		CreatedAt:    now,
		UpdatedAt:    now,
		Status:       "parsed",
	}

	// Извлекаем данные из корневого узла
	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 {
		return nil, fmt.Errorf("invalid YAML document")
	}

	rootNode := node.Content[0]
	if rootNode.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("root node is not a mapping")
	}

	// Обрабатываем все ключи в корневом узле
	for i := 0; i < len(rootNode.Content); i += 2 {
		keyNode := rootNode.Content[i]
		valueNode := rootNode.Content[i+1]

		if keyNode.Kind != yaml.ScalarNode {
			continue
		}

		key := keyNode.Value

		switch key {
		case "version":
			if valueNode.Kind == yaml.ScalarNode {
				project.Version = valueNode.Value
			}

		case "services":
			if valueNode.Kind == yaml.MappingNode {
				// Сохраняем порядок сервисов
				for j := 0; j < len(valueNode.Content); j += 2 {
					serviceKeyNode := valueNode.Content[j]
					serviceValueNode := valueNode.Content[j+1]

					if serviceKeyNode.Kind != yaml.ScalarNode {
						continue
					}

					serviceName := serviceKeyNode.Value

					// Добавляем имя сервиса в порядок
					project.ServiceOrder = append(project.ServiceOrder, serviceName)

					// Парсим сервис
					var serviceRaw interface{}
					if err := serviceValueNode.Decode(&serviceRaw); err != nil {
						return nil, fmt.Errorf("failed to decode service %s: %v", serviceName, err)
					}

					service, err := p.parseService(serviceName, serviceRaw)
					if err != nil {
						return nil, fmt.Errorf("failed to parse service %s: %v", serviceName, err)
					}

					// Устанавливаем порядковый номер
					service.Order = len(project.ServiceOrder)

					project.Services[serviceName] = service
				}
			}

		case "networks":
			if valueNode.Kind == yaml.MappingNode {
				for j := 0; j < len(valueNode.Content); j += 2 {
					networkKeyNode := valueNode.Content[j]
					networkValueNode := valueNode.Content[j+1]

					if networkKeyNode.Kind != yaml.ScalarNode {
						continue
					}

					networkName := networkKeyNode.Value

					var networkRaw interface{}
					if err := networkValueNode.Decode(&networkRaw); err != nil {
						return nil, fmt.Errorf("failed to decode network %s: %v", networkName, err)
					}

					network, err := p.parseNetwork(networkName, networkRaw)
					if err != nil {
						return nil, fmt.Errorf("failed to parse network %s: %v", networkName, err)
					}
					project.Networks[networkName] = network
				}
			}

		case "volumes":
			if valueNode.Kind == yaml.MappingNode {
				// Сохраняем порядок томов
				for j := 0; j < len(valueNode.Content); j += 2 {
					volumeKeyNode := valueNode.Content[j]
					volumeValueNode := valueNode.Content[j+1]

					if volumeKeyNode.Kind != yaml.ScalarNode {
						continue
					}

					volumeName := volumeKeyNode.Value

					// Добавляем имя тома в порядок
					project.VolumeOrder = append(project.VolumeOrder, volumeName)

					var volumeRaw interface{}
					if err := volumeValueNode.Decode(&volumeRaw); err != nil {
						return nil, fmt.Errorf("failed to decode volume %s: %v", volumeName, err)
					}

					volume, err := p.parseVolume(volumeName, volumeRaw)
					if err != nil {
						return nil, fmt.Errorf("failed to parse volume %s: %v", volumeName, err)
					}

					// Устанавливаем порядковый номер
					volume.Order = len(project.VolumeOrder)

					project.Volumes[volumeName] = volume
				}
			}

		case "secrets":
			if valueNode.Kind == yaml.MappingNode {
				for j := 0; j < len(valueNode.Content); j += 2 {
					secretKeyNode := valueNode.Content[j]
					secretValueNode := valueNode.Content[j+1]

					if secretKeyNode.Kind != yaml.ScalarNode {
						continue
					}

					secretName := secretKeyNode.Value

					var secretRaw interface{}
					if err := secretValueNode.Decode(&secretRaw); err != nil {
						return nil, fmt.Errorf("failed to decode secret %s: %v", secretName, err)
					}

					secret, err := p.parseSecret(secretName, secretRaw)
					if err != nil {
						return nil, fmt.Errorf("failed to parse secret %s: %v", secretName, err)
					}
					project.Secrets[secretName] = secret
				}
			}

		case "configs":
			if valueNode.Kind == yaml.MappingNode {
				for j := 0; j < len(valueNode.Content); j += 2 {
					configKeyNode := valueNode.Content[j]
					configValueNode := valueNode.Content[j+1]

					if configKeyNode.Kind != yaml.ScalarNode {
						continue
					}

					configName := configKeyNode.Value

					var configRaw interface{}
					if err := configValueNode.Decode(&configRaw); err != nil {
						return nil, fmt.Errorf("failed to decode config %s: %v", configName, err)
					}

					config, err := p.parseConfig(configName, configRaw)
					if err != nil {
						return nil, fmt.Errorf("failed to parse config %s: %v", configName, err)
					}
					project.Configs[configName] = config
				}
			}
		}
	}

	return project, nil
}

// parseService парсит конфигурацию сервиса
func (p *ComposeParser) parseService(name string, raw interface{}) (*ComposeServiceConfig, error) {
	service := &ComposeServiceConfig{
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    "parsed",
	}

	serviceMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("service configuration must be a map")
	}

	// Базовые поля
	if image, ok := serviceMap["image"].(string); ok {
		service.Image = image
	}

	if buildRaw, ok := serviceMap["build"]; ok {
		build, err := p.parseBuild(buildRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse build config: %v", err)
		}
		service.Build = build
	}

	if commandRaw, ok := serviceMap["command"]; ok {
		service.Command = p.parseStringOrSlice(commandRaw)
	}

	if entrypointRaw, ok := serviceMap["entrypoint"]; ok {
		service.Entrypoint = p.parseStringOrSlice(entrypointRaw)
	}

	if workingDir, ok := serviceMap["working_dir"].(string); ok {
		service.WorkingDir = workingDir
	}

	if user, ok := serviceMap["user"].(string); ok {
		service.User = user
	}

	if platform, ok := serviceMap["platform"].(string); ok {
		service.Platform = platform
	}

	// Зависимости и перезапуск
	if dependsOnRaw, ok := serviceMap["depends_on"]; ok {
		service.DependsOn = p.parseStringOrSlice(dependsOnRaw)
	}

	if restart, ok := serviceMap["restart"].(string); ok {
		service.Restart = restart
	}

	// Сеть и порты
	if portsRaw, ok := serviceMap["ports"]; ok {
		ports, err := p.parsePorts(portsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ports: %v", err)
		}
		service.Ports = ports
	}

	if exposeRaw, ok := serviceMap["expose"]; ok {
		service.Expose = p.parseStringOrSlice(exposeRaw)
	}

	if networksRaw, ok := serviceMap["networks"]; ok {
		service.Networks = p.parseStringOrSlice(networksRaw)
	}

	if networkMode, ok := serviceMap["network_mode"].(string); ok {
		service.NetworkMode = networkMode
	}

	// Переменные окружения
	if envRaw, ok := serviceMap["environment"]; ok {
		env, err := p.parseEnvironment(envRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse environment: %v", err)
		}
		service.Environment = env
	}

	if envFileRaw, ok := serviceMap["env_file"]; ok {
		service.EnvFile = p.parseStringOrSlice(envFileRaw)
	}

	// Тома
	if volumesRaw, ok := serviceMap["volumes"]; ok {
		volumes, err := p.parseVolumeMounts(volumesRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse volumes: %v", err)
		}
		service.Volumes = volumes
	}

	if volumesFromRaw, ok := serviceMap["volumes_from"]; ok {
		service.VolumesFrom = p.parseStringOrSlice(volumesFromRaw)
	}

	// Ресурсы
	if deployRaw, ok := serviceMap["deploy"]; ok {
		deploy, err := p.parseDeploy(deployRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse deploy config: %v", err)
		}
		service.Deploy = deploy
	}

	if cpuShares, ok := serviceMap["cpu_shares"].(int64); ok {
		service.CPUShares = cpuShares
	}

	if cpuset, ok := serviceMap["cpuset"].(string); ok {
		service.CPUSet = cpuset
	}

	if cpuQuota, ok := serviceMap["cpu_quota"].(int64); ok {
		service.CPUQuota = cpuQuota
	}

	if cpus, ok := serviceMap["cpus"].(float64); ok {
		service.CPUs = cpus
	}

	if memory, ok := serviceMap["memory"].(string); ok {
		service.Memory = memory
	}

	if memorySwap, ok := serviceMap["memory_swap"].(string); ok {
		service.MemorySwap = memorySwap
	}

	// Логирование
	if loggingRaw, ok := serviceMap["logging"]; ok {
		logging, err := p.parseLogging(loggingRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse logging config: %v", err)
		}
		service.Logging = logging
	}

	// Здоровье
	if healthcheckRaw, ok := serviceMap["healthcheck"]; ok {
		healthcheck, err := p.parseHealthcheck(healthcheckRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse healthcheck config: %v", err)
		}
		service.HealthCheck = healthcheck
	}

	// Метки
	if labelsRaw, ok := serviceMap["labels"]; ok {
		labels, err := p.parseLabels(labelsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse labels: %v", err)
		}
		service.Labels = labels
	}

	// Расширения
	if extendsRaw, ok := serviceMap["extends"]; ok {
		extends, err := p.parseExtends(extendsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse extends config: %v", err)
		}
		service.Extends = extends
	}

	return service, nil
}

// parseBuild парсит конфигурацию сборки
func (p *ComposeParser) parseBuild(raw interface{}) (*BuildConfig, error) {
	build := &BuildConfig{}

	switch v := raw.(type) {
	case string:
		build.Context = v
	case map[string]interface{}:
		if context, ok := v["context"].(string); ok {
			build.Context = context
		}
		if dockerfile, ok := v["dockerfile"].(string); ok {
			build.Dockerfile = dockerfile
		}
		if argsRaw, ok := v["args"]; ok {
			args, err := p.parseLabels(argsRaw)
			if err != nil {
				return nil, fmt.Errorf("failed to parse build args: %v", err)
			}
			build.Args = args
		}
		if target, ok := v["target"].(string); ok {
			build.Target = target
		}
		if cacheFromRaw, ok := v["cache_from"]; ok {
			build.CacheFrom = p.parseStringOrSlice(cacheFromRaw)
		}
		if labelsRaw, ok := v["labels"]; ok {
			labels, err := p.parseLabels(labelsRaw)
			if err != nil {
				return nil, fmt.Errorf("failed to parse build labels: %v", err)
			}
			build.Labels = labels
		}
	default:
		return nil, fmt.Errorf("invalid build configuration type: %T", raw)
	}

	return build, nil
}

// parsePorts парсит маппинг портов
func (p *ComposeParser) parsePorts(raw interface{}) ([]PortMapping, error) {
	var ports []PortMapping

	switch v := raw.(type) {
	case []interface{}:
		for _, portRaw := range v {
			port, err := p.parsePort(portRaw)
			if err != nil {
				return nil, err
			}
			ports = append(ports, *port)
		}
	default:
		return nil, fmt.Errorf("invalid ports configuration type: %T", raw)
	}

	return ports, nil
}

// parsePort парсит один порт
func (p *ComposeParser) parsePort(raw interface{}) (*PortMapping, error) {
	port := &PortMapping{}

	switch v := raw.(type) {
	case string:
		// Парсим строку формата "8080:80" или "8080:80/tcp"
		parts := strings.Split(v, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid port format: %s", v)
		}

		// Парсим опубликованный порт
		publishedParts := strings.Split(parts[0], "/")
		if len(publishedParts) > 0 {
			if portNum, err := p.parsePortNumber(publishedParts[0]); err == nil {
				port.Published = uint16(portNum)
			}
		}

		// Парсим целевой порт и протокол
		targetParts := strings.Split(parts[1], "/")
		if len(targetParts) > 0 {
			if portNum, err := p.parsePortNumber(targetParts[0]); err == nil {
				port.Target = uint16(portNum)
			}
		}
		if len(targetParts) > 1 {
			port.Protocol = targetParts[1]
		}

	case map[string]interface{}:
		if target, ok := v["target"].(int); ok {
			port.Target = uint16(target)
		}
		if published, ok := v["published"].(int); ok {
			port.Published = uint16(published)
		}
		if protocol, ok := v["protocol"].(string); ok {
			port.Protocol = protocol
		}
		if mode, ok := v["mode"].(string); ok {
			port.Mode = mode
		}
	default:
		return nil, fmt.Errorf("invalid port configuration type: %T", raw)
	}

	return port, nil
}

// parsePortNumber парсит номер порта из строки
func (p *ComposeParser) parsePortNumber(s string) (int, error) {
	var port int
	_, err := fmt.Sscanf(s, "%d", &port)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", s)
	}
	return port, nil
}

// parseEnvironment парсит переменные окружения
func (p *ComposeParser) parseEnvironment(raw interface{}) (map[string]string, error) {
	env := make(map[string]string)

	switch v := raw.(type) {
	case []interface{}:
		for _, envRaw := range v {
			if envStr, ok := envRaw.(string); ok {
				parts := strings.SplitN(envStr, "=", 2)
				if len(parts) == 2 {
					env[parts[0]] = parts[1]
				} else {
					env[parts[0]] = ""
				}
			}
		}
	case map[string]interface{}:
		for key, value := range v {
			if strValue, ok := value.(string); ok {
				env[key] = strValue
			} else {
				env[key] = fmt.Sprintf("%v", value)
			}
		}
	default:
		return nil, fmt.Errorf("invalid environment configuration type: %T", raw)
	}

	return env, nil
}

// parseVolumeMounts парсит монтирования томов
func (p *ComposeParser) parseVolumeMounts(raw interface{}) ([]VolumeMount, error) {
	var volumes []VolumeMount

	switch v := raw.(type) {
	case []interface{}:
		for _, volumeRaw := range v {
			volume, err := p.parseVolumeMount(volumeRaw)
			if err != nil {
				return nil, err
			}
			volumes = append(volumes, *volume)
		}
	default:
		return nil, fmt.Errorf("invalid volumes configuration type: %T", raw)
	}

	return volumes, nil
}

// parseVolumeMount парсит одно монтирование тома
func (p *ComposeParser) parseVolumeMount(raw interface{}) (*VolumeMount, error) {
	volume := &VolumeMount{
		Type: "volume", // значение по умолчанию
	}

	switch v := raw.(type) {
	case string:
		// Парсим строку формата "/host/path:/container/path:ro"
		parts := strings.Split(v, ":")
		if len(parts) < 2 || len(parts) > 3 {
			return nil, fmt.Errorf("invalid volume format: %s", v)
		}

		// Определяем тип
		if strings.HasPrefix(parts[0], "/") || strings.HasPrefix(parts[0], ".") {
			volume.Type = "bind"
			volume.Source = parts[0]
		} else {
			volume.Type = "volume"
			volume.Source = parts[0]
		}

		volume.Target = parts[1]

		if len(parts) == 3 {
			options := parts[2]
			if options == "ro" {
				volume.ReadOnly = true
			} else {
				volume.Consistency = options
			}
		}

	case map[string]interface{}:
		if typ, ok := v["type"].(string); ok {
			volume.Type = typ
		}
		if source, ok := v["source"].(string); ok {
			volume.Source = source
		}
		if target, ok := v["target"].(string); ok {
			volume.Target = target
		}
		if readOnly, ok := v["read_only"].(bool); ok {
			volume.ReadOnly = readOnly
		}
		if consistency, ok := v["consistency"].(string); ok {
			volume.Consistency = consistency
		}
	default:
		return nil, fmt.Errorf("invalid volume configuration type: %T", raw)
	}

	return volume, nil
}

// parseDeploy парсит конфигурацию развертывания
func (p *ComposeParser) parseDeploy(raw interface{}) (*DeployConfig, error) {
	deploy := &DeployConfig{}

	deployMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("deploy configuration must be a map")
	}

	if mode, ok := deployMap["mode"].(string); ok {
		deploy.Mode = mode
	}

	if replicas, ok := deployMap["replicas"].(int); ok {
		deploy.Replicas = uint64(replicas)
	}

	if placementRaw, ok := deployMap["placement"]; ok {
		placement, err := p.parsePlacement(placementRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse placement: %v", err)
		}
		deploy.Placement = placement
	}

	if resourcesRaw, ok := deployMap["resources"]; ok {
		resources, err := p.parseResources(resourcesRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse resources: %v", err)
		}
		deploy.Resources = resources
	}

	if restartPolicyRaw, ok := deployMap["restart_policy"]; ok {
		restartPolicy, err := p.parseRestartPolicy(restartPolicyRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse restart policy: %v", err)
		}
		deploy.RestartPolicy = restartPolicy
	}

	if updateConfigRaw, ok := deployMap["update_config"]; ok {
		updateConfig, err := p.parseUpdateConfig(updateConfigRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse update config: %v", err)
		}
		deploy.UpdateConfig = updateConfig
	}

	if rollbackConfigRaw, ok := deployMap["rollback_config"]; ok {
		rollbackConfig, err := p.parseRollbackConfig(rollbackConfigRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rollback config: %v", err)
		}
		deploy.RollbackConfig = rollbackConfig
	}

	return deploy, nil
}

// parsePlacement парсит конфигурацию размещения
func (p *ComposeParser) parsePlacement(raw interface{}) (*PlacementConfig, error) {
	placement := &PlacementConfig{}

	placementMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("placement configuration must be a map")
	}

	if constraintsRaw, ok := placementMap["constraints"]; ok {
		placement.Constraints = p.parseStringOrSlice(constraintsRaw)
	}

	if preferencesRaw, ok := placementMap["preferences"]; ok {
		placement.Preferences = p.parseStringOrSlice(preferencesRaw)
	}

	if maxReplicas, ok := placementMap["max_replicas"].(int); ok {
		placement.MaxReplicas = uint64(maxReplicas)
	}

	return placement, nil
}

// parseResources парсит требования к ресурсам
func (p *ComposeParser) parseResources(raw interface{}) (*ResourceRequirements, error) {
	resources := &ResourceRequirements{}

	resourcesMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resources configuration must be a map")
	}

	if limitsRaw, ok := resourcesMap["limits"]; ok {
		limits, err := p.parseResourceLimits(limitsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse limits: %v", err)
		}
		resources.Limits = limits
	}

	if reservationsRaw, ok := resourcesMap["reservations"]; ok {
		reservations, err := p.parseResourceLimits(reservationsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse reservations: %v", err)
		}
		resources.Reservations = reservations
	}

	return resources, nil
}

// parseResourceLimits парсит лимиты ресурсов
func (p *ComposeParser) parseResourceLimits(raw interface{}) (*ResourceLimits, error) {
	limits := &ResourceLimits{}

	limitsMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resource limits configuration must be a map")
	}

	if cpus, ok := limitsMap["cpus"].(string); ok {
		limits.CPUs = cpus
	}

	if memory, ok := limitsMap["memory"].(string); ok {
		limits.Memory = memory
	}

	if pids, ok := limitsMap["pids"].(int64); ok {
		limits.Pids = pids
	}

	return limits, nil
}

// parseRestartPolicy парсит политику перезапуска
func (p *ComposeParser) parseRestartPolicy(raw interface{}) (*RestartPolicyConfig, error) {
	policy := &RestartPolicyConfig{}

	policyMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("restart policy configuration must be a map")
	}

	if condition, ok := policyMap["condition"].(string); ok {
		policy.Condition = condition
	}

	if delay, ok := policyMap["delay"].(string); ok {
		policy.Delay = delay
	}

	if maxAttempts, ok := policyMap["max_attempts"].(int); ok {
		policy.MaxAttempts = uint64(maxAttempts)
	}

	if window, ok := policyMap["window"].(string); ok {
		policy.Window = window
	}

	return policy, nil
}

// parseUpdateConfig парсит конфигурацию обновления
func (p *ComposeParser) parseUpdateConfig(raw interface{}) (*UpdateConfig, error) {
	config := &UpdateConfig{}

	configMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("update configuration must be a map")
	}

	if parallelism, ok := configMap["parallelism"].(int); ok {
		config.Parallelism = uint64(parallelism)
	}

	if delay, ok := configMap["delay"].(string); ok {
		config.Delay = delay
	}

	if failureAction, ok := configMap["failure_action"].(string); ok {
		config.FailureAction = failureAction
	}

	if monitor, ok := configMap["monitor"].(string); ok {
		config.Monitor = monitor
	}

	if maxFailureRatio, ok := configMap["max_failure_ratio"].(string); ok {
		config.MaxFailureRatio = maxFailureRatio
	}

	if order, ok := configMap["order"].(string); ok {
		config.Order = order
	}

	return config, nil
}

// parseRollbackConfig парсит конфигурацию отката
func (p *ComposeParser) parseRollbackConfig(raw interface{}) (*RollbackConfig, error) {
	config := &RollbackConfig{}

	configMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("rollback configuration must be a map")
	}

	if parallelism, ok := configMap["parallelism"].(int); ok {
		config.Parallelism = uint64(parallelism)
	}

	if delay, ok := configMap["delay"].(string); ok {
		config.Delay = delay
	}

	if failureAction, ok := configMap["failure_action"].(string); ok {
		config.FailureAction = failureAction
	}

	if monitor, ok := configMap["monitor"].(string); ok {
		config.Monitor = monitor
	}

	if maxFailureRatio, ok := configMap["max_failure_ratio"].(string); ok {
		config.MaxFailureRatio = maxFailureRatio
	}

	if order, ok := configMap["order"].(string); ok {
		config.Order = order
	}

	return config, nil
}

// parseLogging парсит конфигурацию логирования
func (p *ComposeParser) parseLogging(raw interface{}) (*LoggingConfig, error) {
	logging := &LoggingConfig{}

	loggingMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("logging configuration must be a map")
	}

	if driver, ok := loggingMap["driver"].(string); ok {
		logging.Driver = driver
	}

	if optionsRaw, ok := loggingMap["options"]; ok {
		options, err := p.parseLabels(optionsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse logging options: %v", err)
		}
		logging.Options = options
	}

	return logging, nil
}

// parseHealthcheck парсит конфигурацию проверки здоровья
func (p *ComposeParser) parseHealthcheck(raw interface{}) (*HealthCheckConfig, error) {
	healthcheck := &HealthCheckConfig{}

	healthcheckMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("healthcheck configuration must be a map")
	}

	if testRaw, ok := healthcheckMap["test"]; ok {
		healthcheck.Test = p.parseStringOrSlice(testRaw)
	}

	if interval, ok := healthcheckMap["interval"].(string); ok {
		healthcheck.Interval = interval
	}

	if timeout, ok := healthcheckMap["timeout"].(string); ok {
		healthcheck.Timeout = timeout
	}

	if retries, ok := healthcheckMap["retries"].(int); ok {
		healthcheck.Retries = uint64(retries)
	}

	if startPeriod, ok := healthcheckMap["start_period"].(string); ok {
		healthcheck.StartPeriod = startPeriod
	}

	if startInterval, ok := healthcheckMap["start_interval"].(string); ok {
		healthcheck.StartInterval = startInterval
	}

	return healthcheck, nil
}

// parseLabels парсит метки
func (p *ComposeParser) parseLabels(raw interface{}) (map[string]string, error) {
	labels := make(map[string]string)

	switch v := raw.(type) {
	case []interface{}:
		for _, labelRaw := range v {
			if labelStr, ok := labelRaw.(string); ok {
				parts := strings.SplitN(labelStr, "=", 2)
				if len(parts) == 2 {
					labels[parts[0]] = parts[1]
				}
			}
		}
	case map[string]interface{}:
		for key, value := range v {
			if strValue, ok := value.(string); ok {
				labels[key] = strValue
			} else {
				labels[key] = fmt.Sprintf("%v", value)
			}
		}
	default:
		return nil, fmt.Errorf("invalid labels configuration type: %T", raw)
	}

	return labels, nil
}

// parseExtends парсит конфигурацию расширения
func (p *ComposeParser) parseExtends(raw interface{}) (*ExtendsConfig, error) {
	extends := &ExtendsConfig{}

	switch v := raw.(type) {
	case string:
		extends.Service = v
	case map[string]interface{}:
		if file, ok := v["file"].(string); ok {
			extends.File = file
		}
		if service, ok := v["service"].(string); ok {
			extends.Service = service
		}
	default:
		return nil, fmt.Errorf("invalid extends configuration type: %T", raw)
	}

	return extends, nil
}

// parseNetwork парсит конфигурацию сети
func (p *ComposeParser) parseNetwork(name string, raw interface{}) (*NetworkConfig, error) {
	network := &NetworkConfig{}

	networkMap, ok := raw.(map[string]interface{})
	if !ok {
		// Если это просто строка или булево значение (например, external: true)
		if external, ok := raw.(bool); ok {
			network.External = external
		}
		return network, nil
	}

	if driver, ok := networkMap["driver"].(string); ok {
		network.Driver = driver
	}

	if driverOptsRaw, ok := networkMap["driver_opts"]; ok {
		driverOpts, err := p.parseLabels(driverOptsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse driver options: %v", err)
		}
		network.DriverOpts = driverOpts
	}

	if external, ok := networkMap["external"].(bool); ok {
		network.External = external
	}

	if networkName, ok := networkMap["name"].(string); ok {
		network.Name = networkName
	}

	if attachable, ok := networkMap["attachable"].(bool); ok {
		network.Attachable = attachable
	}

	if internal, ok := networkMap["internal"].(bool); ok {
		network.Internal = internal
	}

	if labelsRaw, ok := networkMap["labels"]; ok {
		labels, err := p.parseLabels(labelsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse network labels: %v", err)
		}
		network.Labels = labels
	}

	return network, nil
}

// parseVolume парсит конфигурацию тома
func (p *ComposeParser) parseVolume(name string, raw interface{}) (*VolumeConfig, error) {
	volume := &VolumeConfig{}

	volumeMap, ok := raw.(map[string]interface{})
	if !ok {
		// Если это просто строка или булево значение (например, external: true)
		if external, ok := raw.(bool); ok {
			volume.External = external
		}
		return volume, nil
	}

	if driver, ok := volumeMap["driver"].(string); ok {
		volume.Driver = driver
	}

	if driverOptsRaw, ok := volumeMap["driver_opts"]; ok {
		driverOpts, err := p.parseLabels(driverOptsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse driver options: %v", err)
		}
		volume.DriverOpts = driverOpts
	}

	if external, ok := volumeMap["external"].(bool); ok {
		volume.External = external
	}

	if volumeName, ok := volumeMap["name"].(string); ok {
		volume.Name = volumeName
	}

	if labelsRaw, ok := volumeMap["labels"]; ok {
		labels, err := p.parseLabels(labelsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse volume labels: %v", err)
		}
		volume.Labels = labels
	}

	return volume, nil
}

// parseSecret парсит конфигурацию секрета
func (p *ComposeParser) parseSecret(name string, raw interface{}) (*SecretConfig, error) {
	secret := &SecretConfig{}

	secretMap, ok := raw.(map[string]interface{})
	if !ok {
		// Если это просто строка или булево значение
		if external, ok := raw.(bool); ok {
			secret.External = external
		} else if file, ok := raw.(string); ok {
			secret.File = file
		}
		return secret, nil
	}

	if file, ok := secretMap["file"].(string); ok {
		secret.File = file
	}

	if external, ok := secretMap["external"].(bool); ok {
		secret.External = external
	}

	if secretName, ok := secretMap["name"].(string); ok {
		secret.Name = secretName
	}

	if labelsRaw, ok := secretMap["labels"]; ok {
		labels, err := p.parseLabels(labelsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse secret labels: %v", err)
		}
		secret.Labels = labels
	}

	return secret, nil
}

// parseConfig парсит конфигурацию конфигурации
func (p *ComposeParser) parseConfig(name string, raw interface{}) (*ConfigConfig, error) {
	config := &ConfigConfig{}

	configMap, ok := raw.(map[string]interface{})
	if !ok {
		// Если это просто строка или булево значение
		if external, ok := raw.(bool); ok {
			config.External = external
		} else if file, ok := raw.(string); ok {
			config.File = file
		}
		return config, nil
	}

	if file, ok := configMap["file"].(string); ok {
		config.File = file
	}

	if external, ok := configMap["external"].(bool); ok {
		config.External = external
	}

	if configName, ok := configMap["name"].(string); ok {
		config.Name = configName
	}

	if labelsRaw, ok := configMap["labels"]; ok {
		labels, err := p.parseLabels(labelsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config labels: %v", err)
		}
		config.Labels = labels
	}

	return config, nil
}

// parseStringOrSlice парсит строку или срез строк
func (p *ComposeParser) parseStringOrSlice(raw interface{}) []string {
	switch v := raw.(type) {
	case string:
		return []string{v}
	case []interface{}:
		var result []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return nil
	}
}

// ParseReader парсит Docker Compose файл из io.Reader
func (p *ComposeParser) ParseReader(reader io.Reader) (*ComposeProjectConfig, error) {
	return p.ParseReaderWithName(reader, "docker-compose")
}

// ParseReaderWithName парсит Docker Compose файл из io.Reader с указанным именем проекта
func (p *ComposeParser) ParseReaderWithName(reader io.Reader, projectName string) (*ComposeProjectConfig, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %v", err)
	}

	return p.parseYAML(data, projectName)
}

// ParseFromDirectory ищет и парсит Docker Compose файлы в директории
func (p *ComposeParser) ParseFromDirectory(dirPath string) ([]*ComposeProjectConfig, error) {
	var projects []*ComposeProjectConfig

	// Ищем файлы docker-compose
	patterns := []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}

	for _, pattern := range patterns {
		filePath := filepath.Join(dirPath, pattern)
		if _, err := os.Stat(filePath); err == nil {
			project, err := p.ParseFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse file %s: %v", filePath, err)
			}
			projects = append(projects, project)
		}
	}

	return projects, nil
}

// ParseToReactFlow генерирует структурированный React Flow граф из Docker Compose проекта
func (p *ComposeParser) ParseToReactFlow(project ComposeProjectConfig, options *GraphLayoutOptions) (*ReactFlowGraph, error) {

	if options == nil {
		options = &GraphLayoutOptions{
			Direction:          "LR",
			NodeWidth:          240,
			NodeHeight:         120,
			NodeGapX:           100,
			NodeGapY:           50,
			Padding:            50,
			ColumnGap:          440,
			ColumnTopGap:       120,
			DockerComposeStart: -350,
			VolumeXOffset:      300,
			VolumeYOffset:      180,
			LastY:              -1000,
		}
	}

	// Генерируем узлы и связи
	nodes := make([]ReactFlowNode, 0)
	edges := make([]ReactFlowEdge, 0)
	edgeCounter := 0

	// 1. Рассчитываем размеры и позиции
	serviceCount := len(project.Services)
	volumeCount := len(project.Volumes)
	networkCount := len(project.Networks)

	// 1. Создаем DockerCompose ноду (самая левая)
	dockerComposeX := 0
	dockerComposeY := (serviceCount*options.NodeHeight)/2 - 10

	dockerComposeNode := ReactFlowNode{
		ID:   "docker-compose",
		Type: "compose",
		Position: ReactFlowPosition{
			X: float64(options.DockerComposeStart),
			Y: float64(dockerComposeY),
		},
		Data: ReactFlowNodeData{
			Label: project.Name,
			Type:  "compose",
			Properties: map[string]interface{}{
				"services": serviceCount,
				"networks": len(project.Networks),
				"volumes":  volumeCount,
				"version":  project.Version,
			},
		},
	}
	nodes = append(nodes, dockerComposeNode)

	// 2. Создаем Network ноды (вторая колонка, между DockerCompose и Services)
	// Выравниваем сети из центра относительно высоты всех сервисов
	networkStartX := dockerComposeX + 20 // Позиция между DockerCompose и Services

	// Рассчитываем высоту, занимаемую всеми сервисами
	serviceHeight := serviceCount * options.NodeHeight
	if serviceHeight < options.NodeWidth {
		serviceHeight = options.NodeWidth
	}

	// Рассчитываем высоту, занимаемую всеми сетями
	networkHeight := networkCount * options.NodeHeight // 120px отступ между сетями
	if networkHeight < options.NodeHeight {
		networkHeight = options.NodeHeight // Минимальная высота для сетей
	}

	// Центрируем сети относительно высоты сервисов
	// Если сервисов больше, чем сетей, центрируем сети в середине высоты сервисов
	// Если сетей больше, чем сервисов, центрируем сервисы в середине высоты сетей
	// Используем options.Padding как базовую позицию сервисов (serviceStartY будет равен этому значению)
	serviceBaseY := options.Padding
	var networkStartY int
	if serviceHeight >= networkHeight {
		networkStartY = serviceBaseY + (serviceHeight-networkHeight)/2
	} else {
		networkStartY = serviceBaseY - (networkHeight-serviceHeight)/2
	}

	networksList := make([]networkWithName, 0, len(project.Networks))
	for networkName, network := range project.Networks {
		networksList = append(networksList, networkWithName{
			name:    networkName,
			network: network,
		})
	}

	sort.Slice(networksList, func(i, j int) bool {
		return networksList[i].name < networksList[j].name
	})

	networkIndex := 0
	networkNodes := make(map[string]string)

	for _, item := range networksList {
		networkName := item.name
		network := item.network

		// Позиционируем сети вертикально во второй колонке
		x := networkStartX
		y := networkStartY + networkIndex*120 // 120px отступ между сетями

		nodeID := fmt.Sprintf("network-%s", networkName)
		networkNodes[networkName] = nodeID

		networkNode := ReactFlowNode{
			ID:   nodeID,
			Type: "network",
			Position: ReactFlowPosition{
				X: float64(x),
				Y: float64(y),
			},
			Data: ReactFlowNodeData{
				Label:   networkName,
				Type:    "network",
				Network: network,
				Properties: map[string]interface{}{
					"driver":     network.Driver,
					"internal":   network.Internal,
					"external":   network.External,
					"attachable": network.Attachable,
				},
			},
		}
		nodes = append(nodes, networkNode)

		// Создаем связь от DockerCompose к сети
		edgeCounter++
		edge := ReactFlowEdge{
			ID:     fmt.Sprintf("edge-compose-network-%s", networkName),
			Source: "docker-compose",
			Target: nodeID,
			Type:   "step",
		}
		edges = append(edges, edge)

		networkIndex++
	}

	// 3. Создаем Service ноды (центральная колонка)
	serviceStartX := dockerComposeX + options.ColumnGap
	serviceStartY := options.Padding

	servicesList := make([]serviceWithOrder, 0, len(project.Services))
	for serviceName, service := range project.Services {
		servicesList = append(servicesList, serviceWithOrder{
			name:    serviceName,
			order:   service.Order,
			service: service,
		})
	}

	// Сортируем сервисы по порядку (Order)
	sort.Slice(servicesList, func(i, j int) bool {
		if servicesList[i].order == 0 && servicesList[j].order == 0 {
			idxI := -1
			idxJ := -1
			for k, name := range project.ServiceOrder {
				if name == servicesList[i].name {
					idxI = k
				}
				if name == servicesList[j].name {
					idxJ = k
				}
				if idxI != -1 && idxJ != -1 {
					break
				}
			}

			if idxI != -1 && idxJ != -1 {
				return idxI < idxJ
			}

			if idxI == -1 {
				return false
			}
			if idxJ == -1 {
				return true
			}
		}
		// Сортируем по полю Order
		return servicesList[i].order < servicesList[j].order
	})

	serviceIndex := 0
	for _, item := range servicesList {
		serviceName := item.name
		service := item.service

		// Позиционируем сервисы вертикально в центральной колонке
		x := serviceStartX
		y := serviceStartY + serviceIndex*options.ColumnTopGap

		nodeID := fmt.Sprintf("service-%s", serviceName)

		nodeColor := "#3b82f6"

		serviceNode := ReactFlowNode{
			ID:   nodeID,
			Type: "service",
			Position: ReactFlowPosition{
				X: float64(x),
				Y: float64(y),
			},
			Data: ReactFlowNodeData{
				Label:   serviceName,
				Type:    "service",
				Service: service,
				Status:  "saved",
				Properties: map[string]interface{}{
					"image":      service.Image,
					"ports":      len(service.Ports),
					"volumes":    len(service.Volumes),
					"depends_on": len(service.DependsOn),
					"networks":   len(service.Networks),
					"color":      nodeColor,
					"order":      service.Order,
				},
			},
		}
		nodes = append(nodes, serviceNode)

		// 4. Создаем связи от сетей к сервису (если сервис использует сети)
		hasNetworkConnections := false
		for _, networkName := range service.Networks {
			if networkNodeID, exists := networkNodes[networkName]; exists {
				edgeCounter++
				// Получаем IP-адрес сервиса в этой сети
				// ipAddress := getServiceIPInNetwork(serviceName, networkName)
				edgeLabel := networkName
				labelStyle := map[string]interface{}{
					"fill":      "#3b82f6",
					"opacity":   .4,
					"textAlign": "center",
				}

				edge := ReactFlowEdge{
					ID:     fmt.Sprintf("edge-network-service-%s-%s", networkName, serviceName),
					Source: networkNodeID,
					Target: nodeID,
					Type:   "smoothstep",
					Style: map[string]interface{}{
						"strokeWidth": 0,
						"stroke":      "transparent", // Синий цвет для связей сетей
					},
					Label:      edgeLabel,
					LabelStyle: labelStyle,
				}
				edges = append(edges, edge)
				hasNetworkConnections = true
			}
		}

		// 5. Создаем связь от DockerCompose к сервису если:
		// - у сервиса нет сетей ИЛИ
		// - у сервиса есть сети, но они не определены в проекте (внешние сети)
		if !hasNetworkConnections {
			edgeCounter++
			edge := ReactFlowEdge{
				ID:     fmt.Sprintf("edge-compose-service-%d", edgeCounter),
				Source: "docker-compose",
				Target: nodeID,
				Type:   "smoothstep",
				Style: map[string]interface{}{
					"strokeWidth":     1.5,
					"strokeDasharray": "5,5",
				},
			}
			edges = append(edges, edge)
		}

		serviceIndex++
	}

	// 6. Собираем информацию об использовании томов сервисами
	volumeUsage := make(map[string][]string)
	volumeUsed := make(map[string]bool)
	volumeServiceX := make(map[string]int)
	volumeServiceY := make(map[string]int)
	serviceXPositions := make(map[string]int)
	serviceYPositions := make(map[string]int)

	servicePosIndex := 0
	for _, item := range servicesList {
		serviceName := item.name
		x := serviceStartX
		y := serviceStartY + servicePosIndex*options.ColumnTopGap
		serviceXPositions[serviceName] = x
		serviceYPositions[serviceName] = y
		servicePosIndex++
	}

	for _, item := range servicesList {
		serviceName := item.name
		service := item.service

		for _, volumeMount := range service.Volumes {
			if volumeMount.Type == "volume" && volumeMount.Source != "" {
				volumeName := volumeMount.Source
				volumeUsage[volumeName] = append(volumeUsage[volumeName], serviceName)
				volumeUsed[volumeName] = true
			}
		}
	}

	for volumeName, serviceNames := range volumeUsage {
		if len(serviceNames) > 0 {

			sumX := 0
			sumY := 0
			count := 0
			for _, serviceName := range serviceNames {
				if x, exists := serviceXPositions[serviceName]; exists {
					sumX += x
					count++
				}
				if y, exists := serviceYPositions[serviceName]; exists {
					sumY += y
				}
			}

			if count > 0 {
				volumeServiceX[volumeName] = sumX / count
				volumeServiceY[volumeName] = sumY / count
			}
		}
	}

	volumesList := make([]volumeWithOrder, 0, len(project.Volumes))
	for volumeName, volume := range project.Volumes {
		volumesList = append(volumesList, volumeWithOrder{
			name:   volumeName,
			order:  volume.Order,
			volume: volume,
		})
	}

	sort.Slice(volumesList, func(i, j int) bool {

		if volumesList[i].order == 0 && volumesList[j].order == 0 {

			idxI := -1
			idxJ := -1
			for k, name := range project.VolumeOrder {
				if name == volumesList[i].name {
					idxI = k
				}
				if name == volumesList[j].name {
					idxJ = k
				}
				if idxI != -1 && idxJ != -1 {
					break
				}
			}

			if idxI != -1 && idxJ != -1 {
				return idxI < idxJ
			}

			if idxI == -1 {
				return false
			}
			if idxJ == -1 {
				return true
			}
		}

		return volumesList[i].order < volumesList[j].order
	})

	volumeXOffset := options.VolumeXOffset
	unusedVolumeStartY := dockerComposeY + options.VolumeYOffset

	usedVolumes := make([]positionedVolume, 0)
	unusedVolumes := make([]positionedVolume, 0)

	for _, item := range volumesList {
		volumeName := item.name
		if volumeUsed[volumeName] {
			usedVolumes = append(usedVolumes, positionedVolume{
				name:     volumeName,
				volume:   item.volume,
				targetY:  volumeServiceY[volumeName],
				serviceX: volumeServiceX[volumeName],
				usedBy:   volumeUsage[volumeName],
			})
		} else {
			unusedVolumes = append(unusedVolumes, positionedVolume{
				name:   volumeName,
				volume: item.volume,
				usedBy: volumeUsage[volumeName],
			})
		}
	}

	sort.Slice(usedVolumes, func(i, j int) bool {
		return usedVolumes[i].targetY < usedVolumes[j].targetY
	})

	volumeStartY := options.Padding
	lastY := options.LastY

	for i, vol := range usedVolumes {
		var x, y int
		nodeID := fmt.Sprintf("volume-%s", vol.name)

		desiredY := vol.targetY

		if i == 0 {
			if desiredY < volumeStartY {
				y = volumeStartY
			} else {
				y = desiredY
			}
		} else {
			minY := lastY + options.ColumnTopGap
			if desiredY < minY {
				y = minY
			} else {
				y = desiredY
			}
		}

		lastY = y
		x = vol.serviceX + volumeXOffset

		volumeNode := ReactFlowNode{
			ID:   nodeID,
			Type: "volume",
			Position: ReactFlowPosition{
				X: float64(x),
				Y: float64(y),
			},
			Data: ReactFlowNodeData{
				Label:  vol.name,
				Type:   "volume",
				Volume: vol.volume,
				Properties: map[string]interface{}{
					"driver":   vol.volume.Driver,
					"external": vol.volume.External,
					"order":    vol.volume.Order,
					"used_by":  vol.usedBy, // Список сервисов, использующих том
					"used":     true,       // Флаг использования
				},
			},
		}
		nodes = append(nodes, volumeNode)
	}

	// Позиционируем неиспользуемые тома
	for i, vol := range unusedVolumes {
		var x, y int
		nodeID := fmt.Sprintf("volume-%s", vol.name)

		// Неиспользуемый том: позиционируем под compose нодой (такой же X)
		x = options.DockerComposeStart
		y = unusedVolumeStartY + i*options.NodeGapX
		volumeNode := ReactFlowNode{
			ID:   nodeID,
			Type: "volume",
			Position: ReactFlowPosition{
				X: float64(x),
				Y: float64(y),
			},
			Style: map[string]interface{}{
				"opacity": 0.5,
			},
			Data: ReactFlowNodeData{
				Label:  vol.name,
				Type:   "volume",
				Volume: vol.volume,
				Status: "unused", // Статус неиспользуемого тома
				Properties: map[string]interface{}{
					"driver":   vol.volume.Driver,
					"external": vol.volume.External,
					"order":    vol.volume.Order,
					"used_by":  vol.usedBy, // Список сервисов, использующих том
					"used":     false,      // Флаг использования
					"status":   "unused",   // Статус для фильтрации
				},
			},
		}
		nodes = append(nodes, volumeNode)

		// Добавляем связь от Docker Compose к неиспользуемому тому
		edgeCounter++
		edge := ReactFlowEdge{
			ID:           fmt.Sprintf("edge-compose-unused-volume-%s", vol.name),
			Source:       "docker-compose",
			SourceHandle: "docker-compose-source-2",
			Target:       nodeID,
			Type:         "step",
			Style: map[string]interface{}{
				"strokeWidth":     1,
				"stroke":          "#9ca3af",
				"strokeDasharray": "3,3",
				"opacity":         0.6,
			},
			Label: "unused",
			LabelStyle: map[string]interface{}{
				"fill":         "#9ca3af",
				"fontWeight":   "400",
				"fontSize":     "9px",
				"background":   "rgba(255, 255, 255, 0.8)",
				"padding":      "1px 4px",
				"borderRadius": "3px",
			},
		}
		edges = append(edges, edge)
	}

	// 8. Добавляем связи depends_on между сервисами
	// Сначала создаем карту всех сервисов
	allServices := make(map[string]string) // name -> id
	for _, node := range nodes {
		if node.Type == "service" {
			allServices[node.Data.Label] = node.ID
		}
	}

	// Добавляем зависимости между сервисами
	for _, item := range servicesList {
		serviceName := item.name
		service := item.service
		sourceID := fmt.Sprintf("service-%s", serviceName)

		for _, dependsOn := range service.DependsOn {
			if targetID, exists := allServices[dependsOn]; exists {
				edgeCounter++
				edge := ReactFlowEdge{
					ID:           fmt.Sprintf("edge-depends-%d", edgeCounter),
					Source:       sourceID,
					SourceHandle: fmt.Sprintf("%s-source-2", sourceID),
					Target:       targetID,
					TargetHandle: fmt.Sprintf("%s-target-2", targetID),
					Type:         "smoothstep",
					Animated:     true,
				}
				edges = append(edges, edge)
			}
		}
	}

	// 9. Добавляем связи сервисов с томами (если есть прямые связи)
	// Примечание: связи уже учтены при позиционировании томов
	for _, item := range servicesList {
		serviceName := item.name
		service := item.service
		sourceID := fmt.Sprintf("service-%s", serviceName)

		for _, volumeMount := range service.Volumes {
			if volumeMount.Type == "volume" && volumeMount.Source != "" {
				targetID := fmt.Sprintf("volume-%s", volumeMount.Source)

				// Проверяем, существует ли такой том
				volumeExists := false
				for _, node := range nodes {
					if node.ID == targetID {
						volumeExists = true
						break
					}
				}

				if volumeExists {
					edgeCounter++
					edge := ReactFlowEdge{
						ID:       fmt.Sprintf("edge-service-volume-%d", edgeCounter),
						Source:   sourceID,
						Target:   targetID,
						Type:     "smoothstep",
						Animated: true,
					}
					edges = append(edges, edge)
				}
			}
		}
	}

	// 10. Настраиваем viewport для лучшего отображения
	// Рассчитываем границы графа
	var minX, minY, maxX, maxY float64
	if len(nodes) > 0 {
		minX = nodes[0].Position.X
		minY = nodes[0].Position.Y
		maxX = nodes[0].Position.X + float64(nodes[0].Width)
		maxY = nodes[0].Position.Y + float64(nodes[0].Height)

		for _, node := range nodes[1:] {
			if node.Position.X < minX {
				minX = node.Position.X
			}
			if node.Position.Y < minY {
				minY = node.Position.Y
			}
			if node.Position.X+float64(node.Width) > maxX {
				maxX = node.Position.X + float64(node.Width)
			}
			if node.Position.Y+float64(node.Height) > maxY {
				maxY = node.Position.Y + float64(node.Height)
			}
		}

		// Добавляем padding
		minX -= float64(options.Padding)
		minY -= float64(options.Padding)
		maxX += float64(options.Padding)
		maxY += float64(options.Padding)
	}

	// 11. Создаем финальный граф
	graph := &ReactFlowGraph{
		Nodes:     nodes,
		Edges:     edges,
		Project:   project.Name,
		Layout:    "custom",
		Direction: "LR",
		Viewport: ReactFlowViewport{
			X:    minX,
			Y:    minY,
			Zoom: 0.8,
		},
		CreatedAt: time.Now(),
	}

	return graph, nil
}
