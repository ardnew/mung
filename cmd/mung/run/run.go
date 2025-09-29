// Package run contains the full application logic for the mung CLI.
// The main package should only delegate into this package and exit with
// the returned status code.
package run

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/ardnew/mung"
)

// ExitCode represents a program termination code and implements error.
// It carries a numeric code, a message, and optionally wraps another error.
type ExitCode struct {
	Code int
	Msg  string
	Err  error
}

// Error implements the error interface.
func (e ExitCode) Error() string {
	if e.Msg == "" && e.Code == 0 && e.Err == nil {
		return ""
	}
	if e.Msg == "" {
		if e.Err != nil {
			return e.Err.Error()
		}
		return fmt.Sprintf("exit status %d", e.Code)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

// Unwrap exposes the wrapped error (if any).
func (e ExitCode) Unwrap() error { return e.Err }

// With returns a copy of e that wraps err.
func (e ExitCode) With(err error) ExitCode { e.Err = err; return e }

// Int returns the numeric exit code.
func (e ExitCode) Int() int { return e.Code }

var (
	// success or help/version printed
	ExitOK = ExitCode{Code: 0}
	// flag parse or general argument error
	ExitParseError = ExitCode{Code: 1, Msg: "invalid arguments"}
	// no subjects provided
	ExitNoSubjects = ExitCode{Code: 2, Msg: "no subjects provided"}
	// error expanding subjects (e.g., env lookup)
	ExitSubjectsError = ExitCode{Code: 3, Msg: "failed to expand subjects"}
)

// Main executes the mung CLI and returns an appropriate exit code.
// version is the semantic version for this command passed in by main.
func Main(version string) (string, ExitCode) {
	flags := flagSet{
		FlagSet:    flag.NewFlagSet("mung", flag.ContinueOnError),
		delim:      soloValue{zero: ":", name: "d", desc: "item delimiter"},
		remove:     multiValue{name: "r", desc: "items to remove"},
		prefix:     multiValue{name: "p", desc: "items to prefix subject(s)"},
		suffix:     multiValue{name: "s", desc: "items to suffix subject(s)"},
		filter:     soloValue{name: "t", desc: "command to filter subject(s)"},
		verbose:    incFlag(0),
		version:    incFlag(0),
		cmdVersion: strings.TrimSpace(version),
	}

	// Define command-line flags
	flags.Var(&flags.delim, flags.delim.name, flags.delim.desc)
	flags.Var(&flags.remove, flags.remove.name, flags.remove.desc)
	flags.Var(&flags.prefix, flags.prefix.name, flags.prefix.desc)
	flags.Var(&flags.suffix, flags.suffix.name, flags.suffix.desc)
	flags.Var(&flags.filter, flags.filter.name, flags.filter.desc)
	flags.BoolVar(&flags.nameref, "n", false, "subjects are env NAME references")
	flags.Var(&flags.verbose, "v", "enable verbose output (incremental)")
	flags.Var(&flags.version, "V", "print semantic version of cmd (with module if verbose)")

	flags.Usage = flags.usage
	if err := flags.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return "", ExitOK
		}
		return "", ExitParseError.With(err)
	}

	verb := flags.verbose.get()
	if vers := flags.version.get(); vers > 0 {
		var b strings.Builder
		if vers > 1 || verb > 0 {
			fmt.Fprintf(&b, "github.com/ardnew/mung %s\n", mung.Version())
		}
		fmt.Fprintf(&b, "github.com/ardnew/mung/cmd/mung %s\n", flags.cmdVersion)
		return b.String(), ExitOK
	}

	if len(flags.Args()) == 0 {
		// let main print the error; usage will be printed by caller on demand
		flags.Usage()
		return "", ExitNoSubjects
	}

	subjects, err := flags.subjects()
	if err != nil {
		return "", ExitSubjectsError.With(err)
	}

	opts := []mung.Option[mung.Config]{
		mung.WithSubject(subjects),
		mung.WithDelim(flags.delim.get()),
	}

	if remove := flags.remove.get(); len(remove) > 0 {
		opts = append(opts, mung.WithRemove(remove))
	}
	if prefix := flags.prefix.get(); len(prefix) > 0 {
		opts = append(opts, mung.WithPrefix(prefix))
	}
	if suffix := flags.suffix.get(); len(suffix) > 0 {
		opts = append(opts, mung.WithSuffix(suffix))
	}
	if cmd := flags.filter.get(); cmd != "" {
		opts = append(opts, mung.WithFilter(flags.makeFilter(cmd)))
	}

	seq := mung.Make(opts...).Filtered()
	out := strings.Join(slices.Collect(seq), flags.delim.get())
	return out, ExitOK
}

