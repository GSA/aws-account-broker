package main

import (
	"errors"
	"flag"
	"net/http"
	"net/url"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/BurntSushi/toml"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pivotal-cf/brokerapi"
)

var strUser = flag.String("user", "", "User name")
var strPass = flag.String("pass", "", "Password")

type tomlConfig struct {
	Server server   `toml:"server"`
	DB     database `toml:"database"`
}

type database struct {
	Provider string
	Args     string
}

type server struct {
	Host string
	Port string
}

func init() {
	flag.Parse()
	// Set up username and password
	// flag.StringVar(strUser, "user", "user", "User name")
	// flag.StringVar(strPass, "pass", "pass", "Password")
}

func main() {
	logger := lager.NewLogger("aws-account-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))
	logger.Info("Starting AWS account broker")

	var config tomlConfig
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		logger.Fatal("Problem reading config", err)
		return
	}

	baseEmail, found := os.LookupEnv("BASE_EMAIL")
	if !found {
		logger.Fatal("startup", errors.New("BASE_EMAIL not set"))
	}

	databaseURL, found := os.LookupEnv("DATABASE_URL")
	if found {
		u, err := url.Parse(databaseURL)
		if err != nil {
			logger.Fatal("Failed to parse DATABASE_URL", err)
		}
		config.DB.Provider = u.Scheme
		config.DB.Args = u.Path
	}
	db, err := gorm.Open(config.DB.Provider, config.DB.Args)
	if err != nil {
		logger.Fatal("Failed to connect database", err)
	}
	defer db.Close()

	broker, err := newAWSAccountBroker(baseEmail, logger, db)
	if err != nil {
		logger.Fatal("Problem starting broker", err)
	}

	creds := brokerapi.BrokerCredentials{
		Username: *strUser,
		Password: *strPass,
	}

	brokerAPI := brokerapi.New(broker, logger, creds)
	http.Handle("/", brokerAPI)

	host := config.Server.Host
	port := config.Server.Port
	origin := host + ":" + port
	logger.Info("Broker listening at " + origin)
	logger.Fatal("http-listen", http.ListenAndServe(origin, nil))
}
