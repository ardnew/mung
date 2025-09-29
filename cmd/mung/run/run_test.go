package run

import (
	"flag"
	"os"
	"strings"
	"testing"
)

func withArgs(args []string, fn func()) {
	old := os.Args
	os.Args = append([]string{"mung"}, args...)
	defer func() { os.Args = old }()
	fn()
}

func TestExitCode_Behavior(t *testing.T) {
	base := ExitCode{Code: 2, Msg: "oops"}
	if got := base.Int(); got != 2 {
		t.Fatalf("Int()=%d, want 2", got)
	}
	if got := base.Error(); !strings.Contains(got, "oops") {
		t.Fatalf("Error()=%q, want contains 'oops'", got)
	}
	wrapped := base.With(os.ErrNotExist)
	if w := wrapped.Unwrap(); w == nil || !strings.Contains(w.Error(), "file does not exist") {
		t.Fatalf("Unwrap()=%v, want os.ErrNotExist", w)
	}
}

func TestExitCode_ErrorVariants(t *testing.T) {
	// Msg empty, Err non-nil -> returns wrapped error string
	e1 := ExitCode{Code: 9, Err: os.ErrPermission}
	if got := e1.Error(); !strings.Contains(got, "permission") {
		t.Fatalf("Error()=%q, want contains 'permission'", got)
	}
	// Msg empty, Err nil -> returns formatted exit status
	e2 := ExitCode{Code: 5}
	if got := e2.Error(); got != "exit status 5" {
		t.Fatalf("Error()=%q, want 'exit status 5'", got)
	}
}

func TestNilReceiversAndSetErrors(t *testing.T) {
	// incFlag nil receiver behaviors
	var pf *incFlag
	if pf.get() != 0 {
		t.Fatalf("nil incFlag get != 0")
	}
	if pf.String() != "" {
		t.Fatalf("nil incFlag String should be empty")
	}
	if !pf.IsBoolFlag() {
		t.Fatalf("nil incFlag IsBoolFlag should be true")
	}
	if err := pf.Set("true"); err == nil {
		t.Fatalf("nil incFlag.Set should error")
	}

	// soloValue nil receiver behaviors
	var sv *soloValue
	if !sv.isZero() {
		t.Fatalf("nil soloValue isZero should be true")
	}
	if sv.get() != "" {
		t.Fatalf("nil soloValue get should be empty string")
	}
	if err := sv.Set("x"); err == nil {
		t.Fatalf("nil soloValue.Set should error")
	}

	// multiValue nil receiver behaviors
	var mv *multiValue
	if !mv.isZero() {
		t.Fatalf("nil multiValue isZero should be true")
	}
	if mv.get() != nil {
		t.Fatalf("nil multiValue get should be nil slice")
	}
	if mv.String() != "" {
		t.Fatalf("nil multiValue String should be empty")
	}
	if err := mv.Set("x"); err == nil {
		t.Fatalf("nil multiValue.Set should error")
	}

	// flagSet.subjects error when FlagSet is nil
	var fs flagSet
	if _, err := fs.subjects(); err == nil {
		t.Fatalf("subjects should error when FlagSet is nil")
	}
}

func TestMain_Version_basic(t *testing.T) {
	withArgs([]string{"-V"}, func() {
		out, code := Main("1.2.3")
		if code.Int() != 0 {
			t.Fatalf("code=%d, want 0", code.Int())
		}
		if !strings.Contains(out, "github.com/ardnew/mung/cmd/mung 1.2.3") {
			t.Fatalf("out=%q missing cmd version line", out)
		}
		if strings.Contains(out, "github.com/ardnew/mung ") {
			t.Fatalf("out should not include module version with single -V: %q", out)
		}
	})
}

func TestMain_Version_verbose(t *testing.T) {
	withArgs([]string{"-V", "-v"}, func() {
		out, code := Main("1.2.3")
		if code.Int() != 0 {
			t.Fatalf("code=%d, want 0", code.Int())
		}
		if !strings.Contains(out, "github.com/ardnew/mung ") {
			t.Fatalf("out should include module version with -v: %q", out)
		}
		if !strings.Contains(out, "github.com/ardnew/mung/cmd/mung 1.2.3") {
			t.Fatalf("out=%q missing cmd version line", out)
		}
	})
}

