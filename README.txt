

TLS mutual authentication, also commonly referred to as two-way
authentication, in which both the server and the client validate the other’s
communication,

authorization is necessary when
you have a resource with shared access and varying levels of ownership.

CloudFlare4 wrote a toolkit called CFSSL for signing, verifying, and bundling
TLS certificates.

CFSSL has two tools we’ll need:
• cfssl to sign, verify, and bundle TLS certificates and output the results as JSON.
• cfssljson to take that JSON output and split them into separate key, cer
tificate, CSR, and bundle files.


observablilty - how well you can 
Observability is a measure of how well we understand our system’s inter
nals—its behavior and state—from its external outputs.

metrics 
logs 
traces
sla , slo, sli 

reduce - storage ( and resolution)

histogram , counter, gauges

latecncy, saturation,traffic, errors


races capture request lifecycles and let you track requests as they flow
through your system. 

using uber zap library 
opentelemetry only supports distributed traces, - no metrics and logging at this point 



curl -X PUT http://localhost:8080/kv      -H "Content-Type: application/json"      -d '{"key":"name","value":"Anvitha"}' -vvv

 curl -i http://localhost:8080/kv?key=Name



 Standard gRPC Status CodesgRPC CodeNameTypical HTTP MatchCommon Cause / Usage0OK200 OKThe operation completed successfully.1CANCELLED499 Client ClosedThe client explicitly cancelled the request.2UNKNOWN500 Internal ErrorServer application crashed or threw an unhandled exception.3INVALID_ARGUMENT400 Bad RequestClient passed bad input, like failing validation.4DEADLINE_EXCEEDED504 Gateway TimeoutThe operation timed out before receiving a response.5NOT_FOUND404 Not FoundThe requested entity or resource was not found.6ALREADY_EXISTS409 ConflictAttempted to create an entity that is already present.7PERMISSION_DENIED403 ForbiddenClient lacks permissions for this specific action.8RESOURCE_EXHAUSTED429 Too Many RequestsRate limits hit or server has run out of disk space.9FAILED_PRECONDITION400 Bad RequestSystem state prevents execution (e.g., deleting a non-empty folder).10ABORTED409 ConflictConcurrency issues, like transaction aborts.11OUT_OF_RANGE400 Bad RequestOperation attempted past valid range (e.g., reading past EOF).12UNIMPLEMENTED501 Not ImplementedMethod not recognized or not enabled on the server.13INTERNAL500 Internal ErrorCritical internal framework failure or broken system invariant.14UNAVAILABLE503 Service UnavailableServer shutting down or connection dropped before processing.15DATA_LOSS500 Internal ErrorUnrecoverable data corruption or loss.16UNAUTHENTICATED401 UnauthorizedRequest lacks valid authentication credentials.
 

 An interface value in Go is actually a two‑word struct under the hood:

A pointer to the type information (what concrete type is stored).

A pointer to the actual data (the value itself).

👉 So: passing a struct to an interface is pass by value (copy of the struct), while passing a struct pointer to an interface is pass by reference (the interface holds the pointer).




apt-get update && apt-get install -y iproute2 net-tools
 go run main.go --config-file=/var/run/kvstore/config.yaml



CONFIG_DIR=/src/certs/.kvstore go run getServers.go --addr kvstore.default.svc.cluster.local:8400

go run main.go   --rpc-port=8400   --server-tls-ca-file=/src/certs/.kvstore/ca.pem   --server-tls-cert-file=/src/certs/.kvstore/server.pem   --server-tls-key-file=/src/certs/.kvstore/server-key.pem --bootstrap=true --bind-addr="kvstore-0.kvstore.default.svc.cluster.local:8401" --rpc-port=8400 --data-dir="/var/run/kvstore"