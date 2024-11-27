
# 无法获取gpt-3.5-turbo/gpt-4o 令牌编码器
https://github.com/songquanpeng/one-api/issues/906
```
1.下载 cl100k_base.tiktoken --重命名--> 9b5ad71b2ce5302211f9c61530b329a4922fc6a4

2.下载 o200k_base.tiktoken  --重命名--> fb374d419588a4632f3f557e76b4b70aebbca790

3. 在项目下新建一个目录，比如.cache 把上面两个文件放到该目录下

4. 在.env文件中，添加 TIKTOKEN_CACHE_DIR=./.cache
```