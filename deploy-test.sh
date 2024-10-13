#!/bin/zsh
ssh www@161.189.145.33 << eeooff
rm /data/www/one-api/oneapi
eeooff
echo delete_oneapi

scp ./oneapi www@161.189.145.33:/data/www/one-api/
ssh www@161.189.145.33 << eeooff
cd /data/www/one-api
pm2 restart ecosystem.config.cjs
eeooff
echo done!
