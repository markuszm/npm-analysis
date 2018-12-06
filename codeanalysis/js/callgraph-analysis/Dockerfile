FROM node:8-slim

RUN groupadd -r analysis && useradd --no-log-init --system --gid analysis analysis -d /analyzer
RUN mkdir /analyzer && chown -R analysis:analysis /analyzer
USER analysis

COPY --chown=analysis src /analyzer
COPY --chown=analysis analyze.sh /analyzer/
COPY --chown=analysis package.json /analyzer/
COPY --chown=analysis tsconfig.json /analyzer/
COPY --chown=analysis unittest.sh /analyzer/

WORKDIR /analyzer
RUN yarn install
RUN npm run build

WORKDIR /
CMD ["analyzer/analyze.sh"]
