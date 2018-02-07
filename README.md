# AWS Account Service Broker

This is an API that [creates AWS (sub)accounts in an Organization](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_accounts_create.html). The broker conforms to the [Open Service Broker API](https://www.openservicebrokerapi.org/), so is compatible with [Cloud Foundry](https://cloudfoundry.org/), [OpenShift](https://www.openshift.org/), and [Kubernetes](http://kubernetes.io/).

**The use case:** you offer one of these platforms, with service brokers for specific databases, etc. Those service brokers will have limitations in the ways the underlying services can be configured, which is fine/desirable in many cases. For those cases where more flexibility is needed, though, this broker offers a trap door for users to get self-service access to full AWS accounts. Since those accounts are under the same Organization, they can be centrally configured with any needed policies, etc.
