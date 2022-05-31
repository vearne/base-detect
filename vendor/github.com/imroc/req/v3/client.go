package req

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/imroc/req/v3/internal/util"
	"golang.org/x/net/publicsuffix"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	urlpkg "net/url"
	"os"
	"strings"
	"time"
)

// DefaultClient returns the global default Client.
func DefaultClient() *Client {
	return defaultClient
}

// SetDefaultClient override the global default Client.
func SetDefaultClient(c *Client) {
	if c != nil {
		defaultClient = c
	}
}

var defaultClient *Client = C()

// Client is the req's http client.
type Client struct {
	BaseURL               string
	PathParams            map[string]string
	QueryParams           urlpkg.Values
	Headers               http.Header
	Cookies               []*http.Cookie
	FormData              urlpkg.Values
	DebugLog              bool
	AllowGetMethodPayload bool

	retryOption             *retryOption
	jsonMarshal             func(v interface{}) ([]byte, error)
	jsonUnmarshal           func(data []byte, v interface{}) error
	xmlMarshal              func(v interface{}) ([]byte, error)
	xmlUnmarshal            func(data []byte, v interface{}) error
	trace                   bool
	outputDirectory         string
	disableAutoReadResponse bool
	scheme                  string
	log                     Logger
	t                       *Transport
	t2                      *http2Transport
	dumpOptions             *DumpOptions
	httpClient              *http.Client
	beforeRequest           []RequestMiddleware
	udBeforeRequest         []RequestMiddleware
	afterResponse           []ResponseMiddleware
}

// R create a new request.
func (c *Client) R() *Request {
	return &Request{
		client:      c,
		retryOption: c.retryOption.Clone(),
	}
}

// SetCommonFormDataFromValues set the form data from url.Values for all requests
// which request method allows payload.
func (c *Client) SetCommonFormDataFromValues(data urlpkg.Values) *Client {
	if c.FormData == nil {
		c.FormData = urlpkg.Values{}
	}
	for k, v := range data {
		for _, kv := range v {
			c.FormData.Add(k, kv)
		}
	}
	return c
}

// SetCommonFormData set the form data from map for all requests
// which request method allows payload.
func (c *Client) SetCommonFormData(data map[string]string) *Client {
	if c.FormData == nil {
		c.FormData = urlpkg.Values{}
	}
	for k, v := range data {
		c.FormData.Set(k, v)
	}
	return c
}

// SetBaseURL set the default base URL, will be used if request URL is
// a relative URL.
func (c *Client) SetBaseURL(u string) *Client {
	c.BaseURL = strings.TrimRight(u, "/")
	return c
}

// SetOutputDirectory set output directory that response will
// be downloaded to.
func (c *Client) SetOutputDirectory(dir string) *Client {
	c.outputDirectory = dir
	return c
}

// SetCertFromFile helps to set client certificates from cert and key file.
func (c *Client) SetCertFromFile(certFile, keyFile string) *Client {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		c.log.Errorf("failed to load client cert: %v", err)
		return c
	}
	config := c.GetTLSClientConfig()
	config.Certificates = append(config.Certificates, cert)
	return c
}

// SetCerts set client certificates.
func (c *Client) SetCerts(certs ...tls.Certificate) *Client {
	config := c.GetTLSClientConfig()
	config.Certificates = append(config.Certificates, certs...)
	return c
}

func (c *Client) appendRootCertData(data []byte) {
	config := c.GetTLSClientConfig()
	if config.RootCAs == nil {
		config.RootCAs = x509.NewCertPool()
	}
	config.RootCAs.AppendCertsFromPEM(data)
	return
}

// SetRootCertFromString set root certificates from string.
func (c *Client) SetRootCertFromString(pemContent string) *Client {
	c.appendRootCertData([]byte(pemContent))
	return c
}

// SetRootCertsFromFile set root certificates from files.
func (c *Client) SetRootCertsFromFile(pemFiles ...string) *Client {
	for _, pemFile := range pemFiles {
		rootPemData, err := ioutil.ReadFile(pemFile)
		if err != nil {
			c.log.Errorf("failed to read root cert file: %v", err)
			return c
		}
		c.appendRootCertData(rootPemData)
	}
	return c
}

