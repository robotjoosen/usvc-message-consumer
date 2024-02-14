package config_test

import (
	"testing"

	"github.com/robotjoosen/usvc-messsage-consumer/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSettings struct {
	Name         string `mapstructure:"SERVICE_NAME"`
	LogLevel     string `mapstructure:"LOG_LEVEL"`
	SDPUsvcURL   string `mapstructure:"SDP_USVC_URL"`
	APIURL       string `mapstructure:"API_URL"`
	APIAuthToken string `mapstructure:"API_AUTH_TOKEN"`
}

func TestConfigLoad(t *testing.T) {
	testCases := map[string]struct {
		givenEnvVars            map[string]string
		withDefaults            map[string]any
		thenExpectError         bool
		thenExpectConfiguration testSettings
	}{
		"defaults are set": {
			givenEnvVars: map[string]string{},
			withDefaults: map[string]any{
				"SERVICE_NAME": "test",
				"LOG_LEVEL":    "INFO",
				"SDP_USVC_URL": "http://nil.test",
				"API_URL":      "http://nil.test",
			},
			thenExpectError: false,
			thenExpectConfiguration: testSettings{
				Name:       "test",
				LogLevel:   "INFO",
				SDPUsvcURL: "http://nil.test",
				APIURL:     "http://nil.test",
			},
		},
		"with env vars set or overwritten": {
			givenEnvVars: map[string]string{
				"LOG_LEVEL":      "DEBUG",
				"SDP_USVC_URL":   "http://sdp-usvc-url.test",
				"API_URL":        "http://api-url.test",
				"API_AUTH_TOKEN": "ABC123",
			},
			withDefaults: map[string]any{
				"SERVICE_NAME":   "",
				"LOG_LEVEL":      "INFO",
				"SDP_USVC_URL":   "http://nil.test",
				"API_URL":        "http://nil.test",
				"API_AUTH_TOKEN": "",
			},
			thenExpectError: false,
			thenExpectConfiguration: testSettings{
				Name:         "",
				LogLevel:     "DEBUG",
				SDPUsvcURL:   "http://sdp-usvc-url.test",
				APIURL:       "http://api-url.test",
				APIAuthToken: "ABC123",
			},
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			for envVar, envValue := range tc.givenEnvVars {
				t.Setenv(envVar, envValue)
			}

			cnf := testSettings{}
			_, err := config.Load(&cnf, tc.withDefaults)

			if !tc.thenExpectError {
				require.NoError(t, err)
				assert.Equal(t, tc.thenExpectConfiguration, cnf)
			} else {
				require.Error(t, err)
			}
		})
	}
}
