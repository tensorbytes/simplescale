package scalefactor

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	autoscalev1alpha1 "github.com/tensorbytes/simplescale/api/v1alpha1"
	"github.com/tensorbytes/simplescale/scalefactor/history"
	"github.com/tensorbytes/simplescale/scalefactor/metric"
	"github.com/tensorbytes/simplescale/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klogv2 "k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Recommender interface {
	RunOnce()
	GetRecommendedResourceValue()
	Update()
	Checkpoints()
}

func NewSimpleRecommender() (*SimpleRecommender, error) {
	kubeClient, err := utils.GetControllerClient()
	if err != nil {
		return nil, err
	}
	checker := RecommendationScaleFactorChecker{
		recommenderClient: kubeClient,
	}
	return &SimpleRecommender{
		recommenderClient: kubeClient,
		checker:           &checker,
	}, nil
}

type SimpleRecommender struct {
	collector         history.HistoryProvider
	algorithm         ScaleAlgorithm
	checker           Checker
	recommenderClient client.Client
}

func (r *SimpleRecommender) RegisterCollector(collector history.HistoryProvider) {
	r.collector = collector
}

func (r *SimpleRecommender) RegisterAlgorithm(algorithm ScaleAlgorithm) {
	r.algorithm = algorithm
}

// run once
func (r *SimpleRecommender) RunOnce(ctx context.Context) {
	var recommendationScaleFactorList autoscalev1alpha1.RecommendationScaleFactorList
	err := r.recommenderClient.List(ctx, &recommendationScaleFactorList, &client.ListOptions{})
	if err != nil {
		klogv2.Errorf("list resource of RecommendationScaleFactor error: %v", err)
		return
	}
	// for the recommendationScaleFactor resources
	for _, scaleFactorItem := range recommendationScaleFactorList.Items {
		queryResult, err := r.GetRecommendResourceValue(ctx, scaleFactorItem)
		if err != nil {
			// update QueryExist condition
			condition := autoscalev1alpha1.RecommendationScaleFactorCondition{
				Type:    autoscalev1alpha1.QueryExistCondition,
				Status:  "false",
				Message: err.Error(),
			}
			scaleFactorItem.Status.AppendConditions(condition)
			klogv2.Errorf("controller get recommend resource value failed: %v; queryResult: %s", err, queryResult.String())
			go r.UpdateRecommendationScaleFactorStatus(ctx, scaleFactorItem)
			continue
		} else {
			condition := autoscalev1alpha1.RecommendationScaleFactorCondition{
				Type:   autoscalev1alpha1.QueryExistCondition,
				Status: "true",
			}
			scaleFactorItem.Status.AppendConditions(condition)
		}
		// recommend
		result := r.GetRecommendedResult(ctx, queryResult)
		// check and set default value
		r.checker.CheckAndSetDefault(&scaleFactorItem)
		// update
		r.Update(ctx, scaleFactorItem, result)

	}
}

func (r *SimpleRecommender) GetRecommendResourceValue(ctx context.Context, scaleFactor autoscalev1alpha1.RecommendationScaleFactor) (queryResult metric.HistoryMetric, err error) {
	statements := scaleFactor.Spec.Query
	if statements == "" {
		err = errors.New("recommendationScaleFactor query cannot be nil")
		return
	}
	rangeTime := 1 * time.Hour
	return r.collector.QueryRange(ctx, statements, rangeTime)
}

// get the recommended resource value
func (r *SimpleRecommender) GetRecommendedResult(ctx context.Context, queryResult metric.HistoryMetric) (result ModelResult) {
	// default the next stage is 1 hour
	stageTime := 1 * time.Hour
	result = r.algorithm.PredictNextStage(queryResult, stageTime)
	return result
}

