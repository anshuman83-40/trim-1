// Command trim keeps the lines of a context that a query needs.
//
//	trim --query "why does the build fail" build.log
//	cat build.log | trim --query "why does the build fail" --point balanced
//
// The trimmed context goes to stdout and a one-line summary to stderr.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	trim "github.com/caveman-ai/trim-1"
)

func main() {
	query := flag.String("query", "", "the current ask (required)")
	point := flag.String("point", trim.DefaultPoint, "operating point: conservative, balanced or aggressive")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: trim --query QUERY [--point POINT] [FILE]   (reads stdin without FILE)")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *query == "" || flag.NArg() > 1 {
		flag.Usage()
		os.Exit(2)
	}
	if _, ok := trim.Points[*point]; !ok {
		fmt.Fprintf(os.Stderr, "trim: unknown point %q\n", *point)
		os.Exit(2)
	}

	var (
		context []byte
		err     error
	)
	if flag.NArg() == 1 {
		context, err = os.ReadFile(flag.Arg(0))
	} else {
		context, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "trim:", err)
		os.Exit(1)
	}

	res := trim.Compress(string(context), *query, trim.Options{Point: *point})
	if _, err := os.Stdout.WriteString(res.Output); err != nil {
		fmt.Fprintln(os.Stderr, "trim:", err)
		os.Exit(1)
	}
	if res.Unchanged {
		fmt.Fprintf(os.Stderr, "trim: unchanged (%s)\n", res.Reason)
		return
	}
	fmt.Fprintf(os.Stderr, "trim: kept %d lines, removed %d, %d -> %d bytes\n",
		res.KeptLines, res.RemovedLines, len(context), len(res.Output))
}
