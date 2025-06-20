FROM golang:1.13.0-alpine

WORKDIR /go/src/zxq.co/ripple/hanayo
COPY . .

# Set GOPATH and disable Go modules
ENV GOPATH=/go
ENV GO111MODULE=off

# Build the application
RUN CGO_ENABLED=0 go install -v ./...

FROM alpine:3.10
WORKDIR /hanayo/
COPY --from=0 /go/bin/hanayo ./
COPY --from=0 /go/src/zxq.co/ripple/hanayo/ ./

# Agree to License
RUN mkdir -p ~/.config && touch ~/.config/ripple_license_agreed

EXPOSE 45221
CMD [ "./hanayo" ]
