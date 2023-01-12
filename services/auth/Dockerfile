FROM node:18-alpine
RUN adduser -D nonroot 
USER nonroot
ADD package.json /tmp/package.json
RUN cd /tmp && npm install && \
  mkdir -p /home/nonroot/route && cp -a /tmp/node_modules /home/nonroot/route/
COPY . ./home/nonroot/route
WORKDIR /home/nonroot/route
CMD [ "node", "." ]