func TestMain_NoSubjects_ShowsUsageAndReturnsCode(t *testing.T) {
	withArgs(nil, func() {
		out, code := Main("0.0.0")
		if out != "" {
			t.Fatalf("out=%q, want empty", out)
		}
		if code.Int() != ExitNoSubjects.Int() {
			t.Fatalf("code=%d, want %d", code.Int(), ExitNoSubjects.Int())
		}
	})
}

func TestMain_ParseError(t *testing.T) {
	withArgs([]string{"-Z"}, func() {
		_, code := Main("0.0.0")
		if code.Int() != ExitParseError.Int() {
			t.Fatalf("code=%d, want %d", code.Int(), ExitParseError.Int())
		}
		if code.Error() == "" {
			t.Fatalf("expected non-empty error message")
		}
	})
}

func TestMain_SubjectsAndDelim(t *testing.T) {
	withArgs([]string{"-d", ":", "a::b"}, func() {
		out, code := Main("0")
		if code.Int() != 0 {
			t.Fatalf("code=%d, want 0", code.Int())
		}
		if out != "a:b" {
			t.Fatalf("out=%q, want 'a:b'", out)
		}
	})
}

func TestMain_AllOptions_RemovePrefixSuffix(t *testing.T) {
	withArgs([]string{"-d", ":", "-r", "/bin", "-p", "/usr/local/bin", "-s", "/opt/bin", "/usr/bin:/bin"}, func() {
		out, code := Main("0")
		if code.Int() != 0 {
			t.Fatalf("code=%d, want 0", code.Int())
		}
		if out != "/usr/local/bin:/usr/bin:/opt/bin" {
			t.Fatalf("out=%q, want '/usr/local/bin:/usr/bin:/opt/bin'", out)
		}
	})
}

func TestMain_NameRefEnv(t *testing.T) {
	t.Setenv("MUNG_TEST_PATH", "/bin:/sbin")
	withArgs([]string{"-n", "-d", ":", "MUNG_TEST_PATH"}, func() {
		out, code := Main("0")
		if code.Int() != 0 {
			t.Fatalf("code=%d, want 0", code.Int())
		}
		if out != "/bin:/sbin" {
			t.Fatalf("out=%q, want '/bin:/sbin'", out)
		}
	})
}

func TestMain_NameRefEnv_MissingIsSkipped(t *testing.T) {
	t.Setenv("MUNG_TEST_SET", "/bin")
	withArgs([]string{"-n", "-d", ":", "MUNG_TEST_SET", "MUNG_TEST_MISSING"}, func() {
		out, code := Main("0")
		if code.Int() != 0 {
			t.Fatalf("code=%d, want 0", code.Int())
		}
		if out != "/bin" {
			t.Fatalf("out=%q, want '/bin'", out)
		}
	})
}

func TestMakeFilter_WithBraces(t *testing.T) {
	f := (&flagSet{}).makeFilter("[ -n {} ]")
	if !f("abc") {
		t.Fatalf("filter returned false for non-empty subject")
	}
	if f("") {
		t.Fatalf("filter returned true for empty subject")
	}
}

func TestMakeFilter_WithDollarArg(t *testing.T) {
	f := (&flagSet{}).makeFilter("[ -n \"$1\" ]")
	if !f("abc") {
		t.Fatalf("$1 filter returned false for non-empty subject")
	}
	if f("") {
		t.Fatalf("$1 filter returned true for empty subject")
	}
}

func TestMakeFilter_DefaultAppend(t *testing.T) {
	f := (&flagSet{}).makeFilter("test -n")
	if !f("abc") {
		t.Fatalf("default-append filter returned false for non-empty subject")
	}
	if f("") {
		t.Fatalf("default-append filter returned true for empty subject")
	}
}

