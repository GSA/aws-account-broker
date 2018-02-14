# AWS Account Service Broker

This is an API that [creates AWS (sub)accounts in an Organization](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_accounts_create.html). The broker conforms to the [Open Service Broker API](https://www.openservicebrokerapi.org/), so is compatible with [Cloud Foundry](https://cloudfoundry.org/), [OpenShift](https://www.openshift.org/), and [Kubernetes](http://kubernetes.io/).

**The use case:** you offer one of these platforms, with service brokers for specific databases, etc. Those service brokers will have limitations in the ways the underlying services can be configured, which is fine/desirable in many cases. For those cases where more flexibility is needed, though, this broker offers a trap door for users to get self-service access to full AWS accounts. Since those accounts are under the same Organization, they can be centrally configured with any needed policies, etc.

## Setup

Requires [Go](https://golang.org/). From the repository root, run the following. Note that email addresses for AWS accounts need to be unique, so `BASE_EMAIL` (in this example) will be turned into `something+<ID>@some.com`. This works in GMail, at the very least - you may need to confirm with your mail provider.

```sh
go install github.com/GSA/aws-account-broker
BASE_EMAIL=something@some.com aws-account-broker
```

You can confirm it's running from another terminal with:

```sh
curl --user user:pass http://localhost:8080/v2/catalog
```

### Development

```sh
cd $GOPATH/src/github.com/GSA/aws-account-broker

# 1. make edits

go install
BASE_EMAIL=something@some.com aws-account-broker

# 2. go back to 1
```
