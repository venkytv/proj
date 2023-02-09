package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const Op = "/usr/local/bin/op"
const Tmux = "/opt/homebrew/bin/tmux"

func projDir(name string) string {
	return filepath.Join(os.Getenv("HOME"), "proj", name)
}

func buildProj(name string) {
	if err := os.Mkdir(projDir(name), 0755); err != nil {
		panic(err)
	}
	loadProj(name)
}

func opInject(dir string) []string {
	oprc := filepath.Join(dir, ".oprc")
	env := os.Environ()
	if _, err := os.Stat(oprc); err == nil {
		// Inject and load secrets from .oprc
		cmd := exec.Command(Op, "inject", "--in-file", oprc)
		var out strings.Builder
		cmd.Stdout = &out
		if err = cmd.Run(); err != nil {
			panic(fmt.Sprintf("error: op inject: %v", err))
		}

		scanner := bufio.NewScanner(strings.NewReader(out.String()))
		for scanner.Scan() {
			env = append(env, scanner.Text())
		}
	}

	return env
}

func loadProj(name string) {
	shell, ok := os.LookupEnv("SHELLS")
	if !ok {
		shell = "/bin/zsh"
	}

	dir := projDir(name)

	if err := os.Chdir(dir); err != nil {
		fmt.Fprintf(os.Stderr, "project does not exist: %s: %v\n", name, err)
		os.Exit(2)
	}

	cmd := exec.Command(Tmux, "rename-window", name)
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	// Inject secrets from 1Password
	env := opInject(dir)

	err := syscall.Exec(shell, []string{shell}, env)
	panic(fmt.Sprintf("%s: %s", shell, err))
}

func main() {
	var buildFlag bool

	flag.BoolVar(&buildFlag, "b", false, "build proj area")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-b] <name>\n", os.Args[0])
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	proj := flag.Args()[0]

	if buildFlag {
		buildProj(proj)
	} else {
		loadProj(proj)
	}
}