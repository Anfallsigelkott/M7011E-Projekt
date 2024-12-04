FROM oven/bun

RUN mkdir /app

COPY package.json /app/package.json
WORKDIR /app

RUN bun i

COPY . /app

RUN bun run build

ENTRYPOINT ["bun", "run", "preview", "--host"]
