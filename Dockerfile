ARG RELEASE_TAG=develop 
ARG API_DOMAIN=api.resonate.coop
ARG APP_HOST=https://stream.resonate.coop
ARG STATIC_HOSTNAME=dash.resonate.coop
ARG API_BASE=/api/v3
ARG NODE_ENV=development

# Frontend build stage
FROM node:18-alpine as builder

WORKDIR /build

RUN apk --no-cache add git

ARG RELEASE_TAG

RUN cd /build && git clone --branch ${RELEASE_TAG} --single-branch --depth 1 https://github.com/resonatecoop/id

ENV NODE_ENV development

RUN cd /build/id/frontend && npm install && npm install -g gulp

ARG NODE_ENV
ARG API_DOMAIN
ARG API_BASE
ARG APP_HOST
ARG STATIC_HOSTNAME

ENV NODE_ENV $NODE_ENV
ENV API_DOMAIN $API_DOMAIN
ENV API_BASE $API_BASE
ENV APP_HOST $APP_HOST
ENV STATIC_HOSTNAME $STATIC_HOSTNAME

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
