package config

import "gopkg.in/yaml.v3"

func yamlUnmarshalImpl(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}
