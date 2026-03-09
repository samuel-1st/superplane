package digitalocean

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/superplanehq/superplane/pkg/core"
)

const baseURL = "https://api.digitalocean.com/v2"

type Client struct {
	Token   string
	http    core.HTTPContext
	BaseURL string
}

type DOAPIError struct {
	StatusCode int
	Body       []byte
}

func (e *DOAPIError) Error() string {
	return fmt.Sprintf("request got %d code: %s", e.StatusCode, string(e.Body))
}

func NewClient(http core.HTTPContext, ctx core.IntegrationContext) (*Client, error) {
	apiToken, err := ctx.GetConfig("apiToken")
	if err != nil {
		return nil, fmt.Errorf("error finding API token: %v", err)
	}

	return &Client{
		Token:   string(apiToken),
		http:    http,
		BaseURL: baseURL,
	}, nil
}

func (c *Client) execRequest(method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error building request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %v", err)
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %v", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, &DOAPIError{
			StatusCode: res.StatusCode,
			Body:       responseBody,
		}
	}

	return responseBody, nil
}

// Account represents a DigitalOcean account
type Account struct {
	Email        string `json:"email"`
	UUID         string `json:"uuid"`
	Status       string `json:"status"`
	DropletLimit int    `json:"droplet_limit"`
}

// GetAccount validates the API token by fetching account info
func (c *Client) GetAccount() (*Account, error) {
	url := fmt.Sprintf("%s/account", c.BaseURL)
	responseBody, err := c.execRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Account Account `json:"account"`
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &response.Account, nil
}

// Region represents a DigitalOcean region
type Region struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Available bool   `json:"available"`
}

// ListRegions retrieves all available regions
func (c *Client) ListRegions() ([]Region, error) {
	url := fmt.Sprintf("%s/regions", c.BaseURL)
	responseBody, err := c.execRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Regions []Region `json:"regions"`
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return response.Regions, nil
}

// Size represents a DigitalOcean droplet size
type Size struct {
	Slug         string  `json:"slug"`
	Memory       int     `json:"memory"`
	VCPUs        int     `json:"vcpus"`
	Disk         int     `json:"disk"`
	PriceMonthly float64 `json:"price_monthly"`
	Available    bool    `json:"available"`
}

// ListSizes retrieves all available droplet sizes
func (c *Client) ListSizes() ([]Size, error) {
	url := fmt.Sprintf("%s/sizes?per_page=200", c.BaseURL)
	responseBody, err := c.execRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Sizes []Size `json:"sizes"`
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return response.Sizes, nil
}

// Image represents a DigitalOcean image
type Image struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Type         string `json:"type"`
	Distribution string `json:"distribution"`
}

// ListImages retrieves images of a given type (e.g., "distribution")
func (c *Client) ListImages(imageType string) ([]Image, error) {
	url := fmt.Sprintf("%s/images?type=%s&per_page=200", c.BaseURL, imageType)
	responseBody, err := c.execRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Images []Image `json:"images"`
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return response.Images, nil
}

