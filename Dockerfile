# See https://github.com/zendesk/docker-images-base for more options.
FROM 713408432298.dkr.ecr.us-west-2.amazonaws.com/base/zendesk/docker-images-base/ubuntu:22.04
WORKDIR /app

# Run as non-root. Use GID 1000 to allow access to vault token from vault agent
USER 1000:1000

# Copy required files and run necessary preperations here.
COPY . .

# Update this to your application entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]
