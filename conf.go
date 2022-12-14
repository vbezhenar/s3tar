package main

import (
	"flag"
	"fmt"
	"os"
)

type conf struct {
	src       s3Conf
	tar       s3Conf
	lst       s3Conf
	tarFormat string
}

type s3Conf struct {
	endpoint     string
	region       string
	accessKey    string
	secretKey    string
	sessionToken string
	bucket       string
	prefix       string
}

func loadConf() (*conf, []error) {
	c := &conf{
		tarFormat: "PAX",
	}

	if err := c.readFromEnvs(); err != nil {
		return nil, []error{err}
	}

	c.readFromFlags()

	if c.lst == (s3Conf{}) {
		c.lst = c.tar
	}

	return c, c.validate()
}

func (c *conf) readFromEnvs() error {
	c.src.readFromEnvs("SRC")
	c.tar.readFromEnvs("TAR")
	c.lst.readFromEnvs("LST")
	if s := os.Getenv("TAR_FORMAT"); len(s) > 0 {
		c.tarFormat = s
	}
	return nil
}

func (c *s3Conf) readFromEnvs(prefix string) {
	c.endpoint = os.Getenv(prefix + "_ENDPOINT")
	c.region = os.Getenv(prefix + "_REGION")
	c.accessKey = os.Getenv(prefix + "_ACCESS_KEY")
	c.secretKey = os.Getenv(prefix + "_SECRET_KEY")
	c.sessionToken = os.Getenv(prefix + "_SESSION_TOKEN")
	c.bucket = os.Getenv(prefix + "_BUCKET")
	c.prefix = os.Getenv(prefix + "_PREFIX")
}

func (c *conf) readFromFlags() {
	c.src.flags("src")
	c.tar.flags("tar")
	c.lst.flags("lst")
	flag.StringVar(&c.tarFormat, "tarformat", c.tarFormat, "tar format (USTAR|PAX|GNU)")
	flag.Parse()
}

func (c *s3Conf) flags(prefix string) {
	flag.StringVar(&c.endpoint, prefix+"endpoint", c.endpoint, prefix+" endpoint")
	flag.StringVar(&c.region, prefix+"region", c.region, prefix+" region")
	flag.StringVar(&c.accessKey, prefix+"accesskey", c.accessKey, prefix+" access key")
	flag.StringVar(&c.secretKey, prefix+"secretkey", c.secretKey, prefix+" secret key")
	flag.StringVar(&c.sessionToken, prefix+"sessiontoken", c.sessionToken, prefix+" session token")
	flag.StringVar(&c.bucket, prefix+"bucket", c.bucket, prefix+" bucket")
	flag.StringVar(&c.prefix, prefix+"prefix", c.prefix, prefix+" prefix")
}

func (c *conf) validate() []error {
	var errs []error
	errs = append(errs, c.src.validate("src")...)
	errs = append(errs, c.tar.validate("tar")...)
	errs = append(errs, c.lst.validate("lst")...)
	if c.tarFormat != "USTAR" && c.tarFormat != "PAX" && c.tarFormat != "GNU" {
		errs = append(errs, fmt.Errorf("tar format must be one of USTAR, PAX, GNU"))
	}
	return errs
}

func (c *s3Conf) validate(prefix string) []error {
	var errs []error
	if c.bucket == "" {
		errs = append(errs, fmt.Errorf(prefix+" bucket is not set"))
	}
	return errs
}
