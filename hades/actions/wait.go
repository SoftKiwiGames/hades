package actions

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SoftKiwiGames/hades/hades/schema"
	"github.com/SoftKiwiGames/hades/hades/types"
)

type WaitAction struct {
	Message string
	Timeout string
}

func NewWaitAction(action *schema.ActionWait) Action {
	return &WaitAction{
		Message: action.Message,
		Timeout: action.Timeout,
	}
}

func (a *WaitAction) Execute(ctx context.Context, runtime *types.Runtime) error {
	message := a.Message
	if message == "" {
		message = "Continue?"
	}

	// Parse timeout if provided
	var timeout time.Duration
	var err error
	if a.Timeout != "" {
		timeout, err = time.ParseDuration(a.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout format: %w", err)
		}
	}

	// Create a channel for the user response
	responseChan := make(chan bool, 1)

	// Start goroutine to read user input
	go func() {
		fmt.Printf("\n⏸️  %s [y/N]: ", message)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			responseChan <- false
			return
		}
		response = strings.TrimSpace(strings.ToLower(response))
		responseChan <- (response == "y" || response == "yes")
	}()

	// Wait for response or timeout
	if timeout > 0 {
		select {
		case approved := <-responseChan:
			if !approved {
				return fmt.Errorf("user declined to continue")
			}
		case <-time.After(timeout):
			return fmt.Errorf("wait timed out after %s", timeout)
		case <-ctx.Done():
			return ctx.Err()
		}
	} else {
		select {
		case approved := <-responseChan:
			if !approved {
				return fmt.Errorf("user declined to continue")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (a *WaitAction) DryRun(ctx context.Context, runtime *types.Runtime) string {
	message := a.Message
	if message == "" {
		message = "Continue?"
	}
	if a.Timeout != "" {
		return fmt.Sprintf("wait: %s (timeout: %s)", message, a.Timeout)
	}
	return fmt.Sprintf("wait: %s", message)
}
