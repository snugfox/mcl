package provider

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// VanillaProvider is a provider for Minecraft: Java Edition provided by Mojang.
type VanillaProvider struct {
	versions   []vanillaVersionInfo
	versionMap map[string]*vanillaVersionInfo // Maps version ID to version info
}

type vanillaVersionInfo struct {
	ID          string    `json:"id"` // Version ID
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	Time        time.Time `json:"time"`
	ReleaseTime time.Time `json:"releaseTime"`

	versionResource *vanillaVersionResource // Populated manually on-demand
}

type vanillaVersionResource struct {
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

func (vp *VanillaProvider) fetchManifest(ctx context.Context, force bool) error {
	if force || vp.versions == nil {
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
			Latest   map[string]string    `json:"latest"`
			Versions []vanillaVersionInfo `json:"versions"`
		}
		if err := json.NewDecoder(res.Body).Decode(&launcherManifest); err != nil {
			return err
		}

		// Index and cache versions as we are likely to later lookup specific
		// version information.
		vp.versions = launcherManifest.Versions
		vp.versionMap = make(map[string]*vanillaVersionInfo)
		for i, vInfo := range launcherManifest.Versions {
			vp.versionMap[vInfo.ID] = &launcherManifest.Versions[i]
		}
		for alias, version := range launcherManifest.Latest { // Resolve alias and add to the version map
			vInfo, ok := vp.versionMap[version]
			if !ok {
				return errors.New("manifest references unknown version")
			}
			vp.versionMap[alias] = vInfo
		}
	}

	return nil
}

func (vvi *vanillaVersionInfo) fetchVersionManifest(ctx context.Context, force bool) (*vanillaVersionResource, error) {
	if force || vvi.versionResource == nil {
		// Download and parse JSON version manifest
		req, err := http.NewRequest(http.MethodGet, vvi.URL, nil) // TODO: Test for accepted hostnames
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
				Server vanillaVersionResource `json:"server"`

				// ...unused fields for client resources...
			} `json:"downloads"`

			// ...other unused fields...
		}
		if err := json.NewDecoder(res.Body).Decode(&versionManifest); err != nil {
			return nil, err
		}

		vvi.versionResource = &versionManifest.Downloads.Server // We only need to track the server resource
	}

	return vvi.versionResource, nil
}

func (VanillaProvider) jarPath(baseDir string) string {
	return filepath.Join(baseDir, serverJARFilename) // TODO: Move to constant
}

// Edition returns the edition ID and name for Minecraft: Java Edition.
func (VanillaProvider) Edition() (id string, name string) {
	return "vanilla", "Minecraft: Java Edition"
}

// Versions returns all available server versions for the edition. For
// Minecraft: Java Edition, it also returns channels, such as "release" and
// "snapshot".
func (vp *VanillaProvider) Versions(ctx context.Context) ([]string, error) {
	if err := vp.fetchManifest(ctx, false); err != nil {
		return nil, err
	}

	// TODO: Filter versions without server available (can use time constant)
	versionIDs := make([]string, len(vp.versions))
	for i, vInfo := range vp.versions {
		versionIDs[i] = vInfo.ID
	}
	return versionIDs, nil
}

// DefaultVersion returns the default versions specified by Mojang. For
// Minecraft: Java Edition, it always returns "release".
func (VanillaProvider) DefaultVersion() string {
	return "release"
}

// ResolveVersion resolves a version identifier to a fixed version
// identifier (e.g. release -> 1.7).
func (vp *VanillaProvider) ResolveVersion(ctx context.Context, version string) (string, error) {
	if err := vp.fetchManifest(ctx, false); err != nil {
		return "", err
	}

	vInfo, ok := vp.versionMap[version]
	if !ok {
		return "", errors.New("version not found")
	}
	return vInfo.ID, nil
}

// IsFetchNeeded returns whether the server resources for the edition and a
// specified version are not available locally and require fetching. For
// Minecraft: Java Edition, it checks if the server JAR exists locally, and if
// so, compares the SHA-1 checksum with that provided by Mojang.
func (vp *VanillaProvider) IsFetchNeeded(ctx context.Context, baseDir, version string) (bool, error) {
	if err := vp.fetchManifest(ctx, false); err != nil {
		return false, err
	}

	// Get and extract the hash for the server from the version manifest.
	vInfo, ok := vp.versionMap[version]
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
	jarPath := vp.jarPath(baseDir)
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
// base directory.
func (vp *VanillaProvider) Fetch(ctx context.Context, baseDir, version string) error {
	if err := vp.fetchManifest(ctx, false); err != nil {
		return err
	}

	// Download and parse the version manifest
	vInfo, ok := vp.versionMap[version]
	if !ok {
		return errors.New("version not found")
	}
	vResource, err := vInfo.fetchVersionManifest(ctx, false)
	if err != nil {
		return err
	}

	// Create/open server JAR file
	if err := os.MkdirAll(baseDir, os.ModeDir|0755); err != nil {
		return err
	}
	jarPath := vp.jarPath(baseDir)
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

// IsPrepareNeeded returns whether the server resources for the edition and a
// specified version are not available for immediate use and required
// additional preparation. For Minecraft: Java Edition, it always returns false
// and a nil error.
func (vp *VanillaProvider) IsPrepareNeeded(_ context.Context, _, _ string) (bool, error) {
	return false, nil // There is no preparation step for the vanilla edition
}

// Prepare prepares (preprocesses) fetched server resources such that they are
// immediately useable without any further modifications. For Minecraft: Java
// Edition, it is effectively a no-op.
func (vp *VanillaProvider) Prepare(_ context.Context, _, _ string) error {
	return nil // There is no preparation step for the vanilla edition
}

// Run runs a server within a specified working directory. Server resources
// should have been previously fetched to the same base directory and for the
// same version prior to calling Run. Runtime arguments are passed as JVM
// options and server arguments are passed to the server JAR. Either argument
// parameter may be nil if no arguments need to be specified.
func (vp *VanillaProvider) Run(ctx context.Context, baseDir, workingDir, version string, runtimeArgs, serverArgs []string) error {
	jarPath, err := filepath.Abs(vp.jarPath(baseDir))
	if err != nil {
		return err
	}

	// Concatenate arguments
	args := append(runtimeArgs, "-jar", jarPath)
	if serverArgs != nil {
		args = append(args, serverArgs...)
	}
	cmd := exec.CommandContext(ctx, "java", args...) // TODO: Shutdown gracefully instead of kill when context cancelled
	cmd.Dir = workingDir

	// Vanilla server may use all standard pipes
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
