package elasticache

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/brandnetworks/tcpproxy/backends"
	"sort"
	"strconv"
	"fmt"
	"log"
)


func CreateElasticacheBackend(logLevel int, cacheClusterId string, localPort int, awsConfig *aws.Config) *ElasticacheBackend {
	return &ElasticacheBackend {
		logLevel: logLevel,
		localPort: strconv.Itoa(localPort),
		cacheClusterId: cacheClusterId,
		elasticache: elasticache.New(awsConfig),
	}
}

type ElasticacheBackend struct {
	logLevel int
	localPort string
	cacheClusterId string
	elasticache *elasticache.ElastiCache
}

func (d *ElasticacheBackend) GetProxyConfigurations() ([]backends.ConnectionConfig, error) {

	if d.logLevel > 0 {
		log.Println("Describing cluster", d.cacheClusterId)
	}

	// Paging shouldn't be a concern, as only 0 or 1 clusters should be returned by this call.
	clusters, err := d.elasticache.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{
		CacheClusterId:    aws.String(d.cacheClusterId),
		MaxRecords:        aws.Int64(100),
		ShowCacheNodeInfo: aws.Bool(true),
	})

	if err != nil {
		log.Println("Error describing cluster", d.cacheClusterId, err)
		return nil, err
	}

	var pollResults = make(map[int]backends.ConnectionConfig, 0)
	var nodeIDs = make([]int, 0)

	// Get all of the nodes for this cluster
	for _, cluster := range clusters.CacheClusters {
		for _, node := range cluster.CacheNodes {
			backend, err := backends.ParseConnection(fmt.Sprintf("%v:%s:%v", d.localPort, *node.Endpoint.Address, *node.Endpoint.Port))

			if err != nil {
				return nil, err
			}

			backend.Name = *cluster.CacheClusterId + "::" + *node.CacheNodeId

			id, err := strconv.Atoi(*node.CacheNodeId)

			if err != nil {
				return nil, err
			}

			pollResults[id] = *backend
			nodeIDs = append(nodeIDs, id)
		}
	}

	if d.logLevel > 0 {
		log.Println("Found", len(nodeIDs), "nodeIDs")
	}

	sort.Sort(sort.IntSlice(nodeIDs))

	// Select the lowest id no
	if len(nodeIDs) == 0 {
		return []backends.ConnectionConfig{}, nil
	} else {
		return []backends.ConnectionConfig{pollResults[nodeIDs[0]]}, nil
	}
}

func (b *ElasticacheBackend) IsPollable() bool {
	return true
}
