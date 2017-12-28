package cache

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	// "github.com/segmentio/stats/influxdb"
	// "github.com/segmentio/stats/grafana"
	// "github.com/segmentio/stats/prometheus"
	// "github.com/segmentio/stats/datadog"

	// "github.com/segmentio/events"
	// "github.com/segmentio/events/httpevents"
	// _ "github.com/segmentio/events/text"

	// "github.com/segmentio/stats"
	// "github.com/segmentio/stats/httpstats"
	// "github.com/segmentio/stats/procstats"
	// "github.com/segmentio/stats/iostats"
	// "github.com/segmentio/stats/netstats"
	// "github.com/segmentio/stats/redisstats"

	"github.com/sniperkit/logger"
)

var (
	etagKeys []string = []string{"ETag"}
	// statx    *stats.Engine
)

/*
	Refs:
	- https://github.com/mreiferson/go-httpclient
		- Snippets:
		```Go
		transport := &httpclient.Transport{
		    ConnectTimeout:        1*time.Second,
		    RequestTimeout:        10*time.Second,
		    ResponseHeaderTimeout: 5*time.Second,
		}
		defer transport.Close()

		client := &http.Client{Transport: transport}
		req, _ := http.NewRequest("GET", "http://127.0.0.1/test", nil)
		resp, err := client.Do(req)
		if err != nil {
		    return err
		}
		defer resp.Body.Close()
		# Note: you will want to re-use a single client object rather than creating one for each request, otherwise you will end up leaking connections.
		```
	- https://github.com/heatxsink/go-httprequest
	- https://github.com/ViBiOh/httputils/blob/master/request.go
*/

// Transport is an implementation of http.RoundTripper that will return values from a cache
// where possible (avoiding a network request) and will additionally add validators (etag/if-modified-since)
// to repeated requests allowing servers to return 304 / Not Modified
type Transport struct {
	// The RoundTripper interface actually used to make requests
	// If nil, http.DefaultTransport is used
	// httpstats.NewTransport(t)
	Transport http.RoundTripper
	Cache     Cache
	// If true, responses returned from the cache will be given an extra header, X-From-Cache
	MarkCachedResponses bool
	OnDuplicateAbort    bool
	Stats               bool
	StatsEngine         string
	StatsEndpoint       string
	Debug               bool
	Timeout             time.Duration
	Dial                time.Duration
	processedCount      int
	cachedCount         int
	conditionalCount    int
	// transport *httputil.StatsTransport
}

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

func NewTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(connectTimeout, readWriteTimeout),
		},
	}
}

/*
To have launchd start influxdb now and restart at login:
  brew services start influxdb
Or, if you don't want/need a background service you can just run:
  influxd -config /usr/local/etc/influxdb.conf

To have launchd start chronograf now and restart at login:
  brew services start chronograf
Or, if you don't want/need a background service you can just run:
  chronograf
*/

/*
var (
	// statsProc   = procstats.StartCollector(procstats.NewGoMetrics())
	// statsEngine *stats.Engine
	statsTags    []stats.Tag = []stats.Tag{}
	influxClient *influxdb.Client
	influxConfig influxdb.ClientConfig
	statsEngine  *stats.Engine
	statsProc    io.Closer
	// statsHTTP *stats
)
*/

// NewTransport returns a new Transport with the
// provided Cache implementation and MarkCachedResponses set to true
func NewTransport(c Cache) *Transport {

	// Make a new HTTP client with a transport that will report HTTP metrics,
	// set the engine to nil to use the default.
	/*
		httpc := &http.Client{
			Transport: httpstats.NewTransport(
				&http.Transport{},
			),
		}
	*/
	// ref. https://github.com/segmentio/stats#http-clients
	return &Transport{
		Cache:               c,
		MarkCachedResponses: true,
		// Transport:           httpstats.NewTransport(http.DefaultTransport),
		// Transport:           httpstats.NewTransportWith(statsEngine, http.DefaultClient.Transport),
	}
}

// Client returns an *http.Client that caches responses.
func (t *Transport) Client() *http.Client {
	return &http.Client{Transport: t}
	//return &http.Client{Transport: httpstats.NewTransportWith(statsEngine, http.DefaultClient.Transport)}
}

