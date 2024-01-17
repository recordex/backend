# syntax=docker/dockerfile:1

FROM golang:1.21

# Build diff-pdf
WORKDIR /
RUN git clone https://github.com/vslavik/diff-pdf.git
WORKDIR /diff-pdf
# 参考
# https://docs.docker.jp/engine/articles/dockerfile_best-practice.html
RUN apt-get update && \
    apt-get -y install make automake g++ libglib2.0-dev libpoppler-glib-dev wx-common libwxgtk3.2-dev && \
    apt-get clean
RUN ./bootstrap
RUN ./configure
RUN make
# diff-pdf にパスを通す
ENV PATH=$PATH:/diff-pdf

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY ./ ./

RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /backend

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/engine/reference/builder/#expose
EXPOSE 8080

# Run
CMD ["/backend"]
