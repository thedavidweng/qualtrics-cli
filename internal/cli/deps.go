package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/thedavidweng/qualtrics-cli/internal/config"
	"github.com/thedavidweng/qualtrics-cli/internal/errors"
	"github.com/thedavidweng/qualtrics-cli/internal/output"
	"github.com/thedavidweng/qualtrics-cli/internal/qualtrics"
	"github.com/thedavidweng/qualtrics-cli/internal/safety"
)

type CommandDeps struct {
	Start    time.Time
	Renderer *output.Renderer
	Client   *qualtrics.Client
}

func newDeps(renderer *output.Renderer, command string, start time.Time) (CommandDeps, bool) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		handleError(renderer, command, errors.New(errors.InternalError, "failed to load config", errors.CatInternal, false, err), start)
		return CommandDeps{}, false
	}

	baseURL, err := cfg.BaseURL()
	if err != nil {
		handleError(renderer, command, errors.New(errors.AuthRequired, err.Error(), errors.CatAuth, false, nil), start)
		return CommandDeps{}, false
	}

	token := cfg.Token()
	if token == "" {
		handleError(renderer, command, errors.New(errors.AuthRequired, "no API token configured; run `qualtrics auth set-token`", errors.CatAuth, false, nil), start)
		return CommandDeps{}, false
	}

	return CommandDeps{
		Start:    start,
		Renderer: renderer,
		Client:   qualtrics.NewClient(baseURL, token, timeout),
	}, true
}

func runList[T any](ctx context.Context, command, failMsg string, fn func(context.Context, *qualtrics.Client) ([]T, *output.PaginationMeta, error), human func([]T)) {
	start := time.Now()
	renderer := output.NewRenderer(nil, nil, jsonMode, pretty)

	deps, ok := newDeps(renderer, command, start)
	if !ok {
		return
	}

	items, page, err := fn(ctx, deps.Client)
	if err != nil {
		handleError(renderer, command, wrapError(err, failMsg), start)
		return
	}

	if jsonMode {
		env := output.NewEnvelope(command, profile, output.SchemaVersion, requestID, items, time.Since(start))
		env.Meta.Pagination = page
		renderer.RenderSuccess(env)
		return
	}
	human(items)
}

func runLocal(command string, fn func() any, human func(any)) {
	start := time.Now()
	renderer := output.NewRenderer(nil, nil, jsonMode, pretty)

	data := fn()
	if jsonMode {
		env := output.NewEnvelope(command, profile, output.SchemaVersion, requestID, data, time.Since(start))
		renderer.RenderSuccess(env)
		return
	}
	human(data)
}

func run[T any](ctx context.Context, command, failMsg string, fn func(context.Context, *qualtrics.Client) (T, error), human func(T)) {
	start := time.Now()
	renderer := output.NewRenderer(nil, nil, jsonMode, pretty)

	deps, ok := newDeps(renderer, command, start)
	if !ok {
		return
	}

	data, err := fn(ctx, deps.Client)
	if err != nil {
		handleError(renderer, command, wrapError(err, failMsg), start)
		return
	}

	if jsonMode {
		env := output.NewEnvelope(command, profile, output.SchemaVersion, requestID, data, time.Since(start))
		renderer.RenderSuccess(env)
		return
	}
	human(data)
}

type mutation struct {
	resourceID string
	planAfter  any
	do         func(context.Context, *qualtrics.Client) (any, error)
	human      func()
}

func runMutation(cmd *cobra.Command, command, failMsg string, tier safety.OperationTier, prepare func() (mutation, *errors.Error)) {
	start := time.Now()
	renderer := output.NewRenderer(nil, nil, jsonMode, pretty)

	if err := safety.Check(tier, readOnly, dryRun, confirm); err != nil {
		handleError(renderer, command, err, start)
		return
	}

	m, verr := prepare()
	if verr != nil {
		handleError(renderer, command, verr, start)
		return
	}

	if dryRun {
		plan := map[string]any{
			"command":     command,
			"resource_id": m.resourceID,
			"after":       m.planAfter,
		}
		renderer.RenderSuccess(output.NewEnvelope(command, profile, output.SchemaVersion, requestID, plan, time.Since(start)))
		return
	}

	if tier == safety.TierDestructive {
		if err := confirmDestructive(m.resourceID); err != nil {
			handleError(renderer, command, err, start)
			return
		}
	}

	deps, ok := newDeps(renderer, command, start)
	if !ok {
		return
	}

	data, err := m.do(cmd.Context(), deps.Client)
	if err != nil {
		handleError(renderer, command, wrapError(err, failMsg), start)
		return
	}

	if jsonMode {
		renderer.RenderSuccess(output.NewEnvelope(command, profile, output.SchemaVersion, requestID, data, time.Since(start)))
		return
	}
	m.human()
}

func confirmDestructive(resourceID string) *errors.Error {
	if resourceID == "" {
		return errors.New(errors.ConfirmationRequired, "destructive operation requires --confirm", errors.CatSafety, false, nil)
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		if confirm {
			return nil
		}
		return errors.New(errors.ConfirmationRequired, fmt.Sprintf("destructive operation on %s requires --confirm when stdin is not interactive", resourceID), errors.CatSafety, false, nil)
	}
	fmt.Printf("This deletes %s and cannot be undone. Type %s to confirm: ", resourceID, resourceID)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	if strings.TrimSpace(line) != resourceID {
		return errors.New(errors.ConfirmationRequired, "confirmation did not match; aborted", errors.CatSafety, false, nil)
	}
	return nil
}

func readPayload(path string) ([]byte, error) {
	if path == "" || path == "-" {
		return io.ReadAll(os.Stdin)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read payload %s: %w", path, err)
	}
	return data, nil
}

func writeFile(path string, data []byte) error {
	if path == "" || path == "-" {
		_, err := os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func wrapError(err error, message string) *errors.Error {
	if e, ok := err.(*errors.Error); ok {
		return e
	}
	return errors.New(errors.APIError, message, errors.CatAPI, false, err)
}
func handleError(renderer *output.Renderer, command string, err error, start time.Time) {
	e, ok := err.(*errors.Error)
	if !ok {
		e = errors.New(errors.InternalError, err.Error(), errors.CatInternal, false, err)
	}
	env := output.NewErrorEnvelope(command, profile, output.SchemaVersion, e, time.Since(start))
	renderer.RenderError(env)
	os.Exit(e.ExitCode())
}
