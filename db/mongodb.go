package db

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Prajjawalk/zond-indexer/entity"
	"github.com/Prajjawalk/zond-indexer/services"
	"github.com/Prajjawalk/zond-indexer/types"
	"github.com/Prajjawalk/zond-indexer/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MACHINE_METRICS = "machine_metrics"
var BEACON_CHAIN = "beaconchain"
var VALIDATOR_BALANCES_FAMILY = "vb"
var PROPOSALS_FAMILY = "pr"
var SYNC_COMMITTEES_FAMILY = "sc"
var ATTESTATIONS_FAMILY = "at"

type Mongo struct {
	Client  *mongo.Client
	Db      *mongo.Database
	ChainId string
}

func InitMongodb(connectionString, instance, chainId string) (*Mongo, error) {
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(connectionString).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database(instance)

	mongodb := &Mongo{
		Client:  client,
		Db:      db,
		ChainId: chainId,
	}
	return mongodb, nil
}

func (mongo *Mongo) Close() {
	if err := mongo.Client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}

func (mongo *Mongo) GetClient() interface{} {
	return mongo.Client
}

func (mongodb *Mongo) SaveMachineMetric(process string, userID uint64, machine string, data *entity.MachineMetrics) error {
	data.UserID = userID
	data.Process = process
	data.Machine = machine

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	ts := time.Now()
	lastInsert, err := mongodb.getLastMachineMetricInsertTs(ctx, process, userID, machine)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if lastInsert.Add(59 * time.Second).After(ts) {
		return fmt.Errorf("rate limit, last metric insert was less than 1 min ago")
	}

	_, err = mongodb.Db.Collection(MACHINE_METRICS).InsertOne(ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func (mongodb *Mongo) getLastMachineMetricInsertTs(ctx context.Context, process string, userID uint64, machine string) (time.Time, error) {
	filter := bson.M{"userID": userID, "process": process, "machine": machine}
	var result entity.MachineMetrics

	err := mongodb.Db.Collection(MACHINE_METRICS).FindOne(ctx, filter, &options.FindOneOptions{Sort: bson.D{{Key: "createdat", Value: -1}}}).Decode(&result)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(result.Timestamp.T), 0), nil
}

func (mongodb *Mongo) getMachineMetricNamesMap(userID uint64, searchDepth int) (map[string]bool, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	machineNames := make(map[string]bool)
	var result []entity.MachineMetrics
	filter := bson.M{"userID": userID, "timestamp": bson.M{"$gt": primitive.Timestamp{T: uint32(time.Now().Add(time.Duration(searchDepth*-1) * time.Minute).Unix()), I: 0}, "lt": primitive.Timestamp{T: uint32(time.Now().Unix()), I: 0}}}
	cursor, err := mongodb.Db.Collection(MACHINE_METRICS).Find(ctx, filter, options.Find().SetLimit(int64(searchDepth)))
	if err != nil {
		return machineNames, err
	}

	if err = cursor.All(context.TODO(), &result); err != nil {
		return machineNames, err
	}

	for _, i := range result {
		machineNames[i.Machine] = true
	}

	return machineNames, nil
}

func (mongodb *Mongo) GetMachineMetricsMachineNames(userID uint64) ([]string, error) {
	names, err := mongodb.getMachineMetricNamesMap(userID, 300)
	if err != nil {
		return nil, err
	}

	result := []string{}
	for key := range names {
		result = append(result, key)
	}

	return result, nil
}

func (mongodb *Mongo) GetMachineMetricsMachineCount(userID uint64) (uint64, error) {
	names, err := mongodb.getMachineMetricNamesMap(userID, 15)
	if err != nil {
		return 0, err
	}

	return uint64(len(names)), nil
}

func (mongodb *Mongo) GetMachineMetricsNode(userID uint64, limit, offset int) ([]*types.MachineMetricNode, error) {
	return getMachineMetrics[types.MachineMetricNode](*mongodb, "beaconnode", userID, limit, offset)
}

func (mongodb *Mongo) GetMachineMetricsValidator(userID uint64, limit, offset int) ([]*types.MachineMetricValidator, error) {
	return getMachineMetrics[types.MachineMetricValidator](*mongodb, "validator", userID, limit, offset)
}

func (mongodb *Mongo) GetMachineMetricsSystem(userID uint64, limit, offset int) ([]*types.MachineMetricSystem, error) {
	return getMachineMetrics[types.MachineMetricSystem](*mongodb, "system", userID, limit, offset)
}

func getMachineMetrics[T types.MachineMetricSystem | types.MachineMetricNode | types.MachineMetricValidator](mongodb Mongo, process string, userID uint64, limit, offset int) ([]*T, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	res := make([]*T, 0)
	var result []*T
	if offset <= 0 {
		offset = 1
	}

	gapSize := getMachineStatsGap(uint64(limit))

	filter := bson.M{"userID": userID, "process": process}
	cursor, err := mongodb.Db.Collection(MACHINE_METRICS).Find(ctx, filter, options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)))
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.TODO(), &result); err != nil {
		return nil, err
	}

	for idx, i := range result {
		if idx%gapSize != 0 {
			continue
		}

		res = append(res, i)
	}

	return res, nil
}

