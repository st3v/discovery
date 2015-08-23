package main

import (
	"fmt"
	"log"
	"os"

	"github.com/st3v/discovery"
	"github.com/st3v/discovery/etcd"
)

func main() {
	registry := etcd.NewRegistry(os.Args[1:], "cfkit")

	instances := []discovery.Instance{
		{
			ID:     "1",
			Name:   "foo",
			Env:    "production",
			Region: "us-west",
			Host:   "app-1.foo.io",
			Port:   123,
		},
		{
			ID:     "2",
			Name:   "foo",
			Env:    "production",
			Region: "us-west",
			Host:   "app-2.foo.io",
			Port:   984,
		},
		{
			ID:     "1",
			Name:   "bar",
			Env:    "production",
			Region: "us-west",
			Host:   "app-1.bar.io",
			Port:   123,
		},
		{
			ID:     "1",
			Name:   "foo",
			Env:    "production",
			Region: "us-east",
			Host:   "app-3.foo.io",
			Port:   123,
		},
	}

	for _, i := range instances {
		if err := registry.Register(i, 100); err != nil {
			log.Fatalf("Error registering instance: %s", err)
		}
	}

	result, err := registry.Lookup("foo", "", "us-west")
	if err != nil {
		log.Fatalf("Error looking up instances: %s", err)
	}

	for _, i := range result {
		fmt.Println(i)
	}

}
