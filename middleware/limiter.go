package middleware

import (
	"fmt"
	"context"
	"time"
	"strconv"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/amleshkashyap/limiter/rules"
	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	Handle(ginCtx *gin.Context, rule rules.StoredRule)
	getRules()
	getKey(ginCtx *gin.Context)
	RateLimitExceeded(ginCtx *gin.Context)
}

type FixedWindowCounter struct {
	windowDuration time.Duration
	windowStart int64
}

var fwcLimiter RateLimiter;

func InitRateLimiters(storedRules rules.StoredRule) {
	v, err := json.Marshal(storedRules)
	if err != nil {
		fmt.Println("Failed to marshal")
	}
	err = RedisClient.Set(context.Background(), "stored_rules", v, 0).Err()
	if err != nil {
		fmt.Println("Coudn't set rules to redis", err)
	}
	fwcLimiter = FixedWindowCounter{windowDuration: time.Duration(60000000), windowStart: time.Now().Unix()}
}

// Blanket Error Handling Policy
// If anything goes wrong during rate limiting parsing - let the request pass - this is a disastrous policy
// Not allowing the requests is also disastrous - essentially, rate limiter can't go wrong - that's very hard to ensure - parsing bugs can be completely avoided or monitored but fetching from Redis?
func RateLimiterMiddleware(ginCtx *gin.Context) (RateLimiter, rules.StoredRule) {
	domain := ginCtx.Request.URL.Path
	query_params := ginCtx.Request.URL.Query()
	var storedRules rules.StoredRule

	v, err := RedisClient.Get(context.Background(), "stored_rules").Result()
	if err != nil {
		fmt.Println("Error While Fetching Rule")
		return nil, storedRules
	}

	json.Unmarshal([]byte(v), &storedRules)
	if err != nil {
		return nil, storedRules
	}

	fmt.Println("Found Rules", storedRules)

	if storedRules.Domain != domain {
		return nil, storedRules
	}

	for key, val := range query_params {
		if key == storedRules.Key && val[0] == storedRules.Value {
			return fwcLimiter, storedRules
		}
	}

	return nil, storedRules
}

// There's a scenario where the key has expired - in that case, initialize the key with count = 1
func (fwc FixedWindowCounter) Handle(ginCtx *gin.Context, rule rules.StoredRule) {
	key := fmt.Sprintf("%s:%s:%s", rule.Domain, rule.Key, rule.Value)
	fmt.Println("key: ", key)

	count, err := RedisClient.Get(context.Background(), key).Result()

	if err == redis.Nil {
		fmt.Println("Couldn't find in redis, setting")
                err := RedisClient.Set(context.Background(), key, 1, time.Duration(60000000000)).Err()
                if err != nil {
                        fmt.Println("Error while setting", err)
                }
	} else if err != nil {
		ginCtx.Next()
		return
	}

	v, _ := strconv.Atoi(count)
	if v > rule.MaxRequests {
		ginCtx.JSON(429, gin.H{"msg": "Too Many Request"})
		ginCtx.Abort()
		return
	}
	RedisClient.Incr(context.Background(), key)
	ginCtx.Next()
}

func (fwc FixedWindowCounter) getRules() {
}

func (fwc FixedWindowCounter) getKey(ginCtx *gin.Context) {
}

func (fwc FixedWindowCounter) RateLimitExceeded (ginCtx *gin.Context) {
}
