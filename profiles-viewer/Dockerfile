# Stage 1: Build Node dependencies
FROM --platform=$BUILDPLATFORM node:20.9.0 AS deps
WORKDIR /app
COPY profiles-viewer/webapp/package.json profiles-viewer/webapp/yarn.lock* ./
RUN yarn install --frozen-lockfile || yarn install

# Stage 2: Build the React app
FROM --platform=$BUILDPLATFORM node:20.9.0 AS frontend-builder
WORKDIR /webapp
COPY --from=deps /app/node_modules ./node_modules
COPY profiles-viewer/webapp .
RUN yarn build

# Stage 3: Build Go binary with embedded static files
FROM --platform=$BUILDPLATFORM golang:1.23 AS backend-builder
WORKDIR /app
COPY . .
# Replace webapp dist with built frontend assets, preserving embed.go
RUN rm -rf profiles-viewer/webapp/dist profiles-viewer/webapp/src profiles-viewer/webapp/node_modules
COPY --from=frontend-builder /webapp/dist profiles-viewer/webapp/dist
WORKDIR /app/profiles-viewer
ARG TARGETARCH
ARG LD_FLAGS
RUN CGO_ENABLED=0 GOARCH=$TARGETARCH go build -ldflags="${LD_FLAGS}" -o profiles-viewer ./cmd

# Stage 4: Final image
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=backend-builder /app/profiles-viewer/profiles-viewer .
USER 65532:65532
EXPOSE 8080 4317
ENTRYPOINT ["/app/profiles-viewer"]
