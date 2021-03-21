FROM scratch
COPY metrics-generator /
ENTRYPOINT ["/metrics-generator"]
