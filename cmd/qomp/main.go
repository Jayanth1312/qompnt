package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Set by goreleaser via -ldflags "-X main.version=…".
var version = "dev"

// defaultRegistry is used when neither --registry nor QOMP_REGISTRY is set.
// Override for local: --registry http://localhost:8080 or QOMP_REGISTRY.
const defaultRegistry = "https://qompnt.vercel.app"

func main() {
	var registry string
	var initFlags initOpts

	root := &cobra.Command{
		Use:   "qomp",
		Short: "Install qompnt themes and HTML components into a project",
		Long: `qomp is a thin installer for qompnt. It pulls themes and components from
the registry, caches them locally, and copies them into your project under
paths recorded in qomp.json.`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(resolveRegistry(registry), initFlags)
		},
	}
	root.PersistentFlags().StringVar(&registry, "registry", "", "registry base URL (or set QOMP_REGISTRY)")
	addInitFlags(root, &initFlags)

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Interactive setup (theme and components)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(resolveRegistry(registry), initFlags)
		},
	}
	addInitFlags(initCmd, &initFlags)
	root.AddCommand(initCmd)

	root.AddCommand(&cobra.Command{
		Use:   "add [component]",
		Short: "Add a component by slug",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(args[0])
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "update",
		Short: "Refresh installed theme and components from the registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate()
		},
	})

	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Download and install the latest qomp release",
		RunE: func(cmd *cobra.Command, args []string) error {
			check, _ := cmd.Flags().GetBool("check")
			return runUpgrade(check)
		},
	}
	upgradeCmd.Flags().Bool("check", false, "only report if a newer release is available")
	root.AddCommand(upgradeCmd)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func addInitFlags(cmd *cobra.Command, opts *initOpts) {
	cmd.Flags().StringVar(&opts.Theme, "theme", "", "theme id (non-interactive; default claude)")
	cmd.Flags().StringVar(&opts.Accent, "accent", "", "accent hex color (optional)")
	cmd.Flags().StringVar(&opts.Components, "components", "", "all|minimal|none|slug,slug (non-interactive)")
}

func resolveRegistry(flag string) string {
	if flag != "" {
		return flag
	}
	if v := os.Getenv("QOMP_REGISTRY"); v != "" {
		return v
	}
	return defaultRegistry
}
