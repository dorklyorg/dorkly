package dorkly

import (
	"errors"
	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
	"github.com/launchdarkly/go-sdk-common/v3/ldlog"
	"github.com/launchdarkly/go-server-sdk/v7"
	"github.com/launchdarkly/go-server-sdk/v7/ldcomponents"
	"go.uber.org/zap"
	"reflect"
	"time"
)

type updateMonitor struct {
	logger               *zap.SugaredLogger
	endpoint             string
	envName              string
	sdkKey               string
	expectedFlagVersions map[string]int

	actualFlagVersions map[string]int
	ldClient           *ldclient.LDClient
}

func newUpdateMonitor(endpoint string, envName string, sdkKey string, expectedFlagVersions map[string]int) (*updateMonitor, error) {
	l := logger.Named("updateMonitor").
		With(zap.String("endpoint", endpoint), zap.String("envName", envName), zap.String("sdkKey", sdkKey)).
		With("expectedFlagVersions", expectedFlagVersions)

	ldConfig := ldclient.Config{
		ServiceEndpoints: ldcomponents.RelayProxyEndpoints("https://dorkly-example-test.mbe39aim2pgh2.us-west-2.cs.amazonlightsail.com/"),
		Events:           ldcomponents.NoEvents(),
		Logging:          ldcomponents.Logging().LogEvaluationErrors(true).MinLevel(ldlog.Debug),
	}

	ldClient, err := ldclient.MakeCustomClient(sdkKey, ldConfig, 30*time.Second)

	if err != nil {
		l.Errorf("Failed to create LaunchDarkly client: %v", err)
		return nil, err
	}

	return &updateMonitor{
		logger:               l,
		endpoint:             endpoint,
		envName:              envName,
		sdkKey:               sdkKey,
		expectedFlagVersions: expectedFlagVersions,
		actualFlagVersions:   make(map[string]int),
		ldClient:             ldClient,
	}, nil
}

func (m *updateMonitor) awaitExpectedFlagVersions(maxWaitDuration time.Duration) error {
	tick := time.NewTicker(100 * time.Millisecond)
	done := time.After(maxWaitDuration)
	for {
		select {
		case <-done:
			m.logger.With("actualFlagVersions", m.actualFlagVersions).Errorf("Timed out waiting for expected flag versions")
			return errors.New("timed out waiting for expected flag versions. See log for details")
		case <-tick.C:
			allFlags := m.ldClient.AllFlagsState(ldcontext.NewBuilder("example-user-key").Build())
			valuesMap := allFlags.ToValuesMap()
			for actualFlagKey := range valuesMap {
				flagState, exists := allFlags.GetFlag(actualFlagKey)
				if exists {
					m.actualFlagVersions[actualFlagKey] = flagState.Version
				}
			}
			if reflect.DeepEqual(m.expectedFlagVersions, m.actualFlagVersions) {
				m.logger.Infof("All expected flag versions found")
				return nil
			}
		}
	}
}
