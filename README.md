# AWS Account Service Broker

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
    sqlite3 /tmp/aws-account-broker.db < poc_data.sql  
    ```

1. Alternatively, you can inialize the database with just the schema with no data.

    ```sh
    sqlite3 /tmp/aws-account-broker.db < schema.sql
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
    curl --user user:pass http://localhost:8080/v2/catalog
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
