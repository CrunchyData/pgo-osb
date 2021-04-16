ARG BASE_IMAGE_OS
ARG DOCKERBASEREGISTRY
FROM ${DOCKERBASEREGISTRY}${BASE_IMAGE_OS}

ARG RELVER

LABEL Vendor="Crunchy Data Solutions" \
        Version="${RELVER}" \
        Release="${RELVER}" \
        summary="Crunchy Data pgo-osb open service broker " \
        description="Crunchy Data PostgreSQL Operator - pgo-osb "


ADD ./pgo-osb /usr/local/bin/

CMD /usr/local/bin/pgo-osb --help
