package shclient

import "os/exec"

type shClient struct {

}

func NewSH() *shClient {
	return &shClient{

	}
}

func (c *shClient) Exec(name string, arg ...string) (string, error) {
	output, err := exec.Command(name, arg...).CombinedOutput()
	return string(output), err
}

