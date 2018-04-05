package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"code.cloudfoundry.org/lager"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/pivotal-cf/brokerapi"
	"github.com/stretchr/testify/assert"
)

// TODO: Add tests for assinging "available" accounts and releasing
//  assigned accounts.

type mockOrganizationsClient struct {
	organizationsiface.OrganizationsAPI
	createErr   error
	createState string
}

func (m mockOrganizationsClient) CreateAccount(input *organizations.CreateAccountInput) (*organizations.CreateAccountOutput, error) {
	id := "car-999999999999"
	output := organizations.CreateAccountOutput{
		CreateAccountStatus: &organizations.CreateAccountStatus{
			Id:          &id,
			AccountName: input.AccountName,
			State:       &m.createState,
		},
	}

	return &output, m.createErr
}

func (m mockOrganizationsClient) DescribeCreateAccountStatus(input *organizations.DescribeCreateAccountStatusInput) (*organizations.DescribeCreateAccountStatusOutput, error) {
	accountID := "999999999999"
	instanceID := "test-account"

	output := organizations.DescribeCreateAccountStatusOutput{
		CreateAccountStatus: &organizations.CreateAccountStatus{
			AccountId:   &accountID,
			AccountName: &instanceID,
			Id:          input.CreateAccountRequestId,
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

	dbFileName := filepath.Join(os.TempDir(), "broker_test.db")
	db, err := gorm.Open("sqlite3", dbFileName)
	if err != nil {
		logger.Fatal("startup", errors.New("failed to connect database"))
	}

	//Setup Database structure
	db.AutoMigrate(&serviceInstance{})

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

	spec, err := broker.Provision(ctx, "test-account", details, true)

	assert.NoError(t, err)
	assert.True(t, spec.IsAsync)
}

func TestProvisionFail(t *testing.T) {
	broker := mockBroker(errors.New("failed"), organizations.CreateAccountStateFailed)
	ctx := context.Background()
	details := brokerapi.ProvisionDetails{}

	_, err := broker.Provision(ctx, "test-account", details, true)

	assert.Error(t, err)
}

func TestProvisionSync(t *testing.T) {
	// values are arbitrary
	broker := mockBroker(nil, organizations.CreateAccountStateFailed)
	ctx := context.Background()
	details := brokerapi.ProvisionDetails{}

	_, err := broker.Provision(ctx, "test-account", details, false)

	assert.Error(t, err)
}

func TestLastOperation(t *testing.T) {
	broker := mockBroker(nil, organizations.CreateAccountStateSucceeded)
	ctx := context.Background()

	result, err := broker.LastOperation(ctx, "test-account", "")

	assert.NoError(t, err)
	assert.Equal(t, brokerapi.Succeeded, result.State)
}
