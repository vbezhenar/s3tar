package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	conf, errs := loadConf()
	if len(errs) > 0 {
		log.Print("Cannot load config:")
		for _, err := range errs {
			log.Printf(" %v", err.Error())
		}
		os.Exit(1)
	}

	ctx := context.Background()

	srcS3Client, err := newS3Client(ctx, conf.src)
	if err != nil {
		log.Fatalf("Cannot initialize src S3 client: %v", err)
	}

	tarS3Client, err := newS3Client(ctx, conf.tar)
	if err != nil {
		log.Fatalf("Cannot initialize tar S3 client: %v", err)
	}

	lstS3Client, err := newS3Client(ctx, conf.lst)
	if err != nil {
		log.Fatalf("Cannot initialize lst S3 client: %v", err)
	}

	listings, err := loadListings(ctx, lstS3Client, conf.lst.bucket, conf.lst.prefix)
	if err != nil {
		log.Fatalf("Cannot load listings: %v", err)
	}

	for _, listing := range listings {
		fmt.Println(listing)
	}

	_ = srcS3Client
	_ = tarS3Client
}

func newS3Client(ctx context.Context, s3Conf s3Conf) (*s3.Client, error) {
	var optFns []func(*config.LoadOptions) error

	if s3Conf.endpoint != "" {
		r := aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				e := aws.Endpoint{
					URL:           s3Conf.endpoint,
					SigningRegion: region,
				}
				return e, nil
			},
		)
		optFns = append(optFns, config.WithEndpointResolverWithOptions(r))
	}

	if s3Conf.region != "" {
		optFns = append(optFns, config.WithRegion(s3Conf.region))
	}

	if s3Conf.accessKey != "" || s3Conf.secretKey != "" || s3Conf.sessionToken != "" {
		cp := credentials.NewStaticCredentialsProvider(
			s3Conf.accessKey,
			s3Conf.secretKey,
			s3Conf.sessionToken,
		)
		optFns = append(optFns, config.WithCredentialsProvider(cp))
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("cannot load aws config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig)

	return s3Client, nil
}

func loadListings(ctx context.Context, client *s3.Client, bucket string, prefix string) ([]string, error) {
	listInput := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	}
	paginator := s3.NewListObjectsV2Paginator(client, listInput)
	var lstStrings []string
	re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}-\d{3}Z\.lst$`)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot get next page: %w", err)
		}
		for _, o := range page.Contents {
			if !re.MatchString(*o.Key) {
				continue
			}

			getInput := s3.GetObjectInput{
				Bucket:            &bucket,
				Key:               o.Key,
				IfMatch:           o.ETag,
				IfUnmodifiedSince: o.LastModified,
			}
			getOutput, err := client.GetObject(ctx, &getInput)
			if err != nil {
				return nil, fmt.Errorf("cannot get object %v: %w", o.Key, err)
			}
			defer getOutput.Body.Close()
			scanner := bufio.NewScanner(getOutput.Body)
			for scanner.Scan() {
				lstStrings = append(lstStrings, scanner.Text())
			}
		}
	}

	return lstStrings, nil
}
