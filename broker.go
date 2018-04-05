package main

import (
	"context"
	"encoding/json"
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
			ID:          "1d138a29-ac8b-4360-be9b-db50867fee95",
			Name:        "aws-account",
			Description: "Provisions AWS accounts under the organization",
			Bindable:    true,
			// TODO add plans
			Plans: []brokerapi.ServicePlan{
				brokerapi.ServicePlan{
					ID:          "2e8718e2-0991-48d2-b3be-514303bf762d",
					Name:        "devsecops",
					Description: "Provisions AWS accounts under the organization",
				},
			},
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
	InstanceID string
	RequestID  string
}

func (b awsAccountBroker) Provision(ctx context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	spec := brokerapi.ProvisionedServiceSpec{}
	if !asyncAllowed {
		return spec, brokerapi.ErrAsyncRequired
	}

	email := generateUniqueEmail(b.baseEmail, instanceID)
	createResult, err := b.mgr.CreateAccount(instanceID, email, b.db)
	if err != nil {
		return spec, err
	}

	b.logger.Info("Account created for " + email)

	spec.IsAsync = true
	status, _ := json.Marshal(createResult.CreateAccountStatus)
	spec.OperationData = string(status)
	return spec, nil
}

func (b awsAccountBroker) Deprovision(ctx context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	spec := brokerapi.DeprovisionServiceSpec{}

	var instance serviceInstance
	err := b.db.First(&instance, "instance_id = ?", instanceID).Error

	if err != nil {
		instance.InstanceID = "unassigned"
		b.db.Save(&instance)
	}

	return spec, err
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

	awsStatus, err := b.mgr.GetAccountStatus(instance.RequestID)
	brokerState := awsStatusToBrokerInstanceState(*awsStatus)

	op := brokerapi.LastOperation{
		State:       brokerState,
		Description: awsStatus.GoString(),
	}
	return op, err
}

func newAWSAccountBroker(baseEmail string, logger lager.Logger, db *gorm.DB) (awsAccountBroker, error) {
	mgr, err := newAccountManager()
	return awsAccountBroker{mgr, baseEmail, logger, db}, err
}
