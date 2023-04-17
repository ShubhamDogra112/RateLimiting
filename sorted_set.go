//sorted set

func rateLimit(key string, maxCount int, duration time.Duration) bool {
    now := time.Now().Unix()
    client.ZRemRangeByScore(key, "0", strconv.FormatInt(now-int64(duration.Seconds()), 10))
    client.Expire(key, duration)

    count, _ := client.ZCard(key).Result()
    if count >= maxCount {
        return false
    }

    client.ZAdd(key, &redis.Z{
        Score:  float64(now),
        Member: strconv.FormatInt(now, 10),
    })
    return true
}

//atomic version
func rateLimit(client *redis.Client, key string, maxCount int, duration time.Duration) bool {
	now := time.Now().Unix()
	minScore := strconv.FormatInt(now-int64(duration.Seconds()), 10)

	// Create a new transaction pipeline
	pipeline := client.TxPipeline()

	// Add the commands to the pipeline
	pipeline.ZRemRangeByScore(key, "0", minScore)
	pipeline.Expire(key, duration)
	cardCmd := pipeline.ZCard(key)

	// Execute the pipeline
	_, err := pipeline.Exec()
	if err != nil {
		fmt.Printf("Failed to execute pipeline: %v\n", err)
		return false
	}

	count, err := cardCmd.Result()
	if err != nil {
		fmt.Printf("Failed to get ZCard result: %v\n", err)
		return false
	}

	if count >= maxCount {
		return false
	}

	// Add new request timestamp to the sorted set
	client.ZAdd(key, &redis.Z{
		Score:  float64(now),
		Member: strconv.FormatInt(now, 10),
	})

	return true
}

func main() {
	if rateLimit("my-key", 10, time.Minute) {
		// request was allowed
	} else {
		// request was rate limited
	}
}

