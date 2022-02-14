# Docker-Healthchecks

To use this package, run:
```go
go get github.com/obbap1/docker-healthchecks
```

This package exposes two functions. <br>
`ListNetworks` - returns all the names of your docker networks and their IDs. the network names are namespaced, but it doesnt matter, you can pass in the full name into the `InspectNetworkByName` function. 
<br>

`InspectNetworkByName` - returns all the containers on the network and their health status.

## Eg. 

```go
import (
    "fmt"
    d "github.com/obbap1/docker-healthchecks"
)

networks, err := d.ListNetworks()
if err == nil {
    for network := range networks {
        c, err := d.InspectNetworkByName(network)
        if err == nil {
            fmt.Printf("\n %+s, %+v", network, c)
        }
    }
}
```

## Sample response 
```go
[{name:database status:nil failingStreak:0} {name:cache status:healthy failingStreak:0}]
```