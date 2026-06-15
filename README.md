# s3server

The aim of this repository is to propose a simple *fake* `AWS` `S3` server that accept any authentication.

## How to build and run

Compile

```shell
go build ./cmd/s3server
```

Run

```shell
./s3server
```

## How to test

Start a shell with the AWS cli image

```shell
docker run --rm --interactive --tty \
  --volume $(pwd)/image.jpg:/aws/image.jpg \
  --env AWS_PROFILE=local \
  --env AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID:-test} \
  --env AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY:-test} \
  --env AWS_REGION=${AWS_REGION:-us-east-1} \
  --env AWS_ENDPOINT_URL=${AWS_ENDPOINT_URL:-http://localhost:9000} \
  --entrypoint bash \
  amazon/aws-cli
```

Prepare the *fake* credentials

```shell
aws configure set aws_access_key_id ${AWS_ACCESS_KEY_ID}
aws configure set aws_secret_access_key ${AWS_SECRET_ACCESS_KEY}
aws configure set default.region ${AWS_REGION}
```

Test it

```shell
aws s3 ls # List buckets
aws s3 mb s3://photos # Create a bucket
aws s3 cp image.jpg s3://photos/image.jpg # Upload an object
aws s3 ls s3://photos # List objects in the bucket
aws s3 rm s3://photos/image.jpg # Remove object
aws s3 rb s3://photos # Delete a bucket
```

## cURL commands

- create bucket

  ```bash
  curl -X PUT http://localhost:9000/photos
  ```

- list buckets

  ```bash
  curl http://localhost:9000/
  curl -s http://localhost:9000/ | xmllint --format -
  ```

- verify if a bucket exists

  ```bash
  curl -I http://localhost:9000/photos
  ```

- delete bucket

  ```bash
  curl -X DELETE http://localhost:9000/photos
  ```

- send a file

  ```bash
  curl \
      -T photo.jpg \
      http://localhost:9000/photos/photo.jpg
  ```

- send data from stdin

  ```bash
  echo "Hello S3" | \
  curl -X PUT \
       --data-binary @- \
       http://localhost:9000/photos/hello.txt
  ```

- download a file

  ```bash
  curl -o photo.jpg http://localhost:9000/photos/photo.jpg
  ```

- get metadata only

  ```bash
  curl -I http://localhost:9000/photos/photo.jpg
  ```

- delete an object

  ```bash
  curl -X DELETE \
      http://localhost:9000/photos/photo.jpg
  ```

- list objects in a bucket

  ```bash
  curl \
      "http://localhost:9000/photos?list-type=2"
  ```

- ... using a prefix

  ```bash
  curl \
      "http://localhost:9000/photos?list-type=2&prefix=images/"
  ```

- ... with pagination

  ```bash
  curl \
      "http://localhost:9000/photos?list-type=2&max-keys=10"
  ```

  > use `NextContinuationToken` to continue

  ```bash
  curl \
      "http://localhost:9000/photos?list-type=2&continuation-token=<token>"
  ```
