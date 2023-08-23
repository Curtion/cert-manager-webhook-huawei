# 介绍

该项目为 `cert-manager` 的一个webhook插件, 用于对接华为云DNS服务, 实现自动化证书签发和续期.

# 使用说明

1. 下载`release`中最新helm包, 在k8s中安装它, 记得修改`groupName`的值(例如公司域名)
2. 安装[reflector](https://github.com/EmberStack/kubernetes-reflector), 用于自动同步申请后的证书到其它命名空间.
3. 配置`Issuer`和`Certificate`

   ```yaml
   apiVersion: cert-manager.io/v1
   kind: ClusterIssuer
   metadata:
     name: letsencrypt
   spec:
     acme:
       email: curtion@126.com
       server: https://acme-v02.api.letsencrypt.org/directory
       privateKeySecretRef:
         name: letsencrypt
       solvers:
         - dns01:
             webhook:
               config:
                 region: cn-southwest-2
                 AK: XKCD2EQDHF9XGIS851R7
                 SK: tnYnXON5GBzpfl5Ey50MeTvIwA7IRTVbsRqaLy6D
                 ZoneName: jidian-iot.cn
               groupName: acme.jidian-iot.cn
               solverName: huawei-solver
   ---
   apiVersion: cert-manager.io/v1
   kind: Certificate
   metadata:
     name: jidian-iot-tls
   spec:
     secretName: jidian-iot-tls
     dnsNames:
       - "*.jidian-iot.cn"
     issuerRef:
       name: letsencrypt
       kind: ClusterIssuer
     secretTemplate:
       annotations:
         reflector.v1.k8s.emberstack.com/reflection-allowed: "true"
         reflector.v1.k8s.emberstack.com/reflection-allowed-namespaces: ""
         reflector.v1.k8s.emberstack.com/reflection-auto-enabled: "true"
         reflector.v1.k8s.emberstack.com/reflection-auto-namespaces: ""
   
   ```

   上述配置会尝试申请`*.jidian-iot.cn`泛域名证书, 并且把证书名命名为`jidian-iot-tls`并放置到`default`命名空间中, 然后`reflector`会自动把证书同步到其它命名空间中.

# yaml配置说明

- region: 区域信息,参考[华为云文档](https://developer.huaweicloud.com/endpoint?DNS)
- AK: 华为云AK
- SK: 华为云SK
- ZoneName: 域名
- groupName: 和安装webhook时的值保持一致
- solverName: 固定为`huawei-solver`, 不可修改
- reflector.v1.k8s.emberstack.com/*: 参考[reflector](https://github.com/EmberStack/kubernetes-reflector)说明

# 测试

修改`testdata/huawei-solver`中`config.json.default`文件名为`config.json`, 并修改其中的配置, 然后执行`make test`进行测试.

# 其它

当然你也可以构建自己的docker镜像, 执行`make build`即可, 只需要在安装helm包时修改镜像地址.