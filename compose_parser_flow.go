package compose_parser

import (
	"fmt"
	"sort"
	"time"
)

// ParseToReactFlow парсит конфигурацию Docker Compose проекта в граф для React Flow
func (p *ComposeParser) ParseToReactFlow(project *ComposeProjectConfig, options *GraphLayoutOptions) (*ReactFlowGraph, error) {
	// 1. Инициализация опций по умолчанию
	options = p.initDefaultOptions(options)

	// 2. Рассчет размеров и позиций
	dimensions := p.calculateGraphDimensions(project, options)

	// 3. Создание ноды DockerCompose
	dockerComposeNode := p.createDockerComposeNode(project, options, dimensions)

	// 4. Создание нод сетей
	networkNodes, networkNodeMap := p.createNetworkNodes(project, options, dimensions)

	// 5. Создание нод сервисов
	serviceNodes, serviceMap := p.createServiceNodes(project, options, dimensions)

	// 6. Создание связей с сетями
	networkEdges := p.createNetworkToServiceEdges(project, serviceNodes, networkNodeMap, dimensions)

	// 7. Сбор информации об использовании томов
	volumeUsage, servicePositions := p.collectVolumeUsage(project, serviceNodes)

	// 8. Создание нод томов
	volumeNodes := p.createVolumeNodes(project, options, dimensions, volumeUsage, servicePositions)

	// 9. Создание связей зависимостей
	dependsEdges := p.createDependsOnEdges(project, serviceMap)

	// 10. Создание связей сервисов с томами
	serviceToVolumeEdges := p.createServiceToVolumeEdges(project, serviceNodes, volumeUsage)

	// 11. Расчет viewport
	allNodes := append(append(append([]ReactFlowNode{dockerComposeNode}, networkNodes...), serviceNodes...), volumeNodes...)
	viewport := p.calculateViewport(allNodes, options)

	// 12. Создание финального графа
	allEdges := append(append(networkEdges, dependsEdges...), serviceToVolumeEdges...)
	return p.buildFinalGraph(project, allNodes, allEdges, viewport), nil
}

