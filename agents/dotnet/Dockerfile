FROM busybox

ARG version

WORKDIR /autoinstrumentation

# TODO: Detect musl / libc during lang detection process and choose the right instrumentation
ADD https://github.com/open-telemetry/opentelemetry-dotnet-instrumentation/releases/download//v0.3.1-beta.1/opentelemetry-dotnet-instrumentation-linux-musl.zip .

RUN unzip opentelemetry-dotnet-instrumentation-linux-musl.zip

RUN chmod -R go+r .