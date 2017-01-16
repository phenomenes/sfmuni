package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"gopkg.in/redis.v5"
)

const baseUrl = "http://webservices.nextbus.com/service/publicXMLFeed?command="

type Fetcher interface {
	Fetch(cmd string, u url.Values) (body []byte, err error)
}

type SomeFetcher struct {
	fn func(cmd string, u url.Values) (body []byte, err error)
}

// Get routes information
//	Command: agencyList
func (f SomeFetcher) AgencyListHandler(w http.ResponseWriter, r *http.Request) {
	client := context.Get(r, "redis.Client").(*redis.Client)

	body, err := getFromCache(client, r.URL.String())
	if err == nil {
		fmt.Fprintf(w, "%s", body)
		return
	}

	body, err = f.fn("agencyList", nil)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writeToCache(client, r.URL.String(), body)

	fmt.Fprintf(w, "%s", body)
}

// Get routes information
//	Command: routeList
//	Arguments: a=<agency_tag>
func (f SomeFetcher) RouteListHandler(w http.ResponseWriter, r *http.Request) {
	client := context.Get(r, "redis.Client").(*redis.Client)
	u := url.Values{}

	u.Set("a", "sf-muni")

	body, err := getFromCache(client, r.URL.String())
	if err == nil {
		fmt.Fprintf(w, "%s", body)
		return
	}

	body, err = f.fn("routeList", u)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeToCache(client, r.URL.String(), body)

	fmt.Fprintf(w, "%s", body)
}

// Get routes information
//	Command: routeConfig
//	Arguments: a=<agency_tag>
//		   r=<route_tag> (optional)
func (f SomeFetcher) RouteConfigHandler(w http.ResponseWriter, r *http.Request) {
	client := context.Get(r, "redis.Client").(*redis.Client)
	u := url.Values{}
	v := mux.Vars(r)

	u.Set("a", "sf-muni")

	if v["route_tag"] != "" {
		u.Set("r", v["route_tag"])
	}

	body, err := getFromCache(client, r.URL.String())
	if err == nil {
		fmt.Fprintf(w, "%s", body)
		return
	}

	body, err = f.fn("routeConfig", u)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeToCache(client, r.URL.String(), body)

	fmt.Fprintf(w, "%s", body)
}

// Get predictions for from a given stop (stop_id) or route (route_tag)
//	Command: predictions
//	Arguments: a=<agency_tag>
//	           stopId=<stopId>
//		   route_tag=<route_tag> (optional)
//		   useShortTitles=true (optional)
func PredictionsIdHandler(w http.ResponseWriter, r *http.Request) {
	client := context.Get(r, "redis.Client").(*redis.Client)
	u := url.Values{}
	v := mux.Vars(r)

	u.Set("a", "sf-muni")
	u.Set("stopId", v["stop_id"])

	if v["route_tag"] != "" {
		u.Set("routeTag", v["route_tag"])
	}

	if r.URL.Query().Get("short") == "true" {
		u.Set("useShortTitles", "true")
	}

	body, err := getFromCache(client, r.URL.String())
	if err == nil {
		fmt.Fprintf(w, "%s", body)
		return
	}

	body, err = fetch("predictions", u)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeToCache(client, r.URL.String(), body)

	fmt.Fprintf(w, "%s", body)
}

// Get predictions for a single route from a given stop (stopTag)
//	Command: predictions
//	Arguments: a=<agency_tag>
//		   r=<route tag>
//		   s=<stop tag>
//		   useShortTitles=true (optional)
func PredictionsTagHandler(w http.ResponseWriter, r *http.Request) {
	client := context.Get(r, "redis.Client").(*redis.Client)
	u := url.Values{}
	v := mux.Vars(r)

	u.Set("a", "sf-muni")
	u.Set("r", v["route_tag"])
	u.Set("s", v["stop_tag"])

	if r.URL.Query().Get("short") == "true" {
		u.Set("useShortTitles", "true")
	}

	body, err := getFromCache(client, r.URL.String())
	if err == nil {
		fmt.Fprintf(w, "%s", body)
		return
	}

	body, err = fetch("predictions", u)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeToCache(client, r.URL.String(), body)

	fmt.Fprintf(w, "%s", body)
}

// Get prediction for multiple stops/routes
// 	Command: predictionsForMultiStops
//	Arguments: a=<agency_tag>
//	           stops=<stop_1>..stops=<stop_N>
//		   useShortTitles=true (optional)
func PredictionsForMultiStopsHandler(w http.ResponseWriter, r *http.Request) {
	client := context.Get(r, "redis.Client").(*redis.Client)
	u := url.Values{}

	u.Set("a", "sf-muni")

	stops := strings.Split(mux.Vars(r)["stops"], ",")
	for n := range stops {
		u.Add("stops", stops[n])
	}

	if r.URL.Query().Get("short") == "true" {
		u.Set("useShortTitles", "true")
	}

	body, err := getFromCache(client, r.URL.String())
	if err == nil {
		fmt.Fprintf(w, "%s", body)
		return
	}

	body, err = fetch("predictionsForMultiStops", u)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeToCache(client, r.URL.String(), body)

	fmt.Fprintf(w, "%s", body)
}

