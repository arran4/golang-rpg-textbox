package skill

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FetchSkill fetches a skill from source (local or github) and extracts it to targetDir.
// It returns the revision (commit hash or local digest).
func FetchSkill(source string, targetDir string) (string, error) {
	if strings.HasPrefix(source, ".") || strings.HasPrefix(source, "/") {
		return fetchLocal(source, targetDir)
	}

	// Try parsing as GitHub: owner/repo[/path]
	parts := strings.Split(source, "/")
	if len(parts) >= 2 {
		owner := parts[0]
		repo := parts[1]
		subpath := ""
		if len(parts) > 2 {
			subpath = strings.Join(parts[2:], "/")
		}
		return fetchGitHub(owner, repo, subpath, targetDir)
	}

	return "", fmt.Errorf("unsupported source format: %s", source)
}

func fetchLocal(source string, targetDir string) (string, error) {
	stat, err := os.Stat(source)
	if err != nil {
		return "", err
	}
	if !stat.IsDir() {
		return "", fmt.Errorf("local source %s is not a directory", source)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-regular files to avoid copying symlinks, sockets, devices, etc directly.
		if !info.Mode().IsRegular() && !info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(targetDir, rel)
		if info.IsDir() {
			return os.MkdirAll(dest, info.Mode())
		}
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = srcFile.Close() }()

		destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer func() { _ = destFile.Close() }()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("local-%d", stat.ModTime().Unix()), nil
}

func fetchGitHub(owner string, repo string, subpath string, targetDir string) (string, error) {
	// 1. Get latest commit hash
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/HEAD", owner, repo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get commit info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch from GitHub: %s", resp.Status)
	}

	var commitData struct {
		Sha string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commitData); err != nil {
		return "", err
	}
	revision := commitData.Sha

	// 2. Download zip
	zipURL := fmt.Sprintf("https://github.com/%s/%s/archive/%s.zip", owner, repo, revision)
	reqZip, err := http.NewRequest("GET", zipURL, nil)
	if err != nil {
		return "", err
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		reqZip.Header.Set("Authorization", "token "+token)
	}
	respZip, err := client.Do(reqZip)
	if err != nil {
		return "", fmt.Errorf("failed to download zip: %w", err)
	}
	defer func() { _ = respZip.Body.Close() }()

	if respZip.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download zip: %s", respZip.Status)
	}

	body, err := io.ReadAll(respZip.Body)
	if err != nil {
		return "", err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}

	rootPrefix := fmt.Sprintf("%s-%s/", repo, revision)
	extractPrefix := rootPrefix
	if subpath != "" {
		extractPrefix = rootPrefix + subpath
		if !strings.HasSuffix(extractPrefix, "/") {
			extractPrefix += "/"
		}
	}

	found := false
	for _, file := range zipReader.File {
		if !strings.HasPrefix(file.Name, extractPrefix) {
			continue
		}

		relPath := strings.TrimPrefix(file.Name, extractPrefix)
		if relPath == "" {
			continue
		}

		// Guard against path traversal
		if strings.Contains(relPath, "..") || filepath.IsAbs(relPath) {
			return "", fmt.Errorf("invalid path in zip: %s", relPath)
		}

		found = true
		destPath := filepath.Join(targetDir, relPath)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, file.Mode()); err != nil {
				return "", err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return "", err
		}

		srcFile, err := file.Open()
		if err != nil {
			return "", err
		}

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			_ = srcFile.Close()
			return "", err
		}

		_, err = io.Copy(destFile, srcFile)
		_ = srcFile.Close()
		_ = destFile.Close()

		if err != nil {
			return "", err
		}
	}

	if !found {
		return "", fmt.Errorf("path %s not found in repository", subpath)
	}

	return revision, nil
}
