// Code generated automatically in func ExampleVersion(); DO NOT EDIT

// Package mung manipulates PATH-like environment variables.
//
// The package allows for splitting, joining, prefixing, suffixing, removing,
// and replacing elements in delimited strings commonly found in environment
// variables like PATH, LD_LIBRARY_PATH, and similar.
package mung

import (
	"fmt"
	"strings"
)

// ExampleVersion demonstrates how to get the version of the mung package.
func ExampleVersion() {
	version := Version()

	fmt.Println(version)
	// Output: 0.3.0
}

//go:generate bash -c "sed -i'' ${DOLLAR}((${GOLINE}-3))$'s|.*|\t// Output: '\"${DOLLAR}(cat VERSION)\"'|' '${GOFILE}'"

// ExampleMake demonstrates creating a new Config with options.
func ExampleMake() {
	// Create a new Config with PATH-like variable
	config := Make(
		WithSubject([]string{"/usr/local/bin:/usr/bin:/bin"}),
		WithDelim(":"),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/usr/bin:/bin
}

// ExampleWrap demonstrates applying options to an existing object.
func ExampleWrap() {
	// Create an empty Config
	var config Config

	// Apply options to the existing Config
	config = Wrap(config,
		WithSubject([]string{"/usr/local/bin:/usr/bin"}),
		WithDelim(":"),
		WithSuffix([]string{"/opt/bin"}),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/usr/bin:/opt/bin
}

// ExampleWithSubject demonstrates setting the subject strings.
func ExampleWithSubject() {
	config := Make(
		WithSubject([]string{"/usr/local/bin:/usr/bin:/bin"}),
		WithDelim(":"),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/usr/bin:/bin
}

// ExampleWithSubjectItems demonstrates adding to subject strings.
func ExampleWithSubjectItems() {
	config := Make(
		WithSubject([]string{"/usr/local/bin"}),
		WithDelim(":"),
		WithSubjectItems("/usr/bin", "/bin"),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/usr/bin:/bin
}

// ExampleWithDelim demonstrates setting the delimiter.
func ExampleWithDelim() {
	// Create config using semicolon as delimiter (Windows PATH-like)
	config := Make(
		WithSubject([]string{"C:\\Program Files\\App;C:\\Windows"}),
		WithDelim(";"),
	)

	fmt.Println(config.String())
	// Output: C:\Program Files\App;C:\Windows
}

// ExampleWithRemove demonstrates removing elements.
func ExampleWithRemove() {
	config := Make(
		WithSubject([]string{"/usr/local/bin:/usr/bin:/bin:/opt/bin"}),
		WithDelim(":"),
		WithRemove([]string{"/usr/bin", "/opt/bin"}),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/bin
}

// ExampleWithRemoveItems demonstrates adding items to remove.
func ExampleWithRemoveItems() {
	config := Make(
		WithSubject([]string{"/usr/local/bin:/usr/bin:/bin:/opt/bin"}),
		WithDelim(":"),
		WithRemove([]string{"/usr/bin"}),
		WithRemoveItems("/opt/bin"),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/bin
}

// ExampleWithPrefix demonstrates prepending elements.
func ExampleWithPrefix() {
	config := Make(
		WithSubject([]string{"/usr/bin:/bin"}),
		WithDelim(":"),
		WithPrefix([]string{"/usr/local/bin"}),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/usr/bin:/bin
}

// ExampleWithPrefix_order shows how ordering works with multiple prefixes.
func ExampleWithPrefix_order() {
	// Note: prefixes are prepended in reverse order of arguments
	config := Make(
		WithSubject([]string{"/bin"}),
		WithDelim(":"),
		WithPrefix([]string{"/usr/local/bin", "/usr/bin"}),
	)

	fmt.Println(config.String())
	// Output: /usr/bin:/usr/local/bin:/bin
}

// ExampleWithPrefixItems demonstrates adding prefixes.
func ExampleWithPrefixItems() {
	config := Make(
		WithSubject([]string{"/bin"}),
		WithDelim(":"),
		WithPrefix([]string{"/usr/bin"}),
		WithPrefixItems("/usr/local/bin"),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/usr/bin:/bin
}

// ExampleWithSuffix demonstrates appending elements.
func ExampleWithSuffix() {
	config := Make(
		WithSubject([]string{"/usr/local/bin:/usr/bin"}),
		WithDelim(":"),
		WithSuffix([]string{"/bin", "/opt/bin"}),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/usr/bin:/bin:/opt/bin
}

// ExampleWithSuffixItems demonstrates adding suffixes.
func ExampleWithSuffixItems() {
	config := Make(
		WithSubject([]string{"/usr/local/bin:/usr/bin"}),
		WithDelim(":"),
		WithSuffix([]string{"/bin"}),
		WithSuffixItems("/opt/bin"),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/usr/bin:/bin:/opt/bin
}

// ExampleWithReplace demonstrates replacing elements.
func ExampleWithReplace() {
	config := Make(
		WithSubject([]string{"/usr/local/bin:/usr/bin:/bin"}),
		WithDelim(":"),
		WithReplace(map[string]string{
			"/usr/bin": "/opt/bin",
			"/bin":     "/sbin",
		}),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/opt/bin:/sbin
}

// ExampleWithReplaceItem demonstrates adding a single replacement rule.
func ExampleWithReplaceItem() {
	config := Make(
		WithSubject([]string{"/usr/local/bin:/usr/bin:/bin"}),
		WithDelim(":"),
		WithReplaceItem("/usr/bin", "/opt/bin"),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/opt/bin:/bin
}

// ExampleWithReplaceEach demonstrates adding replacement rules using an iterator.
func ExampleWithReplaceEach() {
	// Define a replacement sequence
	replacements := func(yield func(string, string) bool) {
		if !yield("/usr/bin", "/opt/bin") {
			return
		}
		_ = yield("/bin", "/sbin")
	}

	config := Make(
		WithSubject([]string{"/usr/local/bin:/usr/bin:/bin"}),
		WithDelim(":"),
		WithReplaceEach(replacements),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/opt/bin:/sbin
}

// ExampleWithReplaceItems demonstrates adding replacement rules using maps.
func ExampleWithReplaceItems() {
	replacements := map[string]string{
		"/usr/bin": "/opt/bin",
		"/bin":     "/sbin",
	}

	config := Make(
		WithSubject([]string{"/usr/local/bin:/usr/bin:/bin"}),
		WithDelim(":"),
		WithReplaceItems(replacements),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/opt/bin:/sbin
}

// ExampleWithPredicate demonstrates filtering elements using a predicate.
func ExampleWithPredicate() {
	config := Make(
		WithSubject([]string{"/usr/local/bin:/usr/bin:/bin"}),
		WithDelim(":"),
		WithPredicate(func(s string) bool {
			return !strings.Contains(s, "local")
		}),
	)

	fmt.Println(config.String())
	// Output: /usr/bin:/bin
}

// ExampleConfig_String demonstrates the String method.
func ExampleConfig_String() {
	config := Make(
		WithSubject([]string{"/usr/local/bin:/usr/bin:/bin"}),
		WithDelim(":"),
	)

	// Convert the munged result to a string
	result := config.String()
	fmt.Println(result)
	// Output: /usr/local/bin:/usr/bin:/bin
}

// ExampleConfig_Subject demonstrates accessing the subject strings.
func ExampleConfig_Subject() {
	config := Make(
		WithSubject([]string{"/usr/local/bin", "/usr/bin", "/bin"}),
	)

	subject := config.Subject()
	fmt.Println(strings.Join(subject, ", "))
	// Output: /usr/local/bin, /usr/bin, /bin
}

// ExampleConfig_Delim demonstrates getting the delimiter.
func ExampleConfig_Delim() {
	config := Make(
		WithDelim(":"),
	)

	delim := config.Delim()
	fmt.Printf("Delimiter: %q\n", delim)
	// Output: Delimiter: ":"
}

// ExampleConfig_Remove demonstrates accessing removal strings.
func ExampleConfig_Remove() {
	config := Make(
		WithRemove([]string{"/usr/bin", "/opt/bin"}),
	)

	remove := config.Remove()
	fmt.Println(strings.Join(remove, ", "))
	// Output: /usr/bin, /opt/bin
}

// ExampleConfig_Prefix demonstrates accessing prefix strings.
func ExampleConfig_Prefix() {
	config := Make(
		WithPrefix([]string{"/usr/local/bin", "/usr/bin"}),
	)

	prefix := config.Prefix()
	fmt.Println(strings.Join(prefix, ", "))
	// Output: /usr/local/bin, /usr/bin
}

// ExampleConfig_Suffix demonstrates accessing suffix strings.
func ExampleConfig_Suffix() {
	config := Make(
		WithSuffix([]string{"/bin", "/opt/bin"}),
	)

	suffix := config.Suffix()
	fmt.Println(strings.Join(suffix, ", "))
	// Output: /bin, /opt/bin
}

// ExampleConfig_Replace demonstrates accessing replacement rules.
func ExampleConfig_Replace() {
	config := Make(
		WithReplace(map[string]string{
			"/usr/bin": "/opt/bin",
			"/bin":     "/sbin",
		}),
	)

	replace := config.Replace()
	// Sort for consistent output
	keys := []string{"/bin", "/usr/bin"}
	for _, k := range keys {
		if v, ok := replace[k]; ok {
			fmt.Printf("%s -> %s\n", k, v)
		}
	}
	// Output:
	// /bin -> /sbin
	// /usr/bin -> /opt/bin
}

// ExampleConfig_All demonstrates iterating through the processed sequence.
func ExampleConfig_All() {
	config := Make(
		WithSubject([]string{"/usr/bin:/bin"}),
		WithDelim(":"),
		WithPrefix([]string{"/usr/local/bin"}),
		WithSuffix([]string{"/opt/bin"}),
	)

	// Collect results using the iterator
	var result []string
	config.All()(func(s string) bool {
		result = append(result, s)
		return true
	})

	fmt.Println(strings.Join(result, ", "))
	// Output: /usr/local/bin, /usr/bin, /bin, /opt/bin
}

// ExampleConfig_complex demonstrates a more complex workflow.
func ExampleConfig_complex() {
	// Build a PATH-like variable with multiple operations
	config := Make(
		WithSubject([]string{"/usr/bin:/bin:/usr/local/old:/opt/deprecated"}),
		WithDelim(":"),
		WithPrefix([]string{"/usr/local/bin"}),
		WithSuffix([]string{"/opt/bin"}),
		WithRemove([]string{"/usr/local/old", "/opt/deprecated"}),
		WithReplace(map[string]string{"/bin": "/sbin"}),
	)

	fmt.Println(config.String())
	// Output: /usr/local/bin:/usr/bin:/sbin:/opt/bin
}
