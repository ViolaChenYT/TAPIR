package common

import (
	"math"
)

type ReplicaAddress struct {
	Host string
	Port string
}

func NewReplicaAddress(host, port string) *ReplicaAddress {
	return &ReplicaAddress{Host: host, Port: port}
}

func (addr *ReplicaAddress) SpecificString() string {
	return addr.Host + ":" + addr.Port
}

type ClientConfiguration struct {
	TAPIR_ID         int
	IR_ID            int
	ClosestReplicaID int
}

func NewClientConfiguration(tapir_id, ir_id, closest_replica_id int) *ClientConfiguration {
	return &ClientConfiguration{TAPIR_ID: tapir_id, IR_ID: ir_id, ClosestReplicaID: closest_replica_id}
}

type Configuration struct {
	N        int // Number of replicas
	F        int // Number of failures tolerated
	Client   *ClientConfiguration
	Replicas map[int]*ReplicaAddress // <replica_id, replica_address>
}

func NewConfiguration(client *ClientConfiguration, replicas map[int]*ReplicaAddress) *Configuration {
	return &Configuration{
		N:        len(replicas),
		F:        int(math.Floor(float64((len(replicas) - 1)) / 2)),
		Client:   client,
		Replicas: replicas,
	}
}

func (c *Configuration) QuorumSize() int {
	return c.N - c.F
}

// Example Configs
func GetConfigA() *Configuration {
	client := NewClientConfiguration(0, 0, 0)

	replicas := map[int]*ReplicaAddress{
		0: NewReplicaAddress("localhost", "8000"),
	}

	return NewConfiguration(client, replicas)
}

func GetConfigB() *Configuration {
	client := NewClientConfiguration(123, 666, 101)

	replicas := map[int]*ReplicaAddress{
		101: NewReplicaAddress("localhost", "55209"),
		102: NewReplicaAddress("localhost", "55210"),
		103: NewReplicaAddress("localhost", "55211"),
	}

	return NewConfiguration(client, replicas)
}

func GetConfigC() *Configuration {
	client := NewClientConfiguration(123, 666, 101)

	replicas := map[int]*ReplicaAddress{
		101: NewReplicaAddress("localhost", "55209"),
		102: NewReplicaAddress("localhost", "55210"),
		103: NewReplicaAddress("localhost", "55211"),
		104: NewReplicaAddress("localhost", "55212"),
		105: NewReplicaAddress("localhost", "55213"),
	}

	return NewConfiguration(client, replicas)
}
