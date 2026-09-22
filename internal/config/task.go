package config

// Task 是一个任务，由若干步骤组成。
type Task struct {
	Description string `yaml:"description,omitempty"`
	Steps       []Step `yaml:"steps"`
}

// Step 是任务中的一步，引用一个 Action。
type Step struct {
	Name   string                 `yaml:"name"`
	Action string                 `yaml:"action"`
	With   map[string]interface{} `yaml:"with"`
	Delay  int                    `yaml:"delay"`
}
