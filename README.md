## TCP Proxy

A simple tcpproxy in go. This is useful for proxying connections in and out of AWS VPCs e.g. if there's a database in 
EC2 Classic or another region and you only want to whitelist a single IP address, deploy this proxy onto a box and have
the machines in the VPC connect to it and have it forward those connections to the database.

This fulfils the same role as HAProxy, the difference being that this proxy will obey DNS TTLs. HAProxy only looks up
the domain name on startup, which stops DNS Failover from working.

### Config

The system supports a variety of backends for configuration, the included ones are:
#### static
   This backend is the *default* you configure it by passing in the arguments in the form `--connections [<port>:<url>:<port>]*`.

    tcpproxy --connections [<port>:<url>:<port>]*

#### dynamodb
This backend will poll dynamodb for configurations and kill and create connections as they get added or removed.
It can be enabled by setting the `--backend dynamodb` flag and passing in the `--proxy <name>`flag,
to indicate the proxies name. For example with blue-green deployment.

The `--dynamodb [tablename]` flag can be used to overide the default tablename of `classic-proxy`.

    tcpproxy --backend dynamodb --proxy <deployment name>

#### elasticache
This backend will automatically proxy between the machine and a random node in the elasticache cluster.
It can be enabled by passing the `--backend elasticache` flag. It interrogates the AWS api for all nodes
in the cluster and selects the node with the lowest identifier to proxy to. The `--elasticache-port <number>` indicates the
local port on which the proxy operates.

    tcpproxy --backend elasticache --elasticache-cluster-id <cluster id> --elasticache-port <localport>

### Running it

Run it as follows:

    tcpproxy --connections 8002:example.com:5432
    tcpproxy --backend dynamodb --proxy test
    tcpproxy --backend elasticache --elasticache-cluster-id my-redis-cluster --elasticache-port 6379

Debug can be enabled with the `--debug <level>` where `level` is an integer in the range `0...2`. Where 0 is no logging and 2 is maximum logging.

## Runnit it from docker

## Running

    docker brandnetworks/tcpproxy --connections 8002:example.com:5432

### Monitoring it

The tcpproxy exposes a /status HTTP endpoint on STATUS_ADDRESS (8001 in the example above).

It also exposes a `/connections` HTTP endpoint which returns a JSON blob with the full list of proxied connections.

### Releasing it.

The project includes a Dockerfile, allowing it to be built as a Docker image for deployment.

## Building

    docker build -t builder . && docker run builder | docker build -t eip-associate -
