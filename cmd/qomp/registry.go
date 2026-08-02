package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Manifest struct {
	Version    string      `json:"version"`
	Registry   string      `json:"registry"`
	Minimal    []string    `json:"minimal"`
	Themes     []Theme     `json:"themes"`
	Components []Component `json:"components"`
	Assets     Assets      `json:"assets"`
}

type Theme struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url,omitempty"`
	Hash  string `json:"hash,omitempty"`
}

type Component struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	URL         string   `json:"url"`
	Hash        string   `json:"hash"`
}

type Assets struct {
	Tokens string `json:"tokens"`
	Utils  string `json:"utils"`
}

// registryItem mirrors the site's /r/<slug>.json payload (fields we need).
type registryItem struct {
	Name  string         `json:"name"`
	Files []registryFile `json:"files"`
}

type registryFile struct {
	Path    string `json:"path"`
	Target  string `json:"target"`
	Content string `json:"content"`
}

type Client struct {
	Base   string
	HTTP   *http.Client
	Manifest *Manifest
}

func newClient(base string) *Client {
	return &Client{
		Base: strings.TrimRight(base, "/"),
		HTTP: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) fetchManifest() (*Manifest, error) {
	url := c.Base + "/r/cli/manifest.json"
	data, err := fetchCached(c.HTTP, url, "")
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}
	if m.Registry == "" {
		m.Registry = c.Base
	}
	c.Manifest = &m
	return &m, nil
}

func (c *Client) findComponent(slug string) (Component, bool) {
	if c.Manifest == nil {
		return Component{}, false
	}
	for _, comp := range c.Manifest.Components {
		if comp.Slug == slug {
			return comp, true
		}
	}
	return Component{}, false
}

func (c *Client) findTheme(id string) (Theme, bool) {
	if c.Manifest == nil {
		return Theme{}, false
	}
	for _, t := range c.Manifest.Themes {
		if t.ID == id {
			return t, true
		}
	}
	return Theme{}, false
}

func (c *Client) fetchItem(url string) (registryItem, error) {
	data, err := fetchCached(c.HTTP, url, "")
	if err != nil {
		return registryItem{}, err
	}
	var item registryItem
	if err := json.Unmarshal(data, &item); err != nil {
		return registryItem{}, fmt.Errorf("registry item: %w", err)
	}
	return item, nil
}
