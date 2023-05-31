package types

type MachineMetricSystemUser struct {
	UserID                    uint64
	Machine                   string
	CurrentData               *MachineMetricSystem
	CurrentDataInsertTs       int64
	FiveMinuteOldData         *MachineMetricSystem
	FiveMinuteOldDataInsertTs int64
}

type Eth1AddressSearchItem struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Token   string `json:"token"`
}
