FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY ds-operator /ds-operator
USER nonroot:nonroot
ENTRYPOINT ["/ds-operator"]
