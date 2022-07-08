package main

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

var OsExitCheckAnalyser = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "check for os.Exit() in main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					if fn.Name.Name == "main" {
						ast.Inspect(fn, func(node ast.Node) bool {
							if call, ok := node.(*ast.CallExpr); ok {
								if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "Exit" {
									if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == "os" {
										pass.Reportf(ident.NamePos, "os.Exit() in main")
									}
								}
							}
							return true
						})
					}
				}

			}
		}
	}
	return nil, nil
}

func main() {

	mychecks := []*analysis.Analyzer{
		OsExitCheckAnalyser,          // check for os.Exit() in main
		asmdecl.Analyzer,             // report mismatches between assembly files and Go declarations
		assign.Analyzer,              // check for useless assignments
		atomic.Analyzer,              // check for common mistakes using the sync/atomic package
		atomicalign.Analyzer,         // check for non-64-bits-aligned arguments to sync/atomic functions
		bools.Analyzer,               // check for common mistakes involving boolean operators
		buildssa.Analyzer,            // build SSA-form IR for later passes
		buildtag.Analyzer,            // check that +build tags are well-formed and correctly located
		cgocall.Analyzer,             // detect some violations of the cgo pointer passing rules
		composite.Analyzer,           // check for unkeyed composite literals
		copylock.Analyzer,            // check for locks erroneously passed by value
		ctrlflow.Analyzer,            // build a control-flow graph
		deepequalerrors.Analyzer,     // check for calls of reflect.DeepEqual on error values
		errorsas.Analyzer,            // report passing non-pointer or non-error values to errors.As
		fieldalignment.Analyzer,      // find structs that would use less memory if their fields were sorted
		findcall.Analyzer,            // find calls to a particular function
		framepointer.Analyzer,        // report assembly that clobbers the frame pointer before saving it
		httpresponse.Analyzer,        // check for mistakes using HTTP responses
		ifaceassert.Analyzer,         // detect impossible interface-to-interface type assertions
		inspect.Analyzer,             // optimize AST traversal for later passes
		loopclosure.Analyzer,         // check references to loop variables from within nested functions
		lostcancel.Analyzer,          // check cancel func returned by context.WithCancel is called
		nilfunc.Analyzer,             // check for useless comparisons between functions and nil
		nilness.Analyzer,             // check for redundant or impossible nil comparisons
		pkgfact.Analyzer,             // gather name/value pairs from constant declarations
		printf.Analyzer,              // check consistency of Printf format strings and arguments
		reflectvaluecompare.Analyzer, // check for comparing reflect.Value values with == or reflect.DeepEqual
		shadow.Analyzer,              // check for possible unintended shadowing of variables
		shift.Analyzer,               // check for shifts that equal or exceed the width of the integer
		sigchanyzer.Analyzer,         // check for unbuffered channel of os.Signal
		sortslice.Analyzer,           // check the argument type of sort.Slice
		stdmethods.Analyzer,          // check signature of methods of well-known interfaces
		stringintconv.Analyzer,       // check for string(int) conversions
		structtag.Analyzer,           // check that struct field tags conform to reflect.StructTag.Get
		testinggoroutine.Analyzer,    // report calls to (*testing.T).Fatal from goroutines started by a test.
		tests.Analyzer,               // check for common mistaken usages of tests and examples
		unmarshal.Analyzer,           // report passing non-pointer or non-interface values to unmarshal
		unreachable.Analyzer,         // check for unreachable code
		unsafeptr.Analyzer,           // check for invalid conversions of uintptr to unsafe.Pointer
		unusedresult.Analyzer,        // check for unused results of calls to some functions
		unusedwrite.Analyzer,         // checks for unused writes
		usesgenerics.Analyzer,        // detect whether a package uses generics features
	}

	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name == "ST1000" { // Incorrect or missing package comment
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	for _, v := range simple.Analyzers {
		if v.Analyzer.Name == "S1000" { // Use plain channel send or receive instead of single-case select
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	for _, v := range quickfix.Analyzers {
		if v.Analyzer.Name == "QF1001" { // Apply De Morgan’s law
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	multichecker.Main(mychecks...)

}
