FROM node:8-slim

RUN apt update && apt install -y jq

WORKDIR /analyzer

COPY . .

ENTRYPOINT [ "./analyze.sh" ]