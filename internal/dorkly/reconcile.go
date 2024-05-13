package dorkly

import (
	"github.com/google/go-cmp/cmp"
	"log"
)

func Reconcile(old, new RelayArchive) (RelayArchive, error) {
	compareResult := compareMaps(old.envs, new.envs)
	log.Printf("environments: %+v", compareResult)

	// Process new envs
	for _, envKey := range compareResult.new {
		//set all versions to 1
		newEnv := new.envs[envKey]
		newEnv.EnvMetadata.Env.Version = 1
		newEnv.EnvMetadata.IncrementDataId()
		for flagKey, flag := range newEnv.Flags.Flags {
			flag.Version = 1
			newEnv.Flags.Flags[flagKey] = flag
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
		newEnv.EnvMetadata.Env.Version = oldEnv.EnvMetadata.Env.Version
		if !cmp.Equal(oldEnv.EnvMetadata, newEnv.EnvMetadata) {
			newEnv.EnvMetadata.Env.Version++
			shouldChangeDataId = true
		}

		// compare flags
		compareResult := compareMaps(oldEnv.Flags.Flags, newEnv.Flags.Flags)
		log.Printf("falgs: %+v", compareResult)

		// Process new flags
		for _, flagKey := range compareResult.new {
			flag := newEnv.Flags.Flags[flagKey]
			flag.Version = 1
			newEnv.Flags.Flags[flagKey] = flag
			shouldChangeDataId = true
		}

		// Process deleted flags
		for _, flagKey := range compareResult.deleted {
			deletedFlag := oldEnv.Flags.Flags[flagKey]
			deletedFlag.Version++
			deletedFlag.Deleted = true
			shouldChangeDataId = true
		}

		// Process existing flags
		for _, flagKey := range compareResult.existing {
			// compare flags ignoring versions
			oldFlag := oldEnv.Flags.Flags[flagKey]
			newFlag := newEnv.Flags.Flags[flagKey]
			newFlag.Version = oldFlag.Version
			if !cmp.Equal(oldFlag, newFlag) {
				newFlag.Version++
				newEnv.Flags.Flags[flagKey] = newFlag
				shouldChangeDataId = true
			}
		}

		if shouldChangeDataId {
			newEnv.EnvMetadata.IncrementDataId()
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

func compareMaps[T any](old, new map[string]T) compareResult {
	deletedKeys, newKeys, existingKeys := make([]string, 0), make([]string, 0), make([]string, 0)

	// check for existing/deleted envs
	for envKey, _ := range old {
		_, ok := new[envKey]
		if ok {
			existingKeys = append(existingKeys, envKey)
		} else {
			deletedKeys = append(deletedKeys, envKey)
		}
	}

	// check for new envs
	for envKey, _ := range new {
		_, ok := old[envKey]
		if !ok {
			newKeys = append(newKeys, envKey)
		}
	}

	return compareResult{
		deleted:  deletedKeys,
		new:      newKeys,
		existing: existingKeys,
	}
}
