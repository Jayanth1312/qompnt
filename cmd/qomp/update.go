package main

import (
	"fmt"
	"os"
)

func runUpdate() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg, root, err := findConfig(cwd)
	if err != nil {
		return err
	}

	client := newClient(cfg.Registry)
	fmt.Fprintf(os.Stderr, "Fetching registry from %s …\n", client.Base)
	m, err := client.fetchManifest()
	if err != nil {
		return err
	}
	if cfg.Hashes == nil {
		cfg.Hashes = map[string]string{}
	}

	stylesKey := "styles:" + cfg.Theme + ":" + cfg.Accent + ":" + m.Version
	needStyles := cfg.Hashes["styles"] != stylesKey
	if t, ok := client.findTheme(cfg.Theme); ok && t.Hash != "" {
		if cfg.Hashes["theme"] != t.Hash {
			needStyles = true
		}
	}
	if needStyles {
		fmt.Fprintf(os.Stderr, "Updating styles (theme %q) …\n", cfg.Theme)
		if err := installStyles(root, cfg, client); err != nil {
			return err
		}
		cfg.Hashes["styles"] = stylesKey
		if t, ok := client.findTheme(cfg.Theme); ok && t.Hash != "" {
			cfg.Hashes["theme"] = t.Hash
		}
	} else {
		fmt.Fprintln(os.Stderr, "Styles up to date.")
	}

	updated := 0
	skipped := 0
	missing := 0
	for _, slug := range cfg.Components {
		c, ok := client.findComponent(slug)
		if !ok {
			fmt.Fprintf(os.Stderr, "skip %s: no longer in registry\n", slug)
			missing++
			continue
		}
		key := "component:" + slug
		if cfg.Hashes[key] == c.Hash {
			skipped++
			continue
		}
		fmt.Fprintf(os.Stderr, "Updating %s …\n", slug)
		if err := installComponent(root, cfg, client, slug); err != nil {
			return err
		}
		cfg.Hashes[key] = c.Hash
		updated++
	}

	cfg.Hashes["manifest"] = m.Version
	if err := saveConfig(root, cfg); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Done: %d updated, %d already current", updated, skipped)
	if missing > 0 {
		fmt.Fprintf(os.Stderr, ", %d missing", missing)
	}
	fmt.Fprintln(os.Stderr)
	return nil
}
