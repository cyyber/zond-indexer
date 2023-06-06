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
	// "github.com/Prajjawalk/zond-indexer/services"
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
var max_block_number = uint64(1000000000)
var INCOME_DETAILS_COLUMN_FAMILY = "id"
var STATS_COLUMN_FAMILY = "stats"
var SERIES_FAMILY = "series"

var MongodbClient *Mongo

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

func (mongodb *Mongo) GetMachineMetricsForNotifications(eventList []entity.MachineMetrics) (map[uint64]map[string]*types.MachineMetricSystemUser, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*200))
	defer cancel()
	res := make(map[uint64]map[string]*types.MachineMetricSystemUser) // userID -> machine -> data

	limit := 5
	count := 0

	for _, i := range eventList {
		filter := bson.M{"userID": i.UserID, "machine": i.Machine, "process": "system"}
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

		last, found := res[i.UserID][i.Machine]

		if found && count == limit-1 {
			res[i.UserID][i.Machine] = &types.MachineMetricSystemUser{
				UserID:                    i.UserID,
				Machine:                   i.Machine,
				CurrentData:               last.CurrentData,
				FiveMinuteOldData:         &result,
				CurrentDataInsertTs:       last.CurrentDataInsertTs,
				FiveMinuteOldDataInsertTs: time.Now().Unix(), //this field is gcp_bigtable ReadItem timestamp
			}
		} else {
			res[i.UserID][i.Machine] = &types.MachineMetricSystemUser{
				UserID:                    i.UserID,
				Machine:                   i.Machine,
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

func (mongodb *Mongo) GetValidatorAttestationHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorAttestation, error) {

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	res := make(map[uint64][]*types.ValidatorAttestation, len(validators))
	if endEpoch < startEpoch { // handle overflows
		startEpoch = 0
	}

	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: ATTESTATIONS_FAMILY}, {Key: "validatorId", Value: bson.D{{Key: "validatorId", Value: bson.D{{Key: "$in", Value: validators}}}}}, {Key: "epoch", Value: bson.D{{Key: "$gte", Value: startEpoch}, {Key: "$lte", Value: endEpoch}}}}
	cursor, err := mongodb.Db.Collection(BEACON_CHAIN).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var results []*entity.AttestationFamily
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	for _, result := range results {
		attesterSlot := result.AttestorSlot
		validator := result.ValidatorId

		inclusionSlot := uint64(max_block_number) - uint64(time.Now().Unix())/1000

		status := uint64(1)
		if inclusionSlot == uint64(max_block_number) {
			inclusionSlot = 0
			status = 0
		}

		if res[validator] == nil {
			res[validator] = make([]*types.ValidatorAttestation, 0)
		}

		if len(res[validator]) > 1 && res[validator][len(res[validator])-1].AttesterSlot == attesterSlot {
			res[validator][len(res[validator])-1].InclusionSlot = inclusionSlot
			res[validator][len(res[validator])-1].Status = status
			res[validator][len(res[validator])-1].Delay = int64(inclusionSlot - attesterSlot)
		} else {
			res[validator] = append(res[validator], &types.ValidatorAttestation{
				Index:          validator,
				Epoch:          attesterSlot / utils.Config.Chain.Config.SlotsPerEpoch,
				AttesterSlot:   attesterSlot,
				CommitteeIndex: 0,
				Status:         status,
				InclusionSlot:  inclusionSlot,
				Delay:          int64(inclusionSlot) - int64(attesterSlot) - 1,
			})
		}
	}

	return res, nil
}

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

func (mongodb *Mongo) GetValidatorMissedAttestationsCount(validators []uint64, firstEpoch uint64, lastEpoch uint64) (map[uint64]*types.ValidatorMissedAttestationsStatistic, error) {
	if firstEpoch > lastEpoch {
		return nil, fmt.Errorf("GetValidatorMissedAttestationsCount received an invalid firstEpoch (%d) and lastEpoch (%d) combination", firstEpoch, lastEpoch)
	}

	res := make(map[uint64]*types.ValidatorMissedAttestationsStatistic)
	for e := firstEpoch; e <= lastEpoch; e++ {
		data, err := mongodb.GetValidatorAttestationHistory(validators, e, e)

		if err != nil {
			return nil, err
		}

		logger.Infof("retrieved attestation history for epoch %v", e)

		for validator, attestations := range data {
			for _, attestation := range attestations {
				if attestation.Status == 0 {
					if res[validator] == nil {
						res[validator] = &types.ValidatorMissedAttestationsStatistic{
							Index: validator,
						}
					}
					res[validator].MissedAttestations++
				}
			}
		}
	}

	return res, nil
}

