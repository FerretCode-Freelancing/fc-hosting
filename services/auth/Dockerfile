FROM node:18-alpine
RUN adduser -D nonroot \
  && echo "nonroot ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/nonroot \
  && chmod 0440 /etc/sudoers/nonroot
USER nonroot
ADD package.json /tmp/package.json
RUN cd /tmp && npm install && \
  mkdir -p /usr/route && cp -a /tmp/node_modules /usr/route/
COPY . ./usr/route
WORKDIR /usr/route
CMD [ "node", "." ]