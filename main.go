package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
)

var (
	bzip2Flag = flag.Bool("b", false, "compress bzip2")
	helpFlag  = flag.Bool("h", false, "show help")
)

var helpMessage = `
 Usage: git tarball [options] [commitish]
  -b   create tar.bz2
  -h   show this help

 Examples:
  git tarball          # create tarball with HEAD
  git tarball 0f7fea7  # create tarball with 0f7fea7
  git tarball -b       # create tar.bz2

`

func main() {
	os.Exit(_main())
}

func execCmd(c ...string) *exec.Cmd {
	cmd := exec.Command(c[0], c[1:]...)
	cmd.Stderr = os.Stderr
	return cmd
}

func _main() int {
	flag.Parse()
	if *helpFlag {
		fmt.Print(helpMessage)
		return 1
	}
	commitish := "HEAD"
	if len(flag.Args()) > 1 {
		commitish = flag.Args()[0]
	}

	out, err := execCmd("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return 1
	}
	out = out[:len(out)-1]
	top := path.Base(string(out))

	git := execCmd("git", "archive", "--format", "tar", "--prefix", top+"/", commitish)

	compress := execCmd("gzip")
	outFilename := top + ".tar.gz"
	if *bzip2Flag {
		compress = execCmd("bzip2")
		outFilename = top + ".tar.bz2"
	}
	compress.Stdin, err = git.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	io, err := os.Create(outFilename)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	defer io.Close()
	compress.Stdout = io
	cmds := []*exec.Cmd{git, compress}
	for _, c := range cmds {
		if err = c.Start(); err != nil {
			fmt.Println(err)
			return 1
		}
	}
	for _, c := range cmds {
		if err = c.Wait(); err != nil {
			fmt.Println(err)
			return 1
		}
	}
	return 0
}
