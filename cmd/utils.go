package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	userReadWrite     = 0600 // rw-|---|---
	userReadWriteExec = 0700 // rwx|---|---
)

var (
	//journalOut override-able by tests
	journalWriter   io.Writer = os.Stdout
	fatalExitWriter io.Writer = os.Stderr
)

//SetJournalWriter for testing
func SetJournalWriter(w io.Writer) {
	journalWriter = w
}

//SetFatalExitWriter for testing
func SetFatalExitWriter(w io.Writer) {
	fatalExitWriter = w
}

func fatalExit(err error, fmtArgs ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed:%v\n", err)
		os.Exit(1)
	}
}

func fatalExitf(err error, format string, args ...interface{}) {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(os.Stderr, "%s\n%v\n", msg, err)
		os.Exit(1)
	}
}

func journal(message string, args ...interface{}) {
	if viper.GetBool("verbose") {
		fmt.Fprintf(journalWriter, message, args...)
		fmt.Fprintln(journalWriter)
	}
}

func dumpFileFn(prefix string, fn func() []byte) {
	if !viper.GetBool("dump-work") {
		return
	}
	dumpFile(prefix, fn())
}

func dumpFile(prefix string, content []byte) {
	if !viper.GetBool("dump-work") {
		return
	}
	v := viper.GetBool("verbose")
	viper.Set("verbose", true)
	defer func() { viper.Set("verbose", v) }()

	tmpfile, err := ioutil.TempFile("", prefix)
	if err != nil {
		journal("failed to create temp file:%v", err)
		return
	}
	if _, err := tmpfile.Write(content); err != nil {
		journal("Unable to write to temp file: %v", err)
		return
	}
	if err := tmpfile.Close(); err != nil {
		journal("failed to close temp file: %v", err)
		return
	}
	journal("Dumped '%s'\n", tmpfile.Name())
	return
}

func fileMustExist(filePath string) {
	stat, err := os.Stat(filePath)
	if stat != nil || os.IsExist(err) {
		return
	}

	dir := filepath.Dir(filePath)
	_, err = os.Stat(dir)
	if !os.IsExist(err) {
		journal("Created missing directory '%s'", dir)
		os.MkdirAll(dir, userReadWriteExec)
	}

	f, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, userReadWrite)
	if err == nil {
		f.Close()
	} else {
		fatalExitf(err, "Failed to create required file", filePath)
	}
}

func homeDirectory() string {
	home := os.Getenv("USERPROFILE")
	if home == "" {
		home = os.Getenv("HOME")
	}
	return home
}

func bindFlags(root *cobra.Command, cmd string) {
	for _, c := range root.Commands() {
		if c.Name() == cmd {
			viper.BindPFlags(c.Flags())
			return
		}
	}
}