func (mongodb *Mongo) GetMachineMetricsForNotifications(eventList []services.MachineEvents) (map[uint64]map[string]*types.MachineMetricSystemUser, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*200))
	defer cancel()
	res := make(map[uint64]map[string]*types.MachineMetricSystemUser) // userID -> machine -> data

	limit := 5
	count := 0

	for _, i := range eventList {
		filter := bson.M{"userID": i.UserID, "machine": i.MachineName, "process": "system"}
		cursor, err := mongodb.Db.Collection(MACHINE_METRICS).Find(ctx, filter, options.Find().SetLimit(int64(limit)))
		if err != nil {
			return nil, err
		}

		var result types.MachineMetricSystem
		if err = cursor.All(context.TODO(), &result); err != nil {
			return nil, err
		}

		if _, found := res[i.UserID]; !found {
			res[i.UserID] = make(map[string]*types.MachineMetricSystemUser)
		}

		last, found := res[i.UserID][i.MachineName]

		if found && count == limit-1 {
			res[i.UserID][i.MachineName] = &types.MachineMetricSystemUser{
				UserID:                    i.UserID,
				Machine:                   i.MachineName,
				CurrentData:               last.CurrentData,
				FiveMinuteOldData:         &result,
				CurrentDataInsertTs:       last.CurrentDataInsertTs,
				FiveMinuteOldDataInsertTs: time.Now().Unix(), //this field is gcp_bigtable ReadItem timestamp
			}
		} else {
			res[i.UserID][i.MachineName] = &types.MachineMetricSystemUser{
				UserID:                    i.UserID,
				Machine:                   i.MachineName,
				CurrentData:               &result,
				FiveMinuteOldData:         nil,
				CurrentDataInsertTs:       time.Now().Unix(), //this field is gcp_bigtable ReadItem timestamp
				FiveMinuteOldDataInsertTs: 0,
			}
		}
		count++
	}

	return res, nil
}

func (mongodb *Mongo) SaveValidatorBalances(epoch uint64, validators []*types.Validator) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	start := time.Now()

	for _, validator := range validators {
		_, err := mongodb.Db.Collection(BEACON_CHAIN).InsertOne(ctx, bson.D{{Key: "validatorId", Value: validator.Index}, {Key: "balance", Value: validator.Balance}, {Key: "effectiveBalance", Value: validator.EffectiveBalance}, {Key: "type", Value: VALIDATOR_BALANCES_FAMILY}, {Key: "epoch", Value: epoch}})
		if err != nil {
			return err
		}
	}

	logger.Infof("exported validator balances in %v", time.Since(start))
	return nil
}

