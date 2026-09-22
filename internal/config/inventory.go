package config

// Inventory 是清单文件的顶层结构。
type Inventory struct {
	GlobalVars Host             `yaml:"global_vars"`
	Servers    map[string]Group `yaml:"servers"`
	Tasks      map[string]Task  `yaml:"tasks"`
}
