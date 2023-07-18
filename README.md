# 介绍

该项目为 `cert-manager` 的一个webhook插件, 用于对接华为云DNS服务, 实现自动化证书签发和续期.

# 使用说明

1. clone代码, 执行`helm package ./deploy/huawei-webhook/`生成helm安装包.

2. 修改`MakeFile`的`IMAGE_NAME`值,  执行`make build`生成镜像, 并推送镜像.

3. 在k8s集群内部安装第一步生成的helm包, 修改`values.yaml`的`groupName`值(例如公司域名)和镜像地址.

4. 安装[reflector]([emberstack/kubernetes-reflector: Custom Kubernetes controller that can be used to replicate secrets, configmaps and certificates. (github.com)](https://github.com/EmberStack/kubernetes-reflector)), 用于自动同步申请后的证书到其它命名空间.

5. 配置`Issuer`和`Certificate`

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