func TestMakeFilter_EmptyReturnsTrue(t *testing.T) {
	f := (&flagSet{}).makeFilter("   ")
	if !f("") || !f("abc") {
		t.Fatalf("empty command filter should always return true")
	}
}

func Test_shQuote(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"", "''"},
		{"abc", "'abc'"},
		{"a'b", "'a'\"'\"'b'"},
	}
	for _, tt := range tests {
		if got := shQuote(tt.in); got != tt.out {
			t.Fatalf("shQuote(%q)=%q, want %q", tt.in, got, tt.out)
		}
	}
}

func TestMain_FilterIntegration_Braces(t *testing.T) {
	withArgs([]string{"-t", "[ -n {} ]", "-d", ":", "a::b"}, func() {
		out, code := Main("0")
		if code.Int() != 0 {
			t.Fatalf("code=%d, want 0", code.Int())
		}
		if out != "a:b" {
			t.Fatalf("out=%q, want 'a:b'", out)
		}
	})
}

func TestMain_FilterIntegration_Dollar(t *testing.T) {
	withArgs([]string{"-t", "[ -n \"$1\" ]", "-d", ":", "a::b"}, func() {
		out, code := Main("0")
		if code.Int() != 0 {
			t.Fatalf("code=%d, want 0", code.Int())
		}
		if out != "a:b" {
			t.Fatalf("out=%q, want 'a:b'", out)
		}
	})
}

func TestMain_FilterIntegration_DefaultAppend(t *testing.T) {
	withArgs([]string{"-t", "test -n", "-d", ":", "::a:b::"}, func() {
		out, code := Main("0")
		if code.Int() != 0 {
			t.Fatalf("code=%d, want 0", code.Int())
		}
		if out != "a:b" {
			t.Fatalf("out=%q, want 'a:b'", out)
		}
	})
}

func TestExitCode_ZeroHasEmptyError(t *testing.T) {
	if ExitOK.Error() != "" {
		t.Fatalf("ExitOK.Error()=%q, want empty", ExitOK.Error())
	}
}

func TestIncFlag(t *testing.T) {
	var f incFlag
	if f.get() != 0 || f.String() != "0" {
		t.Fatalf("initial incFlag unexpected: get=%d str=%q", f.get(), f.String())
	}
	if err := f.Set(""); err != nil { // empty is no-op
		t.Fatalf("Set empty err=%v", err)
	}
	if f.get() != 0 {
		t.Fatalf("after empty Set, get=%d; want 0", f.get())
	}
	if err := f.Set("true"); err != nil {
		t.Fatalf("Set true err=%v", err)
	}
	if f.get() != 1 {
		t.Fatalf("after true, get=%d; want 1", f.get())
	}
	if err := f.Set("junk"); err != nil { // non-bool ignored
		t.Fatalf("Set junk err=%v", err)
	}
	if f.get() != 1 {
		t.Fatalf("after junk, get=%d; want 1", f.get())
	}
	if !f.IsBoolFlag() {
		t.Fatalf("IsBoolFlag=false, want true")
	}
}

func TestSoloValue(t *testing.T) {
	v := soloValue{zero: ":"}
	if !v.isZero() {
		t.Fatalf("new soloValue should be zero")
	}
	if v.get() != ":" {
		t.Fatalf("get zero=%q, want ':'", v.get())
	}
	if err := v.Set(";"); err != nil {
		t.Fatalf("Set err=%v", err)
	}
	if v.isZero() {
		t.Fatalf("soloValue should not be zero after Set")
	}
	if v.get() != ";" {
		t.Fatalf("get after Set=%q, want ';'", v.get())
	}
}

