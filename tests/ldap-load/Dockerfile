FROM gcr.io/forgeops-public/ds:7.2-dev

USER root 

RUN apt-get update && apt-get install -y python3

USER forgerock
WORKDIR /home/forgerock
COPY --chown=forgerock:root *.py /home/forgerock
COPY --chown=forgerock:root *.sh /home/forgerock

ENTRYPOINT ["/home/forgerock/load.sh"]