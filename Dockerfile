FROM node:22-bookworm-slim AS frontend
WORKDIR /build
COPY fe_calculator/package*.json ./
RUN npm ci
COPY fe_calculator/ ./
RUN npm run build

FROM golang:1.26-bookworm AS backend
WORKDIR /build
COPY be_calculator/ ./
RUN CGO_ENABLED=0 go build -trimpath -o /out/calculator-api .

FROM scratch
COPY --from=backend /out/calculator-api /calculator-api
COPY --from=frontend /build/dist/ /web/
ENV STATIC_DIR=/web
EXPOSE 3000
USER 65532:65532
ENTRYPOINT ["/calculator-api"]
