package compose_parser

import (
	"time"
)

// ComposeServiceConfig представляет конфигурацию одного сервиса в Docker Compose
type ComposeServiceConfig struct {
	// Основные параметры
	Name       string       `json:"name"`
	Image      string       `json:"image,omitempty"`
	Build      *BuildConfig `json:"build,omitempty"`
	Command    []string     `json:"command,omitempty"`
	Entrypoint []string     `json:"entrypoint,omitempty"`
	WorkingDir string       `json:"working_dir,omitempty"`
	User       string       `json:"user,omitempty"`
	Platform   string       `json:"platform,omitempty"`
	Order      int          `json:"order,omitempty"` // Порядковый номер сервиса в файле

	// Зависимости и перезапуск
	DependsOn []string `json:"depends_on,omitempty"`
	Restart   string   `json:"restart,omitempty"`

	// Сеть и порты
	Ports       []PortMapping `json:"ports,omitempty"`
	Expose      []string      `json:"expose,omitempty"`
	Networks    []string      `json:"networks,omitempty"`
	NetworkMode string        `json:"network_mode,omitempty"`

	// Переменные окружения
	Environment map[string]string `json:"environment,omitempty"`
	EnvFile     []string          `json:"env_file,omitempty"`

	// Тома и монтирования
	Volumes     []VolumeMount `json:"volumes,omitempty"`
	VolumesFrom []string      `json:"volumes_from,omitempty"`

	// Ресурсы
	Deploy     *DeployConfig `json:"deploy,omitempty"`
	CPUShares  int64         `json:"cpu_shares,omitempty"`
	CPUSet     string        `json:"cpuset,omitempty"`
	CPUQuota   int64         `json:"cpu_quota,omitempty"`
	CPUs       float64       `json:"cpus,omitempty"`
	Memory     string        `json:"memory,omitempty"`
	MemorySwap string        `json:"memory_swap,omitempty"`

	// Логирование
	Logging *LoggingConfig `json:"logging,omitempty"`

	// Здоровье
	HealthCheck *HealthCheckConfig `json:"healthcheck,omitempty"`

	// Метки и расширения
	Labels  map[string]string `json:"labels,omitempty"`
	Extends *ExtendsConfig    `json:"extends,omitempty"`

	// Временные метки
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Статус
	Status string `json:"status"` // saved, active, inactive
}

