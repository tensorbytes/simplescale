#!/bin/bash
component=$1
version=$2

function buildWebhook(){
    docker build -t registry.cn-beijing.aliyuncs.com/shikanon/simplescaler-webhook:$version -f ./deploy/webhook/Dockerfile .
    echo "image: " registry.cn-beijing.aliyuncs.com/shikanon/simplescaler-webhook:$version
}

function buildUpdate(){
    docker build -t registry.cn-beijing.aliyuncs.com/shikanon/simplescaler-update:$version -f ./deploy/updater/Dockerfile .
    echo "image: " registry.cn-beijing.aliyuncs.com/shikanon/simplescaler-update:$version
}

function buildRecommond(){
    docker build -t registry.cn-beijing.aliyuncs.com/shikanon/simplescaler-recommender:$version -f ./deploy/recommender/Dockerfile .
    echo "image: " registry.cn-beijing.aliyuncs.com/shikanon/simplescaler-recommender:$version
}

if [[ $component = "webhook" ]];
then
    buildWebhook
elif [[ $component = "updater" ]];
then
    buildUpdate
elif [[ $component = "recommender" ]];
then
    buildRecommond
elif [[ $component = "all" ]];
then
    buildWebhook
    buildUpdate
    buildRecommond
fi