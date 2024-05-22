package dorkly

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/launchdarkly/go-sdk-common/v3/ldtime"
	"github.com/launchdarkly/go-server-sdk-evaluation/v3/ldmodel"
	"github.com/launchdarkly/go-server-sdk/v7/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"log"
	"net/http"
	"testing"
	"time"
)

// This test ensures the ld-relay can load a dorkly-generated archive and serve flags from it.
func TestIntegration_LdRelayCanLoadArchive(t *testing.T) {
	ctx := context.Background()
	containerFlagsArchivePath := "/dorkly/flags.tar.gz"
	containerReq := testcontainers.ContainerRequest{
		Image:        "launchdarkly/ld-relay:8.4.2", // tag should be kept in sync with docker/Dockerfile
		ExposedPorts: []string{"8030/tcp"},
		Env: map[string]string{
			"FILE_DATA_SOURCE": containerFlagsArchivePath,
		},
		Files: []testcontainers.ContainerFile{{
			HostFilePath:      "testdata/flags.tar.gz",
			ContainerFilePath: containerFlagsArchivePath,
			FileMode:          0755,
		}},
		WaitingFor: wait.ForLog("Starting server listening on port 8030"),
	}
	relayContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: containerReq,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}
	defer func() {
		if err := relayContainer.Terminate(ctx); err != nil {
			log.Fatalf("Could not stop container: %s", err)
		}
	}()

	containerBaseUrl, err := relayContainer.Endpoint(ctx, "http")
	require.Nil(t, err)
	t.Logf("Container URL: %s", containerBaseUrl)

	t.Run("status", func(t *testing.T) {
		ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		defer cancelFunc()
		// Check if the relay is healthy and contains expected environments
		request, err := http.NewRequestWithContext(ctx, "GET", containerBaseUrl+"/status", nil)
		require.Nil(t, err)

		response, err := http.DefaultClient.Do(request)
		require.Nil(t, err)

		defer response.Body.Close()
		require.Equal(t, http.StatusOK, response.StatusCode)

		body, err := io.ReadAll(response.Body)
		require.Nil(t, err)
		statusRep := StatusRep{}
		err = json.Unmarshal(body, &statusRep)
		require.Nil(t, err)

		assert.Equal(t, "healthy", statusRep.Status)
		assert.Len(t, statusRep.Environments, 2)
		assert.Contains(t, statusRep.Environments, "production")
		assert.Contains(t, statusRep.Environments, "staging")
	})

	testFlagsForEnv(t, ctx, containerBaseUrl+"/sdk/flags", "production")
	testFlagsForEnv(t, ctx, containerBaseUrl+"/sdk/flags", "staging")
}

func testFlagsForEnv(t *testing.T, ctx context.Context, url, env string) {
	t.Run(fmt.Sprintf("flags: %s", env), func(t *testing.T) {
		ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(10*time.Second))
		defer cancelFunc()
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		require.Nil(t, err)

		req.Header.Set("Authorization", insecureSdkKey(env))
		response, err := http.DefaultClient.Do(req)
		require.Nil(t, err)

		defer response.Body.Close()
		require.Equal(t, http.StatusOK, response.StatusCode)
		body, err := io.ReadAll(response.Body)
		require.Nil(t, err)

		actualFlags := make(map[string]ldmodel.FeatureFlag)
		err = json.Unmarshal(body, &actualFlags)
		require.Nil(t, err)

		relayArchive := testProject1.toRelayArchive()
		expectedFlags := relayArchive.envs[env].data.Flags

		// We need to set the version to 1 because that's always the first delivered version of a flag
		for key := range expectedFlags {
			f := expectedFlags[key]
			f.Version = 1
			expectedFlags[key] = f
		}

		assert.Equal(t, expectedFlags, actualFlags)
	})
}

// all structs below are copied from https://github.com/launchdarkly/ld-relay/blob/7d67cb6e8f4edd2e0ea5a20b7640a3d35a8e165d/internal/api/status_reps.go

// StatusRep is the JSON representation returned by the status endpoint.
//
// This is exported for use in integration test code.
type StatusRep struct {
	Environments  map[string]EnvironmentStatusRep `json:"environments"`
	Status        string                          `json:"status"`
	Version       string                          `json:"version"`
	ClientVersion string                          `json:"clientVersion"`
}

// EnvironmentStatusRep is the per-environment JSON representation returned by the status endpoint.
//
// This is exported for use in integration test code.
type EnvironmentStatusRep struct {
	SDKKey           string               `json:"sdkKey"`
	EnvID            string               `json:"envId,omitempty"`
	EnvKey           string               `json:"envKey,omitempty"`
	EnvName          string               `json:"envName,omitempty"`
	ProjKey          string               `json:"projKey,omitempty"`
	ProjName         string               `json:"projName,omitempty"`
	MobileKey        string               `json:"mobileKey,omitempty"`
	ExpiringSDKKey   string               `json:"expiringSdkKey,omitempty"`
	Status           string               `json:"status"`
	ConnectionStatus ConnectionStatusRep  `json:"connectionStatus"`
	DataStoreStatus  DataStoreStatusRep   `json:"dataStoreStatus"`
	BigSegmentStatus *BigSegmentStatusRep `json:"bigSegmentStatus,omitempty"`
}

// BigSegmentStatusRep is the big segment status representation returned by the status endpoint.
//
// This is exported for use in integration test code.
type BigSegmentStatusRep struct {
	Available          bool                       `json:"available"`
	PotentiallyStale   bool                       `json:"potentiallyStale"`
	LastSynchronizedOn ldtime.UnixMillisecondTime `json:"lastSynchronizedOn"`
}

// ConnectionStatusRep is the data source status representation returned by the status endpoint.
//
// This is exported for use in integration test code.
type ConnectionStatusRep struct {
	State      interfaces.DataSourceState `json:"state"`
	StateSince ldtime.UnixMillisecondTime `json:"stateSince"`
	LastError  *ConnectionErrorRep        `json:"lastError,omitempty"`
}

// ConnectionErrorRep is the optional error information in ConnectionStatusRep.
//
// This is exported for use in integration test code.
type ConnectionErrorRep struct {
	Kind interfaces.DataSourceErrorKind `json:"kind"`
	Time ldtime.UnixMillisecondTime     `json:"time"`
}

// DataStoreStatusRep is the data store status representation returned by the status endpoint.
//
// This is exported for use in integration test code.
type DataStoreStatusRep struct {
	State      string                     `json:"state"`
	StateSince ldtime.UnixMillisecondTime `json:"stateSince"`
	Database   string                     `json:"database,omitempty"`
	DBServer   string                     `json:"dbServer,omitempty"`
	DBPrefix   string                     `json:"dbPrefix,omitempty"`
	DBTable    string                     `json:"dbTable,omitempty"`
}