func (mongodb *Mongo) GetValidatorSyncDutiesStatistics(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]*types.ValidatorSyncDutiesStatistic, error) {
	data, err := mongodb.GetValidatorSyncDutiesHistory(validators, startEpoch, endEpoch)

	if err != nil {
		return nil, err
	}

	res := make(map[uint64]*types.ValidatorSyncDutiesStatistic)
	for validator, duties := range data {
		if res[validator] == nil && len(duties) > 0 {
			res[validator] = &types.ValidatorSyncDutiesStatistic{
				Index: validator,
			}
		}

		for _, duty := range duties {
			if duty.Status == 0 {
				res[validator].MissedSync++
			} else {
				res[validator].ParticipatedSync++
			}
		}
	}

	return res, nil
}

func (mongodb *Mongo) GetValidatorEffectiveness(validators []uint64, epoch uint64) ([]*types.ValidatorEffectiveness, error) {
	data, err := mongodb.GetValidatorAttestationHistory(validators, epoch-100, epoch)

	if err != nil {
		return nil, err
	}

	res := make([]*types.ValidatorEffectiveness, 0, len(validators))
	type readings struct {
		Count uint64
		Sum   float64
	}

	aggEffectiveness := make(map[uint64]*readings)

	for validator, history := range data {
		for _, attestation := range history {
			if aggEffectiveness[validator] == nil {
				aggEffectiveness[validator] = &readings{}
			}
			if attestation.InclusionSlot > 0 {
				// logger.Infof("adding %v for epoch %v %.2f%%", attestation.InclusionSlot, attestation.AttesterSlot, 1.0/float64(attestation.InclusionSlot-attestation.AttesterSlot)*100)
				aggEffectiveness[validator].Sum += 1.0 / float64(attestation.InclusionSlot-attestation.AttesterSlot)
				aggEffectiveness[validator].Count++
			} else {
				aggEffectiveness[validator].Sum += 0 // missed attestations get a penalty of 32 slots
				aggEffectiveness[validator].Count++
			}
		}
	}
	for validator, reading := range aggEffectiveness {
		res = append(res, &types.ValidatorEffectiveness{
			Validatorindex:        validator,
			AttestationEfficiency: float64(reading.Sum) / float64(reading.Count) * 100,
		})
	}

	return res, nil
}

func (mongodb *Mongo) GetValidatorBalanceStatistics(startEpoch, endEpoch uint64) (map[uint64]*types.ValidatorBalanceStatistic, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*10))
	defer cancel()

	res := make(map[uint64]*types.ValidatorBalanceStatistic)
	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: VALIDATOR_BALANCES_FAMILY}, {Key: "epoch", Value: bson.D{{Key: "$gte", Value: startEpoch}, {Key: "$lte", Value: endEpoch}}}}
	cursor, err := mongodb.Db.Collection(BEACON_CHAIN).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var results []*entity.ValidatorBalancesFamily
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	for _, result := range results {
		epoch := result.Epoch
		validator := result.ValidatorId
		balance := result.Balance
		effectiveBalance := result.EffectiveBalance

		if res[validator] == nil {
			res[validator] = &types.ValidatorBalanceStatistic{
				Index:                 validator,
				MinEffectiveBalance:   effectiveBalance,
				MaxEffectiveBalance:   0,
				MinBalance:            balance,
				MaxBalance:            0,
				StartEffectiveBalance: 0,
				EndEffectiveBalance:   0,
				StartBalance:          0,
				EndBalance:            0,
			}
		}

		// logger.Info(epoch, startEpoch)
		if epoch == startEpoch {
			res[validator].StartBalance = balance
			res[validator].StartEffectiveBalance = effectiveBalance
		}

		if epoch == endEpoch {
			res[validator].EndBalance = balance
			res[validator].EndEffectiveBalance = effectiveBalance
		}

		if balance > res[validator].MaxBalance {
			res[validator].MaxBalance = balance
		}
		if balance < res[validator].MinBalance {
			res[validator].MinBalance = balance
		}

		if balance > res[validator].MaxEffectiveBalance {
			res[validator].MaxEffectiveBalance = balance
		}
		if balance < res[validator].MinEffectiveBalance {
			res[validator].MinEffectiveBalance = balance
		}
	}

	return res, nil
}

