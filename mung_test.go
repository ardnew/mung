package mung

import (
	"reflect"
	"slices"
	"testing"
)

// TestMake tests the Make generic function with various option combinations
func TestMake(t *testing.T) {
	tests := []struct {
		name string
		opts []Option[Config]
		want Config
	}{
		{
			name: "empty_options",
			opts: []Option[Config]{},
			want: Config{},
		},
		{
			name: "single_option",
			opts: []Option[Config]{WithSubject("foo")},
			want: Config{subject: []string{"foo"}},
		},
		{
			name: "multiple_options",
			opts: []Option[Config]{
				WithSubject("foo", "bar"),
				WithDelim(":"),
				WithRemove("baz"),
				WithPrefix("pre"),
				WithSuffix("suf"),
				WithReplace(map[string]string{"foo": "replaced"}),
			},
			want: Config{
				subject: []string{"foo", "bar"},
				delim:   ":",
				remove:  []string{"baz"},
				prefix:  []string{"pre"},
				suffix:  []string{"suf"},
				replace: map[string]string{"foo": "replaced"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Make(tt.opts...)
			if !configEqual(got, tt.want) {
				t.Errorf("Make() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test with a different type
	type testStruct struct {
		Value string
	}

	setValueOption := func(v string) Option[testStruct] {
		return func(ts testStruct) testStruct {
			ts.Value = v
			return ts
		}
	}

	ts := Make(setValueOption("test value"))
	if ts.Value != "test value" {
		t.Errorf("Make() with custom type failed, got: %v, want: %v", ts.Value, "test value")
	}
}

// TestWrap tests applying options to an existing object
func TestWrap(t *testing.T) {
	tests := []struct {
		name   string
		init   Config
		opts   []Option[Config]
		expect Config
	}{
		{
			name:   "empty_config_no_options",
			init:   Config{},
			opts:   []Option[Config]{},
			expect: Config{},
		},
		{
			name: "existing_config_new_options",
			init: Config{delim: ",", subject: []string{"original"}},
			opts: []Option[Config]{
				WithSubject("new"),
				WithPrefix("pre"),
			},
			expect: Config{
				delim:   ",",
				subject: []string{"new"},
				prefix:  []string{"pre"},
			},
		},
		{
			name: "multiple_options_chain",
			init: Config{},
			opts: []Option[Config]{
				WithSubject("a", "b"),
				WithDelim(":"),
				WithRemove("x"),
				WithPrefix("pre"),
				WithSuffix("suf"),
				WithReplace(map[string]string{"a": "A"}),
			},
			expect: Config{
				subject: []string{"a", "b"},
				delim:   ":",
				remove:  []string{"x"},
				prefix:  []string{"pre"},
				suffix:  []string{"suf"},
				replace: map[string]string{"a": "A"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrap(tt.init, tt.opts...)
			if !configEqual(got, tt.expect) {
				t.Errorf("Wrap() = %v, want %v", got, tt.expect)
			}
		})
	}
}

// TestConfigString tests the String method of Config
func TestConfigString(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   string
	}{
		{
			name:   "empty_config",
			config: Config{},
			want:   "",
		},
		{
			name: "simple_subject",
			config: Config{
				subject: []string{"hello"},
				delim:   ",",
			},
			want: "hello",
		},
		{
			name: "multiple_subject_items",
			config: Config{
				subject: []string{"hello", "world"},
				delim:   ",",
			},
			want: "hello,world",
		},
		{
			name: "subject_with_delimiters",
			config: Config{
				subject: []string{"hello,there", "world"},
				delim:   ",",
			},
			want: "hello,there,world",
		},
		{
			name: "with_prefix_suffix",
			config: Config{
				subject: []string{"middle"},
				delim:   ",",
				prefix:  []string{"start"},
				suffix:  []string{"end"},
			},
			want: "start,middle,end",
		},
		{
			name: "with_removals",
			config: Config{
				subject: []string{"a", "b", "c", "d"},
				delim:   ",",
				remove:  []string{"b", "d"},
			},
			want: "a,c",
		},
		{
			name: "with_replacements",
			config: Config{
				subject: []string{"hello", "world"},
				delim:   ",",
				replace: map[string]string{"hello": "hi", "world": "earth"},
			},
			want: "hi,earth",
		},
		{
			name: "empty_delimiter",
			config: Config{
				subject: []string{"hello", "world"},
				delim:   "",
			},
			want: "helloworld",
		},
		{
			name: "complex_case",
			config: Config{
				subject: []string{"b", "c", "d"},
				delim:   ":",
				prefix:  []string{"a"},
				suffix:  []string{"e", "f"},
				remove:  []string{"c"},
				replace: map[string]string{"d": "D"},
			},
			want: "a:b:D:e:f",
		},
		{
			name: "duplicate_elimination",
			config: Config{
				subject: []string{"a", "b", "a", "c"},
				delim:   ",",
			},
			want: "a,b,c",
		},
		{
			name: "prefix_overlap_with_subject",
			config: Config{
				subject: []string{"a", "b"},
				delim:   ",",
				prefix:  []string{"a", "c"},
			},
			want: "a,c,b",
		},
		{
			name: "delimiter_character_in_items",
			config: Config{
				subject: []string{"a,b", "c,d"},
				delim:   ",",
			},
			want: "a,b,c,d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.String(); got != tt.want {
				t.Errorf("Config.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestConfigAccessors tests all getter methods of Config
func TestConfigAccessors(t *testing.T) {
	config := Config{
		subject: []string{"a", "b"},
		delim:   ":",
		remove:  []string{"x", "y"},
		prefix:  []string{"p1", "p2"},
		suffix:  []string{"s1", "s2"},
		replace: map[string]string{"a": "A", "b": "B"},
	}

	// Test Subject accessor
	if !slicesEqual(config.Subject(), []string{"a", "b"}) {
		t.Errorf("Config.Subject() = %v, want %v", config.Subject(), []string{"a", "b"})
	}

	// Test Delim accessor
	if config.Delim() != ":" {
		t.Errorf("Config.Delim() = %v, want %v", config.Delim(), ":")
	}

	// Test Remove accessor
	if !slicesEqual(config.Remove(), []string{"x", "y"}) {
		t.Errorf("Config.Remove() = %v, want %v", config.Remove(), []string{"x", "y"})
	}

	// Test Prefix accessor
	if !slicesEqual(config.Prefix(), []string{"p1", "p2"}) {
		t.Errorf("Config.Prefix() = %v, want %v", config.Prefix(), []string{"p1", "p2"})
	}

	// Test Suffix accessor
	if !slicesEqual(config.Suffix(), []string{"s1", "s2"}) {
		t.Errorf("Config.Suffix() = %v, want %v", config.Suffix(), []string{"s1", "s2"})
	}

	// Test Replace accessor makes a copy
	replaceMap := config.Replace()
	if !mapsEqual(replaceMap, map[string]string{"a": "A", "b": "B"}) {
		t.Errorf("Config.Replace() = %v, want %v", replaceMap, map[string]string{"a": "A", "b": "B"})
	}

	// Verify that modifying the returned map doesn't affect the original
	replaceMap["a"] = "Modified"
	if config.replace["a"] != "A" {
		t.Errorf("Config.Replace() did not return a copy; original map was modified")
	}
}

// TestConfigSeq tests the Seq method of Config under various configurations
func TestConfigSeq(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   []string
	}{
		{
			name:   "empty_config",
			config: Config{},
			want:   []string{},
		},
		{
			name: "simple_subject",
			config: Config{
				subject: []string{"hello"},
			},
			want: []string{"hello"},
		},
		{
			name: "subject_with_delimiter",
			config: Config{
				subject: []string{"hello,world"},
				delim:   ",",
			},
			want: []string{"hello", "world"},
		},
		{
			name: "prefix_and_suffix",
			config: Config{
				subject: []string{"middle"},
				prefix:  []string{"start"},
				suffix:  []string{"end"},
			},
			want: []string{"start", "middle", "end"},
		},
		{
			name: "with_removals",
			config: Config{
				subject: []string{"a", "b", "c"},
				remove:  []string{"b"},
			},
			want: []string{"a", "c"},
		},
		{
			name: "with_replacements",
			config: Config{
				subject: []string{"hello", "world"},
				replace: map[string]string{"hello": "hi"},
			},
			want: []string{"hi", "world"},
		},
		{
			name: "complex_config",
			config: Config{
				subject: []string{"item1,item2", "item3"},
				delim:   ",",
				prefix:  []string{"prefix"},
				suffix:  []string{"suffix"},
				remove:  []string{"item2"},
				replace: map[string]string{"item3": "replaced"},
			},
			want: []string{"prefix", "item1", "replaced", "suffix"},
		},
		{
			name: "remove_prefix_items",
			config: Config{
				subject: []string{"a", "b", "c"},
				prefix:  []string{"x", "y"},
				remove:  []string{"x", "a"},
			},
			want: []string{"y", "b", "c"},
		},
		{
			name: "remove_suffix_items",
			config: Config{
				subject: []string{"a", "b"},
				suffix:  []string{"c", "d"},
				remove:  []string{"b", "d"},
			},
			want: []string{"a", "c"},
		},
		{
			name: "eliminate_duplicates",
			config: Config{
				subject: []string{"a", "b", "a", "c", "b"},
				prefix:  []string{"a", "d"},
			},
			want: []string{"a", "d", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.Collect(tt.config.Seq())
			if !slicesEqual(got, tt.want) {
				t.Errorf("Config.Seq() collected = %v, want %v", got, tt.want)
			}

			// Test early termination by stopping after first item
			earlyTermResults := []string{}
			tt.config.Seq()(func(s string) bool {
				earlyTermResults = append(earlyTermResults, s)
				return len(earlyTermResults) < 1 // only process one item
			})

			if len(tt.want) > 0 && (len(earlyTermResults) != 1 || earlyTermResults[0] != tt.want[0]) {
				t.Errorf("Config.Seq() early termination = %v, want [%v]", earlyTermResults, tt.want[0])
			}
		})
	}
}

// TestWithSubject tests the WithSubject option function
func TestWithSubject(t *testing.T) {
	tests := []struct {
		name     string
		subjects []string
		initial  Config
		want     []string
	}{
		{
			name:     "empty_to_empty",
			subjects: []string{},
			initial:  Config{},
			want:     []string{},
		},
		{
			name:     "empty_to_single",
			subjects: []string{"hello"},
			initial:  Config{},
			want:     []string{"hello"},
		},
		{
			name:     "empty_to_multiple",
			subjects: []string{"hello", "world"},
			initial:  Config{},
			want:     []string{"hello", "world"},
		},
		{
			name:     "replace_existing",
			subjects: []string{"new1", "new2"},
			initial:  Config{subject: []string{"old1", "old2"}},
			want:     []string{"new1", "new2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WithSubject(tt.subjects...)(tt.initial)
			if !slicesEqual(config.Subject(), tt.want) {
				t.Errorf("WithSubject() = %v, want %v", config.Subject(), tt.want)
			}
		})
	}
}

// TestWithSubjectItems tests the WithSubjectItems option function
func TestWithSubjectItems(t *testing.T) {
	tests := []struct {
		name     string
		subjects []string
		initial  Config
		want     []string
	}{
		{
			name:     "nil_to_empty",
			subjects: []string{},
			initial:  Config{},
			want:     []string{},
		},
		{
			name:     "nil_to_single",
			subjects: []string{"hello"},
			initial:  Config{},
			want:     []string{"hello"},
		},
		{
			name:     "nil_to_multiple",
			subjects: []string{"hello", "world"},
			initial:  Config{},
			want:     []string{"hello", "world"},
		},
		{
			name:     "append_to_existing",
			subjects: []string{"new1", "new2"},
			initial:  Config{subject: []string{"old1", "old2"}},
			want:     []string{"old1", "old2", "new1", "new2"},
		},
		{
			name:     "append_to_non-nil_empty",
			subjects: []string{"new1"},
			initial:  Config{subject: []string{}},
			want:     []string{"new1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WithSubjectItems(tt.subjects...)(tt.initial)
			if !slicesEqual(config.Subject(), tt.want) {
				t.Errorf("WithSubjectItems() = %v, want %v", config.Subject(), tt.want)
			}
		})
	}
}

// TestWithDelim tests the WithDelim option function
func TestWithDelim(t *testing.T) {
	tests := []struct {
		name    string
		delim   string
		initial Config
		want    string
	}{
		{
			name:    "empty_to_empty",
			delim:   "",
			initial: Config{},
			want:    "",
		},
		{
			name:    "empty_to_comma",
			delim:   ",",
			initial: Config{},
			want:    ",",
		},
		{
			name:    "empty_to_special_chars",
			delim:   ":|:",
			initial: Config{},
			want:    ":|:",
		},
		{
			name:    "replace_existing",
			delim:   "new",
			initial: Config{delim: "old"},
			want:    "new",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WithDelim(tt.delim)(tt.initial)
			if config.Delim() != tt.want {
				t.Errorf("WithDelim() = %v, want %v", config.Delim(), tt.want)
			}
		})
	}
}

// TestWithRemove tests the WithRemove option function
func TestWithRemove(t *testing.T) {
	tests := []struct {
		name    string
		remove  []string
		initial Config
		want    []string
	}{
		{
			name:    "empty_to_empty",
			remove:  []string{},
			initial: Config{},
			want:    []string{},
		},
		{
			name:    "empty_to_single",
			remove:  []string{"hello"},
			initial: Config{},
			want:    []string{"hello"},
		},
		{
			name:    "empty_to_multiple",
			remove:  []string{"hello", "world"},
			initial: Config{},
			want:    []string{"hello", "world"},
		},
		{
			name:    "replace_existing",
			remove:  []string{"new1", "new2"},
			initial: Config{remove: []string{"old1", "old2"}},
			want:    []string{"new1", "new2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WithRemove(tt.remove...)(tt.initial)
			if !slicesEqual(config.Remove(), tt.want) {
				t.Errorf("WithRemove() = %v, want %v", config.Remove(), tt.want)
			}
		})
	}
}

// TestWithRemoveItems tests the WithRemoveItems option function
func TestWithRemoveItems(t *testing.T) {
	tests := []struct {
		name    string
		remove  []string
		initial Config
		want    []string
	}{
		{
			name:    "nil_to_empty",
			remove:  []string{},
			initial: Config{},
			want:    []string{},
		},
		{
			name:    "nil_to_single",
			remove:  []string{"hello"},
			initial: Config{},
			want:    []string{"hello"},
		},
		{
			name:    "nil_to_multiple",
			remove:  []string{"hello", "world"},
			initial: Config{},
			want:    []string{"hello", "world"},
		},
		{
			name:    "append_to_existing",
			remove:  []string{"new1", "new2"},
			initial: Config{remove: []string{"old1", "old2"}},
			want:    []string{"old1", "old2", "new1", "new2"},
		},
		{
			name:    "append_to_non-nil_empty",
			remove:  []string{"new1"},
			initial: Config{remove: []string{}},
			want:    []string{"new1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WithRemoveItems(tt.remove...)(tt.initial)
			if !slicesEqual(config.Remove(), tt.want) {
				t.Errorf("WithRemoveItems() = %v, want %v", config.Remove(), tt.want)
			}
		})
	}
}

// TestWithPrefix tests the WithPrefix option function
func TestWithPrefix(t *testing.T) {
	tests := []struct {
		name    string
		prefix  []string
		initial Config
		want    []string
	}{
		{
			name:    "empty_to_empty",
			prefix:  []string{},
			initial: Config{},
			want:    []string{},
		},
		{
			name:    "empty_to_single",
			prefix:  []string{"hello"},
			initial: Config{},
			want:    []string{"hello"},
		},
		{
			name:    "empty_to_multiple",
			prefix:  []string{"world", "hello"},
			initial: Config{},
			want:    []string{"hello", "world"},
		},
		{
			name:    "replace_existing",
			prefix:  []string{"new2", "new1"},
			initial: Config{prefix: []string{"old1", "old2"}},
			want:    []string{"new1", "new2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WithPrefix(tt.prefix...)(tt.initial)
			if !slicesEqual(config.Prefix(), tt.want) {
				t.Errorf("WithPrefix() = %v, want %v", config.Prefix(), tt.want)
			}
		})
	}
}

// TestWithPrefixItems tests the WithPrefixItems option function
func TestWithPrefixItems(t *testing.T) {
	tests := []struct {
		name    string
		prefix  []string
		initial Config
		want    []string
	}{
		{
			name:    "nil_to_empty",
			prefix:  []string{},
			initial: Config{},
			want:    []string{},
		},
		{
			name:    "nil_to_single",
			prefix:  []string{"hello"},
			initial: Config{},
			want:    []string{"hello"},
		},
		{
			name:    "nil_to_multiple",
			prefix:  []string{"world", "hello"},
			initial: Config{},
			want:    []string{"hello", "world"},
		},
		{
			name:    "append_to_existing",
			prefix:  []string{"new2", "new1"},
			initial: Config{prefix: []string{"old1", "old2"}},
			want:    []string{"new1", "new2", "old1", "old2"},
		},
		{
			name:    "append_to_non-nil_empty",
			prefix:  []string{"new1"},
			initial: Config{prefix: []string{}},
			want:    []string{"new1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WithPrefixItems(tt.prefix...)(tt.initial)
			if !slicesEqual(config.Prefix(), tt.want) {
				t.Errorf("WithPrefixItems() = %v, want %v", config.Prefix(), tt.want)
			}
		})
	}
}

// TestWithSuffix tests the WithSuffix option function
func TestWithSuffix(t *testing.T) {
	tests := []struct {
		name    string
		suffix  []string
		initial Config
		want    []string
	}{
		{
			name:    "empty_to_empty",
			suffix:  []string{},
			initial: Config{},
			want:    []string{},
		},
		{
			name:    "empty_to_single",
			suffix:  []string{"hello"},
			initial: Config{},
			want:    []string{"hello"},
		},
		{
			name:    "empty_to_multiple",
			suffix:  []string{"hello", "world"},
			initial: Config{},
			want:    []string{"hello", "world"},
		},
		{
			name:    "replace_existing",
			suffix:  []string{"new1", "new2"},
			initial: Config{suffix: []string{"old1", "old2"}},
			want:    []string{"new1", "new2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WithSuffix(tt.suffix...)(tt.initial)
			if !slicesEqual(config.Suffix(), tt.want) {
				t.Errorf("WithSuffix() = %v, want %v", config.Suffix(), tt.want)
			}
		})
	}
}

// TestWithSuffixItems tests the WithSuffixItems option function
func TestWithSuffixItems(t *testing.T) {
	tests := []struct {
		name    string
		suffix  []string
		initial Config
		want    []string
	}{
		{
			name:    "nil_to_empty",
			suffix:  []string{},
			initial: Config{},
			want:    []string{},
		},
		{
			name:    "nil_to_single",
			suffix:  []string{"hello"},
			initial: Config{},
			want:    []string{"hello"},
		},
		{
			name:    "nil_to_multiple",
			suffix:  []string{"hello", "world"},
			initial: Config{},
			want:    []string{"hello", "world"},
		},
		{
			name:    "append_to_existing",
			suffix:  []string{"new1", "new2"},
			initial: Config{suffix: []string{"old1", "old2"}},
			want:    []string{"old1", "old2", "new1", "new2"},
		},
		{
			name:    "append_to_non-nil_empty",
			suffix:  []string{"new1"},
			initial: Config{suffix: []string{}},
			want:    []string{"new1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WithSuffixItems(tt.suffix...)(tt.initial)
			if !slicesEqual(config.Suffix(), tt.want) {
				t.Errorf("WithSuffixItems() = %v, want %v", config.Suffix(), tt.want)
			}
		})
	}
}

// TestWithReplace tests the WithReplace option function
func TestWithReplace(t *testing.T) {
	tests := []struct {
		name    string
		replace map[string]string
		initial Config
		want    map[string]string
	}{
		{
			name:    "empty_to_empty",
			replace: map[string]string{},
			initial: Config{},
			want:    map[string]string{},
		},
		{
			name:    "empty_to_single",
			replace: map[string]string{"hello": "world"},
			initial: Config{},
			want:    map[string]string{"hello": "world"},
		},
		{
			name:    "empty_to_multiple",
			replace: map[string]string{"hello": "world", "foo": "bar"},
			initial: Config{},
			want:    map[string]string{"hello": "world", "foo": "bar"},
		},
		{
			name:    "replace_existing",
			replace: map[string]string{"new1": "val1", "new2": "val2"},
			initial: Config{replace: map[string]string{"old1": "val1", "old2": "val2"}},
			want:    map[string]string{"new1": "val1", "new2": "val2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := WithReplace(tt.replace)(tt.initial)
			got := config.Replace()
			if !mapsEqual(got, tt.want) {
				t.Errorf("WithReplace() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWithReplaceItems tests the WithReplaceItems option function
func TestWithReplaceItems(t *testing.T) {
	tests := []struct {
		name         string
		replacePairs [][2]string
		initial      Config
		want         map[string]string
	}{
		{
			name:         "nil_to_empty",
			replacePairs: [][2]string{},
			initial:      Config{},
			want:         map[string]string{},
		},
		{
			name:         "nil_to_single",
			replacePairs: [][2]string{{"hello", "world"}},
			initial:      Config{},
			want:         map[string]string{"hello": "world"},
		},
		{
			name:         "nil_to_multiple",
			replacePairs: [][2]string{{"hello", "world"}, {"foo", "bar"}},
			initial:      Config{},
			want:         map[string]string{"hello": "world", "foo": "bar"},
		},
		{
			name:         "merge_with_existing",
			replacePairs: [][2]string{{"new1", "val1"}, {"new2", "val2"}},
			initial:      Config{replace: map[string]string{"old1": "val1", "old2": "val2"}},
			want:         map[string]string{"old1": "val1", "old2": "val2", "new1": "val1", "new2": "val2"},
		},
		{
			name:         "override_existing_key",
			replacePairs: [][2]string{{"key", "newval"}},
			initial:      Config{replace: map[string]string{"key": "oldval"}},
			want:         map[string]string{"key": "newval"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert pairs to iter.Seq2
			seq := func(yield func(string, string) bool) {
				for _, pair := range tt.replacePairs {
					if !yield(pair[0], pair[1]) {
						return
					}
				}
			}

			config := WithReplaceItems(seq)(tt.initial)
			got := config.Replace()
			if !mapsEqual(got, tt.want) {
				t.Errorf("WithReplaceItems() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test alternate form using WithReplaceItem
	t.Run("WithReplaceItem_helper", func(t *testing.T) {
		initial := Config{}
		config := WithReplaceItems(
			func(yield func(string, string) bool) {
				if !yield("hello", "world") {
					return
				}
				_ = yield("foo", "bar")
			},
		)(initial)
		want := map[string]string{"hello": "world", "foo": "bar"}
		got := config.Replace()
		if !mapsEqual(got, want) {
			t.Errorf("WithReplaceItem() = %v, want %v", got, want)
		}
	})
}

// TestSplit tests the internal split function which is key to Config.Seq behavior
func TestSplit(t *testing.T) {
	tests := []struct {
		name  string
		delim string
		input []string
		want  []string
	}{
		{
			name:  "empty_input",
			delim: ",",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "nil_input",
			delim: ",",
			input: nil,
			want:  []string{},
		},
		{
			name:  "single_item_no_delimiter",
			delim: ",",
			input: []string{"hello"},
			want:  []string{"hello"},
		},
		{
			name:  "single_item_with_delimiter",
			delim: ",",
			input: []string{"hello,world"},
			want:  []string{"hello", "world"},
		},
		{
			name:  "multiple_items_no_delimiters",
			delim: ",",
			input: []string{"hello", "world"},
			want:  []string{"hello", "world"},
		},
		{
			name:  "multiple_items_with_delimiters",
			delim: ",",
			input: []string{"hello,there", "world,wide,web"},
			want:  []string{"hello", "there", "world", "wide", "web"},
		},
		{
			name:  "empty_delimiter",
			delim: "",
			input: []string{"hello", "world"},
			want:  []string{"hello", "world"},
		},
		{
			name:  "empty_elements_skipped",
			delim: ",",
			input: []string{"hello,,world", ",,,", "test"},
			want:  []string{"hello", "world", "test"},
		},
		{
			name:  "empty_strings_skipped",
			delim: ",",
			input: []string{"", "hello"},
			want:  []string{"hello"},
		},
		{
			name:  "only_delimiters_skipped",
			delim: ",",
			input: []string{",", ",,"},
			want:  []string{},
		},
		{
			name:  "duplicate_whole_strings_skipped",
			delim: ",",
			input: []string{"hello", "hello", "world"},
			want:  []string{"hello", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.Collect(split(tt.delim, tt.input...))
			if !slicesEqual(got, tt.want) {
				t.Errorf("split() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test early termination
	t.Run("early_termination", func(t *testing.T) {
		result := []string{}
		split(",", "a,b,c,d,e")(func(s string) bool {
			if s == "c" {
				return false
			}
			result = append(result, s)
			return true
		})

		want := []string{"a", "b"}
		if !slicesEqual(result, want) {
			t.Errorf("split() with early termination = %v, want %v", result, want)
		}
	})
}

// TestMemoType tests the internal memo type and its methods
func TestMemoType(t *testing.T) {
	t.Run("empty_memo", func(t *testing.T) {
		m := memo[string]{}
		if m.contains("test") {
			t.Errorf("empty memo should not contain 'test'")
		}
	})

	t.Run("add_and_contains", func(t *testing.T) {
		m := memo[string]{}
		m.add("test")
		if !m.contains("test") {
			t.Errorf("memo should contain 'test' after adding it")
		}
	})

	t.Run("seen_first_call", func(t *testing.T) {
		m := memo[string]{}
		if m.seen("test") {
			t.Errorf("first call to seen() should return false")
		}
		if !m.contains("test") {
			t.Errorf("memo should contain 'test' after seen() call")
		}
	})

	t.Run("seen_second_call", func(t *testing.T) {
		m := memo[string]{}
		m.seen("test")
		if !m.seen("test") {
			t.Errorf("second call to seen() should return true")
		}
	})

	t.Run("memoItems_single_slice", func(t *testing.T) {
		m := memoItems([]string{"a", "b", "c"})
		if !m.contains("a") || !m.contains("b") || !m.contains("c") {
			t.Errorf("memoItems failed to properly initialize memo")
		}
	})

	t.Run("memoItems_multiple_slices", func(t *testing.T) {
		m := memoItems([]string{"a", "b"}, []string{"c", "d"})
		if !m.contains("a") || !m.contains("b") || !m.contains("c") || !m.contains("d") {
			t.Errorf("memoItems failed to properly initialize memo from multiple slices")
		}
	})

	t.Run("memoItems_with_duplicates", func(t *testing.T) {
		m := memoItems([]string{"a", "b"}, []string{"b", "c"})
		expected := map[string]struct{}{
			"a": {},
			"b": {},
			"c": {},
		}
		if !reflect.DeepEqual(map[string]struct{}(m), expected) {
			t.Errorf("memoItems didn't handle duplicates correctly")
		}
	})

	t.Run("memoItems_with_empty", func(t *testing.T) {
		m := memoItems([]string{}, []string{"a", "b"})
		expected := map[string]struct{}{
			"a": {},
			"b": {},
		}
		if !reflect.DeepEqual(map[string]struct{}(m), expected) {
			t.Errorf("memoItems didn't handle empty slice correctly")
		}
	})
}

// TestSumLen tests the internal sumLen function
func TestSumLen(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  int
	}{
		{
			name:  "empty",
			input: []string{},
			want:  0,
		},
		{
			name:  "empty_string",
			input: []string{""},
			want:  0,
		},
		{
			name:  "single_string",
			input: []string{"hello"},
			want:  5,
		},
		{
			name:  "multiple_strings",
			input: []string{"hello", "world"},
			want:  10,
		},
		{
			name:  "with_empty_strings",
			input: []string{"", "hello", ""},
			want:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sumLen(tt.input...); got != tt.want {
				t.Errorf("sumLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to compare slices for equality
func slicesEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Helper function to compare maps for equality
func mapsEqual[K, V comparable](a, b map[K]V) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		if vb, ok := b[k]; !ok || va != vb {
			return false
		}
	}
	return true
}

// Helper function to compare Config objects for equality
func configEqual(a, b Config) bool {
	if !slicesEqual(a.subject, b.subject) {
		return false
	}
	if a.delim != b.delim {
		return false
	}
	if !slicesEqual(a.remove, b.remove) {
		return false
	}
	if !slicesEqual(a.prefix, b.prefix) {
		return false
	}
	if !slicesEqual(a.suffix, b.suffix) {
		return false
	}
	return mapsEqual(a.replace, b.replace)
}

// TestComplexWorkflow tests a complex workflow combining multiple operations
func TestComplexWorkflow(t *testing.T) {
	// Create a config with all options set
	c := Make(
		WithSubject("one:two:three"),
		WithDelim(":"),
		WithRemove("two"),
		WithPrefix("start"),
		WithSuffix("end"),
		WithReplace(map[string]string{"three": "THREE"}),
	)

	// Test that all options were set correctly
	if !slicesEqual(c.Subject(), []string{"one:two:three"}) {
		t.Errorf("Subject was not set correctly: %v", c.Subject())
	}

	if c.Delim() != ":" {
		t.Errorf("Delim was not set correctly: %v", c.Delim())
	}

	if !slicesEqual(c.Remove(), []string{"two"}) {
		t.Errorf("Remove was not set correctly: %v", c.Remove())
	}

	if !slicesEqual(c.Prefix(), []string{"start"}) {
		t.Errorf("Prefix was not set correctly: %v", c.Prefix())
	}

	if !slicesEqual(c.Suffix(), []string{"end"}) {
		t.Errorf("Suffix was not set correctly: %v", c.Suffix())
	}

	if !mapsEqual(c.Replace(), map[string]string{"three": "THREE"}) {
		t.Errorf("Replace was not set correctly: %v", c.Replace())
	}

	// Test the sequence
	expectedSeq := []string{"start", "one", "THREE", "end"}
	gotSeq := slices.Collect(c.Seq())
	if !slicesEqual(gotSeq, expectedSeq) {
		t.Errorf("Seq() = %v, want %v", gotSeq, expectedSeq)
	}

	// Test the string output
	expectedStr := "start:one:THREE:end"
	gotStr := c.String()
	if gotStr != expectedStr {
		t.Errorf("String() = %v, want %v", gotStr, expectedStr)
	}

	// Modify the config with additional options
	c = Wrap(c,
		WithSubjectItems("four:five"),
		WithRemoveItems("five"),
		WithPrefixItems("prestart"),
		WithSuffixItems("endend"),
	)

	// Verify modified configuration
	expectedSeq = []string{"prestart", "start", "one", "THREE", "four", "end", "endend"}
	gotSeq = slices.Collect(c.Seq())
	if !slicesEqual(gotSeq, expectedSeq) {
		t.Errorf("Modified Seq() = %v, want %v", gotSeq, expectedSeq)
	}

	// Verify string output after modification
	expectedStr = "prestart:start:one:THREE:four:end:endend"
	gotStr = c.String()
	if gotStr != expectedStr {
		t.Errorf("Modified String() = %v, want %v", gotStr, expectedStr)
	}
}

// TestCustomOptionType tests using the Option type with a custom struct
func TestCustomOptionType(t *testing.T) {
	type CustomConfig struct {
		Name  string
		Value int
	}

	withName := func(name string) Option[CustomConfig] {
		return func(c CustomConfig) CustomConfig {
			c.Name = name
			return c
		}
	}

	withValue := func(value int) Option[CustomConfig] {
		return func(c CustomConfig) CustomConfig {
			c.Value = value
			return c
		}
	}

	// Test Make
	c := Make(
		withName("test"),
		withValue(42),
	)

	if c.Name != "test" || c.Value != 42 {
		t.Errorf("Custom Make() failed, got Name=%s Value=%d, want Name=test Value=42", c.Name, c.Value)
	}

	// Test Wrap
	initial := CustomConfig{Name: "initial", Value: 0}
	c = Wrap(initial, withValue(99))

	if c.Name != "initial" || c.Value != 99 {
		t.Errorf("Custom Wrap() failed, got Name=%s Value=%d, want Name=initial Value=99", c.Name, c.Value)
	}
}
