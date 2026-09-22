package config

// Group 是服务器组。
type Group struct {
	Vars  Host            `yaml:"vars"`
	Hosts map[string]Host `yaml:"hosts"`
}