func (t *Transport) Info() {
	logx.InfoWithFields(logger.Fields{
		"Transport.MarkCachedResponses": t.MarkCachedResponses,
		"Transport.processedCount":      t.processedCount,
		"Transport.cachedCount":         t.cachedCount,
		"Transport.conditionalCount":    t.conditionalCount,
	}, "Transport.Info()")
	/*
		log.WithFields(log.Fields{
			"Transport.MarkCachedResponses": t.MarkCachedResponses,
			"Transport.processedCount":      t.processedCount,
			"Transport.cachedCount":         t.cachedCount,
			"Transport.conditionalCount":    t.conditionalCount,
		}).Info("Info()")
	*/
}

// RoundTrip takes a Request and returns a Response
//
// If there is a fresh Response already in cache, then it will be returned without connecting to
// the server.
//
// If there is a stale Response, then any validators it contains will be set on the new request
// to give the server a chance to respond with NotModified. If this happens, then the cached Response
// will be returned.
func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	// events.Log("bad registry URL: %{url}s", config.Registry)

	t.processedCount++
	cacheKey := cacheKey(req)
	// cacheable := (req.Method == "GET" || req.Method == "POST" || req.Method == "HEAD") && req.Header.Get("range") == ""
	cacheable := (req.Method == "GET" || req.Method == "HEAD") && req.Header.Get("Range") == ""
	var cachedResp *http.Response
	if cacheable {
		cachedResp, err = CachedResponse(t.Cache, req)
	} else {
		// Need to invalidate an existing value
		t.Cache.Delete(cacheKey)
	}

	transport := t.Transport
	if transport == nil {

		// var transport = http.DefaultTransport
		// transport = httpstats.NewTransport(transport)
		// transport = httpevents.NewTransport(transport)
		// http.DefaultTransport = transport
		// transport = httpstats.NewTransportWith(statsEngine, http.DefaultClient.Transport)
		transport = http.DefaultTransport
	}

	if cacheable && cachedResp != nil && err == nil {
		if t.MarkCachedResponses {
			cachedResp.Header.Set(XFromCache, "1")
			logx.DebugWithFields(logger.Fields{
				"cacheKey":     cacheKey,
				"X-From-Cache": req.Header.Get(XFromCache),
			}, "Transport.RoundTrip() MarkCachedResponses")
		}

		if varyMatches(cachedResp, req) {
			// Can only use cached value if the new request doesn't Vary significantly
			freshness := getFreshness(cachedResp.Header, req.Header)
			if freshness == fresh {
				logx.DebugWithFields(logger.Fields{
					"cacheKey":  cacheKey,
					"freshness": freshness,
				}, "Transport.RoundTrip() getFreshness")
				return cachedResp, nil
			}

			if freshness == stale {
				var req2 *http.Request
				// Add validators if caller hasn't already done so
				var etagKey, etag string
				for _, v := range etagKeys {
					etag = cachedResp.Header.Get(v)
					if etag != "" {
						etagKey = v
						logx.DebugWithFields(logger.Fields{
							"cacheKey":        cacheKey,
							"etagKey":         etagKey,
							"etag":            etag,
							"req.Header.etag": req.Header.Get(etagKey),
						}, "Transport.RoundTrip() ETAG Match.")
						break
					}
				}

				if etag != "" && req.Header.Get(etagKey) == "" {
					req2 = cloneRequest(req)
					req2.Header.Set("If-None-Match", etag)
					logx.DebugWithFields(logger.Fields{
						"cacheKey":                  cacheKey,
						"etag":                      etag,
						"etagKey":                   etagKey,
						"req.Header.etag":           req.Header.Get(etagKey),
						"req2.Header.If-None-Match": req2.Header.Get("If-None-Match"),
					}, "Transport.RoundTrip() cloneRequest, etag not empty in cachedResp, but header request was empty.")
				}
				lastModified := cachedResp.Header.Get("Last-Modified")
				if lastModified != "" && req.Header.Get("Last-Modified") == "" {
					if req2 == nil {
						logx.DebugWithFields(logger.Fields{
							"cacheKey":        cacheKey,
							"etagKey":         etagKey,
							"etag":            etag,
							"req.Header.etag": req.Header.Get(etagKey),
						}, "Transport.RoundTrip() cloneRequest, cloning request as it was nil.")
						req2 = cloneRequest(req)
					}
					req2.Header.Set("If-Modified-Since", lastModified)
				}
				if req2 != nil {
					req = req2
					logx.Debug("Transport.RoundTrip() cloneRequest success...")
				}
				logx.DebugWithFields(logger.Fields{
					"cacheKey":        cacheKey,
					"freshness":       freshness,
					"etag":            etag,
					"etagKey":         etagKey,
					"req.Header.etag": req.Header.Get(etagKey),
					"lastModified":    lastModified,
				}, "Transport.RoundTrip() ETAG")
			}
		}

		resp, err = transport.RoundTrip(req)

		logx.DebugWithFields(logger.Fields{
			"resp.StatusCode":        resp.StatusCode,
			"resp.Header.Length":     len(resp.Header),
			"req.Method":             req.Method,
			"http.StatusNotModified": http.StatusNotModified,
			"cacheKey":               cacheKey,
		}, "Transport.RoundTrip() transport.RoundTrip")

		if err == nil && req.Method == "GET" && resp.StatusCode == http.StatusNotModified {
			// Replace the 304 response with the one from cache, but update with some new headers
			endToEndHeaders := getEndToEndHeaders(resp.Header)
			for _, header := range endToEndHeaders {
				cachedResp.Header[header] = resp.Header[header]
			}
			// cachedResp.Status = fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK))
			// cachedResp.StatusCode = http.StatusOK
			resp = cachedResp

		} else if (err != nil || (cachedResp != nil && resp.StatusCode >= 500)) &&
			req.Method == "GET" && canStaleOnError(cachedResp.Header, req.Header) {
			// In case of transport failure and stale-if-error activated, returns cached content
			// when available
			// cachedResp.Status = fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK))
			// cachedResp.StatusCode = http.StatusOK
			return cachedResp, nil

		} else {
			if err != nil || resp.StatusCode != http.StatusOK {
				t.Cache.Delete(cacheKey)
			}
			if err != nil {
				return nil, err
			}
		}
	} else {
		reqCacheControl := parseCacheControl(req.Header)
		logx.DebugWithFields(logger.Fields{
			"cacheable":       cacheable,
			"cacheKey":        cacheKey,
			"reqCacheControl": reqCacheControl,
		}, "Transport.RoundTrip() reqCacheControl")
		if _, ok := reqCacheControl["Only-If-Cached"]; ok {
			resp = newGatewayTimeoutResponse(req)
		} else {
			resp, err = transport.RoundTrip(req)
			if err != nil {
				return nil, err
			}
		}

	}

	if cacheable && canStore(resp.StatusCode, parseCacheControl(req.Header), parseCacheControl(resp.Header)) {
		for _, varyKey := range headerAllCommaSepValues(resp.Header, "Vary") {
			varyKey = http.CanonicalHeaderKey(varyKey)
			fakeHeader := "X-Varied-" + varyKey
			reqValue := req.Header.Get(varyKey)
			if reqValue != "" {
				resp.Header.Set(fakeHeader, reqValue)
			}
		}
		logx.DebugWithFields(logger.Fields{
			"cacheKey":        cacheKey,
			"req.Method":      req.Method,
			"resp.StatusCode": resp.StatusCode,
		}, "Transport.RoundTrip() cachingReadCloser")

		switch req.Method {
		case "GET":
			// Delay caching until EOF is reached.
			resp.Body = &cachingReadCloser{
				R: resp.Body,
				OnEOF: func(r io.Reader) {
					resp := *resp
					resp.Body = ioutil.NopCloser(r)
					respBytes, err := httputil.DumpResponse(&resp, true)
					if err == nil {
						t.Cache.Set(cacheKey, respBytes)
					}
				},
			}
		default:
			respBytes, err := httputil.DumpResponse(resp, true)
			if err == nil {
				t.Cache.Set(cacheKey, respBytes)
			}
		}
	} else {
		logx.DebugWithFields(logger.Fields{
			"cacheable":   cacheable,
			"cacheKey":    cacheKey,
			"resp.Header": resp.Header,
		}, "Transport.RoundTrip() Delete")

		t.Cache.Delete(cacheKey)
	}
	return resp, nil
}

type ETagTransport struct {
	ETag string
}

func (t *ETagTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.ETag == "" {
		return nil, errors.New("t.ETag is empty")
	}

	// To set extra querystring params, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	req = cloneRequest(req)
	req.Header.Add("If-None-Match", t.ETag)

	// Make the HTTP request.
	resp, err := http.DefaultTransport.RoundTrip(req)

	logx.DebugWithFields(logger.Fields{
		"ETag": resp.Header.Get("ETag"),
	}, "ETagTransport.RoundTrip().")

	// log.Println("ETag: ", resp.Header.Get("ETag"))
	return resp, err
}
