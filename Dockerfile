ARG RELEASE_TAG=develop 
ARG API_DOMAIN=api.resonate.coop
ARG APP_HOST=https://stream.resonate.coop
ARG STATIC_HOSTNAME=dash.resonate.coop
ARG API_BASE=/api/v3
ARG NODE_ENV=development
ARG BUILD_ENV=production

# Frontend build stage
FROM node:18-alpine as builder

ARG RELEASE_TAG
ARG NODE_ENV
ARG BUILD_ENV
ARG API_DOMAIN
ARG API_BASE
ARG APP_HOST
ARG STATIC_HOSTNAME

ENV API_DOMAIN $API_DOMAIN
ENV NODE_ENV $NODE_ENV
ENV API_BASE $API_BASE
ENV APP_HOST $APP_HOST
ENV STATIC_HOSTNAME $STATIC_HOSTNAME

WORKDIR /build

RUN apk --no-cache add git

RUN cd /build && git clone --branch ${RELEASE_TAG} --single-branch --depth 1 https://github.com/resonatecoop/id

RUN cd /build/id/frontend && npm install && npm install -g gulp

ENV NODE_ENV $BUILD_ENV

RUN cd /build/id/frontend && npm run build

# Backend build stage
FROM golang:latest

ARG RELEASE_TAG

RUN mkdir /build

WORKDIR /build

RUN export GO111MODULE=on
RUN apt-get -y update
RUN cd /build && git clone --branch ${RELEASE_TAG} --single-branch --depth 1 https://github.com/resonatecoop/id

RUN cd id && go build

EXPOSE 11000

WORKDIR /build/id

COPY --from=builder /build/id/data /build/id/data/
COPY --from=builder /build/id/public /build/id/public/

ENTRYPOINT ["sh", "docker-entrypoint.sh"]
