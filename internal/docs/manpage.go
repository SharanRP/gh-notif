package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ManPageGenerator generates man pages for commands
type ManPageGenerator struct {
	// OutputDir is the directory where man pages will be written
	OutputDir string

	// Section is the man page section (1-9)
	Section int

	// Date is the date to include in the man page
	Date string

	// Footer is the footer to include in the man page
	Footer string

	// Header is the header to include in the man page
	Header string
}

// NewManPageGenerator creates a new man page generator
func NewManPageGenerator(outputDir string) *ManPageGenerator {
	return &ManPageGenerator{
		OutputDir: outputDir,
		Section:   1,
		Date:      time.Now().Format("January 2006"),
		Footer:    "gh-notif",
		Header:    "User Commands",
	}
}

// GenerateManPages generates man pages for all commands
func (g *ManPageGenerator) GenerateManPages(rootCmd *cobra.Command) error {
	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// Generate the man page for the root command
	if err := g.GenerateManPage(rootCmd, ""); err != nil {
		return fmt.Errorf("error generating man page for root command: %w", err)
	}

	// Generate man pages for all subcommands
	return g.generateManPagesForSubcommands(rootCmd, "")
}

// generateManPagesForSubcommands generates man pages for all subcommands
func (g *ManPageGenerator) generateManPagesForSubcommands(cmd *cobra.Command, parentName string) error {
	for _, subCmd := range cmd.Commands() {
		// Skip hidden commands
		if subCmd.Hidden {
			continue
		}

		// Generate the command name
		var cmdName string
		if parentName == "" {
			cmdName = subCmd.Name()
		} else {
			cmdName = fmt.Sprintf("%s-%s", parentName, subCmd.Name())
		}

		// Generate the man page for this command
		if err := g.GenerateManPage(subCmd, cmdName); err != nil {
			return fmt.Errorf("error generating man page for command %s: %w", cmdName, err)
		}

		// Generate man pages for this command's subcommands
		if err := g.generateManPagesForSubcommands(subCmd, cmdName); err != nil {
			return err
		}
	}

	return nil
}

// GenerateManPage generates a man page for a command
func (g *ManPageGenerator) GenerateManPage(cmd *cobra.Command, cmdName string) error {
	// Determine the command name
	name := cmd.Name()
	if cmdName != "" {
		name = cmdName
	}

	// Determine the file path
	filePath := filepath.Join(g.OutputDir, fmt.Sprintf("%s.%d", name, g.Section))

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating man page file: %w", err)
	}
	defer file.Close()

	// Write the man page
	if err := g.writeManPage(file, cmd, name); err != nil {
		return fmt.Errorf("error writing man page: %w", err)
	}

	return nil
}

