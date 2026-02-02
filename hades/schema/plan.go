package schema

type Plan struct {
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Job     string         `yaml:"job"`
	Targets []string       `yaml:"targets"`
	Env     map[string]Env `yaml:"env"`
}
