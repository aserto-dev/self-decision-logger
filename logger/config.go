package logger

type DecisionLogConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}
