package env_test

import (
	"os"
	"testing"

	"github.com/zxdev/env"
)

type pullCfg struct {
	Path  *env.Path `env:"-"`
	Since string    `env:"s" default:"24h" help:"lookback window"`
	Limit int       `default:"100" help:"max records"`
}

type exportCfg struct {
	Path   *env.Path `env:"-"`
	Format string    `env:"f" default:"json" help:"output format"`
}

func commands(pull *pullCfg, export *exportCfg) (*env.Path, string) {
	return env.Commands(&env.Options{NoExit: true, Silent: true},
		env.Command{Name: "pull", Help: "retrieve and stage records", Cfg: pull},
		env.Command{Name: "export", Help: "emit processed records", Cfg: export},
	)
}

// TestCommandsMenu spoofs a bare invocation; the top-level menu enumerates
// every subcommand and no command is selected.
func TestCommandsMenu(t *testing.T) {
	os.Args = []string{"prog"}
	var pull pullCfg
	var export exportCfg
	if _, name := commands(&pull, &export); name != "" {
		t.Fatalf("menu: expected no command selected, got %q", name)
	}
}

// TestCommandsHelp spoofs `prog help pull`; the drill-in selects pull and
// renders its field table without parsing/running it.
func TestCommandsHelp(t *testing.T) {
	os.Args = []string{"prog", "help", "pull"}
	var pull pullCfg
	var export exportCfg
	if _, name := commands(&pull, &export); name != "pull" {
		t.Fatalf("help pull: expected %q, got %q", "pull", name)
	}
}

// TestCommandsDispatch spoofs `prog pull -since 48h`; the subcommand word is
// consumed and Configure parses only pull's flags, leaving the default intact
// for unset fields.
func TestCommandsDispatch(t *testing.T) {
	os.Args = []string{"prog", "pull", "-since", "48h"}
	var pull pullCfg
	var export exportCfg
	_, name := commands(&pull, &export)
	if name != "pull" {
		t.Fatalf("dispatch: expected %q, got %q", "pull", name)
	}
	if pull.Since != "48h" {
		t.Fatalf("dispatch: Since = %q, want %q", pull.Since, "48h")
	}
	if pull.Limit != 100 {
		t.Fatalf("dispatch: Limit = %d, want default 100", pull.Limit)
	}
}

// TestCommandsUnknown spoofs an unknown subcommand; nothing is selected and the
// menu is shown (NoExit prevents the os.Exit(2)).
func TestCommandsUnknown(t *testing.T) {
	os.Args = []string{"prog", "nope"}
	var pull pullCfg
	var export exportCfg
	if _, name := commands(&pull, &export); name != "" {
		t.Fatalf("unknown: expected no command selected, got %q", name)
	}
}

// TestCommandsRun spoofs `prog pull -since 12h`; the matched command's Run
// handler fires after Cfg is populated and receives the environment Path.
func TestCommandsRun(t *testing.T) {
	os.Args = []string{"prog", "pull", "-since", "12h"}
	var pull pullCfg
	var ran bool
	var gotPath *env.Path

	env.Commands(&env.Options{NoExit: true, Silent: true},
		env.Command{Name: "pull", Help: "retrieve and stage records", Cfg: &pull,
			Run: func(path *env.Path) { ran, gotPath = true, path }},
	)

	if !ran {
		t.Fatal("run: Run handler did not fire")
	}
	if gotPath == nil {
		t.Fatal("run: Run handler received nil Path")
	}
	if pull.Since != "12h" {
		t.Fatalf("run: Since = %q, want %q (Cfg must be populated before Run)", pull.Since, "12h")
	}
}