func (mongodb *Mongo) SaveAttestationAssignments(epoch uint64, assignments map[string]uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	start := time.Now()

	validatorsPerSlot := make(map[uint64][]uint64)
	for key, validator := range assignments {
		keySplit := strings.Split(key, "-")

		attesterslot, err := strconv.ParseUint(keySplit[0], 10, 64)
		if err != nil {
			return err
		}

		if validatorsPerSlot[attesterslot] == nil {
			validatorsPerSlot[attesterslot] = make([]uint64, 0, len(assignments)/int(utils.Config.Chain.Config.SlotsPerEpoch))
		}
		validatorsPerSlot[attesterslot] = append(validatorsPerSlot[attesterslot], validator)
	}

	for slot, validators := range validatorsPerSlot {
		for _, validator := range validators {
			_, err := mongodb.Db.Collection(BEACON_CHAIN).InsertOne(ctx, bson.D{{Key: "chainID", Value: mongodb.ChainId}, {Key: "epoch", Value: epoch}, {Key: "validatorId", Value: validator}, {Key: "attestorSlot", Value: slot}, {Key: "type", Value: ATTESTATIONS_FAMILY}})
			if err != nil {
				return err
			}
		}
	}

	logger.Infof("exported attestation assignments in %v", time.Since(start))
	return nil
}

func (mongodb *Mongo) SaveProposalAssignments(epoch uint64, assignments map[uint64]uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	start := time.Now()

	for slot, validator := range assignments {
		_, err := mongodb.Db.Collection(BEACON_CHAIN).InsertOne(ctx, bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "validatorId", Value: validator}, {Key: "type", Value: PROPOSALS_FAMILY}, {Key: "status", Value: uint64(1)}, {Key: "epoch", Value: epoch}, {Key: "slot", Value: slot}})
		if err != nil {
			return err
		}
	}

	logger.Infof("exported proposal assignments to bigtable in %v", time.Since(start))
	return nil
}

func (mongodb *Mongo) SaveSyncCommitteesAssignments(startSlot, endSlot uint64, validators []uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	start := time.Now()

	for i := startSlot; i <= endSlot; i++ {
		for _, validator := range validators {
			_, err := mongodb.Db.Collection(BEACON_CHAIN).InsertOne(ctx, bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "validatorId", Value: validator}, {Key: "type", Value: SYNC_COMMITTEES_FAMILY}, {Key: "epoch", Value: i / utils.Config.Chain.Config.SlotsPerEpoch}, {Key: "slot", Value: i}})
			if err != nil {
				return err
			}
		}
	}

	logger.Infof("exported sync committee assignments in %v", time.Since(start))
	return nil
}

func (mongodb *Mongo) SaveAttestations(blocks map[uint64]map[string]*types.Block) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	start := time.Now()

	attestationsBySlot := make(map[uint64]map[uint64]uint64) //map[attestedSlot]map[validator]includedSlot

	slots := make([]uint64, 0, len(blocks))
	for slot := range blocks {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool {
		return slots[i] < slots[j]
	})

	for _, slot := range slots {
		for _, b := range blocks[slot] {
			logger.Infof("processing slot %v", slot)
			for _, a := range b.Attestations {
				for _, validator := range a.Attesters {
					inclusionSlot := slot
					attestedSlot := a.Data.Slot
					if attestationsBySlot[attestedSlot] == nil {
						attestationsBySlot[attestedSlot] = make(map[uint64]uint64)
					}

					if attestationsBySlot[attestedSlot][validator] == 0 || inclusionSlot < attestationsBySlot[attestedSlot][validator] {
						attestationsBySlot[attestedSlot][validator] = inclusionSlot
					}
				}
			}
		}
	}

	for attestedSlot, inclusions := range attestationsBySlot {
		for validator := range inclusions {
			_, err := mongodb.Db.Collection(BEACON_CHAIN).InsertOne(ctx, bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "validatorId", Value: validator}, {Key: "epoch", Value: attestedSlot / utils.Config.Chain.Config.SlotsPerEpoch}, {Key: "attestorSlot", Value: attestedSlot}, {Key: "type", Value: ATTESTATIONS_FAMILY}})
			if err != nil {
				return err
			}
		}
	}

	logger.Infof("exported attestations in %v", time.Since(start))
	return nil
}

