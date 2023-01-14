package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/web-stuff-98/go-social-media/pkg/db"
)

type BlockInfo struct {
	LastRequest      time.Time `json:"last_request"`
	RequestsInWindow uint16    `json:"reqs_in_window"`
}

type SimpleLimiterOpts struct {
	Window        time.Duration `json:"window"`
	MaxReqs       uint16        `json:"max_reqs"`
	BlockDuration time.Duration `json:"block_dur"`
	Message       string        `json:"msg"`
	RouteName     string        `json:"-"`
}

func errMsg(w http.ResponseWriter, s int, m string) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(s)
	json.NewEncoder(w).Encode(map[string]string{
		"msg": m,
	})
}

func BasicRateLimiter(next http.HandlerFunc, opts SimpleLimiterOpts, rdb redis.Client, colls db.Collections) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Find IP block info on redis
		ipInfoKey := "SIMPLE-LIMITER-INFO=" + r.RemoteAddr + "=" + opts.RouteName
		ipInfoCmd := rdb.Get(r.Context(), ipInfoKey)
		ipInfo := &BlockInfo{}
		if ipInfoCmd.Err() == nil {
			// IP block info for route found
			ipInfoString := ipInfoCmd.Val()
			err := json.Unmarshal([]byte(ipInfoString), ipInfo)
			if err != nil {
				errMsg(w, http.StatusInternalServerError, "Internal error")
				return
			}
			// Check if blocked
			if ipInfo.RequestsInWindow >= opts.MaxReqs {
				if time.Now().After(ipInfo.LastRequest.Add(opts.BlockDuration)) {
					// The IP was blocked, but is now no longer blocked, so delete the block info and do next
					delCmd := rdb.Del(r.Context(), ipInfoKey)
					if delCmd.Err() != nil {
						if delCmd.Err() != redis.Nil {
							errMsg(w, http.StatusInternalServerError, "Internal error")
							return
						}
					}
					next.ServeHTTP(w, r)
					return
				} else {
					// The IP is blocked, extend redis key expiration to end of block duration and send err msg
					secsRemaining := ipInfo.LastRequest.Add(opts.BlockDuration).Second() - time.Now().Second()
					setCmd := rdb.Set(r.Context(), ipInfoKey, ipInfoString, time.Duration(secsRemaining*1000000000))
					if setCmd.Err() != nil {
						errMsg(w, http.StatusInternalServerError, "Internal error")
						return
					}
					var msg string
					if opts.Message != "" {
						msg = opts.Message
					} else {
						msg = "Too many requests"
					}
					errMsg(w, http.StatusTooManyRequests, msg)
					return
				}
			}
			// If not blocked add to the number of requests
			if ipInfo.LastRequest.Before(time.Now().Add(-opts.Window)) {
				ipInfo.RequestsInWindow = 1
			} else {
				ipInfo.RequestsInWindow++
			}
			ipInfo.LastRequest = time.Now()
			ipInfoBytes, err := json.Marshal(ipInfo)
			if err != nil {
				errMsg(w, http.StatusInternalServerError, "Internal error")
				return
			}
			// Set the ip block info
			setCmd := rdb.Set(r.Context(), ipInfoKey, string(ipInfoBytes), opts.Window)
			if setCmd.Err() != nil {
				errMsg(w, http.StatusInternalServerError, "Internal error")
				return
			}
		} else {
			// IP block info for route was not found, set it, but first check if its because there was an internal error
			if ipInfoCmd.Err() != redis.Nil {
				errMsg(w, http.StatusInternalServerError, "Internal error")
				return
			} else {
				// No internal error, create IP block info on redis and do next
				ipInfo = &BlockInfo{
					LastRequest:      time.Now(),
					RequestsInWindow: 1,
				}
				ipInfoBytes, err := json.Marshal(ipInfo)
				if err != nil {
					errMsg(w, http.StatusInternalServerError, "Internal error")
					return
				}
				// Set the ip block info
				setCmd := rdb.Set(r.Context(), ipInfoKey, string(ipInfoBytes), opts.Window)
				if setCmd.Err() != nil {
					errMsg(w, http.StatusInternalServerError, "Internal error")
					return
				}
				next.ServeHTTP(w, r)
			}
		}
	})
}
