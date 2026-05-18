package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:     "format <file.md>",
	Short:   "Format a markdown file",
	Aliases: []string{"fmt", "f"},
	Args:    cobra.ExactArgs(1),
	RunE:    runFormat,
}

func runFormat(_ *cobra.Command, args []string) error {
	file := args[0]
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", file)
	}
	return formatFile(file)
}

func formatFile(file string) error {
	if path, err := exec.LookPath("prettier"); err == nil {
		fmt.Printf("Formatting %s with prettier\n", file)
		c := exec.Command(path, "--parser", "markdown", "--write", file)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	if path, err := exec.LookPath("deno"); err == nil {
		fmt.Printf("Formatting %s with deno fmt\n", file)
		c := exec.Command(path, "fmt", file)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	if path, err := exec.LookPath("bun"); err == nil {
		fmt.Printf("Formatting %s with bunx prettier\n", file)
		c := exec.Command(path, "x", "prettier", "--parser", "markdown", "--write", file)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	return fmt.Errorf("no markdown formatter found — install prettier (npm i -g prettier), deno (https://deno.land), or bun (https://bun.sh)")
}
