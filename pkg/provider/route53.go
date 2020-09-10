package provider

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"math"
	"regexp"
	"strings"
)

const (
	Route53ProviderType = "route53"
)

type Route53Confg struct {
	Sess *session.Session

	Client Route53Client

	HostedZoneId          string
	RecordName            string
	Type                  string
	SourceIdentifier      string
	DestinationIdentifier string
}

type Route53Client interface {
	ListResourceRecordSets(input *route53.ListResourceRecordSetsInput) (*route53.ListResourceRecordSetsOutput, error)
	ChangeResourceRecordSets(input *route53.ChangeResourceRecordSetsInput) (*route53.ChangeResourceRecordSetsOutput, error)
}

type Route53Provider struct {
	client Route53Client
	config *Route53Confg
}

func (p *Route53Provider) TargetName() string {
	return fmt.Sprintf("AWS/Route53/%s", p.config.RecordName)
}

func (p *Route53Provider) matchPattern(pattern string, s string) bool {
	if pattern == s {
		return true
	}
	if strings.Index(s, pattern) != -1 {
		return true
	}
	if matched, err := regexp.MatchString(pattern, s); matched == true && err != nil {
		return true
	}

	return false
}

func (p *Route53Provider) getResourceRecordSets() (src *route53.ResourceRecordSet, dest *route53.ResourceRecordSet, err error) {
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId:          aws.String(p.config.HostedZoneId),
	}
	if strings.HasSuffix(p.config.RecordName, ".") {
		input.StartRecordName = aws.String(p.config.RecordName)
	}
	if p.config.Type != "" {
		input.StartRecordType = aws.String(p.config.Type)
	}
	for isTruncated := true; isTruncated == true && (src == nil || dest == nil); {
		res, err := p.client.ListResourceRecordSets(input)
		if err != nil {
			return nil, nil, err
		}
		for _, r := range res.ResourceRecordSets {
			if src != nil && dest != nil {
				break
			}
			if p.config.Type != "" && !p.matchPattern(p.config.Type, aws.StringValue(r.Type)) {
				continue
			}
			if p.config.RecordName != "" && !p.matchPattern(p.config.RecordName, aws.StringValue(r.Name)) {
				continue
			}
			if p.matchPattern(p.config.SourceIdentifier, aws.StringValue(r.SetIdentifier)) {
				src = r
				continue
			}
			if p.matchPattern(p.config.DestinationIdentifier, aws.StringValue(r.SetIdentifier))  {
				dest = r
				continue
			}
		}
		input.StartRecordIdentifier = res.NextRecordIdentifier
		input.StartRecordName = res.NextRecordName
		input.StartRecordType = res.NextRecordType

		isTruncated = res.IsTruncated != nil && *res.IsTruncated == true
	}
	if src == nil || dest == nil {
		return nil, nil, err
	}
	return src, dest, nil
}

func (p *Route53Provider) Get() (percentage float64, err error) {
	sourceResourceRecordSet, destinationResourceRecordSet, err := p.getResourceRecordSets()
	if err != nil {
		return -1, nil
	}

	totalWeight := aws.Int64Value(sourceResourceRecordSet.Weight) + aws.Int64Value(destinationResourceRecordSet.Weight)
	if totalWeight == 0 {
		return -1, nil
	}

	return float64(aws.Int64Value(destinationResourceRecordSet.Weight)) / float64(totalWeight) * 100, nil
}

func (p *Route53Provider) Update(percentage float64) error {
	sourceResourceRecordSet, destinationResourceRecordSet, err := p.getResourceRecordSets()
	if err != nil {
		return nil
	}

	sourceResourceRecordSet.Weight = aws.Int64(int64(100 - math.Floor(percentage)))
	destinationResourceRecordSet.Weight = aws.Int64(int64(math.Round(percentage)))

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action:            aws.String("UPSERT"),
					ResourceRecordSet: sourceResourceRecordSet,
				},
				{
					Action:            aws.String("UPSERT"),
					ResourceRecordSet: destinationResourceRecordSet,
				},
			},
		},
		HostedZoneId: aws.String(p.config.HostedZoneId),
	}

	if _, err := p.client.ChangeResourceRecordSets(input); err != nil {
		return err
	}

	return nil
}

func NewRoute53Provider(config *Route53Confg) (*Route53Provider, error) {
	if config.HostedZoneId == "" {
		return nil, errors.New("Route53Config.HostedZoneId is missing")
	}
	if config.SourceIdentifier == "" {
		return nil, errors.New("Route53Config.SourceIdentifier must be a string or a regular expression")
	}
	if config.DestinationIdentifier == "" {
		return nil, errors.New("Route53Config.DestinationIdentifier must be a string or a regular expression")
	}

	client := config.Client

	if client == nil {
		if config.Sess == nil {
			return nil, errors.New("Route53Config.Sess must be set when Route53Config.Client is missing")
		}
		client = route53.New(config.Sess)
	}

	return &Route53Provider{
		client: client,
		config: config,
	}, nil
}
