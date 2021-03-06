package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// APIInformation represents the information returned back from /v2/info
type APIInformation struct {
	APIVersion                   string `json:"api_version"`
	AuthorizationEndpoint        string `json:"authorization_endpoint"`
	DopplerEndpoint              string `json:"doppler_logging_endpoint"`
	MinCLIVersion                string `json:"min_cli_version"`
	MinimumRecommendedCLIVersion string `json:"min_recommended_cli_version"`
	Name                         string `json:"name"`
	RoutingEndpoint              string `json:"routing_endpoint"`
	TokenEndpoint                string `json:"token_endpoint"`
}

// API returns the Cloud Controller API URL for the targeted Cloud Controller.
func (client *Client) API() string {
	return client.cloudControllerURL
}

// APIVersion returns Cloud Controller API Version for the targeted Cloud
// Controller.
func (client *Client) APIVersion() string {
	return client.cloudControllerAPIVersion
}

// AuthorizationEndpoint returns the authorization endpoint for the targeted
// Cloud Controller.
func (client *Client) AuthorizationEndpoint() string {
	return client.authorizationEndpoint
}

// DopplerEndpoint returns the Doppler endpoint for the targetd Cloud
// Controller.
func (client *Client) DopplerEndpoint() string {
	return client.dopplerEndpoint
}

// MinCLIVersion returns the minimum CLI version required for the targeted
// Cloud Controller
func (client *Client) MinCLIVersion() string {
	return client.minCLIVersion
}

// RoutingEndpoint returns the Routing endpoint for the targeted Cloud
// Controller.
func (client *Client) RoutingEndpoint() string {
	return client.routingEndpoint
}

// TokenEndpoint returns the Token endpoint for the targeted Cloud Controller.
func (client *Client) TokenEndpoint() string {
	return client.tokenEndpoint
}

// Info returns back endpoint and API information from /v2/info.
func (client *Client) Info() (APIInformation, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetInfoRequest,
	})
	if err != nil {
		return APIInformation{}, nil, err
	}

	var info APIInformation
	response := cloudcontroller.Response{
		Result: &info,
	}

	err = client.connection.Make(request, &response)
	if _, ok := err.(ccerror.NotFoundError); ok {
		return APIInformation{}, nil, ccerror.APINotFoundError{URL: client.cloudControllerURL}
	}
	return info, response.Warnings, err
}
