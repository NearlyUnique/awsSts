package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/vito/go-interact/interact"
)

var (
	verbose  bool
	dumpWork bool
)

func while(predicate func() bool) {
	for ok := true; ok; {
		ok = predicate()
	}
}
func exitErr(err error, msg string, args ...interface{}) {
	if err != nil {
		if len(args) > 0 {
			msg = fmt.Sprintf(msg, args...)
		}
		fmt.Fprintf(os.Stderr, msg+"\nErr:%v\n", err)
		os.Exit(1)
	}
}
func anyKeyOrQuit() bool {
	choice := "a"
	_ = interact.NewInteraction(
		"Auto refresh key...",
		interact.Choice{Display: "Quit", Value: "q"},
		interact.Choice{Display: "Refresh", Value: "r"},
	).Resolve(&choice)
	return choice != "q"
}
func log(message string, args ...interface{}) {
	if verbose {
		fmt.Printf(message, args...)
	}
}
func dumpFile(prefix string, content []byte) {
	if !dumpWork {
		return
	}
	tmpfile, err := ioutil.TempFile("", prefix)
	if err != nil {
		log("failed to create temp file:%v", err)
		return
	}
	if _, err := tmpfile.Write(content); err != nil {
		log("Unable to write to temp file: %v", err)
		return
	}
	if err := tmpfile.Close(); err != nil {
		log("failed to close temp file: %v", err)
		return
	}
	v := verbose
	verbose = true
	log("Dumped '%s'\n", tmpfile.Name())
	verbose = v
	return
}
