package main

import (
	"fmt"
	"os"
)

func runAdd(slug string) error {
	if slug == "" {
		return fmt.Errorf("usage: qomp add <component>")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg, root, err := findConfig(cwd)
	if err != nil {
		return err
	}

	client := newClient(cfg.Registry)
	if _, err := client.fetchManifest(); err != nil {
		return err
	}
	if _, ok := client.findComponent(slug); !ok {
		return fmt.Errorf("no component named %q available", slug)
	}

	fmt.Fprintf(os.Stderr, "Installing %s …\n", slug)
	if err := installComponent(root, cfg, client, slug); err != nil {
		return err
	}
	ensureComponentListed(&cfg, slug)
	if cfg.Hashes == nil {
		cfg.Hashes = map[string]string{}
	}
	if c, ok := client.findComponent(slug); ok {
		cfg.Hashes["component:"+slug] = c.Hash
	}
	if err := saveConfig(root, cfg); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Added %s → %s\n", slug, cfg.Paths.Components)
	return nil
}
