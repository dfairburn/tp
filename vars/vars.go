package vars

import (
  "log"
  "os"
  "io"

  "gopkg.in/yaml.v3"
)

func LoadVars(path string) map[interface{}]interface{} {
	y := make(map[interface{}]interface{})

	f, err := os.Open(path)
	if err != nil {
    log.Fatalf("error: %v", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
  if err != nil {
    log.Fatalf("error: %v", err)
  }

	err = yaml.Unmarshal(data, &y)
	if err != nil {
    log.Fatalf("error: %v", err)
	}

  return y
}
