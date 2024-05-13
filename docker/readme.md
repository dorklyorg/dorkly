Docker

Env vars:
1. `S3_URL`: S3 url containing ld-relay archive. Example: 's3://dorkly/flags.tar.gz'
1. `AWS_REGION`: AWS region containing the SQS queue. This should match the sqs queue's region. Example: 'us-west-2'. 
2. `SQS_QUEUE_URL`: SQS queue where we listen for changes to S3 bucket. Example: 'https://sqs.us-west-1.amazonaws.com/310766071441/dorkly'
3. `AWS_ACCESS_KEY_ID`: AWS access key id (or omit if using instance role etc)
4. `AWS_SECRET_ACCESS_KEY`: AWS secret access key (or omit if using instance role etc)

To Build and push: (platform flag is required for AWS Lightsail)
`docker build --platform=linux/amd64 -t  drichelson/dorkly:latest . && docker push drichelson/dorkly:latest`

You'll want to expose port 8030.
Health check is at port `8030/status`