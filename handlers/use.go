package handlers

import (
	"os"
  "fmt"
	"text/template"

  "github.com/dfairburn/tp/config"
  "github.com/dfairburn/tp/vars"
)

type Req struct {
	Token *string
	Name  string
}

// Use gets a filepath and "uses" that template
func Use(path string, config config.Config) error {
	// if no file given
	// if file given
	//t := "token"
	//r := Req{
		//Token: &t,
		//Name:  "Dan",
	//}

  fmt.Println(config.VariableDefinitionFile)
  v := vars.LoadVars(config.VariableDefinitionFile)


	tp := template.Must(template.ParseFiles(path))
	return tp.Execute(os.Stdout, v)
}
