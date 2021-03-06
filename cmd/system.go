package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/blang/semver"
	"github.com/hashicorp/go-getter"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
	"github.com/theckman/yacspin"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "INTERCEPT / SYSTEM - Setup and Update intercept and its core system tools to run AUDIT",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		if systemUpdate {

			confirmAndSelfUpdate()

		}

		if systemSetup {

			updateCore()

		} else {

			fmt.Println("│")
			fmt.Println("│ TEMP ")
			fmt.Println("│")

		}

		PrintClose()

	},
}

func init() {

	systemCmd.PersistentFlags().BoolP("setup", "s", false, "Setup core tools")
	systemCmd.PersistentFlags().BoolP("auto", "a", false, "Non interactive mode")
	systemCmd.PersistentFlags().BoolP("update", "u", false, "Update to the lastest version")

	rootCmd.AddCommand(systemCmd)

}

func confirmAndSelfUpdate() {

	var current string

	fmt.Println("│")
	fmt.Println("├ Binary Update")
	fmt.Println("│")

	latest, found, err := selfupdate.DetectLatest("xfhg/intercept")
	if err != nil {
		fmt.Println("│ Error occurred while detecting version:", err)
		return
	}
	if buildVersion != "" {
		current = buildVersion[1:len(buildVersion)]
	} else {
		current = "0.0.1"
	}
	v := semver.MustParse(current)
	if !found || latest.Version.LTE(v) {
		fmt.Println("│ Current version is the latest")
		return
	}

	cfg := yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[51],
		Suffix:          " Downloading Update",
		SuffixAutoColon: true,
		Message:         latest.Version.String(),
		StopCharacter:   "│ ✓",
		StopColors:      []string{"fgGreen"},
	}

	if !systemAuto {
		fmt.Print("│ Do you want to update to v", latest.Version, " ? (y/n): ")
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil || (input != "y\n" && input != "n\n") {
			fmt.Println("│ Invalid input")
			return
		}
		if input == "n\n" {
			return
		}
	} else {
		fmt.Println("│ Automatic Update")

	}

	spinner, err := yacspin.New(cfg)
	if err != nil {
		LogError(err)
	}
	fmt.Println("│")
	spinner.Start()

	exe, err := os.Executable()
	if err != nil {
		spinner.StopColors("fgRed")
		spinner.StopCharacter("│ x")
		spinner.Message("Could not locate executable path")
		spinner.Stop()
		return
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		spinner.StopColors("fgRed")
		spinner.StopCharacter("│ x")
		spinner.Message("Error occurred while updating binary")
		spinner.Stop()
		fmt.Println("│")
		fmt.Println("│ Error occurred while updating binary:", err)
		return
	}

	spinner.Stop()

	fmt.Println("│")
	fmt.Println("│ Successfully updated to version", latest.Version)

}

func updateCore() {

	fmt.Println("│")
	fmt.Println("├ Core Tools Setup")
	fmt.Println("│")

	core := ""
	coreDst := GetExecutablePath()

	switch runtime.GOOS {
	case "windows":
		core = "https://github.com/xfhg/intercept/releases/latest/download/i-ripgrep-x86_64-windows.zip"
	case "darwin":
		core = "https://github.com/xfhg/intercept/releases/latest/download/i-ripgrep-x86_64-darwin.zip"
	case "linux":
		core = "https://github.com/xfhg/intercept/releases/latest/download/i-ripgrep-x86_64-linux.zip"
	default:
		colorRedBold.Println("│ OS not supported")
		PrintClose()
		os.Exit(1)
	}

	cfg := yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[51],
		Suffix:          " Downloading Core",
		SuffixAutoColon: true,
		Message:         core,
		StopCharacter:   "│ ✓",
		StopColors:      []string{"fgGreen"},
	}

	spinner, err := yacspin.New(cfg)
	if err != nil {
		LogError(err)
	}
	spinner.Start()

	client := &getter.Client{
		Ctx:  context.Background(),
		Dst:  coreDst,
		Dir:  true,
		Src:  core,
		Mode: getter.ClientModeDir,
	}

	if err := client.Get(); err != nil {
		spinner.StopColors("fgRed")
		spinner.StopCharacter("│ x")
		spinner.Message("Error getting path")
		spinner.Stop()
		fmt.Fprintf(os.Stderr, "│ Error getting path %s : %v", client.Src, err)
		PrintClose()
		os.Exit(1)
	}

	spinner.Message("Creating .ignore file")

	d := []string{"search_regex", "config.yaml"}
	err = WriteLinesOnFile(d, filepath.Join(coreDst, ".ignore"))
	err = WriteLinesOnFile(d, filepath.Join(GetWd(), ".ignore"))

	spinner.Stop()

}
