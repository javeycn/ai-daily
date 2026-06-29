# AI Daily 爬虫容器
# 基于 Alpine 极简镜像（Go 二进制是静态链接，无需 glibc）
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 二进制通过 volume 挂载，不在镜像中
# 启动命令由外部 docker compose run 指定
ENTRYPOINT ["/app/ai-news-crawler"]
CMD ["--config", "/app/config-prod.yaml"]
