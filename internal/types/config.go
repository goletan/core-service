package types

type CoreConfig struct {
	ServiceName string `mapstructure:"service_name"`

	Discovery struct {
		Filter struct {
			Labels map[string]string `mapstructure:"labels"`
			Tags   map[string]string `mapstructure:"tags"`
		} `mapstructure:"filter"`
	} `mapstructure:"discovery"`

	Orchestrator struct {
		Strategy        string         `mapstructure:"strategy"`
		PriorityMap     map[string]int `mapstructure:"priority_map"`
		PriorityTag     string         `mapstructure:"priority_tag"`
		DefaultPriority int            `mapstructure:"default_priority"`
	} `mapstructure:"orchestrator"`
}
