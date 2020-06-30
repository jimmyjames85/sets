package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func readfile(fileloc string) (map[string]struct{}, error) {
	f, err := os.Open(fileloc)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readMap(f)
}

func readMap(r io.Reader) (map[string]struct{}, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	m := make(map[string]struct{})
	elems := strings.Split(string(b), "\n")
	for _, e := range elems {
		m[e] = struct{}{}
		// TODO if e already exists spit out to STDERR (detect DUPLICATE)
	}
	return m, nil
}

func writeMap(w io.Writer, m map[string]struct{}) error {
	for k := range m {
		_, err := fmt.Fprintf(w, "%s\n", k)
		if err != nil {
			return err
		}
	}
	return nil
}

func union(m1, m2 map[string]struct{}) map[string]struct{} {
	m := make(map[string]struct{})

	for k := range m1 {
		m[k] = struct{}{}
	}

	for k := range m2 {
		m[k] = struct{}{}
	}

	return m
}

// returns m1 - m2
func difference(m1, m2 map[string]struct{}) map[string]struct{} {
	m := make(map[string]struct{})

	for k := range m1 {
		m[k] = struct{}{}
	}

	for k := range m2 {
		delete(m, k)
	}

	return m
}

func intersection(m1, m2 map[string]struct{}) map[string]struct{} {
	if len(m2) < len(m1) {
		m1, m2 = m2, m1 // iterate through the smaller of the two maps
	}

	m := make(map[string]struct{})
	for k := range m1 {
		if _, ok := m2[k]; !ok {
			continue
		}
		m[k] = struct{}{}
	}
	return m
}

func usage() {
	// TODO //             --verify
	fmt.Printf(`%s
        Reads two inputs in as sets, one entry per line. Performs a set operation defined by OP and prints results to STDOUT.

Usage:
        %s file1 OP file2

OP:
	union			+, add, union
	difference		-, diff, sub, difference
	intersection		n, int, intersect, intersection

Flags:
        -h, --help              Displays this message

`, os.Args[0], os.Args[0])

}

var opts = struct {
	//// validate bool

	lvFileloc string
	rvFileloc string
	op        string
}{}

func exitf(code int, format string, a ...interface{}) {
	w := os.Stderr
	if code == 0 {
		w = os.Stdout
	}

	fmt.Fprintf(w, format, a...)
	os.Exit(code)
}

func mustParseOpts() {
	var operands []string
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			usage()
			os.Exit(0)
		// case "--validate":
		// 	opts.validate = true
		default:
			operands = append(operands, args[i])
		}
	}

	if len(operands) != 3 {
		exitf(-1, "expecting 3 arguments: got %d\n", len(operands))
	}

	opts.lvFileloc = operands[0]
	opts.op = operands[1]
	opts.rvFileloc = operands[2]

}

func main() {
	mustParseOpts()

	var op func(map[string]struct{}, map[string]struct{}) map[string]struct{}

	switch opts.op {
	case "difference", "-", "diff", "sub":
		op = difference
	case "union", "+", "add":
		op = union
	case "intersection", "n", "intersect", "int":
		op = intersection
	default:
		exitf(-1, "unknown op: %s\n", opts.op)
	}

	m1, err := readfile(opts.lvFileloc)
	if err != nil {
		exitf(-1, "file err: %s: %s\n", opts.lvFileloc, err.Error())
	}
	m2, err := readfile(opts.rvFileloc)
	if err != nil {
		exitf(-1, "file err: %s: %s\n", opts.rvFileloc, err.Error())
	}

	err = writeMap(os.Stdout, op(m1, m2))
	if err != nil {
		exitf(-1, "io err: %s\n", err.Error())
	}

}