func (mongodb *Mongo) GetValidatorProposalHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64][]*types.ValidatorProposal, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	res := make(map[uint64][]*types.ValidatorProposal, len(validators))
	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: PROPOSALS_FAMILY}, {Key: "validatorId", Value: bson.D{{Key: "validatorId", Value: bson.D{{Key: "$in", Value: validators}}}}}, {Key: "epoch", Value: bson.D{{Key: "$gte", Value: startEpoch}, {Key: "$lte", Value: endEpoch}}}}
	cursor, err := mongodb.Db.Collection(BEACON_CHAIN).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var results []*entity.ProposalsFamily
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	for _, result := range results {
		proposalSlot := result.Slot
		validator := result.ValidatorId
		status := uint64(1)
		id := result.ID
		inclusionSlot := max_block_number - uint64(id.Timestamp().UnixNano()/1e3)/1000

		if inclusionSlot == max_block_number {
			inclusionSlot = 0
			status = 2
		}

		if res[validator] == nil {
			res[validator] = make([]*types.ValidatorProposal, 0)
		}

		if len(res[validator]) > 1 && res[validator][len(res[validator])-1].Slot == proposalSlot {
			res[validator][len(res[validator])-1].Slot = proposalSlot
			res[validator][len(res[validator])-1].Status = status
		} else {
			res[validator] = append(res[validator], &types.ValidatorProposal{
				Index:  validator,
				Status: status,
				Slot:   proposalSlot,
			})
		}

	}

	return res, nil
}

func (mongodb *Mongo) SaveValidatorIncomeDetails(epoch uint64, rewards map[uint64]*types.ValidatorEpochIncome) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	start := time.Now()
	total := &entity.Stats{}
	ts := utils.EpochToTime(epoch).UnixMicro()
	muts := 0
	for i, rewardDetails := range rewards {
		muts++
		doc, err := utils.ToDoc(&entity.IncomeDetailsColumnFamily{
			ValidatorId:                        i,
			AttestationSourceReward:            rewardDetails.AttestationSourceReward,
			AttestationSourcePenalty:           rewardDetails.AttestationSourcePenalty,
			AttestationTargetReward:            rewardDetails.AttestationTargetReward,
			AttestationTargetPenalty:           rewardDetails.AttestationTargetPenalty,
			AttestationHeadReward:              rewardDetails.AttestationHeadReward,
			FinalityDelayPenalty:               rewardDetails.FinalityDelayPenalty,
			ProposerSlashingInclusionReward:    rewardDetails.ProposerSlashingInclusionReward,
			ProposerAttestationInclusionReward: rewardDetails.ProposerAttestationInclusionReward,
			ProposerSyncInclusionReward:        rewardDetails.ProposerSyncInclusionReward,
			SyncCommitteeReward:                rewardDetails.SyncCommitteeReward,
			SyncCommitteePenalty:               rewardDetails.SyncCommitteePenalty,
			SlashingReward:                     rewardDetails.SlashingReward,
			SlashingPenalty:                    rewardDetails.SlashingPenalty,
			TxFeeRewardWei:                     rewardDetails.TxFeeRewardWei,
			ProposalsMissed:                    rewardDetails.ProposalsMissed,
			Timestamp:                          ts,
			Type:                               INCOME_DETAILS_COLUMN_FAMILY,
			ChainId:                            mongodb.ChainId,
			Epoch:                              epoch,
		})
		if err != nil {
			return err
		}
		_, err = mongodb.Db.Collection(BEACON_CHAIN).InsertOne(ctx, doc)
		if err != nil {
			return err
		}

		total.AttestationHeadReward += rewardDetails.AttestationHeadReward
		total.AttestationSourceReward += rewardDetails.AttestationSourceReward
		total.AttestationSourcePenalty += rewardDetails.AttestationSourcePenalty
		total.AttestationTargetReward += rewardDetails.AttestationTargetReward
		total.AttestationTargetPenalty += rewardDetails.AttestationTargetPenalty
		total.FinalityDelayPenalty += rewardDetails.FinalityDelayPenalty
		total.ProposerSlashingInclusionReward += rewardDetails.ProposerSlashingInclusionReward
		total.ProposerAttestationInclusionReward += rewardDetails.ProposerAttestationInclusionReward
		total.ProposerSyncInclusionReward += rewardDetails.ProposerSyncInclusionReward
		total.SyncCommitteeReward += rewardDetails.SyncCommitteeReward
		total.SyncCommitteePenalty += rewardDetails.SyncCommitteePenalty
		total.SlashingReward += rewardDetails.SlashingReward
		total.SlashingPenalty += rewardDetails.SlashingPenalty
		total.TxFeeRewardWei = utils.AddBigInts(total.TxFeeRewardWei, rewardDetails.TxFeeRewardWei)
	}

	total.Timestamp = ts
	total.Type = STATS_COLUMN_FAMILY
	total.ChainId = mongodb.ChainId
	total.Epoch = epoch

	statsDoc, err := utils.ToDoc(total)
	if err != nil {
		return err
	}

	_, err = mongodb.Db.Collection(BEACON_CHAIN).InsertOne(ctx, statsDoc)
	if err != nil {
		return err
	}

	logger.Infof("exported validator income details for epoch %v in %v", epoch, time.Since(start))
	return nil
}