// GetTLSClientConfig return the underlying tls.Config.
func (c *Client) GetTLSClientConfig() *tls.Config {
	if c.t.TLSClientConfig == nil {
		c.t.TLSClientConfig = &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
		}
	}
	return c.t.TLSClientConfig
}

func (c *Client) defaultCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	if c.DebugLog {
		c.log.Debugf("<redirect> %s %s", req.Method, req.URL.String())
	}
	return nil
}

// SetRedirectPolicy set the RedirectPolicy which controls the behavior of receiving redirect
// responses (usually responses with 301 and 302 status code), see the predefined
// AllowedDomainRedirectPolicy, AllowedHostRedirectPolicy, MaxRedirectPolicy, NoRedirectPolicy,
// SameDomainRedirectPolicy and SameHostRedirectPolicy.
func (c *Client) SetRedirectPolicy(policies ...RedirectPolicy) *Client {
	if len(policies) == 0 {
		return c
	}
	c.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		for _, f := range policies {
			if f == nil {
				continue
			}
			err := f(req, via)
			if err != nil {
				return err
			}
		}
		if c.DebugLog {
			c.log.Debugf("<redirect> %s %s", req.Method, req.URL.String())
		}
		return nil
	}
	return c
}

// DisableKeepAlives disable the HTTP keep-alives (enabled by default)
// and will only use the connection to the server for a single
// HTTP request.
//
// This is unrelated to the similarly named TCP keep-alives.
func (c *Client) DisableKeepAlives() *Client {
	c.t.DisableKeepAlives = true
	return c
}

// EnableKeepAlives enables HTTP keep-alives (enabled by default).
func (c *Client) EnableKeepAlives() *Client {
	c.t.DisableKeepAlives = false
	return c
}

// DisableCompression disables the compression (enabled by default),
// which prevents the Transport from requesting compression
// with an "Accept-Encoding: gzip" request header when the
// Request contains no existing Accept-Encoding value. If
// the Transport requests gzip on its own and gets a gzipped
// response, it's transparently decoded in the Response.Body.
// However, if the user explicitly requested gzip it is not
// automatically uncompressed.
func (c *Client) DisableCompression() *Client {
	c.t.DisableCompression = true
	return c
}

// EnableCompression enables the compression (enabled by default).
func (c *Client) EnableCompression() *Client {
	c.t.DisableCompression = false
	return c
}

// SetTLSClientConfig set the TLS client config. Be careful! Usually
// you don't need this, you can directly set the tls configuration with
// methods like EnableInsecureSkipVerify, SetCerts etc. Or you can call
// GetTLSClientConfig to get the current tls configuration to avoid
// overwriting some important configurations, such as not setting NextProtos
// will not use http2 by default.
func (c *Client) SetTLSClientConfig(conf *tls.Config) *Client {
	c.t.TLSClientConfig = conf
	return c
}

// EnableInsecureSkipVerify enable send https without verifing
// the server's certificates (disabled by default).
func (c *Client) EnableInsecureSkipVerify() *Client {
	c.GetTLSClientConfig().InsecureSkipVerify = true
	return c
}

// DisableInsecureSkipVerify disable send https without verifing
// the server's certificates (disabled by default).
func (c *Client) DisableInsecureSkipVerify() *Client {
	c.GetTLSClientConfig().InsecureSkipVerify = false
	return c
}

// SetCommonQueryParams set URL query parameters with a map
// for all requests.
func (c *Client) SetCommonQueryParams(params map[string]string) *Client {
	for k, v := range params {
		c.SetCommonQueryParam(k, v)
	}
	return c
}

// AddCommonQueryParam add a URL query parameter with a key-value
// pair for all requests.
func (c *Client) AddCommonQueryParam(key, value string) *Client {
	if c.QueryParams == nil {
		c.QueryParams = make(urlpkg.Values)
	}
	c.QueryParams.Add(key, value)
	return c
}

func (c *Client) pathParams() map[string]string {
	if c.PathParams == nil {
		c.PathParams = make(map[string]string)
	}
	return c.PathParams
}

