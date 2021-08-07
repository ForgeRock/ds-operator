FROM gcr.io/forgerock-io/ds/pit1:7.2.0-5d3bee3563e4d4b239aadd1362bcd411e775d8c9

COPY --chown=forgerock:root . /opt/opendj/

# This creates sample "prototype" backends for the CDK deployment. This is not required
# and is done to accelerate the cdk startup.
# When complete, a tar file is created with the prototype backends.
RUN default-scripts/setup && tar cvfz idrepo.tar.z config db changelogDb locks var

FROM gcr.io/forgerock-io/ds/pit1:7.2.0-5d3bee3563e4d4b239aadd1362bcd411e775d8c9
WORKDIR /opt/opendj

COPY --from=0 --chown=forgerock:root /opt/opendj/idrepo.tar.z /opt/opendj/
COPY --chown=forgerock:root default-scripts /opt/opendj/default-scripts
COPY --chown=forgerock:root ldif-ext /opt/opendj/ldif-ext
COPY --chown=forgerock:root *.sh /opt/opendj/

