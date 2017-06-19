package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"golang.org/x/net/html/charset"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/xmlpath.v2"
	"log"
	"os"
	"os/exec"
  "strconv"
)

const (
	VERSION = "1.0.2"
)

func read_xml() *xmlpath.Node {
	path, err := exec.LookPath("passenger-status")
	if err != nil {
		// passenger-status not found in path
		if _, err := os.Stat("/usr/local/rvm/wrappers/default/passenger-status"); err == nil {
			// default rvm wrapper exists so use that!
			path = "/usr/local/rvm/wrappers/default/passenger-status"
		}
	}

	cmd := exec.Command(path, "--show=xml")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Stuff to handle the iso-8859-1 xml encoding
	// http://stackoverflow.com/a/32224438/606167
	decoder := xml.NewDecoder(stdout)
	decoder.CharsetReader = charset.NewReaderLabel

	xmlData, err := xmlpath.ParseDecoder(decoder)
	if err != nil {
		log.Fatal(err)
	}

	// Check version
	version_path := xmlpath.MustCompile("/info/@version")
	if version, ok := version_path.String(xmlData); !ok || (version != "3" && version != "2") {
		log.Fatal("Unsupported Passenger version (xml version ", version, ")")
	}

	return xmlData
}

func print_simple_selector(pathString string) {
	path := xmlpath.MustCompile(pathString)

	if value, ok := path.String(read_xml()); ok {
		fmt.Println(value)
	}
}

func print_selector_sum(pathString string) {
  // xmlpath doesn't support sum selector so we sum in Go

	path := xmlpath.MustCompile(pathString)

  sum := 0

  iter := path.Iter(read_xml())

  for iter.Next() {
    if intVal, err := strconv.Atoi(iter.Node().String()); err == nil {
      sum = sum + intVal
    }
  }

  fmt.Println(sum)
}

func print_app_groups_json() {
	groupPath := xmlpath.MustCompile("//supergroup/group")
	uuidPath := xmlpath.MustCompile("uuid")
	namePath := xmlpath.MustCompile("name")

	group_iter := groupPath.Iter(read_xml())

	var entries []map[string]string

	for group_iter.Next() {
		uuid, ok1 := uuidPath.String(group_iter.Node())
		name, ok2 := namePath.String(group_iter.Node())
    if ok1 && ok2 {
      entries = append(entries, map[string]string{"{#UUID}": uuid, "{#NAME}": name})
    }
	}

	data := map[string][]map[string]string{"data": entries}

	json, _ := json.Marshal(data)
	fmt.Println(string(json))
}

var (
	app     = kingpin.New("zabbix-passenger", "A utility to parse passenger-status output for usage with Zabbix")
	appPath = app.Flag("app", "UUID of application (leave out for global value)").String()

	appGroupsJson = app.Command("app-groups-json", "Get list of application groups in JSON format (for LLD)")
	queue         = app.Command("queue", "Get number of requests in queue, optionally specify app with --app")
	capacityUsed  = app.Command("capacity-used", "Get global capacity used, optionally specify app with --app")
	sessions      = app.Command("sessions", "GEt number of sessions, optionally specify app with --app")
)

func main() {
	app.Version(VERSION)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case appGroupsJson.FullCommand():
		print_app_groups_json()
	case queue.FullCommand():
		if *appPath != "" {
			print_simple_selector(fmt.Sprintf("//group[uuid='%v']/get_wait_list_size", *appPath))
		} else {
			print_simple_selector("//info/get_wait_list_size")
		}
	case capacityUsed.FullCommand():
		if *appPath != "" {
			print_simple_selector(fmt.Sprintf("//group[uuid='%v']/capacity_used", *appPath))
		} else {
			print_simple_selector("//info/capacity_used")
		}
  case sessions.FullCommand():
    if *appPath != "" {
			print_selector_sum(fmt.Sprintf("//group[uuid='%v']/processes/process/sessions", *appPath))
    } else {
			print_selector_sum(fmt.Sprintf("//group/processes/process/sessions", *appPath))
    }
	}
}
