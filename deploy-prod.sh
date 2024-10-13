#!/bin/zsh
ssh www@52.83.39.13 << eeooff
rm /data/www/one-api/oneapi
eeooff
echo delete_oneapi

scp ./oneapi www@52.83.39.13:/data/www/one-api/
ssh www@52.83.39.13 << eeooff
cd /data/www/one-api
pm2 restart ecosystem.config.cjs
eeooff
echo done!
