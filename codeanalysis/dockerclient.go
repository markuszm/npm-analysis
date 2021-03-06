package codeanalysis

import (
	"bytes"
	"context"
	"github.com/dustinkirkland/golang-petname"
	"github.com/pkg/errors"
	"math/rand"
	"os/exec"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func BuildImage(contextPath, tag string) error {
	cmd := exec.Command("docker", "build", "-t", tag, contextPath)

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()

	if errOut.String() != "" {
		return errors.New(errOut.String())
	}
	return err
}

func RunDockerContainer(tag string, arguments ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	containerName := petname.Generate(2, "-")

	cmdArgs := append([]string{"run", "--rm", "--name", containerName, tag}, arguments...)
	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		// no error handling as if this fails there is nothing we can do
		cmd := exec.Command("docker", "stop", containerName)
		cmd.Run()

		cmd = exec.Command("docker", "rm", containerName)
		cmd.Run()
		return "", errors.New("command timed out")
	}

	if err != nil {
		return "", errors.Wrapf(err, errOut.String())
	}
	return out.String(), nil
}
