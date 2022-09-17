package err3_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/gregwebs/err3/err3"
	"github.com/gregwebs/err3/try"
)

func throw() (string, error) {
	return "", fmt.Errorf("this is an ERROR")
}

func twoStrNoThrow() (string, string, error)        { return "test", "test", nil }
func intStrNoThrow() (int, string, error)           { return 1, "test", nil }
func boolIntStrNoThrow() (bool, int, string, error) { return true, 1, "test", nil }
func noThrow() (string, error)                      { return "test", nil }

func recursion(a int) (r int, err error) {
	defer err3.Handle(&err, nil)

	if a == 0 {
		return 0, nil
	}
	s := try.Check1(noThrow())
	try.Check(err)
	_ = s
	r = try.Check1(recursion(a - 1))
	r += a
	return r, nil
}

func cleanRecursion(a int) int {
	if a == 0 {
		return 0
	}
	s := try.Check1(noThrow())
	_ = s
	return a + cleanRecursion(a-1)
}

func recursionWithErrorCheck(a int) (int, error) {
	if a == 0 {
		return 0, nil
	}
	s, err := noThrow()
	if err != nil {
		return 0, err
	}
	_ = s
	v, err := recursionWithErrorCheck(a - 1)
	if err != nil {
		return 0, err
	}
	return a + v, nil
}

func errHandlefOnly() (err error) {
	defer err3.Handlef(&err, "handle top")
	defer err3.Handlef(&err, "handle error")
	_ = try.Check1(throw())
	defer err3.Handlef(&err, "handle error")
	_ = try.Check1(throw())
	defer err3.Handlef(&err, "handle error")
	_ = try.Check1(throw())
	return err
}

func errTry1_Fmt() (err error) {
	defer err3.Handlef(&err, "handle top")
	// _ = try.Try1(throw())(func(err error) error { return fmt.Errorf("handle error: %v", err) })
	_ = try.Try1(throw())(try.Fmt("handle error"))
	_ = try.Try1(throw())(try.Fmt("handle error"))
	_ = try.Try1(throw())(try.Fmt("handle error"))
	return err
}

func errId(err error) error { return err }
func empty() error          { return nil }

func errTry1_id() (err error) {
	defer err3.Handlef(&err, "handle top")
	_ = try.Try1(throw())(errId)
	_ = try.Try1(throw())(errId)
	_ = try.Try1(throw())(errId)
	return err
}

func errHandle_Only() (err error) {
	defer err3.Handlef(&err, "handle top")
	defer err3.Handle(&err, empty)
	_ = try.Check1(throw())
	defer err3.Handle(&err, empty)
	_ = try.Check1(throw())
	defer err3.Handle(&err, empty)
	_ = try.Check1(throw())
	return err
}

func errTry1_inlineHandler() (err error) {
	defer err3.Handlef(&err, "handle top")
	_ = try.Try1(throw())(func(err error) error { return fmt.Errorf("handle error: %v", err) })
	_ = try.Try1(throw())(func(err error) error { return fmt.Errorf("handle error: %v", err) })
	_ = try.Try1(throw())(func(err error) error { return fmt.Errorf("handle error: %v", err) })
	return err
}

func noErr() error {
	return nil
}

func TestTry_noError(t *testing.T) {
	try.Check1(noThrow())
	try.Check2(twoStrNoThrow())
	try.Check2(intStrNoThrow())
	try.Check3(boolIntStrNoThrow())
}

func TestDefault_Error(t *testing.T) {
	var err error
	defer err3.Handle(&err, nil)

	try.Check1(throw())

	t.Fail() // If everything works we are newer here
}

func TestTry_Error(t *testing.T) {
	var err error
	defer err3.Handle(&err, nil)

	try.Check1(throw())

	t.Fail() // If everything works we are newer here
}

func TestPanickingCatchAll(t *testing.T) {
	type args struct {
		f func()
	}
	tests := []struct {
		name  string
		args  args
		wants error
	}{
		{"general panic",
			args{
				func() {
					defer err3.CatchAll(func(err error) {}, func(v any) {})
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					defer err3.CatchAll(func(err error) {}, func(v any) {})
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() != nil {
					t.Error("panics should not fly thru")
				}
			}()
			tt.args.f()
		})
	}
}

func TestPanickingCatchTrace(t *testing.T) {
	noPanic := func(v any) {}
	noError := func(err error) {}

	type args struct {
		f func()
	}
	tests := []struct {
		name  string
		args  args
		wants error
	}{
		{"general panic",
			args{
				func() {
					defer err3.CatchAll(noError, noPanic)
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					defer err3.CatchAll(noError, noPanic)
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() != nil {
					t.Error("panics should NOT carry on when tracing")
				}
			}()
			tt.args.f()
		})
	}
}

func TestPanickingCarryOn_Handle(t *testing.T) {
	type args struct {
		f func()
	}
	tests := []struct {
		name  string
		args  args
		wants error
	}{
		{"general panic",
			args{
				func() {
					var err error
					defer err3.Handle(&err, nil)
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					var err error
					defer err3.Handle(&err, nil)
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("panics should went thru when not our errors")
				}
			}()
			tt.args.f()
		})
	}
}

