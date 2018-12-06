FROM node:8-slim

RUN groupadd -r analysis && useradd --no-log-init --system --gid analysis analysis -d /analyzer
RUN mkdir /analyzer && chown -R analysis:analysis /analyzer
USER analysis
# TODO(markuszm): Either run this after building with typescript (copying from dist) or we can
# shuffle this directory to have all the typescript under 'src', then build it from within the
# dockerfile
COPY --chown=analysis dist /analyzer
COPY --chown=analysis package.json /analyzer/

WORKDIR /analyzer
RUN yarn install

WORKDIR /
CMD ["analyzer/analyze.sh"]