func (mongodb *Mongo) GetEpochIncomeHistoryDescending(startEpoch uint64, endEpoch uint64) (*types.ValidatorEpochIncome, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	res := types.ValidatorEpochIncome{}
	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: STATS_COLUMN_FAMILY}, {Key: "epoch", Value: bson.D{{Key: "$gte", Value: startEpoch}, {Key: "$lte", Value: endEpoch}}}}
	var results []entity.IncomeDetailsColumnFamily

	cursor, err := mongodb.Db.Collection(BEACON_CHAIN).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("unable to get income history details for epoch %d", startEpoch)
	}

	res.AttestationHeadReward = results[0].AttestationHeadReward
	res.AttestationSourceReward = results[0].AttestationSourceReward
	res.AttestationSourcePenalty = results[0].AttestationSourcePenalty
	res.AttestationTargetReward = results[0].AttestationTargetReward
	res.AttestationTargetPenalty = results[0].AttestationTargetPenalty
	res.FinalityDelayPenalty = results[0].FinalityDelayPenalty
	res.ProposerSlashingInclusionReward = results[0].ProposerSlashingInclusionReward
	res.ProposerAttestationInclusionReward = results[0].ProposerAttestationInclusionReward
	res.ProposerSyncInclusionReward = results[0].ProposerSyncInclusionReward
	res.SyncCommitteeReward = results[0].SyncCommitteeReward
	res.SyncCommitteePenalty = results[0].SyncCommitteePenalty
	res.SlashingReward = results[0].SlashingReward
	res.SlashingPenalty = results[0].SlashingPenalty
	res.TxFeeRewardWei = results[0].TxFeeRewardWei
	res.ProposalsMissed = results[0].ProposalsMissed

	return &res, nil
}

