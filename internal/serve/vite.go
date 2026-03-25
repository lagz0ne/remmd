package serve

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// runViteDev spawns the Vite dev server as a subprocess and blocks until it exits.
// It polls TCP readiness with a 30s timeout before returning control.
func runViteDev(ctx context.Context, dir string, port int) error {
	portStr := fmt.Sprintf("%d", port)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	args := []string{"--port", portStr, "--host", "127.0.0.1", "--strictPort"}

	viteBin := resolveViteBin(dir)
	var cmd *exec.Cmd
	if viteBin != "" {
		cmd = exec.CommandContext(ctx, "node", append([]string{viteBin}, args...)...)
	} else {
		cmd = exec.CommandContext(ctx, "npx", append([]string{"vite"}, args...)...)
	}

	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
	cmd.WaitDelay = 5 * time.Second

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("vite start: %w", err)
	}

	// Poll for TCP readiness
	deadline := time.After(30 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			cmd.Cancel()
			return fmt.Errorf("vite readiness timeout after 30s")
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
			if err == nil {
				conn.Close()
				slog.Info("vite dev server ready", "addr", addr)
				return cmd.Wait()
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// resolveViteBin finds the vite binary, checking dir/node_modules/.bin/vite
// and then walking one level up.
func resolveViteBin(dir string) string {
	candidates := []string{
		filepath.Join(dir, "node_modules", ".bin", "vite"),
		filepath.Join(dir, "..", "node_modules", ".bin", "vite"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}
