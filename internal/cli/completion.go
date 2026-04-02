package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for openskills.

To load completions:

  bash:
    source <(openskills completion bash)
    # or add to ~/.bashrc:
    echo 'source <(openskills completion bash)' >> ~/.bashrc

  zsh:
    source <(openskills completion zsh)
    # or install permanently:
    openskills completion zsh > "${fpath[1]}/_openskills"

  fish:
    openskills completion fish | source
    # or install permanently:
    openskills completion fish > ~/.config/fish/completions/openskills.fish

  powershell:
    openskills completion powershell | Out-String | Invoke-Expression
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}
	return cmd
}