func (mongodb *Mongo) GetEpochIncomeHistory(epoch uint64) (*types.ValidatorEpochIncome, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancel()

	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: STATS_COLUMN_FAMILY}, {Key: "epoch", Value: epoch}}
	var results []entity.IncomeDetailsColumnFamily

	cursor, err := mongodb.Db.Collection(BEACON_CHAIN).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) != 0 {
		res := types.ValidatorEpochIncome{}
		res.AttestationHeadReward = results[0].AttestationHeadReward
		res.AttestationSourceReward = results[0].AttestationSourceReward
		res.AttestationSourcePenalty = results[0].AttestationSourcePenalty
		res.AttestationTargetReward = results[0].AttestationTargetReward
		res.AttestationTargetPenalty = results[0].AttestationTargetPenalty
		res.FinalityDelayPenalty = results[0].FinalityDelayPenalty
		res.ProposerSlashingInclusionReward = results[0].ProposerSlashingInclusionReward
		res.ProposerAttestationInclusionReward = results[0].ProposerAttestationInclusionReward
		res.ProposerSyncInclusionReward = results[0].ProposerSyncInclusionReward
		res.SyncCommitteeReward = results[0].SyncCommitteeReward
		res.SyncCommitteePenalty = results[0].SyncCommitteePenalty
		res.SlashingReward = results[0].SlashingReward
		res.SlashingPenalty = results[0].SlashingPenalty
		res.TxFeeRewardWei = results[0].TxFeeRewardWei
		res.ProposalsMissed = results[0].ProposalsMissed
		return &res, nil
	}

	// if there is no result we have to calculate the sum
	income, err := mongodb.GetValidatorIncomeDetailsHistory([]uint64{}, epoch, 1)
	if err != nil {
		logger.WithError(err).Error("error getting validator income history")
	}

	total := &types.ValidatorEpochIncome{}

	for _, epochs := range income {
		for _, details := range epochs {
			total.AttestationHeadReward += details.AttestationHeadReward
			total.AttestationSourceReward += details.AttestationSourceReward
			total.AttestationSourcePenalty += details.AttestationSourcePenalty
			total.AttestationTargetReward += details.AttestationTargetReward
			total.AttestationTargetPenalty += details.AttestationTargetPenalty
			total.FinalityDelayPenalty += details.FinalityDelayPenalty
			total.ProposerSlashingInclusionReward += details.ProposerSlashingInclusionReward
			total.ProposerAttestationInclusionReward += details.ProposerAttestationInclusionReward
			total.ProposerSyncInclusionReward += details.ProposerSyncInclusionReward
			total.SyncCommitteeReward += details.SyncCommitteeReward
			total.SyncCommitteePenalty += details.SyncCommitteePenalty
			total.SlashingReward += details.SlashingReward
			total.SlashingPenalty += details.SlashingPenalty
			total.TxFeeRewardWei = utils.AddBigInts(total.TxFeeRewardWei, details.TxFeeRewardWei)
		}
	}

	return total, nil
}

// GetValidatorIncomeDetailsHistory returns the validator income details
// startEpoch & endEpoch are inclusive
func (mongodb *Mongo) GetValidatorIncomeDetailsHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]map[uint64]*types.ValidatorEpochIncome, error) {
	if startEpoch > endEpoch {
		startEpoch = 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*180)
	defer cancel()

	// logger.Infof("range: %v to %v", rangeStart, rangeEnd)
	res := make(map[uint64]map[uint64]*types.ValidatorEpochIncome, len(validators))

	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: INCOME_DETAILS_COLUMN_FAMILY}, {Key: "validatorId", Value: bson.D{{Key: "validatorId", Value: bson.D{{Key: "$in", Value: validators}}}}}, {Key: "epoch", Value: bson.D{{Key: "$gte", Value: startEpoch}, {Key: "$lte", Value: endEpoch}}}}
	cursor, err := mongodb.Db.Collection(BEACON_CHAIN).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var results []*entity.IncomeDetailsColumnFamily
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	for _, result := range results {
		epoch := result.Epoch
		validator := result.ValidatorId
		incomeDetails := &types.ValidatorEpochIncome{}
		incomeDetails.AttestationHeadReward = result.AttestationHeadReward
		incomeDetails.AttestationSourceReward = result.AttestationSourceReward
		incomeDetails.AttestationSourcePenalty = result.AttestationSourcePenalty
		incomeDetails.AttestationTargetReward = result.AttestationTargetReward
		incomeDetails.AttestationTargetPenalty = result.AttestationTargetPenalty
		incomeDetails.FinalityDelayPenalty = result.FinalityDelayPenalty
		incomeDetails.ProposerSlashingInclusionReward = result.ProposerSlashingInclusionReward
		incomeDetails.ProposerAttestationInclusionReward = result.ProposerAttestationInclusionReward
		incomeDetails.ProposerSyncInclusionReward = result.ProposerSyncInclusionReward
		incomeDetails.SyncCommitteeReward = result.SyncCommitteeReward
		incomeDetails.SyncCommitteePenalty = result.SyncCommitteePenalty
		incomeDetails.SlashingReward = result.SlashingReward
		incomeDetails.SlashingPenalty = result.SlashingPenalty
		incomeDetails.TxFeeRewardWei = result.TxFeeRewardWei
		incomeDetails.ProposalsMissed = result.ProposalsMissed

		if res[validator] == nil {
			res[validator] = make(map[uint64]*types.ValidatorEpochIncome)
		}

		res[validator][epoch] = incomeDetails
	}

	return res, nil
}

