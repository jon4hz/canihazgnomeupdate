package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Masterminds/semver"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/jon4hz/canihazgnomeupdate/extensions"
	"github.com/jon4hz/canihazgnomeupdate/gnome"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "canihazgnomeupdate",
	Run:  root,
	Args: cobra.ExactArgs(1),
}

var rootCmdFlags struct {
	enabled bool
}

func init() {
	rootCmd.Flags().BoolVar(&rootCmdFlags.enabled, "enabled", true, "Only search for enabled extensions")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func root(cmd *cobra.Command, args []string) {
	gnomeVersion := args[0]

	// list local extensions
	exts, err := extensions.List(rootCmdFlags.enabled)
	if err != nil {
		log.Fatalln("Failed to get installed extensions: ", err)
	}

	results := make([]checkResult, 0, len(exts))

	spinner := spinner.New().
		Type(spinner.Points).
		Title("  Checking extensions...").
		Style(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		Action(func() {
			for _, ext := range exts {
				result, err := gnome.SearchExtension(ext)
				if err != nil {
					log.Fatalln("Failed to search extensions: ", err)
				}

				results = append(results, checkResult{
					UUID:            ext,
					URL:             result.URL,
					LatestVersion:   latestExtensionVersion(result),
					UpdateSupported: supportsUpdate(result, gnomeVersion),
				})
			}
		})

	if err := spinner.Run(); err != nil {
		log.Fatalln("Failed to run spinner: ", err)
	}

	printResults(results)
	summary(results)
}

func latestExtensionVersion(ext *gnome.Extension) string {
	newestSem := semver.MustParse("0")
	var newest string
	for shellVersion := range ext.ShellVersionMap {
		ver, err := semver.NewVersion(shellVersion)
		if err != nil {
			log.Printf("Failed to parse version %q of extension %q: %s", shellVersion, ext.UUID, err)
			continue
		}
		if ver.GreaterThan(newestSem) {
			newestSem = ver
			newest = shellVersion
		}
	}
	return newest
}

func supportsUpdate(ext *gnome.Extension, gnomeVersion string) bool {
	_, ok := ext.ShellVersionMap[gnomeVersion]
	return ok
}

type checkResult struct {
	UUID            string
	URL             string
	LatestVersion   string
	UpdateSupported bool
}

func printResults(res []checkResult) {
	rows := make([][]string, 0, len(res))
	for _, r := range res {
		var updateSupported string
		if r.UpdateSupported {
			updateSupported = "✅"
		} else {
			updateSupported = "❌"
		}
		rows = append(rows, []string{
			updateSupported,
			r.UUID,
			r.LatestVersion,
			r.URL,
		})
	}

	re := lipgloss.NewRenderer(os.Stdout)
	var (
		cellStyle = re.NewStyle().Padding(0, 1)
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return cellStyle.Copy().Bold(true).Align(lipgloss.Center)
			case col == 2:
				return cellStyle.Copy().Align(lipgloss.Center)
			default:
				return cellStyle
			}
		}).
		Headers("Ok?", "UUID", "Latest", "URL").
		Rows(rows...)

	fmt.Println()
	fmt.Println(t)
	fmt.Println()
}

func summary(res []checkResult) {
	ok := true
	for _, r := range res {
		if !r.UpdateSupported {
			ok = false
			break
		}
	}

	if ok {
		fmt.Println("😸 You can haz gnome update!")
	} else {
		fmt.Println("😿 You can't haz gnome update!")
	}
}
