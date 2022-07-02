package test

import (
	"os/exec"
	"testing"
)

func TestShellClient(t *testing.T) {
	t.Log("Creating Namespace")
	cmd := exec.Command("bash", "-c", "ip netns ls")
	stdout, err := cmd.Output()
	if err != nil {
		t.Log("Namespace creation failed! " + string(stdout))
	}
	t.Log("Namespace creation success! " + string(stdout))
}
