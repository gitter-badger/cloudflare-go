package cloudflare

import (
	"encoding/json"
	"net/url"

	"github.com/pkg/errors"
)

// Zone describes a CloudFlare zone.
type Zone struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	DevMode           int      `json:"development_mode"`
	OriginalNS        []string `json:"original_name_servers"`
	OriginalRegistrar string   `json:"original_registrar"`
	OriginalDNSHost   string   `json:"original_dnshost"`
	CreatedOn         string   `json:"created_on"`
	ModifiedOn        string   `json:"modified_on"`
	NameServers       []string `json:"name_servers"`
	Owner             Owner    `json:"owner"`
	Permissions       []string `json:"permissions"`
	Plan              ZonePlan `json:"plan"`
	Status            string   `json:"status"`
	Paused            bool     `json:"paused"`
	Type              string   `json:"type"`
	Host              struct {
		Name    string
		Website string
	} `json:"host"`
	VanityNS    []string `json:"vanity_name_servers"`
	Betas       []string `json:"betas"`
	DeactReason string   `json:"deactivation_reason"`
	Meta        ZoneMeta `json:"meta"`
}

// ZoneMeta metadata about a zone.
type ZoneMeta struct {
	// custom_certificate_quota is broken - sometimes it's a string, sometimes a number!
	// CustCertQuota     int    `json:"custom_certificate_quota"`
	PageRuleQuota     int  `json:"page_rule_quota"`
	WildcardProxiable bool `json:"wildcard_proxiable"`
	PhishingDetected  bool `json:"phishing_detected"`
}

// ZonePlan contains the plan information for a zone.
type ZonePlan struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Price        int    `json:"price"`
	Currency     string `json:"currency"`
	Frequency    string `json:"frequency"`
	LegacyID     string `json:"legacy_id"`
	IsSubscribed bool   `json:"is_subscribed"`
	CanSubscribe bool   `json:"can_subscribe"`
}

// ZoneID contains only the zone ID.
type ZoneID struct {
	ID string `json:"id"`
}

// ZoneResponse represents the response from the Zone endpoint containing an array of zones.
type ZoneResponse struct {
	Response
	Result []Zone `json:"result"`
}

// ZoneIDResponse represents the response from the Zone endpoint, containing only a zone ID.
type ZoneIDResponse struct {
	Response
	Result ZoneID `json:"result"`
}

// AvailablePlansResponse represents the response from the Available Plans endpoint.
type AvailablePlansResponse struct {
	Response
	Result []ZonePlan `json:"result"`
	ResultInfo
}

// AvailablePlansResponse represents the response from the Plan Details endpoint.
type ZonePlanResponse struct {
	Response
	Result ZonePlan `json:"result"`
}

// ZoneSetting contains settings for a zone.
type ZoneSetting struct {
	ID            string      `json:"id"`
	Editable      bool        `json:"editable"`
	ModifiedOn    string      `json:"modified_on"`
	Value         interface{} `json:"value"`
	TimeRemaining int         `json:"time_remaining"`
}

// ZoneSettingResponse represents the response from the Zone Setting endpoint.
type ZoneSettingResponse struct {
	Response
	Result []ZoneSetting `json:"result"`
}

// CreateZone creates a zone on an account.
// API reference:
// 	https://api.cloudflare.com/#zone-create-a-zone
//	POST /zones
func (api *API) CreateZone(z Zone) {
	// res, err := api.makeRequest("POST", "/zones", z)
}

// ZoneActivationCheck initiates another zone activation check for newly-created zones.
// API reference:
//   https://api.cloudflare.com/#zone-initiate-another-zone-activation-check
//   PUT /zones/:identifier/activation_check
func (api *API) ZoneActivationCheck(zoneID string) (Response, error) {
	res, err := api.makeRequest("PUT", "/zones/"+zoneID+"/activation_check", nil)
	if err != nil {
		return Response{}, errors.Wrap(err, errMakeRequestError)
	}
	var r Response
	err = json.Unmarshal(res, &r)
	return r, nil
}

