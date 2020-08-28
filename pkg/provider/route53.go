package provider

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"math"
)

const (
	Route53ProviderType = "route53"
)

type Route53Confg struct {
	Sess *session.Session

	HostedZoneId          string
	RecordName            string
	SourceIdentifier      string
	DestinationIdentifier string
}

type Route53Provider struct {
	client *route53.Route53
	config *Route53Confg
}

func (p *Route53Provider) getResourceRecordSets() (sourceResourceRecordSet *route53.ResourceRecordSet, destinationResourceRecordSet *route53.ResourceRecordSet, err error) {
	var startRecordIdentifier *string
	var startRecordName = aws.String(p.config.RecordName)
	var startRecordType *string
	var isTruncated = false
	for !isTruncated {
		res, err := p.client.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
			HostedZoneId:          aws.String(p.config.HostedZoneId),
			StartRecordIdentifier: startRecordIdentifier,
			StartRecordName:       startRecordName,
			StartRecordType:       startRecordType,
		})
		if err != nil {
			return nil, nil, err
		}
		for _, r := range res.ResourceRecordSets {
			if aws.StringValue(r.Name) != p.config.RecordName {
				continue
			}
			if aws.StringValue(r.SetIdentifier) == p.config.SourceIdentifier {
				sourceResourceRecordSet = r
				continue
			}
			if aws.StringValue(r.SetIdentifier) == p.config.DestinationIdentifier {
				destinationResourceRecordSet = r
				continue
			}
		}
		startRecordIdentifier = res.NextRecordIdentifier
		startRecordName = res.NextRecordName
		startRecordType = res.NextRecordType

		isTruncated = res.IsTruncated != nil && *res.IsTruncated == true
	}
	if sourceResourceRecordSet == nil {
		return nil, nil, err
	}
	if destinationResourceRecordSet == nil {
		return nil, nil, err
	}

	return sourceResourceRecordSet, destinationResourceRecordSet, nil
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
					Action: aws.String("UPSERT"),
					ResourceRecordSet: sourceResourceRecordSet,
				},
				{
					Action: aws.String("UPSERT"),
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
	client := route53.New(config.Sess)

	return &Route53Provider{
		client: client,
		config: config,
	}, nil
}
