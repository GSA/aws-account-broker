package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"code.cloudfoundry.org/lager"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pivotal-cf/brokerapi"
)

type notImplementedError struct{}

func (e notImplementedError) Error() string {
	return "Not implemented"
}

type awsAccountBroker struct {
	mgr       accountManager
	baseEmail string
	logger    lager.Logger
	db        *gorm.DB
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

func generateUniqueEmail(baseEmail string, id string) string {
	emailParts := strings.SplitN(baseEmail, "@", 2)
	return fmt.Sprintf("%s+%s@%s", emailParts[0], id, emailParts[1])
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

type serviceInstance struct {
	gorm.Model
	InstanceId string
	RequestId  string
}

func (b awsAccountBroker) Provision(ctx context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	spec := brokerapi.ProvisionedServiceSpec{}
	if !asyncAllowed {
		return spec, brokerapi.ErrAsyncRequired
	}

	email := generateUniqueEmail(b.baseEmail, instanceID)
	createResult, err := b.mgr.CreateAccount(instanceID, email)
	if err != nil {
		return spec, err
	}

	b.logger.Info("Account created for " + email)
	var requestID = *createResult.CreateAccountStatus.Id
	b.logger.Info("RequestId: " + requestID)

	b.db.Create(&serviceInstance{InstanceId: instanceID, RequestId: requestID})

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

	var instance serviceInstance
	b.db.First(&instance, "instance_id = ?", instanceID)

	awsStatus, err := b.mgr.GetAccountStatus(instance.RequestId)
	brokerState := awsStatusToBrokerInstanceState(*awsStatus)

	op := brokerapi.LastOperation{
		State:       brokerState,
		Description: awsStatus.GoString(),
	}
	return op, err
}

func NewAWSAccountBroker(baseEmail string, logger lager.Logger, db *gorm.DB) (awsAccountBroker, error) {
	mgr, err := newAccountManager()
	return awsAccountBroker{mgr, baseEmail, logger, db}, err
}
