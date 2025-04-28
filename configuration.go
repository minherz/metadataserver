package metadataserver

import (
	"encoding/json"
	"fmt"
	"os"
)

type Metadata func() string

type Configuration struct {
	Port            int
	Address         string
	Endpoint        string
	Handlers        map[string]Metadata
	ShutdownTimeout int
}

type jsonConfiguration struct {
	Address         string         `json:"address"`
	Endpoint        string         `json:"endpoint"`
	Handlers        map[string]any `json:"metadata"`
	Port            int            `json:"port"`
	ShutdownTimeout int            `json:"shutdownTimeout"`
}

const (
	DefaultAddress         = "169.254.169.254"
	DefaultEndpoint        = "/computeMetadata/v1"
	DefaultPort            = 80
	DefaultShutdownTimeout = 5
)

var DefaultConfigurationHandlers = map[string]Metadata{
	"project/project-id": func() string {
		return "test-project-id"
	},
}
var EmptyConfigurationHandlers = map[string]Metadata{}

func NewConfigFromFile(path string) (*Configuration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var jc jsonConfiguration
	if err := json.Unmarshal(data, &jc); err != nil {
		return nil, err
	}
	c := NewConfiguration(DefaultConfigurationHandlers)
	if jc.Port > 0 {
		c.Port = jc.Port
	}
	if jc.ShutdownTimeout > 0 {
		c.ShutdownTimeout = jc.ShutdownTimeout
	}
	if jc.Address != "" {
		c.Address = jc.Address
	}
	if jc.Endpoint != "" {
		if jc.Endpoint[0] != '/' {
			jc.Endpoint = "/" + jc.Endpoint
		}
		c.Endpoint = jc.Endpoint
	}
	c.Handlers = convert(jc.Handlers)
	return c, nil
}

func NewConfiguration(handlers map[string]Metadata) *Configuration {
	if len(handlers) == 0 {
		handlers = DefaultConfigurationHandlers
	}
	return &Configuration{
		Address:         DefaultAddress,
		Endpoint:        DefaultEndpoint,
		Handlers:        handlers,
		Port:            DefaultPort,
		ShutdownTimeout: DefaultShutdownTimeout,
	}
}

func convert(m map[string]any) map[string]Metadata {
	result := make(map[string]Metadata)
	for k, v := range m {
		if dataMap, ok := v.(map[string]any); ok {
			if v2, ok := dataMap["value"]; ok {
				s := fmt.Sprintf("%v", v2)
				result[k] = func() string {
					return s
				}
				continue
			}
			if v2, ok := dataMap["env"]; ok {
				s := fmt.Sprintf("%v", v2)
				result[k] = func() string {
					return os.Getenv(s)
				}
				continue
			}
		}
	}
	return result
}
