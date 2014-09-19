package sentinel

import (
	"github.com/mdevilliers/redishappy/types"
	"github.com/mdevilliers/redishappy/services/logger"
	"testing"
)

func TestBasicEventChannel(t *testing.T) {
	logger.InitLogging("log")
	switchmasterchannel := make(chan MasterSwitchedEvent)
	manager := NewManager(switchmasterchannel)
	defer manager.ClearState()
	manager.Notify(&SentinelAdded{Sentinel: &types.Sentinel{Host: "10.1.1.1", Port: 12345}})

	responseChannel := make(chan SentinelTopology)

	manager.GetState(TopologyRequest{ReplyChannel: responseChannel})
	topologyState := <-responseChannel

	if len(topologyState.Sentinels) != 1 {
		t.Error("Topology count should be 1")
	}

	manager2 := NewManager(switchmasterchannel)
	manager2.Notify(&SentinelAdded{Sentinel: &types.Sentinel{Host: "10.1.1.2", Port: 12345}})

	manager2.GetState(TopologyRequest{ReplyChannel: responseChannel})

	topologyState = <-responseChannel

	if len(topologyState.Sentinels) != 2 {
		t.Errorf("Topology count should be 2 : it is %d", len(topologyState.Sentinels))
	}

	// fmt.Printf("%s\n",util.String(topologyState))
}

func TestAddingAndLoseingASentinel(t *testing.T) {
	logger.InitLogging("log")
	switchmasterchannel := make(chan MasterSwitchedEvent)
	manager := NewManager(switchmasterchannel)
	defer manager.ClearState()

	sentinel := &types.Sentinel{Host: "10.1.1.5", Port: 12345}

	manager.Notify(&SentinelAdded{Sentinel: sentinel})
	manager.Notify(&SentinelLost{Sentinel: sentinel})

	responseChannel := make(chan SentinelTopology)

	manager.GetState(TopologyRequest{ReplyChannel: responseChannel})
	topologyState := <-responseChannel

	if len(topologyState.Sentinels) != 1 {
		t.Error("Topology count should be 1")
	}

	// fmt.Printf("%s\n",util.String(topologyState))
}

func TestAddingInfoToADiscoveredSentinel(t *testing.T) {
	logger.InitLogging("log")
	switchmasterchannel := make(chan MasterSwitchedEvent)
	manager := NewManager(switchmasterchannel)
	defer manager.ClearState()

	sentinel := &types.Sentinel{Host: "10.1.1.6", Port: 12345}

	manager.Notify(&SentinelAdded{Sentinel: sentinel})

	ping := &SentinelPing{Sentinel: sentinel, Clusters: []string{"one", "two", "three"}}
	ping2 := &SentinelPing{Sentinel: sentinel, Clusters: []string{"four", "five"}}
	manager.Notify(ping)
	manager.Notify(ping2)

	responseChannel := make(chan SentinelTopology)

	manager.GetState(TopologyRequest{ReplyChannel: responseChannel})
	topologyState := <-responseChannel

	info, ok := topologyState.FindSentinelInfo(sentinel)

	if ok {
		if len(info.KnownClusters) != 2 {
			t.Error("There should only be 2 known clusters")
		}
	} else {
		t.Error("Added sentinel not found")
	}

	// fmt.Printf("%s\n",util.String(topologyState))
}
