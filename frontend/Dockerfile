FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

# NEXT_PUBLIC_* vars are baked into the client bundle at build time, not
# read at container start — must be passed as a build-arg, not a runtime
# "environment:" entry in docker-compose. Default matches local dev (browser
# on the host reaches the backend via the host's published port). CD
# pipelines override this with the real public URL — see cd-*.yml.
ARG NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
RUN npm run build

EXPOSE 3000

CMD ["npm", "run", "start"]
