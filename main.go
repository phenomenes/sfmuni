package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var redisHostPort, port string

func main() {
	flag.StringVar(&port, "port", "8080", "Listen port")
	flag.Parse()

	c := newRedisClient()

	f := &SomeFetcher{
		fn: fetch,
	}

	r := mux.NewRouter()
	s := r.PathPrefix("/v1").
		Methods("GET").
		Subrouter()

	s.HandleFunc("/agency-list", withRedis(c, f.AgencyListHandler)).
		Name("agency-list")
	s.HandleFunc("/route-list", withRedis(c, f.RouteListHandler)).
		Name("route-list")
	s.HandleFunc("/route-config", withRedis(c, f.RouteConfigHandler)).
		Name("route-config")
	s.HandleFunc("/route-config/{route_tag}", withRedis(c, f.RouteConfigHandler)).
		Name("route-config")
	s.HandleFunc("/predictions-id/{stop_id}", withRedis(c, PredictionsIdHandler)).
		Name("predictions-id")
	s.HandleFunc("/predictions-id/{stop_id}/{route_tag}",
		withRedis(c, PredictionsIdHandler)).
		Name("predictions-id")
	s.HandleFunc("/predictions-tag/{route_tag}/{stop_tag}",
		withRedis(c, PredictionsTagHandler)).
		Name("predictions-tag")
	s.HandleFunc("/predictions-multi", withRedis(c, PredictionsForMultiStopsHandler)).
		Queries("stops", "{stops}").
		Name("predictions-multi")
	s.HandleFunc("/schedule/{route_tag}", withRedis(c, ScheduleHandler)).
		Name("schedule")
	s.HandleFunc("/messages", withRedis(c, MessagesHandler)).
		Name("messages")
	s.HandleFunc("/vehicle-locations/{route_tag}/{epoch}",
		withRedis(c, VehicleLocationsHandler)).
		Name("vehicle-locations")
	s.HandleFunc("/not-in-service/{time:[0-9]+:[0-9]+}",
		withRedis(c, NotInServiceHandler)).
		Name("not-in-service")
	s.HandleFunc("/stats", withRedis(c, withRouter(r, StatsHandler))).
		Name("stats")
	s.HandleFunc("/slow-queries", withRedis(c, SlowQueriesHandler)).
		Name("slow-queries")

	log.Fatal(http.ListenAndServe(":"+port, Middleware(r, c)))
}
