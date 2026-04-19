package main

import (
	"context"
	"os/exec"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "status")
	err := cmd.Run()
	spew.Dump(ctx.Err(), err)
}
