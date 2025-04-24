package exoscale

import (
	"context"
	"fmt"

	v3 "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"
)

type APIAccess struct {
	Zone   string
	Key    string
	Secret string
	Creds  *credentials.Credentials
}

func NewAPIAccess(zone, key, secret string) *APIAccess {
	return &APIAccess{
		Zone:   zone,
		Key:    key,
		Secret: secret,
		Creds:  credentials.NewStaticCredentials(key, secret),
	}
}

type Instance struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	IP     string            `json:"ip"`
}

func (a *APIAccess) GetInstances() ([]*Instance, error) {
	instances := make([]*Instance, 0)
	client, err := v3.NewClient(a.Creds)
	if err != nil {
		return nil, fmt.Errorf("create client: %v", err)
	}
	resp, err := client.ListInstances(context.Background())
	if err != nil {
		return nil, fmt.Errorf("list instancers: %v", err)
	}
	for _, instance := range resp.Instances {
		instances = append(instances, &Instance{
			ID:     instance.ID.String(),
			Name:   instance.Name,
			Labels: instance.Labels,
			IP:     instance.PublicIP.To4().String(),
		})
	}
	return instances, nil
}
