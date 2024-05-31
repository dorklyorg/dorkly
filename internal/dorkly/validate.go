package dorkly

// ValidateYamlProject loads the project yaml files from the given path and validates them.
// It returns a Project struct if the yaml files are valid, otherwise it returns an error.
func ValidateYamlProject(path string) (*Project, error) {
	return loadProjectYamlFiles(path)
}