func (mongodb *Mongo) SaveProposals(blocks map[uint64]map[string]*types.Block) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	start := time.Now()

	slots := make([]uint64, 0, len(blocks))
	for slot := range blocks {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool {
		return slots[i] < slots[j]
	})

	for _, slot := range slots {
		for _, b := range blocks[slot] {
			if len(b.BlockRoot) != 32 { // skip dummy blocks
				continue
			}

			_, err := mongodb.Db.Collection(BEACON_CHAIN).InsertOne(ctx, bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "validatorId", Value: b.Proposer}, {Key: "epoch", Value: b.Slot / utils.Config.Chain.Config.SlotsPerEpoch}, {Key: "slot", Value: slot}, {Key: "type", Value: PROPOSALS_FAMILY}})
			if err != nil {
				return err
			}
		}
	}

	logger.Infof("exported proposals in %v", time.Since(start))
	return nil
}

func (mongodb *Mongo) SaveSyncComitteeDuties(blocks map[uint64]map[string]*types.Block) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	start := time.Now()

	dutiesBySlot := make(map[uint64]map[uint64]bool) //map[dutiesSlot]map[validator]bool

	slots := make([]uint64, 0, len(blocks))
	for slot := range blocks {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool {
		return slots[i] < slots[j]
	})

	for _, slot := range slots {
		for _, b := range blocks[slot] {
			if b.Status == 2 {
				continue
			} else if b.SyncAggregate != nil && len(b.SyncAggregate.SyncCommitteeValidators) > 0 {
				bitLen := len(b.SyncAggregate.SyncCommitteeBits) * 8
				valLen := len(b.SyncAggregate.SyncCommitteeValidators)
				if bitLen < valLen {
					return fmt.Errorf("error getting sync_committee participants: bitLen != valLen: %v != %v", bitLen, valLen)
				}
				for i, valIndex := range b.SyncAggregate.SyncCommitteeValidators {
					if dutiesBySlot[b.Slot] == nil {
						dutiesBySlot[b.Slot] = make(map[uint64]bool)
					}
					dutiesBySlot[b.Slot][valIndex] = utils.BitAtVector(b.SyncAggregate.SyncCommitteeBits, i)
				}
			}
		}
	}

	if len(dutiesBySlot) == 0 {
		logger.Infof("no sync duties to export")
		return nil
	}

	for slot, validators := range dutiesBySlot {
		for validator := range validators {
			_, err := mongodb.Db.Collection(BEACON_CHAIN).InsertOne(ctx, bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "validatorId", Value: validator}, {Key: "type", Value: SYNC_COMMITTEES_FAMILY}, {Key: "epoch", Value: slot / utils.Config.Chain.Config.SlotsPerEpoch}, {Key: "slot", Value: slot}})
			if err != nil {
				return err
			}
		}
	}

	logger.Infof("exported sync committee duties in %v", time.Since(start))
	return nil
}

func (mongodb *Mongo) GetValidatorBalanceHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorBalance, error) {
	valLen := len(validators)
	getAllThreshold := 1000
	validatorMap := make(map[uint64]bool, valLen)
	for _, validatorIndex := range validators {
		validatorMap[validatorIndex] = true
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	res := make(map[uint64][]*types.ValidatorBalance, valLen)
	if endEpoch < startEpoch { // handle overflows
		startEpoch = 0
	}

	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: VALIDATOR_BALANCES_FAMILY}, {Key: "validatorId", Value: bson.D{{Key: "validatorId", Value: bson.D{{Key: "$in", Value: validators}}}}}, {Key: "epoch", Value: bson.D{{Key: "$gte", Value: startEpoch}, {Key: "$lte", Value: endEpoch}}}}
	cursor, err := mongodb.Db.Collection(BEACON_CHAIN).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var results []*entity.ValidatorBalancesFamily
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	for _, result := range results {
		// If we requested more than getAllThreshold validators we will
		// get data for all validators and need to filter out all
		// unwanted ones
		if valLen >= getAllThreshold && !validatorMap[result.ValidatorId] {
			continue
		}

		if res[result.ValidatorId] == nil {
			res[result.ValidatorId] = make([]*types.ValidatorBalance, 0)
		}

		res[result.ValidatorId] = append(res[result.ValidatorId], &types.ValidatorBalance{
			Epoch:            result.Epoch,
			Balance:          result.Balance,
			EffectiveBalance: result.EffectiveBalance,
			Index:            result.ValidatorId,
			PublicKey:        []byte{},
		})
	}

	return res, nil
}

