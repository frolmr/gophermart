package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSecretEnv(t *testing.T) {
	type want struct {
		JWTKey []byte
	}
	tests := []struct {
		envs map[string]string
		want want
	}{
		{
			envs: map[string]string{
				"JWT_SECRET": "zna40k_dura40k",
			},
			want: want{
				JWTKey: []byte("zna40k_dura40k"),
			},
		},
	}

	for _, test := range tests {
		for k, v := range test.envs {
			os.Setenv(k, v)
		}

		config, err := NewAuthConfig()
		os.Clearenv()

		assert.Equal(t, test.want.JWTKey, config.JWTKey)
		assert.NoError(t, err)
	}
}
