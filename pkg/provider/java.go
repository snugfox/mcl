package provider

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/snugfox/mcl/pkg/store"
)

// JavaProvider is a provider for Minecraft: Java Edition provided by Mojang.
type JavaProvider struct {
	versions   []javaVersionInfo
	versionMap map[string]*javaVersionInfo // Maps version ID to version info
}

type javaVersionInfo struct {
	ID          string    `json:"id"` // Version ID
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	Time        time.Time `json:"time"`
	ReleaseTime time.Time `json:"releaseTime"`

	versionResource *javaVersionResource // Populated manually on-demand
}

type javaVersionResource struct {
	SHA1 string `json:"sha1"` // Hex-encoded
	Size int64  `json:"size"` // In bytes
	URL  string `json:"url"`
}

const (
	// URL of the launcher manifest provided by Mojang
	launcherManifestURL string = "https://launchermeta.mojang.com/mc/game/version_manifest.json"

	// Filename of the server JAR
	serverJARFilename string = "server.jar"
)

func isAcceptedHostname(rawurl string, acceptedHostnames []string) bool {
	return true
}

func (jp *JavaProvider) fetchManifest(ctx context.Context, force bool) error {
	if force || jp.versions == nil {
		// Download and parse the JSON launcher manifest
		req, err := http.NewRequest(http.MethodGet, launcherManifestURL, nil) // TODO: Test for accepted hostnames
		if err != nil {
			return err
		}
		req = req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		var launcherManifest struct {
			Latest   map[string]string `json:"latest"`
			Versions []javaVersionInfo `json:"versions"`
		}
		if err := json.NewDecoder(res.Body).Decode(&launcherManifest); err != nil {
			return err
		}

		// Index and cache versions as we are likely to later lookup specific
		// version information.
		jp.versions = launcherManifest.Versions
		jp.versionMap = make(map[string]*javaVersionInfo)
		for i, vInfo := range launcherManifest.Versions {
			jp.versionMap[vInfo.ID] = &launcherManifest.Versions[i]
		}
		for alias, version := range launcherManifest.Latest { // Resolve alias and add to the version map
			vInfo, ok := jp.versionMap[version]
			if !ok {
				return errors.New("manifest references unknown version")
			}
			jp.versionMap[alias] = vInfo
		}
	}

	return nil
}

func (jvi *javaVersionInfo) fetchVersionManifest(ctx context.Context, force bool) (*javaVersionResource, error) {
	if force || jvi.versionResource == nil {
		// Download and parse JSON version manifest
		req, err := http.NewRequest(http.MethodGet, jvi.URL, nil) // TODO: Test for accepted hostnames
		if err != nil {
			return nil, err
		}
		req = req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		var versionManifest struct {
			Downloads struct {
				Server javaVersionResource `json:"server"`

				// ...unused fields for client resources...
			} `json:"downloads"`

			// ...other unused fields...
		}
		if err := json.NewDecoder(res.Body).Decode(&versionManifest); err != nil {
			return nil, err
		}

		jvi.versionResource = &versionManifest.Downloads.Server // We only need to track the server resource
	}

	return jvi.versionResource, nil
}

func (JavaProvider) jarPath(baseDir string) string {
	return filepath.Join(baseDir, serverJARFilename)
}

// Edition returns the edition ID and name for Minecraft: Java Edition.
func (JavaProvider) Edition() (id string, name string) {
	return "java", "Minecraft: Java Edition"
}

