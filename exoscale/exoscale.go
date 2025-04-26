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
	State  string            `json:"state"`
}

func (a *APIAccess) GetInstances() ([]*Instance, error) {
	instances := make([]*Instance, 0)
	client, err := v3.NewClient(a.Creds)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	resp, err := client.ListInstances(context.Background())
	if err != nil {
		return nil, fmt.Errorf("list instancers: %w", err)
	}
	for _, instance := range resp.Instances {
		instances = append(instances, fromListInstance(&instance))
	}
	return instances, nil
}

func (a *APIAccess) GetInstance(id string) (*Instance, error) {
	client, err := v3.NewClient(a.Creds)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	instance, err := client.GetInstance(context.Background(), v3.UUID(id))
	if err != nil {
		return nil, fmt.Errorf("get instance %s: %w", id, err)
	}
	return fromInstance(instance), nil
}

func (a *APIAccess) StartInstance(id string) error {
	client, err := v3.NewClient(a.Creds)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	_, err = client.StartInstance(context.Background(), v3.UUID(id), v3.StartInstanceRequest{})
	if err != nil {
		return fmt.Errorf("start instance %s: %w", id, err)
	}
	return nil
}

func (a *APIAccess) StopInstance(id string) error {
	client, err := v3.NewClient(a.Creds)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	_, err = client.StopInstance(context.Background(), v3.UUID(id))
	if err != nil {
		return fmt.Errorf("start instance %s: %w", id, err)
	}
	return nil
}

func fromListInstance(instance *v3.ListInstancesResponseInstances) *Instance {
	return &Instance{
		ID:     instance.ID.String(),
		Name:   instance.Name,
		Labels: instance.Labels,
		IP:     instance.PublicIP.To4().String(),
		State:  string(instance.State),
	}
}

func fromInstance(instance *v3.Instance) *Instance {
	return &Instance{
		ID:     instance.ID.String(),
		Name:   instance.Name,
		Labels: instance.Labels,
		IP:     instance.PublicIP.To4().String(),
		State:  string(instance.State),
	}
}
