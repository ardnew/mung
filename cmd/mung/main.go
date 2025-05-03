package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/ardnew/mung"
)

type flagSet struct {
	*flag.FlagSet
	delim   soloValue
	remove  multiValue
	prefix  multiValue
	suffix  multiValue
	nameref bool
	verbose incFlag
	version incFlag
}

func main() {
	flags := flagSet{
		FlagSet: flag.NewFlagSet("mung", flag.ContinueOnError),
		delim:   soloValue{zero: ":", name: "d", desc: "item delimiter"},
		remove:  multiValue{name: "r", desc: "items to remove"},
		prefix:  multiValue{name: "a", desc: "items to prepend"},
		suffix:  multiValue{name: "e", desc: "items to append"},
		verbose: incFlag(0),
		version: incFlag(0),
	}
	// Define command-line flags
	flags.Var(&flags.delim, flags.delim.name, flags.delim.desc)
	flags.Var(&flags.remove, flags.remove.name, flags.remove.desc)
	flags.Var(&flags.prefix, flags.prefix.name, flags.prefix.desc)
	flags.Var(&flags.suffix, flags.suffix.name, flags.suffix.desc)
	flags.BoolVar(&flags.nameref, "n", false, "subjects are env NAME references")
	flags.Var(&flags.verbose, "v", "enable verbose output (incremental)")
	flags.Var(&flags.version, "V", "print semantic version of cmd (or module if verbose)")

	flags.Usage = flags.usage
	if err := flags.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	
	verb := flags.verbose.get()
	if vers := flags.version.get(); vers > 0 {
		if vers > 1 || verb > 0 {
			fmt.Printf("github.com/ardnew/mung %s\n", mung.Version)
		} else {
			fmt.Printf("github.com/ardnew/mung/cmd/mung %s\n", Version)
		}
		os.Exit(0)
	}

	if len(flags.Args()) == 0 {
		fmt.Fprint(os.Stderr, "error: no subjects provided\n")
		flags.Usage()
		os.Exit(2)
	}

	subjects, err := flags.subjects()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(3)
	}

	opts := []mung.Option[mung.Config]{
		mung.WithSubject(subjects...),
		mung.WithDelim(flags.delim.get()),
	}

	if remove := flags.remove.get(); len(remove) > 0 {
		opts = append(opts, mung.WithRemove(remove...))
	}
	if prefix := flags.prefix.get(); len(prefix) > 0 {
		opts = append(opts, mung.WithPrefix(prefix...))
	}
	if suffix := flags.suffix.get(); len(suffix) > 0 {
		opts = append(opts, mung.WithSuffix(suffix...))
	}

	seq := mung.Make(opts...).Seq()

	println(strings.Join(slices.Collect(seq), flags.delim.get()))
}

func (f *flagSet) usage() {
	fmt.Fprintf(f.Output(), "Usage: %s [options] <subjects>\n", f.Name())
	fmt.Fprintln(f.Output(), "\nOptions:")
	f.PrintDefaults()
	fmt.Fprintln(f.Output(), "\nSubjects:")
	fmt.Fprintln(f.Output(), "  <subjects> are the strings to process")
}

func (f *flagSet) subjects() ([]string, error) {
	if f == nil || f.FlagSet == nil || !f.Parsed() {
		return nil, errors.New("flagSet is not initialized or parsed")
	}
	s := []string{}
	if !f.nameref {
		return f.Args(), nil
	}
	for _, name := range f.Args() {
		if value, ok := os.LookupEnv(name); ok {
			s = append(s, value)
		}
	}
	return s, nil
}

type (
	incFlag int
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

func (v *soloValue) String() string {
	return v.get()
}

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
