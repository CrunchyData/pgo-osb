ARG BASEOS
ARG DOCKERBASEREGISTRY
FROM ${DOCKERBASEREGISTRY}${BASEOS}

ARG RELVER

LABEL Vendor="Crunchy Data Solutions" \
        Version="${RELVER}" \
        Release="${RELVER}" \
        summary="Crunchy Data pgo-osb open service broker " \
        description="Crunchy Data PostgreSQL Operator - pgo-osb "

COPY redhat/atomic/help.1 /help.1
COPY redhat/atomic/help.md /help.md
COPY licenses /licenses

ADD ./pgo-osb /usr/local/bin/

CMD /usr/local/bin/pgo-osb --help
