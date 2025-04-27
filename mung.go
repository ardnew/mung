// Package mung provides methods to manipulate strings commonly found in
// PATH-like environment variables.
package mung

import (
	"iter"
	"maps"
	"slices"
	"strings"
)

// Option functions return their argument with modifications applied.
type Option[T any] func(T) T

// Make returns a new object of type T with the given options applied.
func Make[T any](opts ...Option[T]) (t T) { return Wrap(t, opts...) }

// Wrap returns t after applying the given options.
func Wrap[T any](t T, opts ...Option[T]) T {
	for _, o := range opts {
		t = o(t)
	}
	return t
}

// Config represents the configuration for string munging operations.
type Config struct {
	subject []string
	delim   string
	remove  []string
	prefix  []string
	suffix  []string
	replace map[string]string
}

// String returns the munged strings joined with the configuration's delimiter.
func (c Config) String() string {
	n := sumLen(c.prefix...) + sumLen(c.suffix...) +
		sumLen(c.subject...) + sumLen(slices.Collect(maps.Values(c.replace))...) +
		max(0, len(c.delim)*
			(len(c.prefix)+len(c.suffix)+len(c.subject)+len(c.replace)-1))
	var sb strings.Builder
	sb.Grow(n)
	c.Seq()(func(s string) bool {
		if sb.Len() > 0 {
			sb.WriteString(c.delim)
		}
		sb.WriteString(s)
		return true
	})
	return sb.String()
}

// Subject returns the subject strings to be processed.
func (c Config) Subject() []string { return c.subject }

// Delim returns the delimiter used for splitting and joining strings.
func (c Config) Delim() string { return c.delim }

// Remove returns the list of strings to be removed during processing.
func (c Config) Remove() []string { return c.remove }

// Prefix returns the list of strings to be prepended to the result.
func (c Config) Prefix() []string { return c.prefix }

// Suffix returns the list of strings to be appended to the result.
func (c Config) Suffix() []string { return c.suffix }

// Replace returns a copy of the string replacement map.
func (c Config) Replace() map[string]string { return maps.Clone(c.replace) }

// Seq returns each string item from the munged sequence.
func (c Config) Seq() iter.Seq[string] {
	prev := memo[string]{}
	yieldSeq := func(seq []string, omit memo[string], yield func(string) bool) bool {
		for s := range split(c.delim, seq...) {
			if !prev.seen(s) && !omit.contains(s) {
				if r, ok := c.replace[s]; ok {
					if !yield(r) {
						return false
					}
				} else if !yield(s) {
					return false
				}
			}
		}
		return true
	}
	return func(yield func(string) bool) {
		if yieldSeq(c.prefix, memoItems(c.remove), yield) {
			if yieldSeq(c.subject, memoItems(c.remove, c.prefix, c.suffix), yield) {
				_ = yieldSeq(c.suffix, memoItems(c.remove), yield)
			}
		}
	}
}

// WithSubject returns an option that sets all subject strings to be processed.
func WithSubject(subjects ...string) Option[Config] {
	return func(config Config) Config {
		config.subject = subjects
		return config
	}
}

// WithSubjectItems returns an option that adds subject strings to be processed.
func WithSubjectItems(subjects ...string) Option[Config] {
	return func(config Config) Config {
		if config.subject == nil {
			config.subject = make([]string, 0, len(subjects))
		}
		config.subject = append(config.subject, subjects...)
		return config
	}
}

// WithDelim returns an option that sets the string tokenizing delimiter.
func WithDelim(delim string) Option[Config] {
	return func(config Config) Config {
		config.delim = delim
		return config
	}
}

// WithRemove returns an option that sets all strings to remove
// during processing.
func WithRemove(removes ...string) Option[Config] {
	return func(config Config) Config {
		config.remove = removes
		return config
	}
}

// WithRemoveItems returns an option that adds strings to remove
// during processing.
func WithRemoveItems(removes ...string) Option[Config] {
	return func(config Config) Config {
		if config.remove == nil {
			config.remove = make([]string, 0, len(removes))
		}
		config.remove = append(config.remove, removes...)
		return config
	}
}

// WithPrefix returns an option that sets all strings to prepend
// after processing.
//
// Strings are prepended in the order they are given,
// meaning the trailing argument will lead the result;
// or, the leading argument is the first to be prepended.
func WithPrefix(prefixes ...string) Option[Config] {
	return func(config Config) Config {
		slices.Reverse(prefixes)
		config.prefix = prefixes
		return config
	}
}

// WithPrefixItems returns an option that adds strings to prepend
// after processing.
//
// Strings are prepended in the order they are given,
// meaning the trailing argument will lead the result;
// or, the leading argument is the first to be prepended.
func WithPrefixItems(prefixes ...string) Option[Config] {
	return func(config Config) Config {
		if config.prefix == nil {
			config.prefix = make([]string, 0, len(prefixes))
		}
		slices.Reverse(prefixes)
		config.prefix = append(prefixes, config.prefix...)
		return config
	}
}

// WithSuffix returns an option that sets all strings to append
// after processing.
//
// Strings are appended in the order they are given,
// meaning the trailing argument will trail the result;
// or, the leading argument is the first to be appended.
func WithSuffix(suffixes ...string) Option[Config] {
	return func(config Config) Config {
		config.suffix = suffixes
		return config
	}
}

// WithSuffixItems returns an option that adds strings to append
// after processing.
func WithSuffixItems(suffixes ...string) Option[Config] {
	return func(config Config) Config {
		if config.suffix == nil {
			config.suffix = make([]string, 0, len(suffixes))
		}
		config.suffix = append(config.suffix, suffixes...)
		return config
	}
}

// WithReplace returns an option that sets all whole/fixed-string substitution
// rules applied after processing.
func WithReplace(replace map[string]string) Option[Config] {
	return func(config Config) Config {
		config.replace = replace
		return config
	}
}

// WithReplaceItem returns an option that adds individual whole/fixed-string
// substitution rules applied after processing.
//
// The given sequence seq must yield paired strings in which
// the first string is the item to replace, and the second is the replacement.
func WithReplaceItems(replacements iter.Seq2[string, string]) Option[Config] {
	return func(config Config) Config {
		if config.replace == nil {
			config.replace = make(map[string]string)
		}
		maps.Insert(config.replace, replacements)
		return config
	}
}

func split(delim string, each ...string) iter.Seq[string] {
	return func(yield func(string) bool) {
		undelimited := delim == ""
		memo := memo[string]{}
		for _, s := range each {
			switch {
			case memo.seen(s):
				continue
			case strings.ReplaceAll(s, delim, "") == "": // skip empty elements
				continue
			case undelimited || !strings.Contains(s, delim): // yield entire element
				if !yield(s) {
					return
				}
			default: // yield delimited elements
				for e := range strings.SplitSeq(s, delim) {
					if e != "" && !yield(e) { // skip empty elements after split
						return
					}
				}
			}
		}
	}
}

func sumLen(each ...string) int {
	sum := 0
	for _, s := range each {
		sum += len(s)
	}
	return sum
}

type memo[T comparable] map[T]struct{}

func (m memo[T]) contains(item T) bool {
	_, ok := m[item]
	return ok
}

func (m memo[T]) add(item T) {
	m[item] = struct{}{}
}

func (m memo[T]) seen(item T) bool {
	if m.contains(item) {
		return true
	}
	m.add(item)
	return false
}

func memoItems[T comparable](slices ...[]T) memo[T] {
	memo := make(memo[T], len(slices))
	for _, slice := range slices {
		for _, item := range slice {
			memo[item] = struct{}{}
		}
	}
	return memo
}
