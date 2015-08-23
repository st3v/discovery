package etcd

import (
	"encoding/json"
	"strings"

	etcderr "github.com/coreos/etcd/error"
	"github.com/coreos/go-etcd/etcd"

	"github.com/st3v/discovery"
)

const (
	wildcard  = "*"
	separator = "/"
)

type EtcdClient interface {
	Set(key string, value string, ttl uint64) (*etcd.Response, error)
	Get(key string, sort, recursive bool) (*etcd.Response, error)
}

type registry struct {
	namespace string
	client    EtcdClient
}

var etcdClient = func(machines []string) EtcdClient {
	return EtcdClient(etcd.NewClient(machines))
}

func NewRegistry(machines []string, namespace string) *registry {
	return &registry{
		namespace: namespace,
		client:    etcdClient(machines),
	}
}

func (r *registry) Register(i discovery.Instance, ttl uint64) error {
	_, err := r.client.Set(
		r.storePath(i),
		i.String(),
		ttl,
	)
	return err
}

func errKeyNotFound(err error) bool {
	if etcdErr, ok := err.(*etcd.EtcdError); ok {
		return etcdErr.ErrorCode == etcderr.EcodeKeyNotFound
	}
	return false
}

func (r *registry) Lookup(name, env, region string) ([]discovery.Instance, error) {
	instances := []discovery.Instance{}

	key := r.lookupPath(name, env, region)

	resp, err := r.client.Get(key, false, true)
	if err != nil {
		if errKeyNotFound(err) {
			return []discovery.Instance{}, nil
		}
		return instances, err
	}

	return lift(resp.Node, r.wildcardPath(name, env, region), 0)
}

func (r *registry) storePath(i discovery.Instance) string {
	path := path("", false, r.namespace, i.Region, i.Env, i.Name, i.ID)
	return strings.Join(path, separator)
}

func (r *registry) lookupPath(name, env, region string) string {
	path := path("", true, r.namespace, region, env, name)
	return strings.Join(path, separator)
}

func (r *registry) wildcardPath(name, env, region string) []string {
	return path(wildcard, false, "", r.namespace, region, env, name, wildcard)
}

func path(def string, skip bool, parts ...string) []string {
	var path []string

	for _, s := range parts {
		if s == "" {
			if skip {
				break
			} else if def != "" {
				s = def
			}
		}
		path = append(path, s)
	}

	return path

}

func lift(node *etcd.Node, path []string, depth int) ([]discovery.Instance, error) {
	result := []discovery.Instance{}

	for i, s := range strings.Split(node.Key, separator) {
		if path[i] != s && path[i] != wildcard {
			return result, nil
		}
	}

	if node.Dir {
		for _, n := range node.Nodes {
			instances, err := lift(n, path, depth+1)
			if err != nil {
				return []discovery.Instance{}, err
			}
			result = append(result, instances...)
		}

		return result, nil
	}

	instance := discovery.Instance{}
	if err := json.Unmarshal([]byte(node.Value), &instance); err != nil {
		return []discovery.Instance{}, err
	}

	result = append(result, instance)

	return result, nil
}
