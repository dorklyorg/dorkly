package dorkly

import (
	"context"
	"log"
	"reflect"
)

type Reconciler struct {
	archiveService  RelayArchiveService
	projectYamlPath string
}

func NewReconciler(archiveService RelayArchiveService, projectYamlPath string) *Reconciler {
	return &Reconciler{
		archiveService:  archiveService,
		projectYamlPath: projectYamlPath,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context) error {
	existingArchive, err := r.archiveService.fetchExisting(ctx)
	if err != nil {
		return err
	}

	project, err := loadProjectYamlFiles(r.projectYamlPath)
	if err != nil {
		return err
	}

	newArchive := project.toRelayArchive()
	reconciledArchive, err := reconcile(*existingArchive, *newArchive)
	if err != nil {
		return err
	}

	return r.archiveService.saveNew(ctx, reconciledArchive)
}

func reconcile(old, new RelayArchive) (RelayArchive, error) {
	compareResult := compareMapKeys(old.envs, new.envs)
	log.Printf("environments: %+v", compareResult)

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
		log.Printf("flags: %+v", compareResult)

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

type compareResult struct {
	deleted  []string
	new      []string
	existing []string
}

// compareMapKeys compares two maps and returns the keys that are new, existing, and deleted
func compareMapKeys[T any](old, new map[string]T) compareResult {
	deletedKeys, newKeys, existingKeys := make([]string, 0), make([]string, 0), make([]string, 0)

	// check for existing/deleted keys
	for key, _ := range old {
		_, ok := new[key]
		if ok {
			existingKeys = append(existingKeys, key)
		} else {
			deletedKeys = append(deletedKeys, key)
		}
	}

	// check for new keys
	for key, _ := range new {
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
