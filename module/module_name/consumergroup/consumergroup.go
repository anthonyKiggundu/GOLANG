package consumergroup

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/metricbeat/mb"
	"github.com/elastic/beats/metricbeat/module/kafka"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/Shopify/sarama"
	"errors"
)

// init registers the MetricSet with the central registry.
// The New method will be called after the setup of the module and before starting to fetch data
func init() {
	if err := mb.Registry.AddMetricSet("kafka_module", "consumergroup", New); err != nil {
		panic(err)
	}
}

var errFailQueryOffset = errors.New("operation failed")
// MetricSet type defines all fields of the MetricSet
// As a minimum it must inherit the mb.BaseMetricSet fields, but can be extended with
// additional entries. These variables can be used to persist data or configuration between
// multiple fetch calls.
type MetricSet struct {
	mb.BaseMetricSet
	broker *kafka.Broker	
	counter int
	topics []string 
}

type groupAssignment struct {
        clientID   string
        memberID   string
        clientHost string
}

// New create a new instance of the MetricSet
// Part of new is also setting up the configuration by processing additional
// configuration entries if needed.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {

	config := defaultConfig    //struct{}{}

	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}
	
	timeout := base.Module().Config().Timeout
	
	cfg := kafka.BrokerSettings{
		MatchID:     true,
		DialTimeout: timeout,
		ReadTimeout: timeout,
		ClientID:    config.ClientID,
		Retries:     config.Retries,
		Backoff:     config.Backoff,
		//TLS:         tls,
		//Username:    config.Username,
		//Password:    config.Password,	
		// consumer groups API requires at least 0.9.0.0
		Version: kafka.Version{String: "0.9.0.0"},
}
	return &MetricSet{
		BaseMetricSet: base,
		broker:        kafka.NewBroker(base.Host(), cfg),
		counter:       1,
	}, nil
}

func (m *MetricSet) connect() (*kafka.Broker, error) {
	err := m.broker.Connect()
	return m.broker, err
}

// Fetch methods implements the data gathering and data conversion to the right format
// It returns the event which is then forward to the output. In case of an error, a
// descriptive error must be returned.
func (m *MetricSet) Fetch() ([]common.MapStr, error) {

	b, err := m.connect()
	if err != nil {
		return nil, err
	}
	
	defer b.Close()
	
	topics, err := b.GetTopicsMetadata(m.topics...)
	if err != nil {
		return nil, err
	}
	events := []common.MapStr{}
	brokerInfo := common.MapStr{
		"id":      b.ID(),
		"address": b.AdvertisedAddr(),	
	}	
	
	for _, topic := range topics {
		logp.Err("fetch events for topic: ", topic.Name)
		evtTopic := common.MapStr{
			"name": topic.Name,
		}

		if topic.Err != 0 {
			evtTopic["error"] = common.MapStr{
				"code": topic.Err,
			}
		}
		
		for _, partition := range topic.Partitions {
			// partition offsets can be queried from leader only
			if b.ID() != partition.Leader {
				logp.Err("broker is not leader (broker=%v, leader=%v)", b.ID(), partition.Leader)
				continue
			}

			// collect offsets for all replicas
			for _, id := range partition.Replicas {

				// Get oldest and newest available offsets
				offOldest, offNewest, offOK, err := queryOffsetRange(b, id, topic.Name, partition.ID)
				
				if !offOK {
					if err == nil {
						err = errFailQueryOffset
					}

					logp.Err("Failed to query kafka partition (%v:%v) offsets: %v",
						topic.Name, partition.ID, err)
					continue
				}
				
				// computer consumer lag
				consumer_lag, err := consumerPartitionLag(b,partition.ID, topic.Name, offOldest, offNewest)
				partitionEvent := common.MapStr{
					"Partition_ID":             partition.ID,
					"Partition_Leader":         partition.Leader,
					"Partition_Replica":        id,
					"consumer_lag": 	consumer_lag,
					//"insync_replica": hasID(id, partition.Isr),
				}
				// logp.Err("Partition data %v",partitionEvent)
				if partition.Err != 0 {
					partitionEvent["error"] = common.MapStr{
						"code": partition.Err,
					}
				}
	 	
				event := common.MapStr{
					"counter": m.counter,
					"broker": brokerInfo,
					"Topic_name": evtTopic,
					"partition": partitionEvent,
					"offset": common.MapStr{
						"newest": offNewest,
						"oldest": offOldest,
				},
			}
		events = append(events, event)
		}
	}
   }
	
	m.counter++	
	return events, nil
}


// queryOffsetRange queries the broker for the oldest and the newest offsets in
// a kafka topics partition for a given replica.
func queryOffsetRange(b *kafka.Broker,replicaID int32,topic string,partition int32,) (int64, int64, bool, error) {
	oldest, err := b.PartitionOffset(replicaID, topic, partition, sarama.OffsetOldest)
	if err != nil {
		return -1, -1, false, err
	}

	newest, err := b.PartitionOffset(replicaID, topic, partition, sarama.OffsetNewest)
	if err != nil {
		return -1, -1, false, err
	}

	okOld := oldest != -1
	okNew := newest != -1
	return oldest, newest, okOld && okNew, nil
}

func hasID(id int32, lst []int32) bool {
	for _, other := range lst {
		if id == other {
			return true
		}
	}
return false
}

// Function to compute the consumer lag for each partition ID
func consumerPartitionLag(b *kafka.Broker,partition int32,topic string, offOld int64, offNew int64) (int64, error) {
	diffOffset := offNew - offOld
	return diffOffset, nil
}

	//if err := m.broker.Connect(); err != nil {
//		logp.Err("broker connect failed: %v", err)
//		return nil, err
//	}
//-------------
	
	//fmt.Printf("%v\n",slice(b.ListGroups())[1])
		//"groups": slice(b.ListGroups()),
		//"Metadata":  slice(b.GetMetadata()),
		//"Topics Metadata": slice(b.GetTopicsMetadata()),
		

// Read from MapStr string array and dump array
/*func slice(args ...interface{}) []interface{} {
    return args
} */

