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
	"path"
	"path/filepath"
	"time"
)

type VanillaProvider struct {
	versions   []vanillaVersionInfo
	versionMap map[string]*vanillaVersionInfo
}

type vanillaVersionInfo struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	Time        time.Time `json:"time"`
	ReleaseTime time.Time `json:"releaseTime"`

	versionResource *vanillaVersionResource
}

type vanillaVersionResource struct {
	SHA1 string `json:"sha1"`
	Size int64  `json:"size"`
	URL  string `json:"url"`
}

const launcherManifestURL string = "https://launchermeta.mojang.com/mc/game/version_manifest.json"

func isAcceptedHostname(rawurl string, acceptedHostnames []string) bool {
	return true
}

func (vp *VanillaProvider) fetchManifest(ctx context.Context, force bool) error {
	if force || vp.versions == nil {
		req, err := http.NewRequest(http.MethodGet, launcherManifestURL, nil) // TODO: Test for accepted hostnames
		if err != nil {
			return err
		}
		req = req.WithContext(ctx)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		var vm struct {
			Latest   map[string]string    `json:"latest"`
			Versions []vanillaVersionInfo `json:"versions"`
		}
		if err := json.NewDecoder(res.Body).Decode(&vm); err != nil {
			return err
		}

		vp.versions = vm.Versions
		vp.versionMap = make(map[string]*vanillaVersionInfo)
		for i, vInfo := range vm.Versions {
			vp.versionMap[vInfo.ID] = &vm.Versions[i]
		}
		for alias, version := range vm.Latest {
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

				// ...unused fields for client resources
			} `json:"downloads"`

			// ...other unused fields
		}
		if err := json.NewDecoder(res.Body).Decode(&versionManifest); err != nil {
			return nil, err
		}
		vvi.versionResource = &versionManifest.Downloads.Server
	}

	return vvi.versionResource, nil
}

func (VanillaProvider) jarPath(outputDir string) string {
	return path.Join(outputDir, "server.jar")
}

func (VanillaProvider) Edition() (id string, name string) {
	return "vanilla", "Minecraft: Java Edition"
}

func (vp *VanillaProvider) Versions(ctx context.Context) ([]string, error) {
	if err := vp.fetchManifest(ctx, false); err != nil {
		return nil, err
	}

	versionIDs := make([]string, len(vp.versions))
	for i, vInfo := range vp.versions {
		versionIDs[i] = vInfo.ID
	}
	return versionIDs, nil
}

func (VanillaProvider) DefaultVersion() string {
	return "release"
}

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

func (vp *VanillaProvider) IsFetchNeeded(ctx context.Context, baseDir, version string) (bool, error) {
	if err := vp.fetchManifest(ctx, false); err != nil {
		return false, err
	}

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

	return !bytes.Equal(h.Sum(nil), expectedSHA1), nil
}

func (vp *VanillaProvider) Fetch(ctx context.Context, baseDir, version string) error {
	if err := vp.fetchManifest(ctx, false); err != nil {
		return err
	}

	vInfo, ok := vp.versionMap[version]
	if !ok {
		return errors.New("version not found")
	}
	vResource, err := vInfo.fetchVersionManifest(ctx, false)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(baseDir, os.ModeDir|0755); err != nil {
		return err
	}
	file, err := os.OpenFile(path.Join(baseDir, "server.jar"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	res, err := http.Get(vResource.URL) // TODO: Utilize context
	if err != nil {
		return err
	}
	defer res.Body.Close()

	io.Copy(file, res.Body)

	return nil
}

func (vp *VanillaProvider) IsPrepareNeeded(_ context.Context, _, _ string) (bool, error) {
	return false, nil // There is no preparation step for the vanilla edition
}

func (vp *VanillaProvider) Prepare(_ context.Context, _, _ string) error {
	return nil // There is no preparation step for the vanilla edition
}

func (vp *VanillaProvider) Run(ctx context.Context, baseDir, workingDir, version string, runtimeArgs, serverArgs []string) error {
	jarPath, err := filepath.Abs(filepath.Join(baseDir, "server.jar"))
	if err != nil {
		return err
	}

	args := append(runtimeArgs, "-jar", jarPath)
	if serverArgs != nil {
		args = append(args, serverArgs...)
	}
	cmd := exec.CommandContext(ctx, "java", args...)
	cmd.Dir = workingDir

	// Vanilla server may use all standard pipes
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
