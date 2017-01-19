# sfmuni XML RESTful API

This package wraps the NextBus XML API feed and exposes RESTful endpoints.
It consist of an app container `phenomenes/sfmuni` and a backend `redis` for
caching requests.
The `kubebernetes.yml` definition provides three services:

  - api
  - redis-master
  - redis-slave

## Requirements

- Kubernetes 1.4
- Kubernetes dns service
- go 1.7
- kubectl v1.5.1

## Run

```
go get github.com/phenomenes/sfmuni
cd $GOPATH/src/github.com/phenomenes/sfmuni
kubectl -f ./kubernetes.yml
```

You can check if the pods were successfully created by querying the Kubernetes API

```
kubectl get pods
```

The `api` service is exposed on port 8080, in order to access the api you need
to get the api service url/ip address:

```
kubectl get service api
```

## API Endpoints

| Method | Endpoint                                              | Usage                                        |
|--------|-------------------------------------------------------|----------------------------------------------|
| GET    | /v1/agency-list                                       | Get agencies                                 |
| GET    | /v1/route-list                                        | Get routes                                   |
| GET    | /v1/route-config                                      | Get route configuration                      |
| GET    | /v1/route-config/{route_tag}                          | Get route configuration for route_tag        |
| GET    | /v1/predictions-id/{stop_id}                          | Get predictions for stop_id                  |
| GET    | /v1/predictions-id/{stop_id}/{route_tag}              | Get predictions for stop_id and route_tag    |
| GET    | /v1/predictions-tag/{route_tag}/{stop_tag}            | Get prediction for route_tag and stop_tag    |
| GET    | /v1/predictions-multi?stops={stop_id}|{route_tag},... | Get predictions for multiple stops           |
| GET    | /v1/schedule/{route_tag}                              | Get schedule for route                       |
| GET    | /v1/messages                                          | Get messages for all routes                  |
| GET    | /v1/messages?route_tags={route_tag},...               | Get messages for route_tag(s)                |
| GET    | /v1/vehicle-locations/{route_tag}/{epoch}             | Get vehicle locations for route_tag and time |
| GET    | /v1/stats                                             | Get API stats                                |
| GET    | /v1/slow-queries                                      | Get slow queries                             |
| GET    | /v1/not-in-service/{HH:mm}                            | Get routes not in service at HH:mm           |

# Examples URLs

```
http://127.0.0.1:8080/v1/agency-list
http://127.0.0.1:8080/v1/route-list
http://127.0.0.1:8080/v1/route-config
http://127.0.0.1:8080/v1/route-config/E
http://127.0.0.1:8080/v1/schedule/N
http://127.0.0.1:8080/v1/route-config/E
http://127.0.0.1:8080/v1/messages
http://127.0.0.1:8080/v1/messages?route_tags=49,45
http://127.0.0.1:8080/v1/vehicle-locations/N/0
http://127.0.0.1:8080/v1/predictions-multi?stops=N|3909,N|6997
http://127.0.0.1:8080/v1/stats
http://127.0.0.1:8080/v1/predictions-tag/E/5184
http://127.0.0.1:8080/v1/predictions-id/15205
http://127.0.0.1:8080/v1/predictions-id/15184/E
http://127.0.0.1:8080/v1/slow-queries
http://127.0.0.1:8080/v1/not-in-service/02:40
```

## TO-DO

- Persistent redis storage
- Unit tests
- Parameterised configuration
- Make NotInServiceHandler requests concurrent
- Cache NotInServiceHandler requests
- SSL support
- Implement JSON
