func tokenBucket(client *redis.Client, key string, tokens int, refillInterval time.Duration, bucketSize int) bool {
	now := time.Now().UnixNano()

	// Lua script for token bucket rate limiter
	luaScript := `
	local key = KEYS[1]
	local tokens = tonumber(ARGV[1])
	local refillInterval = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])
	local bucketSize = tonumber(ARGV[4])
	
	local bucket = redis.call("hmget", key, "tokens", "lastRefill")
	local tokensInBucket = tonumber(bucket[1])
	local lastRefill = tonumber(bucket[2])
	
	if lastRefill == nil then
	  tokensInBucket = tokens
	  lastRefill = now
	end
	
	local elapsedTime = now - lastRefill
	
	if elapsedTime >= refillInterval then
	  local tokensToAdd = math.floor(elapsedTime / refillInterval)
	  tokensInBucket = math.min(tokensInBucket + tokensToAdd, bucketSize)
	  lastRefill = now
	end
	
	if tokensInBucket >= 1 then
	  tokensInBucket = tokensInBucket - 1
	  redis.call("hmset", key, "tokens", tokensInBucket, "lastRefill", lastRefill)
	  redis.call("expire", key, math.ceil(refillInterval / 1000000000))
	  return 1
	else
	  return 0
	end
`
	// Load and run Lua script
	script := redis.NewScript(luaScript)
	result, err := script.Run(client, []string{key}, tokens, refillInterval.Nanoseconds(), now, bucketSize).Result()
	if err != nil {
		fmt.Printf("Failed to execute Lua script: %v\n", err)
		return false
	}

	return result == int64(1)
}

func main() {
	// Initialize the Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Example usage of the tokenBucket function
	if tokenBucket(client, "my-token-bucket", 10, time.Minute, 10) {
		// request was allowed
		fmt.Println("Request allowed")
	} else {
		// request was rate limited
		fmt.Println("Request rate limited")
	}
}