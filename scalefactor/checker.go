package scalefactor

import (
	"context"
	"fmt"

	autoscalev1alpha1 "github.com/tensorbytes/simplescale/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultDesiredValue = "50"
	DefaultCooldown     = "30s"
	DefaultMinScope     = 20
)

type Checker interface {
	CheckAndSetDefault(*autoscalev1alpha1.RecommendationScaleFactor) error
}

type RecommendationScaleFactorChecker struct {
	recommenderClient client.Client
}

func (c *RecommendationScaleFactorChecker) CheckAndSetDefault(factor *autoscalev1alpha1.RecommendationScaleFactor) error {
	var fixed bool
	if factor.Spec.Cooldown == "" {
		factor.Spec.Cooldown = DefaultCooldown
		fixed = true
	}
	if factor.Spec.DesiredValue == "" {
		factor.Spec.DesiredValue = DefaultDesiredValue
		fixed = true
	}
	if factor.Spec.MinScope == 0 {
		factor.Spec.MinScope = DefaultMinScope
		fixed = true
	}
	if fixed {
		if err := c.recommenderClient.Patch(context.TODO(), factor, client.Merge); err != nil {
			return fmt.Errorf("RecommendationScaleFactor spec update for default value error: %v", err)
		}
	}
	return nil
}
