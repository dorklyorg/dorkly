package dorkly

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/launchdarkly/go-server-sdk-evaluation/v3/ldmodel"
	"hash"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	fileCreateMode = 0755
)

// RelayArchive represents the data consumed by ld-relay in offline mode
type RelayArchive struct {
	envs map[string]Env
}

func (ra *RelayArchive) String() string {
	envStrings := make([]string, 0, len(ra.envs))
	for _, env := range ra.envs {
		envStrings = append(envStrings, env.String())
	}

	return fmt.Sprintf("Environments: %v", strings.Join(envStrings, ", "))
}

// Env is a representation of the <env>.json and <env>-data.json files in the relay archive
type Env struct {
	metadata RelayArchiveEnv
	data     RelayArchiveData
}

func (e *Env) String() string {
	return fmt.Sprintf("(Name: %v, Version: %d, DataID: %v, Data: %v)",
		e.metadata.EnvMetadata.EnvName, e.metadata.EnvMetadata.Version, e.metadata.DataId, e.data.String())
}

func (e *Env) computeDataId() {
	dataId := e.metadata.EnvMetadata.Version
	for _, flag := range e.data.Flags {
		dataId += flag.Version
	}
	e.metadata.DataId = fmt.Sprintf("%d", dataId)
}

// RelayArchiveEnv is a representation of the <env>.json file in the relay archive
type RelayArchiveEnv struct {
	EnvMetadata RelayArchiveEnvMetadata `json:"env"`

	// this must be changed in order for flag changes to be picked up.
	// the official relay archive uses a double-quoted string like this: "\"<value>\""
	DataId string `json:"dataId"`
}

func (a *RelayArchiveEnv) String() string {
	return fmt.Sprintf("Env: %v, DataId: %v", a.EnvMetadata.String(), a.DataId)
}

type RelayArchiveEnvMetadata struct {
	EnvID      string    `json:"envID"`
	EnvKey     string    `json:"envKey"`
	EnvName    string    `json:"envName"`
	MobKey     string    `json:"mobKey"`
	ProjKey    string    `json:"projKey"`
	ProjName   string    `json:"projName"`
	SDKKey     SDKKeyRep `json:"sdkKey"`
	DefaultTTL int       `json:"defaultTtl"`
	SecureMode bool      `json:"secureMode"`
	Version    int       `json:"version"`
}

func (r *RelayArchiveEnvMetadata) String() string {
	return fmt.Sprintf("EnvName: %v, ProjName: %s, Version: %v", r.EnvName, r.ProjName, r.Version)
}

type SDKKeyRep struct {
	Value string `json:"value"`
}

// RelayArchiveData is a representation of the <env>-data.json file in the relay archive
type RelayArchiveData struct {
	Segments map[string]ldmodel.Segment     `json:"segments"`
	Flags    map[string]ldmodel.FeatureFlag `json:"flags"`
}

func (rad *RelayArchiveData) String() string {
	return fmt.Sprintf("Flag count: %d", len(rad.Flags))
}

func (ra *RelayArchive) injectSecrets(secretsService SecretsService) error {
	ctx := context.Background()
	for _, env := range ra.envs {
		sdkKey, err := secretsService.getSdkKey(ctx, env.metadata.EnvMetadata.ProjKey, env.metadata.EnvMetadata.EnvName)
		if err != nil {
			return err
		}
		env.metadata.EnvMetadata.SDKKey.Value = sdkKey

		mobKey, err := secretsService.getMobileKey(ctx, env.metadata.EnvMetadata.ProjKey, env.metadata.EnvMetadata.EnvName)
		if err != nil {
			return err
		}
		env.metadata.EnvMetadata.MobKey = mobKey

		ra.envs[env.metadata.EnvMetadata.EnvName] = env
	}
	return nil
}

// marshalArchiveFilesJson returns a map of relay archive filenames to their json bytes
func (ra *RelayArchive) marshalArchiveFilesJson() (map[string][]byte, error) {
	archiveContents := make(map[string][]byte)

	for envName, env := range ra.envs {
		// create env metadata file: <envName>.json
		jsonBytes, err := json.MarshalIndent(env.metadata, "", "  ")
		if err != nil {
			return nil, err
		}
		archiveContents[envName+".json"] = jsonBytes

		// create env data file: <envName>-data.json
		jsonBytes, err = json.MarshalIndent(env.data, "", "  ")
		if err != nil {
			return nil, err
		}
		archiveContents[envName+"-data.json"] = jsonBytes
	}

	return archiveContents, nil
}

func (ra *RelayArchive) toTarGzFile(pathToArchive string) error {
	archiveFilesPath := filepath.Join(os.TempDir(), fmt.Sprintf("dorkly-%v", time.Now().UnixNano()))
	err := ra.createArchiveFilesAndComputeChecksum(archiveFilesPath)
	if err != nil {
		return err
	}

	err = directoryToTarGz(archiveFilesPath, pathToArchive)
	return err
}