func TestMultiValue(t *testing.T) {
	v := multiValue{zero: []string{"Z"}}
	if !v.isZero() {
		t.Fatalf("new multiValue should be zero")
	}
	if got := v.get(); len(got) != 1 || got[0] != "Z" {
		t.Fatalf("get zero=%v", got)
	}
	if err := v.Set("a"); err != nil {
		t.Fatalf("Set a err=%v", err)
	}
	if err := v.Set("b"); err != nil {
		t.Fatalf("Set b err=%v", err)
	}
	if v.isZero() {
		t.Fatalf("multiValue should not be zero after Set")
	}
	got := v.get()
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("get=%v", got)
	}
	if s := v.String(); !strings.Contains(s, "a") || !strings.Contains(s, "b") {
		t.Fatalf("String()=%q, want contains 'a' and 'b'", s)
	}
}

func TestHelpFlag(t *testing.T) {
	withArgs([]string{"-h"}, func() {
		out, code := Main("0")
		if code.Int() != 0 {
			t.Fatalf("code=%d, want 0", code.Int())
		}
		if out != "" {
			t.Fatalf("out=%q, want empty on help", out)
		}
	})
}

func TestDoubleV_ShowsModuleWithoutVerbose(t *testing.T) {
	withArgs([]string{"-V", "-V"}, func() {
		out, code := Main("1.0.0")
		if code.Int() != 0 {
			t.Fatalf("code=%d, want 0", code.Int())
		}
		if !strings.Contains(out, "github.com/ardnew/mung ") {
			t.Fatalf("out should include module version with -V -V: %q", out)
		}
	})
}

func TestUsagePrinter(t *testing.T) {
	// Indirectly exercised via NoSubjects, but call directly to increase coverage
	fs := flagSet{FlagSet: nil, cmdVersion: "X"}
	// Initialize the flag set properly
	fs.FlagSet = newFlagSetForTest()
	fs.cmdVersion = "X"
	fs.delim = soloValue{zero: ":", name: "d", desc: "item delimiter"}
	fs.remove = multiValue{name: "r", desc: "items to remove"}
	fs.prefix = multiValue{name: "p", desc: "items to prefix subject(s)"}
	fs.suffix = multiValue{name: "s", desc: "items to suffix subject(s)"}
	fs.filter = soloValue{name: "t", desc: "command to filter subject(s)"}
	fs.verbose = incFlag(0)
	fs.version = incFlag(0)
	fs.Var(&fs.delim, fs.delim.name, fs.delim.desc)
	fs.Var(&fs.remove, fs.remove.name, fs.remove.desc)
	fs.Var(&fs.prefix, fs.prefix.name, fs.prefix.desc)
	fs.Var(&fs.suffix, fs.suffix.name, fs.suffix.desc)
	fs.Var(&fs.filter, fs.filter.name, fs.filter.desc)
	fs.Var(&fs.verbose, "v", "enable verbose")
	fs.Var(&fs.version, "V", "print version")
	buf := &strings.Builder{}
	fs.SetOutput(buf)
	fs.usage()
	out := buf.String()
	for _, need := range []string{"USAGE", "OPTIONS", "FILTERING", "-t", "-d"} {
		if !strings.Contains(out, need) {
			t.Fatalf("usage missing %q in output: %q", need, out)
		}
	}
}

func newFlagSetForTest() *flag.FlagSet {
	fs := flag.NewFlagSet("mung", flag.ContinueOnError)
	fs.SetOutput(new(strings.Builder))
	return fs
}

func TestFlagSetSubjects_NotParsed(t *testing.T) {
	fs := flagSet{FlagSet: newFlagSetForTest()}
	fs.nameref = false
	if _, err := fs.subjects(); err == nil {
		t.Fatalf("expected error when subjects called before Parse")
	}
}

func TestFlagSetSubjects_NameRefOnlySetOnes(t *testing.T) {
	fs := flagSet{FlagSet: newFlagSetForTest()}
	fs.nameref = true
	// Simulate parsed args
	_ = fs.FlagSet.Parse([]string{"FOO", "BAR"})
	t.Setenv("FOO", "X")
	got, err := fs.subjects()
	if err != nil {
		t.Fatalf("subjects err=%v", err)
	}
	if len(got) != 1 || got[0] != "X" {
		t.Fatalf("subjects got=%v, want [X]", got)
	}
}
