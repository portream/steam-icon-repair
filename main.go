package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const cdnURL = "https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/%s/%s.ico"

type shortcut struct {
	file     string
	appID    string
	iconPath string
	iconHash string
}

var (
	rxURL  = regexp.MustCompile(`(?i)^URL\s*=\s*steam://rungameid/(\d+)`)
	rxIcon = regexp.MustCompile(`(?i)^IconFile\s*=\s*(.+)$`)
)

func main() {
	dryRun := flag.Bool("dry-run", false, "only list broken icons, do not download")
	verbose := flag.Bool("v", false, "verbose output")
	flag.Parse()

	fmt.Println("Steam Icon Repair")
	fmt.Println("=================")

	dirs := shortcutDirs()
	var all []shortcut
	for _, d := range dirs {
		if *verbose {
			fmt.Printf("Scanning: %s\n", d)
		}
		all = append(all, scan(d)...)
	}

	if len(all) == 0 {
		fmt.Println("No Steam shortcuts found.")
		return
	}
	fmt.Printf("Steam shortcuts found: %d\n", len(all))

	var broken []shortcut
	for _, s := range all {
		if _, err := os.Stat(s.iconPath); os.IsNotExist(err) {
			broken = append(broken, s)
		} else if *verbose {
			fmt.Printf("  OK  %s (appid %s)\n", filepath.Base(s.file), s.appID)
		}
	}

	if len(broken) == 0 {
		fmt.Println("All icons are present. Nothing to repair.")
		return
	}
	fmt.Printf("Missing icons: %d\n\n", len(broken))

	if *dryRun {
		for _, s := range broken {
			fmt.Printf("  MISSING  appid=%s  %s\n", s.appID, s.iconPath)
		}
		fmt.Println("\n(dry-run: nothing downloaded)")
		return
	}

	client := &http.Client{Timeout: 25 * time.Second}
	fixed, failed := 0, 0
	for _, s := range broken {
		fmt.Printf("  appid=%s  %s ... ", s.appID, filepath.Base(s.iconPath))
		if err := download(client, s); err != nil {
			fmt.Printf("FAIL (%v)\n", err)
			failed++
			continue
		}
		fmt.Println("OK")
		fixed++
	}

	fmt.Printf("\nRestored: %d   Failed: %d\n", fixed, failed)
	if fixed > 0 {
		fmt.Println("\nTo refresh icons in Explorer, run:")
		fmt.Println("  ie4uinit.exe -show")
		fmt.Println("If icons still look stale, sign out and back in.")
	}
}

func shortcutDirs() []string {
	var out []string
	add := func(p string) {
		if p == "" {
			return
		}
		if _, err := os.Stat(p); err == nil {
			out = append(out, p)
		}
	}
	if h, err := os.UserHomeDir(); err == nil {
		add(filepath.Join(h, "Desktop"))
	}
	add(filepath.Join(os.Getenv("PUBLIC"), "Desktop"))
	add(filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs"))
	add(filepath.Join(os.Getenv("PROGRAMDATA"), "Microsoft", "Windows", "Start Menu", "Programs"))
	return out
}

func scan(root string) []shortcut {
	var out []shortcut
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(path), ".url") {
			return nil
		}
		if s, ok := parse(path); ok {
			out = append(out, s)
		}
		return nil
	})
	return out
}

func parse(path string) (shortcut, bool) {
	f, err := os.Open(path)
	if err != nil {
		return shortcut{}, false
	}
	defer f.Close()

	s := shortcut{file: path}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if m := rxURL.FindStringSubmatch(line); m != nil {
			s.appID = m[1]
		} else if m := rxIcon.FindStringSubmatch(line); m != nil {
			s.iconPath = strings.TrimSpace(m[1])
		}
	}
	if s.appID == "" || s.iconPath == "" {
		return shortcut{}, false
	}
	base := filepath.Base(s.iconPath)
	s.iconHash = strings.TrimSuffix(base, filepath.Ext(base))
	return s, true
}

func download(client *http.Client, s shortcut) error {
	url := fmt.Sprintf(cdnURL, s.appID, s.iconHash)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "steam-icon-repair/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if err := os.MkdirAll(filepath.Dir(s.iconPath), 0o755); err != nil {
		return err
	}
	tmp := s.iconPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, s.iconPath)
}
