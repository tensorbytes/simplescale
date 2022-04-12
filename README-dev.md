# 开发文档

---
初始化：
```bash
kubebuilder init --domain scale.shikanon.com --repo github.com/tensorbytes/simplescale
```

## 添加新的api
```bash
kubebuilder create api --group autoscale --version v1alpha1 --kind SimpleAutoScaler
```

```bash
kubebuilder create api --group autoscale --version v1alpha1 --kind RecommendationScaleFactor
```

## 创建CRD
```bash
controller-gen "crd:trivialVersions=true,preserveUnknownFields=false" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
```

```
controller-gen object:headerFile="hack\\boilerplate.go.txt" paths="./..."
```