// Versions returns all available server versions for the edition. For
// Minecraft: Java Edition, it also returns channels, such as "release" and
// "snapshot".
func (jp *JavaProvider) Versions(ctx context.Context) ([]string, error) {
	if err := jp.fetchManifest(ctx, false); err != nil {
		return nil, err
	}

	// Determine release time cutoff for supported versions. The version manifest
	// returns server JAR as far back as 1.2.5; however, servers are available for
	// older versions through a different endpoint.
	// TODO: Support versions available through http://s3.amazonaws.com/Minecraft.Download/versions/<VERSION>/<VERSION>.json
	jvi125, ok := jp.versionMap["1.2.5"]
	if !ok {
		return nil, errors.New("version 1.2.5 not found (oldest supported server)")
	}

	versionIDs := make([]string, 0)
	for _, vInfo := range jp.versions {
		if vInfo.ReleaseTime.After(jvi125.ReleaseTime) || vInfo.ReleaseTime.Equal(jvi125.ReleaseTime) { // Filter unsupported versions prior to 1.2.5
			versionIDs = append(versionIDs, vInfo.ID)
		}
	}
	return versionIDs, nil
}

// DefaultVersion returns the default versions specified by Mojang. For
// Minecraft: Java Edition, it always returns "release".
func (JavaProvider) DefaultVersion() string {
	return "release"
}

// ResolveVersion resolves a version identifier to a fixed version
// identifier (e.g. release -> 1.7).
func (jp *JavaProvider) ResolveVersion(ctx context.Context, version string) (string, error) {
	if err := jp.fetchManifest(ctx, false); err != nil {
		return "", err
	}

	vInfo, ok := jp.versionMap[version]
	if !ok {
		return "", errors.New("version not found")
	}
	return vInfo.ID, nil
}

// IsFetchNeeded returns whether the server resources for the instance are not
// available locally and require fetching. For Minecraft: Java Edition, it
// checks if the server JAR exists locally, and if so, validates the SHA-1
// checksum with that provided by Mojang.
func (jp *JavaProvider) IsFetchNeeded(ctx context.Context, inst Instance) (bool, error) {
	ji := inst.(*JavaInstance)

	if err := jp.fetchManifest(ctx, false); err != nil {
		return false, err
	}

	// Get and extract the hash for the server from the version manifest.
	vInfo, ok := jp.versionMap[ji.ver]
	if !ok {
		return false, errors.New("version not found")
	}
	vResource, err := vInfo.fetchVersionManifest(ctx, false)
	if err != nil {
		return false, err
	}

	expectedSHA1, err := hex.DecodeString(vResource.SHA1)
	if err != nil {
		return false, nil
	}

	// Determine the hash of the JAR file available locally (if one exists)
	// TODO: Do this in parallel with fetching the version manifest
	jarPath := jp.jarPath(ji.baseDir)
	jarFile, err := os.Open(jarPath)
	if err != nil {
		return true, nil // TODO: Check existence first
	}
	defer jarFile.Close()

	h := sha1.New()
	if _, err := io.Copy(h, jarFile); err != nil {
		return false, err
	}

	// Return whether the hashes are equal
	return !bytes.Equal(h.Sum(nil), expectedSHA1), nil
}