// BuildConfig представляет конфигурацию сборки
type BuildConfig struct {
	Context    string            `json:"context"`
	Dockerfile string            `json:"dockerfile,omitempty"`
	Args       map[string]string `json:"args,omitempty"`
	Target     string            `json:"target,omitempty"`
	CacheFrom  []string          `json:"cache_from,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// PortMapping представляет маппинг портов
type PortMapping struct {
	Target    uint16 `json:"target"`
	Published uint16 `json:"published,omitempty"`
	Protocol  string `json:"protocol,omitempty"`
	Mode      string `json:"mode,omitempty"`
}

// VolumeMount представляет монтирование тома
type VolumeMount struct {
	Type        string `json:"type"` // bind, volume, tmpfs, npipe
	Source      string `json:"source,omitempty"`
	Target      string `json:"target"`
	ReadOnly    bool   `json:"read_only,omitempty"`
	Consistency string `json:"consistency,omitempty"`
}

// DeployConfig представляет конфигурацию развертывания
type DeployConfig struct {
	Mode           string                `json:"mode,omitempty"`
	Replicas       uint64                `json:"replicas,omitempty"`
	Placement      *PlacementConfig      `json:"placement,omitempty"`
	Resources      *ResourceRequirements `json:"resources,omitempty"`
	RestartPolicy  *RestartPolicyConfig  `json:"restart_policy,omitempty"`
	UpdateConfig   *UpdateConfig         `json:"update_config,omitempty"`
	RollbackConfig *RollbackConfig       `json:"rollback_config,omitempty"`
}

// PlacementConfig представляет конфигурацию размещения
type PlacementConfig struct {
	Constraints []string `json:"constraints,omitempty"`
	Preferences []string `json:"preferences,omitempty"`
	MaxReplicas uint64   `json:"max_replicas,omitempty"`
}

// ResourceRequirements представляет требования к ресурсам
type ResourceRequirements struct {
	Limits       *ResourceLimits `json:"limits,omitempty"`
	Reservations *ResourceLimits `json:"reservations,omitempty"`
}

// ResourceLimits представляет лимиты ресурсов
type ResourceLimits struct {
	CPUs   string `json:"cpus,omitempty"`
	Memory string `json:"memory,omitempty"`
	Pids   int64  `json:"pids,omitempty"`
}

// RestartPolicyConfig представляет политику перезапуска
type RestartPolicyConfig struct {
	Condition   string `json:"condition,omitempty"`
	Delay       string `json:"delay,omitempty"`
	MaxAttempts uint64 `json:"max_attempts,omitempty"`
	Window      string `json:"window,omitempty"`
}

// UpdateConfig представляет конфигурацию обновления
type UpdateConfig struct {
	Parallelism     uint64 `json:"parallelism,omitempty"`
	Delay           string `json:"delay,omitempty"`
	FailureAction   string `json:"failure_action,omitempty"`
	Monitor         string `json:"monitor,omitempty"`
	MaxFailureRatio string `json:"max_failure_ratio,omitempty"`
	Order           string `json:"order,omitempty"`
}

// RollbackConfig представляет конфигурацию отката
type RollbackConfig struct {
	Parallelism     uint64 `json:"parallelism,omitempty"`
	Delay           string `json:"delay,omitempty"`
	FailureAction   string `json:"failure_action,omitempty"`
	Monitor         string `json:"monitor,omitempty"`
	MaxFailureRatio string `json:"max_failure_ratio,omitempty"`
	Order           string `json:"order,omitempty"`
}

// LoggingConfig представляет конфигурацию логирования
type LoggingConfig struct {
	Driver  string            `json:"driver,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}

// HealthCheckConfig представляет конфигурацию проверки здоровья
type HealthCheckConfig struct {
	Test          []string `json:"test,omitempty"`
	Interval      string   `json:"interval,omitempty"`
	Timeout       string   `json:"timeout,omitempty"`
	Retries       uint64   `json:"retries,omitempty"`
	StartPeriod   string   `json:"start_period,omitempty"`
	StartInterval string   `json:"start_interval,omitempty"`
}

// ExtendsConfig представляет конфигурацию расширения
type ExtendsConfig struct {
	File    string `json:"file,omitempty"`
	Service string `json:"services,omitempty"`
}

// ComposeProjectConfig представляет полную конфигурацию Docker Compose проекта
type ComposeProjectConfig struct {
	// Версия Compose
	Version string `json:"version,omitempty"`

	// Сервисы
	Services     map[string]*ComposeServiceConfig `json:"services"`
	ServiceOrder []string                         `json:"service_order,omitempty"` // Порядок сервисов в файле
	VolumeOrder  []string                         `json:"volume_order,omitempty"`  // Порядок томов в файле

	// Сети
	Networks map[string]*NetworkConfig `json:"networks,omitempty"`

	// Тома
	Volumes map[string]*VolumeConfig `json:"volumes,omitempty"`

	// Секреты
	Secrets map[string]*SecretConfig `json:"secrets,omitempty"`

	// Конфигурации
	Configs map[string]*ConfigConfig `json:"configs,omitempty"`

	// Метаданные
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Статус
	Status string `json:"status"` // draft, active, archived
}

// NetworkConfig представляет конфигурацию сети
type NetworkConfig struct {
	Driver     string            `json:"driver,omitempty"`
	DriverOpts map[string]string `json:"driver_opts,omitempty"`
	External   bool              `json:"external,omitempty"`
	Name       string            `json:"name,omitempty"`
	Attachable bool              `json:"attachable,omitempty"`
	Internal   bool              `json:"internal,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// VolumeConfig представляет конфигурацию тома
type VolumeConfig struct {
	Driver     string            `json:"driver,omitempty"`
	DriverOpts map[string]string `json:"driver_opts,omitempty"`
	External   bool              `json:"external,omitempty"`
	Name       string            `json:"name,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Order      int               `json:"order,omitempty"` // Порядковый номер тома в файле
}

// SecretConfig представляет конфигурацию секрета
type SecretConfig struct {
	File     string            `json:"file,omitempty"`
	External bool              `json:"external,omitempty"`
	Name     string            `json:"name,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
}

// ConfigConfig представляет конфигурацию конфигурации
type ConfigConfig struct {
	File     string            `json:"file,omitempty"`
	External bool              `json:"external,omitempty"`
	Name     string            `json:"name,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
}

// ComposeServiceStatus представляет статус сервиса Docker Compose
type ComposeServiceStatus struct {
	Name         string    `json:"name"`
	Status       string    `json:"status"` // running, stopped, paused, restarting, dead
	Health       string    `json:"health"` // healthy, unhealthy, starting, none
	RestartCount int       `json:"restart_count"`
	CreatedAt    time.Time `json:"created_at"`
	StartedAt    time.Time `json:"started_at,omitempty"`
	FinishedAt   time.Time `json:"finished_at,omitempty"`
	ExitCode     int       `json:"exit_code,omitempty"`
}

// ComposeProjectStatus представляет статус проекта Docker Compose
type ComposeProjectStatus struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"` // running, stopped, partially_running, error
	Services  []ComposeServiceStatus `json:"services"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// ReactFlowNode представляет узел в React Flow графе
type ReactFlowNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // 'services', 'network', 'volume', 'secret', 'config'
	Position ReactFlowPosition      `json:"position"`
	Data     ReactFlowNodeData      `json:"data"`
	Width    int                    `json:"width,omitempty"`
	Height   int                    `json:"height,omitempty"`
	Selected bool                   `json:"selected,omitempty"`
	Dragging bool                   `json:"dragging,omitempty"`
	Style    map[string]interface{} `json:"style,omitempty"`
}

// ReactFlowEdge представляет связь в React Flow графе
type ReactFlowEdge struct {
	ID           string                 `json:"id"`
	Source       string                 `json:"source"`
	SourceHandle string                 `json:"sourceHandle"`
	Target       string                 `json:"target"`
	TargetHandle string                 `json:"targetHandle"`
	Type         string                 `json:"type,omitempty"` // 'default', 'smoothstep', 'step', 'straight'
	Animated     bool                   `json:"animated,omitempty"`
	Style        map[string]interface{} `json:"style,omitempty"`
	Label        string                 `json:"label,omitempty"`
	NetworkName  string                 `json:"networkName,omitempty"`
	ServiceName  string                 `json:"serviceName,omitempty"`
	LabelStyle   map[string]interface{} `json:"labelStyle,omitempty"`
	LabelBgStyle map[string]interface{} `json:"labelBgStyle,omitempty"`
}

// ReactFlowPosition представляет позицию узла в графе
type ReactFlowPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ReactFlowNodeData представляет данные узла
type ReactFlowNodeData struct {
	Label       string                 `json:"label"`
	Type        string                 `json:"type"` // 'services', 'network', 'volume', 'secret', 'config'
	Service     *ComposeServiceConfig  `json:"services,omitempty"`
	Network     *NetworkConfig         `json:"network,omitempty"`
	Volume      *VolumeConfig          `json:"volume,omitempty"`
	Secret      *SecretConfig          `json:"secret,omitempty"`
	Config      *ConfigConfig          `json:"config,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// ReactFlowGraph представляет полный граф для React Flow
type ReactFlowGraph struct {
	Nodes     []ReactFlowNode   `json:"nodes"`
	Edges     []ReactFlowEdge   `json:"edges"`
	Project   string            `json:"project"`
	Layout    string            `json:"layout"`    // 'dagre', 'elk', 'custom'
	Direction string            `json:"direction"` // 'TB', 'BT', 'LR', 'RL'
	Viewport  ReactFlowViewport `json:"viewport"`
	CreatedAt time.Time         `json:"created_at"`
}

// ReactFlowViewport представляет viewport графа
type ReactFlowViewport struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom float64 `json:"zoom"`
}

