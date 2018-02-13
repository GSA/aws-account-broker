package main

import (
	"context"
	"errors"

	"code.cloudfoundry.org/lager"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/pivotal-cf/brokerapi"
)

type notImplementedError struct{}

func (e notImplementedError) Error() string {
	return "Not implemented"
}

type awsAccountBroker struct {
	mgr    accountManager
	logger lager.Logger
}

func awsStatusToBrokerInstanceState(status organizations.CreateAccountStatus) brokerapi.LastOperationState {
	switch *status.State {
	case "IN_PROGRESS":
		return brokerapi.InProgress
	case "SUCCEEDED":
		return brokerapi.Succeeded
	}

	// fallback, including "FAILED"
	// https://docs.aws.amazon.com/organizations/latest/APIReference/API_ListCreateAccountStatus.html#API_ListCreateAccountStatus_RequestSyntax
	return brokerapi.Failed
}

func (b awsAccountBroker) Services(ctx context.Context) []brokerapi.Service {
	return []brokerapi.Service{
		brokerapi.Service{
			// TODO change to GUID?
			// https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#service-objects
			ID:          "aws-account-broker",
			Name:        "aws-account",
			Description: "Provisions AWS accounts under the organization",
			Bindable:    true,
			// TODO add plans
			Plans: []brokerapi.ServicePlan{},
			Metadata: &brokerapi.ServiceMetadata{
				DisplayName: "AWS account broker",
				// LongDescription:     "...",
				DocumentationUrl: "https://github.com/GSA/aws-account-broker",
				SupportUrl:       "https://github.com/GSA/aws-account-broker/issues/new",
				// ImageUrl:            "...",
				ProviderDisplayName: "The IDI team in GSA IT",
			},
			Tags: []string{
				"aws",
				"iaas",
			},
		},
	}
}

func (b awsAccountBroker) Provision(ctx context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	spec := brokerapi.ProvisionedServiceSpec{}
	if !asyncAllowed {
		return spec, errors.New("Accounts can only be created asynchronously")
	}

	// TODO don't hard-code these
	email := "aidan.feldman+broker@gsa.gov"
	_, err := b.mgr.CreateAccount("Service Broker account", email)
	if err != nil {
		return spec, err
	}
	// TODO use the result?

	b.logger.Info("Account created for " + email)

	spec.IsAsync = true
	// TODO set OperationData?
	return spec, nil
}

func (b awsAccountBroker) Deprovision(ctx context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	spec := brokerapi.DeprovisionServiceSpec{}
	return spec, errors.New("Not able to close accout through the API - see https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_accounts_close.html")
}

func (b awsAccountBroker) Bind(ctx context.Context, instanceID, bindingID string, details brokerapi.BindDetails) (brokerapi.Binding, error) {
	binding := brokerapi.Binding{}
	return binding, notImplementedError{}
}

func (b awsAccountBroker) Unbind(ctx context.Context, instanceID, bindingID string, details brokerapi.UnbindDetails) error {
	return notImplementedError{}
}

func (b awsAccountBroker) Update(ctx context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	spec := brokerapi.UpdateServiceSpec{}
	return spec, notImplementedError{}
}

func (b awsAccountBroker) LastOperation(ctx context.Context, instanceID, operationData string) (brokerapi.LastOperation, error) {
	awsStatus, err := b.mgr.GetAccountStatus()
	brokerState := awsStatusToBrokerInstanceState(*awsStatus)

	op := brokerapi.LastOperation{
		State:       brokerState,
		Description: awsStatus.GoString(),
	}
	return op, err
}

func NewAWSAccountBroker(logger lager.Logger) (awsAccountBroker, error) {
	mgr, err := newAccountManager()
	return awsAccountBroker{mgr, logger}, err
}
