package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				response.Header().Set("Connection", "close")
				app.serverErrorResponse(response, request, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(response, request)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// a go routine which removes old entries from the client map
	go func() {

		for {
			time.Sleep(time.Minute)
			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}

	}()

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

		if app.config.limiter.enabled {

			ip, _, err := net.SplitHostPort(request.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(response, request, err)
				return
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(response, request)
				return
			}

			mu.Unlock()

			next.ServeHTTP(response, request)

		}

	})
}
