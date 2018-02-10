package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
)

type accountManager struct {
	svc organizationsiface.OrganizationsAPI
}

func (am accountManager) CreateAccount(acctName string, email string) (*organizations.CreateAccountOutput, error) {
	input := &organizations.CreateAccountInput{
		AccountName: aws.String(acctName),
		Email:       aws.String(email),
	}

	result, err := am.svc.CreateAccount(input)
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

		return result, err
	}

	fmt.Println(result)

	return result, err
}

// TODO accept an ID - currently just gets the first one
func (am accountManager) GetAccountStatus() (*organizations.CreateAccountStatus, error) {
	input := &organizations.ListCreateAccountStatusInput{}
	input.SetMaxResults(1)

	result, err := am.svc.ListCreateAccountStatus(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case organizations.ErrCodeAccessDeniedException:
				fmt.Println(organizations.ErrCodeAccessDeniedException, aerr.Error())
			case organizations.ErrCodeAWSOrganizationsNotInUseException:
				fmt.Println(organizations.ErrCodeAWSOrganizationsNotInUseException, aerr.Error())
			case organizations.ErrCodeInvalidInputException:
				fmt.Println(organizations.ErrCodeInvalidInputException, aerr.Error())
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

		return nil, err
	}

	fmt.Println(result)

	return result.CreateAccountStatuses[0], nil
}

func newAccountManager() (accountManager, error) {
	// cache session, per
	// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/sessions.html
	sess, err := session.NewSession()
	svc := organizations.New(sess)
	return accountManager{svc}, err
}