func TestPanicking_Return(t *testing.T) {
	type args struct {
		f func()
	}
	tests := []struct {
		name  string
		args  args
		wants error
	}{
		{"general panic",
			args{
				func() {
					var err error
					defer err3.Handle(&err, nil)
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					var err error
					defer err3.Handle(&err, nil)
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("panics should carry on")
				}
			}()
			tt.args.f()
		})
	}
}

func TestPanicking_Catch(t *testing.T) {
	type args struct {
		f func()
	}
	tests := []struct {
		name  string
		args  args
		wants error
	}{
		{"general panic",
			args{
				func() {
					defer err3.CatchError(func(err error) {})
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					defer err3.CatchError(func(err error) {})
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("panics should carry on")
				}
			}()
			tt.args.f()
		})
	}
}

func TestCatch_Error(t *testing.T) {
	defer err3.CatchError(func(err error) {
		//fmt.Printf("error and defer handling:%s\n", err)
	})

	try.Check1(throw())

	t.Fail() // If everything works we are newer here
}

func Example_copyFile() {
	copyFile := func(src, dst string) (err error) {
		defer err3.Handlef(&err, "copy %s %s", src, dst)

		// These try.To() checkers are as fast as `if err != nil {}`

		r := try.Check1(os.Open(src))
		defer r.Close()

		rmFile := try.Cleanup(func() {
			os.Remove(dst)
		})
		w := try.Try1(os.Create(dst))(rmFile)
		defer w.Close()
		_ = try.Try1(io.Copy(w, r))(rmFile, try.Fmt("copy failure"))
		return nil
	}

	err := copyFile("/notfound/path/file.go", "/notfound/path/file.bak")
	if err != nil {
		fmt.Println(err)
	}
	// Output: copy /notfound/path/file.go /notfound/path/file.bak: open /notfound/path/file.go: no such file or directory
}

func ExampleHandle() {
	var err error
	defer err3.Handle(&err, nil)
	try.Check1(noThrow())
	// Output:
}

func ExampleHandlef() {
	annotated := func() (err error) {
		defer err3.Handlef(&err, "annotated")
		try.Check1(throw())
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated: this is an ERROR
}

func ExampleHandlef_format_args() {
	annotated := func() (err error) {
		defer err3.Handlef(&err, "annotated: %s", "err3")
		try.Check1(throw())
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated: err3: this is an ERROR
}

func ExampleHandlef_panic() {
	type fn func(v int) int
	var recursion fn
	const recursionLimit = 77 // 12+11+10+9+8+7+6+5+4+3+2+1 = 78

	recursion = func(i int) int {
		if i > recursionLimit { // simulated error case
			panic(fmt.Errorf("helper failed at: %d", i))
		} else if i == 0 {
			return 0 // recursion without error ends here
		}
		return i + recursion(i-1)
	}

	annotated := func() (err error) {
		defer err3.Handlef(&err, "annotated: %s", "err3")

		r := recursion(12) // call recursive algorithm successfully
		recursion(r)       // call recursive algorithm unsuccessfully
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated: err3: helper failed at: 78
}

func ExampleHandlef_deferStack() {
	annotated := func() (err error) {
		defer err3.Handlef(&err, "3rd")
		defer err3.Handlef(&err, "2nd")
		_ = try.Try1(throw())(try.Fmt("1st"))
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: 3rd: 2nd: 1st: this is an ERROR
}

func ExampleHandle_with_handler() {
	doSomething := func(a, b int) (err error) {
		defer err3.Handle(&err, func() error {
			return fmt.Errorf("error with (%d, %d): %v", a, b, err)
		})
		try.Check1(throw())
		return err
	}
	err := doSomething(1, 2)
	fmt.Printf("%v", err)
	// Output: error with (1, 2): this is an ERROR
}

func BenchmarkOldErrorCheckingWithIfClause(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := noThrow()
		if err != nil {
			return
		}
	}
}

func Benchmark_Err_HandleNil(b *testing.B) {
	for n := 0; n < b.N; n++ {
		errHandle_Only()
	}
}

func Benchmark_Err_Try1_id(b *testing.B) {
	for n := 0; n < b.N; n++ {
		errTry1_id()
	}
}

func Benchmark_Err_HandlersOnly(b *testing.B) {
	for n := 0; n < b.N; n++ {
		errHandlefOnly()
	}
}

func Benchmark_Err_Try1_Fmt(b *testing.B) {
	for n := 0; n < b.N; n++ {
		errTry1_Fmt()
	}
}

func Benchmark_NoErr_Check1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = try.Check1(noThrow()) // we show here what can take time
	}
}

func Benchmark_Noerr3_Check_NilErr(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := noThrow()
		try.Check(err)
	}
}

func Benchmark_NoErr_Check2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = try.Check2(twoStrNoThrow())
	}
}

func Benchmark_NoErr_Check(b *testing.B) {
	for n := 0; n < b.N; n++ {
		try.Check(noErr())
	}
}

func Benchmark_NoErr_Check_NilErr(b *testing.B) {
	for n := 0; n < b.N; n++ {
		err := noErr()
		try.Check(err)
	}
}

func BenchmarkCleanRecursionWithTryCall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = cleanRecursion(100)
	}
}

func BenchmarkRecursionWithCheckAndDefer(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = recursion(100)
	}
}

func BenchmarkRecursionWithOldErrorCheck(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := recursionWithErrorCheck(100)
		if err != nil {
			return
		}
	}
}
