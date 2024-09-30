package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"sync"
)

type Prop struct {
	Name string
	Type string
}

type Interpreter struct {
	components map[string]*Component
	mutex      sync.RWMutex
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		components: make(map[string]*Component),
	}
}

func (i *Interpreter) ParseFile(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	var currentComponent *Component
	var contentBuilder strings.Builder
	var packageName string

	packageRegex := regexp.MustCompile(`^package\s+(\w+)`)
	componentRegex := regexp.MustCompile(`^templ\s+(\w+)\((.*?)\)\s*{`)
	childrenRegex := regexp.MustCompile(`{\s*children...\s*}`)

	for scanner.Scan() {
		line := scanner.Text()
		if match := packageRegex.FindStringSubmatch(line); match != nil {
			packageName = match[1]
		} else if match := componentRegex.FindStringSubmatch(line); match != nil {
			if currentComponent != nil {
				currentComponent.Content = contentBuilder.String()
				i.mutex.Lock()
				i.components[currentComponent.Package+"."+currentComponent.Name] = currentComponent
				i.mutex.Unlock()
			}

			//fmt.Println("match", parseProps(match[2]))

			currentComponent = &Component{
				Name:    match[1],
				Props:   parseProps(match[2]),
				Package: packageName,
			}
			contentBuilder.Reset()
		} else if currentComponent != nil {
			if strings.TrimSpace(line) == "}" && contentBuilder.Len() > 0 && strings.Count(contentBuilder.String(), "{") == strings.Count(contentBuilder.String(), "}") {
				currentComponent.Content = contentBuilder.String()
				if childrenRegex.MatchString(currentComponent.Content) {
					currentComponent.Children = &Component{}
				}
				i.mutex.Lock()
				fmt.Println(currentComponent.Package + "." + currentComponent.Name)
				i.components[currentComponent.Package+"."+currentComponent.Name] = currentComponent
				i.mutex.Unlock()
				currentComponent = nil
				contentBuilder.Reset()
			} else {
				contentBuilder.WriteString(line + "\n")
			}
		}
	}

	if currentComponent != nil {
		currentComponent.Content = contentBuilder.String()
		if childrenRegex.MatchString(currentComponent.Content) {
			currentComponent.Children = &Component{}
		}
		i.mutex.Lock()
		i.components[currentComponent.Package+"."+currentComponent.Name] = currentComponent
		i.mutex.Unlock()
	}

	return scanner.Err()
}

func parseProps(propsString string) []Prop {
	props := []Prop{}
	if propsString == "" {
		return props
	}
	propPairs := strings.Split(propsString, ",")
	for _, pair := range propPairs {
		parts := strings.Fields(pair)
		if len(parts) == 2 {
			props = append(props, Prop{Name: parts[0], Type: parts[1]})
		}
	}
	return props
}