// Fetch fetches (downloads) server resources into a specified base directory.
// For Minecraft: Java Edition, it downloads the server JAR from Mojang to the
// instance's base directory.
func (jp *JavaProvider) Fetch(ctx context.Context, inst Instance) error {
	ji := inst.(*JavaInstance)

	if err := jp.fetchManifest(ctx, false); err != nil {
		return err
	}

	// Download and parse the version manifest
	vInfo, ok := jp.versionMap[ji.ver]
	if !ok {
		return errors.New("version not found")
	}
	vResource, err := vInfo.fetchVersionManifest(ctx, false)
	if err != nil {
		return err
	}

	// Create/open server JAR file
	if err := os.MkdirAll(ji.baseDir, os.ModeDir|0755); err != nil {
		return err
	}
	jarPath := ji.jp.jarPath(ji.baseDir)
	jarFile, err := os.OpenFile(jarPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	// Download the server JAR from the URL specified by the verison manifest.
	res, err := http.Get(vResource.URL) // TODO: Utilize context
	if err != nil {
		return err
	}
	defer res.Body.Close()
	io.Copy(jarFile, res.Body) // TODO: Check for error

	return nil
}

// IsPrepareNeeded returns whether the server resources for the instance are not
// available for immediate use and required additional preparation. For
// Minecraft: Java Edition, it always returns false and a nil error.
func (jp *JavaProvider) IsPrepareNeeded(_ context.Context, inst Instance) (bool, error) {
	return false, nil // There is no preparation step for the java edition
}

// Prepare prepares (preprocesses) fetched server resources such that they are
// immediately useable without any further modifications. For Minecraft: Java
// Edition, it is effectively a no-op.
func (jp *JavaProvider) Prepare(_ context.Context, inst Instance) error {
	return nil // There is no preparation step for the Java edition
}

// Run runs a server within a specified working directory. Server resources
// should have been previously fetched to the instance's base directory and
// version prior to calling Run. Runtime arguments are passed as JVM
// options and server arguments are passed to the server JAR. Either argument
// parameter may be nil if no arguments need to be specified. If the instance is
// already running, it will return an error.
//
// TODO: Allow interaction over STDIN
func (jp *JavaProvider) Run(ctx context.Context, inst Instance, workingDir string, runtimeArgs, serverArgs []string) error {
	ji := inst.(*JavaInstance)

	jarPath, err := filepath.Abs(ji.jp.jarPath(ji.baseDir))
	if err != nil {
		return err
	}

	// Concatenate arguments
	args := append(runtimeArgs, "-jar", jarPath)
	if serverArgs != nil {
		args = append(args, serverArgs...)
	}

	// Create a new command with all standard pipes attached.
	ji.mu.Lock()
	if ji.cmd != nil {
		ji.mu.Unlock()
		return errors.New("instance already running")
	}
	ji.cmd = exec.CommandContext(ctx, "java", args...) // TODO: Shutdown gracefully instead of kill when context cancelled
	ji.cmd.Dir = workingDir
	ji.stdinPipe, _ = ji.cmd.StdinPipe()
	ji.cmd.Stdout = os.Stdout
	ji.cmd.Stderr = os.Stderr
	ji.cmdExit = make(chan struct{})
	ji.mu.Unlock()

	err = ji.cmd.Run()
	close(ji.cmdExit)
	return err
}

// Stop stops the server instance. If the instance is not running, it will
// return an error.
//
// TODO: Be more specific in function doc
// TODO: Handle ctx
func (jp *JavaProvider) Stop(_ context.Context, inst Instance) error {
	ji := inst.(*JavaInstance)
	ji.mu.Lock()
	defer ji.mu.Unlock()

	if ji.cmd == nil {
		return errors.New("instance not running")
	}

	fmt.Fprintln(ji.stdinPipe, "\rstop")
	<-ji.cmdExit
	return nil
}

// NewInstance creates a new JavaInstance for a given version stored in baseDir.
func (jp *JavaProvider) NewInstance(ver, baseTmpl string) (Instance, error) {
	ed, _ := jp.Edition()
	baseDir, err := store.BaseDir(".", baseTmpl, ed, ver)
	if err != nil {
		return nil, err
	}
	return &JavaInstance{
		jp:      jp,
		ver:     ver,
		baseDir: baseDir,
	}, nil
}

// JavaInstance is a server instance for Minecraft: Java Edition.
type JavaInstance struct {
	mu sync.Mutex

	// Immutable
	jp      *JavaProvider
	ver     string
	baseDir string

	cmd       *exec.Cmd
	cmdExit   chan struct{}
	stdinPipe io.WriteCloser
}

// Provider returns the JavaProvider used to create the instance.
func (ji *JavaInstance) Provider() Provider {
	return ji.jp
}

// Version returns the instance's version.
func (ji *JavaInstance) Version() string {
	return ji.ver
}

// BaseDir returns the instance's base directory for server resources.
func (ji *JavaInstance) BaseDir() string {
	return ji.baseDir
}