func (mongodb *Mongo) GetAggregatedValidatorIncomeDetailsHistory(validators []uint64, startEpoch uint64, endEpoch uint64) (map[uint64]*types.ValidatorEpochIncome, error) {
	if startEpoch > endEpoch {
		startEpoch = 0
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*10))
	defer cancel()
	incomeStats := make(map[uint64]*types.ValidatorEpochIncome, len(validators))
	filter := bson.D{{Key: "chainId", Value: mongodb.ChainId}, {Key: "type", Value: INCOME_DETAILS_COLUMN_FAMILY}, {Key: "validatorId", Value: bson.D{{Key: "validatorId", Value: bson.D{{Key: "$in", Value: validators}}}}}, {Key: "epoch", Value: bson.D{{Key: "$gte", Value: startEpoch}, {Key: "$lte", Value: endEpoch}}}}
	cursor, err := mongodb.Db.Collection(BEACON_CHAIN).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var results []*entity.IncomeDetailsColumnFamily
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	start := time.Now()
	for _, result := range results {
		epoch := result.Epoch
		validator := result.ValidatorId

		if incomeStats[validator] == nil {
			incomeStats[validator] = &types.ValidatorEpochIncome{}
		}

		incomeStats[validator].AttestationHeadReward += result.AttestationHeadReward
		incomeStats[validator].AttestationSourceReward += result.AttestationSourceReward
		incomeStats[validator].AttestationSourcePenalty += result.AttestationSourcePenalty
		incomeStats[validator].AttestationTargetReward += result.AttestationTargetReward
		incomeStats[validator].AttestationTargetPenalty += result.AttestationTargetPenalty
		incomeStats[validator].FinalityDelayPenalty += result.FinalityDelayPenalty
		incomeStats[validator].ProposerSlashingInclusionReward += result.ProposerSlashingInclusionReward
		incomeStats[validator].ProposerAttestationInclusionReward += result.ProposerAttestationInclusionReward
		incomeStats[validator].ProposerSyncInclusionReward += result.ProposerSyncInclusionReward
		incomeStats[validator].SyncCommitteeReward += result.SyncCommitteeReward
		incomeStats[validator].SyncCommitteePenalty += result.SyncCommitteePenalty
		incomeStats[validator].SlashingReward += result.SlashingReward
		incomeStats[validator].SlashingPenalty += result.SlashingPenalty
		incomeStats[validator].TxFeeRewardWei = utils.AddBigInts(incomeStats[validator].TxFeeRewardWei, result.TxFeeRewardWei)

		logger.Infof("processed income data for epoch %v in %v", epoch, time.Since(start))
	}

	return incomeStats, nil
}

func (mongodb *Mongo) DeleteEpoch(epoch uint64) error {
	// First receive all keys that were written by this block (entities & indices)
	startSlot := epoch * utils.Config.Chain.Config.SlotsPerEpoch
	endSlot := (epoch+1)*utils.Config.Chain.Config.SlotsPerEpoch - 1
	logger.Infof("deleting epoch %v (slot %v to %v)", epoch, startSlot, endSlot)

	filter := bson.D{{Key: "Epoch", Value: epoch}}
	res, err := mongodb.Db.Collection(BEACON_CHAIN).DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	logger.Infof("deleted %v documents\n", res.DeletedCount)
	return nil
}
