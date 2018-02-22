package main

import (
	"context"
	"errors"
	"testing"

	"code.cloudfoundry.org/lager"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pivotal-cf/brokerapi"
	"github.com/stretchr/testify/assert"
)

type mockOrganizationsClient struct {
	organizationsiface.OrganizationsAPI
	Id          string
	createErr   error
	createState string
}

func (m mockOrganizationsClient) CreateAccount(input *organizations.CreateAccountInput) (*organizations.CreateAccountOutput, error) {
	output := organizations.CreateAccountOutput{
		CreateAccountStatus: &organizations.CreateAccountStatus{
			Id:          &m.Id,
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

	baseEmail := "foo@bar.com"
	logger := lager.NewLogger("test")

	db, err := gorm.Open("sqlite3", "broker_test.db")
	if err != nil {
		logger.Fatal("startup", errors.New("failed to connect database"))
	}

	return awsAccountBroker{mgr, baseEmail, logger, db}
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

func TestGenerateUniqueEmail(t *testing.T) {
	assert.Equal(t, generateUniqueEmail("foo@bar.com", "1"), "foo+1@bar.com")
	assert.Equal(t, generateUniqueEmail("foo.bar@baz.com", "1"), "foo.bar+1@baz.com")
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