// SetCommonPathParam set a path parameter for all requests.
func (c *Client) SetCommonPathParam(key, value string) *Client {
	c.pathParams()[key] = value
	return c
}

// SetCommonPathParams set path parameters for all requests.
func (c *Client) SetCommonPathParams(pathParams map[string]string) *Client {
	m := c.pathParams()
	for k, v := range pathParams {
		m[k] = v
	}
	return c
}

// SetCommonQueryParam set a URL query parameter with a key-value
// pair for all requests.
func (c *Client) SetCommonQueryParam(key, value string) *Client {
	if c.QueryParams == nil {
		c.QueryParams = make(urlpkg.Values)
	}
	c.QueryParams.Set(key, value)
	return c
}

// SetCommonQueryString set URL query parameters with a raw query string
// for all requests.
func (c *Client) SetCommonQueryString(query string) *Client {
	params, err := urlpkg.ParseQuery(strings.TrimSpace(query))
	if err != nil {
		c.log.Warnf("failed to parse query string (%s): %v", query, err)
		return c
	}
	if c.QueryParams == nil {
		c.QueryParams = make(urlpkg.Values)
	}
	for p, v := range params {
		for _, pv := range v {
			c.QueryParams.Add(p, pv)
		}
	}
	return c
}

// SetCommonCookies set HTTP cookies for all requests.
func (c *Client) SetCommonCookies(cookies ...*http.Cookie) *Client {
	c.Cookies = append(c.Cookies, cookies...)
	return c
}

// DisableDebugLog disable debug level log (disabled by default).
func (c *Client) DisableDebugLog() *Client {
	c.DebugLog = false
	return c
}

// EnableDebugLog enable debug level log (disabled by default).
func (c *Client) EnableDebugLog() *Client {
	c.DebugLog = true
	return c
}

// DevMode enables:
// 1. Dump content of all requests and responses to see details.
// 2. Output debug level log for deeper insights.
// 3. Trace all requests, so you can get trace info to analyze performance.
// 4. Set User-Agent to pretend to be a web browser, avoid returning abnormal data from some sites.
func (c *Client) DevMode() *Client {
	return c.EnableDumpAll().
		EnableDebugLog().
		EnableTraceAll().
		SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36")
}

// SetScheme set the default scheme for client, will be used when
// there is no scheme in the request URL (e.g. "github.com/imroc/req").
func (c *Client) SetScheme(scheme string) *Client {
	if !util.IsStringEmpty(scheme) {
		c.scheme = strings.TrimSpace(scheme)
	}
	return c
}

// SetLogger set the customized logger for client, will disable log if set to nil.
func (c *Client) SetLogger(log Logger) *Client {
	if log == nil {
		c.log = &disableLogger{}
		return c
	}
	c.log = log
	return c
}

func (c *Client) getResponseOptions() *ResponseOptions {
	if c.t.ResponseOptions == nil {
		c.t.ResponseOptions = &ResponseOptions{}
	}
	return c.t.ResponseOptions
}

// SetTimeout set timeout for all requests.
func (c *Client) SetTimeout(d time.Duration) *Client {
	c.httpClient.Timeout = d
	return c
}

func (c *Client) getDumpOptions() *DumpOptions {
	if c.dumpOptions == nil {
		c.dumpOptions = newDefaultDumpOptions()
	}
	return c.dumpOptions
}

// EnableDumpAll enable dump for all requests, including
// all content for the request and response by default.
func (c *Client) EnableDumpAll() *Client {
	if c.t.dump != nil { // dump already started
		return c
	}
	c.t.EnableDump(c.getDumpOptions())
	return c
}

// EnableDumpAllToFile enable dump for all requests and output
// to the specified file.
func (c *Client) EnableDumpAllToFile(filename string) *Client {
	file, err := os.Create(filename)
	if err != nil {
		c.log.Errorf("create dump file error: %v", err)
		return c
	}
	c.getDumpOptions().Output = file
	c.EnableDumpAll()
	return c
}

