lbagent
=======
nginx负载均衡agent

守护模式运行，通过rest方式交互

###功能（已实现）
1. 新增domain
2. 插入real server
3. 删除real server
4. 查询domain real servers
5. 删除 domain

###功能（要实现）
1. 替换所有 real server

###其他
**测试**

当前只在ubuntu server上测试过

**domain配置模板**

目录 template/nginx

**依赖**

  github.com/siddontang/go-log/log
  
  github.com/ant0ine/go-json-rest/rest
