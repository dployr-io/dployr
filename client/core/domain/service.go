// domain/service.go
package domain

import (
	"fmt"
	"time"

	"dployr/core/types"
)

type DomainService struct {
	baseURL    string
}

func NewDomainService(baseURL string) *DomainService {
	return &DomainService{
		baseURL:    baseURL,
	}
}

func (d *DomainService) AddDomain(domain string, projectID string) (types.Domain, error) {
	// _, err := http.Post(d.baseURL+"/foo/bar", "application/json" , &types.Domain{
	// 	domain: "foo.bar",
	// })

	time.Sleep(3 * time.Second)

	res := types.Domain{
		Provider:           "cloudflare",
		AutoSetupAvailable: true,
		ManualRecords:      d.generateManualInstructions(domain, "202.121.80.311"),
	}

	err := error(nil)

	return res, err
}

func (d *DomainService) GetDomains() []types.Domain {
	return []types.Domain{
		{
			Id:                 "39189134002340941",
			Subdomain:          "foo.bar",
			Provider:           "namecheap",
			AutoSetupAvailable: true,
			Verified:           false,
			UpdatedAt:          time.Now(),
		},
		{
			Id:                 "39189134002340940",
			Subdomain:          "29500390932930390332.dployr.io",
			Provider:           "cloudflare",
			AutoSetupAvailable: true,
			Verified:           true,
			UpdatedAt:          time.Now(),
		},
	}
}

func (d *DomainService) generateManualInstructions(domain, serverIP string) string {
	return fmt.Sprintf(`
A Record:
Name: @
Value: %s
TTL: 300

CNAME Record:
Name: www
Value: %s
TTL: 300
`, serverIP, domain)
}
