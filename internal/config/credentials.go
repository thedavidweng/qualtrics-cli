package config

import (
	"os"
	"strconv"
)

func parseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func yamlUnmarshal(data []byte, v any) error {
	return yamlUnmarshalImpl(data, v)
}

func ResolveSecret(value string) string {
	const prefix = "env:"
	if len(value) > len(prefix) && value[:len(prefix)] == prefix {
		return os.Getenv(value[len(prefix):])
	}
	return value
}
