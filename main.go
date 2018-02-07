package main

import (
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi"
)

func main() {
	broker := awsAccountBroker{}

	logger := lager.NewLogger("aws-account-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))

	logger.Info("Starting AWS account broker")

	creds := brokerapi.BrokerCredentials{
		// TODO specify these another way
		Username: "user",
		Password: "pass",
	}

	brokerAPI := brokerapi.New(broker, logger, creds)
	http.Handle("/", brokerAPI)

	host := "127.0.0.1"
	port := "8080"
	logger.Fatal("http-listen", http.ListenAndServe(host+":"+port, nil))
}
