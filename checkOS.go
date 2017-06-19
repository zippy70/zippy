package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/mitchellh/ps"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	processName *string
	typof       *bool
)

func check(e error) {
	if e != nil {
		recover()
	}
}

func listProcs() []string {
	procFiles, err := ioutil.ReadDir("/proc")
	check(err)
	dirs := []string{}
	for _, fileInfo := range procFiles {
		isDir := fileInfo.IsDir()
		fileName := fileInfo.Name()
		matched, _ := regexp.MatchString("^[0-9]+$", fileName)
		if isDir == true && matched == true {
			dirs = append(dirs, fileName)
		}
	}
	return dirs
}

func countFds(dir string) int {
	fds, err := ioutil.ReadDir("/proc/" + dir + "/fd")
	check(err)
	return len(fds)
}

func countOpenFiles(dirs []string) int {
	openFiles := 0
	for _, dir := range dirs {
		openFiles = openFiles + countFds(dir)
	}
	return openFiles
}

func tcpConnections(name string) string {
	out := "ss"
	args := []string{"-l", "-p", "-n"}
	cmdout, err := exec.Command(out, args...).Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	x := string(cmdout)
	var count int
	temp := strings.Split(x, "\n")
	for index, line := range temp {
		if strings.Contains(line, name) {
			count++
		}
		_ = index
	}
	t := strconv.Itoa(int(count))
	return t

}
func PS() {
	ps, _ := ps.Processes()
	fmt.Println(ps[0].Executable())
	for pp, _ := range ps {
		fmt.Printf("%d %s\n", ps[pp].Pid(), ps[pp].Executable())
	}
}

func FindProcess(key string) (int, string, error) {
	pname := ""
	pid := 0
	err := errors.New("not found")
	ps, _ := ps.Processes()
	for i, _ := range ps {
		if ps[i].Executable() == key {
			pid = ps[i].Pid()
			pname = ps[i].Executable()
			err = nil
			break
		}
	}
	return pid, pname, err
}

func init() {

	processName = flag.String("ProcessName", "java", "process name eg:java")
	typof = flag.Bool("typeOf", false, "openfile = true, tcp connections = false")
}

func main() {
	flag.Parse()
	dirs := listProcs()
	openFiles := countOpenFiles(dirs)
	connection := tcpConnections(*processName)
	pid, s, a := FindProcess(*processName)
	pid1 := strconv.Itoa(pid)
	pid2 := []string{pid1}
	openFiles = countOpenFiles(pid2)
	_ = s
	_ = a
	if *typof == true {
		fmt.Printf("OK - Total number of open files is:%d | openfiles=%d\n", openFiles, openFiles)
		os.Exit(0)
	} else {
		fmt.Printf("OK - Total number of tcp connections %s is:%s | tcpconnections_%s=%s\n", *processName, connection, *processName, connection)
		os.Exit(0)
	}
}
