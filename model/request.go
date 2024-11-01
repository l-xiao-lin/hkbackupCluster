package model

type ParamSSHConfig struct {
	Host    string `json:"host"`
	User    string `json:"user"`
	Port    int    `json:"port"`
	KeyPath string `json:"keyPath"`
	Command string `json:"command"`
}
