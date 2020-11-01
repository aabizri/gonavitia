package navitia

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type query interface {
	toURL() (url.Values, error)
}

// results is implemented by every Result type
type results interface {
	creating()
	sending()
	parsing()
}

// requestURL requests a url, with the query already encoded in, and decodes the result in res.
func (s *Session) requestURL(ctx context.Context, url string, res results) error {
	// Store creation time
	res.creating()

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.Wrapf(err, "couldn't create new request (for %s)", url)
	}

	// Add basic auth
	req.SetBasicAuth(s.APIKey, "")

	// Execute the request
	resp, err := s.client.Do(req)
	res.sending()

	// Check the response
	if err != nil {
		return errors.Wrap(err, "error while executing request")
	}
	if resp.StatusCode != http.StatusOK {
		return parseRemoteError(resp)
	}

	// Defer the close
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Limit the reader
	reader := io.LimitReader(resp.Body, maxSize)

	// Parse the now limited body
	dec := json.NewDecoder(reader)
	err = dec.Decode(res)
	if err != nil {
		return errors.Wrap(err, "JSON decoding failed")
	}
	res.parsing()

	// Return
	return err
}

// request does a request given a url, query and results to populate
func (s *Session) request(ctx context.Context, baseURL string, query query, res results) error {
	// Encode the parameters
	values, err := query.toURL()
	if err != nil {
		return errors.Wrap(err, "error while retrieving url values to be encoded")
	}
	reqURL := baseURL + "?" + values.Encode()

	// Call requestURL
	return s.requestURL(ctx, reqURL, res)
}
