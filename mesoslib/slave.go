package mesoslib

import (
	"encoding/json"
	"net/http"
)

type Resources struct {
	Cpus   float64 `json:"cpus,omitempty"`
	Memory int     `json:"mem,omitempty"`
	Disk   int     `json:"disk,omitempty"`
}

type SlaveInfo struct {
	ID         string            `json:"id,omitempty"`
	Hostname   string            `json:"hostname,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Port       int               `json:"port,omitempty"`
	Resources  *Resources        `json:"resources,omitempty"`
}

func (m *MesosLib) Slaves() ([]*SlaveInfo, error) {
	resp, err := http.Get("http://" + m.master + "/master/state.json")
	if err != nil {
		return nil, err
	}

	data := struct {
		Slaves []*SlaveInfo
	}{}

	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	resp.Body.Close()

	return data.Slaves, nil
}
