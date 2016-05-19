# Stupid Simple Bus

The `ssbus` project creates a small and simple web service that uses
 long running http calls to maintain a bus style of communication.

# Installation

```
go install github.com/get-go/ssbus/...
```

# Usage

After starting the `ssbus` server on the default port, you can
 connect as many clients as you want to the stream endpoint.
 Any message posted to that endpoint will be sent to all
 attached clients.
 
```
# Server
ssbus

# Client
curl localhost:8675/_

# Send Message

curl -d '{"data":"Message Payload"}' localhost:8675/_
```
