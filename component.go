package main

type Component struct {
	Name     string
	Props    []Prop
	Content  string
	Package  string
	Children *Component
}
