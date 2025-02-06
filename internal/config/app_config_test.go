package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	type want struct {
		address              string
		dbURI                string
		accrualSystemAddress string
	}
	tests := []struct {
		args []string
		envs map[string]string
		want want
	}{
		{
			args: []string{"-a", "localhost:8080", "-d", "postgres:tra-ta-ta", "-r", "localhost:3333"},
			envs: map[string]string{
				"RUN_ADDRESS":            "localhost:9090",
				"DATABASE_URI":           "postgres:ko-ko-ko",
				"ACCRUAL_SYSTEM_ADDRESS": "localhost:9090",
			},
			want: want{
				address:              "localhost:9090",
				dbURI:                "postgres:ko-ko-ko",
				accrualSystemAddress: "localhost:9090",
			},
		},
		{
			args: []string{"-a", "localhost:8080", "-d", "postgres:tra-ta-ta", "-r", "localhost:3333"},
			envs: map[string]string{},
			want: want{
				address:              "localhost:8080",
				dbURI:                "postgres:tra-ta-ta",
				accrualSystemAddress: "localhost:3333",
			},
		},
	}

	for _, test := range tests {
		for k, v := range test.envs {
			os.Setenv(k, v)
		}
		os.Args = append([]string{"cmd"}, test.args...)

		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		config, err := NewAppConfig()
		os.Clearenv()

		assert.Equal(t, test.want.address, config.RunAddress)
		assert.NoError(t, err)
	}
}

func TestEmptyFlags(t *testing.T) {
	tests := []struct {
		args           []string
		expectedErrors []error
		expectedMsg    string
	}{
		{
			args:           []string{"-a", "localhost:8080", "-d", "postgres:tra-ta-ta"},
			expectedErrors: []error{ErrMissingAccrualSystemAddres},
			expectedMsg:    "missing accrual system address",
		},
		{
			args:           []string{"-a", "localhost:8080"},
			expectedErrors: []error{ErrMissingDDURI, ErrMissingAccrualSystemAddres},
			expectedMsg:    "missing database URI\nmissing accrual system address",
		},
		{
			args:           []string{},
			expectedErrors: []error{ErrMissingAddress, ErrMissingDDURI, ErrMissingAccrualSystemAddres},
			expectedMsg:    "missing run address\nmissing database URI\nmissing accrual system address",
		},
	}

	for _, test := range tests {
		os.Args = append([]string{"cmd"}, test.args...)

		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		_, err := NewAppConfig()
		os.Clearenv()

		for _, expectedError := range test.expectedErrors {
			assert.ErrorIs(t, err, expectedError)
		}
		assert.EqualError(t, err, test.expectedMsg)
	}
}
