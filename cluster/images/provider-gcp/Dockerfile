FROM BASEIMAGE
RUN apk --no-cache add ca-certificates bash

ARG ARCH
ARG TINI_VERSION

ADD provider /usr/local/bin/crossplane-gcp-provider

COPY stack-package /

EXPOSE 8080
ENTRYPOINT ["crossplane-gcp-provider"]
