package dorkly

import (
	"github.com/go-errors/errors"
	"github.com/launchdarkly/go-server-sdk-evaluation/v3/ldmodel"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Project is the dorkly representation of a LaunchDarkly project
type Project struct {
	// Name is reused as key
	Name string `yaml:"name"`

	Description string          `yaml:"description"`
	Flags       map[string]Flag `yaml:"flags"`

	path         string
	environments []string
}

// Flag contains everything needed to serve a flag for all environments in a project
type Flag struct {
	FlagBase

	// envConfigs is a map of environment name to environment-specific flag configuration
	envConfigs map[string]FlagConfigForEnv
}

func loadProjectYamlFiles(path string) (*Project, error) {
	if !isDirectory(path) {
		return nil, errors.Errorf("path [%s] is not a directory", path)
	}

	projectYmlPath := path + "/project.yml"
	logger.Infof("Loading project config from file [%s]", projectYmlPath)
	f, err := os.Open(projectYmlPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var project Project
	dec := yaml.NewDecoder(f)
	err = dec.Decode(&project)
	if err != nil {
		return nil, err
	}
	project.path = path

	err = project.loadEnvironmentNames()
	if err != nil {
		return nil, err
	}

	project.Flags = make(map[string]Flag)
	err = project.loadFlagsYamlFiles()
	if err != nil {
		return nil, err
	}

	return &project, err
}

func (p *Project) loadFlagsYamlFiles() error {
	flagsPath := filepath.Join(p.path, "flags")
	logger.Infof("Loading flags from path [%s]", flagsPath)
	files, err := os.ReadDir(flagsPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			logger.Infof("Skipping unexpected directory [%s]", file.Name())
			continue
		}

		if strings.HasSuffix(file.Name(), ".yml") {
			filePath := filepath.Join(flagsPath, file.Name())
			logger.Infof("Loading flag from file [%s]", filePath)
			flag, err := p.loadFlagYamlFile(filePath)
			if err != nil {
				return err
			}

			p.Flags[flag.key] = *flag
		}
	}
	return nil
}

func (p *Project) loadFlagYamlFile(filePath string) (*Flag, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var flagBase FlagBase
	dec := yaml.NewDecoder(f)
	err = dec.Decode(&flagBase)
	if err != nil {
		return nil, err
	}
	flag := Flag{FlagBase: flagBase}
	flag.key = getFileNameNoSuffix(f.Name())
	flag.envConfigs = make(map[string]FlagConfigForEnv)
	for _, env := range p.environments {
		flagEnvConfig, err := p.loadFlagConfigForEnvYamlFile(flag, filepath.Join(p.path, "environments", env, flag.key+".yml"))
		if err != nil {
			return nil, err
		}
		err = flagEnvConfig.Validate(flag.FlagBase)
		if err != nil {
			return nil, err
		}
		flag.envConfigs[env] = flagEnvConfig

	}

	return &flag, nil
}

func (p *Project) loadEnvironmentNames() error {
	envPath := filepath.Join(p.path, "environments")
	logger.Infof("Loading environments from path [%s]", envPath)
	files, err := os.ReadDir(envPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			logger.Infof("Adding environment [%s]", file.Name())
			p.environments = append(p.environments, file.Name())
		}
	}
	sort.Strings(p.environments)
	return nil
}

func (p *Project) loadFlagConfigForEnvYamlFile(flag Flag, filePath string) (FlagConfigForEnv, error) {
	logger.Infof("Loading environment-specific flag config from file [%s]", filePath)
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)

	switch flag.Type {
	case FlagTypeBoolean:
		envData := &FlagBoolean{}
		err = dec.Decode(&envData)
		if err != nil {
			return nil, err
		}
		return envData, nil
	case FlagTypeBooleanRollout:
		envData := &FlagBooleanRollout{}
		err = dec.Decode(&envData)
		if err != nil {
			return nil, err
		}
		return envData, nil
	}
	return nil, errors.Errorf("unsupported flag type [%s] for flag [%s]", p.Flags[flag.key].Type, flag.key)
}

// toRelayArchive converts a dorkly Project to a RelayArchive for consumption by ld-relay
func (p *Project) toRelayArchive() *RelayArchive {
	envs := make(map[string]Env)
	for _, env := range p.environments {
		envs[env] = Env{
			metadata: RelayArchiveEnv{
				EnvMetadata: RelayArchiveEnvMetadata{
					EnvID:    env,
					EnvKey:   env,
					EnvName:  env,
					MobKey:   insecureMobileKey(env),
					ProjKey:  p.Name,
					ProjName: p.Name,
					SDKKey: SDKKeyRep{
						Value: insecureSdkKey(env),
					},
					DefaultTTL: 0,
					SecureMode: false,
					Version:    0,
				},
			},
			data: RelayArchiveData{
				Segments: make(map[string]ldmodel.Segment),
				Flags:    make(map[string]ldmodel.FeatureFlag),
			},
		}
		for _, flag := range p.Flags {
			envs[env].data.Flags[flag.key] = flag.envConfigs[env].ToLdFlag(flag.FlagBase)
		}
	}
	return &RelayArchive{envs: envs}
}

func insecureSdkKey(env string) string {
	return "sdk-" + env + "-not-secure"
}

func insecureMobileKey(env string) string {
	return "mob-" + env + "-not-secure"
}

func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func getFileNameNoSuffix(path string) string {
	fileName := filepath.Base(path)
	noSuffix := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return noSuffix
}
