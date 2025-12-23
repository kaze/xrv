package cli

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	if cmd == nil {
		t.Fatal("NewRootCommand() returned nil")
	}

	if cmd.Use != "xrv" {
		t.Errorf("Use = %v, want xrv", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestRootCommandFlags(t *testing.T) {
	cmd := NewRootCommand()

	flags := []string{"config", "cache-dir", "debug"}

	for _, flag := range flags {
		if cmd.PersistentFlags().Lookup(flag) == nil {
			t.Errorf("Flag %s not defined", flag)
		}
	}
}

func TestRootCommandExecution(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}