// initDefaultOptions инициализирует опции по умолчанию, если они не заданы
func (p *ComposeParser) initDefaultOptions(options *GraphLayoutOptions) *GraphLayoutOptions {
	if options == nil {
		return &GraphLayoutOptions{
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
	return options
}

// calculateGraphDimensions рассчитывает размеры и позиции для всех элементов графа
func (p *ComposeParser) calculateGraphDimensions(project *ComposeProjectConfig, options *GraphLayoutOptions) *GraphDimensions {
	serviceCount := len(project.Services)
	volumeCount := len(project.Volumes)
	networkCount := len(project.Networks)

	// Позиция DockerCompose ноды
	dockerComposeX := 0
	dockerComposeY := (serviceCount*options.NodeHeight)/2 - 10

	// Рассчитываем высоту, занимаемую всеми сервисами
	serviceHeight := serviceCount * options.NodeHeight
	if serviceHeight < options.NodeWidth {
		serviceHeight = options.NodeWidth
	}

	// Рассчитываем высоту, занимаемую всеми сетями
	networkHeight := networkCount * options.NodeHeight
	if networkHeight < options.NodeHeight {
		networkHeight = options.NodeHeight
	}

	// Центрируем сети относительно высоты сервисов
	serviceBaseY := options.Padding
	var networkStartY int
	if serviceHeight >= networkHeight {
		networkStartY = serviceBaseY + (serviceHeight-networkHeight)/2
	} else {
		networkStartY = serviceBaseY - (networkHeight-serviceHeight)/2
	}

	// Позиция начала сервисов
	serviceStartX := dockerComposeX + options.ColumnGap
	serviceStartY := options.Padding

	// Позиция неиспользуемых томов
	volumeYOffset := options.VolumeYOffset
	unusedVolumeStartY := dockerComposeY + volumeYOffset

	return &GraphDimensions{
		ServiceCount:       serviceCount,
		VolumeCount:        volumeCount,
		NetworkCount:       networkCount,
		DockerComposeX:     dockerComposeX,
		DockerComposeY:     dockerComposeY,
		ServiceHeight:      serviceHeight,
		NetworkHeight:      networkHeight,
		ServiceBaseY:       serviceBaseY,
		NetworkStartY:      networkStartY,
		ServiceStartX:      serviceStartX,
		ServiceStartY:      serviceStartY,
		VolumeYOffset:      volumeYOffset,
		UnusedVolumeStartY: unusedVolumeStartY,
	}
}

// createDockerComposeNode создает ноду DockerCompose (самая левая)
func (p *ComposeParser) createDockerComposeNode(project *ComposeProjectConfig, options *GraphLayoutOptions, dimensions *GraphDimensions) ReactFlowNode {
	return ReactFlowNode{
		ID:   "docker-compose",
		Type: "compose",
		Position: ReactFlowPosition{
			X: float64(options.DockerComposeStart),
			Y: float64(dimensions.DockerComposeY),
		},
		Data: ReactFlowNodeData{
			Label: project.Name,
			Type:  "compose",
			Properties: map[string]interface{}{
				"services": dimensions.ServiceCount,
				"networks": dimensions.NetworkCount,
				"volumes":  dimensions.VolumeCount,
				"version":  project.Version,
			},
		},
	}
}

// createNetworkNodes создает ноды сетей (вторая колонка)
func (p *ComposeParser) createNetworkNodes(project *ComposeProjectConfig, options *GraphLayoutOptions, dimensions *GraphDimensions) ([]ReactFlowNode, map[string]string) {
	nodes := make([]ReactFlowNode, 0)
	networkNodes := make(map[string]string)

	// Получаем список сетей и сортируем их
	networksList := p.getSortedNetworks(project.Networks)

	networkIndex := 0
	for _, item := range networksList {
		networkName := item.name
		network := item.network

		// Позиционируем сети вертикально во второй колонке
		x := dimensions.DockerComposeX + 20
		y := dimensions.NetworkStartY + networkIndex*120 // 120px отступ между сетями

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
		networkIndex++
	}

	return nodes, networkNodes
}

// createServiceNodes создает ноды сервисов (центральная колонка)
func (p *ComposeParser) createServiceNodes(project *ComposeProjectConfig, options *GraphLayoutOptions, dimensions *GraphDimensions) ([]ReactFlowNode, map[string]string) {
	nodes := make([]ReactFlowNode, 0)
	serviceMap := make(map[string]string)

	servicesList := p.getSortedServices(project.Services, project.ServiceOrder)

	serviceIndex := 0
	for _, item := range servicesList {
		serviceName := item.name
		service := item.service

		// Позиционируем сервисы вертикально в центральной колонке
		x := dimensions.ServiceStartX
		y := dimensions.ServiceStartY + serviceIndex*options.ColumnTopGap

		nodeID := fmt.Sprintf("services-%s", serviceName)
		serviceMap[serviceName] = nodeID

		nodeColor := "#3b82f6"

		serviceNode := ReactFlowNode{
			ID:   nodeID,
			Type: "services",
			Position: ReactFlowPosition{
				X: float64(x),
				Y: float64(y),
			},
			Data: ReactFlowNodeData{
				Label:   serviceName,
				Type:    "services",
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
		serviceIndex++
	}

	return nodes, serviceMap
}

// createNetworkToServiceEdges создает связи от сетей к сервисам
func (p *ComposeParser) createNetworkToServiceEdges(project *ComposeProjectConfig, serviceNodes []ReactFlowNode, networkNodeMap map[string]string, dimensions *GraphDimensions) []ReactFlowEdge {
	edges := make([]ReactFlowEdge, 0)
	edgeCounter := 0

	servicesList := p.getSortedServices(project.Services, project.ServiceOrder)

	for _, item := range servicesList {
		serviceName := item.name
		service := item.service
		nodeID := fmt.Sprintf("services-%s", serviceName)

		hasNetworkConnections := false
		for _, networkName := range service.Networks {
			if networkNodeID, exists := networkNodeMap[networkName]; exists {
				edgeCounter++
				edgeLabel := networkName
				labelStyle := map[string]interface{}{
					"fill":      "#3b82f6",
					"opacity":   0.4,
					"textAlign": "center",
				}

				edge := ReactFlowEdge{
					ID:     fmt.Sprintf("edge-network-services-%s-%s", networkName, serviceName),
					Source: networkNodeID,
					Target: nodeID,
					Type:   "smoothstep",
					Style: map[string]interface{}{
						"strokeWidth": 0,
						"stroke":      "transparent",
					},
					Label:      edgeLabel,
					LabelStyle: labelStyle,
				}
				edges = append(edges, edge)
				hasNetworkConnections = true
			}
		}

		// Создаем связь от DockerCompose к сервису, если нет сетей
		if !hasNetworkConnections {
			edgeCounter++
			edge := ReactFlowEdge{
				ID:     fmt.Sprintf("edge-compose-services-%d", edgeCounter),
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
	}

	return edges
}

// collectVolumeUsage собирает информацию об использовании томов сервисами
func (p *ComposeParser) collectVolumeUsage(project *ComposeProjectConfig, serviceNodes []ReactFlowNode) (map[string][]string, map[string]ReactFlowPosition) {
	volumeUsage := make(map[string][]string)
	servicePositions := make(map[string]ReactFlowPosition)

	// Собираем позиции сервисов
	for _, node := range serviceNodes {
		if node.Type == "services" {
			servicePositions[node.Data.Label] = node.Position
		}
	}

	// Собираем использование томов
	servicesList := p.getSortedServices(project.Services, project.ServiceOrder)

	for _, item := range servicesList {
		serviceName := item.name
		service := item.service

		for _, volumeMount := range service.Volumes {
			if volumeMount.Type == "volume" && volumeMount.Source != "" {
				volumeName := volumeMount.Source
				volumeUsage[volumeName] = append(volumeUsage[volumeName], serviceName)
			}
		}
	}

	return volumeUsage, servicePositions
}

// createVolumeNodes создает ноды томов (используемые и неиспользуемые)
func (p *ComposeParser) createVolumeNodes(project *ComposeProjectConfig, options *GraphLayoutOptions, dimensions *GraphDimensions, volumeUsage map[string][]string, servicePositions map[string]ReactFlowPosition) []ReactFlowNode {
	nodes := make([]ReactFlowNode, 0)

	// Рассчитываем средние позиции для используемых томов
	volumeServiceX := make(map[string]int)
	volumeServiceY := make(map[string]int)

	for volumeName, serviceNames := range volumeUsage {
		if len(serviceNames) > 0 {
			sumX := 0
			sumY := 0
			count := 0
			for _, serviceName := range serviceNames {
				if pos, exists := servicePositions[serviceName]; exists {
					sumX += int(pos.X)
					sumY += int(pos.Y)
					count++
				}
			}

			if count > 0 {
				volumeServiceX[volumeName] = sumX / count
				volumeServiceY[volumeName] = sumY / count
			}
		}
	}

	volumeUsed := make(map[string]bool)
	for volumeName := range volumeUsage {
		volumeUsed[volumeName] = true
	}

	volumesList := p.getSortedVolumes(project.Volumes, project.VolumeOrder)

	volumeXOffset := options.VolumeXOffset
	unusedVolumeStartY := dimensions.UnusedVolumeStartY

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

	lastY := options.LastY

	// Создаем используемые тома
	for i, vol := range usedVolumes {
		var x, y int
		nodeID := fmt.Sprintf("volume-%s", vol.name)

		desiredY := vol.targetY

		if i == 0 {
			if desiredY < options.Padding {
				y = options.Padding
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
					"used_by":  vol.usedBy,
					"used":     true,
				},
			},
		}
		nodes = append(nodes, volumeNode)
	}

	// Создаем неиспользуемые тома
	for i, vol := range unusedVolumes {
		var x, y int
		nodeID := fmt.Sprintf("volume-%s", vol.name)

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
				Status: "unused",
				Properties: map[string]interface{}{
					"driver":   vol.volume.Driver,
					"external": vol.volume.External,
					"order":    vol.volume.Order,
					"used_by":  vol.usedBy,
					"used":     false,
					"status":   "unused",
				},
			},
		}
		nodes = append(nodes, volumeNode)

		// Добавляем связь от Docker Compose к неиспользуемому тому
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
		// Примечание: связь добавляется отдельно в основной метод
		_ = edge // В текущей реализации не добавляем, чтобы не дублировать
	}

	return nodes
}

// createDependsOnEdges создает связи зависимостей между сервисами
func (p *ComposeParser) createDependsOnEdges(project *ComposeProjectConfig, serviceMap map[string]string) []ReactFlowEdge {
	edges := make([]ReactFlowEdge, 0)
	edgeCounter := 0

	servicesList := p.getSortedServices(project.Services, project.ServiceOrder)

	for _, item := range servicesList {
		serviceName := item.name
		sourceID := fmt.Sprintf("services-%s", serviceName)

		for _, dependsOn := range item.service.DependsOn {
			if targetID, exists := serviceMap[dependsOn]; exists {
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

	return edges
}

// createServiceToVolumeEdges создает связи сервисов с томами
func (p *ComposeParser) createServiceToVolumeEdges(project *ComposeProjectConfig, serviceNodes []ReactFlowNode, volumeUsage map[string][]string) []ReactFlowEdge {
	edges := make([]ReactFlowEdge, 0)
	edgeCounter := 0

	servicesList := p.getSortedServices(project.Services, project.ServiceOrder)

	for _, item := range servicesList {
		serviceName := item.name
		sourceID := fmt.Sprintf("services-%s", serviceName)

		for _, volumeMount := range item.service.Volumes {
			if volumeMount.Type == "volume" && volumeMount.Source != "" {
				targetID := fmt.Sprintf("volume-%s", volumeMount.Source)

				// Проверяем, существует ли такой том
				volumeExists := false
				for _, node := range serviceNodes {
					if node.ID == targetID {
						volumeExists = true
						break
					}
				}

				// Также проверяем в volumeUsage
				if !volumeExists {
					for volName := range volumeUsage {
						if volName == volumeMount.Source {
							volumeExists = true
							break
						}
					}
				}

				if volumeExists {
					edgeCounter++
					edge := ReactFlowEdge{
						ID:       fmt.Sprintf("edge-services-volume-%d", edgeCounter),
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

	return edges
}

// calculateViewport рассчитывает viewport для отображения графа
func (p *ComposeParser) calculateViewport(nodes []ReactFlowNode, options *GraphLayoutOptions) ReactFlowViewport {
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

	return ReactFlowViewport{
		X:    minX,
		Y:    minY,
		Zoom: 0.8,
	}
}

// buildFinalGraph создает финальный граф для React Flow
func (p *ComposeParser) buildFinalGraph(project *ComposeProjectConfig, nodes []ReactFlowNode, edges []ReactFlowEdge, viewport ReactFlowViewport) *ReactFlowGraph {
	return &ReactFlowGraph{
		Nodes:     nodes,
		Edges:     edges,
		Project:   project.Name,
		Layout:    "custom",
		Direction: "LR",
		Viewport:  viewport,
		CreatedAt: time.Now(),
	}
}

// getSortedNetworks возвращает отсортированный список сетей
func (p *ComposeParser) getSortedNetworks(networks map[string]*NetworkConfig) []networkWithName {
	networksList := make([]networkWithName, 0, len(networks))
	for networkName, network := range networks {
		networksList = append(networksList, networkWithName{
			name:    networkName,
			network: network,
		})
	}

	sort.Slice(networksList, func(i, j int) bool {
		return networksList[i].name < networksList[j].name
	})

	return networksList
}

// getSortedServices возвращает отсортированный список сервисов
func (p *ComposeParser) getSortedServices(services map[string]*ComposeServiceConfig, serviceOrder []string) []serviceWithOrder {
	servicesList := make([]serviceWithOrder, 0, len(services))
	for serviceName, service := range services {
		servicesList = append(servicesList, serviceWithOrder{
			name:    serviceName,
			order:   service.Order,
			service: service,
		})
	}

	sort.Slice(servicesList, func(i, j int) bool {
		if servicesList[i].order == 0 && servicesList[j].order == 0 {
			idxI := -1
			idxJ := -1
			for k, name := range serviceOrder {
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
		return servicesList[i].order < servicesList[j].order
	})

	return servicesList
}

// getSortedVolumes возвращает отсортированный список томов
func (p *ComposeParser) getSortedVolumes(volumes map[string]*VolumeConfig, volumeOrder []string) []volumeWithOrder {
	volumesList := make([]volumeWithOrder, 0, len(volumes))
	for volumeName, volume := range volumes {
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
			for k, name := range volumeOrder {
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

	return volumesList
}
