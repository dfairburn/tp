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

	for _, resource := range export.Resources {

		if resource.Type == "environment" {
			// TODO: convert env json into yaml (probably a lib that can do this)
			fmt.Println("=========================")
			fmt.Printf("type: %s\n", resource.Type)
			fmt.Printf("data: %s\n", resource.Data)
			fmt.Println("=========================")
		} else if resource.Type == "request" {
			fmt.Println("=========================")
			fmt.Printf("url: %s\n", resource.Url)
			fmt.Printf("method: %s\n", resource.Url)
			fmt.Printf("headers: %s\n", resource.Headers)
			fmt.Println("=========================")
		} else if resource.Type == "request_group" {
			fmt.Println("=========================")
			fmt.Printf("service name: %s\n", resource.Name)
			fmt.Println("=========================")
		}
	}

	return nil, nil
}
