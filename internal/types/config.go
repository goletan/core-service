package types

type CoreConfig struct {
	ServiceName  string `mapstructure:"service_name"`
	Orchestrator struct {
		Filter          map[string]string `mapstructure:"filter"`
		Strategy        string            `mapstructure:"strategy"`
		PriorityMap     map[string]int    `mapstructure:"priority_map"`
		PriorityTag     string            `mapstructure:"priority_tag"`
		DefaultPriority int               `mapstructure:"default_priority"`
	} `mapstructure:"orchestrator"`
}
