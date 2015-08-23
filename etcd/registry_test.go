package etcd

import (
	"encoding/json"
	"testing"

	"github.com/coreos/go-etcd/etcd"
	"github.com/st3v/discovery"
	"github.com/st3v/discovery/etcd/fake"
)

func TestRegister(t *testing.T) {
	var (
		namespace = "namespace"
		machines  = []string{"foo", "bar"}
		ttl       = uint64(1234)
		instance  = discovery.Instance{
			ID:     "id",
			Name:   "name",
			Env:    "env",
			Region: "region",
			Host:   "host",
			Port:   987,
		}
		path   = "namespace/region/env/name/id"
		json   = instance.String()
		client = new(fake.EtcdClient)
	)

	oldEtcdClient := injectClient(client, machines, t)
	defer func() { etcdClient = oldEtcdClient }()

	registry := NewRegistry(machines, namespace)

	err := registry.Register(instance, ttl)
	if want, have := error(nil), err; want != have {
		t.Fatalf("want %v, have %v", want, have)
	}

	if want, have := 1, client.SetCallCount(); want != have {
		t.Fatalf("want %v, have %v", want, have)
	}

	actualPath, actualJSON, actualTTL := client.SetArgsForCall(0)

	if want, have := path, actualPath; want != have {
		t.Fatalf("want %v, have %v", want, have)
	}

	if want, have := json, actualJSON; want != have {
		t.Fatalf("want %v, have %v", want, have)
	}

	if want, have := ttl, actualTTL; want != have {
		t.Fatalf("want %v, have %v", want, have)
	}
}

func TestLookupPath(t *testing.T) {
	var (
		namespace = "namespace"
		machines  = []string{"foo", "bar"}
		client    = new(fake.EtcdClient)

		pathTests = []struct {
			name       string
			env        string
			region     string
			lookupPath string
		}{
			{"name", "env", "region", "namespace/region/env/name"},
			{"name", "env", "", "namespace"},
			{"name", "", "region", "namespace/region"},
			{"", "env", "region", "namespace/region/env"},
		}
	)

	payload := `{
		"action": "get",
		"node": {
			"key": "/namespace/region",
			"dir": true,
			"nodes": [
				{
					"key": "/namespace/region/env",
					"dir": true,
					"nodes": [
						{
							"key": "/namespace/region/env/name",
							"dir": true,
							"nodes": [
								{
									"key": "/namespace/region/env/name/id",
									"value": "{\"name\":\"name\",\"id\":\"id\",\"region\":\"region\",\"env\":\"env\",\"host\":\"host\",\"port\":987}",
									"expiration": "2015-08-23T19:22:45.549917226Z",
									"ttl": 79,
									"modifiedIndex": 214,
									"createdIndex": 214
								}
							],
							"modifiedIndex": 157,
							"createdIndex": 157
						}
					],
					"modifiedIndex": 10,
					"createdIndex": 10
				}
			],
			"modifiedIndex": 10,
			"createdIndex": 10
		}
	}`

	var resp = new(etcd.Response)
	if want, have := error(nil), json.Unmarshal([]byte(payload), resp); want != have {
		t.Fatalf("want %v, have %v", want, have)
	}
	client.GetReturns(resp, nil)

	oldEtcdClient := injectClient(client, machines, t)
	defer func() { etcdClient = oldEtcdClient }()

	registry := NewRegistry(machines, namespace)

	for i, test := range pathTests {
		instances, err := registry.Lookup(test.name, test.env, test.region)
		if want, have := error(nil), err; want != have {
			t.Fatalf("want %v, have %v", want, have)
		}

		actualPath, actualSort, actualRecursive := client.GetArgsForCall(i)

		if want, have := test.lookupPath, actualPath; want != have {
			t.Fatalf("want %v, have %v", want, have)
		}

		if want, have := false, actualSort; want != have {
			t.Fatalf("want %v, have %v", want, have)
		}

		if want, have := true, actualRecursive; want != have {
			t.Fatalf("want %v, have %v", want, have)
		}

		if want, have := 1, len(instances); want != have {
			t.Fatalf("want %v, have %v", want, have)
		}

		instance := instances[0]

		if want, have := "region", instance.Region; want != have {
			t.Fatalf("want %v, have %v", want, have)
		}

		if want, have := "env", instance.Env; want != have {
			t.Fatalf("want %v, have %v", want, have)
		}

		if want, have := "name", instance.Name; want != have {
			t.Fatalf("want %v, have %v", want, have)
		}

		if want, have := "id", instance.ID; want != have {
			t.Fatalf("want %v, have %v", want, have)
		}
	}
}

func injectClient(client EtcdClient, machines []string, t *testing.T) func([]string) EtcdClient {
	oldEtcdClient := etcdClient
	etcdClient = func(etcdMachines []string) EtcdClient {
		if want, have := len(machines), len(etcdMachines); want != have {
			t.Fatalf("want %v, have %v", want, have)
		}

		for i, m := range etcdMachines {
			if want, have := machines[i], m; want != have {
				t.Fatalf("want %v, have %v", want, have)
			}
		}

		return client
	}
	return oldEtcdClient
}
