package middleware

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var slidingWindowScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = ARGV[4]

redis.call("ZREMRANGEBYSCORE", key, 0, now - window_ms)
redis.call("ZADD", key, now, member)
local count = redis.call("ZCARD", key)
redis.call("PEXPIRE", key, window_ms)

if count > limit then
	redis.call("ZREM", key, member)
	return 0
end

return 1
`)

type RateLimiter interface {
	Allow(ctx context.Context, ip string) (bool, error)
}

type RedisRateLimiter struct {
	client *redis.Client
	limit  int
	window time.Duration
	prefix string
}

func NewRedisRateLimiter(redisURL string, limit int, window time.Duration, prefix string) (*RedisRateLimiter, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	return &RedisRateLimiter{
		client: redis.NewClient(options),
		limit:  limit,
		window: window,
		prefix: prefix,
	}, nil
}

func (r *RedisRateLimiter) Close() error {
	return r.client.Close()
}

func (r *RedisRateLimiter) Allow(ctx context.Context, ip string) (bool, error) {
	now := time.Now().UnixMilli()
	member := fmt.Sprintf("%d-%d", now, time.Now().UnixNano())

	result, err := slidingWindowScript.Run(
		ctx,
		r.client,
		[]string{r.key(ip)},
		now,
		r.window.Milliseconds(),
		r.limit,
		member,
	).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func UnaryRateLimitInterceptor(limiter RateLimiter, logger *log.Logger) grpc.UnaryServerInterceptor {
	if logger == nil {
		logger = log.Default()
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if limiter == nil {
			return handler(ctx, req)
		}

		ip := clientIPFromContext(ctx)
		limiterCtx, cancel := withTimeout(ctx)
		allowed, err := limiter.Allow(limiterCtx, ip)
		cancel()
		if err != nil {
			logger.Printf("appointment rate limiter error for %q on %s: %v", ip, info.FullMethod, err)
			return handler(ctx, req)
		}

		if !allowed {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

func withTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 2*time.Second)
}

func clientIPFromContext(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok || p.Addr == nil {
		return "unknown"
	}

	host := p.Addr.String()
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return parsedHost
	}

	return strings.Trim(host, "[]")
}

func (r *RedisRateLimiter) key(ip string) string {
	return r.prefix + ":" + ip
}
