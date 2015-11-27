package dynamodb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/brandnetworks/tcpproxy/backends"
)

func createProxy(tablename string, proxy_name string, proxy_configuration string) *dynamodb.PutItemInput {
	params := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{// Required
			"proxy_name": {S: aws.String(proxy_name)},
			"proxy_configuration": {S: aws.String(proxy_configuration)},
		},

		TableName: aws.String(tablename), // Required
	}

	return params
}

func deleteProxy(tablename string, proxy_name string, proxy_configuration string) *dynamodb.DeleteItemInput {
	params := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{// Required
			"proxy_name": {S: aws.String(proxy_name)},
			"proxy_configuration": {S: aws.String(proxy_configuration)},
		},
		TableName: aws.String(tablename), // Required
	}

	return params
}

func getProxiesWithName(tablename string, proxy_name string) *dynamodb.QueryInput {
	params := &dynamodb.QueryInput{
		TableName: aws.String(tablename),

		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":proxy_name": {S: aws.String(proxy_name)},
		},

		KeyConditionExpression: aws.String("proxy_name = :proxy_name"),

		// Pagination kicks in at 1MB which is a good limit
		// http://docs.aws.amazon.com/amazondynamodb/latest/developerguide/QueryAndScan.html#Pagination
//		Limit: aws.Int64(100),

		ProjectionExpression: aws.String("proxy_name, proxy_configuration"),
	}

	return params
}

func CreateDynamoDbBackend(proxy_name string, tablename string, awsConfig *aws.Config) *DynamoDbBackend {
	return &DynamoDbBackend{
		proxy_name: proxy_name,
		tablename: tablename,
		database: dynamodb.New(session.New(), awsConfig),
	}
}

type DynamoDbBackend struct {
	tablename string
	proxy_name string
	database  *dynamodb.DynamoDB
}

func (d *DynamoDbBackend) CreateProxyConfiguration(proxy_configuration string) error {
	_, err := d.database.PutItem(createProxy(d.tablename, d.proxy_name, proxy_configuration))

	return err
}

func (d *DynamoDbBackend) DeleteProxyConfiguration(proxy_configuration string) error {
	_, err := d.database.DeleteItem(deleteProxy(d.tablename, d.proxy_name, proxy_configuration))

	return err
}

func (d *DynamoDbBackend) GetProxyConfigurations() ([]backends.ConnectionConfig, error) {
	// TODO For now this is limited to 1MB of items before it hits pagination, which it doesnt implement yet.
	result, err := d.database.Query(getProxiesWithName(d.tablename, d.proxy_name))

	if err != nil {
		return nil, err
	}

	var connections = make([]backends.ConnectionConfig, len(result.Items))

	for i := range result.Items {
		connection, err := backends.ParseConnection(*result.Items[i]["proxy_configuration"].S)

		if err != nil {
			return nil, err
		}

		connections[i] = *connection
	}

	return connections, nil
}

func (b *DynamoDbBackend) IsPollable() bool {
	return true
}
