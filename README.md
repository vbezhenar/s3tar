# S3 tar

Archives selected objects from the source S3 storage using tar format into the
destination S3 storage.

This project is in "Work in Progress" state. Do not use it.

## Sample usage

```
export SRC_ENDPOINT=https://object.pscloud.io/
export SRC_REGION=kz-ala-1
export SRC_BUCKET=12345-test-src
export SRC_ACCESS_KEY=ABC123
export SRC_SECRET_KEY=AbCd1234

export SRC_PREFIX=data/

export TAR_ENDPOINT=https://object.pscloud.io/
export TAR_REGION=kz-ala-1
export TAR_BUCKET=12345-test-dst
export TAR_ACCESS_KEY=ABC123
export TAR_SECRET_KEY=AbCd1234

export TAR_PREFIX=archive/tar/
export LST_PREFIX=archive/list/

export TAR_FORMAT=USTAR

./s3tar
```

## Algorithm

1. Download all lst objects from the listing storage, and build a set of (key,
   last-modified, etag) values.
2. List objects in the source storage and exclude objects already contained in
   the set built in the first step.
3. Archive objects from the source storage into tar format and upload into the
   tar storage.
4. Write corresponding (key, last-modified, etag) values into listing object and
   upload into the listing storage;

Tar archive and lst object use key names derived from the program start UTC
timestamp. Example: `2022-10-26_13-05-59.tar`.

## Parameters

All parameters can be specified using environment variables or command line
switches.

For example `SRC_PREFIX` is an environment variable name and `-srcprefix` is a
corresponding command line switch.

Command line switches override corresponding environment variables.

`(SRC|TAR|LST)_(ENDPOINT|REGION|BUCKET|ACCESS_KEY|SECRET_KEY|SESSION_TOKEN)`: S3
connection parameters for source/tar/listing storage. If all LST connection
parameters are omitted, TAR connection parameters will be used.
[Default configuration source](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/config)
will be used if some parameters are omitted.

`SRC_PREFIX` select only keys with the specified prefix.

`TAR_PREFIX` use this prefix for storing tar objects.

`LST_PREFIX` use this prefix for locating and storing listing objects.

`TAR_FORMAT` one of `USTAR`, `PAX`, `GNU`. `USTAR` allows only ASCII keys and
has various size limitations, most notably 8GiB per archive. `PAX` and `GNU` do
not have any realistic limits.
[Documenation](https://pkg.go.dev/archive/tar#Format) suggests to prefer `PAX`
format. Default value is `PAX`.

## Directions for future improvements

- Use bloom filter to reduce RAM usage for huge storages.
- Gzip tar archives before uploading them into the tar storage. Other
  compression algorithms might be useful as well.
- Gunzip source objects before archiving them into the tar arhive. It might be
  useful to achieve better compression.
- Encrypt tar archives using age library and asymmetric or symmetric encryption.
- Encrypt listing objects using symmetric encryption.
- Deduplication: if some object was archived before but now has a new name,
  store symbolic link instead of its data.
- Allow using multiple sources and a single tar/lst destination. It means that
  multiple s3tar invocations will use the same destination. It meant for
  deduplication. Use-case: some object is moved from fast storage to the cold
  storage. It should not be archived twice.
