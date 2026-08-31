package libs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	LibsDir     = "libs"
	GitHubBase  = "https://raw.githubusercontent.com"
	OutLibsRepo = "mark-https-gif/out-GO-Programming/main/libs"
)

func GetLibsDir() string {
	if _, err := os.Stat(LibsDir); err == nil {
		return LibsDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return LibsDir
	}
	return filepath.Join(home, ".out", LibsDir)
}

func InitLibsDir() error {
	dir := GetLibsDir()
	return os.MkdirAll(dir, 0755)
}

func Get(url string) error {
	if err := InitLibsDir(); err != nil {
		return fmt.Errorf("cannot create libs dir: %s", err)
	}

	name := extractName(url)
	if name == "" {
		return fmt.Errorf("cannot extract library name from: %s", url)
	}

	if !strings.HasSuffix(name, ".out") {
		name = name + ".out"
	}

	fullURL := normalizeURL(url)
	fmt.Printf("Downloading %s ...\n", fullURL)

	resp, err := http.Get(fullURL)
	if err != nil {
		return fmt.Errorf("download failed: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, fullURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read failed: %s", err)
	}

	dst := filepath.Join(GetLibsDir(), name)
	if err := os.WriteFile(dst, body, 0644); err != nil {
		return fmt.Errorf("write failed: %s", err)
	}

	fmt.Printf("Installed: %s (%d bytes)\n", name, len(body))
	return nil
}

func normalizeURL(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}

	parts := strings.Split(url, "/")
	if len(parts) == 1 {
		return fmt.Sprintf("%s/%s/%s.out", GitHubBase, OutLibsRepo, parts[0])
	}

	if len(parts) == 2 {
		return fmt.Sprintf("%s/%s/%s/main.out", GitHubBase, parts[0], parts[1])
	}

	repo := strings.Join(parts[:2], "/")
	path := strings.Join(parts[2:], "/")
	return fmt.Sprintf("%s/%s/%s/%s", GitHubBase, parts[0], repo, path)
}

func extractName(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return ""
	}
	name := parts[len(parts)-1]
	name = strings.TrimSuffix(name, ".out")
	return name
}

func List() error {
	dir := GetLibsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("No libraries installed.")
		fmt.Printf("Install: out get <library>\n")
		fmt.Printf("Dir: %s\n", dir)
		return nil
	}

	count := 0
	fmt.Println("Installed libraries:")
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".out") {
			info, _ := e.Info()
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			fmt.Printf("  %s (%d bytes)\n", e.Name(), size)
			count++
		}
	}

	if count == 0 {
		fmt.Println("  (empty)")
	}
	fmt.Printf("\nDir: %s\n", dir)
	fmt.Printf("Install: out get <library>\n")
	return nil
}

func Find(name string) (string, bool) {
	dir := GetLibsDir()

	// Try direct name
	path := filepath.Join(dir, name+".out")
	if _, err := os.Stat(path); err == nil {
		return path, true
	}

	// Try with .out extension
	path = filepath.Join(dir, name)
	if _, err := os.Stat(path); err == nil {
		return path, true
	}

	// Try subdirectory
	path = filepath.Join(dir, name, "main.out")
	if _, err := os.Stat(path); err == nil {
		return path, true
	}

	return "", false
}
