# Specify the version of Go to use
FROM golang:1.15

# Copy all the files from the host into the container
WORKDIR /src
COPY . .
COPY $GITHUB_EVENT_PATH $GITHUB_EVENT_PATH

# Enable Go modules
ENV GO111MODULE=on

RUN echo $GITHUB_EVENT_PATH
RUN cat $GITHUB_EVENT_PATH

# Compile the action
RUN go build -o /bin/action

# Specify the container's entrypoint as the action
ENTRYPOINT ["/bin/action"]
