# ---- 构建阶段 (builder) ----
# 类比 Java：这一层像 `FROM maven ... RUN mvn package`——用带完整工具链的大镜像编译。
FROM golang:1.26-alpine AS build
WORKDIR /src

# 先只拷 go.mod/go.sum 再 download：利用 Docker 层缓存。
# 只要依赖没变，这一层就命中缓存，不会每次改代码都重新拉依赖。
COPY go.mod go.sum ./
RUN go mod download

# 再拷源码、编译。CGO_ENABLED=0 → 编出静态二进制，不依赖 glibc，能丢进极简镜像跑。
COPY . .
RUN CGO_ENABLED=0 go build -o /app/server .

# ---- 运行阶段 (runtime) ----
# 类比 Java：这一层像 `FROM eclipse-temurin:jre` + `COPY app.jar`——只装运行所需，镜像小、攻击面小。
FROM alpine:3.20
WORKDIR /app

# 只从构建阶段拷出编译好的二进制 + 配置目录，编译工具链一概不带进来。
COPY --from=build /app/server .
COPY config ./config

EXPOSE 8080
CMD ["./server"]
