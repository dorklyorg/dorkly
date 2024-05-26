package dorkly

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"reflect"
)

type Reconciler struct {
	archiveService  RelayArchiveService
	secretsService  SecretsService
	projectYamlPath string
	logger          *zap.SugaredLogger
}

func NewReconciler(archiveService RelayArchiveService, secretsService SecretsService, projectYamlPath string) *Reconciler {
	return &Reconciler{
		archiveService:  archiveService,
		secretsService:  secretsService,
		projectYamlPath: projectYamlPath,
		logger:          logger.Named("Reconciler"),
	}
}

func (r *Reconciler) Reconcile(ctx context.Context) error {
	r.logger.
		With("archiveService", r.archiveService).
		With("secretsService", r.secretsService).
		Info("Begin Reconcile")

	var existingArchive *RelayArchive
	err := runStep("Fetch existing archive", func() error {
		var err error
		existingArchive, err = r.archiveService.fetchExisting(ctx)
		if err != nil {
			if errors.Is(err, ErrExistingArchiveNotFound) {
				r.logger.Warn("Existing archive not found. Creating new empty archive.")
				existingArchive = &RelayArchive{}
				return nil
			}
		}
		return err
	})
	if err != nil {
		return err
	}

	var newArchive *RelayArchive
	err = runStep("Load local yaml project files", func() error {
		project, err := loadProjectYamlFiles(r.projectYamlPath)
		if err != nil {
			return err
		}

		newArchive = project.toRelayArchive()
		err = newArchive.injectSecrets(r.secretsService)
		return err
	})
	if err != nil {
		return err
	}

	var reconciledArchive RelayArchive
	err = runStep("Reconcile existing archive and local yaml project files into reconciled archive", func() error {
		var err error
		reconciledArchive, err = reconcile(*existingArchive, *newArchive)
		return err
	})
	if err != nil {
		return err
	}

	err = runStep("Publish reconciled archive", func() error {
		return r.archiveService.saveNew(ctx, reconciledArchive)
	})

	return err
}

func reconcile(old, new RelayArchive) (RelayArchive, error) {
	compareResult := compareMapKeys(old.envs, new.envs)
	logger.Infof("environments: %+v", compareResult)

	// Process new envs
	for _, envKey := range compareResult.new {
		//set all versions to 1
		newEnv := new.envs[envKey]
		newEnv.metadata.EnvMetadata.Version = 1
		newEnv.metadata.incrementDataId()
		for flagKey, flag := range newEnv.data.Flags {
			flag.Version = 1
			newEnv.data.Flags[flagKey] = flag
		}
		new.envs[envKey] = newEnv
	}

	// TODO: Process deleted envs.. wtf.

	// Process existing envs
	for _, envKey := range compareResult.existing {
		shouldChangeDataId := false
		// compare env metadata:
		oldEnv := old.envs[envKey]
		newEnv := new.envs[envKey]

		// compare env metadata ignoring versions
		newEnv.metadata.EnvMetadata.Version = oldEnv.metadata.EnvMetadata.Version
		if !reflect.DeepEqual(oldEnv.metadata, newEnv.metadata) {
			newEnv.metadata.EnvMetadata.Version++
			shouldChangeDataId = true
		}

		// compare flags
		compareResult := compareMapKeys(oldEnv.data.Flags, newEnv.data.Flags)
		logger.Infof("flags: %+v", compareResult)

		// Process new flags
		for _, flagKey := range compareResult.new {
			flag := newEnv.data.Flags[flagKey]
			flag.Version = 1
			newEnv.data.Flags[flagKey] = flag
			shouldChangeDataId = true
		}

		// Process deleted flags
		for _, flagKey := range compareResult.deleted {
			deletedFlag := oldEnv.data.Flags[flagKey]
			deletedFlag.Version++
			deletedFlag.Deleted = true
			newEnv.data.Flags[flagKey] = deletedFlag
			shouldChangeDataId = true
		}

		// Process existing flags
		for _, flagKey := range compareResult.existing {
			// compare flags ignoring versions
			oldFlag := oldEnv.data.Flags[flagKey]
			newFlag := newEnv.data.Flags[flagKey]
			newFlag.Version = oldFlag.Version
			if !reflect.DeepEqual(oldFlag, newFlag) {
				newFlag.Version++
				newEnv.data.Flags[flagKey] = newFlag
				shouldChangeDataId = true
			}
		}

		if shouldChangeDataId {
			newEnv.metadata.incrementDataId()
		}
		new.envs[envKey] = newEnv
	}
	return new, nil
}

// runStep utilizes GitHub Actions' log grouping feature:
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#grouping-log-lines
func runStep(step string, f func() error) error {
	//fmt.Printf("\n[%s] BEGIN\n", step)
	fmt.Printf("::group::%s\n", step)
	err := f()
	//if err != nil {
	//	fmt.Printf("[%s] ERROR: %v\n", step, err)
	//}
	fmt.Printf("::endgroup::\n")
	//fmt.Printf("[%s] END\n", step)
	return err
}

type compareResult struct {
	deleted  []string
	new      []string
	existing []string
}

// compareMapKeys compares two maps and returns the keys that are new, existing, and deleted
func compareMapKeys[T any](old, new map[string]T) compareResult {
	deletedKeys, newKeys, existingKeys := make([]string, 0), make([]string, 0), make([]string, 0)

	// check for existing/deleted keys
	for key := range old {
		_, ok := new[key]
		if ok {
			existingKeys = append(existingKeys, key)
		} else {
			deletedKeys = append(deletedKeys, key)
		}
	}

	// check for new keys
	for key := range new {
		_, ok := old[key]
		if !ok {
			newKeys = append(newKeys, key)
		}
	}

	return compareResult{
		deleted:  deletedKeys,
		new:      newKeys,
		existing: existingKeys,
	}
}
