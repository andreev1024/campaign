package main

type User struct {
	Name    string            `json:"user"`
	Profile map[string]string `json:"profile"`
}
