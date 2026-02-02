package schema

type Job struct {
	Local     bool                `yaml:"local"`
	Env       map[string]Env      `yaml:"env"`
	Artifacts map[string]Artifact `yaml:"artifacts"`
	Actions   []Action            `yaml:"actions"`
}

type Artifact struct {
	Path string `yaml:"path"`
}

type Action struct {
	Run  *ActionRun  `yaml:"run,omitempty"`
	Copy *ActionCopy `yaml:"copy,omitempty"`
}

type ActionRun string

type ActionCopy struct {
	Src string `yaml:"src"`
	Dst string `yaml:"dst"`
}
