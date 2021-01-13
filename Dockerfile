FROM golang:1.15-alpine AS build

WORKDIR /shoelaces
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w -extldflags "-static"' -o /tmp/shoelaces . && \
printf "---\nnetworkMaps:\n" > /tmp/mappings.yaml

# Final container has basically nothing in it but the executable
FROM scratch
COPY --from=build /tmp/shoelaces /shoelaces

WORKDIR /data
COPY --from=build /tmp/mappings.yaml mappings.yaml
COPY --from=build /shoelaces/web /web

ENV BIND_ADDR=0.0.0.0:8081
EXPOSE 8081

ENTRYPOINT ["/shoelaces"]
CMD ["-data-dir", "/data", "-static-dir", "/web"]
