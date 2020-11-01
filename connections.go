package navitia

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"
	"unsafe"

	"github.com/pkg/errors"

	"github.com/govitia/navitia/types"
)

// A Connection is either a Departure or an Arrival
type Connection struct {
	Display   types.Display
	StopPoint types.StopPoint
	Route     types.Route
	// StopDateTime
}

// ConnectionsResults holds the results of a departures or arrivals request.
type ConnectionsResults struct {
	Connections []Connection

	Paging Paging `json:"links"`

	Logging `json:"-"`
}

// UnmarshalJSON implements unmarshalling for ConnectionsResults.
func (cr *ConnectionsResults) UnmarshalJSON(b []byte) error {
	// First let's create the analogous structure
	// We define some of the value as pointers to the real values, allowing us to bypass copying in cases where we don't need to process the data
	data := &struct {
		// Pointers to the corresponding real values
		Paging *Paging `json:"links"`

		// Value to process
		Departures *[]Connection `json:"departures"`
		Arrivals   *[]Connection `json:"arrivals"`
	}{
		Paging: &cr.Paging,
	}

	// Now unmarshall the raw data into the analogous structure
	err := json.Unmarshal(b, data)
	if err != nil {
		return errors.Wrap(err, "ConnectionsResults.UnmarshalJSON: error while unmarshalling Line")
	}

	// Now process the values
	switch {
	case data.Departures != nil:
		cr.Connections = *data.Departures
	case data.Arrivals != nil:
		cr.Connections = *data.Arrivals
	}
	// else there's nor Departures nor Arrivals found

	return nil
}

// ConnectionsRequest contains the optional parameters for a Departures request.
type ConnectionsRequest struct {
	// From what time on do you want to see the results ?
	From time.Time

	// Maximum duration between From and the retrieved results.
	//
	// Default value is 24 hours
	Duration time.Duration

	// The maximum amount of results
	//
	// Default value is 10 results
	Count uint

	// ForbiddenURIs
	Forbidden []types.ID

	// Freshness of the data
	Freshness types.DataFreshness

	// Enables GeoJSON data in the reply. GeoJSON objects can be VERY large ! >1MB.
	Geo bool
}

func (req ConnectionsRequest) toURL() (url.Values, error) {
	values := url.Values{}

	if datetime := req.From; !datetime.IsZero() {
		str := datetime.Format(types.DateTimeFormat)
		values.Add("datetime", str)
	}

	// If count is defined don't bother with the minimimal and maximum amount of items to return
	if count := req.Count; count != 0 {
		countStr := strconv.FormatUint(uint64(count), 10)
		values.Add("count", countStr)
	}

	// Deal with the forbidden URIs
	if forbidden := req.Forbidden; len(forbidden) != 0 {
		magic := *(*[]string)(unsafe.Pointer(&forbidden))
		values["forbidden_uris[]"] = magic
	}

	// Set the freshness
	if freshness := req.Freshness; freshness != "" {
		values.Add("data_freshness", string(freshness))
	}

	// Add GEO
	if !req.Geo {
		values.Add("disable_geojson", "true")
	}

	return values, nil
}

// departures is the internal function used by Departures & Arrivals functions
func (s *Session) connections(ctx context.Context, url string, req ConnectionsRequest) (*ConnectionsResults, error) {
	var results = &ConnectionsResults{}
	err := s.request(ctx, url, req, results)
	return results, err
}

const (
	departuresEndpoint string = "departures"
	arrivalsEndpoint          = "arrivals"
)

// DeparturesSA requests the departures for a given StopArea
func (scope *Scope) DeparturesSA(ctx context.Context, req ConnectionsRequest, resource types.ID) (*ConnectionsResults, error) {
	// Create the URL
	scopeURL := scope.session.APIURL + "/coverage/" + string(scope.region) + "/stop_areas/" + string(resource) + "/" + departuresEndpoint

	return scope.session.connections(ctx, scopeURL, req)
}

// DeparturesSP requests the departures for a given StopPoint
func (scope *Scope) DeparturesSP(ctx context.Context, req ConnectionsRequest, resource types.ID) (*ConnectionsResults, error) {
	// Create the URL
	scopeURL := scope.session.APIURL + "/coverage/" + string(scope.region) + "/stop_points/" + string(resource) + "/" + departuresEndpoint

	return scope.session.connections(ctx, scopeURL, req)
}

// DeparturesC requests the departures from a point described by coordinates.
func (s *Session) DeparturesC(ctx context.Context, req ConnectionsRequest, coords types.Coordinates) (*ConnectionsResults, error) {
	// Create the URL
	coordsQ := string(coords.ID())
	scopeURL := s.APIURL + "/coverage/" + coordsQ + "/coords/" + coordsQ + "/" + departuresEndpoint

	return s.connections(ctx, scopeURL, req)
}

// ArrivalsSA requests the arrivals for a given StopArea in a given region.
func (scope *Scope) ArrivalsSA(ctx context.Context, req ConnectionsRequest, resource types.ID) (*ConnectionsResults, error) {
	// Create the URL
	scopeURL := scope.session.APIURL + "/coverage/" + string(scope.region) + "/stop_areas/" + string(resource) + "/" + arrivalsEndpoint

	return scope.session.connections(ctx, scopeURL, req)
}

// ArrivalsSP requests the arrivals for a given StopPoint in a given region.
func (scope *Scope) ArrivalsSP(ctx context.Context, req ConnectionsRequest, resource types.ID) (*ConnectionsResults, error) {
	// Create the URL
	scopeURL := scope.session.APIURL + "/coverage/" + string(scope.region) + "/stop_points/" + string(resource) + "/" + arrivalsEndpoint

	return scope.session.connections(ctx, scopeURL, req)
}

// ArrivalsC requests the arrivals from a point described by coordinates.
func (s *Session) ArrivalsC(ctx context.Context, req ConnectionsRequest, coords types.Coordinates) (*ConnectionsResults, error) {
	// Create the URL
	coordsQ := string(coords.ID())
	scopeURL := s.APIURL + "/coverage/" + coordsQ + "/coords/" + coordsQ + "/" + arrivalsEndpoint

	return s.connections(ctx, scopeURL, req)
}