func (ra *RelayArchive) createArchiveFilesAndComputeChecksum(path string) error {
	err := ensureEmptyDirExists(path)
	if err != nil {
		return err
	}
	logger.Infoln("Creating archive files in path: ", path)

	archiveContents, err := ra.marshalArchiveFilesJson()
	if err != nil {
		return err
	}

	// collect all filepaths to be checksummed
	filepaths := make([]string, 0, len(ra.envs)*2)

	for filename, jsonBytes := range archiveContents {
		envMetadataPath := filepath.Join(path, filename)
		filepaths = append(filepaths, envMetadataPath)
		err = os.WriteFile(envMetadataPath, jsonBytes, fileCreateMode)
		if err != nil {
			return err
		}
	}

	// create checksum file
	checksumPath := filepath.Join(path, "checksum.md5")
	checkSumBytes, err := md5ChecksumForFiles(filepaths)
	if err != nil {
		return err
	}
	err = os.WriteFile(checksumPath, checkSumBytes, fileCreateMode)
	if err != nil {
		return err
	}

	return nil
}

// directoryToTarGz creates a tar.gz archive of the directory at the given path
// We could use the archive/tar package to create the archive, but it's easier to use the tar command
func directoryToTarGz(dir, pathToArchive string) error {
	logger.Infof("Creating tar.gz archive of directory: %s to %s", dir, pathToArchive)
	cmd := exec.Command("tar", "-czvf", pathToArchive, "-C", dir, ".")

	output, err := cmd.CombinedOutput()
	logger.Infoln("tar command output: \n" + string(output))
	return err
}

func ensureEmptyDirExists(path string) error {
	// Check if directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create directory if it does not exist
		err = os.MkdirAll(path, fileCreateMode)
		if err != nil {
			return err
		}
	} else {
		// Check if directory is empty
		files, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		if len(files) > 0 {
			return errors.Errorf("directory %s is not empty. Refusing to write to non-empty directory", path)
		}
	}
	return nil
}

func md5ChecksumForFiles(files []string) ([]byte, error) {
	sort.Strings(files)
	h := md5.New()
	for _, f := range files {
		if err := addFileToHash(h, f); err != nil {
			return nil, err // COVERAGE: can't cause this condition in unit tests
		}
	}
	return h.Sum(nil), nil
}

func addFileToHash(h hash.Hash, file string) error {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(h, f)
	return err
}

func loadRelayArchiveFromTarGzFile(path string) (*RelayArchive, error) {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	logger.Infoln("Loading RelayArchive from file:", fullPath)

	ad := RelayArchive{envs: make(map[string]Env)}

	archiveContents, err := readTarGz(fullPath)
	if err != nil {
		return nil, err
	}

	// iterate over *-data.json files in archive to get each env
	for name, fileBytes := range archiveContents {
		if name == "." {
			continue
		}
		logger.With("file", name).Debugln("Processing file...")
		if strings.HasSuffix(name, "-data.json") {
			// load flag data
			envName := strings.TrimSuffix(name, "-data.json")
			logger.With("file", name).With("env", envName).Infof("Found data file for env")
			data := RelayArchiveData{}
			err := json.Unmarshal(fileBytes, &data)
			if err != nil {
				return nil, err
			}

			// load env metadata
			envMetadata := RelayArchiveEnv{}
			envMetadataFileName := envName + ".json"
			if envBytes, ok := archiveContents[envMetadataFileName]; ok {
				logger.With("file", envMetadataFileName).With("env", envName).Infof("Found metadata file for env")
				err = json.Unmarshal(envBytes, &envMetadata)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, errors.Errorf("missing metadata file for env: %s", envName)
			}

			ad.envs[envName] = Env{
				metadata: envMetadata,
				data:     data,
			}
		}
	}

	// basic validation
	//logger.Infoln("Performing sanity checks on loaded files...")
	//if len(ad.envs) == 0 {
	//	return nil, errors.New((fmt.Sprintf("no envs found in dir: %s", path)
	//}
	//for envName, env := range ad.envs {
	//	if env.metadata.DataId == "" {
	//		return nil, errors.New((fmt.Sprintf("env [%s] has no dataId field! Check for well-formed %s.json file", envName, envName)
	//	}
	//}

	logger.Infof("Existing archive: %v", ad.String())
	return &ad, nil

}

func readTarGz(srcFile string) (map[string][]byte, error) {
	l := logger.With("archiveFile", srcFile)
	file, err := os.Open(srcFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	contents := make(map[string][]byte)
	// Iterate through the files in the archive adding them to the map
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive or root directory
		}
		if err != nil {
			return nil, err
		}

		filename := filepath.Base(header.Name)
		if filename == "." {
			continue
		}
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, tr); err != nil {
			return nil, err
		}
		// We expect all files to be in the root directory of the archive so we strip out the leading ./ from the filename
		contents[filename] = buf.Bytes()
	}

	filenames := make([]string, 0, len(contents))
	for filename := range contents {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	l.With("contents", filenames).Infof("Found %d files in archive", len(filenames))

	return contents, nil
}
