# MediaStorm

MediaStorm generates traffic to AWS Elemental MediaStore.

## Building

```
go get -u -d github.com/aws/aws-sdk-go
go build
```

## Usage

Use the [AWS Console][aws-console], CLI or API to create a MediaStore container
or get the *Data endpoint* for an existing container.

[aws-console]: https://console.aws.amazon.com/mediastore/home

Configure the [AWS credentials][aws-creds] for CLI usage, e.g., by setting the
[environment variables][aws-envs] `AWS_ACCESS_KEY_ID` and
`AWS_SECRET_ACCESS_KEY`.

[aws-creds]: https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html
[aws-envs]: https://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html

Set the AWS region using config files or the `AWS_REGION` environment variable.

Run:

```
mediastorm -endpoint https://CONTAINER.data.mediastore.REGION.amazonaws.com
```

For more options, see a list of available flags:

```
mediastorm -h
```
