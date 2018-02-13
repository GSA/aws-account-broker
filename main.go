package main

import (
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi"
)

func makeLogger() lager.Logger {
	logger := lager.NewLogger("aws-account-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))
	return logger
}

func main() {
	logger := makeLogger()
	logger.Info("Starting AWS account broker")

	broker, err := NewAWSAccountBroker(logger)
	if err != nil {
		logger.Fatal("Problem starting broker", err)
	}

	creds := brokerapi.BrokerCredentials{
		// TODO specify these another way
		Username: "user",
		Password: "pass",
	}

	brokerAPI := brokerapi.New(broker, logger, creds)
	http.Handle("/", brokerAPI)

	host := "127.0.0.1"
	port := "8080"
	origin := host + ":" + port
	logger.Info("Broker listening at " + origin)
	logger.Fatal("http-listen", http.ListenAndServe(origin, nil))
}
