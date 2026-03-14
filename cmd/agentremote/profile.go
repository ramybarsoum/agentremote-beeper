package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/beeper/agentremote/cmd/internal/beeperauth"
)

const defaultProfile = "default"

type authConfig = beeperauth.Config

// configRoot returns ~/.config/agentremote
func configRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "agentremote"), nil
}

// profileRoot returns ~/.config/agentremote/profiles/<profile>
func profileRoot(profile string) (string, error) {
	root, err := configRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "profiles", profile), nil
}

// authConfigPath returns the path to the auth config for a profile.
func authConfigPath(profile string) (string, error) {
	root, err := profileRoot(profile)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "config.json"), nil
}

// instanceRoot returns the instances directory for a profile.
func instanceRoot(profile string) (string, error) {
	root, err := profileRoot(profile)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "instances"), nil
}

type instancePaths struct {
	Root             string
	ConfigPath       string
	RegistrationPath string
	LogPath          string
	PIDPath          string
	MetaPath         string
}

func getInstancePaths(profile, instanceName string) (*instancePaths, error) {
	root, err := instanceRoot(profile)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(root, instanceName)
	return &instancePaths{
		Root:             dir,
		ConfigPath:       filepath.Join(dir, "config.yaml"),
		RegistrationPath: filepath.Join(dir, "registration.yaml"),
		LogPath:          filepath.Join(dir, "bridge.log"),
		PIDPath:          filepath.Join(dir, "bridge.pid"),
		MetaPath:         filepath.Join(dir, "meta.json"),
	}, nil
}

func ensureInstanceLayout(profile, instanceName string) (*instancePaths, error) {
	sp, err := getInstancePaths(profile, instanceName)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(sp.Root, 0o700); err != nil {
		return nil, err
	}
	return sp, nil
}

func loadAuthConfig(profile string) (authConfig, error) {
	store, err := authStore(profile)
	if err != nil {
		return authConfig{}, err
	}
	return beeperauth.Load(store)
}

func saveAuthConfig(profile string, cfg authConfig) error {
	path, err := authConfigPath(profile)
	if err != nil {
		return err
	}
	return beeperauth.Save(path, cfg)
}

func getAuthOrEnv(profile string) (authConfig, error) {
	store, err := authStore(profile)
	if err != nil {
		return authConfig{}, err
	}
	return beeperauth.ResolveFromEnvOrStore(store)
}

func listProfiles() ([]string, error) {
	root, err := configRoot()
	if err != nil {
		return nil, err
	}
	profilesDir := filepath.Join(root, "profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var profiles []string
	for _, e := range entries {
		if e.IsDir() {
			profiles = append(profiles, e.Name())
		}
	}
	return profiles, nil
}

func listInstancesForProfile(profile string) ([]string, error) {
	root, err := instanceRoot(profile)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var instances []string
	for _, e := range entries {
		if e.IsDir() {
			instances = append(instances, e.Name())
		}
	}
	return instances, nil
}

func authStore(profile string) (beeperauth.Store, error) {
	path, err := authConfigPath(profile)
	if err != nil {
		return beeperauth.Store{}, err
	}
	return beeperauth.Store{
		Path: path,
		MissingError: func() error {
			return fmt.Errorf("not logged in (profile %q). Run: agentremote login --profile %s", profile, profile)
		},
	}, nil
}
