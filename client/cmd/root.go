package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"zenith/client/api"
	data "zenith/models"

	"github.com/spf13/cobra"
)

type options struct {
	baseURL string
	timeout time.Duration
	client  *api.Client
}

// NewRootCommand creates fresh flags, writers and options for each invocation.
// Separate command trees can be executed concurrently; one tree is not shared.
func NewRootCommand() *cobra.Command {
	opts := &options{}
	root := &cobra.Command{
		Use: "zenith-client", Short: "Register, inspect and monitor Zenith services",
		SilenceUsage: true, SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.timeout <= 0 {
				return fmt.Errorf("--timeout must be positive")
			}
			client, err := api.New(opts.baseURL, &http.Client{Timeout: opts.timeout, CheckRedirect: api.NoRedirect})
			if err != nil {
				return err
			}
			opts.client = client
			return nil
		},
	}
	root.PersistentFlags().StringVar(&opts.baseURL, "url", "http://127.0.0.1:8080", "Base URL for API")
	root.PersistentFlags().DurationVar(&opts.timeout, "timeout", 10*time.Second, "Timeout per HTTP request, e.g. 500ms or 10s")
	root.AddCommand(newAddCommand(opts), newRemoveCommand(opts), newStatusCommand(opts), newPingCommand(opts), newCheckCommand(opts))
	return root
}

func serviceArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	name, ok := data.NormalizeServiceName(args[0])
	if !ok {
		return "", fmt.Errorf("invalid service name: use 1-256 UTF-8 bytes without control characters")
	}
	return name, nil
}

func ExecuteWithArgs(args []string) (string, error) {
	return ExecuteWithContext(context.Background(), args)
}

func ExecuteWithContext(ctx context.Context, args []string) (string, error) {
	var out bytes.Buffer
	root := NewRootCommand()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.ExecuteContext(ctx)
	return out.String(), err
}

func Execute() {
	// os.Interrupt handles Ctrl+C on every OS and Ctrl+Break on Windows.
	// Windows close/logoff SIGTERM permits cleanup but cannot guarantee its duration.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	root := NewRootCommand()
	err := root.ExecuteContext(ctx)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
