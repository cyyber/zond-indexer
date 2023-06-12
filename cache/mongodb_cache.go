package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Prajjawalk/zond-indexer/entity"
	"github.com/Prajjawalk/zond-indexer/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DATABASE           = "zondDb"
	TABLE_CACHE        = "cache"
	FAMILY_TEN_MINUTES = "10_min"
	FAMILY_ONE_HOUR    = "1_hour"
	FAMILY_ONE_DAY     = "1_day"
	COLUMN_DATA        = "d"
)

type MongodbCache struct {
	client *mongo.Client

	tableCache *mongo.Collection

	chainId string
}

func InitMongodbCache(client *mongo.Client, chainId string) *MongodbCache {
	bt := &MongodbCache{
		client:     client,
		tableCache: client.Database(DATABASE).Collection(TABLE_CACHE),
		chainId:    chainId,
	}

	return bt
}

func (cache *MongodbCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {

	family := FAMILY_TEN_MINUTES
	if expiration.Minutes() >= 60 {
		family = FAMILY_ONE_HOUR
	}
	if expiration.Hours() > 1 {
		family = FAMILY_ONE_DAY
	}

	inputCache := &entity.Cache{}
	inputCache.Expiration = primitive.DateTime(expiration)
	inputCache.Key = []byte(fmt.Sprintf("C:%s", key))
	inputCache.Type = family
	valueMarshal, err := json.Marshal(value)
	if err != nil {
		return err
	}
	inputCache.Value = valueMarshal

	doc, err := utils.ToDoc(inputCache)
	if err != nil {
		return err
	}

	_, err = cache.tableCache.InsertOne(ctx, doc)
	if err != nil {
		return err
	}

	return nil
}

func (cache *MongodbCache) setByte(ctx context.Context, key string, value []byte, expiration time.Duration) error {

	family := FAMILY_TEN_MINUTES
	if expiration.Minutes() >= 60 {
		family = FAMILY_ONE_HOUR
	}
	if expiration.Hours() > 1 {
		family = FAMILY_ONE_DAY
	}

	inputCache := &entity.Cache{}
	inputCache.Expiration = primitive.DateTime(expiration)
	inputCache.Key = []byte(fmt.Sprintf("C:%s", key))
	inputCache.Type = family
	inputCache.Value = value

	doc, err := utils.ToDoc(inputCache)
	if err != nil {
		return err
	}

	_, err = cache.tableCache.InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	return nil
}

func (cache *MongodbCache) SetString(ctx context.Context, key, value string, expiration time.Duration) error {
	return cache.setByte(ctx, key, []byte(value), expiration)
}

func (cache *MongodbCache) SetUint64(ctx context.Context, key string, value uint64, expiration time.Duration) error {
	return cache.setByte(ctx, key, ui64tob(value), expiration)
}

func (cache *MongodbCache) SetBool(ctx context.Context, key string, value bool, expiration time.Duration) error {
	return cache.setByte(ctx, key, booltob(value), expiration)
}

func (cache *MongodbCache) Get(ctx context.Context, key string, returnValue any) (any, error) {
	res, err := cache.getByte(ctx, key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(res), returnValue)
	if err != nil {
		// cache.remoteRedisCache.Del(ctx, key).Err()
		logrus.Errorf("error (bigtable_cache.go / Get) unmarshalling data for key %v: %v", key, err)
		return nil, err
	}

	return returnValue, nil
}

func (cache *MongodbCache) getByte(ctx context.Context, key string) ([]byte, error) {
	filter := bson.D{{Key: "key", Value: fmt.Sprintf("C:%s", key)}}
	var result *entity.Cache
	err := cache.tableCache.FindOne(ctx, filter, options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Value, nil
}

func (cache *MongodbCache) GetString(ctx context.Context, key string) (string, error) {

	res, err := cache.getByte(ctx, key)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func (cache *MongodbCache) GetUint64(ctx context.Context, key string) (uint64, error) {

	res, err := cache.getByte(ctx, key)
	if err != nil {
		return 0, err
	}

	return btoi64(res), nil
}

func (cache *MongodbCache) GetBool(ctx context.Context, key string) (bool, error) {

	res, err := cache.getByte(ctx, key)
	if err != nil {
		return false, err
	}

	return btobool(res), nil
}

func ui64tob(val uint64) []byte {
	r := make([]byte, 8)
	for i := uint64(0); i < 8; i++ {
		r[i] = byte((val >> (i * 8)) & 0xff)
	}
	return r
}

func btoi64(val []byte) uint64 {
	r := uint64(0)
	for i := uint64(0); i < 8; i++ {
		r |= uint64(val[i]) << (8 * i)
	}
	return r
}

func booltob(val bool) []byte {
	r := make([]byte, 1)
	if val {
		r[0] = 1
	} else {
		r[0] = 0
	}
	return r
}

func btobool(val []byte) bool {
	return val[0] == 1
}