// Get a route's schedule
//	Command: schedule
//	Argument: a=<agency_tag>
//		  r=<route_tag>
func ScheduleHandler(w http.ResponseWriter, r *http.Request) {
	client := context.Get(r, "redis.Client").(*redis.Client)
	u := url.Values{}
	v := mux.Vars(r)

	u.Set("a", "sf-muni")
	u.Set("r", v["route_tag"])

	body, err := getFromCache(client, r.URL.String())
	if err == nil {
		fmt.Fprintf(w, "%s", body)
		return
	}

	body, err = fetch("schedule", u)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeToCache(client, r.URL.String(), body)

	fmt.Fprintf(w, "%s", body)
}

// Get messages for all or a given route
//	Command: messages
//	Arguments: a=<agency tag>
//		   r=<route tag1>..r=<route tagN> (optional)
func MessagesHandler(w http.ResponseWriter, r *http.Request) {
	client := context.Get(r, "redis.Client").(*redis.Client)
	u := url.Values{}

	u.Set("a", "sf-muni")

	routeTags := r.URL.Query().Get("route_tags")
	if routeTags != "" {
		routes := strings.Split(routeTags, ",")
		for n := range routes {
			u.Add("r", routes[n])
		}
	}

	body, err := getFromCache(client, r.URL.String())
	if err == nil {
		fmt.Fprintf(w, "%s", body)
		return
	}

	body, err = fetch("messages", u)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeToCache(client, r.URL.String(), body)

	fmt.Fprintf(w, "%s", body)
}

// Get vehicles location
//	Command: vehicleLocations
//	Arguments: a=<agency_tag>
//		   r=<route tag>
//		   t=<epoch time in msec>
func VehicleLocationsHandler(w http.ResponseWriter, r *http.Request) {
	client := context.Get(r, "redis.Client").(*redis.Client)
	u := url.Values{}
	v := mux.Vars(r)

	u.Set("a", "sf-muni")
	u.Set("r", v["route_tag"])
	u.Set("t", v["epoch"])

	body, err := getFromCache(client, r.URL.String())
	if err == nil {
		fmt.Fprintf(w, "%s", body)
		return
	}

	body, err = fetch("vehicleLocations", u)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeToCache(client, r.URL.String(), body)

	fmt.Fprintf(w, "%s", body)
}

type endpoint struct {
	Name string `xml:"name,attr"`
	Hits string `xml:"hits,attr"`
}

type endpoints struct {
	XMLName xml.Name   `xml:"body"`
	List    []endpoint `xml:"endpoint"`
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	client := context.Get(r, "redis.Client").(*redis.Client)
	router := context.Get(r, "mux.Router").(*mux.Router)

	list := []endpoint{}
	em := make(map[string]string)

	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		if name := route.GetName(); name != "" {
			em[name] = "0"
		}

		return nil
	})

	for k, _ := range em {
		if val := client.Get(k).Val(); val != "" {
			em[k] = val
		}
		list = append(list, endpoint{Name: k, Hits: em[k]})
	}

	buf, _ := xml.MarshalIndent(&endpoints{List: list}, "", "  ")

	fmt.Fprintf(w, "%s%s\n", xml.Header, string(buf))
}

type query struct {
	Name string  `xml:"request,attr"`
	Time float64 `xml:"time,attr"`
}

type queries struct {
	XMLName xml.Name `xml:"body"`
	List    []query  `xml:"slow-queries"`
}

func SlowQueriesHandler(w http.ResponseWriter, r *http.Request) {
	client := context.Get(r, "redis.Client").(*redis.Client)

	// zrevrangebyscore slow-queries +inf -inf withscores
	vals, _ := client.ZRevRangeByScoreWithScores("slow-queries", redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}).Result()

	list := []query{}

	for n := range vals {
		list = append(list, query{Name: vals[n].Member.(string),
			Time: vals[n].Score})
	}

	buf, _ := xml.MarshalIndent(&queries{List: list}, "", "  ")

	fmt.Fprintf(w, "%s%s\n", xml.Header, string(buf))

}

type Stop struct {
	Tag       string `xml:"tag,attr"`
	TimeValue string `xml:",chardata"`
}

type Tr struct {
	BlockId string `xml:"blockID,attr"`
	Stops   []Stop `xml:"stop"`
}

type RouteSchedule struct {
	Trs          []Tr   `xml:"tr"`
	Tag          string `xml:"tag,attr"`
	ServiceClass string `xml:"serviceClass,attr"`
}

