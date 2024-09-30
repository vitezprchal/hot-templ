package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

func (i *Interpreter) Render(componentName string, props map[string]string) (string, error) {
	i.mutex.RLock()
	comp, ok := i.components[componentName]
	i.mutex.RUnlock()

	if !ok {
		return "", fmt.Errorf("component %s not found", componentName)
	}

	content := comp.Content

	// @component calls
	componentCallRegex := regexp.MustCompile(`@((?:\w+\.)?(?:\w+))\(((?:[^()]|\((?:[^()]|\([^()]*\))*\))*)\)\s*({)?((?:[^{}]|{[^{}]*})*)(})?`)
	content = componentCallRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := componentCallRegex.FindStringSubmatch(match)

		if len(submatches) < 2 {
			return match
		}
		childName := submatches[1]
		if !strings.Contains(childName, ".") {
			childName = comp.Package + "." + childName
		}
		childArgs := parseComponentArgs(submatches[2])
		childContent := ""
		if len(submatches) > 4 && submatches[4] != "" {
			childContent = submatches[4]
		}

		childProps := make(map[string]string)
		for i, arg := range childArgs {

			childProps[fmt.Sprintf("arg%d", i)] = arg
		}

		for k, v := range props {
			if _, ok := childProps[k]; !ok {
				childProps[k] = v
			}
		}

		childProps["children"] = childContent

		renderedChild, err := i.Render(childName, childProps)
		if err != nil {
			log.Printf("Error rendering %s: %v", childName, err)
			return fmt.Sprintf("<!-- Error rendering %s -->", childName)
		}
		return renderedChild
	})

	// apply props
	propRegex := regexp.MustCompile(`{(?:\s*)([a-zA-Z0-9_]+(?:\.[a-zA-Z0-9_]+)*)(?:\s*)}`)
	content = propRegex.ReplaceAllStringFunc(content, func(match string) string {
		propName := propRegex.FindStringSubmatch(match)[1]
		if value, ok := props[propName]; ok {
			return value
		}
		return match
	})

	// children
	childrenRegex := regexp.MustCompile(`{(?:\s*)children(?:\.{3})?(?:\s*)}`)
	if childrenRegex.MatchString(content) {
		if childContent, ok := props["children"]; ok {
			content = childrenRegex.ReplaceAllString(content, childContent)
		} else {
			content = childrenRegex.ReplaceAllString(content, "")
		}
	}

	return content, nil
}

func parseComponentArgs(argsString string) []string {
	var args []string
	re := regexp.MustCompile(`"([^"]*)"|(\S+)`)
	matches := re.FindAllStringSubmatch(argsString, -1)

	for _, match := range matches {
		if match[1] != "" {
			args = append(args, match[1])
		} else {
			args = append(args, match[2])
		}
	}
	return args
}