// CreateDropletRequest is the payload for creating a droplet
type CreateDropletRequest struct {
	Name     string   `json:"name"`
	Region   string   `json:"region"`
	Size     string   `json:"size"`
	Image    string   `json:"image"`
	SSHKeys  []string `json:"ssh_keys,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	UserData string   `json:"user_data,omitempty"`
}

// Droplet represents a DigitalOcean droplet
type Droplet struct {
	ID       int             `json:"id"`
	Name     string          `json:"name"`
	Memory   int             `json:"memory"`
	VCPUs    int             `json:"vcpus"`
	Disk     int             `json:"disk"`
	Status   string          `json:"status"`
	Region   DropletRegion   `json:"region"`
	Image    DropletImage    `json:"image"`
	SizeSlug string          `json:"size_slug"`
	Networks DropletNetworks `json:"networks"`
	Tags     []string        `json:"tags"`
}

type DropletRegion struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type DropletImage struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type DropletNetworks struct {
	V4 []DropletNetworkV4 `json:"v4"`
}

type DropletNetworkV4 struct {
	IPAddress string `json:"ip_address"`
	Type      string `json:"type"`
}

// CreateDroplet creates a new droplet
func (c *Client) CreateDroplet(req CreateDropletRequest) (*Droplet, error) {
	url := fmt.Sprintf("%s/droplets", c.BaseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	responseBody, err := c.execRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var response struct {
		Droplet Droplet `json:"droplet"`
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &response.Droplet, nil
}

// GetDroplet retrieves a droplet by its ID
func (c *Client) GetDroplet(dropletID int) (*Droplet, error) {
	url := fmt.Sprintf("%s/droplets/%d", c.BaseURL, dropletID)
	responseBody, err := c.execRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Droplet Droplet `json:"droplet"`
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &response.Droplet, nil
}

// DOAction represents a DigitalOcean action
type DOAction struct {
	ID           int    `json:"id"`
	Status       string `json:"status"`
	Type         string `json:"type"`
	StartedAt    string `json:"started_at"`
	CompletedAt  string `json:"completed_at"`
	ResourceID   int    `json:"resource_id"`
	ResourceType string `json:"resource_type"`
	RegionSlug   string `json:"region_slug"`
}

// ListActions retrieves actions filtered by resource type.
// The DigitalOcean /v2/actions API does not support resource_type as a query
// parameter, so we fetch all recent actions and filter client-side.
func (c *Client) ListActions(resourceType string) ([]DOAction, error) {
	url := fmt.Sprintf("%s/actions?page=1&per_page=50", c.BaseURL)
	responseBody, err := c.execRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Actions []DOAction `json:"actions"`
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	filtered := make([]DOAction, 0, len(response.Actions))
	for _, a := range response.Actions {
		if a.ResourceType == resourceType {
			filtered = append(filtered, a)
		}
	}

	return filtered, nil
}

// GetAction retrieves a single action by ID
func (c *Client) GetAction(actionID int) (*DOAction, error) {
	url := fmt.Sprintf("%s/actions/%d", c.BaseURL, actionID)
	responseBody, err := c.execRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var response struct {
		Action DOAction `json:"action"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}
	return &response.Action, nil
}

// PostDropletAction performs an action on a droplet
func (c *Client) PostDropletAction(dropletID int, body map[string]any) (*DOAction, error) {
	url := fmt.Sprintf("%s/droplets/%d/actions", c.BaseURL, dropletID)
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}
	responseBody, err := c.execRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var response struct {
		Action DOAction `json:"action"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}
	return &response.Action, nil
}

// DeleteDroplet deletes a droplet by ID
func (c *Client) DeleteDroplet(dropletID int) error {
	url := fmt.Sprintf("%s/droplets/%d", c.BaseURL, dropletID)
	_, err := c.execRequest(http.MethodDelete, url, nil)
	return err
}

// Snapshot represents a DigitalOcean snapshot
type Snapshot struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	ResourceID    string   `json:"resource_id"`
	ResourceType  string   `json:"resource_type"`
	Regions       []string `json:"regions"`
	SizeGigabytes float64  `json:"size_gigabytes"`
	CreatedAt     string   `json:"created_at"`
}

// DeleteSnapshot deletes a snapshot by ID
func (c *Client) DeleteSnapshot(snapshotID string) error {
	url := fmt.Sprintf("%s/snapshots/%s", c.BaseURL, snapshotID)
	_, err := c.execRequest(http.MethodDelete, url, nil)
	return err
}

// ListSnapshots retrieves all snapshots
func (c *Client) ListSnapshots() ([]Snapshot, error) {
	url := fmt.Sprintf("%s/snapshots", c.BaseURL)
	responseBody, err := c.execRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var response struct {
		Snapshots []Snapshot `json:"snapshots"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}
	return response.Snapshots, nil
}

// Domain represents a DigitalOcean domain
type Domain struct {
	Name string `json:"name"`
	TTL  int    `json:"ttl"`
}

// DNSRecord represents a DigitalOcean DNS record
type DNSRecord struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Data     string `json:"data"`
	TTL      int    `json:"ttl"`
	Priority int    `json:"priority,omitempty"`
}

// DNSRecordRequest is the payload for creating/updating a DNS record
type DNSRecordRequest struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Data     string `json:"data"`
	TTL      int    `json:"ttl"`
	Priority int    `json:"priority,omitempty"`
}