// EnableDumpAllTo enable dump for all requests and output to
// the specified io.Writer.
func (c *Client) EnableDumpAllTo(output io.Writer) *Client {
	c.getDumpOptions().Output = output
	c.EnableDumpAll()
	return c
}

// EnableDumpAllAsync enable dump for all requests and output
// asynchronously, can be used for debugging in production
// environment without affecting performance.
func (c *Client) EnableDumpAllAsync() *Client {
	o := c.getDumpOptions()
	o.Async = true
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutRequestBody enable dump for all requests without
// request body, can be used in the upload request to avoid dumping the
// unreadable binary content.
func (c *Client) EnableDumpAllWithoutRequestBody() *Client {
	o := c.getDumpOptions()
	o.RequestBody = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutResponseBody enable dump for all requests without
// response body, can be used in the download request to avoid dumping the
// unreadable binary content.
func (c *Client) EnableDumpAllWithoutResponseBody() *Client {
	o := c.getDumpOptions()
	o.ResponseBody = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutResponse enable dump for all requests without response,
// can be used if you only care about the request.
func (c *Client) EnableDumpAllWithoutResponse() *Client {
	o := c.getDumpOptions()
	o.ResponseBody = false
	o.ResponseHeader = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutRequest enables dump for all requests without request,
// can be used if you only care about the response.
func (c *Client) EnableDumpAllWithoutRequest() *Client {
	o := c.getDumpOptions()
	o.RequestHeader = false
	o.RequestBody = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutHeader enable dump for all requests without header,
// can be used if you only care about the body.
func (c *Client) EnableDumpAllWithoutHeader() *Client {
	o := c.getDumpOptions()
	o.RequestHeader = false
	o.ResponseHeader = false
	c.EnableDumpAll()
	return c
}

// EnableDumpAllWithoutBody enable dump for all requests without body,
// can be used if you only care about the header.
func (c *Client) EnableDumpAllWithoutBody() *Client {
	o := c.getDumpOptions()
	o.RequestBody = false
	o.ResponseBody = false
	c.EnableDumpAll()
	return c
}

// NewRequest is the alias of R()
func (c *Client) NewRequest() *Request {
	return c.R()
}

// DisableAutoReadResponse disable read response body automatically (enabled by default).
func (c *Client) DisableAutoReadResponse() *Client {
	c.disableAutoReadResponse = true
	return c
}

// EnableAutoReadResponse enable read response body automatically (enabled by default).
func (c *Client) EnableAutoReadResponse() *Client {
	c.disableAutoReadResponse = false
	return c
}

// SetAutoDecodeContentType set the content types that will be auto-detected and decode
// to utf-8 (e.g. "json", "xml", "html", "text").
func (c *Client) SetAutoDecodeContentType(contentTypes ...string) *Client {
	opt := c.getResponseOptions()
	opt.AutoDecodeContentType = autoDecodeContentTypeFunc(contentTypes...)
	return c
}

// SetAutoDecodeContentTypeFunc set the function that determines whether the
// specified `Content-Type` should be auto-detected and decode to utf-8.
func (c *Client) SetAutoDecodeContentTypeFunc(fn func(contentType string) bool) *Client {
	opt := c.getResponseOptions()
	opt.AutoDecodeContentType = fn
	return c
}

// SetAutoDecodeAllContentType enable try auto-detect charset and decode all
// content type to utf-8.
func (c *Client) SetAutoDecodeAllContentType() *Client {
	opt := c.getResponseOptions()
	opt.AutoDecodeContentType = func(contentType string) bool {
		return true
	}
	return c
}

// DisableAutoDecode disable auto-detect charset and decode to utf-8
// (enabled by default).
func (c *Client) DisableAutoDecode() *Client {
	c.getResponseOptions().DisableAutoDecode = true
	return c
}

// EnableAutoDecode enable auto-detect charset and decode to utf-8
// (enabled by default).
func (c *Client) EnableAutoDecode() *Client {
	c.getResponseOptions().DisableAutoDecode = false
	return c
}

// SetUserAgent set the "User-Agent" header for all requests.
func (c *Client) SetUserAgent(userAgent string) *Client {
	return c.SetCommonHeader(hdrUserAgentKey, userAgent)
}

// SetCommonBearerAuthToken set the bearer auth token for all requests.
func (c *Client) SetCommonBearerAuthToken(token string) *Client {
	return c.SetCommonHeader("Authorization", "Bearer "+token)
}

// SetCommonBasicAuth set the basic auth for all requests.
func (c *Client) SetCommonBasicAuth(username, password string) *Client {
	c.SetCommonHeader("Authorization", util.BasicAuthHeaderValue(username, password))
	return c
}

// SetCommonHeaders set headers for all requests.
func (c *Client) SetCommonHeaders(hdrs map[string]string) *Client {
	for k, v := range hdrs {
		c.SetCommonHeader(k, v)
	}
	return c
}

// SetCommonHeader set a header for all requests.
func (c *Client) SetCommonHeader(key, value string) *Client {
	if c.Headers == nil {
		c.Headers = make(http.Header)
	}
	c.Headers.Set(key, value)
	return c
}

// SetCommonHeaderNonCanonical set a header for all requests which key is a
// non-canonical key (keep case unchanged), only valid for HTTP/1.1.
func (c *Client) SetCommonHeaderNonCanonical(key, value string) *Client {
	if c.Headers == nil {
		c.Headers = make(http.Header)
	}
	c.Headers[key] = append(c.Headers[key], value)
	return c
}

// SetCommonHeadersNonCanonical set headers for all requests which key is a
// non-canonical key (keep case unchanged), only valid for HTTP/1.1.
func (c *Client) SetCommonHeadersNonCanonical(hdrs map[string]string) *Client {
	for k, v := range hdrs {
		c.SetCommonHeaderNonCanonical(k, v)
	}
	return c
}

// SetCommonContentType set the `Content-Type` header for all requests.
func (c *Client) SetCommonContentType(ct string) *Client {
	c.SetCommonHeader(hdrContentTypeKey, ct)
	return c
}

// DisableDumpAll disable dump for all requests.
func (c *Client) DisableDumpAll() *Client {
	c.t.DisableDump()
	return c
}

// SetCommonDumpOptions configures the underlying Transport's DumpOptions
// for all requests.
func (c *Client) SetCommonDumpOptions(opt *DumpOptions) *Client {
	if opt == nil {
		return c
	}
	if opt.Output == nil {
		if c.dumpOptions != nil {
			opt.Output = c.dumpOptions.Output
		} else {
			opt.Output = os.Stdout
		}
	}
	c.dumpOptions = opt
	if c.t.dump != nil {
		c.t.dump.DumpOptions = opt
	}
	return c
}

// SetProxy set the proxy function.
func (c *Client) SetProxy(proxy func(*http.Request) (*urlpkg.URL, error)) *Client {
	c.t.Proxy = proxy
	return c
}

// OnBeforeRequest add a request middleware which hooks before request sent.
func (c *Client) OnBeforeRequest(m RequestMiddleware) *Client {
	c.udBeforeRequest = append(c.udBeforeRequest, m)
	return c
}

// OnAfterResponse add a response middleware which hooks after response received.
func (c *Client) OnAfterResponse(m ResponseMiddleware) *Client {
	c.afterResponse = append(c.afterResponse, m)
	return c
}

// SetProxyURL set proxy from the proxy URL.
func (c *Client) SetProxyURL(proxyUrl string) *Client {
	u, err := urlpkg.Parse(proxyUrl)
	if err != nil {
		c.log.Errorf("failed to parse proxy url %s: %v", proxyUrl, err)
		return c
	}
	c.t.Proxy = http.ProxyURL(u)
	return c
}

// DisableTraceAll disable trace for all requests.
func (c *Client) DisableTraceAll() *Client {
	c.trace = false
	return c
}

// EnableTraceAll enable trace for all requests.
func (c *Client) EnableTraceAll() *Client {
	c.trace = true
	return c
}

// SetCookieJar set the `CookeJar` to the underlying `http.Client`.
func (c *Client) SetCookieJar(jar http.CookieJar) *Client {
	c.httpClient.Jar = jar
	return c
}

// SetJsonMarshal set the JSON marshal function which will be used
// to marshal request body.
func (c *Client) SetJsonMarshal(fn func(v interface{}) ([]byte, error)) *Client {
	c.jsonMarshal = fn
	return c
}

// SetJsonUnmarshal set the JSON unmarshal function which will be used
// to unmarshal response body.
func (c *Client) SetJsonUnmarshal(fn func(data []byte, v interface{}) error) *Client {
	c.jsonUnmarshal = fn
	return c
}

// SetXmlMarshal set the XML marshal function which will be used
// to marshal request body.
func (c *Client) SetXmlMarshal(fn func(v interface{}) ([]byte, error)) *Client {
	c.xmlMarshal = fn
	return c
}

// SetXmlUnmarshal set the XML unmarshal function which will be used
// to unmarshal response body.
func (c *Client) SetXmlUnmarshal(fn func(data []byte, v interface{}) error) *Client {
	c.xmlUnmarshal = fn
	return c
}

// SetDialTLS set the customized `DialTLSContext` function to Transport.
// Make sure the returned `conn` implements TLSConn if you want your
// customized `conn` supports HTTP2.
func (c *Client) SetDialTLS(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Client {
	c.t.DialTLSContext = fn
	return c
}

// SetDial set the customized `DialContext` function to Transport.
func (c *Client) SetDial(fn func(ctx context.Context, network, addr string) (net.Conn, error)) *Client {
	c.t.DialContext = fn
	return c
}

// SetTLSHandshakeTimeout set the TLS handshake timeout.
func (c *Client) SetTLSHandshakeTimeout(timeout time.Duration) *Client {
	c.t.TLSHandshakeTimeout = timeout
	return c
}

// EnableForceHTTP1 enable force using HTTP1 (disabled by default).
func (c *Client) EnableForceHTTP1() *Client {
	c.t.ForceHttpVersion = HTTP1
	return c
}

// EnableForceHTTP2 enable force using HTTP2 for https requests
// (disabled by default).
func (c *Client) EnableForceHTTP2() *Client {
	c.t.ForceHttpVersion = HTTP2
	return c
}

// DisableForceHttpVersion disable force using HTTP1 (disabled by default).
func (c *Client) DisableForceHttpVersion() *Client {
	c.t.ForceHttpVersion = ""
	return c
}

// DisableAllowGetMethodPayload disable sending GET method requests with body.
func (c *Client) DisableAllowGetMethodPayload() *Client {
	c.AllowGetMethodPayload = false
	return c
}

// EnableAllowGetMethodPayload allows sending GET method requests with body.
func (c *Client) EnableAllowGetMethodPayload() *Client {
	c.AllowGetMethodPayload = true
	return c
}

func (c *Client) isPayloadForbid(m string) bool {
	return (m == http.MethodGet && !c.AllowGetMethodPayload) || m == http.MethodHead || m == http.MethodOptions
}

// GetClient returns the underlying `http.Client`.
func (c *Client) GetClient() *http.Client {
	return c.httpClient
}

func (c *Client) getRetryOption() *retryOption {
	if c.retryOption == nil {
		c.retryOption = newDefaultRetryOption()
	}
	return c.retryOption
}

// SetCommonRetryCount enables retry and set the maximum retry count for all requests.
func (c *Client) SetCommonRetryCount(count int) *Client {
	c.getRetryOption().MaxRetries = count
	return c
}

// SetCommonRetryInterval sets the custom GetRetryIntervalFunc for all requests,
// you can use this to implement your own backoff retry algorithm.
// For example:
// 	 req.SetCommonRetryInterval(func(resp *req.Response, attempt int) time.Duration {
//      sleep := 0.01 * math.Exp2(float64(attempt))
//      return time.Duration(math.Min(2, sleep)) * time.Second
// 	 })
func (c *Client) SetCommonRetryInterval(getRetryIntervalFunc GetRetryIntervalFunc) *Client {
	c.getRetryOption().GetRetryInterval = getRetryIntervalFunc
	return c
}

// SetCommonRetryFixedInterval set retry to use a fixed interval for all requests.
func (c *Client) SetCommonRetryFixedInterval(interval time.Duration) *Client {
	c.getRetryOption().GetRetryInterval = func(resp *Response, attempt int) time.Duration {
		return interval
	}
	return c
}

// SetCommonRetryBackoffInterval set retry to use a capped exponential backoff with jitter
// for all requests.
// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
func (c *Client) SetCommonRetryBackoffInterval(min, max time.Duration) *Client {
	c.getRetryOption().GetRetryInterval = backoffInterval(min, max)
	return c
}

// SetCommonRetryHook set the retry hook which will be executed before a retry.
// It will override other retry hooks if any been added before.
func (c *Client) SetCommonRetryHook(hook RetryHookFunc) *Client {
	c.getRetryOption().RetryHooks = []RetryHookFunc{hook}
	return c
}

// AddCommonRetryHook adds a retry hook for all requests, which will be
// executed before a retry.
func (c *Client) AddCommonRetryHook(hook RetryHookFunc) *Client {
	ro := c.getRetryOption()
	ro.RetryHooks = append(ro.RetryHooks, hook)
	return c
}

// SetCommonRetryCondition sets the retry condition, which determines whether the
// request should retry.
// It will override other retry conditions if any been added before.
func (c *Client) SetCommonRetryCondition(condition RetryConditionFunc) *Client {
	c.getRetryOption().RetryConditions = []RetryConditionFunc{condition}
	return c
}

// AddCommonRetryCondition adds a retry condition, which determines whether the
// request should retry.
func (c *Client) AddCommonRetryCondition(condition RetryConditionFunc) *Client {
	ro := c.getRetryOption()
	ro.RetryConditions = append(ro.RetryConditions, condition)
	return c
}

// SetUnixSocket set client to dial connection use unix socket.
// For example:
//   client.SetUnixSocket("/var/run/custom.sock")
func (c *Client) SetUnixSocket(file string) *Client {
	return c.SetDial(func(ctx context.Context, network, addr string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, "unix", file)
	})
}

// NewClient is the alias of C
func NewClient() *Client {
	return C()
}

// Clone copy and returns the Client
func (c *Client) Clone() *Client {
	t := c.t.Clone()
	t2, _ := http2ConfigureTransports(t)
	client := *c.httpClient
	client.Transport = t

	cc := *c
	cc.httpClient = &client
	cc.t = t
	cc.t2 = t2

	cc.Headers = cloneHeaders(c.Headers)
	cc.Cookies = cloneCookies(c.Cookies)
	cc.PathParams = cloneMap(c.PathParams)
	cc.QueryParams = cloneUrlValues(c.QueryParams)
	cc.FormData = cloneUrlValues(c.FormData)
	cc.beforeRequest = cloneRequestMiddleware(c.beforeRequest)
	cc.udBeforeRequest = cloneRequestMiddleware(c.udBeforeRequest)
	cc.afterResponse = cloneResponseMiddleware(c.afterResponse)
	cc.dumpOptions = c.dumpOptions.Clone()
	cc.retryOption = c.retryOption.Clone()

	cc.log = c.log
	cc.jsonUnmarshal = c.jsonUnmarshal
	cc.jsonMarshal = c.jsonMarshal
	cc.xmlMarshal = c.xmlMarshal
	cc.xmlUnmarshal = c.xmlUnmarshal

	return &cc
}

// C create a new client.
func C() *Client {
	t := &Transport{
		ResponseOptions:       &ResponseOptions{},
		ForceAttemptHTTP2:     true,
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	t2, _ := http2ConfigureTransports(t)
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	httpClient := &http.Client{
		Transport: t,
		Jar:       jar,
		Timeout:   2 * time.Minute,
	}
	beforeRequest := []RequestMiddleware{
		parseRequestHeader,
		parseRequestURL,
		parseRequestBody,
		parseRequestCookie,
	}
	afterResponse := []ResponseMiddleware{
		parseResponseBody,
		handleDownload,
	}
	c := &Client{
		beforeRequest: beforeRequest,
		afterResponse: afterResponse,
		log:           createDefaultLogger(),
		httpClient:    httpClient,
		t:             t,
		t2:            t2,
		jsonMarshal:   json.Marshal,
		jsonUnmarshal: json.Unmarshal,
		xmlMarshal:    xml.Marshal,
		xmlUnmarshal:  xml.Unmarshal,
	}
	httpClient.CheckRedirect = c.defaultCheckRedirect

	t.Debugf = func(format string, v ...interface{}) {
		if c.DebugLog {
			c.log.Debugf(format, v...)
		}
	}
	return c
}

func (c *Client) do(r *Request) (resp *Response, err error) {
	resp = &Response{
		Request: r,
	}

	for {
		for _, f := range r.client.udBeforeRequest {
			if err = f(r.client, r); err != nil {
				return
			}
		}
		for _, f := range r.client.beforeRequest {
			if err = f(r.client, r); err != nil {
				return
			}
		}

		// setup trace
		if r.trace == nil && r.client.trace {
			r.trace = &clientTrace{}
		}
		if r.trace != nil {
			r.ctx = r.trace.createContext(r.Context())
		}

		// setup url and host
		var host string
		if h := r.getHeader("Host"); h != "" {
			host = h // Host header override
		} else {
			host = r.URL.Host
		}

		// setup header
		var header http.Header
		if r.Headers == nil {
			header = make(http.Header)
		} else {
			header = r.Headers
		}
		contentLength := int64(len(r.body))

		var reqBody io.ReadCloser
		if r.getBody != nil {
			reqBody, err = r.getBody()
			if err != nil {
				return
			}
		}
		req := &http.Request{
			Method:        r.Method,
			Header:        header,
			URL:           r.URL,
			Host:          host,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			ContentLength: contentLength,
			Body:          reqBody,
			GetBody:       r.getBody,
		}
		for _, cookie := range r.Cookies {
			req.AddCookie(cookie)
		}
		ctx := r.ctx
		if r.isSaveResponse && r.downloadCallback != nil {
			var wrap wrapResponseBodyFunc = func(rc io.ReadCloser) io.ReadCloser {
				return &callbackReader{
					ReadCloser: rc,
					callback: func(read int64) {
						r.downloadCallback(DownloadInfo{
							Response:       resp,
							DownloadedSize: read,
						})
					},
					lastTime: time.Now(),
					interval: r.downloadCallbackInterval,
				}
			}
			if ctx == nil {
				ctx = context.Background()
			}
			ctx = context.WithValue(ctx, wrapResponseBodyKey, wrap)
		}
		if ctx != nil {
			req = req.WithContext(ctx)
		}
		r.RawRequest = req
		r.StartTime = time.Now()
		var httpResponse *http.Response
		httpResponse, err = c.httpClient.Do(req)
		resp.Response = httpResponse

		// auto-read response body if possible
		if err == nil && !c.disableAutoReadResponse && !r.isSaveResponse {
			_, err = resp.ToBytes()
			if err != nil {
				return
			}
		}

		if r.retryOption == nil || r.RetryAttempt >= r.retryOption.MaxRetries { // absolutely cannot retry.
			if err != nil { // return immediately if error occurs.
				return
			}
			break // jump out to execute the ResponseMiddlewares if possible.
		}

		// check retry whether is needed.
		needRetry := err != nil                                   // default behaviour: retry if error occurs
		for _, condition := range r.retryOption.RetryConditions { // override default behaviour if custom RetryConditions has been set.
			needRetry = condition(resp, err)
			if needRetry {
				break
			}
		}
		if !needRetry { // no retry is needed.
			break // jump out to execute the ResponseMiddlewares.
		}

		// need retry, attempt to retry
		r.RetryAttempt++
		for _, hook := range r.retryOption.RetryHooks { // run retry hooks
			hook(resp, err)
		}
		time.Sleep(r.retryOption.GetRetryInterval(resp, r.RetryAttempt))

		// clean up before retry
		if r.dumpBuffer != nil {
			r.dumpBuffer.Reset()
		}
		r.trace = nil
		r.ctx = nil
		resp.body = nil
	}

	for _, f := range r.client.afterResponse {
		if err = f(r.client, resp); err != nil {
			return
		}
	}
	return
}
