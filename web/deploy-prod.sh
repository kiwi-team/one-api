#!/bin/zsh
# scp -r ./build/default ec2-user@llmprod:~
# ssh ec2-user@llmprod > /dev/null 2>&1 << eeooff
# sudo rm -rf /data/www/one-api/default
# sudo mv ~/default /data/www/one-api/
# sudo chown www:www /data/www/one-api/default
# eeooff
# echo done!


#!/bin/zsh
ssh www@52.83.39.13 << eeooff
rm -rf /data/www/one-api/default
eeooff
echo delete_old_file

scp -r  ./build/default www@161.189.145.33:/data/www/one-api/