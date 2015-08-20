package elasticache

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/brandnetworks/tcpproxy/backends"
	"sort"
	"strconv"
	"fmt"
)


func CreateElasticacheBackend(cacheClusterID string, localPort int, awsConfig *aws.Config) *ElasticacheBackend {
	return &ElasticacheBackend {
		localPort: strconv.Itoa(localPort),
		cacheClusterID: cacheClusterID,
		elasticache: elasticache.New(awsConfig),
	}
}

type ElasticacheBackend struct {
	localPort string
	cacheClusterID string
	elasticache *elasticache.ElastiCache
}

func (d *ElasticacheBackend) GetProxyConfigurations() ([]backends.ConnectionConfig, error) {

	// Paging shouldn't be a concern, as only 0 or 1 clusters should be returned by this call.
	clusters, err := d.elasticache.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{
		CacheClusterID:    aws.String(d.cacheClusterID),
		MaxRecords:        aws.Int64(100),
		ShowCacheNodeInfo: aws.Bool(true),
	})

	if err != nil {
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

			backend.Name = *cluster.CacheClusterID + "::" + *node.CacheNodeID

			id, _ := strconv.Atoi(*node.CacheNodeID)

			pollResults[id] = *backend
			nodeIDs = append(nodeIDs, id)
		}
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