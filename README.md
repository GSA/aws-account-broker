# AWS Account Service Broker [![CircleCI](https://circleci.com/gh/GSA/aws-account-broker.svg?style=shield)](https://circleci.com/gh/GSA/aws-account-broker)


This is an API that [creates AWS (sub)accounts in an Organization](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_accounts_create.html). The broker conforms to the [Open Service Broker API](https://www.openservicebrokerapi.org/), so is compatible with [Cloud Foundry](https://cloudfoundry.org/), [OpenShift](https://www.openshift.org/), and [Kubernetes](http://kubernetes.io/).

**The use case:** you offer one of these platforms, with service brokers for specific databases, etc. Those service brokers will have limitations in the ways the underlying services can be configured, which is fine/desirable in many cases. For those cases where more flexibility is needed, though, this broker offers a trap door for users to get self-service access to full AWS accounts. Since those accounts are under the same Organization, they can be centrally configured with any needed policies, etc.

## Setup

1. Install system dependencies.
    1. [Go](https://golang.org/)
    1. [Dep](https://golang.github.io/dep/docs/installation.html)
    1. [SQLite](https://www.sqlite.orga) - Proof of concept testing is using SQLite3 for persistence

1. Clone the repository.

    ```sh
    export GOPATH=~/go # or whatever go workspace you prefer
    mkdir -p $GOPATH/src/github.com/GSA
    cd $GOPATH/src/github.com/GSA
    git clone https://github.com/GSA/aws-account-broker.git
    cd aws-account-broker
    ```

1. Install Go package dependencies.

    ```sh
    dep ensure
    ```

1. Compile the broker.

    ```sh
    go build
    ```

1. Setup the database with Proof-of-Concept data.

    ```sh
    sqlite3 aws-account-broker.db < poc_data.sql  
    ```

1. Alternatively, you can inialize the database with just the schema with no data.

    ```sh
    sqlite3 aws-account-broker.db < schema.sql
    ```

1. Change any settings in the `config.toml` file.  See comments in file for
instructions.

2. You can override the database settings using a `DATABASE_URL` environment variable.
(Note: Only tested with sqlite3).

    ```sh
    export DATABASE_URL="sqlite3:/tmp/alt_database.db"
    ```

1. Pick a base email.
    * Email addresses for AWS accounts need to be unique, so `BASE_EMAIL` (below) will be turned into `something+<ID>@some.com`. This works in GMail, at the very least - you may need to confirm with your mail provider.
1. Run the broker.

    ```sh
    BASE_EMAIL=something@some.com ./aws-account-broker -user=<a username> -pass=<a password>
    ```

1. Confirm it's running and responding to requests. From another terminal, run:

    ```sh
    curl --user user:pass -H "X-Broker-API-Version: 2.13" http://localhost:8080/v2/catalog
    ```

    Make sure to use the user and pass that you specified in the run command above.

1. To create an account (also known as [Provisioning](https://github.com/openservicebrokerapi/servicebroker/blob/v2.13/spec.md#provisioning)):

    ```sh
    curl "http://user:pass@localhost:8080/v2/service_instances/<INSTANCE_ID>?accepts_incomplete=true" -d '{
      "service_id": "aws-account-broker",
      "plan_id": "IGNORED",
      "organization_guid": "IGNORED",
      "space_guid": "IGNORED"
    }' -X PUT -H "X-Broker-API-Version: 2.13" -H "Content-Type: application/json"
    ```

    Note that the `INSTANCE_ID` needs to be unique value for all the accounts in your Organization, as it's used to produce the unique email. The command also contains some dummy parameters - marked as `IGNORED` - which are required by the API spec but not yet used.

### Development

1. make edits
2. build and run

  ```sh
  go build
  BASE_EMAIL=something@some.com ./aws-account-broker -user=<a username> -pass=<a password>
  ```

3. CONTROL+C, then go back to 1

### Deploy to Cloud.gov

1. Initialize the database; For proof-of-concept testing, initialize with the
`poc_data.sql` file, otherwise use the `schema.sql` file.

    ```sh
    sqlite3 aws-account-broker.db < poc_data.sql
    ```

1. Log in to Cloud.gov and setup the command line.  [See documentation](https://cloud.gov/docs/getting-started/setup/#set-up-the-command-line)
1. For now, target your sandbox

    ```sh
    cf target -o <ORG> -s <SPACE>
    ```

1. Push the app. **Note:** The app will fail because required environment
variables are not set yet.

    ```sh
    cf push --random-route aws-account-broker
    ```

1. Set the environment variables from the command line or Cloud.gov dashboard.
environment variables:

    ```sh
    cf set-env aws-account-broker BASE_EMAIL ${BASE_EMAIL}
    cf set-env aws-account-broker BROKER_USER ${BROKER_USER}
    cf set-env aws-account-broker BROKER_PASSWORD ${BROKER_PASSWORD}
    cf set-env aws-account-broker AWS_ACCESS_KEY_ID ${AWS_ACCESS_KEY_ID}
    cf set-env aws-account-broker AWS_SECRET_ACCESS_KEY ${AWS_SECRET_ACCESS_KEY}
    ```

1. Restage the application

    ```sh
    cf restage aws-account-broker
    ```

1. Get the random route

    ```sh
    broker_url=$(cf app aws-account-broker | grep routes: | awk '{print $2}')
    ```

1. Check the service catalog

    ```sh
    curl -u ${BROKER_USER}:${BROKER_PASSWORD} -H "X-Broker-API-Version: 2.13" https://${broker_url}/v2/catalog
    ```

1. Check last operation

    ```sh
    curl -u ${BROKER_USER}:${BROKER_PASSWORD} -H "X-Broker-API-Version: 2.13" https://${broker_url}/v2/service_instances/gsa-devsecops-test4/last_operation
    ```

1. Register the broker

    ```sh
    cf create-service-broker aws-account-broker  \
    ${BROKER_USER} ${BROKER_PASSWORD} https://${broker_url} \
    --space-scoped
    ```

1. Display the broker in marketplace

    ```sh
    cf marketplace -s aws-account
    ```

1. Create an AWS account.

    ```sh
    cf create-service aws-account devsecops gsa-devsecops-test<#>
    ```

1. Check the status of the service.  Once the account is created, this will give you the account number.

    ```sh
    cf service gsa-devsecops-test<#>
    ```