// get rate of scale
func (r *SimpleRecommender) EvaluateScaleRate(ctx context.Context, scaleFactor *autoscalev1alpha1.RecommendationScaleFactor, result ModelResult) (scaleRate float64, err error) {
	currentValue := result.Value
	desiredValue, err := strconv.ParseFloat(scaleFactor.Spec.DesiredValue, 64)
	if err != nil {
		return
	}
	scaleRate = r.evaluateScaleRate(currentValue, desiredValue)
	if scaleRate <= 0 {
		return scaleRate, fmt.Errorf("scaleRate(%f) cannot be smaller than 0", scaleRate)
	}
	// scale rate change than old scale must larger than min scope
	if scaleFactor.Status.ScaleFactor != "" {
		oldScaleRate, err := strconv.ParseFloat(scaleFactor.Status.ScaleFactor, 64)
		if err != nil {
			return scaleRate, nil
		}
		minScope := float64(scaleFactor.Spec.MinScope)
		if (scaleRate > oldScaleRate*(1-minScope/100)) && (scaleRate < oldScaleRate*(1+minScope/100)) {
			scaleRate = 1
		}
	}

	return
}

// how to make the current value to desire value
func (r *SimpleRecommender) evaluateScaleRate(currentValue, desiredValue float64) (scaleRate float64) {
	scaleRate = currentValue / desiredValue
	return
}

// update the recommended value
func (r *SimpleRecommender) Update(ctx context.Context, scaleFactor autoscalev1alpha1.RecommendationScaleFactor, result ModelResult) {
	// checking cooldown time
	cooldowntime, err := time.ParseDuration(scaleFactor.Spec.Cooldown)
	if err != nil {
		klogv2.Errorf("cooldown field parse time error: %v", err)
		return
	}
	// checking condition
	if result.ErrorMessage != nil {
		condition := autoscalev1alpha1.RecommendationScaleFactorCondition{
			Type:    autoscalev1alpha1.ValidRecommendValueCondition,
			Status:  "false",
			Message: result.ErrorMessage.Error(),
		}
		scaleFactor.Status.AppendConditions(condition)
		go r.UpdateRecommendationScaleFactorStatus(ctx, scaleFactor)
		return
	} else {
		condition := autoscalev1alpha1.RecommendationScaleFactorCondition{
			Type:   autoscalev1alpha1.ValidRecommendValueCondition,
			Status: "true",
		}
		scaleFactor.Status.AppendConditions(condition)
	}

	if scaleFactor.Status.LastCoolDownTime.Add(cooldowntime).Before(time.Now()) {
		scaleFactor.Status.LastCoolDownTime = metav1.Now()
		scaleFactor.Status.CurrentValue = fmt.Sprintf("%f", result.Value)
		scaleRate, err := r.EvaluateScaleRate(ctx, &scaleFactor, result)
		if err != nil {
			klogv2.Error(err)
			return
		}
		scaleFactor.Status.ScaleFactor = fmt.Sprintf("%f", scaleRate)
		go r.UpdateRecommendationScaleFactorStatus(ctx, scaleFactor)
	}
}

// update the recommendscalefactors status
func (r *SimpleRecommender) UpdateRecommendationScaleFactorStatus(ctx context.Context, scaleFactor autoscalev1alpha1.RecommendationScaleFactor) {
	retry := 3
	err := r.updateRecommendationScaleFactorStatus(ctx, &scaleFactor, retry)
	if err != nil {
		klogv2.Error(err)
	}
}

func (r *SimpleRecommender) updateRecommendationScaleFactorStatus(ctx context.Context, scaleFactor *autoscalev1alpha1.RecommendationScaleFactor, retry int) error {
	if err := r.recommenderClient.Status().Update(ctx, scaleFactor, &client.UpdateOptions{}); err != nil {
		if apierrors.IsConflict(err) {
			retry = retry - 1
			if retry > 0 {
				err = r.updateRecommendationScaleFactorStatus(ctx, scaleFactor, retry)
			}
		}
		return fmt.Errorf("RecommendationScaleFactor status update error: %v", err)
	}
	return nil
}

// update the recommendscalefactors spec
func (r *SimpleRecommender) updateRecommendationScaleFactorSpec(ctx context.Context, scaleFactor *autoscalev1alpha1.RecommendationScaleFactor) error {
	if err := r.recommenderClient.Patch(ctx, scaleFactor, client.Merge); err != nil {
		return fmt.Errorf("RecommendationScaleFactor spec update error: %v", err)
	}
	return nil
}
