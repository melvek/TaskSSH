package config

// Inventory 是清单文件的顶层结构。
type Inventory struct {
	Include  []string         `yaml:"include,omitempty"`
	Defaults Host             `yaml:"defaults"`
	Servers  map[string]Group `yaml:"servers"`
	Tasks    map[string]Task  `yaml:"tasks"`
}

// Group 是服务器组。
type Group struct {
	Vars  Host            `yaml:"vars"`
	Hosts map[string]Host `yaml:"hosts"`
}
