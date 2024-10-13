# #!/bin/zsh
# scp -r ./build/default www@newtoio:~
# ssh ec2-user@newtoio > /dev/null 2>&1 << eeooff
# sudo rm -rf /data/www/one-api/default
# sudo mv ~/default /data/www/one-api/
# sudo chown www:www /data/www/one-api/default
# eeooff
# echo done!


#!/bin/zsh
ssh www@161.189.145.33 << eeooff
rm -rf /data/www/one-api/default
eeooff
echo delete_old_file

scp -r  ./build/default www@161.189.145.33:/data/www/one-api/