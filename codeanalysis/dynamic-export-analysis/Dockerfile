FROM node:8-slim

RUN apt update && apt install -y jq

WORKDIR /analyzer

COPY . .

RUN yarn add readdirp --silent --non-interactive --no-lockfile 1>&2

ENTRYPOINT [ "./analyze.sh" ]