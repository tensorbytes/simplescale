# SimpleAutoscaler

SimpleAutoscaler 项目是一个简单的 autoscaler 实现，分为两个CRD资源：
- `RecommendationScaleFactor`，一个用于根据用户自定义的指标计算出目标值扩缩容比例的资源对象。
- `SimpleAutoScaler`，一个用于落实扩缩容行为策略的资源对象，`SimpleAutoScaler`提供了灵活的方法读取扩缩容比例值来落实扩缩容策略。

做这个项目最开始目的是为了提高k8s集群的资源使用率，但在做的过程中希望有一个组件可以帮助落实一些k8s集群的pod配置策略，比如 Pod 的 requests、limits 怎么设置更合理。正好趁着过年休息期间，开发一个自动扩缩容的operator并记录下来当作一个 operator 的新手教程，让大家了解怎么去做 k8s 的开发。

[架构图](https://github.com/tensorbytes/simplescale/blob/main/docs/Architecture.md)


## 部署

部署到k8s集群：
```bash
kubectl apply -f deploy/k8s.yaml
```

测试案例部署：
```bash
kubectl apply -f deploy/k8s_test.yaml
```

## 使用说明

### SimpleAutoScaler

SimpleAutoScaler 是一个根据特定CRD值对目标资源的值进行修改的控制器，下面是 SimpleAutoScaler 的案例：

```yaml
kind: SimpleAutoScaler 
apiVersion: autoscale.scale.shikanon.com/v1alpha1
metadata:
  name: autoscale-test-custom
  namespace: default
spec:
  targetRef: # 扩缩容目标资源对象，指被控制进行扩缩容的资源对象
    kind: Deployment # 目标资源类型
    apiVersion: apps/v1 # 目标资源apiVersion
    name: autoscale-test # 目标资源名称，默认是和SimpleAutoScaler在同一命名空间下的
  policy: # 针对扩缩容目标资源的扩缩容策略
  - name: "requests-cpu" # 容策略名称
    field: "spec.template.spec.containers.0.resources.requests.cpu" # 目标资源被修改的字段路径，SimpleAutoScaler 通过修改这个值来控制目标资源
    type: "cpu" # 目标资源扩缩容字段的类型，这里支持 cpu, memory， replicas 和 other
    update: # 字段的更新策略
      downscaleWindow: 1m # 字段缩容策略的更新冷却时间
      upscaleWindow: 1m # 字段扩容策略的更新冷却时间
      minAllowed: 10m # 字段的最小值
      maxAllowed: "10" # 字段的最大值
    scaleFactorObject: # 影响扩缩容的因子，这里是指向特定的CRD的字段
      kind: RecommendationScaleFactor # 影响因子的资源类型
      apiVersion: autoscale.scale.shikanon.com/v1alpha1 # 影响因子的资源apiVersion
      name: autoscale-test-custom # 影响因子的资源名称
      namespace: default  # 影响因子的命名空间
      field: "status.scaleFactor"  # 影响因子的字段
```

### RecommendationScaleFactor

RecommendationScaleFactor 是一个用来计算扩缩容指标的控制器，可以根据用户自定义的指标进行计算出扩缩容比例，由 SimpleAutoScaler 采集执行, RecommendationScaleFactor 案例：

```yaml
kind: RecommendationScaleFactor
apiVersion: autoscale.scale.shikanon.com/v1alpha1
metadata:
  name: autoscale-test-cpu
  namespace: default
spec:
  # query 表示指标的查询语句
  query: |
    100 * avg (rate (container_cpu_usage_seconds_total{
    image!="",container="autoscale-test",pod=~"^autoscale-test()().*$",kubernetes_io_hostname=~"^.*$"}[5m]
    )) /avg (kube_pod_container_resource_requests_cpu_cores{pod=~"^autoscale-test()().*$"})
  desiredValue: "1" # 查询语句计算出来的目标期望值
  cooldown: "30s" # 更新时间
  minScope: 10 # 最小变化幅度,百分比值%
  maxScope: 300 # 最大变化幅度,百分比值%
```