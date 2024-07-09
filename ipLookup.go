package ipLookup

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
)

type IPLookupEndpoint interface {
	GetIP() (ip net.IP, err error)
}

type awsLookup struct{}

func (lookup *awsLookup) GetIP() (ip net.IP, err error) {
	return getPublicIP("https://checkip.amazonaws.com")
}

type apifyLookup struct{}

func (lookup *apifyLookup) GetIP() (ip net.IP, err error) {
	return getPublicIP("https://api.ipify.org")
}

type wtfIsMyIPLookup struct{}

func (lookup *wtfIsMyIPLookup) GetIP() (ip net.IP, err error) {
	return getPublicIP("https://wtfismyip.com/text")
}

type localLookup struct{}

func (lookup *localLookup) GetIP() (ip net.IP, err error) {
	return getPrivateIP()
}

// IPLookup defines a set of endpoints for which to look up this IP
type IPLookup struct {
	endpoints []IPLookupEndpoint
}

// New creates a new IPLookup
func New() *IPLookup {
	return &IPLookup{}
}

// WithAWS adds the AWS IP lookup endpoint to the IPLookup object
func (lookup *IPLookup) WithAWS() *IPLookup {
	lookup.endpoints = append(lookup.endpoints, &awsLookup{})
	return lookup
}

// WithAPIfy adds the APIfy lookup endpoint to the IPLookup object
func (lookup *IPLookup) WithAPIfy() *IPLookup {
	lookup.endpoints = append(lookup.endpoints, &apifyLookup{})
	return lookup
}

// WithWTFIsMyIP adds the wtfismyip lookup endpoint to the IPLookup object
func (lookup *IPLookup) WithWTFIsMyIP() *IPLookup {
	lookup.endpoints = append(lookup.endpoints, &wtfIsMyIPLookup{})
	return lookup
}

// WithLocal adds the local lookup endpoint to the IPLookup object
func (lookup *IPLookup) WithLocal() *IPLookup {
	lookup.endpoints = append(lookup.endpoints, &localLookup{})
	return lookup
}

// GetIP gets the IP address for the given IPLookup
func (lookup *IPLookup) GetIP() (ip net.IP, err error) {
	if len(lookup.endpoints) < 1 {
		return ip, errors.New("Must specify at least one public or local endpoint")
	}
	for _, endpoint := range lookup.endpoints {
		ip, err = endpoint.GetIP()
		if err == nil && ip != nil {
			return ip, err
		}
	}
	return ip, errors.New("Failed to lookup IP")
}

// getPrivateIP gets the private IP
func getPrivateIP() (ip net.IP, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ip, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

// getPublicIP gets the public IP by querying endpoint
func getPublicIP(endpoint string) (ip net.IP, err error) {
	resp, err := http.Get(string(endpoint))
	if err != nil {
		return ip, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ip, err
	}
	bodyString := strings.TrimSpace(string(body))

	ip = net.ParseIP(bodyString)
	if ip == nil {
		return ip, errors.New("Failed to unmarshal public IP")
	}

	return ip, nil
}
