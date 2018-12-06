FROM node:latest

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       python3 \
       jq \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV PYTHONIOENCODING utf-8

RUN groupadd -r analysis && useradd --no-log-init --system --gid analysis analysis -d /analyzer
RUN mkdir /analyzer && chown -R analysis:analysis /analyzer
COPY src /analyzer
RUN mkdir -p /analysis/inputs && mkdir -p /analysis/output && chmod -R 777 /analysis

USER analysis
CMD ["/analyzer/run.sh"]
