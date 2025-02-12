package rules

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v2"
)

type Units int

const (
	Second Units = iota
	Minute
	Hour
)

var UnitName = map[Units]string{
	Second: "seconds",
	Minute: "minutes",
	Hour  : "hour",
}

type StoredRule struct {
	Domain		string	`yaml:"domain"`
	Method		string	`yaml:"method"`
	Key		string	`yaml:"key"`
	Value		string	`yaml:"value"`
	MaxRequests	int	`yaml:"maxRequests"`
	Unit		string	`yaml:"unit"`
}

func FetchRules() StoredRule {
	yamlFile, err := os.ReadFile("rules/rules.yaml")
	if err != nil {
		fmt.Println(err)
	}
	var storedRules StoredRule

	err = yaml.Unmarshal(yamlFile, &storedRules)
	fmt.Println(err)
	fmt.Println(storedRules)
	return storedRules
}
