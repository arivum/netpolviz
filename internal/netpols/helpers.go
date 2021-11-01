package netpols

import (
	"flag"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// SetLogLevel sets global log level
func SetLogLevel(level string) {
	var (
		logLevel log.Level
		err      error
	)
	if logLevel, err = log.ParseLevel(level); err != nil {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(logLevel)
	}
}

// ParseFlags parses all command-line flags
func ParseFlags() *Config {
	var (
		cfg  = &Config{}
		home string
	)

	if home = HomeDir(); home != "" {
		flag.StringVar(&cfg.KubeConfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&cfg.KubeConfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.StringVar(&cfg.LogLevel, "loglevel", "info", "log level (one of info, warn, error, debug)")
	flag.StringVar(&cfg.ListenHost, "host", "", "Host to listen on")
	flag.StringVar(&cfg.ListenPort, "port", "11819", "Port to listen on")
	flag.StringVar(&cfg.WWWDirectory, "wwwdir", "assets/web", "Directory containing the WebUI")
	flag.Parse()
	return cfg
}

// HomeDir return the home directory of the current user
func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// LabelsMatch returns true if all givenLabels matches the ones needed
func LabelsMatch(neededLabels map[string]string, givenLabels map[string]string) bool {
	match := true
	for neededKey, neededValue := range neededLabels {
		if givenLabelValue, ok := givenLabels[neededKey]; ok {
			if givenLabelValue != neededValue {
				match = false
			}
		} else {
			match = false
		}
	}
	return match
}
