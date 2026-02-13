package cloud

import (
	"context"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type AWSConfig struct {
	Profile string
	Region  string
}

func AWSInstances(ctx context.Context, cfg AWSConfig) ([]CloudInstance, error) {
	var opts []func(*config.LoadOptions) error
	if cfg.Profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(cfg.Profile))
	}
	if cfg.Region != "" {
		opts = append(opts, config.WithRegion(cfg.Region))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("aws: failed to load config: %w", err)
	}

	if awsCfg.Region == "" {
		return nil, fmt.Errorf("aws: region is required (use config, AWS_REGION)")
	}

	client := ec2.NewFromConfig(awsCfg)

	var instances []CloudInstance
	var nextToken *string

	for {
		out, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			Filters: []types.Filter{
				{
					Name:   strPtr("instance-state-name"),
					Values: []string{"running"},
				},
			},
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("aws: failed to describe instances: %w", err)
		}

		for _, res := range out.Reservations {
			for _, inst := range res.Instances {
				ci := CloudInstance{
					Name: tagValue(inst.Tags, "Name"),
					Tags: tagsToMap(inst.Tags),
				}

				if inst.PublicIpAddress != nil {
					ci.PublicIPv4 = net.ParseIP(*inst.PublicIpAddress)
				}
				if len(inst.NetworkInterfaces) > 0 {
					for _, addr := range inst.NetworkInterfaces[0].Ipv6Addresses {
						if addr.Ipv6Address != nil {
							ci.PublicIPv6 = net.ParseIP(*addr.Ipv6Address)
							break
						}
					}
				}

				instances = append(instances, ci)
			}
		}

		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}

	return instances, nil
}

func tagValue(tags []types.Tag, key string) string {
	for _, t := range tags {
		if t.Key != nil && *t.Key == key {
			if t.Value != nil {
				return *t.Value
			}
			return ""
		}
	}
	return ""
}

func tagsToMap(tags []types.Tag) map[string]string {
	m := make(map[string]string, len(tags))
	for _, t := range tags {
		if t.Key != nil && t.Value != nil {
			m[*t.Key] = *t.Value
		}
	}
	return m
}

func strPtr(s string) *string {
	return &s
}