type flagSet struct {
	*flag.FlagSet
	delim      soloValue
	remove     multiValue
	prefix     multiValue
	suffix     multiValue
	filter     soloValue
	nameref    bool
	verbose    incFlag
	version    incFlag
	cmdVersion string
}

func (f *flagSet) usage() {
	version := fmt.Sprintf(
		"%s version %s (module %s)", f.Name(), f.cmdVersion, mung.Version(),
	)
	fmt.Fprintln(f.Output(), version)
	fmt.Fprintln(f.Output())
	fmt.Fprintln(f.Output(), "USAGE")
	fmt.Fprintln(f.Output())
	fmt.Fprintf(f.Output(), "  %s [options] <subjects>\n", f.Name())
	fmt.Fprintln(f.Output())
	fmt.Fprintln(f.Output(), "OPTIONS")
	fmt.Fprintln(f.Output())
	f.PrintDefaults()
	fmt.Fprintln(f.Output())
	fmt.Fprintln(f.Output(), "SUBJECTS")
	fmt.Fprintln(f.Output())
	fmt.Fprintln(f.Output(), "  <subjects> are the initial strings to be munged.")
	fmt.Fprintln(f.Output(), "  These can be literal strings or, if flag -n is set,")
	fmt.Fprintln(f.Output(), "  environment variable names whose values will be used.")
	fmt.Fprintln(f.Output(), "  Subjects containing the delimiter (-d, default ':')")
	fmt.Fprintln(f.Output(), "  will be split into multiple items.")
	fmt.Fprintln(f.Output())
	fmt.Fprintln(f.Output(), "FILTERING")
	fmt.Fprintln(f.Output())
	fmt.Fprintln(f.Output(), "  The -t flag specifies a command-line to filter subjects.")
	fmt.Fprintln(f.Output(), "  The line is executed for each subject. A subject passes if")
	fmt.Fprintln(f.Output(), "  the command exits with status 0; otherwise, it is filtered out.")
	fmt.Fprintln(f.Output())
	fmt.Fprintln(f.Output(), "  Subject substitution:")
	fmt.Fprintln(f.Output(), "    - If '{}' appears in the command-line, it is replaced with")
	fmt.Fprintln(f.Output(), "      the subject (POSIX shell-quoted) before execution.")
	fmt.Fprintln(f.Output(), "    - If '{}' is not present, the subject is available as $1")
	fmt.Fprintln(f.Output(), "      or $@ to the shell (sh -c ...).")
	fmt.Fprintln(f.Output(), "    - If neither '{}' nor $1/$@ are present, the subject is")
	fmt.Fprintln(f.Output(), "      appended to the command-line as an argument.")
	fmt.Fprintln(f.Output())
	fmt.Fprintln(f.Output(), "  Examples:")
	fmt.Fprintln(f.Output(), "    -t '[ -d {} ]'  # subject substituted directly")
	fmt.Fprintln(f.Output(), "    -t '[ -d $1 ]'  # subject via $1 argument")
	fmt.Fprintln(f.Output(), "    -t 'test -d'    # subject appended to command")
}

func (f *flagSet) subjects() ([]string, error) {
	if f == nil || f.FlagSet == nil || !f.Parsed() {
		return nil, errors.New("flagSet is not initialized or parsed")
	}

	if !f.nameref {
		return f.Args(), nil
	}

	s := []string{}
	for _, name := range f.Args() {
		if value, ok := os.LookupEnv(name); ok {
			s = append(s, value)
		}
	}
	return s, nil
}

