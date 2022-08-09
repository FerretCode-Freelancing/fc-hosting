FROM node:18-alpine
ADD package.json /tmp/package.json
RUN cd /tmp && npm install
RUN mkdir -p /usr/route && cp -a /tmp/node_modules /usr/route
COPY . ./usr/route
WORKDIR /usr/route
CMD [ "node", "." ]