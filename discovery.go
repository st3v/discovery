package discovery

import "encoding/json"

type Instance struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Region string `json:"region"`
	Env    string `json:"env"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
}

type Registry interface {
	Lookup(name, env, region string) ([]Instance, error)
	Register(Instance) error
}

func (i Instance) String() string {
	b, _ := json.Marshal(i)
	return string(b)
}
