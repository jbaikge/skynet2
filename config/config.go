package config

import (
	"flag"
	"github.com/robfig/config"
	"github.com/skynetservices/skynet2/log"
	"os"
	"time"
)

var defaultConfigFiles = []string{
	"/etc/skynet/skynet.conf",
	"./skynet.conf",
}

var configFile string
var uuid string
var conf *config.Config

func init() {
	flagset := flag.NewFlagSet("config", flag.ContinueOnError)
	flagset.StringVar(&configFile, "config", "", "Config File")
	flagset.StringVar(&uuid, "uuid", "", "uuid")

	args, _ := SplitFlagsetFromArgs(flagset, os.Args[1:])
	flagset.Parse(args)

	// Ensure we have a UUID
	if uuid == "" {
		uuid = NewUUID()
	}

	if configFile == "" {
		for _, f := range defaultConfigFiles {
			if _, err := os.Stat(f); err == nil {
				configFile = f
				break
			}
		}
	}

	if configFile == "" {
		log.Println(log.ERROR, "Failed to find config file")
		conf = config.NewDefault()
		return
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Println(log.ERROR, "Config file does not exist", err)
		conf = config.NewDefault()
		return
	}

	var err error
	if conf, err = config.ReadDefault(configFile); err != nil {
		conf = config.NewDefault()
		log.Fatal(err)
	}
}

func String(service, version, option string) (string, error) {
	s := getSection(service, version)

	return conf.String(s, option)
}

func Bool(service, version, option string) (bool, error) {
	s := getSection(service, version)

	return conf.Bool(s, option)
}

func Int(service, version, option string) (int, error) {
	s := getSection(service, version)

	return conf.Int(s, option)
}

func RawString(service, version, option string) (string, error) {
	s := getSection(service, version)

	return conf.RawString(s, option)
}

func RawStringDefault(option string) (string, error) {
	return conf.RawStringDefault(option)
}

func getSection(service, version string) string {
	s := service + "-" + version
	if conf.HasSection(s) {
		return s
	}

	return service
}

func getFlagName(f string) (name string) {
	if f[0] == '-' {
		minusCount := 1

		if f[1] == '-' {
			minusCount++
		}

		f = f[minusCount:]

		for i := 0; i < len(f); i++ {
			if f[i] == '=' || f[i] == ' ' {
				break
			}

			name += string(f[i])
		}
	}

	return
}

func UUID() string {
	return uuid
}

func SplitFlagsetFromArgs(flagset *flag.FlagSet, args []string) (flagsetArgs []string, additionalArgs []string) {
	for _, f := range args {
		if flagset.Lookup(getFlagName(f)) != nil {
			flagsetArgs = append(flagsetArgs, f)
		} else {
			additionalArgs = append(additionalArgs, f)
		}
	}

	return
}

// Client
type ClientConfig struct {
	Host                      string
	Region                    string
	IdleConnectionsToInstance int
	MaxConnectionsToInstance  int
	IdleTimeout               time.Duration
}

func GetDefaultEnvVar(name, def string) (v string) {
	v = os.Getenv(name)
	if v == "" {
		v = def
	}
	return
}

func FlagsForClient(ccfg *ClientConfig, flagset *flag.FlagSet) {
	flagset.DurationVar(&ccfg.IdleTimeout, "timeout", DefaultIdleTimeout, "amount of idle time before timeout")
	flagset.IntVar(&ccfg.IdleConnectionsToInstance, "maxidle", DefaultIdleConnectionsToInstance, "maximum number of idle connections to a particular instance")
	flagset.IntVar(&ccfg.MaxConnectionsToInstance, "maxconns", DefaultMaxConnectionsToInstance, "maximum number of concurrent connections to a particular instance")
	flagset.StringVar(&ccfg.Region, "region", GetDefaultEnvVar("SKYNET_REGION", DefaultRegion), "region client is located in")
	flagset.StringVar(&ccfg.Host, "host", GetDefaultEnvVar("SKYNET_HOST", DefaultRegion), "host client is located in")
}

func GetClientConfig() (config *ClientConfig, args []string) {
	return GetClientConfigFromFlags(os.Args[1:])
}

func GetClientConfigFromFlags(argv []string) (config *ClientConfig, args []string) {
	config = &ClientConfig{}

	flagset := flag.NewFlagSet("config", flag.ContinueOnError)

	FlagsForClient(config, flagset)

	err := flagset.Parse(argv)

	args = flagset.Args()
	if err == flag.ErrHelp {
		// -help was given, pass it on to caller who
		// may decide to quit instead of continuing
		args = append(args, "-help")
	}

	return
}