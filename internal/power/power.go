package power

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

func Reboot() error   { return run("reboot") }
func Poweroff() error { return run("poweroff") }

func run(verb string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "systemctl", verb).CombinedOutput()
	if err != nil && len(out) > 0 {
		return fmt.Errorf("%s", out)
	}
	return err
}
