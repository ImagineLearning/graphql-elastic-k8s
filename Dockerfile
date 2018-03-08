FROM alpine:3.6
EXPOSE 6000

COPY graphql /

ENTRYPOINT ["/graphql"]
