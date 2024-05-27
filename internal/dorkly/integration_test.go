package dorkly

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/launchdarkly/go-server-sdk-evaluation/v3/ldmodel"
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
// It serves as a partial integration test for the dorkly project.
func TestDocker_LdRelayCanLoadArchive(t *testing.T) {
	ctx := context.Background()
	containerReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../../Docker/",
			Dockerfile: "Dockerfile",
			Repo:       "dorkly-testcontainers",
			Tag:        "LdRelayCanLoadArchive",
		},
		ExposedPorts: []string{"8030/tcp"},
		Env: map[string]string{
			"LOG_LEVEL": "debug",
			"S3_URL":    "required to be non-empty but ok to be this bogus value so we can ensure resiliency if S3 connection fails",
		},
		Files: []testcontainers.ContainerFile{{
			HostFilePath:      "testdata/flags.tar.gz",
			ContainerFilePath: "/dorkly/flags.tar.gz",
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

// StatusRep is the JSON representation returned by the status endpoint.
type StatusRep struct {
	Environments  map[string]interface{} `json:"environments"`
	Status        string                 `json:"status"`
	Version       string                 `json:"version"`
	ClientVersion string                 `json:"clientVersion"`
}
