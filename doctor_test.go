package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDoctorOptions(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantOpts doctorOptions
		wantErr  bool
	}{
		{
			name:     "no args",
			args:     []string{},
			wantOpts: doctorOptions{},
		},
		{
			name:     "fix flag",
			args:     []string{"--fix"},
			wantOpts: doctorOptions{fix: true},
		},
		{
			name:    "help",
			args:    []string{"--help"},
			wantErr: true,
		},
		{
			name:    "unknown arg",
			args:    []string{"--unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := parseDoctorOptions(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOpts, opts)
		})
	}
}

func TestPrintDoctorUsage(t *testing.T) {
	var buf bytes.Buffer
	printDoctorUsage(&buf)
	output := buf.String()

	assert.Contains(t, output, "relay doctor [--fix]")
	assert.Contains(t, output, "Diagnose")
}

func TestRunDoctorCommand(t *testing.T) {
	ui := cliUI{color: false}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := runDoctorCommand([]string{}, stdout, stderr, ui)

	assert.True(t, code == 0 || code == 1, "doctor command should return 0 or 1")
	output := stdout.String()
	assert.Contains(t, output, "relay doctor")

	if strings.Contains(output, "relay binary") {
		assert.True(t, strings.Contains(output, "+") || strings.Contains(output, "-"))
	}
}
