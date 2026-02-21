package makecom

// CreateScenarioRequest is the payload for POST /scenarios.
type CreateScenarioRequest struct {
	Blueprint  string `json:"blueprint"`
	TeamID     int    `json:"teamId"`
	Scheduling string `json:"scheduling"`
	FolderID   int    `json:"folderId,omitempty"`
}

// CreateScenarioResponse is the response from POST /scenarios.
type CreateScenarioResponse struct {
	Scenario ScenarioMeta `json:"scenario"`
}

// ScenarioMeta holds the identifiers returned after scenario creation.
type ScenarioMeta struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Blueprint is the top-level scenario blueprint object.
type Blueprint struct {
	Name     string           `json:"name"`
	Flow     []Module         `json:"flow"`
	Metadata BlueprintMetadata `json:"metadata"`
}

// Module represents a single step in the scenario flow.
type Module struct {
	ID         int            `json:"id"`
	Module     string         `json:"module"`
	Version    int            `json:"version"`
	Parameters map[string]any `json:"parameters,omitempty"`
	Filter     *ModuleFilter  `json:"filter,omitempty"`
	Mapper     map[string]any `json:"mapper"`
	Metadata   ModuleMetadata `json:"metadata"`
}

// ModuleFilter defines a condition that must be true for the module to execute.
type ModuleFilter struct {
	Name       string              `json:"name"`
	Conditions [][]FilterCondition `json:"conditions"`
}

// FilterCondition is a single predicate: a <op> b.
type FilterCondition struct {
	A string `json:"a"`
	B string `json:"b"`
	O string `json:"o"`
}

// ModuleMetadata holds designer positioning and restore hints for a module.
type ModuleMetadata struct {
	Designer Designer       `json:"designer"`
	Restore  map[string]any `json:"restore,omitempty"`
}

// Designer holds the visual position of a module in the Make.com editor.
type Designer struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// BlueprintMetadata holds scenario-level execution settings.
type BlueprintMetadata struct {
	Instant  bool            `json:"instant"`
	Version  int             `json:"version"`
	Scenario ScenarioOptions `json:"scenario"`
	Designer DesignerMeta    `json:"designer"`
	Zone     string          `json:"zone"`
	Notes    []any           `json:"notes"`
}

// DesignerMeta holds visual editor metadata.
type DesignerMeta struct {
	Orphans []any `json:"orphans"`
}

// ScenarioOptions controls execution behaviour.
type ScenarioOptions struct {
	RoundTrips            int  `json:"roundtrips"`
	MaxErrors             int  `json:"maxErrors"`
	AutoCommit            bool `json:"autoCommit"`
	AutoCommitTriggerLast bool `json:"autoCommitTriggerLast"`
	Sequential            bool `json:"sequential"`
	Confidential          bool `json:"confidential"`
	Dataloss              bool `json:"dataloss"`
	DLQ                   bool `json:"dlq"`
	FreshVariables        bool `json:"freshVariables"`
}

// Scheduling defines how often the scenario runs.
type Scheduling struct {
	Type     string `json:"type"`
	Interval int    `json:"interval"`
}
