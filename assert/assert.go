package assert

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"
)

var (
	// P is a production Asserter that sets panic objects to errors which
	// allows handle handlers to catch them.
	P = AsserterToError

	// D is a development Asserter that sets panic objects to strings that
	// doesn't by caught by handle handlers.
	D Asserter = AsserterDebug

	// DefaultAsserter is a default asserter used for package-level functions
	// like assert.That(). It is the same as the production asserter P, which
	// treats assert failures as Go errors, but in addition to that, it formats
	// the assertion message properly. Naturally, only if handle handlers are
	// found in the call stack, these errors are caught.
	//
	// You are free to set it according to your current preferences. For
	// example, it might be better to panic about every assertion fault during
	// the tests. When in other cases, throw an error.
	DefaultAsserter = AsserterToError | AsserterFormattedCallerInfo
)

var (
	// testers is must be set if assertion package is used for the unit testing.
	testers map[int]testing.TB = make(map[int]testing.TB)
)

// PushTester sets the current testing context for default asserter. This must
// be called at the beginning of every test.
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			assert.PushTester(t) // <- IMPORTANT!
//			defer assert.PopTester()
//			...
//			assert.That(something, "test won't work")
//		})
//	}
func PushTester(t testing.TB) {
	if DefaultAsserter&AsserterUnitTesting == 0 {
		// if this is forgotten or tests don't have proper place to set it
		// it's good to keep the API as simple as possible
		DefaultAsserter |= AsserterUnitTesting
	}
	testers[goid()] = t
}

// PopTester pops the testing context reference from the memory. This isn't
// totally necessary, but if you want play by book, please do it. Usually done
// by defer after PushTester.
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			assert.PushTester(t) // <- important!
//			defer assert.PopTester() // <- for good girls and not so bad boys
//			...
//			assert.That(something, "test won't work")
//		})
//	}
func PopTester() {
	delete(testers, goid())
}

func tester() testing.TB {
	return testers[goid()]
}

// NotImplemented always panics with 'not implemented' assertion message.
func NotImplemented(a ...any) {
	D.reportAssertionFault("not implemented", a...)
}

// ThatNot asserts that the term is NOT true. If is it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func ThatNot(term bool, a ...any) {
	That(!term, a...)
}

// That asserts that the term is true. If not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func That(term bool, a ...any) {
	if !term {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := "assertion violation"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// NotNil asserts that the value is not nil. If it is it panics/errors (default
// Asserter) with the given message.
func NotNil[T any](p *T, a ...any) {
	if p == nil {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := "assertion violation: pointer is nil"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// SNil asserts that the slice IS nil. If it is it panics/errors (default
// Asserter) with the given message.
func SNil[T any](s []T, a ...any) {
	if s != nil {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := "assertion violation: slice MUST be nil"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// SNotNil asserts that the slice is not nil. If it is it panics/errors (default
// Asserter) with the given message.
func SNotNil[T any](s []T, a ...any) {
	if s == nil {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := "assertion violation: slice is nil"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// CNotNil asserts that the channel is not nil. If it is it panics/errors
// (default Asserter) with the given message.
func CNotNil[T any](c chan T, a ...any) {
	if c == nil {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := "assertion violation: channel is nil"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// MNotNil asserts that the map is not nil. If it is it panics/errors (default
// Asserter) with the given message.
func MNotNil[T comparable, U any](m map[T]U, a ...any) {
	if m == nil {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := "assertion violation: map is nil"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// NotEqual asserts that the values aren't equal. If they are it panics/errors
// (current Asserter) with the given message.
func NotEqual[T comparable](val, want T, a ...any) {
	if want == val {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf("assertion violation: got %v, want %v", val, want)
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// Equal asserts that the values are equal. If not it panics/errors (current
// Asserter) with the given message.
func Equal[T comparable](val, want T, a ...any) {
	if want != val {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf("assertion violation: got %v, want %v", val, want)
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// SLen asserts that the length of the slice is equal to the given. If not it
// panics/errors (current Asserter) with the given message. Note! This is
// reasonably fast but not as fast as 'That' because of lacking inlining for the
// current implementation of Go's type parametric functions.
func SLen[T any](obj []T, length int, a ...any) {
	l := len(obj)

	if l != length {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf("assertion violation: got %d, want %d", l, length)
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// MLen asserts that the length of the map is equal to the given. If not it
// panics/errors (current Asserter) with the given message. Note! This is
// reasonably fast but not as fast as 'That' because of lacking inlining for the
// current implementation of Go's type parametric functions.
func MLen[T comparable, U any](obj map[T]U, length int, a ...any) {
	l := len(obj)

	if l != length {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf("assertion violation: got %d, want %d", l, length)
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// NotEmpty asserts that the string is not empty. If it is, it panics/errors
// (current Asserter) with the given message.
func NotEmpty(obj string, a ...any) {
	if obj == "" {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := "assertion violation: string shouldn't be empty"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// SNotEmpty asserts that the slice is not empty. If it is, it panics/errors
// (current Asserter) with the given message. Note! This is reasonably fast but
// not as fast as 'That' because of lacking inlining for the current
// implementation of Go's type parametric functions.
func SNotEmpty[T any](obj []T, a ...any) {
	l := len(obj)

	if l == 0 {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := "assertion violation: slice shouldn't be empty"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// MNotEmpty asserts that the map is not empty. If it is, it panics/errors
// (current Asserter) with the given message. Note! This is reasonably fast but
// not as fast as 'That' because of lacking inlining for the current
// implementation of Go's type parametric functions.
func MNotEmpty[T comparable, U any](obj map[T]U, length int, a ...any) {
	l := len(obj)

	if l == 0 {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := "assertion violation: map shouldn't be empty"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// NoError asserts that the error is nil. If is not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func NoError(err error, a ...any) {
	if err != nil {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := "assertion violation: " + err.Error()
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// Error asserts that the err is not nil. If it is it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func Error(err error, a ...any) {
	if err == nil {
		if DefaultAsserter.isUnitTesting() {
			tester().Helper()
		}
		defMsg := "assertion violation: missing error"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

func combineArgs(format string, a []any) []any {
	args := make([]any, 1, len(a)+1)
	args[0] = format
	args = append(args, a...)
	return args
}

func goid() int {
	var buf [64]byte
	runtime.Stack(buf[:], false)
	var id int
	_, err := fmt.Fscanf(bytes.NewReader(buf[:]), "goroutine %d", &id)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}
