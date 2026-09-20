package cli

import (
	"context"
	"os"
	"time"

	"github.com/thedavidweng/qualtrics-cli/internal/config"
	"github.com/thedavidweng/qualtrics-cli/internal/errors"
	"github.com/thedavidweng/qualtrics-cli/internal/output"
	"github.com/thedavidweng/qualtrics-cli/internal/qualtrics"
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
