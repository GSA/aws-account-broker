package main

import (
	"context"
	"testing"

	"code.cloudfoundry.org/lager"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
	"github.com/pivotal-cf/brokerapi"
	"github.com/stretchr/testify/assert"
)

type mockOrganizationsClient struct {
	organizationsiface.OrganizationsAPI
}

func (m mockOrganizationsClient) CreateAccount(input *organizations.CreateAccountInput) (*organizations.CreateAccountOutput, error) {
	state := organizations.CreateAccountStateInProgress

	output := organizations.CreateAccountOutput{
		CreateAccountStatus: &organizations.CreateAccountStatus{
			AccountName: input.AccountName,
			State:       &state,
		},
	}

	return &output, nil
}

func mockBroker() awsAccountBroker {
	svc := mockOrganizationsClient{}
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

func TestProvisionSuccess(t *testing.T) {
	broker := mockBroker()

	ctx := context.Background()
	details := brokerapi.ProvisionDetails{}

	spec, err := broker.Provision(ctx, "123", details, true)

	assert.NoError(t, err)
	assert.True(t, spec.IsAsync)
}

func TestProvisionSync(t *testing.T) {
	broker := mockBroker()

	ctx := context.Background()
	details := brokerapi.ProvisionDetails{}

	_, err := broker.Provision(ctx, "123", details, false)

	assert.Error(t, err)
}
