package claudecli

import (
	"fmt"
	"os/exec"
	"strings"
)

type CLI struct {
	bin string
}

func New() (*CLI, error) {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return nil, fmt.Errorf("claude binary not found in PATH")
	}
	return &CLI{bin: bin}, nil
}

func (c *CLI) MarketplaceAdd(url string) (string, error) {
	return c.run("plugin", "marketplace", "add", url)
}

func (c *CLI) MarketplaceRemove(name string) error {
	_, err := c.run("plugin", "marketplace", "remove", name)
	return err
}

func (c *CLI) MarketplaceUpdate(name string) error {
	_, err := c.run("plugin", "marketplace", "update", name)
	return err
}

func (c *CLI) PluginInstall(ref string) error {
	_, err := c.run("plugin", "install", ref)
	return err
}

func (c *CLI) PluginUninstall(name string) error {
	_, err := c.run("plugin", "uninstall", name)
	return err
}

func (c *CLI) PluginEnable(name string) error {
	_, err := c.run("plugin", "enable", name)
	return err
}

func (c *CLI) PluginDisable(name string) error {
	_, err := c.run("plugin", "disable", name)
	return err
}

func (c *CLI) run(args ...string) (string, error) {
	cmd := exec.Command(c.bin, args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if output != "" {
			return "", fmt.Errorf("claude %s failed: %s", strings.Join(args, " "), output)
		}
		return "", fmt.Errorf("claude %s failed: %w", strings.Join(args, " "), err)
	}
	return output, nil
}
