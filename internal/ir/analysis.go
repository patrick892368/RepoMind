package ir

type Analysis struct {
	SchemaVersion string         `json:"schema_version"`
	GeneratedBy   string         `json:"generated_by"`
	Language      string         `json:"language,omitempty"`
	Repository    RepositoryInfo `json:"repository"`
	Stack         StackInfo      `json:"stack"`
	Scan          ScanSummary    `json:"scan"`
	Summary       ProjectSummary `json:"summary,omitempty"`
	Packages      []PackageInfo  `json:"packages,omitempty"`
	Models        []DBModel      `json:"models,omitempty"`
	Routes        []APIRoute     `json:"routes,omitempty"`
	CallEdges     []CallEdge     `json:"call_edges,omitempty"`
	Diagrams      DiagramSet     `json:"diagrams,omitempty"`
}

type RepositoryInfo struct {
	Name       string `json:"name"`
	Root       string `json:"root"`
	AnalyzedAt string `json:"analyzed_at"`
}

type StackInfo struct {
	Backend        string   `json:"backend,omitempty"`
	Frontend       string   `json:"frontend,omitempty"`
	Database       string   `json:"database,omitempty"`
	Cache          string   `json:"cache,omitempty"`
	Queue          string   `json:"queue,omitempty"`
	Languages      []string `json:"languages,omitempty"`
	PackageManager []string `json:"package_manager,omitempty"`
	ConfigFiles    []string `json:"config_files,omitempty"`
}

type PackageInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Parent       string    `json:"parent,omitempty"`
	Type         string    `json:"type,omitempty"`
	Dependencies []string  `json:"dependencies,omitempty"`
	Stack        StackInfo `json:"stack,omitempty"`
	Files        int       `json:"files"`
	Directories  int       `json:"directories"`
	Models       int       `json:"models"`
	Routes       int       `json:"routes"`
}

type ScanSummary struct {
	TotalFiles       int            `json:"total_files"`
	TotalDirectories int            `json:"total_directories"`
	TotalBytes       int64          `json:"total_bytes"`
	Truncated        bool           `json:"truncated"`
	Files            []FileEntry    `json:"files"`
	Directories      []string       `json:"directories"`
	Ignored          []string       `json:"ignored"`
	LanguageCounts   map[string]int `json:"language_counts"`
	Errors           []ScanError    `json:"errors,omitempty"`
}

type FileEntry struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Extension string `json:"extension,omitempty"`
	Language  string `json:"language,omitempty"`
}

type ScanError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type DBModel struct {
	Name       string       `json:"name"`
	Table      string       `json:"table,omitempty"`
	File       string       `json:"file"`
	Line       int          `json:"line,omitempty"`
	Source     string       `json:"source"`
	Confidence string       `json:"confidence,omitempty"`
	Evidence   string       `json:"evidence,omitempty"`
	Fields     []DBField    `json:"fields"`
	Relations  []DBRelation `json:"relations,omitempty"`
}

type DBField struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Required   bool   `json:"required"`
	PrimaryKey bool   `json:"primary_key,omitempty"`
	Unique     bool   `json:"unique,omitempty"`
}

type DBRelation struct {
	Name       string   `json:"name"`
	Target     string   `json:"target"`
	Type       string   `json:"type"`
	Fields     []string `json:"fields,omitempty"`
	References []string `json:"references,omitempty"`
}

type DiagramSet struct {
	ER      string `json:"er,omitempty"`
	API     string `json:"api,omitempty"`
	Call    string `json:"call,omitempty"`
	Package string `json:"package,omitempty"`
}

type APIRoute struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	Handler    string `json:"handler,omitempty"`
	File       string `json:"file"`
	Line       int    `json:"line,omitempty"`
	Source     string `json:"source"`
	Confidence string `json:"confidence,omitempty"`
	Evidence   string `json:"evidence,omitempty"`
}

type CallEdge struct {
	Caller string `json:"caller"`
	Callee string `json:"callee"`
	File   string `json:"file"`
	Line   int    `json:"line"`
	Source string `json:"source"`
}

type ProjectSummary struct {
	Title      string   `json:"title,omitempty"`
	Overview   string   `json:"overview,omitempty"`
	Modules    []string `json:"modules,omitempty"`
	Stack      []string `json:"stack,omitempty"`
	KeyFlows   []string `json:"key_flows,omitempty"`
	StartHints []string `json:"start_hints,omitempty"`
}
