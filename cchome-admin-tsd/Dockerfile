FROM registry.cn-shanghai.aliyuncs.com/goiot/alpine:3.16.3
RUN echo "Asia/Shanghai" > /etc/timezone
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone && mkdir -p /app/conf /app/views /app/static /app/swagger
COPY views /app/views/
COPY static /app/static/
COPY app /app/admin
WORKDIR /app/
EXPOSE 2010
EXPOSE 2011
CMD ["./admin"]