FROM scratch
COPY dist/linux_amd64/terraform-provider-circleci /terraform-provider-circleci
ENTRYPOINT ["/terraform-provider-circleci"]
