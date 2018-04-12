package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
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

	dbFileName := "broker-test.db"
	_ = os.Remove(dbFileName)
	db, err := gorm.Open("sqlite3", dbFileName)
	if err != nil {
		logger.Fatal("startup", errors.New("failed to connect database"))
	}

	//Setup Database structure
	db.AutoMigrate(&serviceInstance{})
	//Add available account to Database
	db.Create(&serviceInstance{InstanceID: "available", RequestID: "car-111111111111"})
	//Add existing account for testing Deprovision
	db.Create(&serviceInstance{InstanceID: "test-deprovision", RequestID: "car-222222222222"})
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

	spec, err := broker.Provision(ctx, "test-reassign", details, true)

	//Decode the OperationData JSON
	var dat map[string]interface{}
	json.Unmarshal([]byte(spec.OperationData), &dat)

	assert.NoError(t, err)
	assert.True(t, spec.IsAsync)
	assert.Equal(t, "car-111111111111", dat["Id"].(string))

	spec, err = broker.Provision(ctx, "test-create", details, true)
	json.Unmarshal([]byte(spec.OperationData), &dat)

	assert.NoError(t, err)
	assert.True(t, spec.IsAsync)
	assert.Equal(t, "car-999999999999", dat["Id"].(string))
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

func TestDeprovision(t *testing.T) {
	broker := mockBroker(nil, organizations.CreateAccountStateSucceeded)
	ctx := context.Background()
	details := brokerapi.DeprovisionDetails{
		PlanID:    "2e8718e2-0991-48d2-b3be-514303bf762d",
		ServiceID: "1d138a29-ac8b-4360-be9b-db50867fee95",
	}

	_, err := broker.Deprovision(ctx, "test-deprovision", details, false)

	assert.NoError(t, err)
}