// GraphLayoutOptions представляет опции для размещения графа
type GraphLayoutOptions struct {
	Direction          string `json:"direction"` // 'TB', 'BT', 'LR', 'RL'
	NodeWidth          int    `json:"node_width"`
	NodeHeight         int    `json:"node_height"`
	NodeGapX           int    `json:"node_gap_x"`
	NodeGapY           int    `json:"node_gap_y"`
	Padding            int    `json:"padding"`
	ColumnGap          int    `json:"column_gap"`
	ColumnTopGap       int    `json:"column_top_gap"`
	DockerComposeStart int    `json:"docker_compose_start"`
	VolumeXOffset      int    `json:"volume_x_offset"`
	VolumeYOffset      int    `json:"volume_y_offset"`
	LastY              int    `json:"last_y"`
}

type networkWithName struct {
	name    string
	network *NetworkConfig
}

type volumeWithOrder struct {
	name   string
	order  int
	volume *VolumeConfig
}

type positionedVolume struct {
	name     string
	volume   *VolumeConfig
	targetY  int
	serviceX int // Средняя X сервисов
	usedBy   []string
}

type serviceWithOrder struct {
	name    string
	order   int
	service *ComposeServiceConfig
}

// GraphDimensions содержит предварительно рассчитанные размеры и позиции
type GraphDimensions struct {
	ServiceCount       int
	VolumeCount        int
	NetworkCount       int
	DockerComposeX     int
	DockerComposeY     int
	ServiceHeight      int
	NetworkHeight      int
	ServiceBaseY       int
	NetworkStartY      int
	ServiceStartX      int
	ServiceStartY      int
	VolumeYOffset      int
	UnusedVolumeStartY int
}
