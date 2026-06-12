package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Command pairs a subcommand name and one-line help with its tagged config
// struct; the Cfg struct uses the same env:/default:/help:/name: tag vocabulary
// as any other env.Configure target. Run, when set, is invoked after Configure
// populates Cfg, receiving the environment Path.
//
//	env.Command{
//		Name: "pull",
//		Help: "retrieve and stage records",
//		Cfg:  &pull,
//		Run:  func(path *env.Path) { pull.exec(path) },
//	}
type Command struct {
	Name string           // subcommand word, eg. "pull"
	Help string           // one-line description shown in the top-level menu
	Cfg  any              // pointer to a tagged config struct
	Run  func(path *Path) // optional handler, invoked after Cfg is populated
}

// Commands provides go-tool style subcommand dispatch where the full set of
// subcommands is enumerated from the top-level help; the selected command name
// is returned alongside the environment Path.
//
//	prog                -> menu (every command)
//	prog help           -> menu
//	prog help <cmd>     -> the command's flag table
//	prog <cmd> help     -> the command's flag table
//	prog version        -> version banner
//	prog <cmd> [flags]  -> Configure(cmd.Cfg) on the remaining args
//
// Pass Options to control Silent/NoHelp/SetENV/NoExit, as with env.Configure;
// a nil Options uses the defaults. The subcommand word is consumed from os.Args
// before Configure runs, so each command parses only its own flags and ordered
// positionals.
func Commands(opt *Options, cmds ...Command) (path *Path, name string) {

	if opt == nil {
		opt = new(Options)
	}

	argv := os.Args[1:]
	switch {

	case len(argv) == 0:
		menu(cmds)
		if opt.NoExit {
			return nil, ""
		}
		os.Exit(0)

	case argv[0] == "version", argv[0] == "-version", argv[0] == "--version":
		// defer to Configure's existing version banner
		return Configure(opt), ""

	case argv[0] == "help", argv[0] == "-h", argv[0] == "-help", argv[0] == "--help":
		if len(argv) > 1 {
			if c := lookup(cmds, argv[1]); c != nil {
				commandHelp(c, opt)
				if opt.NoExit {
					return nil, c.Name
				}
				os.Exit(0)
			}
		}
		menu(cmds)
		if opt.NoExit {
			return nil, ""
		}
		os.Exit(0)
	}

	c := lookup(cmds, argv[0])
	if c == nil {
		fmt.Fprintf(os.Stderr, "%s: unknown command %q\n",
			filepath.Base(os.Args[0]), argv[0])
		menu(cmds)
		if opt.NoExit {
			return nil, ""
		}
		os.Exit(2)
	}

	// honor `prog <cmd> help|-help` as a drill-in for the command
	if len(argv) > 1 {
		switch strings.TrimLeft(argv[1], "-") {
		case "help", "h":
			commandHelp(c, opt)
			if opt.NoExit {
				return nil, c.Name
			}
			os.Exit(0)
		}
	}

	// strip the subcommand word so Configure parses only the command's own
	// flags and ordered positionals from os.Args
	os.Args = append(os.Args[:1], os.Args[2:]...)
	path = Configure(opt, c.Cfg)
	if c.Run != nil {
		c.Run(path)
	}
	return path, c.Name
}

// lookup returns the named command or nil.
func lookup(cmds []Command, name string) *Command {
	for i := range cmds {
		if cmds[i].Name == name {
			return &cmds[i]
		}
	}
	return nil
}

// menu renders the go-tool style top-level help enumerating every subcommand.
func menu(cmds []Command) {

	name := filepath.Base(os.Args[0])
	fmt.Printf("\n %s\n%s\n version %s\n build   %s\n\n",
		name, strings.Repeat("-", 40), Version, Build)
	if len(Description) > 0 {
		fmt.Printf(" %s\n\n", Description)
	}

	fmt.Println(" commands:")
	for i := range cmds {
		fmt.Printf("   %-12s %s\n", cmds[i].Name, cmds[i].Help)
	}
	fmt.Printf("\n use \"%s help <command>\" for details\n\n", name)
}

// commandHelp renders a command-scoped header followed by the field table.
func commandHelp(c *Command, opt *Options) {

	name := filepath.Base(os.Args[0])
	fmt.Printf("\n %s %s\n%s\n", name, c.Name, strings.Repeat("-", 40))
	if len(c.Help) > 0 {
		fmt.Printf(" %s\n", c.Help)
	}
	fmt.Println()
	if !opt.NoHelp {
		usage(c.Cfg)
	}
	fmt.Println()
}
