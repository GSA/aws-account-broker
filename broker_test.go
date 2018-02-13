package main

import (
	"context"
	"testing"

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

func TestProvision(t *testing.T) {
	svc := mockOrganizationsClient{}
	mgr := accountManager{svc}
	broker := awsAccountBroker{mgr}

	ctx := context.Background()
	details := brokerapi.ProvisionDetails{}

	spec, err := broker.Provision(ctx, "123", details, true)

	assert.NoError(t, err)
	assert.True(t, spec.IsAsync)
}
