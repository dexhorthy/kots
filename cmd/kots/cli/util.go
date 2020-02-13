package cli

import (
	"os"
	"path/filepath"
	"strings"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	kubernetesConfigFlags *genericclioptions.ConfigFlags
)

func ExpandDir(input string) string {
	if !strings.HasPrefix(input, "~") {
		return input
	}

	return filepath.Join(homeDir(), input[1:])
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