// CreateDNSRecord creates a new DNS record for a domain
func (c *Client) CreateDNSRecord(domain string, req DNSRecordRequest) (*DNSRecord, error) {
	url := fmt.Sprintf("%s/domains/%s/records", c.BaseURL, domain)
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}
	responseBody, err := c.execRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var response struct {
		DomainRecord DNSRecord `json:"domain_record"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}
	return &response.DomainRecord, nil
}

// DeleteDNSRecord deletes a DNS record from a domain
func (c *Client) DeleteDNSRecord(domain string, recordID int) error {
	url := fmt.Sprintf("%s/domains/%s/records/%d", c.BaseURL, domain, recordID)
	_, err := c.execRequest(http.MethodDelete, url, nil)
	return err
}

// ListDNSRecords retrieves DNS records for a domain, with optional query params
func (c *Client) ListDNSRecords(domain string, queryParams ...string) ([]DNSRecord, error) {
	url := fmt.Sprintf("%s/domains/%s/records", c.BaseURL, domain)
	if len(queryParams) > 0 {
		url += "?" + queryParams[0]
	}
	responseBody, err := c.execRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var response struct {
		DomainRecords []DNSRecord `json:"domain_records"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}
	return response.DomainRecords, nil
}

// UpdateDNSRecord updates an existing DNS record
func (c *Client) UpdateDNSRecord(domain string, recordID int, req DNSRecordRequest) (*DNSRecord, error) {
	url := fmt.Sprintf("%s/domains/%s/records/%d", c.BaseURL, domain, recordID)
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}
	responseBody, err := c.execRequest(http.MethodPatch, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var response struct {
		DomainRecord DNSRecord `json:"domain_record"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}
	return &response.DomainRecord, nil
}

// ListDomains retrieves all domains
func (c *Client) ListDomains() ([]Domain, error) {
	url := fmt.Sprintf("%s/domains", c.BaseURL)
	responseBody, err := c.execRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var response struct {
		Domains []Domain `json:"domains"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}
	return response.Domains, nil
}

// ForwardingRule represents a load balancer forwarding rule
type ForwardingRule struct {
	EntryProtocol  string `json:"entry_protocol"`
	EntryPort      int    `json:"entry_port"`
	TargetProtocol string `json:"target_protocol"`
	TargetPort     int    `json:"target_port"`
}

// LoadBalancer represents a DigitalOcean load balancer
type LoadBalancer struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	IP              string           `json:"ip"`
	Status          string           `json:"status"`
	Region          DropletRegion    `json:"region"`
	ForwardingRules []ForwardingRule `json:"forwarding_rules"`
	DropletIDs      []int            `json:"droplet_ids"`
	Tags            []string         `json:"tag"`
	SizeSlug        string           `json:"size_slug"`
}

// CreateLoadBalancerRequest is the payload for creating a load balancer
type CreateLoadBalancerRequest struct {
	Name            string           `json:"name"`
	Region          string           `json:"region"`
	SizeSlug        string           `json:"size_slug,omitempty"`
	ForwardingRules []ForwardingRule `json:"forwarding_rules"`
	DropletIDs      []int            `json:"droplet_ids,omitempty"`
	Tags            []string         `json:"tag,omitempty"`
}

// CreateLoadBalancer creates a new load balancer
func (c *Client) CreateLoadBalancer(req CreateLoadBalancerRequest) (*LoadBalancer, error) {
	url := fmt.Sprintf("%s/load_balancers", c.BaseURL)
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}
	responseBody, err := c.execRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var response struct {
		LoadBalancer LoadBalancer `json:"load_balancer"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}
	return &response.LoadBalancer, nil
}

// DeleteLoadBalancer deletes a load balancer by ID
func (c *Client) DeleteLoadBalancer(lbID string) error {
	url := fmt.Sprintf("%s/load_balancers/%s", c.BaseURL, lbID)
	_, err := c.execRequest(http.MethodDelete, url, nil)
	return err
}

// ListLoadBalancers retrieves all load balancers
func (c *Client) ListLoadBalancers() ([]LoadBalancer, error) {
	url := fmt.Sprintf("%s/load_balancers", c.BaseURL)
	responseBody, err := c.execRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var response struct {
		LoadBalancers []LoadBalancer `json:"load_balancers"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}
	return response.LoadBalancers, nil
}

// ReservedIPAction represents the result of a reserved IP action
type ReservedIPAction struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Type   string `json:"type"`
}

// PostReservedIPAction performs an action on a reserved IP
func (c *Client) PostReservedIPAction(reservedIP string, body map[string]any) (*ReservedIPAction, error) {
	url := fmt.Sprintf("%s/reserved_ips/%s/actions", c.BaseURL, reservedIP)
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}
	responseBody, err := c.execRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var response struct {
		Action ReservedIPAction `json:"action"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}
	return &response.Action, nil
}
