# Integration harness: build the yup-split binary for Linux, then run it inside a
# Debian image (GNU coreutils) so its flags, inputs, and outputs can be checked
# against the real `split` reference. Invoked by `make integration`.
FROM golang:1.26 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/yup-split .

FROM debian:stable-slim
COPY --from=build /out/yup-split /usr/local/bin/yup-split
COPY integration/run.sh /run.sh
ENTRYPOINT ["/bin/sh", "/run.sh"]
