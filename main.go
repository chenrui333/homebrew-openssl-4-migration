package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/chenrui333/homebrew-openssl-4-migration/internal/checklist"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/deptree"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/migrate"
	"github.com/chenrui333/homebrew-openssl-4-migration/internal/status"
)

const defaultHomebrewCore = "/opt/homebrew/Library/Taps/homebrew/homebrew-core"

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "sslmigrate",
		Short: "Harness for migrating homebrew-core formulae from openssl@3 to openssl@4",
	}
	root.AddCommand(depTreeCmd(), statusCmd(), checklistCmd(), migrateCmd())
	return root
}

func depTreeCmd() *cobra.Command {
	var homebrewCore, output string
	cmd := &cobra.Command{
		Use:   "dep-tree",
		Short: "Scan homebrew-core and write data/dep_tree.json",
		RunE: func(_ *cobra.Command, _ []string) error {
			tree, err := deptree.Build(homebrewCore)
			if err != nil {
				return err
			}
			pending := 0
			done := 0
			for _, f := range tree.Formulae {
				if f.OpenSSLDependency == "openssl@3" {
					pending++
				} else {
					done++
				}
			}
			if err := deptree.Save(tree, output); err != nil {
				return err
			}
			fmt.Printf("Wrote %s (%d formulae: %d pending, %d done)\n",
				output, tree.FormulaCount, pending, done)
			return nil
		},
	}
	cmd.Flags().StringVar(&homebrewCore, "homebrew-core", envOr("HOMEBREW_CORE", defaultHomebrewCore), "path to homebrew-core checkout")
	cmd.Flags().StringVar(&output, "output", "data/dep_tree.json", "output JSON path")
	return cmd
}

func statusCmd() *cobra.Command {
	var homebrewCore, depTree, output string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print migration dashboard and regenerate TRACKING.md",
		RunE: func(_ *cobra.Command, _ []string) error {
			return status.Run(homebrewCore, depTree, output)
		},
	}
	cmd.Flags().StringVar(&homebrewCore, "homebrew-core", envOr("HOMEBREW_CORE", defaultHomebrewCore), "path to homebrew-core checkout")
	cmd.Flags().StringVar(&depTree, "dep-tree", "data/dep_tree.json", "dependency inventory JSON")
	cmd.Flags().StringVar(&output, "output", "TRACKING.md", "tracking markdown output path")
	return cmd
}

func checklistCmd() *cobra.Command {
	var homebrewCore, depTree, output string
	cmd := &cobra.Command{
		Use:   "checklist",
		Short: "Write CHECKLIST.md with markdown checkboxes grouped by migration batch",
		RunE: func(_ *cobra.Command, _ []string) error {
			return checklist.Run(homebrewCore, depTree, output)
		},
	}
	cmd.Flags().StringVar(&homebrewCore, "homebrew-core", envOr("HOMEBREW_CORE", defaultHomebrewCore), "path to homebrew-core checkout")
	cmd.Flags().StringVar(&depTree, "dep-tree", "data/dep_tree.json", "dependency inventory JSON")
	cmd.Flags().StringVar(&output, "output", "CHECKLIST.md", "checklist markdown output path")
	return cmd
}

func migrateCmd() *cobra.Command {
	var homebrewCore, depTree, pushRemote string
	var dryRun, noPR, resetExisting bool
	cmd := &cobra.Command{
		Use:   "migrate <formula>",
		Short: "Migrate a formula from openssl@3 to openssl@4 and open a PR",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return migrate.Run(args[0], migrate.Options{
				HomebrewCore:  homebrewCore,
				DepTreePath:   depTree,
				DryRun:        dryRun,
				NoPR:          noPR,
				PushRemote:    pushRemote,
				ResetExisting: resetExisting,
			})
		},
	}
	cmd.Flags().StringVar(&homebrewCore, "homebrew-core", envOr("HOMEBREW_CORE", defaultHomebrewCore), "path to homebrew-core checkout")
	cmd.Flags().StringVar(&depTree, "dep-tree", "data/dep_tree.json", "dependency inventory JSON")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print planned diff without modifying homebrew-core")
	cmd.Flags().BoolVar(&noPR, "no-pr", false, "create local commit but skip push and PR")
	cmd.Flags().BoolVar(&resetExisting, "reset-existing", false, "reset an existing branch to origin/<base> (destructive)")
	cmd.Flags().StringVar(&pushRemote, "push-remote", envOr("HOMEBREW_CORE_PUSH_REMOTE", ""), "fork remote to push to")
	return cmd
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
