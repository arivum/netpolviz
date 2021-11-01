### NetpolViz

Before using NetpolViz, please read our [documentation](https://github.com/arivum/netpolviz)


#### Available chart properties

| Property       |   Description         | Default               |
| -------------- | --------------------- | --------------------- |
| `replicaCount` | Number of replicas    | `1`                   |
| `image.repository`  | Image repository | `ghcr.io/arivum/netpolviz` |
| `image.tag`  | Image tag  | `0.1.0` |
| `image.pullPolicy`  | Image pull policy | `IfNotPresent` |
| `imagePullSecrets`  | A list of image pull secrets when using your own image | `[]` |
| `nameOverride`  | Overrides the basename for deployments etc. | `""` |
| `fullnameOverride` | Completly overrides the auto-generated names for deployments etc. | `""` |
| `serviceAccount.create` | Enable serviceAccount creation | `true` |
| `serviceAccount.annotations` | Additional annotations for created serviceAccount | `{}` |
| `serviceAccount.name` | Overrides the generated serviceAccount name | `""` |
| `podAnnotations` | Additional pod annotations | `{}` |
| `podSecurityContext` | Specify security context options that will be applied to the pod | `{}` |
| `securityContext` | Specify security context options that will be applied to all containers inside the deployed pod | `{}` |
| `service.type` | Kubernetes serivce type | `ClusterIP` |
| `service.port` | Kubernetes service port | `443` |
| `ingress.enabled` | Enable installation of an ingress object | `false` |
| `ingress.annotations` | Additonal annotations for the ingress object | `{}` |
| `ingress.hosts` | A list of hosts that the ingress object serves. | `{"host": "chart-example.local", "paths": []}` |
| `ingress.tls` | A list of kubernetes ingress TLS configurations | `[]` |
| `resources` | Pod resource requests and limits | `{}` |
| `nodeSelector` | Kubernetes node selector | `{}` |
| `tolerations` | Kuberentes tolerations | `[]` |
| `affinity` | Kubernetes pod affinity | `{}` |