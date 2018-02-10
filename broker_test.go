package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
	"github.com/pivotal-cf/brokerapi"
	"github.com/stretchr/testify/assert"
)

type mockOrganizationsClient struct {
	organizationsiface.OrganizationsAPI
}

func (m *mockOrganizationsClient) AcceptHandshake(input *organizations.AcceptHandshakeInput) (*organizations.AcceptHandshakeOutput, error) {
	// mock response/functionality
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
}
