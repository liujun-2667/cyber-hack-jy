package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init() error {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	Client = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
		DB:   0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

func SetRoomState(roomID string, state interface{}, ttl time.Duration) error {
	ctx := context.Background()
	jsonData, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return Client.Set(ctx, "room:"+roomID, jsonData, ttl).Err()
}

func GetRoomState(roomID string, dest interface{}) error {
	ctx := context.Background()
	data, err := Client.Get(ctx, "room:"+roomID).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func DeleteRoomState(roomID string) error {
	ctx := context.Background()
	return Client.Del(ctx, "room:"+roomID).Err()
}

func AddToMatchQueue(queueName, playerID string, score float64) error {
	ctx := context.Background()
	return Client.ZAdd(ctx, "matchqueue:"+queueName, redis.Z{
		Score:  score,
		Member: playerID,
	}).Err()
}

func RemoveFromMatchQueue(queueName, playerID string) error {
	ctx := context.Background()
	return Client.ZRem(ctx, "matchqueue:"+queueName, playerID).Err()
}

func GetMatchQueueSize(queueName string) (int64, error) {
	ctx := context.Background()
	return Client.ZCard(ctx, "matchqueue:"+queueName).Result()
}

func GetMatchedPlayers(queueName string, count int, minScore, maxScore float64) ([]string, error) {
	ctx := context.Background()
	result, err := Client.ZRangeByScore(ctx, "matchqueue:"+queueName, &redis.ZRangeBy{
		Min:    strconv.FormatFloat(minScore, 'f', -1, 64),
		Max:    strconv.FormatFloat(maxScore, 'f', -1, 64),
		Count:  int64(count),
		Offset: 0,
	}).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

var _ = json.Marshal
