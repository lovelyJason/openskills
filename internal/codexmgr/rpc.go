package codexmgr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
)

type rpcClient struct {
	proc   *exec.Cmd
	stdin  *json.Encoder
	stdout *bufio.Scanner
	raw    *os.File
}

func newRPCClient() (*rpcClient, error) {
	bin, err := exec.LookPath("codex")
	if err != nil {
		return nil, fmt.Errorf("codex binary not found in PATH")
	}

	proc := exec.Command(bin, "app-server", "--listen", "stdio://")
	stdinPipe, err := proc.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdoutPipe, err := proc.StdoutPipe()
	if err != nil {
		return nil, err
	}
	proc.Stderr = nil

	if err := proc.Start(); err != nil {
		return nil, fmt.Errorf("failed to start codex app-server: %w", err)
	}

	return &rpcClient{
		proc:   proc,
		stdin:  json.NewEncoder(stdinPipe),
		stdout: bufio.NewScanner(stdoutPipe),
	}, nil
}

func (c *rpcClient) close() {
	if c.proc.Process != nil {
		c.proc.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() { done <- c.proc.Wait() }()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			c.proc.Process.Kill()
			<-done
		}
	}
}

func (c *rpcClient) send(msg map[string]interface{}) error {
	return c.stdin.Encode(msg)
}

func (c *rpcClient) readUntilID(targetID string) (map[string]interface{}, error) {
	for c.stdout.Scan() {
		line := c.stdout.Text()
		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		if id, ok := msg["id"].(string); ok && id == targetID {
			return msg, nil
		}
	}
	return nil, fmt.Errorf("unexpected EOF from codex app-server")
}

func (c *rpcClient) initialize() error {
	initID := uuid.New().String()
	if err := c.send(map[string]interface{}{
		"method": "initialize",
		"id":     initID,
		"params": map[string]interface{}{
			"clientInfo": map[string]interface{}{
				"name":    "openskills",
				"title":   "OpenSkills",
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{
				"experimentalApi": true,
			},
		},
	}); err != nil {
		return fmt.Errorf("send initialize: %w", err)
	}

	resp, err := c.readUntilID(initID)
	if err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	if errObj, ok := resp["error"]; ok {
		return fmt.Errorf("initialize error: %v", errObj)
	}

	return c.send(map[string]interface{}{
		"method": "initialized",
	})
}

func (c *rpcClient) call(method string, params map[string]interface{}) (json.RawMessage, error) {
	reqID := uuid.New().String()
	if err := c.send(map[string]interface{}{
		"method": method,
		"id":     reqID,
		"params": params,
	}); err != nil {
		return nil, fmt.Errorf("send %s: %w", method, err)
	}

	resp, err := c.readUntilID(reqID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", method, err)
	}
	if errObj, ok := resp["error"]; ok {
		return nil, fmt.Errorf("%s error: %v", method, errObj)
	}

	result, _ := json.Marshal(resp["result"])
	return result, nil
}

func callCodexAppServer(method string, params map[string]interface{}) (bool, string) {
	client, err := newRPCClient()
	if err != nil {
		return false, err.Error()
	}
	defer client.close()

	if err := client.initialize(); err != nil {
		return false, err.Error()
	}

	result, err := client.call(method, params)
	if err != nil {
		return false, err.Error()
	}

	return true, string(result)
}

func rpcPluginInstall(pluginName string) (bool, string) {
	return callCodexAppServer("plugin/install", map[string]interface{}{
		"marketplacePath": marketplacePath(),
		"pluginName":      pluginName,
	})
}

func rpcPluginUninstall(pluginName string) (bool, string) {
	pluginID := fmt.Sprintf("%s@%s", pluginName, LocalMarketplaceName)
	return callCodexAppServer("plugin/uninstall", map[string]interface{}{
		"pluginId": pluginID,
	})
}

func hintFromRPCError(detail string) string {
	lower := strings.ToLower(detail)
	switch {
	case strings.Contains(lower, "not found in path"):
		return "Install codex CLI first."
	case strings.Contains(lower, "eof"):
		return "codex app-server exited unexpectedly. Is codex installed correctly?"
	default:
		return detail
	}
}
