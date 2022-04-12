package tests

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
)

func loadCA(caFile string) *x509.CertPool {
	pool := x509.NewCertPool()

	if ca, e := ioutil.ReadFile(caFile); e != nil {
		log.Fatal("ReadFile: ", e)
	} else {
		pool.AppendCertsFromPEM(ca)
	}
	return pool
}

func TestWebhook(t *testing.T) {
	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: loadCA("server.crt")},
		}}
	url := "https://simpleautoscale-webhook.operator.svc:8000/mutating-resource"
	contentType := "application/json"
	bodystring := `{
		"apiVersion": "admission.k8s.io/v1",
		"kind": "AdmissionReview",
		"request": {
			"uid": "f6d43324-e48c-4a02-a366-f67118ffe2ab",
			"kind": {
				"group": "apps",
				"version": "v1",
				"kind": "Deployment"
			},
			"resource": {
				"group": "apps",
				"version": "v1",
				"resource": "deployments"
			},
			"requestKind": {
				"group": "apps",
				"version": "v1",
				"kind": "Deployment"
			},
			"requestResource": {
				"group": "apps",
				"version": "v1",
				"resource": "deployments"
			},
			"name": "autoscale-test",
			"namespace": "default",
			"operation": "UPDATE",
			"userInfo": {
				"username": "user-f9mr4",
				"groups": ["system:authenticated", "system:cattle:authenticated"]
			},
			"object": {
				"kind": "Deployment",
				"apiVersion": "apps/v1",
				"metadata": {
					"name": "autoscale-test",
					"namespace": "default",
					"uid": "804db24f-7ce5-4870-8075-55b66fd290f2",
					"resourceVersion": "409197734",
					"generation": 7289,
					"creationTimestamp": "2022-02-08T02:31:14Z",
					"labels": {
						"app": "autoscale-test"
					},
					"annotations": {
						"deployment.kubernetes.io/revision": "7289",
						"kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"labels\":{\"app\":\"autoscale-test\"},\"name\":\"autoscale-test\",\"namespace\":\"default\"},\"spec\":{\"selector\":{\"matchLabels\":{\"app\":\"autoscale-test\"}},\"template\":{\"metadata\":{\"annotations\":{\"prometheus.io/port\":\"8000\",\"prometheus.io/scrape\":\"true\"},\"labels\":{\"app\":\"autoscale-test\"}},\"spec\":{\"containers\":[{\"command\":[\"/app/random-metrics\",\"--listen-address=:8000\"],\"image\":\"docker.io/shikanon096/random-prometheus-metrics:latest\",\"imagePullPolicy\":\"Always\",\"name\":\"autoscale-test\",\"ports\":[{\"containerPort\":8000,\"name\":\"metrics\",\"protocol\":\"TCP\"}],\"resources\":{\"limits\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"},\"requests\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"}}}]}}}}\n"
					},
					"managedFields": [{
						"manager": "kubectl-client-side-apply",
						"operation": "Update",
						"apiVersion": "apps/v1",
						"time": "2022-03-22T13:42:23Z",
						"fieldsType": "FieldsV1",
						"fieldsV1": {
							"f:metadata": {
								"f:annotations": {
									".": {},
									"f:kubectl.kubernetes.io/last-applied-configuration": {}
								},
								"f:labels": {
									".": {},
									"f:app": {}
								}
							},
							"f:spec": {
								"f:progressDeadlineSeconds": {},
								"f:replicas": {},
								"f:revisionHistoryLimit": {},
								"f:selector": {
									"f:matchLabels": {
										".": {},
										"f:app": {}
									}
								},
								"f:strategy": {
									"f:rollingUpdate": {
										".": {},
										"f:maxSurge": {},
										"f:maxUnavailable": {}
									},
									"f:type": {}
								},
								"f:template": {
									"f:metadata": {
										"f:annotations": {
											".": {},
											"f:prometheus.io/port": {},
											"f:prometheus.io/scrape": {}
										},
										"f:labels": {
											".": {},
											"f:app": {}
										}
									},
									"f:spec": {
										"f:containers": {
											"k:{\"name\":\"autoscale-test\"}": {
												".": {},
												"f:command": {},
												"f:image": {},
												"f:imagePullPolicy": {},
												"f:name": {},
												"f:ports": {
													".": {},
													"k:{\"containerPort\":8000,\"protocol\":\"TCP\"}": {
														".": {},
														"f:containerPort": {},
														"f:name": {},
														"f:protocol": {}
													}
												},
												"f:resources": {
													".": {},
													"f:limits": {
														".": {},
														"f:cpu": {},
														"f:memory": {}
													},
													"f:requests": {
														".": {},
														"f:memory": {}
													}
												},
												"f:terminationMessagePath": {},
												"f:terminationMessagePolicy": {}
											}
										},
										"f:dnsPolicy": {},
										"f:restartPolicy": {},
										"f:schedulerName": {},
										"f:securityContext": {},
										"f:terminationGracePeriodSeconds": {}
									}
								}
							}
						}
					}, {
						"manager": "kube-controller-manager",
						"operation": "Update",
						"apiVersion": "apps/v1",
						"time": "2022-03-22T13:45:08Z",
						"fieldsType": "FieldsV1",
						"fieldsV1": {
							"f:metadata": {
								"f:annotations": {
									"f:deployment.kubernetes.io/revision": {}
								}
							},
							"f:status": {
								"f:availableReplicas": {},
								"f:conditions": {
									".": {},
									"k:{\"type\":\"Available\"}": {
										".": {},
										"f:lastTransitionTime": {},
										"f:lastUpdateTime": {},
										"f:message": {},
										"f:reason": {},
										"f:status": {},
										"f:type": {}
									},
									"k:{\"type\":\"Progressing\"}": {
										".": {},
										"f:lastTransitionTime": {},
										"f:lastUpdateTime": {},
										"f:message": {},
										"f:reason": {},
										"f:status": {},
										"f:type": {}
									}
								},
								"f:observedGeneration": {},
								"f:readyReplicas": {},
								"f:replicas": {},
								"f:updatedReplicas": {}
							}
						}
					}, {
						"manager": "kubectl-edit",
						"operation": "Update",
						"apiVersion": "apps/v1",
						"time": "2022-03-22T13:46:28Z",
						"fieldsType": "FieldsV1",
						"fieldsV1": {
							"f:spec": {
								"f:template": {
									"f:spec": {
										"f:containers": {
											"k:{\"name\":\"autoscale-test\"}": {
												"f:resources": {
													"f:requests": {
														"f:cpu": {}
													}
												}
											}
										}
									}
								}
							}
						}
					}]
				},
				"spec": {
					"replicas": 1,
					"selector": {
						"matchLabels": {
							"app": "autoscale-test"
						}
					},
					"template": {
						"metadata": {
							"creationTimestamp": null,
							"labels": {
								"app": "autoscale-test"
							},
							"annotations": {
								"prometheus.io/port": "8000",
								"prometheus.io/scrape": "true"
							}
						},
						"spec": {
							"containers": [{
								"name": "autoscale-test",
								"image": "docker.io/shikanon096/random-prometheus-metrics:latest",
								"command": ["/app/random-metrics", "--listen-address=:8000"],
								"ports": [{
									"name": "metrics",
									"containerPort": 8000,
									"protocol": "TCP"
								}],
								"resources": {
									"limits": {
										"cpu": "100m",
										"memory": "100Mi"
									},
									"requests": {
										"cpu": "100m",
										"memory": "50Mi"
									}
								},
								"terminationMessagePath": "/dev/termination-log",
								"terminationMessagePolicy": "File",
								"imagePullPolicy": "Always"
							}],
							"restartPolicy": "Always",
							"terminationGracePeriodSeconds": 30,
							"dnsPolicy": "ClusterFirst",
							"securityContext": {},
							"schedulerName": "default-scheduler"
						}
					},
					"strategy": {
						"type": "RollingUpdate",
						"rollingUpdate": {
							"maxUnavailable": "25%",
							"maxSurge": "25%"
						}
					},
					"revisionHistoryLimit": 10,
					"progressDeadlineSeconds": 600
				},
				"status": {
					"observedGeneration": 7289,
					"replicas": 1,
					"updatedReplicas": 1,
					"readyReplicas": 1,
					"availableReplicas": 1,
					"conditions": [{
						"type": "Available",
						"status": "True",
						"lastUpdateTime": "2022-03-12T12:03:11Z",
						"lastTransitionTime": "2022-03-12T12:03:11Z",
						"reason": "MinimumReplicasAvailable",
						"message": "Deployment has minimum availability."
					}, {
						"type": "Progressing",
						"status": "True",
						"lastUpdateTime": "2022-03-22T13:45:08Z",
						"lastTransitionTime": "2022-02-08T02:31:14Z",
						"reason": "NewReplicaSetAvailable",
						"message": "ReplicaSet \"autoscale-test-7bf946f967\" has successfully progressed."
					}]
				}
			},
			"oldObject": {
				"kind": "Deployment",
				"apiVersion": "apps/v1",
				"metadata": {
					"name": "autoscale-test",
					"namespace": "default",
					"uid": "804db24f-7ce5-4870-8075-55b66fd290f2",
					"resourceVersion": "409197734",
					"generation": 7289,
					"creationTimestamp": "2022-02-08T02:31:14Z",
					"labels": {
						"app": "autoscale-test"
					},
					"annotations": {
						"deployment.kubernetes.io/revision": "7289",
						"kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"labels\":{\"app\":\"autoscale-test\"},\"name\":\"autoscale-test\",\"namespace\":\"default\"},\"spec\":{\"selector\":{\"matchLabels\":{\"app\":\"autoscale-test\"}},\"template\":{\"metadata\":{\"annotations\":{\"prometheus.io/port\":\"8000\",\"prometheus.io/scrape\":\"true\"},\"labels\":{\"app\":\"autoscale-test\"}},\"spec\":{\"containers\":[{\"command\":[\"/app/random-metrics\",\"--listen-address=:8000\"],\"image\":\"docker.io/shikanon096/random-prometheus-metrics:latest\",\"imagePullPolicy\":\"Always\",\"name\":\"autoscale-test\",\"ports\":[{\"containerPort\":8000,\"name\":\"metrics\",\"protocol\":\"TCP\"}],\"resources\":{\"limits\":{\"cpu\":\"100m\",\"memory\":\"100Mi\"},\"requests\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"}}}]}}}}\n"
					}
				},
				"spec": {
					"replicas": 1,
					"selector": {
						"matchLabels": {
							"app": "autoscale-test"
						}
					},
					"template": {
						"metadata": {
							"creationTimestamp": null,
							"labels": {
								"app": "autoscale-test"
							},
							"annotations": {
								"prometheus.io/port": "8000",
								"prometheus.io/scrape": "true"
							}
						},
						"spec": {
							"containers": [{
								"name": "autoscale-test",
								"image": "docker.io/shikanon096/random-prometheus-metrics:latest",
								"command": ["/app/random-metrics", "--listen-address=:8000"],
								"ports": [{
									"name": "metrics",
									"containerPort": 8000,
									"protocol": "TCP"
								}],
								"resources": {
									"limits": {
										"cpu": "100m",
										"memory": "100Mi"
									},
									"requests": {
										"cpu": "10m",
										"memory": "50Mi"
									}
								},
								"terminationMessagePath": "/dev/termination-log",
								"terminationMessagePolicy": "File",
								"imagePullPolicy": "Always"
							}],
							"restartPolicy": "Always",
							"terminationGracePeriodSeconds": 30,
							"dnsPolicy": "ClusterFirst",
							"securityContext": {},
							"schedulerName": "default-scheduler"
						}
					},
					"strategy": {
						"type": "RollingUpdate",
						"rollingUpdate": {
							"maxUnavailable": "25%",
							"maxSurge": "25%"
						}
					},
					"revisionHistoryLimit": 10,
					"progressDeadlineSeconds": 600
				},
				"status": {
					"observedGeneration": 7289,
					"replicas": 1,
					"updatedReplicas": 1,
					"readyReplicas": 1,
					"availableReplicas": 1,
					"conditions": [{
						"type": "Available",
						"status": "True",
						"lastUpdateTime": "2022-03-12T12:03:11Z",
						"lastTransitionTime": "2022-03-12T12:03:11Z",
						"reason": "MinimumReplicasAvailable",
						"message": "Deployment has minimum availability."
					}, {
						"type": "Progressing",
						"status": "True",
						"lastUpdateTime": "2022-03-22T13:45:08Z",
						"lastTransitionTime": "2022-02-08T02:31:14Z",
						"reason": "NewReplicaSetAvailable",
						"message": "ReplicaSet \"autoscale-test-7bf946f967\" has successfully progressed."
					}]
				}
			},
			"dryRun": false,
			"options": {
				"kind": "UpdateOptions",
				"apiVersion": "meta.k8s.io/v1",
				"fieldManager": "kubectl-edit"
			}
		}
	  }`
	body := strings.NewReader(bodystring)
	resp, err := c.Post(url, contentType, body)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(respbody)
}
