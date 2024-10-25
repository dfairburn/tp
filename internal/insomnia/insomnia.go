package insomnia

import (
	"encoding/json"
	"fmt"
)

type Collection struct {
	Requests []Request
}

func ParseInsomniaExport(file []byte) (*Collection, error) {
	export := &Export{}
	err := json.Unmarshal(file, export)
	if err != nil {
		return nil, err
	}

	b, _ := json.Marshal(export)

	fmt.Printf("export: %+v\n", string(b))

	return nil, nil
}
