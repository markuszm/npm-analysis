package codeanalysis

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"os/exec"
	"time"
)

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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	cmdArgs := append([]string{"run", "--rm", tag}, arguments...)
	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		return "", errors.New("command timed out")
	}

	if err != nil {
		return "", errors.Wrapf(err, errOut.String())
	}
	return out.String(), nil
}
