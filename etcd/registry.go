package etcd

import (
	"encoding/json"
	"path/filepath"

	"github.com/VoltFramework/volt/task"
	"github.com/coreos/go-etcd/etcd"
)

type Registry struct {
	client *etcd.Client
}

func New(machines []string) *Registry {
	return &Registry{
		client: etcd.NewClient(machines),
	}
}

func (r *Registry) Register(id string, t *task.Task) error {
	data, err := r.marshal(t)
	if err != nil {
		return err
	}

	_, err = r.client.Set(r.key(id), data, 0)

	return err
}

func (r *Registry) Fetch(id string) (*task.Task, error) {
	resp, err := r.client.Get(r.key(id), false, false)
	if err != nil {
		return nil, err
	}

	var t *task.Task
	if err := r.unmarshal(resp.Node.Value, &t); err != nil {
		return nil, err
	}

	return t, nil
}

func (r *Registry) Tasks() ([]*task.Task, error) {
	resp, err := r.client.Get("/volt/tasks", true, true)
	if err != nil {
		return nil, err
	}

	out := []*task.Task{}

	for _, n := range resp.Node.Nodes {
		var t *task.Task
		if err := r.unmarshal(n.Value, &t); err != nil {
			return nil, err
		}

		out = append(out, t)
	}

	return out, nil
}

func (r *Registry) Update(id string, t *task.Task) error {
	data, err := r.marshal(t)
	if err != nil {
		return err
	}

	_, err = r.client.Update(r.key(id), data, 0)

	return err
}

func (r *Registry) Delete(id string) error {
	_, err := r.client.Delete(r.key(id), true)
	return err
}

func (r *Registry) key(id string) string {
	return filepath.Join("/volt/tasks", id)
}

func (r *Registry) marshal(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (r *Registry) unmarshal(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}
