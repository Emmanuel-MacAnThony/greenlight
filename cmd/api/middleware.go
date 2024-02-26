package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Emmanuel-MacAnThony/greenlight/internal/data"
	"github.com/Emmanuel-MacAnThony/greenlight/internal/validator"
	"github.com/felixge/httpsnoop"
	"github.com/tomasen/realip"
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

			// ip, _, err := net.SplitHostPort(request.RemoteAddr)
			// if err != nil {
			// 	app.serverErrorResponse(response, request, err)
			// 	return
			// }

			ip := realip.FromRequest(request)

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

func (app *application) authenticate(next http.Handler) http.Handler {

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

		response.Header().Add("Vary", "Authorization")

		authorizationHeader := request.Header.Get("Authorization")

		if authorizationHeader == "" {
			request = app.contextSetUser(request, data.AnonymousUser)
			next.ServeHTTP(response, request)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(response, request)
			return
		}

		token := headerParts[1]

		v := validator.New()

		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(response, request)
			return
		}

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(response, request)
			default:
				app.serverErrorResponse(response, request, err)
			}

			return
		}

		request = app.contextSetUser(request, user)

		next.ServeHTTP(response, request)
	})
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		user := app.contextGetUser(request)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(response, request)
			return
		}
		next.ServeHTTP(response, request)
	})
}

func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {

	fn := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

		user := app.contextGetUser(request)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(response, request)
			return
		}

		if !user.Activated {
			app.inactiveAccountResponse(response, request)
			return
		}

		next.ServeHTTP(response, request)

	})

	return app.requireAuthenticatedUser(fn)
}

func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {

	fn := func(response http.ResponseWriter, request *http.Request) {

		user := app.contextGetUser(request)

		permissions, err := app.models.Permissions.GetAllForUser(user.ID)

		if err != nil {
			app.serverErrorResponse(response, request, err)
			return
		}

		if !permissions.Include(code) {
			app.notPermittedResponse(response, request)
			return
		}

		next.ServeHTTP(response, request)

	}

	return app.requireActivatedUser(fn)

}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

		response.Header().Add("Vary", "Origin")
		response.Header().Add("Vary", "Access-Control-Request-Method")

		origin := request.Header.Get("Origin")

		if origin != "" {

			for i := range app.config.cors.trustedOrigins {

				if origin == app.config.cors.trustedOrigins[i] {
					response.Header().Set("Access-Control-Allow-Origin", origin)

					if request.Method == http.MethodOptions && request.Header.Get("Access-Control-Request-Method") != "" {

						response.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						response.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						response.WriteHeader(http.StatusOK)
						return
					}

					break
				}
			}
		}

		next.ServeHTTP(response, request)

	})
}

func (app *application) metrics(next http.Handler) http.Handler {

	totalRequestReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_μs")
	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

		//start := time.Now()
		totalRequestReceived.Add(1)
		metrics := httpsnoop.CaptureMetrics(next, response, request)

		//next.ServeHTTP(response, request)

		totalResponsesSent.Add(1)
		//duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())
		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)

	})
}
