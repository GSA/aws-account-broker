package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/pivotal-cf/brokerapi"
)

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
		if awsStatusToBrokerInstanceState(awsStatusObj) != expected {
			t.Fatal("doesn't match")
		}
	}
}
