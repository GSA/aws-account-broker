package main

import (
	"errors"
	"flag"
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi"
)

var strUser = flag.String("user", "", "User name")
var strPass = flag.String("pass", "", "Password")

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

	baseEmail, found := os.LookupEnv("BASE_EMAIL")
	if !found {
		logger.Fatal("startup", errors.New("BASE_EMAIL not set"))
	}

	broker, err := NewAWSAccountBroker(baseEmail, logger)
	if err != nil {
		logger.Fatal("Problem starting broker", err)
	}

	creds := brokerapi.BrokerCredentials{
		Username: *strUser,
		Password: *strPass,
	}

	brokerAPI := brokerapi.New(broker, logger, creds)
	http.Handle("/", brokerAPI)

	host := "127.0.0.1"
	port := "8080"
	origin := host + ":" + port
	logger.Info("Broker listening at " + origin)
	logger.Fatal("http-listen", http.ListenAndServe(origin, nil))
}
