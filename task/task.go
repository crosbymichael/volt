package task

import (
	"github.com/VoltFramework/volt/mesoslib"
	"github.com/VoltFramework/volt/mesosproto"
)

type Task struct {
	ID          string   `json:"id,omitempty"`
	Command     string   `json:"cmd,omitempty"`
	Cpus        float64  `json:"cpus,omitempty"`
	Disk        float64  `json:"disk,omitempty"`
	Mem         float64  `json:"mem,omitempty"`
	Files       []string `json:"files,omitempty"`
	DockerImage string   `json:"docker_image,omitempty"`

	SlaveId *string               `json:"slave_id,omitempty"`
	State   *mesosproto.TaskState `json:"state,omitempty"`
	Volumes []*mesoslib.Volume    `json:"volumes,omitempty"`
}