// writeManPage writes a man page to a file
func (g *ManPageGenerator) writeManPage(file *os.File, cmd *cobra.Command, name string) error {
	// Write the header
	if _, err := fmt.Fprintf(file, ".TH %s %d \"%s\" \"%s\" \"%s\"\n", strings.ToUpper(name), g.Section, g.Date, g.Footer, g.Header); err != nil {
		return err
	}

	// Write the name section
	if _, err := fmt.Fprintf(file, ".SH NAME\n%s \\- %s\n", name, cmd.Short); err != nil {
		return err
	}

	// Write the synopsis section
	if _, err := fmt.Fprintf(file, ".SH SYNOPSIS\n.B %s\n", name); err != nil {
		return err
	}
	if cmd.HasSubCommands() {
		if _, err := fmt.Fprintf(file, "[\\fICOMMAND\\fR] "); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(file, "[\\fIOPTIONS\\fR]\n"); err != nil {
		return err
	}

	// Write the description section
	if _, err := fmt.Fprintf(file, ".SH DESCRIPTION\n.B %s\n%s\n", name, cmd.Long); err != nil {
		return err
	}

	// Write the commands section if there are subcommands
	if cmd.HasSubCommands() {
		if _, err := fmt.Fprintf(file, ".SH COMMANDS\n"); err != nil {
			return err
		}

		for _, subCmd := range cmd.Commands() {
			// Skip hidden commands
			if subCmd.Hidden {
				continue
			}

			if _, err := fmt.Fprintf(file, ".TP\n.BR %s\n%s\n", subCmd.Name(), subCmd.Short); err != nil {
				return err
			}
		}
	}

	// Write the options section
	if _, err := fmt.Fprintf(file, ".SH OPTIONS\n"); err != nil {
		return err
	}

	// Add global flags
	cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		// Skip hidden flags
		if flag.Hidden {
			return
		}

		// Format the flag
		if flag.Shorthand != "" {
			fmt.Fprintf(file, ".TP\n.BR \\-\\-%s \" \" \\fI%s\\fR, \" \" \\-%s\n%s\n", flag.Name, flag.Value.Type(), flag.Shorthand, flag.Usage)
		} else {
			fmt.Fprintf(file, ".TP\n.BR \\-\\-%s \" \" \\fI%s\\fR\n%s\n", flag.Name, flag.Value.Type(), flag.Usage)
		}
	})

	// Add local flags
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		// Skip hidden flags
		if flag.Hidden {
			return
		}

		// Format the flag
		if flag.Shorthand != "" {
			fmt.Fprintf(file, ".TP\n.BR \\-\\-%s \" \" \\fI%s\\fR, \" \" \\-%s\n%s\n", flag.Name, flag.Value.Type(), flag.Shorthand, flag.Usage)
		} else {
			fmt.Fprintf(file, ".TP\n.BR \\-\\-%s \" \" \\fI%s\\fR\n%s\n", flag.Name, flag.Value.Type(), flag.Usage)
		}
	})

	// Write the examples section if there are examples
	if cmd.Example != "" {
		if _, err := fmt.Fprintf(file, ".SH EXAMPLES\n"); err != nil {
			return err
		}

		// Split the examples by newline
		examples := strings.Split(cmd.Example, "\n")
		for _, example := range examples {
			// Skip empty lines
			if strings.TrimSpace(example) == "" {
				continue
			}

			// Check if this is a comment or a command
			if strings.HasPrefix(strings.TrimSpace(example), "#") {
				// This is a comment
				if _, err := fmt.Fprintf(file, ".TP\n%s\n", strings.TrimSpace(example)); err != nil {
					return err
				}
			} else {
				// This is a command
				if _, err := fmt.Fprintf(file, ".B %s\n", strings.TrimSpace(example)); err != nil {
					return err
				}
			}
		}
	}

	// Write the see also section
	if _, err := fmt.Fprintf(file, ".SH SEE ALSO\n"); err != nil {
		return err
	}

	// Add the root command
	rootCmd := cmd
	for rootCmd.Parent() != nil {
		rootCmd = rootCmd.Parent()
	}
	if _, err := fmt.Fprintf(file, ".BR %s (1),\n", rootCmd.Name()); err != nil {
		return err
	}

	// Add sibling commands
	if cmd.Parent() != nil {
		for _, siblingCmd := range cmd.Parent().Commands() {
			// Skip hidden commands and the current command
			if siblingCmd.Hidden || siblingCmd == cmd {
				continue
			}

			if _, err := fmt.Fprintf(file, ".BR %s-%s (1),\n", rootCmd.Name(), siblingCmd.Name()); err != nil {
				return err
			}
		}
	}

	// Write the author section
	if _, err := fmt.Fprintf(file, ".SH AUTHOR\n%s Contributors\n", rootCmd.Name()); err != nil {
		return err
	}

	// Write the copyright section
	if _, err := fmt.Fprintf(file, ".SH COPYRIGHT\nCopyright © %s %s Contributors. License MIT.\n", time.Now().Format("2006"), rootCmd.Name()); err != nil {
		return err
	}

	return nil
}

// GenerateManPagesForCommand generates man pages for a command and its subcommands
func GenerateManPagesForCommand(cmd *cobra.Command, outputDir string) error {
	generator := NewManPageGenerator(outputDir)
	return generator.GenerateManPages(cmd)
}
