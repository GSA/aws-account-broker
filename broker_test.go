package main

import (
	"context"
	"errors"
	"testing"

	"code.cloudfoundry.org/lager"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
	"github.com/pivotal-cf/brokerapi"
	"github.com/stretchr/testify/assert"
)

type mockOrganizationsClient struct {
	organizationsiface.OrganizationsAPI
	createErr   error
	createState string
}

func (m mockOrganizationsClient) CreateAccount(input *organizations.CreateAccountInput) (*organizations.CreateAccountOutput, error) {
	output := organizations.CreateAccountOutput{
		CreateAccountStatus: &organizations.CreateAccountStatus{
			AccountName: input.AccountName,
			State:       &m.createState,
		},
	}

	return &output, m.createErr
}

func mockBroker(createErr error, createState string) awsAccountBroker {
	svc := mockOrganizationsClient{
		createErr:   createErr,
		createState: createState,
	}
	mgr := accountManager{svc}
	logger := lager.NewLogger("test")
	return awsAccountBroker{mgr, logger}
}

func TestAWSStatusToBrokerInstanceState(t *testing.T) {
	cases := map[string]brokerapi.LastOperationState{
		"IN_PROGRESS": brokerapi.InProgress,
		"SUCCEEDED":   brokerapi.Succeeded,
		"FAILED":      brokerapi.Failed,
		"foo":         brokerapi.Failed,
	}

	for awsStatus, expected := range cases {
		awsStatusObj := organizations.CreateAccountStatus{
			State: &awsStatus,
		}
		actual := awsStatusToBrokerInstanceState(awsStatusObj)
		assert.Equal(t, expected, actual)
	}
}

func TestServices(t *testing.T) {
	// values are arbitrary here
	broker := mockBroker(nil, organizations.CreateAccountStateInProgress)
	ctx := context.Background()

	services := broker.Services(ctx)

	assert.Len(t, services, 1)
}

func TestProvisionSuccess(t *testing.T) {
	broker := mockBroker(nil, organizations.CreateAccountStateInProgress)
	ctx := context.Background()
	details := brokerapi.ProvisionDetails{}

	spec, err := broker.Provision(ctx, "123", details, true)

	assert.NoError(t, err)
	assert.True(t, spec.IsAsync)
}

func TestProvisionFail(t *testing.T) {
	broker := mockBroker(errors.New("failed"), organizations.CreateAccountStateFailed)
	ctx := context.Background()
	details := brokerapi.ProvisionDetails{}

	_, err := broker.Provision(ctx, "123", details, true)

	assert.Error(t, err)
}

func TestProvisionSync(t *testing.T) {
	// values are arbitrary
	broker := mockBroker(nil, organizations.CreateAccountStateFailed)
	ctx := context.Background()
	details := brokerapi.ProvisionDetails{}

	_, err := broker.Provision(ctx, "123", details, false)

	assert.Error(t, err)
}
