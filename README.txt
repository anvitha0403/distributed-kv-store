

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