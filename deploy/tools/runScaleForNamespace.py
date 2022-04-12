
from kubernetes import dynamic, config
from kubernetes.client import api_client
import click

@click.group()
def cli():
    pass


k8sclient = dynamic.DynamicClient(api_client.ApiClient(configuration=config.load_kube_config()))

deployAPI = k8sclient.resources.get(api_version="apps/v1", kind="Deployment")
scaleAPI = k8sclient.resources.get(api_version="autoscale.scale.shikanon.com/v1alpha1", kind="SimpleAutoScaler")
factorAPI = k8sclient.resources.get(api_version="autoscale.scale.shikanon.com/v1alpha1", kind="RecommendationScaleFactor")

query = """100 * avg (rate (container_cpu_usage_seconds_total{{
                            image!="",container="{name}",pod=~"^{name}-[0-9a-zA-Z]+-[0-9a-zA-Z]+",kubernetes_io_hostname=~"^.*$"}}[5m]
                            )) /avg (kube_pod_container_resource_requests_cpu_cores{{pod=~"^{name}-[0-9a-zA-Z]+-[0-9a-zA-Z]+"}})"""


def createObj(namespace):
    for item in deployAPI.get().items:
        if item.metadata.namespace == namespace:
            deployment = deployAPI.get(name=item.metadata.name,namespace=item.metadata.namespace)
            print(deployment.metadata.name)
            facotrObject = {
                "apiVersion": "autoscale.scale.shikanon.com/v1alpha1",
                "kind": "RecommendationScaleFactor",
                "metadata": {
                    "name": deployment.metadata.name+"-factor",
                    "namespace": deployment.metadata.namespace,
                },
                "spec": {
                    "query": query.format(name=deployment.metadata.name),
                    "desiredValue": "100",
                    "cooldown": "15s",
                    "minScope": 10,
                    "maxScope": 300,
                },
            }
            scaleObject = {
                "apiVersion": "autoscale.scale.shikanon.com/v1alpha1",
                "kind": "SimpleAutoScaler",
                "metadata": {
                    "name": deployment.metadata.name+"-scaler",
                    "namespace": deployment.metadata.namespace,
                },
                "spec": {
                    "targetRef": {
                        "kind": "Deployment",
                        "apiVersion": "apps/v1",
                        "name": deployment.metadata.name,
                    },
                    "policy": [
                        {
                            "name": "requests-cpu",
                            "field": "spec.template.spec.containers.0.resources.requests.cpu",
                            "type": "cpu",
                            "update": {
                                "downscaleWindow": "3h",
                                "upscaleWindow": "1h",
                                "minAllowed": "10m",
                                "maxAllowed": "10",
                            },
                            "scaleFactorObject": {
                                "kind": "RecommendationScaleFactor",
                                "apiVersion": "autoscale.scale.shikanon.com/v1alpha1",
                                "name": deployment.metadata.name+"-factor",
                                "namespace": deployment.metadata.namespace,
                                "field": "status.scaleFactor",
                            },
                        },
                    ],
                },
            }
            factorAPI.create(facotrObject)
            scaleAPI.create(scaleObject)




def deleteObj():
    for item in scaleAPI.get().items:
        print(scaleAPI.delete(name=item.metadata.name,namespace=item.metadata.namespace))

    for item in factorAPI.get().items:
        print(factorAPI.delete(name=item.metadata.name,namespace=item.metadata.namespace))

@cli.command()
@click.option('--namespace', default="default", help='the namespace of create simpleautoscaler.')
def create(namespace):
    """Create simpleautoscaler in special namespace."""
    createObj(namespace)

@cli.command()
def delete():
    """Delete simpleautoscaler in all namespace."""
    deleteObj()

if __name__ == '__main__':
    cli()