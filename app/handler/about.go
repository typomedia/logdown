package handler

import (
	"os"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// about shows system information. Ports AboutController::info.
func (h *Handler) about(c *fiber.Ctx) error {
	p := h.page("app_about_info")
	p.System = map[string]string{
		"os":    osRelease(),
		"cpu":   cpuModel(),
		"go":    strings.TrimPrefix(runtime.Version(), "go"),
		"fiber": fiber.Version,
	}
	return h.render(c, p)
}

// osRelease returns a human-readable OS description, falling back to GOOS.
func osRelease() string {
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		if v := field(string(data), "PRETTY_NAME="); v != "" {
			return v
		}
	}
	return runtime.GOOS + "/" + runtime.GOARCH
}

// cpuModel returns the CPU model name, falling back to GOARCH.
func cpuModel() string {
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "model name") {
				if i := strings.Index(line, ":"); i >= 0 {
					return strings.TrimSpace(line[i+1:])
				}
			}
		}
	}
	return runtime.GOARCH
}

func field(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, key) {
			return strings.Trim(strings.TrimPrefix(line, key), `"`)
		}
	}
	return ""
}
