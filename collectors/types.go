package collectors

type SectionResult[T any] struct {
	Data  *T     `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type Profile struct {
	Name    string `json:"name"`
	Model   string `json:"model"`
	Gateway string `json:"gateway"`
	Active  bool   `json:"active"`
}

type ProfilesData struct {
	Profiles []Profile `json:"profiles"`
}

type KanbanTask struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Assignee string `json:"assignee"`
	Status   string `json:"status"`
	Priority int    `json:"priority"`
}

type KanbanData struct {
	Counts map[string]int `json:"counts"`
	Tasks  []KanbanTask   `json:"tasks"`
}

type MCPServer struct {
	Name      string `json:"name"`
	Transport string `json:"transport"`
	Tools     string `json:"tools"`
	Status    string `json:"status"`
	Enabled   bool   `json:"enabled"`
}

type MCPData struct {
	Servers []MCPServer `json:"servers"`
	Enabled int         `json:"enabled"`
	Total   int         `json:"total"`
}

type GatewayProfile struct {
	Name string `json:"name"`
	PID  int    `json:"pid,omitempty"`
}

type GatewayData struct {
	Running  bool             `json:"running"`
	PID      int              `json:"pid,omitempty"`
	Profiles []GatewayProfile `json:"profiles,omitempty"`
	Summary  string           `json:"summary"`
}

type SystemMetrics struct {
	Label       string `json:"label"`
	Used        string `json:"used"`
	Total       string `json:"total"`
	Available   string `json:"available,omitempty"`
	PercentUsed int    `json:"percent_used"`
}

type UptimeData struct {
	Uptime string  `json:"uptime"`
	Users  int     `json:"users"`
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}

type SystemData struct {
	Memory SystemMetrics  `json:"memory"`
	Disk   SystemMetrics  `json:"disk"`
	Swap   *SystemMetrics `json:"swap,omitempty"`
	Uptime UptimeData     `json:"uptime"`
}

type StatusResponse struct {
	Profiles SectionResult[ProfilesData] `json:"profiles"`
	Kanban   SectionResult[KanbanData]   `json:"kanban"`
	MCP      SectionResult[MCPData]      `json:"mcp"`
	Gateway  SectionResult[GatewayData]  `json:"gateway"`
	System   SectionResult[SystemData]   `json:"system"`
	Updated  string                      `json:"updated"`
}
