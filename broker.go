package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/pivotal-cf/brokerapi"
)

type notImplementedError struct{}

func (e notImplementedError) Error() string {
	return "Not implemented"
}

type awsAccountBroker struct {
	// cache session, per
	// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/sessions.html
	sess *session.Session
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

func (b awsAccountBroker) Provision(ctx context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	spec := brokerapi.ProvisionedServiceSpec{}

	// follows this example
	// https://docs.aws.amazon.com/sdk-for-go/api/service/organizations/#example_Organizations_CreateAccount_shared00

	svc := organizations.New(b.sess)
	input := &organizations.CreateAccountInput{
		// TODO don't hard-code these
		AccountName: aws.String("Production Account"),
		Email:       aws.String("susan@example.com"),
	}

	result, err := svc.CreateAccount(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case organizations.ErrCodeAccessDeniedException:
				fmt.Println(organizations.ErrCodeAccessDeniedException, aerr.Error())
			case organizations.ErrCodeAWSOrganizationsNotInUseException:
				fmt.Println(organizations.ErrCodeAWSOrganizationsNotInUseException, aerr.Error())
			case organizations.ErrCodeConcurrentModificationException:
				fmt.Println(organizations.ErrCodeConcurrentModificationException, aerr.Error())
			case organizations.ErrCodeConstraintViolationException:
				fmt.Println(organizations.ErrCodeConstraintViolationException, aerr.Error())
			case organizations.ErrCodeInvalidInputException:
				fmt.Println(organizations.ErrCodeInvalidInputException, aerr.Error())
			case organizations.ErrCodeFinalizingOrganizationException:
				fmt.Println(organizations.ErrCodeFinalizingOrganizationException, aerr.Error())
			case organizations.ErrCodeServiceException:
				fmt.Println(organizations.ErrCodeServiceException, aerr.Error())
			case organizations.ErrCodeTooManyRequestsException:
				fmt.Println(organizations.ErrCodeTooManyRequestsException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

		return spec, err
	}

	// TODO do something with this
	fmt.Println(result)

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
	op := brokerapi.LastOperation{}
	return op, notImplementedError{}
}

func createBroker() brokerapi.ServiceBroker {
	return awsAccountBroker{}
}

func newAWSAccountBroker() awsAccountBroker {
	return awsAccountBroker{sess: session.New()}
}
