package gateway

import (
	"fmt"
	nethttp "net/http"
	"strings"
	"time"
)

type DoctorRESTGateway struct {
	baseURL    string
	httpClient *nethttp.Client
}

func NewDoctorRESTGateway(baseURL string) *DoctorRESTGateway {
	return &DoctorRESTGateway{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &nethttp.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (g *DoctorRESTGateway) Exists(id string) (bool, error) {
	resp, err := g.httpClient.Get(fmt.Sprintf("%s/doctors/%s", g.baseURL, id))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case nethttp.StatusOK:
		return true, nil
	case nethttp.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("doctor service returned status %d", resp.StatusCode)
	}
}
