package parser

import (
	"regexp"
)

var removeTag = "<REMOVE>"

// RepositoryName tbd
func RepositoryName(repositoryNameTemplate string, tagTemplates []string, templateVars TemplateVars) []string {
	names := make([]string, len(tagTemplates))

	templateVarsWithRemoveTag := make(map[string]string)
	for k, v := range templateVars {
		templateVarsWithRemoveTag[k] = v

		if v == "" && (k == "REGISTRY" || k == "NAMESPACE") {
			templateVarsWithRemoveTag[k] = removeTag
		}
	}

	for i, tagTemplate := range tagTemplates {
		name := regexp.MustCompile(removeTag+`\/`).ReplaceAllString(Template(repositoryNameTemplate, templateVarsWithRemoveTag), "")
		tag := Template(tagTemplate, templateVars)

		names[i] = name + ":" + tag
	}

	return names
}
