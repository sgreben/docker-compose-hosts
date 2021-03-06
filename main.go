package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v2"
)

var (
	appName = "docker-compose-hosts"
	version = "SNAPSHOT"
	flags   struct {
		Quiet            bool
		Verbose          bool
		Version          bool
		Parallel         bool
		File             string
		RemoteSocketAddr string
	}
	config            ConfigV1
	configVersionsMap = map[string]bool{
		"0":   true,
		"0.1": true,
		"1.0": true,
	}
	configVersions = func() (out []string) {
		for v := range configVersionsMap {
			out = append(out, v)
		}
		sort.Strings(out)
		return
	}()
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("[%s] ", appName))
	flags.File = "docker-compose-hosts.yml"
	flags.RemoteSocketAddr = "unix:///var/run/docker.sock"
	flag.BoolVar(&flags.Quiet, "logs-off", flags.Quiet, "disable all logging")
	flag.BoolVar(&flags.Parallel, "parallel", flags.Parallel, "run commands in parallel")
	flag.BoolVar(&flags.Quiet, "q", flags.Quiet, "(alias for -logs-off)")
	flag.BoolVar(&flags.Verbose, "logs-verbose", flags.Verbose, "enable extra logging")
	flag.BoolVar(&flags.Verbose, "v", flags.Verbose, "(alias for -logs-verbose)")
	flag.BoolVar(&flags.Version, "version", flags.Version, "print version and exit")
	flag.StringVar(&flags.File, "file", flags.File, "specify an alternate compose-hosts file")
	flag.StringVar(&flags.File, "f", flags.File, "(alias for -file)")

	flag.Parse()
	if flags.Quiet {
		log.SetOutput(ioutil.Discard)
	}
}

func main() {
	if flags.Version {
		fmt.Println(version)
		return
	}
	f, err := os.Open(flags.File)
	if err != nil {
		log.Fatalf("open compose-hosts file %q: %v", flags.File, err)
	}
	dec := yaml.NewDecoder(f)
	dec.SetStrict(true)
	if err := dec.Decode(&config); err != nil {
		log.Fatalf("parse compose-hosts file %q: %v", flags.File, err)
	}
	if config.Version == "" {
		log.Fatalf("a version must be specified in %q. valid choices: %q", flags.File, configVersions)
	}
	if !configVersionsMap[config.Version] {
		log.Fatalf("invalid version %q specified in %q. valid choices: %q", config.Version, flags.File, configVersions)
	}
	wd := filepath.Dir(flags.File)
	if err := os.Chdir(wd); err != nil {
		log.Fatalf("set working directory to %q: %v", wd, err)
	}
	if flags.Parallel {
		if err := config.CommandParallel(flag.Args()); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := config.Command(flag.Args()); err != nil {
		log.Fatal(err)
	}
}
