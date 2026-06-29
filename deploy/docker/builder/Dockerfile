# AI Daily 前端构建容器
# 基于 Node.js 22 LTS（与 tc106 构建环境一致）
FROM node:22-alpine

RUN apk add --no-cache rsync tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 构建脚本
COPY build.sh /app/build.sh
RUN chmod +x /app/build.sh

ENTRYPOINT ["/app/build.sh"]
