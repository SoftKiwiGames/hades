package inventory

type Provider struct {
	Provider string            `yaml:"provider"`
	Config   map[string]string `yaml:"config"`
	Selector string            `yaml:"selector"`
	Targets  []string          `yaml:"targets"`
	SSH      ProviderSSH       `yaml:"ssh"`
}

type ProviderSSH struct {
	User         string `yaml:"user"`
	Port         int    `yaml:"port"`
	IdentityFile string `yaml:"identity_file"`
}
