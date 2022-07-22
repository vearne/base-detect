package model

import (
	"time"
)

type RespStatus struct {
	Code    string `json:"code"`
	Message string `json:"msg"`
}

type HttpDetectResult struct {
	HttpCode int `json:"httpCode"`

	DataSize int `json:"dataSize"`

	// DNSLookup is a duration that transport took to perform
	// DNS lookup.
	DNSLookup time.Duration `json:"DNSLookup"`

	// ConnTime is a duration that took to obtain a successful connection.
	ConnTime time.Duration `json:"connTime"`

	// TCPConnTime is a duration that took to obtain the TCP connection.
	TCPConnTime time.Duration `json:"TCPConnTime"`

	// TLSHandshake is a duration that TLS handshake took place.
	TLSHandshake time.Duration `json:"TLSHandshake"`

	// ServerTime is a duration that server took to respond first byte.
	ServerTime time.Duration `json:"serverTime"`

	// ResponseTime is a duration since first response byte from server to
	// request completion.
	ResponseTime time.Duration `json:"responseTime"`

	// TotalTime is a duration that total request took end-to-end.
	TotalTime time.Duration `json:"totalTime"`

	// RemoteAddr returns the remote network address.
	RemoteAddr string `json:"remoteAddr"`
}
