package main


import (
	"os"
	"log"
	"flag"
	"strings"
	"net/http"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/brandnetworks/tcpproxy/web"
	"github.com/brandnetworks/tcpproxy/proxy"
	"github.com/brandnetworks/tcpproxy/backends"
	"github.com/brandnetworks/tcpproxy/backends/static"
	"github.com/brandnetworks/tcpproxy/backends/dynamodb"
	"github.com/brandnetworks/tcpproxy/backends/elasticache"
)

type TcpProxyArgs struct {
	htmlEndpointBind *string
	logLevel *int
	awsRegion *string
	backend *string
	proxyName *string
	staticConnectionsConfigurationList *string
	dynamodbTableName *string
	elasticacheClusterID *string
	elasticacheClusterLocalPort *int
}

type TcpProxyError struct {
	msg    string // description of error
}

func (e *TcpProxyError) Error() string { return e.msg }
func NewTcpProxyError(msg string) (*TcpProxyError) {
	return &TcpProxyError{msg: msg}
}

func GetBackend(args TcpProxyArgs) (backends.ReadOnly, error) {
	switch strings.ToLower(*args.backend) {
	case "static":
		if *args.staticConnectionsConfigurationList != "" {
			log.Println("Proxying CLI configurations...")

			return static.CreateStaticBackend(*args.staticConnectionsConfigurationList)

		} else {
			return nil, NewTcpProxyError("Error: No connection configuations specified.")
		}

	case "dynamodb":
		if *args.proxyName != "" {
			log.Println("Proxying configurations from dynamodb...")

			awsConfig := &aws.Config{Region: aws.String(*args.awsRegion), MaxRetries: aws.Int(15)}

			backend := dynamodb.CreateDynamoDbBackend(*args.proxyName, *args.dynamodbTableName, awsConfig)

			return backend, nil

		} else {
			return nil, NewTcpProxyError("Error: No proxy name specified, please provide one for this backend.")
		}

	case "elasticache":
		if  *args.elasticacheClusterID != "" && *args.elasticacheClusterLocalPort > 0 {
			log.Println("Proxying configurations from elasticache...")

			awsConfig := &aws.Config{Region: aws.String(*args.awsRegion), MaxRetries: aws.Int(15)}

			backend := elasticache.CreateElasticacheBackend(*args.elasticacheClusterID, *args.elasticacheClusterLocalPort, awsConfig)

			return backend, nil

		} else {
			if *args.elasticacheClusterID == "" {
				return nil, NewTcpProxyError("Error: No Elasticache Cluster specified, please provide one for this backend.")
			}

			if *args.elasticacheClusterLocalPort < 0 {
				return nil, NewTcpProxyError("Error: No Elasticache localport specified or an invalid port was specified.")
			}

			// This should never be reached, but just incase someone adds more failure conditions in the future.
			return nil, NewTcpProxyError("Error: Unrecognised error initialising the elasticache backend.")
		}

	default:
		return nil, NewTcpProxyError("Error: unrecognised backend chosen.")
	}
}

func main() {

	args := TcpProxyArgs{}

	// General cli flags
	args.htmlEndpointBind = flag.String("status", ":8001", "Address:port used by the status endpoint")
	args.logLevel = flag.Int("debug", 0, "Enable debugging. Default disabled")

	// General backend flags
	args.awsRegion = flag.String("region", "us-east-1", "The AWS region in which the DynamoDB instance is located")
	args.backend = flag.String("backend", "static", "The backend to use of 'static' and 'dynamodb'")
	args.proxyName = flag.String("proxy", "", "This flag sets the name of the proxy")

	// Specific backend configuration flags
	args.staticConnectionsConfigurationList = flag.String("connections", "", "Comma separated list: srcPort:destHost:destPort,srcPort2:destHost2:destPort2")
	args.dynamodbTableName = flag.String("dynamodb", "classic-proxy", "This flag indicates the table on which the application operates, it must already exist")
	args.elasticacheClusterID = flag.String("elasticache-cluster-id", "", "This flag indicates the id of the Elasticache Cluster for which this program should proxy")
	args.elasticacheClusterLocalPort = flag.Int("elasticache-port", -1, "The local port from which the selected elasticache instance is proxied")

	flag.Parse()

	logLevel := *args.logLevel

	if *args.backend == "" {
		log.Fatal("Error: Blank backend specified.")
		flag.Usage()
		os.Exit(-1)
	}

	backend, err := GetBackend(args)

	if err != nil {
		log.Fatal(err)
		flag.Usage()
		os.Exit(-1)
	}

	tcpBackend := func(proxyInstance *proxy.Proxy) {
		proxy.RunTcpProxy(logLevel, proxyInstance.CreateChannel, proxyInstance.KillChannel, func() {
			mux := web.InitialiseEndpoints(logLevel, *args.proxyName, proxyInstance)
			http.ListenAndServe(*args.htmlEndpointBind, mux)
		})
	}

	err = proxy.RunProxy(backend, logLevel, tcpBackend)

	if err != nil {
		log.Fatal("Error fetching connections", nil)
	}

}