// func (mongodb *Mongo) GetValidatorAttestationHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorAttestation, error) {
// 	valLen := len(validators)

// 	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
// 	defer cancel()

// 	res := make(map[uint64][]*types.ValidatorAttestation, len(validators))
// 	if endEpoch < startEpoch { // handle overflows
// 		startEpoch = 0
// 	}

// 	filter := bson.D{{"chainId", mongodb.ChainId}, {"type", ATTESTATIONS_FAMILY}, {Key: "validatorId", Value: bson.D{{Key: "validatorId", Value: bson.D{{Key: "$in", Value: validators}}}}}, {Key: "epoch", Value: bson.D{{Key: "$gte", Value: startEpoch}, {Key: "$lte", Value: endEpoch}}}}
// 	return res, nil
// }

func (mongodb *Mongo) GetValidatorSyncDutiesHistoryOrdered(validatorIndex uint64, startEpoch uint64, endEpoch uint64, reverseOrdering bool) ([]*types.ValidatorSyncParticipation, error) {
	res, err := mongodb.GetValidatorSyncDutiesHistory([]uint64{validatorIndex}, startEpoch, endEpoch)
	if err != nil {
		return nil, err
	}
	if reverseOrdering {
		utils.ReverseSlice(res[validatorIndex])
	}
	return res[validatorIndex], nil
}

func (mongodb *Mongo) GetValidatorSyncDutiesHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorSyncParticipation, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	res := make(map[uint64][]*types.ValidatorSyncParticipation, len(validators))
	if endEpoch < startEpoch { // handle overflows
		startEpoch = 0
	}

	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: SYNC_COMMITTEES_FAMILY}, {Key: "validatorId", Value: bson.D{{Key: "validatorId", Value: bson.D{{Key: "$in", Value: validators}}}}}, {Key: "epoch", Value: bson.D{{Key: "$gte", Value: startEpoch}, {Key: "$lte", Value: endEpoch}}}}
	cursor, err := mongodb.Db.Collection(BEACON_CHAIN).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var results []*entity.SyncCommitteesFamily
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	for _, result := range results {
		slot := result.Slot
		status := uint64(1)
		validator := result.ValidatorId

		if res[validator] == nil {
			res[validator] = make([]*types.ValidatorSyncParticipation, 0)
		}

		if len(res[validator]) > 1 && res[validator][len(res[validator])-1].Slot == slot {
			res[validator][len(res[validator])-1].Status = status
		} else {
			res[validator] = append(res[validator], &types.ValidatorSyncParticipation{
				Slot:   slot,
				Status: status,
			})
		}
	}

	return res, nil
}

// func (mongodb *Mongo) GetValidatorMissedAttestationsCount(validators []uint64, firstEpoch uint64, lastEpoch uint64) {
// 	if firstEpoch > lastEpoch {
// 		return nil, fmt.Errorf("GetValidatorMissedAttestationsCount received an invalid firstEpoch (%d) and lastEpoch (%d) combination", firstEpoch, lastEpoch)
// 	}

// 	res := make(map[uint64]*types.ValidatorMissedAttestationsStatistic)
// 	for e := firstEpoch; e <= lastEpoch; e++ {
// 		data, err := mongodb.GetValidatorAttestationHistory(validators, e, e)

// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// }
