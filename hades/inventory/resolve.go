package inventory

import (
	"context"
	"fmt"

	"github.com/SoftKiwiGames/hades/hades/cloud"
	"github.com/SoftKiwiGames/hades/hades/selector"
	"github.com/SoftKiwiGames/hades/hades/ssh"
	"github.com/SoftKiwiGames/hades/hades/utils"
)

func resolveProviders(ctx context.Context, providers []Provider, hosts map[string]ssh.Host, targets map[string][]string) error {
	for _, p := range providers {
		instances, err := fetchInstances(ctx, p)
		if err != nil {
			return err
		}

		for _, inst := range instances {
			if inst.Name == "" {
				continue
			}

			if p.Selector != "" {
				match, errs := selector.Eval(p.Selector, inst.Tags)
				if errs != nil {
					return fmt.Errorf("provider %q: selector error: %w", p.Provider, errs)
				}
				if !match {
					continue
				}
			}

			if _, exists := hosts[inst.Name]; exists {
				continue
			}

			host, err := instanceToHost(inst, p)
			if err != nil {
				return fmt.Errorf("provider %q: host %q: %w", p.Provider, inst.Name, err)
			}
			hosts[inst.Name] = host

			for _, t := range p.Targets {
				targets[t] = append(targets[t], inst.Name)
			}
		}
	}
	return nil
}

func fetchInstances(ctx context.Context, p Provider) ([]cloud.CloudInstance, error) {
	switch p.Provider {
	case "hetzner":
		token := p.Config["token"]
		return cloud.HetznerInstances(ctx, cloud.HetznerConfig{Token: token})
	case "aws":
		return cloud.AWSInstances(ctx, cloud.AWSConfig{
			Profile: p.Config["profile"],
			Region:  p.Config["region"],
		})
	default:
		return nil, fmt.Errorf("unknown provider %q", p.Provider)
	}
}

func instanceToHost(inst cloud.CloudInstance, p Provider) (ssh.Host, error) {
	addr := ""
	if inst.PublicIPv4 != nil {
		addr = inst.PublicIPv4.String()
	} else if inst.PublicIPv6 != nil {
		addr = inst.PublicIPv6.String()
	}

	host := ssh.Host{
		Name:    inst.Name,
		Address: addr,
		User:    p.SSH.User,
		Port:    p.SSH.Port,
	}

	if p.SSH.IdentityFile != "" {
		keyPath, err := utils.ExpandPath(p.SSH.IdentityFile)
		if err != nil {
			return ssh.Host{}, fmt.Errorf("failed to expand identity_file: %w", err)
		}
		host.KeyPath = keyPath
	}

	return host, nil
}
