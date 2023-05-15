package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type MachineMetrics struct {
	UserID                          uint64
	Machine                         string
	Process                         string
	Timestamp                       primitive.Timestamp
	ExporterVersion                 string
	CpuProcessSecondsTotal          uint64
	MemoryProcessBytes              uint64
	ClientName                      string
	ClientVersion                   string
	ClientBuild                     uint64
	SyncEth2FallbackConfigured      bool
	SyncEth2FallbackConnected       bool
	DiskBeaconchainBytesTotal       uint64
	NetworkLibp2PBytesTotalReceive  uint64
	NetworkLibp2PBytesTotalTransmit uint64
	NetworkPeersConnected           uint64
	SyncEth1Connected               bool
	SyncEth2Synced                  bool
	SyncBeaconHeadSlot              uint64
	SyncEth1FallbackConfigured      bool
	SyncEth1FallbackConnected       bool
}