// ListZones lists zones on an account. Optionally takes a list of zone names
// to filter against.
// API reference:
//	https://api.cloudflare.com/#zone-list-zones
//	GET /zones
func (api *API) ListZones(z ...string) ([]Zone, error) {
	v := url.Values{}
	var res []byte
	var r ZoneResponse
	var zones []Zone
	var err error
	if len(z) > 0 {
		for _, zone := range z {
			v.Set("name", zone)
			res, err = api.makeRequest("GET", "/zones?"+v.Encode(), nil)
			if err != nil {
				return []Zone{}, errors.Wrap(err, errMakeRequestError)
			}
			err = json.Unmarshal(res, &r)
			if err != nil {
				return []Zone{}, errors.Wrap(err, errUnmarshalError)
			}
			if !r.Success {
				// TODO: Provide an actual error message instead of always returning nil
				return []Zone{}, err
			}
			for zi := range r.Result {
				zones = append(zones, r.Result[zi])
			}
		}
	} else {
		// TODO: Paginate here. We only grab the first page of results.
		// Could do this concurrently after the first request by creating a
		// sync.WaitGroup or just a channel + workers.
		res, err = api.makeRequest("GET", "/zones", nil)
		if err != nil {
			return []Zone{}, errors.Wrap(err, errMakeRequestError)
		}
		err = json.Unmarshal(res, &r)
		if err != nil {
			return []Zone{}, errors.Wrap(err, errUnmarshalError)
		}
		zones = r.Result
	}

	return zones, nil
}

// ZoneDetails fetches information about a zone.
// API reference:
// 	https://api.cloudflare.com/#zone-zone-details
// 	GET /zones/:id
func (api *API) ZoneDetails(zoneID string) ([]Zone, error) {
	res, err := api.makeRequest("GET", "/zones"+zoneID, nil)
	if err != nil {
		return []Zone{}, errors.Wrap(err, errMakeRequestError)
	}
	var r ZoneResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return []Zone{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// EditZone edits the given zone.
// API reference:
// 	https://api.cloudflare.com/#zone-edit-zone-properties
// 	PATCH /zones/:id
func EditZone() {
}

// PurgeEverything purges the cache for the given zone.
// Note: this will substantially increase load on the origin server for that
// zone if there is a high cached vs. uncached request ratio.
// API reference:
// https://api.cloudflare.com/#zone-purge-all-files
// DELETE /zones/:id/purge_cache
func (api *API) PurgeEverything(zoneID string) (PurgeCacheResponse, error) {
	uri := "/zones/" + zoneID + "/purge_cache"
	res, err := api.makeRequest("DELETE", uri, PurgeCacheRequest{true, nil, nil})
	if err != nil {
		return PurgeCacheResponse{}, errors.Wrap(err, errMakeRequestError)
	}
	var r PurgeCacheResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return PurgeCacheResponse{}, errors.Wrap(err, errUnmarshalError)
	}
	return r, nil
}

// PurgeCache purges the cache using the given PurgeCacheRequest (zone/url/tag).
// API reference:
// 	https://api.cloudflare.com/#zone-purge-individual-files-by-url-and-cache-tags
// 	DELETE /zones/:id/purge_cache
func (api *API) PurgeCache(zoneID string, pcr PurgeCacheRequest) (PurgeCacheResponse, error) {
	uri := "/zones/" + zoneID + "/purge_cache"
	res, err := api.makeRequest("DELETE", uri, pcr)
	if err != nil {
		return PurgeCacheResponse{}, errors.Wrap(err, errMakeRequestError)
	}
	var r PurgeCacheResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return PurgeCacheResponse{}, errors.Wrap(err, errUnmarshalError)
	}
	return r, nil
}

// DeleteZone deletes the given zone.
// API reference:
//   https://api.cloudflare.com/#zone-delete-a-zone
//   DELETE /zones/:id
func (api *API) DeleteZone(zoneID string) (ZoneID, error) {
	res, err := api.makeRequest("DELETE", "/zones"+zoneID, nil)
	if err != nil {
		return ZoneID{}, errors.Wrap(err, errMakeRequestError)
	}
	var r ZoneIDResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return ZoneID{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// AvailablePlans returns information about all plans available to the specified zone.
// API reference:
//  https://api.cloudflare.com/#zone-plan-available-plans
//  GET /zones/:zone_identifier/available_plans
func (api *API) AvailablePlans(zoneID string) ([]ZonePlan, error) {
	uri := "/zones/" + zoneID + "/available_plans"
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return []ZonePlan{}, errors.Wrap(err, errMakeRequestError)
	}
	var r AvailablePlansResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return []ZonePlan{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// PlanDetails returns information about a zone plan.
// API reference:
//  https://api.cloudflare.com/#zone-plan-plan-details
//  GET /zones/:zone_identifier/available_plans/:identifier
func (api *API) PlanDetails(zoneID, planID string) (ZonePlan, error) {
	uri := "/zones/" + zoneID + "/available_plans/" + planID
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return ZonePlan{}, errors.Wrap(err, errMakeRequestError)
	}
	var r ZonePlanResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return ZonePlan{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// Zone Settings
// https://api.cloudflare.com/#zone-settings-for-a-zone-get-all-zone-settings
// e.g.
// https://api.cloudflare.com/#zone-settings-for-a-zone-get-always-online-setting
// https://api.cloudflare.com/#zone-settings-for-a-zone-change-always-online-setting
