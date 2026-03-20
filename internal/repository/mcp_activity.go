package repository

import "time"

const MCPActivitySchemaVersion = 1

type MCPActivityRecord struct {
	SchemaVersion      int                   `json:"schema_version"`
	UpdatedAt          string                `json:"updated_at"`
	LastSessionStartAt string                `json:"last_session_start_at,omitempty"`
	LastInitializeAt   string                `json:"last_initialize_at,omitempty"`
	LastToolsListAt    string                `json:"last_tools_list_at,omitempty"`
	RecentToolCalls    []MCPObservedToolCall `json:"recent_tool_calls,omitempty"`
}

type MCPObservedToolCall struct {
	Name string `json:"name"`
	At   string `json:"at"`
}

type DoctorMCPActivitySection struct {
	Status             DoctorStatus
	LastSessionStartAt time.Time
	LastInitializeAt   time.Time
	LastToolsListAt    time.Time
	LastToolCallAt     time.Time
	RecentToolCalls    []MCPObservedToolCall
}

type HostRegistrationState string

const (
	HostRegistrationDetected    HostRegistrationState = "detected"
	HostRegistrationNotDetected HostRegistrationState = "not_detected"
	HostRegistrationUnverified  HostRegistrationState = "unverified"
)

type GuidanceState string

const (
	GuidanceStateConfigured    GuidanceState = "configured"
	GuidanceStateNotConfigured GuidanceState = "not_configured"
	GuidanceStateUnsupported   GuidanceState = "unsupported"
	GuidanceStateUnverified    GuidanceState = "unverified"
)

type DoctorHostRegistrationSection struct {
	Status DoctorStatus
	Hosts  []DoctorHostRegistration
}

type DoctorHostRegistration struct {
	Client               SupportedClient
	RegistrationState    HostRegistrationState
	RegistrationEvidence string
	RegistrationPath     string
	GuidanceState        GuidanceState
	GuidanceEvidence     string
	GuidancePath         string
}
