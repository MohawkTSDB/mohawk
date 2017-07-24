FROM fedora

# make sqlite base directory
RUN mkdir /root/server
RUN mkdir /root/ssh

# install bin files
COPY mohawk /usr/bin/

# set env variables
ENV HAWKULAR_FILE_PEM="/root/ssh/server.pem" \
  HAWKULAR_FILE_KEY="/root/ssh/server.key" \
  HAWKULAR_PORT=8443 \
  HAWKULAE_DB_DIR=./server \
  HAWKULAE_DB_URL=127.0.0.1 \
  HAWKULAE_FLAGS="--tls --gzip" \
  HAWKULAE_BACKEND=memory

# declare volume
VOLUME /root/ssh

# tell the port number the container should expose
EXPOSE $HAWKULAR_PORT

# run the application
WORKDIR /root
RUN chmod -R ugo+rwx /root/server
CMD /usr/bin/mohawk $HAWKULAE_FLAGS --port $HAWKULAR_PORT \
  --cert $HAWKULAR_FILE_PEM --key $HAWKULAR_FILE_KEY \
  --backend $HAWKULAE_BACKEND \
  --options "db-dirname=${HAWKULAE_DB_DIR}&db-url=${HAWKULAE_DB_URL}"
