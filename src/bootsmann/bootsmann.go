package main

import (
	"errors"
	"flag"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var log = logging.MustGetLogger("bootsmann")
var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.5s} %{id:03x}%{color:reset} %{message}",
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/bootsmann.cf", "Configuration file")
}

func main() {
	flag.Parse()

	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter)
	doProcess(configFile)
}

func doProcess(cf string) {
	buf, err := ioutil.ReadFile(cf)
	if err != nil {
		log.Error("No such config: %s", cf)
		os.Exit(1)
	}
	lines := strings.Split(string(buf), "\n")

	directives := make(map[string]map[string]string)
	vars := make(map[string]map[string]string)
	global := false
	var currentPath string
	sparks := 0
	sparkStop := make(chan bool)

	globalPath := "$global"

	for lineNum, l := range lines {
		if (len(l) > 2) && (l[1:len(l)-1] == globalPath) {
			if global {
				log.Critical("Second [$global] section on line %d", lineNum)
				os.Exit(1)
			}
			global = true
			currentPath = getPath(l, lineNum)
			directives[currentPath] = make(map[string]string)
			vars[currentPath] = make(map[string]string)
		} else if (len(l) > 2) && (string(l[0]) == "[") {
			// do processing of previous path section in goroutine
			if currentPath != globalPath {
				sparks += processPath(currentPath, sparkStop,
					vars["$global"], vars[currentPath],
					directives["$global"], directives[currentPath])
			}

			currentPath = getPath(l, lineNum)
			directives[currentPath] = make(map[string]string)
			vars[currentPath] = make(map[string]string)
		} else if (len(l) > 0) && (l[0] == '#') {
			// ignore comments
		} else {
			if vn, vv, ok := getVar(l, lineNum); ok {
				if vn[0] == '$' {
					// Special directive
					directives[currentPath][vn] = vv
				} else {
					vars[currentPath][vn] = vv
				}
			}
		}
	}
	sparks += processPath(currentPath, sparkStop,
		vars["$global"], vars[currentPath],
		directives["$global"], directives[currentPath])

	for i := 0; i < sparks; i++ {
		<-sparkStop
	}
}

func getPath(line string, _ int) (path string) {
	path = line[1 : len(line)-1]
	return
}

func getVar(line string, lineNum int) (vn string, vv string, ok bool) {
	out := strings.SplitN(line, "=", 2)
	if len(out) < 2 {
		if len(strings.Trim(line, " ")) > 0 {
			log.Error("Bad variable declaration on %d", lineNum)
		}
		return
	} else {
		vn = strings.Trim(out[0], " ")
		vv = strings.Trim(out[1], " ")
		ok = true
	}
	return
}

func mergeMaps(to, from map[string]string) (out map[string]string) {
	out = make(map[string]string)
	for k, v := range to {
		out[k] = v
	}
	for k, v := range from {
		out[k] = v
	}
	return
}

func processPath(path string, sparkStop chan bool,
	gvars map[string]string, lvars map[string]string,
	gdirs map[string]string, ldirs map[string]string) int {
	log.Info(path)
	vars := mergeMaps(gvars, lvars)
	dirs := mergeMaps(gdirs, ldirs)

	files, err := filepath.Glob(path)
	if err != nil {
		log.Warning("Bad pattern: %s", path)
	}
	if files == nil {
		return 0
	} else {
		for _, file := range files {
			go doFile(file, sparkStop, vars, dirs)
		}
		return len(files)
	}
}

func makeTemplate(dirs map[string]string, vn string) (tmpl string, err error) {
	t := "#{{ }}"
	if template, ok := dirs["$template"]; ok {
		t = template
	}
	ts := strings.Split(t, " ")
	if len(ts) != 2 {
		err = errors.New("Malformed template format: %s")
	} else {
		tmpl = strings.Join([]string{ts[0], vn, ts[1]}, "")
	}
	return
}

func doFile(file string, sparkStop chan bool, vars, dirs map[string]string) {
	log.Info(file)

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		log.Error("Cannot open file %s: %s", file, err)
	}

	s := string(buf)
	for vn, vv := range vars {
		tmpl, err := makeTemplate(dirs, vn)
		if err != nil {
			log.Error(err.Error())
		} else {
			s = strings.Replace(s, tmpl, vv, -1)
		}
	}

	err = ioutil.WriteFile(file, []byte(s), 0644)
	if err != nil {
		log.Error("Cannot write to file %s: %s", file, err)
	}

	sparkStop <- true
}