type Schedule struct {
	XMLName        xml.Name        `xml:"body"`
	RouteSchedules []RouteSchedule `xml:"route"`
}

type bus struct {
	Tag string `xml:"tag,attr"`
}

type notInService struct {
	XMLName xml.Name `xml:"body"`
	Buses   []bus    `xml:"bus"`
}

const (
	shortFormat = "15:05"
	longFormat  = "15:05:00"
)

// schedule
func NotInServiceHandler(w http.ResponseWriter, r *http.Request) {
	u := url.Values{}
	v := mux.Vars(r)
	f := &SomeFetcher{fn: fetch}
	s := Schedule{}
	today := getWeekDay()
	buses := make(map[string]bool)

	u.Set("a", "sf-muni")
	routes := f.getRouteList(w, r).Routes
	schedule, _ := time.Parse(shortFormat, v["time"])

	for i := range routes {
		u.Set("r", routes[i].Tag)

		body, err := fetch("schedule", u)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}

		err = xml.Unmarshal([]byte(body), &s)
		if err != nil {
			panic(err)
		}

		log.Printf("Get schedule for route %s\n", routes[i].Tag)

		for route := range s.RouteSchedules {
			if today != s.RouteSchedules[route].ServiceClass {
				continue
			}

			for block := range s.RouteSchedules[route].Trs {
				first, _ := time.Parse(longFormat,
					findFirst(s.RouteSchedules[route].Trs[block].Stops))
				last, _ := time.Parse(longFormat,
					findLast(s.RouteSchedules[route].Trs[block].Stops))

				buses[routes[i].Tag] = false

				if first.Before(last) {
					if schedule.After(first) && schedule.Before(last) {
						buses[routes[i].Tag] = true
					}
				} else {
					if schedule.After(first) || schedule.Before(last) {
						buses[routes[i].Tag] = true
					}
				}
			}
		}
	}

	list := []bus{}

	for k, v := range buses {
		if v == false {
			list = append(list, bus{Tag: k})
		}
	}

	buf, err := xml.MarshalIndent(&notInService{Buses: list}, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(w, "%s%s\n", xml.Header, string(buf))
}

type Route struct {
	Tag string `xml:"tag,attr"`
}

type RouteList struct {
	XMLName xml.Name `xml:"body"`
	Routes  []Route  `xml:"route"`
}

func (f SomeFetcher) getRouteList(w http.ResponseWriter, r *http.Request) RouteList {
	u := url.Values{}

	u.Set("a", "sf-muni")

	body, _ := f.fn("routeList", u)
	rl := RouteList{}

	err := xml.Unmarshal([]byte(body), &rl)
	if err != nil {
		panic(err)
	}

	return rl
}

func getWeekDay() (today string) {
	wd := int(time.Now().Weekday())

	switch {
	case wd >= 0 || wd <= 4:
		return "wkd"
	case wd == 5:
		return "sat"
	default:
		return "sunday"
	}
}

func findFirst(stops []Stop) string {
	for i := range stops {
		t := stops[i].TimeValue
		if t != "--" {

			return t
		}
	}

	return ""
}

func findLast(stops []Stop) string {
	for i := len(stops) - 1; i >= 0; i-- {
		t := stops[i].TimeValue
		if t != "--" {

			return t
		}
	}

	return ""
}

func fetch(cmd string, u url.Values) ([]byte, error) {
	var uri string

	if u == nil {
		uri = fmt.Sprintf("%s%s", baseUrl, cmd)
	} else {
		uri = fmt.Sprintf("%s%s&%s", baseUrl, cmd, u.Encode())
	}

	//fmt.Println(uri)
	//client := &http.Client{}

	//req, err := http.NewRequest("GET", url, nil)
	//req.Header.Add("Accept-Encoding", "gzip")
	//resp, err := client.Do(req)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//_, err := io.Copy(os.Stdout, resp.Body)

	return body, nil
}

func Middleware(router *mux.Router, client *redis.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var match mux.RouteMatch

		var matched bool
		if router.Match(r, &match) {
			if err := client.Incr(match.Route.GetName()).Err(); err != nil {
				log.Printf("%s", err)
			}
			matched = true
		}

		start := time.Now()

		router.ServeHTTP(w, r)

		if matched {
			_ = client.ZAdd("slow-queries",
				redis.Z{float64(time.Since(start) / time.Millisecond),
					r.URL.Path},
			)
		}

		log.Printf("- %s %s \"%s\" %v",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			time.Since(start),
		)

	})
}

func withRedis(client *redis.Client, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		context.Set(r, "redis.Client", client)
		handler(w, r)
	}
}

func withRouter(router *mux.Router, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		context.Set(r, "mux.Router", router)
		handler(w, r)
	}
}
