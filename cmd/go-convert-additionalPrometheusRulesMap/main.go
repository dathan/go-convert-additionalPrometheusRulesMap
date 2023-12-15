package main

import (
	"fmt"
	"log"
	"os"

	//"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v2"
)

type PrometheusRules struct {
	KubePrometheusStack struct {
		AdditionalPrometheusRulesMap map[string]Groups `yaml:"additionalPrometheusRulesMap"`
	} `yaml:"kube-prometheus-stack"`
}

type Groups struct {
	Groups []Group `yaml:"groups"`
}

type Group struct {
	Name  string `yaml:"name"`
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	Alert       string            `yaml:"alert,omitempty"`
	Expr        string            `yaml:"expr"`
	For         string            `yaml:"for,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type RuleConvert struct {
	Name  string `yaml:"name"`
	Rules []Rule `yaml:"rules"`
}

type PrometheusRuleConfig struct {
	PrometheusRule PrometheusRule `yaml:"prometheusRule"`
}

type PrometheusRule struct {
	Enabled          bool              `yaml:"enabled"`
	Name             string            `yaml:"name"`
	AdditionalLabels map[string]string `yaml:"additionalLabels"`
	Namespace        string            `yaml:"namespace"`
	Rules            []RuleConvert     `yaml:"rules"`
}

func main() {
	file, err := os.ReadFile("prod.yaml")
	if err != nil {
		panic(err)
	}

	var rules PrometheusRules
	if err := yaml.Unmarshal(file, &rules); err != nil {
		panic(err)
	}

	//spew.Dump(rules)
	rulesMap := make(map[string]map[string][]Rule)

	for _, group := range rules.KubePrometheusStack.AdditionalPrometheusRulesMap {
		for _, ruleGroup := range group.Groups {
			if _, exists := rulesMap[ruleGroup.Name]; !exists {
				// If the category doesn't exist, create it
				rulesMap[ruleGroup.Name] = make(map[string][]Rule)
			}
			for _, rule := range ruleGroup.Rules {
				if len(rule.Labels) > 0 {
					if _, exists := rule.Labels["owner"]; exists {
						rulesMap[ruleGroup.Name][rule.Labels["owner"]] = append(rulesMap[ruleGroup.Name][rule.Labels["owner"]], rule)
					} else {
						// set the default to compute-region
						rulesMap[ruleGroup.Name]["compute-region"] = append(rulesMap[ruleGroup.Name]["compute-region"], rule)
					}
				} else {
					// if we do not have labels set the default owner to compute-region
					rulesMap[ruleGroup.Name]["compute-region"] = append(rulesMap[ruleGroup.Name]["compute-region"], rule)
				}

			}
		}
	}
	rulesOutput := make(map[string][]RuleConvert)
	// rulesMap is a group of ruleGroups by Owner
	for groupName, ownerRules := range rulesMap {
		for owner, rules := range ownerRules {
			if _, exists := rulesOutput[owner]; !exists {
				rulesOutput[owner] = []RuleConvert{{Name: groupName, Rules: rules}}
			} else {
				rulesOutput[owner] = append(rulesOutput[owner], RuleConvert{groupName, rules})
			}
		}
	}

	//driver
	var config PrometheusRuleConfig
	team := "cloud-sre-team"
	config.PrometheusRule.Name = team
	config.PrometheusRule.Enabled = true
	config.PrometheusRule.Namespace = "prometheus"
	config.PrometheusRule.Rules = rulesOutput[team]
	// Marshal to YAML
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(string(yamlData))

}
