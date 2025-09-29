// Package mung provides methods to manipulate strings commonly found in
// PATH-like environment variables.
package mung

import (
	"iter"
	"maps"
	"slices"
	"strings"

	_ "embed"
)

//go:embed VERSION
var version string

// Version returns the version of the mung package.
func Version() string { return strings.TrimSpace(version) }

// Option functions return their argument with modifications applied.
type Option[T any] func(T) T

// Make returns a new object of type T with the given options applied.
//
//nolint:ireturn
func Make[T any](opts ...Option[T]) (t T) { return Wrap(t, opts...) }

// Wrap returns t after applying the given options.
//
//nolint:ireturn
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

	predicate func(string) bool
}

// String returns the munged strings joined with the configuration's delimiter.
func (c Config) String() string {
	bufLen := sumLen(c.prefix) + sumLen(c.suffix) +
		sumLen(c.subject) + sumLen(slices.Collect(maps.Values(c.replace))) +
		max(0, len(c.delim)*
			(len(c.prefix)+len(c.suffix)+len(c.subject)+len(c.replace)-1))

	var sb strings.Builder

	sb.Grow(bufLen)

	c.Filtered()(func(s string) bool {
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

// Predicate returns the predicate function used to select yielded strings.
func (c Config) Predicate() func(string) bool {
	if c.predicate == nil {
		// Until [Config.predicate] is initialized by the user,
		// return a default predicate that accepts all elements unconditionally.
		return func(string) bool { return true }
	}

	return c.predicate
}

// All returns each string item from the munged sequence.
func (c Config) All() iter.Seq[string] { return c.seq(false) }

// Filtered returns each string item from the munged sequence
// that satisfies the predicate function [Config.Predicate].
func (c Config) Filtered() iter.Seq[string] { return c.seq(true) }

// seq returns a sequence that yields munged strings using rules defined in the
// receiver configuration [Config].
func (c Config) seq(filter bool) iter.Seq[string] {
	prev := memo[string]{}
	yieldSeq := func(
		seq []string, omit memo[string], yield func(string) bool,
	) bool {
		var itemSeq iter.Seq[string]

		if filter {
			// Every element must satisfy the predicate method [Config.filter]
			itemSeq = c.filter(split(c.delim, seq))
		} else {
			itemSeq = split(c.delim, seq)
		}

		for s := range itemSeq {
			if !omit.contains(s) && !prev.seen(s) {
				if r, ok := c.replace[s]; ok {
					s = r
				}

				if !yield(s) {
					return false
				}
			}
		}

		return true
	}

	// [Prefix] documentation states that the trailing element leads the result.
	// Since a Config's prefix elements can be modified multiple times,
	// the ordering semantics must be applied after Config is finalized.
	// The Config is finalized upon realizing the sequence here ([Config.seq]).
	//
	// Each element in prefix may already be delimited.
	// This allows the user to override the reversal logic mentioned above.
	// So we only want to reverse the order of the outer elements in prefix,
	// while preserving the order of the inner, pre-delimited elements.
	//
	// Items yielded via yieldSeq are memoized to prevent duplicates.
	// So the only items that need to be omitted are the ones
	// that have not yet been (or never will be) yielded.

	return func(yield func(string) bool) {
		if yieldSeq(reverse(c.prefix), memoize(split(c.delim, c.remove)), yield) {
			if yieldSeq(
				c.subject,
				memoize(split(c.delim, c.remove, c.suffix)),
				yield,
			) {
				_ = yieldSeq(c.suffix, memoize(split(c.delim, c.remove)), yield)
			}
		}
	}
}

// filter returns a sequence that yields only the elements that satisfy the
// predicate function [Config.Predicate].
func (c Config) filter(seq iter.Seq[string]) iter.Seq[string] {
	// Fast-path instead of the default "accept-all" from [Config.Predicate],
	// just return the given sequence unmodified.
	//
	// The sequence operations initialized in the receiver will still be applied
	// in [Config.seq]. This just affects which elements actually reach those
	// operations.
	if c.predicate == nil {
		return seq // unfiltered
	}

	return func(yield func(string) bool) {
		for s := range seq {
			if c.predicate(s) && !yield(s) {
				return
			}
		}
	}
}

// WithSubject returns an option that sets all subject strings to be processed.
func WithSubject(subjects []string) Option[Config] {
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
func WithRemove(removes []string) Option[Config] {
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
func WithPrefix(prefixes []string) Option[Config] {
	return func(config Config) Config {
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

		config.prefix = append(config.prefix, prefixes...)

		return config
	}
}

// WithSuffix returns an option that sets all strings to append
// after processing.
//
// Strings are appended in the order they are given,
// meaning the trailing argument will trail the result;
// or, the leading argument is the first to be appended.
func WithSuffix(suffixes []string) Option[Config] {
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
// rules to apply after processing.
func WithReplace(replace map[string]string) Option[Config] {
	return func(config Config) Config {
		config.replace = replace

		return config
	}
}

// WithReplaceItem returns an option that adds an individual whole/fixed-string
// substitution rule to apply after processing.
func WithReplaceItem(from, to string) Option[Config] {
	return func(config Config) Config {
		if config.replace == nil {
			config.replace = make(map[string]string)
		}

		config.replace[from] = to

		return config
	}
}

// WithReplaceEach returns an option that adds individual whole/fixed-string
// substitution rules to apply after processing.
//
// The given sequence seq must yield paired strings in which
// the first string is the item to replace, and the second is the replacement.
func WithReplaceEach(replacements iter.Seq2[string, string]) Option[Config] {
	return func(config Config) Config {
		if config.replace == nil {
			config.replace = make(map[string]string)
		}

		maps.Insert(config.replace, replacements)

		return config
	}
}

// WithReplaceItems returns an option that adds whole/fixed-string substitution
// rules to apply after processing.
//
// The given replacements must yield maps in which
// each map's key is the item to replace, and the value is the replacement.
func WithReplaceItems(replacements ...map[string]string) Option[Config] {
	return func(config Config) Config {
		if config.replace == nil {
			config.replace = make(map[string]string)
		}

		for _, r := range replacements {
			maps.Copy(config.replace, r)
		}

		return config
	}
}

// WithFilter returns an option that sets the predicate function used to
// select yielded strings.
func WithFilter(predicate func(string) bool) Option[Config] {
	return func(config Config) Config {
		config.predicate = predicate

		return config
	}
}

// Reverse returns a copy of the given slice in reverse order.
// The given slice is not modified.
// Use [slices.reverse] to reverse a slice in-place.
func reverse[T any](s []T) []T {
	r := make([]T, len(s))
	for i := len(s) - 1; i >= 0; i-- {
		r[len(s)-1-i] = s[i]
	}

	return r
}

// Split returns a sequence of strings, split by the given delimiter,
// from each of the given slices.
//
// Wrap the result in [unique] to elide duplicates.
//
// The given slices are not modified.
func split(delim string, slices ...[]string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, slice := range slices {
			for _, str := range slice {
				for s := range strings.SplitSeq(str, delim) {
					if strings.ReplaceAll(s, delim, "") == "" { // skip empty elements
						continue
					}

					if !yield(s) {
						return
					}
				}
			}
		}
	}
}

func sumLen(each []string) int {
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

func (m memo[T]) add(item ...T) {
	for _, it := range item {
		m[it] = struct{}{}
	}
}

func (m memo[T]) seen(item T) bool {
	if m.contains(item) {
		return true
	}

	m.add(item)

	return false
}

func memoize[T comparable](items iter.Seq[T]) memo[T] {
	m := memo[T]{}
	m.add(slices.Collect(uniq(items))...)

	return m
}

// uniq returns a sequence that yields only unique items
// from the given sequence, preserving the order of first appearance.
func uniq[T comparable](items iter.Seq[T]) iter.Seq[T] {
	if items == nil {
		return func(func(T) bool) {}
	}

	return func(yield func(T) bool) {
		memo := memo[T]{}
		for item := range items {
			if !memo.seen(item) {
				if !yield(item) {
					return
				}
			}
		}
	}
}
