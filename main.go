package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/amleshkashyap/limiter/middleware"
	"github.com/amleshkashyap/limiter/rules"
	"time"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		fmt.Println("Time Elapsed: ", time.Since(t))
		fmt.Println(c.Writer.Status())
	}
}

func Limiter() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		rateLimiter, rule := middleware.RateLimiterMiddleware(ginCtx)
		if rateLimiter != nil {
			rateLimiter.Handle(ginCtx, rule)
		} else {
			ginCtx.Next()
		}
	}
}

func main() {
	middleware.InitRedisClient()
	storedRules := rules.FetchRules()
	middleware.InitRateLimiters(storedRules)
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(Logger())
	r.Use(Limiter())
	r.GET("/resource", func(c *gin.Context) {
		fmt.Println("Done")
		c.JSON(200, gin.H{"msg": "Success"})
		return
	})
	r.Run(":9005")
}