func (f *flagSet) makeFilter(cmd string) func(string) bool {
	return func(subject string) bool {
		c := strings.TrimSpace(cmd)
		if c == "" {
			return true
		}

		var command *exec.Cmd
		switch {
		case strings.Contains(c, "{}"):
			// Replace '{}' with a safely shell-quoted subject and execute.
			script := strings.ReplaceAll(c, "{}", shQuote(subject))
			command = exec.Command("sh", "-c", script)

		case strings.Contains(c, "$1") || strings.Contains(c, "$@"):
			// Subject is available as $1 (or $@) to the shell.
			command = exec.Command("sh", "-c", c, "sh", subject)

		default:
			// No subject substitution; pass subject as argument $1,
			// and append $1 to the command line.
			command = exec.Command("sh", "-c", c+" "+shQuote(subject), "sh", subject)
		}

		// If verbose, capture output and log details to stderr.
		verbose := f != nil && f.verbose.get() > 0
		var stdoutBuf, stderrBuf bytes.Buffer
		if verbose {
			command.Stdout = &stdoutBuf
			command.Stderr = &stderrBuf
			fmt.Fprintf(os.Stderr, "filter: %s\n", strings.Join(command.Args, " "))
		}

		err := command.Run()

		if verbose {
			// Determine exit status
			status := 0
			if err != nil {
				if ee, ok := err.(*exec.ExitError); ok {
					status = ee.ExitCode()
				} else {
					status = -1
				}
			}
			fmt.Fprintf(os.Stderr, "filter: status=%d\n", status)
			fmt.Fprintf(os.Stderr, "filter: stdout:\n%s", stdoutBuf.String())
			if !strings.HasSuffix(stdoutBuf.String(), "\n") {
				fmt.Fprintln(os.Stderr)
			}
			fmt.Fprintf(os.Stderr, "filter: stderr:\n%s", stderrBuf.String())
			if !strings.HasSuffix(stderrBuf.String(), "\n") {
				fmt.Fprintln(os.Stderr)
			}
		}

		if err != nil {
			return false
		}
		return true
	}
}

// shQuote returns a POSIX-shell-escaped version of s using single quotes.
// It is safe to paste into sh -c command strings.
func shQuote(s string) string {
	if s == "" {
		return "''"
	}
	// Close quote, insert escaped single quote, reopen: '"'"'
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

type (
	incFlag   int
	soloValue struct {
		solo string
		zero string
		name string
		desc string
	}
	multiValue struct {
		mult []string
		zero []string
		name string
		desc string
	}
)

func (f *incFlag) Set(value string) error {
	if f == nil {
		return errors.New("uninitialized flag")
	}
	if value = strings.TrimSpace(value); value != "" {
		if v, err := strconv.ParseBool(value); err == nil && v {
			*f += 1
		}
	}
	return nil
}

func (f *incFlag) get() int {
	if f != nil {
		return int(*f)
	}
	return 0
}

func (f *incFlag) IsBoolFlag() bool { return true }

func (f *incFlag) String() string {
	if f == nil {
		return ""
	}
	return strconv.Itoa(int(*f))
}

func (v *soloValue) Set(value string) error {
	if v == nil {
		return errors.New("uninitialized flag")
	}
	if strings.TrimSpace(value) != "" {
		v.solo = value
	}
	return nil
}

func (v *soloValue) String() string { return v.get() }

func (v *soloValue) isZero() bool {
	if v == nil {
		return true
	}
	return v.solo == ""
}

func (v *soloValue) get() string {
	if v == nil {
		return ""
	}
	if v.solo != "" {
		return v.solo
	}
	return v.zero
}

func (v *multiValue) Set(value string) error {
	if v == nil {
		return errors.New("uninitialized flag")
	}
	if v.mult == nil {
		v.mult = []string{}
	}
	if strings.TrimSpace(value) != "" {
		v.mult = append(v.mult, value)
	}
	return nil
}

func (v *multiValue) String() string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%+v", v.get())
}

func (v *multiValue) isZero() bool {
	if v == nil {
		return true
	}
	return len(v.mult) == 0
}

func (v *multiValue) get() []string {
	if v == nil {
		return nil
	}
	if len(v.mult) > 0 {
		return v.mult
	}
	return v.zero
}
