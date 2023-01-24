# CRD API Demo: Employee Roster

This is a simple example of a CRD based application, with which a client
interacts by manipulating custom resources.

This application tracks employees in a company a aggregates their information
in a ConfigMap. users of the application can add and remove employees from
the rosterm or update their information, by manipulating `Employee` custom
resources. A reconciler, which is registered as an extension of the Kubernetes
API server, is responsible for reconciling the state of the roster in accordance
with the user's requirement as expressed by `Employee` objects.
